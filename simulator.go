package nand16

import "fmt"

// Simulator runs cycle-based simulation on a flattened gate list.
type Simulator struct {
	Gates  []*Gate     // topologically sorted
	FFs    []*FlipFlop
	Cycles int
}

// NewSimulator flattens the module and topologically sorts gates.
func NewSimulator(top *Module) *Simulator {
	gates := top.AllGates()
	ffs := top.AllFFs()
	sorted := topoSort(gates)
	return &Simulator{Gates: sorted, FFs: ffs}
}

// Cycle runs one clock cycle: evaluate all gates, then tick FFs.
func (s *Simulator) Cycle() {
	for _, g := range s.Gates {
		g.Eval()
	}
	for _, ff := range s.FFs {
		ff.Tick()
	}
	s.Cycles++
}

// Run executes n clock cycles.
func (s *Simulator) Run(n int) {
	for i := 0; i < n; i++ {
		s.Cycle()
	}
}

// Settle evaluates combinational logic until stable (no FF tick).
// Useful for testing combinational circuits.
func (s *Simulator) Settle() {
	for _, g := range s.Gates {
		g.Eval()
	}
}

// Stats returns gate/FF counts.
func (s *Simulator) Stats() string {
	return fmt.Sprintf("Gates: %d, FFs: %d, Cycles: %d", len(s.Gates), len(s.FFs), s.Cycles)
}

// topoSort sorts gates so that a gate's inputs are evaluated before it.
func topoSort(gates []*Gate) []*Gate {
	// Map wire -> gate that produces it
	producer := make(map[*Wire]*Gate)
	for _, g := range gates {
		producer[g.Out] = g
	}

	// Build adjacency: gate -> gates that depend on it
	gateIndex := make(map[*Gate]int)
	for i, g := range gates {
		gateIndex[g] = i
	}

	n := len(gates)
	inDeg := make([]int, n)
	dependents := make([][]int, n) // dependents[i] = list of gate indices that depend on gate i

	for i, g := range gates {
		for _, w := range []*Wire{g.A, g.B} {
			if p, ok := producer[w]; ok {
				pi := gateIndex[p]
				dependents[pi] = append(dependents[pi], i)
				inDeg[i]++
			}
		}
	}

	// Kahn's algorithm
	queue := make([]int, 0, n)
	for i := 0; i < n; i++ {
		if inDeg[i] == 0 {
			queue = append(queue, i)
		}
	}

	sorted := make([]*Gate, 0, n)
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		sorted = append(sorted, gates[cur])
		for _, dep := range dependents[cur] {
			inDeg[dep]--
			if inDeg[dep] == 0 {
				queue = append(queue, dep)
			}
		}
	}

	if len(sorted) != n {
		panic(fmt.Sprintf("combinational loop detected: sorted %d of %d gates", len(sorted), n))
	}
	return sorted
}
