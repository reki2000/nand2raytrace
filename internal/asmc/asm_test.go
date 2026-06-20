package asmc

import (
	"bytes"
	"fmt"
	"nand16/internal/cpu"
	"nand16/internal/memory"
	"nand16/internal/system"
	"testing"
)

func TestAssembler_Basic(t *testing.T) {
	src := `
	addi r1, r0, 5    ; R1 = 5
	addi r2, r0, 10   ; R2 = 10
	add  r3, r1, r2   ; R3 = 15
	halt
`
	code, err := Assemble(src)
	if err != nil {
		t.Fatal(err)
	}

	mem := memory.NewMemory(65536)
	mem.LoadBinary(0, code)
	cpu := cpu.NewCPU(mem)
	cpu.PC = 0
	cpu.Run(10)

	if cpu.Regs[3] != 15 {
		t.Errorf("R3=%d, want 15", cpu.Regs[3])
	}
}

func TestAssembler_Labels(t *testing.T) {
	src := `
	addi r1, r0, 5    ; counter
	addi r2, r0, 0    ; sum
loop:
	add  r2, r2, r1
	addi r1, r1, -1
	bne  r1, r0, loop
	halt
`
	code, err := Assemble(src)
	if err != nil {
		t.Fatal(err)
	}

	mem := memory.NewMemory(65536)
	mem.LoadBinary(0, code)
	cpu := cpu.NewCPU(mem)
	cpu.PC = 0
	cpu.Run(200)

	if cpu.Regs[2] != 15 {
		t.Errorf("sum(1..5)=%d, want 15", cpu.Regs[2])
	}
}

func TestAssembler_HelloWorld(t *testing.T) {
	src := `
	; Print a newline by writing it to the UART data port (MMIO).
	; UART data lives at 0xF800; build the address via LUI.
	addi r2, r0, 10    ; newline char
	lui  r1, -2        ; r1 = (0x3E << 10) = 0xF800
	sb   r2, 0(r1)     ; mem[0xF800] = '\n' -> UART -> stdout
	halt
`
	code, err := Assemble(src)
	if err != nil {
		t.Fatal(err)
	}

	sys := system.NewSystem()
	sys.LoadRAM(0, code)
	sys.CPU.PC = 0
	var out bytes.Buffer
	sys.Stdout = &out
	sys.Run(20)

	if out.String() != "\n" {
		t.Errorf("output=%q, want newline", out.String())
	}
}

func TestSystem_UART(t *testing.T) {
	// Write a char via MMIO UART directly
	src := `
	; Write 'A' (65) to UART data port
	; 65 doesn't fit in imm6, so use LUI+ORI
	; Actually: 65 = 0b01000001
	; LUI loads imm6<<10: need upper bits
	; 65 is small, try: addi r1, r0, 31; addi r1, r1, 31; addi r1, r1, 3 = 65
	addi r1, r0, 31
	addi r1, r1, 31
	addi r1, r1, 3    ; R1 = 65 = 'A'
	
	; Build UART address 0xF800
	; 0xF800 = LUI (0x3E << 10 = 0xF800). 0x3E = 62, but imm6 is -32..31
	; 0xF800: upper 6 bits [15:10] = 0b111110 = 62 unsigned, but sign-ext = -2
	; LUI rd, imm6: rd = (imm6 & 0x3F) << 10
	lui  r2, -2       ; R2 = 0x3E << 10? No: (-2 & 0x3F) = 0x3E, 0x3E<<10 = 0xF800 ✓
	sw   r1, 0(r2)    ; mem[0xF800] = 65
	halt
`
	code, err := Assemble(src)
	if err != nil {
		t.Fatal(err)
	}

	sys := system.NewSystem()
	sys.LoadRAM(0, code)
	sys.CPU.PC = 0
	var out bytes.Buffer
	sys.Stdout = &out
	sys.Run(30)

	if out.String() != "A" {
		t.Errorf("UART output=%q, want 'A'", out.String())
	}
}

func TestAssembler_LoadImmediate(t *testing.T) {
	// li must construct arbitrary 16-bit values, exercising every expandLI path:
	// imm6, 2-ADDI, LUI shortcut, and the general shift-based sequence.
	cases := []struct {
		val  int
		want uint16
	}{
		{0, 0},
		{31, 31},
		{-32, 0xFFE0},    // imm6 negative
		{50, 50},         // 2-ADDI positive
		{-50, 0xFFCE},    // 2-ADDI negative
		{0xF000, 0xF000}, // LUI shortcut
		{0xEF00, 0xEF00}, // general path (forth SP)
		{0xDF00, 0xDF00}, // general path (forth RSP)
		{12345, 12345},   // general path
		{0xFFFF, 0xFFFF}, // all ones
		{256, 256},       // hi=1, lo=0
		{0x0101, 0x0101}, // hi=1, lo=1
	}
	for _, c := range cases {
		src := fmt.Sprintf("li r1, %d\nhalt\n", c.val)
		code, err := Assemble(src)
		if err != nil {
			t.Fatalf("li %d: assemble error: %v", c.val, err)
		}
			mem := memory.NewMemory(65536)
		mem.LoadBinary(0, code)
		cpu := cpu.NewCPU(mem)
		cpu.PC = 0
		cpu.Run(50)
		if cpu.Regs[1] != c.want {
			t.Errorf("li r1, %d -> R1=0x%04X (%d), want 0x%04X", c.val, cpu.Regs[1], cpu.Regs[1], c.want)
		}
	}
}

func TestAssembler_LoadImmediateLabelsConsistent(t *testing.T) {
	// A variable-length li before a label must not desync pass1/pass2 addresses.
	src := `
	li   r1, 12345    ; multi-instruction expansion
	jal  skip
	addi r1, r0, 7    ; must be skipped
skip:
	halt
`
	code, err := Assemble(src)
	if err != nil {
		t.Fatal(err)
	}
	mem := memory.NewMemory(65536)
	mem.LoadBinary(0, code)
	cpu := cpu.NewCPU(mem)
	cpu.PC = 0
	cpu.Run(50)
	if cpu.Regs[1] != 12345 {
		t.Errorf("label after li: R1=%d, want 12345 (jal target miscomputed?)", cpu.Regs[1])
	}
}

func TestAssembler_Subroutine(t *testing.T) {
	src := `
	addi r1, r0, 3
	addi r2, r0, 4
	jal  multiply     ; call multiply(R1, R2) -> R3
	halt
	
multiply:             ; R3 = R1 * R2 (software multiply using add)
	addi r3, r0, 0    ; result = 0
mul_loop:
	beq  r2, r0, mul_done
	add  r3, r3, r1   ; result += R1
	addi r2, r2, -1
	j    mul_loop
mul_done:
	ret
`
	code, err := Assemble(src)
	if err != nil {
		t.Fatal(err)
	}

	mem := memory.NewMemory(65536)
	mem.LoadBinary(0, code)
	cpu := cpu.NewCPU(mem)
	cpu.PC = 0
	cpu.Run(200)

	if cpu.Regs[3] != 12 {
		t.Errorf("3*4: R3=%d, want 12", cpu.Regs[3])
	}
}
