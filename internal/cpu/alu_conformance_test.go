package cpu

import (
	"nand16/internal/nand16"
	"testing"
)

// TestALUConformance guards against the behavioral ALU (aluExec) drifting from
// the gate-level ALU (nand16.ALU16). Both must implement the same op encoding
// (000=ADD,001=SUB,010=AND,011=OR,100=XOR,101=SHL,110=SHR,111=SRA) with
// identical results for every operand pair we sample.
func TestALUConformance(t *testing.T) {
	// Build the gate-level ALU once and drive it via the simulator.
	m := nand16.NewModule("alu")
	aBus, bBus := nand16.NewBus(16), nand16.NewBus(16)
	opBus := nand16.NewBus(3)
	result, _, _ := nand16.ALU16(m, aBus, bBus, opBus)
	sim := nand16.NewSimulator(m)

	gate := func(fn int, a, b uint16) uint16 {
		aBus.SetVal(int(a))
		bBus.SetVal(int(b))
		opBus.SetVal(fn)
		sim.Settle()
		return uint16(result.GetVal())
	}

	operands := []uint16{
		0x0000, 0x0001, 0x0002, 0x000F, 0x00FF,
		0x7FFF, 0x8000, 0xABCD, 0x1234, 0xF000, 0xFFFF,
	}
	opNames := []string{"ADD", "SUB", "AND", "OR", "XOR", "SHL", "SHR", "SRA"}

	var c CPU
	for fn := 0; fn < 8; fn++ {
		for _, a := range operands {
			for _, b := range operands {
				want := gate(fn, a, b)
				got := c.aluExec(fn, a, b)
				if got != want {
					t.Errorf("%s(0x%04X, 0x%04X): behavioral=0x%04X gate=0x%04X",
						opNames[fn], a, b, got, want)
				}
			}
		}
	}
}
