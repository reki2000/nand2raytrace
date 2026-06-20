package cpu

import (
	"nand16/internal/memory"
	"nand16/internal/op"
	"testing"
)

func BenchmarkCPU_Step(b *testing.B) {
	// Tight loop: decrement counter until 0
	code := []uint16{
		op.EncodeI(op.OpADDI, 1, 0, 0),      // R1 = 0 (will be overwritten)
		op.EncodeI(op.OpADDI, 1, 1, -1),     // R1-- (loop body)
		op.EncodeB(op.OpBNE, 1, 0, -1),      // if R1!=0 goto -2
		op.EncodeR(op.OpSYSTEM, 0, 0, 0, 1), // HALT
	}
	mem := memory.NewMemory(65536)
	for i, instr := range code {
		mem.Data[i*2] = byte(instr)
		mem.Data[i*2+1] = byte(instr >> 8)
	}
	c := NewCPU(mem)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.PC = 2 // start at the loop
		c.Regs[1] = 1000
		c.Halt = false
		c.Run(10000)
	}
}
