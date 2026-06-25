package nand16

import (
	"fmt"
	"runtime"
	"sync"
	"unsafe"
)

// FlatSimulator is a high-performance drop-in replacement for Simulator.
// It flattens pointer-based Wire/Gate structures into contiguous arrays
// and evaluates independent gates in parallel using goroutines.
//
// Key optimisations over the pointer-based Simulator:
//   - Wire values stored as byte array (0/1), enabling branchless NAND
//   - Gate operands packed as [a,b,out,a,b,out,...] for cache locality
//   - Bounds-check-free inner loop via unsafe pointer arithmetic
//   - Level-partitioned parallel evaluation for multi-core
type FlatSimulator struct {
	// Wire values: wires[i] is 0 or 1 (byte, not bool — enables arithmetic NAND).
	wires []byte

	// Packed gate operands in topo order: [a0, b0, out0, a1, b1, out1, ...].
	// Using int avoids uint32→int conversion on every array index.
	gates []int
	nGate int // number of gates = len(gates)/3

	// Level boundaries for parallel mode: level i spans
	// gates 3*levelStart[i] .. 3*levelStart[i+1] in the packed array.
	levelStart []int

	// Flip-flop operands as parallel int arrays.
	ffD  []int
	ffQ  []int
	ffQn []int

	// Wire-pointer-to-index map for WireIdx/BusIdx.
	wireID map[*Wire]int

	numWorkers int
	wg         sync.WaitGroup

	Cycles int
}

// NewFlatSimulator builds a flattened simulator from an existing Simulator.
func NewFlatSimulator(src *Simulator, workers int) *FlatSimulator {
	sorted := src.Gates
	ffs := src.FFs

	// --- Assign unique wire IDs ---
	wireID := make(map[*Wire]int)
	var wireList []*Wire
	assign := func(w *Wire) {
		if _, ok := wireID[w]; !ok {
			wireID[w] = len(wireList)
			wireList = append(wireList, w)
		}
	}
	for _, g := range sorted {
		assign(g.A); assign(g.B); assign(g.Out)
	}
	for _, ff := range ffs {
		assign(ff.D); assign(ff.Q); assign(ff.Qn)
	}

	// --- Wire byte array ---
	nWires := len(wireList)
	wires := make([]byte, nWires)
	for i, w := range wireList {
		if w.Val {
			wires[i] = 1
		}
	}

	// --- Packed gate array (topo order) ---
	n := len(sorted)
	gatesPacked := make([]int, n*3)
	for i, g := range sorted {
		j := i * 3
		gatesPacked[j] = wireID[g.A]
		gatesPacked[j+1] = wireID[g.B]
		gatesPacked[j+2] = wireID[g.Out]
	}

	// --- Topological depth ---
	producer := make(map[int]int, n) // out wire ID → gate index
	for i := 0; i < n; i++ {
		producer[gatesPacked[i*3+2]] = i
	}
	depth := make([]int, n)
	maxDepth := 0
	for i := 0; i < n; i++ {
		d := 0
		aID := gatesPacked[i*3]
		bID := gatesPacked[i*3+1]
		if pi, ok := producer[aID]; ok && depth[pi]+1 > d {
			d = depth[pi] + 1
		}
		if pi, ok := producer[bID]; ok && depth[pi]+1 > d {
			d = depth[pi] + 1
		}
		depth[i] = d
		if d > maxDepth {
			maxDepth = d
		}
	}

	// --- Sort by level (stable) ---
	levelCount := make([]int, maxDepth+1)
	for _, d := range depth {
		levelCount[d]++
	}
	levelStart := make([]int, maxDepth+2)
	for i := 1; i <= maxDepth+1; i++ {
		levelStart[i] = levelStart[i-1] + levelCount[i-1]
	}
	sortedPacked := make([]int, n*3)
	pos := make([]int, maxDepth+1)
	copy(pos, levelStart[:maxDepth+1])
	for i := 0; i < n; i++ {
		j := pos[depth[i]]
		k := j * 3
		s := i * 3
		sortedPacked[k] = gatesPacked[s]
		sortedPacked[k+1] = gatesPacked[s+1]
		sortedPacked[k+2] = gatesPacked[s+2]
		pos[depth[i]]++
	}

	// --- Flat FF arrays ---
	nFF := len(ffs)
	ffD := make([]int, nFF)
	ffQ := make([]int, nFF)
	ffQn := make([]int, nFF)
	for i, ff := range ffs {
		ffD[i] = wireID[ff.D]
		ffQ[i] = wireID[ff.Q]
		ffQn[i] = wireID[ff.Qn]
	}

	if workers <= 0 {
		workers = runtime.NumCPU()
	}

	return &FlatSimulator{
		wires:      wires,
		gates:      sortedPacked,
		nGate:      n,
		levelStart: levelStart,
		ffD:        ffD,
		ffQ:        ffQ,
		ffQn:       ffQn,
		wireID:     wireID,
		numWorkers: workers,
	}
}

