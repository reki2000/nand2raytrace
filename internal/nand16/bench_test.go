package nand16

import "testing"

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
