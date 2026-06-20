package nand16

// GateCPU is a single-cycle NAND-16 processor built entirely from the gate
// primitives in this package: a PC register, an 8x16 register file, the ALU,
// adders/comparators and a control unit decoded from the opcode.
//
// Memory is modeled as a Go []byte driven by Step: the gate core exposes an
// address/data/control port (the same role as MemoryBus), and the driver
// performs the actual byte access between settle phases, mirroring how
// nand2tetris treats RAM/ROM as built-in chips.
//
// Implemented: ALU, MUL (signed, low/high), ADDI/ANDI/ORI, LUI, LW/SW/LB/SB,
// BEQ/BNE/BLT/BGE, JAL, JALR, HALT, NOP. SYSCALL has no hardware effect (it is
// a NOP); I/O is performed by the OS in software via MMIO.
type GateCPU struct {
	sim *Simulator

	pc       Bus    // program counter (FF outputs)
	q        [8]Bus // register file outputs (q[0] is constant 0)
	instr    Bus    // instruction input (set by the fetch driver)
	memData  Bus    // load-data input (set by the load driver)
	memAddr  Bus    // data memory address output
	memWData Bus    // data memory write-data output
	memRead  *Wire
	memWrite *Wire
	byteMode *Wire
	halt     *Wire
}

func and4(m *Module, a, b, c, d *Wire) *Wire {
	return AND(m, AND(m, a, b), AND(m, c, d))
}

