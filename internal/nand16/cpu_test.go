package nand16

import (
	"nand16/internal/memory"
	"nand16/internal/op"
	"testing"
)

// loadGate assembles a program into a fresh 64KB RAM and a gate CPU.
func loadGate(code []uint16) (*memory.Memory, *GateCPU) {
	mem := memory.NewMemory(65536)
	for i, instr := range code {
		mem.Data[i*2] = byte(instr)
		mem.Data[i*2+1] = byte(instr >> 8)
	}
	return mem, NewGateCPU()
}

func TestGateCPU_ADDI(t *testing.T) {
	mem, g := loadGate([]uint16{
		op.EncodeI(op.OpADDI, 1, 0, 5),  // R1 = 5
		op.EncodeI(op.OpADDI, 2, 1, 10), // R2 = 15
		op.EncodeI(op.OpADDI, 3, 2, -1), // R3 = 14
		op.EncodeR(op.OpSYSTEM, 0, 0, 0, 1),
	})
	g.Run(mem, 10)
	for r, want := range map[int]uint16{1: 5, 2: 15, 3: 14} {
		if got := g.Reg(r); got != want {
			t.Errorf("R%d=%d, want %d", r, got, want)
		}
	}
}

func TestGateCPU_ALU(t *testing.T) {
	mem, g := loadGate([]uint16{
		op.EncodeI(op.OpADDI, 1, 0, 20),
		op.EncodeI(op.OpADDI, 2, 0, 7),
		op.EncodeR(op.OpALU, 3, 1, 2, 0), // ADD -> 27
		op.EncodeR(op.OpALU, 4, 1, 2, 1), // SUB -> 13
		op.EncodeR(op.OpALU, 5, 1, 2, 2), // AND -> 4
		op.EncodeR(op.OpALU, 6, 1, 2, 3), // OR  -> 23
		op.EncodeR(op.OpSYSTEM, 0, 0, 0, 1),
	})
	g.Run(mem, 20)
	for r, want := range map[int]uint16{3: 27, 4: 13, 5: 4, 6: 23} {
		if got := g.Reg(r); got != want {
			t.Errorf("R%d=%d, want %d", r, got, want)
		}
	}
}

func TestGateCPU_LoadStore(t *testing.T) {
	mem, g := loadGate([]uint16{
		op.EncodeI(op.OpADDI, 1, 0, 30), // R1 = 30 (addr)
		op.EncodeI(op.OpADDI, 2, 0, 25), // R2 = 25 (value)
		op.EncodeI(op.OpSW, 1, 2, 0),    // mem[R1] = R2
		op.EncodeI(op.OpLW, 3, 1, 0),    // R3 = mem[R1]
		op.EncodeR(op.OpSYSTEM, 0, 0, 0, 1),
	})
	g.Run(mem, 10)
	if g.Reg(3) != 25 {
		t.Errorf("R3=%d, want 25", g.Reg(3))
	}
	if mem.Data[30] != 25 {
		t.Errorf("mem.Data[30]=%d, want 25", mem.Data[30])
	}
}

func TestGateCPU_ByteLoadStore(t *testing.T) {
	mem, g := loadGate([]uint16{
		op.EncodeI(op.OpADDI, 1, 0, 28),   // addr (fits in imm6: -32..31)
		op.EncodeI(op.OpADDI, 2, 0, -1),   // 0xFFFF
		op.EncodeI(op.OpSB, 1, 2, 0),      // mem[28] = 0xFF (low byte)
		op.EncodeI(op.OpLB, 3, 1, 0),      // R3 = zero-extended byte = 0xFF
		op.EncodeR(op.OpSYSTEM, 0, 0, 0, 1),
	})
	g.Run(mem, 10)
	if g.Reg(3) != 0xFF {
		t.Errorf("R3=0x%X, want 0xFF", g.Reg(3))
	}
	if mem.Data[28] != 0xFF {
		t.Errorf("mem.Data[28]=0x%X, want 0xFF", mem.Data[28])
	}
}

func TestGateCPU_Branch(t *testing.T) {
	mem, g := loadGate([]uint16{
		op.EncodeI(op.OpADDI, 1, 0, 5), // R1 = 5
		op.EncodeI(op.OpADDI, 2, 0, 0), // R2 = 0
		op.EncodeR(op.OpALU, 2, 2, 1, 0), // R2 += R1
		op.EncodeI(op.OpADDI, 1, 1, -1),  // R1--
		op.EncodeB(op.OpBNE, 1, 0, -2),   // loop while R1 != 0
		op.EncodeR(op.OpSYSTEM, 0, 0, 0, 1),
	})
	g.Run(mem, 100)
	if g.Reg(2) != 15 {
		t.Errorf("sum 1..5: R2=%d, want 15", g.Reg(2))
	}
}

