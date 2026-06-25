// Package integration cross-checks the two NAND-16 machine models — the
// behavioral CPU (internal/cpu) and the gate-level CPU (internal/nand16) — and
// verifies that the gate CPU drives the real system memory map and MMIO devices.
package integration

import (
	"bytes"
	"testing"

	"nand16/internal/cpu"
	"nand16/internal/memory"
	"nand16/internal/nand16"
	"nand16/internal/op"
	"nand16/internal/system"
)

func loadMem(code []uint16) *memory.Memory {
	mem := memory.NewMemory(65536)
	for i, instr := range code {
		mem.Data[i*2] = byte(instr)
		mem.Data[i*2+1] = byte(instr >> 8)
	}
	return mem
}

var programs = []struct {
	name   string
	code   []uint16
	cycles int
}{
	{"addi", []uint16{
		op.EncodeI(op.OpADDI, 1, 0, 5),
		op.EncodeI(op.OpADDI, 2, 1, 10),
		op.EncodeI(op.OpADDI, 3, 2, -1),
		op.EncodeR(op.OpSYSTEM, 0, 0, 0, 1),
	}, 10},
	{"alu", []uint16{
		op.EncodeI(op.OpADDI, 1, 0, 20),
		op.EncodeI(op.OpADDI, 2, 0, 7),
		op.EncodeR(op.OpALU, 3, 1, 2, 0),
		op.EncodeR(op.OpALU, 4, 1, 2, 1),
		op.EncodeR(op.OpALU, 5, 1, 2, 4), // XOR
		op.EncodeR(op.OpALU, 6, 1, 2, 5), // SHL
		op.EncodeR(op.OpSYSTEM, 0, 0, 0, 1),
	}, 20},
	{"loadstore", []uint16{
		op.EncodeI(op.OpADDI, 1, 0, 30),
		op.EncodeI(op.OpADDI, 2, 0, 25),
		op.EncodeI(op.OpSW, 1, 2, 0),
		op.EncodeI(op.OpLW, 3, 1, 0),
		op.EncodeR(op.OpSYSTEM, 0, 0, 0, 1),
	}, 10},
	{"branch", []uint16{
		op.EncodeI(op.OpADDI, 1, 0, 5),
		op.EncodeI(op.OpADDI, 2, 0, 0),
		op.EncodeR(op.OpALU, 2, 2, 1, 0),
		op.EncodeI(op.OpADDI, 1, 1, -1),
		op.EncodeB(op.OpBNE, 1, 0, -2),
		op.EncodeR(op.OpSYSTEM, 0, 0, 0, 1),
	}, 100},
	{"jal_jalr", []uint16{
		op.EncodeJ(op.OpJAL, 3),
		op.EncodeR(op.OpSYSTEM, 0, 0, 0, 1),
		0,
		op.EncodeI(op.OpADDI, 1, 0, 0x11),
		op.EncodeR(op.OpSYSTEM, 0, 7, 0, 0),
	}, 10},
	{"lui", []uint16{
		op.EncodeI(op.OpLUI, 1, 0, 0x3F),
		op.EncodeI(op.OpORI, 1, 1, 0x05),
		op.EncodeR(op.OpSYSTEM, 0, 0, 0, 1),
	}, 10},
	{"mul", []uint16{
		op.EncodeI(op.OpADDI, 1, 0, -3),
		op.EncodeI(op.OpADDI, 2, 0, 7),
		op.EncodeR(op.OpMUL, 3, 1, 2, 0), // low
		op.EncodeR(op.OpMUL, 4, 1, 2, 1), // high
		op.EncodeR(op.OpSYSTEM, 0, 0, 0, 1),
	}, 10},
	{"fibonacci", []uint16{
		op.EncodeI(op.OpADDI, 1, 0, 0),
		op.EncodeI(op.OpADDI, 2, 0, 1),
		op.EncodeI(op.OpADDI, 3, 0, 10),
		op.EncodeR(op.OpALU, 4, 1, 2, 0),
		op.EncodeR(op.OpALU, 1, 2, 0, 0),
		op.EncodeR(op.OpALU, 2, 4, 0, 0),
		op.EncodeI(op.OpADDI, 3, 3, -1),
		op.EncodeB(op.OpBNE, 3, 0, -4),
		op.EncodeR(op.OpSYSTEM, 0, 0, 0, 1),
	}, 200},
}

// TestBehavioralGateEquivalence runs every program on both CPUs and requires
// identical final register state.
func TestBehavioralGateEquivalence(t *testing.T) {
	for _, p := range programs {
		t.Run(p.name, func(t *testing.T) {
			bc := cpu.NewCPU(loadMem(p.code))
			bc.PC = 0
			bc.Run(p.cycles)

			g := nand16.NewGateCPU()
			g.Run(loadMem(p.code), p.cycles)

			for r := 1; r < 8; r++ {
				if bc.Regs[r] != g.Reg(r) {
					t.Errorf("R%d: behavioral=0x%04X gate=0x%04X", r, bc.Regs[r], g.Reg(r))
				}
			}
		})
	}
}

// TestGateCPUDrivesSystem runs the gate CPU against the real system memory map
// so that a store to the UART data port reaches stdout through the MMIO device.
func TestGateCPUDrivesSystem(t *testing.T) {
	sys := system.NewSystem()
	var out bytes.Buffer
	sys.Stdout = &out

	// Write a newline to the UART data port (0xF800) then halt.
	code := []uint16{
		op.EncodeI(op.OpADDI, 2, 0, 10), // R2 = '\n'
		op.EncodeI(op.OpLUI, 1, 0, 0x3E), // R1 = 0x3E<<10 = 0xF800
		op.EncodeI(op.OpSB, 1, 2, 0),    // mem[0xF800] = R2 -> UART
		op.EncodeR(op.OpSYSTEM, 0, 0, 0, 1),
	}
	for i, instr := range code {
		sys.Mem.Data[i*2] = byte(instr)
		sys.Mem.Data[i*2+1] = byte(instr >> 8)
	}

	g := nand16.NewGateCPU() // PC starts at 0
	g.Run(sys.Mem, 20)

	if out.String() != "\n" {
		t.Errorf("gate CPU UART output = %q, want newline", out.String())
	}
}

// TestFlatGateEquivalence runs every program on the flat gate-level CPU
// and requires identical final register state to the behavioral CPU.
func TestFlatGateEquivalence(t *testing.T) {
	for _, p := range programs {
		t.Run(p.name, func(t *testing.T) {
			bc := cpu.NewCPU(loadMem(p.code))
			bc.PC = 0
			bc.Run(p.cycles)

			g := nand16.NewGateCPUParallel(1)
			g.Run(loadMem(p.code), p.cycles)

			for r := 1; r < 8; r++ {
				if bc.Regs[r] != g.Reg(r) {
					t.Errorf("R%d: behavioral=0x%04X flat=0x%04X", r, bc.Regs[r], g.Reg(r))
				}
			}
		})
	}
}
