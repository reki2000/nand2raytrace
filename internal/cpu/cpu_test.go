package cpu

import (
	"nand16/internal/memory"
	"nand16/internal/op"
	"testing"
)

func makeSystem(code []uint16) (*CPU, *memory.Memory) {
	mem := memory.NewMemory(65536)
	// Load code at address 0
	for i, instr := range code {
		mem.Data[i*2] = byte(instr)
		mem.Data[i*2+1] = byte(instr >> 8)
	}
	cpu := NewCPU(mem)
	return cpu, mem
}

func TestCPU_ADDI(t *testing.T) {
	code := []uint16{
		op.EncodeI(op.OpADDI, 1, 0, 5),      // R1 = R0 + 5 = 5
		op.EncodeI(op.OpADDI, 2, 1, 10),     // R2 = R1 + 10 = 15
		op.EncodeI(op.OpADDI, 3, 2, -1),     // R3 = R2 + (-1) = 14
		op.EncodeR(op.OpSYSTEM, 0, 0, 0, 1), // HALT
	}
	cpu, _ := makeSystem(code)
	cpu.Run(10)

	if cpu.Regs[1] != 5 {
		t.Errorf("R1=%d, want 5", cpu.Regs[1])
	}
	if cpu.Regs[2] != 15 {
		t.Errorf("R2=%d, want 15", cpu.Regs[2])
	}
	if cpu.Regs[3] != 14 {
		t.Errorf("R3=%d, want 14", cpu.Regs[3])
	}
}

func TestCPU_ALU(t *testing.T) {
	code := []uint16{
		op.EncodeI(op.OpADDI, 1, 0, 20),     // R1 = 20
		op.EncodeI(op.OpADDI, 2, 0, 7),      // R2 = 7
		op.EncodeR(op.OpALU, 3, 1, 2, 0),    // R3 = R1 + R2 = 27 (ADD)
		op.EncodeR(op.OpALU, 4, 1, 2, 1),    // R4 = R1 - R2 = 13 (SUB)
		op.EncodeR(op.OpALU, 5, 1, 2, 2),    // R5 = R1 & R2 = 4 (AND: 10100 & 00111)
		op.EncodeR(op.OpALU, 6, 1, 2, 3),    // R6 = R1 | R2 = 23 (OR: 10100 | 00111)
		op.EncodeR(op.OpSYSTEM, 0, 0, 0, 1), // HALT
	}
	cpu, _ := makeSystem(code)
	cpu.Run(20)

	checks := map[int]uint16{3: 27, 4: 13, 5: 4, 6: 23}
	for r, want := range checks {
		if cpu.Regs[r] != want {
			t.Errorf("R%d=%d, want %d", r, cpu.Regs[r], want)
		}
	}
}

func TestCPU_LoadStore(t *testing.T) {
	// Use LUI+ORI to build larger values
	code := []uint16{
		op.EncodeI(op.OpADDI, 1, 0, 30), // R1 = 30 (data addr, fits in imm6)
		op.EncodeI(op.OpADDI, 2, 0, 25), // R2 = 25 (value to store)
		op.EncodeI(op.OpSW, 1, 2, 0),    // mem[R1+0] = R2
		op.EncodeI(op.OpLW, 3, 1, 0),    // R3 = mem[R1+0]
		op.EncodeR(op.OpSYSTEM, 0, 0, 0, 1),
	}
	cpu, mem := makeSystem(code)
	cpu.Run(10)

	if cpu.Regs[3] != 25 {
		t.Errorf("R3=%d, want 25", cpu.Regs[3])
	}
	if mem.Data[30] != 25 {
		t.Errorf("mem[30]=%d, want 25", mem.Data[30])
	}
}