// NewGateCPU builds the datapath and returns a ready-to-step CPU.
func NewGateCPU() *GateCPU {
	m := NewModule("gatecpu")
	zero := NewWire() // constant 0

	instr := NewBus(16)
	memData := NewBus(16)
	pc := NewBus(16)

	// --- Opcode decode: 4-to-16 one-hot from instr[15:12] ---
	ob := [4]*Wire{instr[12], instr[13], instr[14], instr[15]}
	nb := [4]*Wire{NOT(m, ob[0]), NOT(m, ob[1]), NOT(m, ob[2]), NOT(m, ob[3])}
	var isOp [16]*Wire
	for k := 0; k < 16; k++ {
		sel := func(bit int) *Wire {
			if k&(1<<bit) != 0 {
				return ob[bit]
			}
			return nb[bit]
		}
		isOp[k] = and4(m, sel(0), sel(1), sel(2), sel(3))
	}
	isALU, isMUL := isOp[0x0], isOp[0x1]
	isADDI, isANDI, isORI := isOp[0x2], isOp[0x3], isOp[0x4]
	isLUI := isOp[0x5]
	isLW, isSW, isLB, isSB := isOp[0x6], isOp[0x7], isOp[0x8], isOp[0x9]
	isBEQ, isBNE, isBLT, isBGE := isOp[0xA], isOp[0xB], isOp[0xC], isOp[0xD]
	isJAL, isSYS := isOp[0xE], isOp[0xF]

	// --- SYSTEM sub-decode from func = instr[2:0] ---
	// func 0=JALR, 1=HALT; 2=SYSCALL and 7=NOP have no hardware effect (I/O is
	// done by the OS via MMIO), so they need no decode here.
	f0, f1, f2 := instr[0], instr[1], instr[2]
	funcIs0 := AND3(m, NOT(m, f0), NOT(m, f1), NOT(m, f2))
	funcIs1 := AND3(m, f0, NOT(m, f1), NOT(m, f2))
	isJALR := AND(m, isSYS, funcIs0)
	isHALT := AND(m, isSYS, funcIs1)

	// --- Derived control ---
	isStore := OR(m, isSW, isSB)
	isLoad := OR(m, isLW, isLB)
	isBranch := OR(m, OR(m, isBEQ, isBNE), OR(m, isBLT, isBGE))
	isLink := OR(m, isJAL, isJALR)
	byteMode := OR(m, isLB, isSB)
	storeOrBranch := OR(m, isStore, isBranch)
	useImm := OR(m, OR(m, OR(m, isADDI, isANDI), OR(m, isORI, isLW)),
		OR(m, isSW, OR(m, isLB, isSB)))
	writeEnable := OR(m,
		OR(m, OR(m, isALU, isMUL), OR(m, isADDI, isANDI)),
		OR(m, OR(m, isORI, isLUI), OR(m, isLW, OR(m, isLB, isLink))))

	// --- Register file outputs (q[0] hardwired to 0) ---
	var q [8]Bus
	q[0] = NewBusConst(16, 0)
	for i := 1; i < 8; i++ {
		q[i] = NewBus(16)
	}

	// Instruction register fields.
	rdField := Bus{instr[9], instr[10], instr[11]}
	rs1Field := Bus{instr[6], instr[7], instr[8]}
	rs2Field := Bus{instr[3], instr[4], instr[5]}

	// Read-address selection: stores/branches read rd (and rs1); others rs1/rs2.
	readAddr1 := make(Bus, 3)
	Mux2Bus(m, storeOrBranch, rs1Field, rdField, readAddr1)
	readAddr2 := make(Bus, 3)
	Mux2Bus(m, storeOrBranch, rs2Field, rs1Field, readAddr2)

	rd1 := Mux8Bus(m, readAddr1, q[0], q[1], q[2], q[3], q[4], q[5], q[6], q[7])
	rd2 := Mux8Bus(m, readAddr2, q[0], q[1], q[2], q[3], q[4], q[5], q[6], q[7])

	// --- Immediates / constants (pure wiring) ---
	immSE := make(Bus, 16) // sign-extended imm6
	for i := 0; i < 6; i++ {
		immSE[i] = instr[i]
	}
	for i := 6; i < 16; i++ {
		immSE[i] = instr[5]
	}
	immX2 := make(Bus, 16) // imm6 * 2
	immX2[0] = zero
	for i := 1; i < 16; i++ {
		immX2[i] = immSE[i-1]
	}
	offSE := make(Bus, 16) // sign-extended off12
	for i := 0; i < 12; i++ {
		offSE[i] = instr[i]
	}
	for i := 12; i < 16; i++ {
		offSE[i] = instr[11]
	}
	offX2 := make(Bus, 16) // off12 * 2
	offX2[0] = zero
	for i := 1; i < 16; i++ {
		offX2[i] = offSE[i-1]
	}
	lui := make(Bus, 16) // imm6 << 10
	for i := 0; i < 16; i++ {
		lui[i] = zero
	}
	for k := 0; k < 6; k++ {
		lui[10+k] = instr[k]
	}
	const2 := NewBusConst(16, 2)

	// --- ALU ---
	aluOp := make(Bus, 3)
	aluOp[0] = Mux2(m, isALU, isORI, f0)            // I-type bit0=ORI, else func
	aluOp[1] = Mux2(m, isALU, OR(m, isANDI, isORI), f1) // I-type bit1=ANDI|ORI
	aluOp[2] = Mux2(m, isALU, zero, f2)
	opB := make(Bus, 16)
	Mux2Bus(m, useImm, rd2, immSE, opB)
	aluRes, _, _ := ALU16(m, rd1, opB, aluOp)

	// MUL: func=0 selects the low half of the signed product, else the high half.
	mulLow, mulHigh := mul16(m, rd1, rd2)
	mulRes := make(Bus, 16)
	Mux2Bus(m, funcIs0, mulHigh, mulLow, mulRes)
	aluOrMul := make(Bus, 16)
	Mux2Bus(m, isMUL, aluRes, mulRes, aluOrMul)

	// --- Branch decision ---
	eq, lt := Comparator16Signed(m, rd1, rd2)
	branchTaken := OR(m,
		OR(m, AND(m, isBEQ, eq), AND(m, isBNE, NOT(m, eq))),
		OR(m, AND(m, isBLT, lt), AND(m, isBGE, NOT(m, lt))))

	// --- Next-PC ---
	pcPlus2, _ := Adder16(m, pc, const2, zero)
	branchTgt, _ := Adder16(m, pc, immX2, zero)
	jalTgt, _ := Adder16(m, pc, offX2, zero)
	t1 := make(Bus, 16)
	Mux2Bus(m, branchTaken, pcPlus2, branchTgt, t1) // branch ? target : pc+2
	t2 := make(Bus, 16)
	Mux2Bus(m, isJAL, t1, jalTgt, t2) // jal ? jalTgt : t1
	nextPC := make(Bus, 16)
	Mux2Bus(m, isJALR, t2, rd1, nextPC) // jalr ? rs1 : t2
	for i := 0; i < 16; i++ {
		m.DFFTo(nextPC[i], pc[i])
	}

	// --- Write-back data + address ---
	wb1 := make(Bus, 16)
	Mux2Bus(m, isLUI, aluOrMul, lui, wb1) // lui ? lui : alu/mul
	wb2 := make(Bus, 16)
	Mux2Bus(m, isLoad, wb1, memData, wb2) // load ? memData : wb1
	writeData := make(Bus, 16)
	Mux2Bus(m, isLink, wb2, pcPlus2, writeData) // jal/jalr ? pc+2 : wb2

	const7 := NewBusConst(3, 7)
	writeAddr := make(Bus, 3)
	Mux2Bus(m, isLink, rdField, const7, writeAddr) // jal/jalr -> R7

	// --- Register-file write port (FF update at clock edge) ---
	wSel := decode3to8(m, writeAddr, writeEnable)
	for i := 1; i < 8; i++ {
		d := make(Bus, 16)
		Mux2Bus(m, wSel[i], q[i], writeData, d) // selected -> writeData, else hold
		for b := 0; b < 16; b++ {
			m.DFFTo(d[b], q[i][b])
		}
	}

	cpu := &GateCPU{
		pc:       pc,
		q:        q,
		instr:    instr,
		memData:  memData,
		memAddr:  aluRes,
		memWData: rd2,
		memRead:  isLoad,
		memWrite: isStore,
		byteMode: byteMode,
		halt:     isHALT,
	}
	cpu.sim = NewSimulator(m)
	return cpu
}

