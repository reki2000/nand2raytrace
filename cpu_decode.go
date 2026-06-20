package nand16

// Instruction format constants
const (
	// Opcodes (4 bits)
	OpALU    = 0x0 // R-type: ALU rd, rs1, rs2
	OpMUL    = 0x1 // R-type: MUL rs1, rs2 -> {R1:R0}
	OpADDI   = 0x2 // I-type: ADDI rd, rs1, imm6
	OpANDI   = 0x3 // I-type: ANDI rd, rs1, imm6
	OpORI    = 0x4 // I-type: ORI rd, rs1, imm6
	OpLUI    = 0x5 // I-type: LUI rd, imm6 (loads imm6 << 10 into upper bits)
	OpLW     = 0x6 // I-type: LW rd, imm6(rs1)
	OpSW     = 0x7 // I-type: SW rs1, imm6(rd) -- note: rd field is base, rs1 is src
	OpLB     = 0x8 // I-type: LB rd, imm6(rs1)
	OpSB     = 0x9 // I-type: SB rs1, imm6(rd)
	OpBEQ    = 0xA // B-type: BEQ rs1, rs2, off6
	OpBNE    = 0xB // B-type: BNE rs1, rs2, off6
	OpBLT    = 0xC // B-type: BLT rs1, rs2, off6
	OpBGE    = 0xD // B-type: BGE rs1, rs2, off6
	OpJAL    = 0xE // J-type: JAL off12
	OpSYSTEM = 0xF // R-type: JALR(func=0), HALT(func=1), SYSCALL(func=2), NOP(func=7)
)

// Decoded instruction fields
type Decoded struct {
	Op   int // 4-bit opcode [15:12]
	Rd   int // 3-bit [11:9]
	Rs1  int // 3-bit [8:6]
	Rs2  int // 3-bit [5:3]
	Func int // 3-bit [2:0]
	Imm6 int // 6-bit [5:0] sign-extended to 16
	Off12 int // 12-bit [11:0] sign-extended to 16
	Raw  uint16
}

// Decode extracts fields from a 16-bit instruction word.
func Decode(instr uint16) Decoded {
	d := Decoded{Raw: instr}
	d.Op = int(instr >> 12)
	d.Rd = int((instr >> 9) & 7)
	d.Rs1 = int((instr >> 6) & 7)
	d.Rs2 = int((instr >> 3) & 7)
	d.Func = int(instr & 7)

	// Sign-extend imm6
	imm6 := int(instr & 0x3F)
	if imm6 >= 32 {
		imm6 -= 64
	}
	d.Imm6 = imm6

	// Sign-extend off12
	off12 := int(instr & 0xFFF)
	if off12 >= 2048 {
		off12 -= 4096
	}
	d.Off12 = off12

	return d
}

// Encode helpers for assembler
func EncodeR(op, rd, rs1, rs2, fn int) uint16 {
	return uint16(op<<12 | rd<<9 | rs1<<6 | rs2<<3 | fn)
}

func EncodeI(op, rd, rs1, imm6 int) uint16 {
	return uint16(op<<12 | rd<<9 | rs1<<6 | (imm6 & 0x3F))
}

func EncodeB(op, rs1, rs2, off6 int) uint16 {
	return uint16(op<<12 | rs1<<9 | rs2<<6 | (off6 & 0x3F))
}

func EncodeJ(op, off12 int) uint16 {
	return uint16(op<<12 | (off12 & 0xFFF))
}
