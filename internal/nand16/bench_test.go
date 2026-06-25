package nand16

import (
	"nand16/internal/memory"
	"nand16/internal/op"
	"testing"
)

func encI(opcode, rd, rs1, imm6 int) uint16 {
	return uint16((opcode << 12) | (rd << 9) | (rs1 << 6) | (imm6 & 0x3F))
}
func encR(opcode, rd, rs1, rs2, fn int) uint16 {
	return uint16((opcode << 12) | (rd << 9) | (rs1 << 6) | (rs2 << 3) | fn)
}

// --- Correctness: FlatSimulator vs Simulator produce identical results ---

func TestFlatSimulator_ALU_Conformance(t *testing.T) {
	m1 := NewModule("alu-orig")
	a1, b1, op1 := NewBus(16), NewBus(16), NewBus(3)
	r1, _, _ := ALU16(m1, a1, b1, op1)
	sim1 := NewSimulator(m1)

	m2 := NewModule("alu-flat")
	a2, b2, op2 := NewBus(16), NewBus(16), NewBus(3)
	r2, _, _ := ALU16(m2, a2, b2, op2)
	fs := NewFlatSimulator(NewSimulator(m2), 1)

	cases := [][3]int{
		{0x1234, 0x5678, 0}, {0x5678, 0x1234, 1},
		{0xAAAA, 0x5555, 2}, {0xAAAA, 0x5555, 3},
		{0xAAAA, 0x5555, 4}, {0x0001, 0x0004, 5},
		{0x8000, 0x0004, 6}, {0x8000, 0x0004, 7},
		{0xFFFF, 0x0001, 0}, {0x0000, 0x0001, 1},
	}
	rIdx := fs.BusIdx(r2)
	for _, tc := range cases {
		a1.SetVal(tc[0]); b1.SetVal(tc[1]); op1.SetVal(tc[2])
		sim1.Settle()
		expected := r1.GetVal()

		a2.SetVal(tc[0]); b2.SetVal(tc[1]); op2.SetVal(tc[2])
		for _, w := range a2 { fs.SetWireVal(fs.WireIdx(w), w.Val) }
		for _, w := range b2 { fs.SetWireVal(fs.WireIdx(w), w.Val) }
		for _, w := range op2 { fs.SetWireVal(fs.WireIdx(w), w.Val) }
		fs.Settle()
		got := fs.GetBusVal(rIdx)

		if got != expected {
			t.Errorf("op=%d a=0x%04X b=0x%04X: Sim=%04X Flat=%04X",
				tc[2], tc[0], tc[1], expected, got)
		}
	}
}

func TestFlatSimulator_CPU_Conformance(t *testing.T) {
	mem1 := memory.NewMemory(0x10000)
	mem2 := memory.NewMemory(0x10000)

	// ADDI R1,R0,5 ; ADDI R2,R0,3 ; ADD R3,R1,R2 ; HALT
	prog := []uint16{
		encI(op.OpADDI, 1, 0, 5),       // R1 = 5
		encI(op.OpADDI, 2, 0, 3),       // R2 = 3
		encR(op.OpALU, 3, 1, 2, 0),     // R3 = R1+R2 = 8
		encR(op.OpSYSTEM, 0, 0, 0, 1),  // HALT
	}
	for i, w := range prog {
		mem1.Write16(uint16(i*2), w)
		mem2.Write16(uint16(i*2), w)
	}

	cpu1 := NewGateCPU()
	cpu2 := NewGateCPUParallel(1)

	for step := 0; step < 4; step++ {
		ok1 := cpu1.Step(mem1)
		ok2 := cpu2.Step(mem2)
		if ok1 != ok2 {
			t.Fatalf("step %d: halted mismatch orig=%v flat=%v", step, !ok1, !ok2)
		}
		if cpu1.PC() != cpu2.PC() {
			t.Errorf("step %d: PC orig=0x%04X flat=0x%04X", step, cpu1.PC(), cpu2.PC())
		}
		for r := 0; r < 8; r++ {
			if cpu1.Reg(r) != cpu2.Reg(r) {
				t.Errorf("step %d: R%d orig=0x%04X flat=0x%04X",
					step, r, cpu1.Reg(r), cpu2.Reg(r))
			}
		}
	}
	if cpu2.Reg(1) != 5 { t.Errorf("R1 expected 5 got %d", cpu2.Reg(1)) }
	if cpu2.Reg(2) != 3 { t.Errorf("R2 expected 3 got %d", cpu2.Reg(2)) }
	if cpu2.Reg(3) != 8 { t.Errorf("R3 expected 8 got %d", cpu2.Reg(3)) }
}