// Memory16 is the memory port the CPU drives: word/byte access with the same
// MMIO semantics as package memory (which satisfies this interface). Keeping it
// an interface lets the gate core run against the full system memory map
// without this package depending on the functional memory model.
type Memory16 interface {
	Read16(addr uint16) uint16
	Write16(addr uint16, v uint16)
	Read8(addr uint16) byte
	Write8(addr uint16, v byte)
}

// Step executes one instruction against mem. Returns false once halted.
func (g *GateCPU) Step(mem Memory16) bool {
	g.instr.SetVal(int(mem.Read16(uint16(g.pc.GetVal()))))
	g.sim.Settle()

	if g.halt.Val {
		return false
	}

	if g.memRead.Val { // LW / LB
		addr := uint16(g.memAddr.GetVal())
		var d int
		if g.byteMode.Val {
			d = int(mem.Read8(addr))
		} else {
			d = int(mem.Read16(addr))
		}
		g.memData.SetVal(d)
		g.sim.Settle() // propagate the loaded value into write-back
	}

	if g.memWrite.Val { // SW / SB
		addr := uint16(g.memAddr.GetVal())
		wd := uint16(g.memWData.GetVal())
		if g.byteMode.Val {
			mem.Write8(addr, byte(wd))
		} else {
			mem.Write16(addr, wd)
		}
	}

	for _, ff := range g.sim.FFs { // clock edge: latch PC and registers
		ff.Tick()
	}
	return true
}

// Run executes up to maxCycles instructions or until halt.
func (g *GateCPU) Run(mem Memory16, maxCycles int) int {
	n := 0
	for ; n < maxCycles; n++ {
		if !g.Step(mem) {
			break
		}
	}
	return n
}

// PC returns the current program counter.
func (g *GateCPU) PC() int { return g.pc.GetVal() }

// Reg returns the value of register i (R0 is always 0).
func (g *GateCPU) Reg(i int) uint16 { return uint16(g.q[i].GetVal()) }

// Halted reports whether the last Step decoded a HALT.
func (g *GateCPU) Halted() bool { return g.halt.Val }
