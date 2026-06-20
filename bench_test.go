package nand16

import "testing"

func BenchmarkCPU_Step(b *testing.B) {
	// Tight loop: decrement counter until 0
	code := []uint16{
		EncodeI(OpADDI, 1, 0, 0),     // R1 = 0 (will be overwritten)
		EncodeI(OpADDI, 1, 1, -1),    // R1-- (loop body)
		EncodeB(OpBNE, 1, 0, -1),     // if R1!=0 goto -2
		EncodeR(OpSYSTEM, 0, 0, 0, 1), // HALT
	}
	mi := NewMemoryInterface()
	mem := NewMemory(65536, mi)
	for i, instr := range code {
		mem.Data[i*2] = byte(instr)
		mem.Data[i*2+1] = byte(instr >> 8)
	}
	cpu := NewCPU(mem)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cpu.PC = 2 // start at the loop
		cpu.Regs[1] = 1000
		cpu.Halt = false
		cpu.Run(10000)
	}
}

func BenchmarkGateLevel_ALU(b *testing.B) {
	m := NewModule("bench")
	a, bx := NewBus(16), NewBus(16)
	op := NewBus(3)
	result, _, _ := ALU16(m, a, bx, op)
	sim := NewSimulator(m)

	b.ResetTimer()
	b.ReportMetric(float64(len(sim.Gates)), "gates")
	for i := 0; i < b.N; i++ {
		a.SetVal(i)
		bx.SetVal(i + 1)
		op.SetVal(0) // ADD
		sim.Settle()
		_ = result.GetVal()
	}
}