func TestFlatSimulator_CPU_LoadStore(t *testing.T) {
	mem1 := memory.NewMemory(0x10000)
	mem2 := memory.NewMemory(0x10000)

	prog := []uint16{
		encI(op.OpADDI, 1, 0, 25),      // R1 = 42
		encI(op.OpADDI, 2, 0, 0x10),    // R2 = 0x10 (base addr)
		encI(op.OpSW, 1, 0, 0),         // mem[R1+0] = ... wait, SW encoding is different
	}
	// SW is: SW rs1, imm6(rd) -> mem[rd + imm6] = rs1
	// encoding: op=7, rd(base)=bits[11:9], rs1(src)=bits[8:6], imm6=bits[5:0]
	// SW R1, 0(R2): store R1 at [R2+0]
	//   op=7, rd=R2(010), rs1=R1(001), imm=0
	prog = []uint16{
		encI(op.OpADDI, 1, 0, 25),      // R1 = 42
		encI(op.OpADDI, 2, 0, 0x10),    // R2 = 16
		encI(op.OpSW, 2, 1, 0),         // mem[R2+0] = R1
		encI(op.OpLW, 3, 2, 0),         // R3 = mem[R2+0]
		encR(op.OpSYSTEM, 0, 0, 0, 1),  // HALT
	}
	for i, w := range prog {
		mem1.Write16(uint16(i*2), w)
		mem2.Write16(uint16(i*2), w)
	}

	cpu1 := NewGateCPU()
	cpu2 := NewGateCPUParallel(1)

	for step := 0; step < 5; step++ {
		ok1 := cpu1.Step(mem1)
		ok2 := cpu2.Step(mem2)
		if ok1 != ok2 {
			t.Fatalf("step %d: halted mismatch", step)
		}
		for r := 0; r < 8; r++ {
			if cpu1.Reg(r) != cpu2.Reg(r) {
				t.Errorf("step %d: R%d orig=0x%04X flat=0x%04X",
					step, r, cpu1.Reg(r), cpu2.Reg(r))
			}
		}
	}
	if cpu2.Reg(3) != 25 { t.Errorf("R3 expected 25 got %d", cpu2.Reg(3)) }
}

// --- Benchmarks ---

func BenchmarkGateLevel_ALU(b *testing.B) {
	m := NewModule("bench")
	a, bx := NewBus(16), NewBus(16)
	opb := NewBus(3)
	result, _, _ := ALU16(m, a, bx, opb)
	sim := NewSimulator(m)
	b.ResetTimer()
	b.ReportMetric(float64(len(sim.Gates)), "gates")
	for i := 0; i < b.N; i++ {
		a.SetVal(i); bx.SetVal(i + 1); opb.SetVal(0)
		sim.Settle()
		_ = result.GetVal()
	}
}

func BenchmarkGateLevel_CPUStep(b *testing.B) {
	cpu := NewGateCPU()
	mem := memory.NewMemory(0x10000)
	mem.Write16(0, 0xF007) // NOP
	b.ResetTimer()
	b.ReportMetric(float64(len(cpu.sim.(*Simulator).Gates)), "gates")
	for i := 0; i < b.N; i++ {
		cpu.Step(mem)
	}
}

func BenchmarkFlatSim_CPUStep(b *testing.B) {
	cpu := NewGateCPUParallel(1)
	mem := memory.NewMemory(0x10000)
	mem.Write16(0, 0xF007)
	b.ResetTimer()
	b.ReportMetric(float64(cpu.flat.nGate), "gates")
	for i := 0; i < b.N; i++ {
		cpu.Step(mem)
	}
}

func BenchmarkSettle_Original(b *testing.B) {
	cpu := NewGateCPU()
	sim := cpu.sim.(*Simulator)
	b.ResetTimer()
	b.ReportMetric(float64(len(sim.Gates)), "gates")
	for i := 0; i < b.N; i++ { sim.Settle() }
}

func BenchmarkSettle_Flat(b *testing.B) {
	cpu := NewGateCPU()
	fs := NewFlatSimulator(cpu.sim.(*Simulator), 1)
	b.ResetTimer()
	b.ReportMetric(float64(fs.nGate), "gates")
	for i := 0; i < b.N; i++ { fs.Settle() }
}