// --- Direct array access for GateCPU ---

func (fs *FlatSimulator) WireIdx(w *Wire) int   { return fs.wireID[w] }
func (fs *FlatSimulator) BusIdx(b Bus) []int {
	idx := make([]int, len(b))
	for i, w := range b {
		idx[i] = fs.wireID[w]
	}
	return idx
}

func (fs *FlatSimulator) WireVal(idx int) bool  { return fs.wires[idx] != 0 }
func (fs *FlatSimulator) SetWireVal(idx int, v bool) {
	if v { fs.wires[idx] = 1 } else { fs.wires[idx] = 0 }
}

func (fs *FlatSimulator) GetBusVal(idx []int) int {
	v := 0
	for i, id := range idx {
		v |= int(fs.wires[id]) << i
	}
	return v
}

func (fs *FlatSimulator) SetBusVal(idx []int, v int) {
	for i, id := range idx {
		fs.wires[id] = byte((v >> i) & 1)
	}
}

// --- SimEngine ---

func (fs *FlatSimulator) Settle() {
	if fs.numWorkers <= 1 {
		fs.settleSingle()
	} else {
		fs.settleParallel()
	}
}

// settleSingle: branchless NAND using byte arithmetic, bounds-check-free via unsafe.
//
//go:nosplit
func (fs *FlatSimulator) settleSingle() {
	wires := fs.wires
	gates := fs.gates
	n3 := fs.nGate * 3

	// unsafe base pointer to eliminate per-access bounds checks
	wp := unsafe.Pointer(&wires[0])

	for i := 0; i < n3; i += 3 {
		a := *(*byte)(unsafe.Add(wp, gates[i]))
		b := *(*byte)(unsafe.Add(wp, gates[i+1]))
		*(*byte)(unsafe.Add(wp, gates[i+2])) = 1 - (a & b)
	}
}

func (fs *FlatSimulator) settleParallel() {
	wires := fs.wires
	gates := fs.gates
	nLevels := len(fs.levelStart) - 1

	const parallelThreshold = 256

	wp := unsafe.Pointer(&wires[0])

	for lv := 0; lv < nLevels; lv++ {
		start := fs.levelStart[lv] * 3
		end := fs.levelStart[lv+1] * 3
		width := (end - start) / 3

		if width < parallelThreshold {
			for i := start; i < end; i += 3 {
				a := *(*byte)(unsafe.Add(wp, gates[i]))
				b := *(*byte)(unsafe.Add(wp, gates[i+1]))
				*(*byte)(unsafe.Add(wp, gates[i+2])) = 1 - (a & b)
			}
		} else {
			fs.evalParallel(start, end, wp)
		}
	}
}

func (fs *FlatSimulator) evalParallel(start, end int, wp unsafe.Pointer) {
	gates := fs.gates
	width := (end - start) / 3
	nw := fs.numWorkers
	if nw > width {
		nw = width
	}
	// Round chunk to multiple of 3
	chunk := (width / nw) * 3

	fs.wg.Add(nw)
	for w := 0; w < nw; w++ {
		lo := start + w*chunk
		hi := lo + chunk
		if w == nw-1 {
			hi = end
		}
		go func(lo, hi int) {
			for i := lo; i < hi; i += 3 {
				a := *(*byte)(unsafe.Add(wp, gates[i]))
				b := *(*byte)(unsafe.Add(wp, gates[i+1]))
				*(*byte)(unsafe.Add(wp, gates[i+2])) = 1 - (a & b)
			}
			fs.wg.Done()
		}(lo, hi)
	}
	fs.wg.Wait()
}

func (fs *FlatSimulator) TickFFs() {
	wires := fs.wires
	for i := range fs.ffD {
		v := wires[fs.ffD[i]]
		wires[fs.ffQ[i]] = v
		wires[fs.ffQn[i]] = 1 - v
	}
}

func (fs *FlatSimulator) Cycle() {
	fs.Settle()
	fs.TickFFs()
	fs.Cycles++
}

func (fs *FlatSimulator) Run(n int) {
	for i := 0; i < n; i++ {
		fs.Cycle()
	}
}

func (fs *FlatSimulator) Stats() string {
	nLevels := len(fs.levelStart) - 1
	maxW := 0
	for lv := 0; lv < nLevels; lv++ {
		if w := fs.levelStart[lv+1] - fs.levelStart[lv]; w > maxW {
			maxW = w
		}
	}
	return fmt.Sprintf("Gates: %d, FFs: %d, Wires: %d, Levels: %d, MaxLevelWidth: %d, Workers: %d",
		fs.nGate, len(fs.ffD), len(fs.wires), nLevels, maxW, fs.numWorkers)
}
