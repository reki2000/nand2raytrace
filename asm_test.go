package nand16

import (
	"bytes"
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

	mi := NewMemoryInterface()
	mem := NewMemory(65536, mi)
	mem.LoadBinary(0, code)
	cpu := NewCPU(mem)
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

	mi := NewMemoryInterface()
	mem := NewMemory(65536, mi)
	mem.LoadBinary(0, code)
	cpu := NewCPU(mem)
	cpu.PC = 0
	cpu.Run(200)

	if cpu.Regs[2] != 15 {
		t.Errorf("sum(1..5)=%d, want 15", cpu.Regs[2])
	}
}

func TestAssembler_HelloWorld(t *testing.T) {
	src := `
	; Print "Hi" via SYSCALL putchar
	addi r1, r0, 1     ; syscall 1 = putchar
	addi r2, r0, 72    ; 'H' = 72
	syscall
	addi r2, r0, 105   ; 'i' = 105... but 105 > 31! 
`
	// 105 doesn't fit in imm6. Need to build it differently.
	// Let's use a simpler test: print chars that fit in imm6

	src = `
	; Print "!" via SYSCALL putchar
	addi r1, r0, 1     ; syscall 1 = putchar
	addi r2, r0, 10    ; newline
	syscall
	halt
`
	code, err := Assemble(src)
	if err != nil {
		t.Fatal(err)
	}

	sys := NewSystem()
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

	sys := NewSystem()
	sys.LoadRAM(0, code)
	sys.CPU.PC = 0
	var out bytes.Buffer
	sys.Stdout = &out
	sys.Run(30)

	if out.String() != "A" {
		t.Errorf("UART output=%q, want 'A'", out.String())
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

	mi := NewMemoryInterface()
	mem := NewMemory(65536, mi)
	mem.LoadBinary(0, code)
	cpu := NewCPU(mem)
	cpu.PC = 0
	cpu.Run(200)

	if cpu.Regs[3] != 12 {
		t.Errorf("3*4: R3=%d, want 12", cpu.Regs[3])
	}
}