func TestCPU_Branch(t *testing.T) {
	// Sum 1..5 using loop
	code := []uint16{
		op.EncodeI(op.OpADDI, 1, 0, 5), // R1 = 5 (counter)
		op.EncodeI(op.OpADDI, 2, 0, 0), // R2 = 0 (sum)
		// loop (PC=4):
		op.EncodeR(op.OpALU, 2, 2, 1, 0), // R2 = R2 + R1
		op.EncodeI(op.OpADDI, 1, 1, -1),  // R1 = R1 - 1
		op.EncodeB(op.OpBNE, 1, 0, -2),   // if R1 != R0(0), goto PC-4 (loop)
		op.EncodeR(op.OpSYSTEM, 0, 0, 0, 1),
	}
	cpu, _ := makeSystem(code)
	cpu.Run(100)

	if cpu.Regs[2] != 15 {
		t.Errorf("sum 1..5: R2=%d, want 15", cpu.Regs[2])
	}
}

func TestCPU_JAL_JALR(t *testing.T) {
	code := []uint16{
		op.EncodeJ(op.OpJAL, 3),             // 0: JAL +6 -> jump to addr 6, R7=2
		op.EncodeR(op.OpSYSTEM, 0, 0, 0, 1), // 2: HALT (should be skipped)
		0,                                   // 4: padding
		op.EncodeI(op.OpADDI, 1, 0, 0x11),   // 6: R1 = 0x11 (target)
		op.EncodeR(op.OpSYSTEM, 0, 7, 0, 0), // 8: JALR R7 -> jump back to addr 2
	}
	cpu, _ := makeSystem(code)
	cpu.Run(10)

	if cpu.Regs[1] != 0x11 {
		t.Errorf("R1=%04X, want 0011", cpu.Regs[1])
	}
	if !cpu.Halt {
		t.Error("CPU should be halted")
	}
}

func TestCPU_MUL(t *testing.T) {
	code := []uint16{
		op.EncodeI(op.OpADDI, 1, 0, 25),  // R1 = 25
		op.EncodeI(op.OpADDI, 2, 0, 13),  // R2 = 13
		op.EncodeR(op.OpMUL, 3, 1, 2, 0), // R3 = low16(25*13) = 325
		op.EncodeR(op.OpSYSTEM, 0, 0, 0, 1),
	}
	cpu, _ := makeSystem(code)
	cpu.Run(10)

	if cpu.Regs[3] != 325 {
		t.Errorf("R3=%d, want 325", cpu.Regs[3])
	}
}

func TestCPU_LUI(t *testing.T) {
	code := []uint16{
		op.EncodeI(op.OpLUI, 1, 0, 0x3F), // R1 = 0x3F << 10 = 0xFC00
		op.EncodeI(op.OpORI, 1, 1, 0x05), // R1 = 0xFC00 | 0x05 = 0xFC05
		op.EncodeR(op.OpSYSTEM, 0, 0, 0, 1),
	}
	cpu, _ := makeSystem(code)
	cpu.Run(10)

	if cpu.Regs[1] != 0xFC05 {
		t.Errorf("R1=%04X, want FC05", cpu.Regs[1])
	}
}

func TestCPU_Fibonacci(t *testing.T) {
	// Compute fib(10) = 55
	code := []uint16{
		op.EncodeI(op.OpADDI, 1, 0, 0),  // R1 = 0 (a)
		op.EncodeI(op.OpADDI, 2, 0, 1),  // R2 = 1 (b)
		op.EncodeI(op.OpADDI, 3, 0, 10), // R3 = 10 (counter)
		// loop:
		op.EncodeR(op.OpALU, 4, 1, 2, 0), // R4 = R1 + R2 (next)
		op.EncodeR(op.OpALU, 1, 2, 0, 0), // R1 = R2 + R0 (a = b)
		op.EncodeR(op.OpALU, 2, 4, 0, 0), // R2 = R4 + R0 (b = next)
		op.EncodeI(op.OpADDI, 3, 3, -1),  // R3--
		op.EncodeB(op.OpBNE, 3, 0, -4),   // if R3 != 0 goto loop
		op.EncodeR(op.OpSYSTEM, 0, 0, 0, 1),
	}
	cpu, _ := makeSystem(code)
	cpu.Run(200)

	if cpu.Regs[1] != 55 {
		t.Errorf("fib(10): R1=%d, want 55", cpu.Regs[1])
	}
}