func TestGateCPU_JAL_JALR(t *testing.T) {
	mem, g := loadGate([]uint16{
		op.EncodeJ(op.OpJAL, 3),             // 0: jump to addr 6, R7=2
		op.EncodeR(op.OpSYSTEM, 0, 0, 0, 1), // 2: HALT (skipped)
		0,                                   // 4: padding
		op.EncodeI(op.OpADDI, 1, 0, 0x11),   // 6: R1 = 0x11
		op.EncodeR(op.OpSYSTEM, 0, 7, 0, 0), // 8: JALR R7 -> addr 2
	})
	g.Run(mem, 10)
	if g.Reg(1) != 0x11 {
		t.Errorf("R1=0x%X, want 0x11", g.Reg(1))
	}
	if !g.Halted() {
		t.Error("expected halt")
	}
}

func TestGateCPU_LUI(t *testing.T) {
	mem, g := loadGate([]uint16{
		op.EncodeI(op.OpLUI, 1, 0, 0x3F), // R1 = 0x3F << 10 = 0xFC00
		op.EncodeI(op.OpORI, 1, 1, 0x05), // R1 |= 5 = 0xFC05
		op.EncodeR(op.OpSYSTEM, 0, 0, 0, 1),
	})
	g.Run(mem, 10)
	if g.Reg(1) != 0xFC05 {
		t.Errorf("R1=0x%04X, want 0xFC05", g.Reg(1))
	}
}

func TestGateCPU_MUL(t *testing.T) {
	mem, g := loadGate([]uint16{
		op.EncodeI(op.OpADDI, 1, 0, 25),  // R1 = 25
		op.EncodeI(op.OpADDI, 2, 0, 13),  // R2 = 13
		op.EncodeR(op.OpMUL, 3, 1, 2, 0), // R3 = low16(25*13) = 325
		op.EncodeR(op.OpSYSTEM, 0, 0, 0, 1),
	})
	g.Run(mem, 10)
	if g.Reg(3) != 325 {
		t.Errorf("R3=%d, want 325", g.Reg(3))
	}
}

func TestGateCPU_MULSigned(t *testing.T) {
	// (-3) * 7 = -21. low16 = 0xFFEB, high16 = 0xFFFF.
	mem, g := loadGate([]uint16{
		op.EncodeI(op.OpADDI, 1, 0, -3),
		op.EncodeI(op.OpADDI, 2, 0, 7),
		op.EncodeR(op.OpMUL, 3, 1, 2, 0), // low
		op.EncodeR(op.OpMUL, 4, 1, 2, 1), // high
		op.EncodeR(op.OpSYSTEM, 0, 0, 0, 1),
	})
	g.Run(mem, 10)
	if g.Reg(3) != 0xFFEB {
		t.Errorf("low=0x%04X, want 0xFFEB", g.Reg(3))
	}
	if g.Reg(4) != 0xFFFF {
		t.Errorf("high=0x%04X, want 0xFFFF", g.Reg(4))
	}
}

func TestGateCPU_Fibonacci(t *testing.T) {
	mem, g := loadGate([]uint16{
		op.EncodeI(op.OpADDI, 1, 0, 0),  // a = 0
		op.EncodeI(op.OpADDI, 2, 0, 1),  // b = 1
		op.EncodeI(op.OpADDI, 3, 0, 10), // counter
		op.EncodeR(op.OpALU, 4, 1, 2, 0), // next = a+b
		op.EncodeR(op.OpALU, 1, 2, 0, 0), // a = b
		op.EncodeR(op.OpALU, 2, 4, 0, 0), // b = next
		op.EncodeI(op.OpADDI, 3, 3, -1),  // counter--
		op.EncodeB(op.OpBNE, 3, 0, -4),   // loop
		op.EncodeR(op.OpSYSTEM, 0, 0, 0, 1),
	})
	g.Run(mem, 200)
	if g.Reg(1) != 55 {
		t.Errorf("fib(10): R1=%d, want 55", g.Reg(1))
	}
}
