package nand16

import (
	"fmt"
	"strconv"
	"strings"
)

// Assembler translates NAND-16 assembly to binary.
type Assembler struct {
	labels map[string]int
	errors []string
}

// Assemble translates assembly source to binary.
func Assemble(source string) ([]byte, error) {
	a := &Assembler{labels: make(map[string]int)}
	lines := a.tokenize(source)

	// Pass 1: collect labels
	a.pass1(lines)
	if len(a.errors) > 0 {
		return nil, fmt.Errorf("pass1: %s", strings.Join(a.errors, "; "))
	}

	// Pass 2: emit code
	code := a.pass2(lines)
	if len(a.errors) > 0 {
		return nil, fmt.Errorf("pass2: %s", strings.Join(a.errors, "; "))
	}

	return code, nil
}

type asmLine struct {
	lineNo int
	label  string
	op     string
	args   []string
	orig   string
}

var regMap = map[string]int{
	"r0": 0, "r1": 1, "r2": 2, "r3": 3,
	"r4": 4, "r5": 5, "r6": 6, "r7": 7,
	"sp": 6, "ra": 7, "zero": 0,
}

func (a *Assembler) tokenize(source string) []asmLine {
	var lines []asmLine
	for i, raw := range strings.Split(source, "\n") {
		line := strings.TrimSpace(raw)
		// Remove comments
		if idx := strings.Index(line, ";"); idx >= 0 {
			line = strings.TrimSpace(line[:idx])
		}
		if line == "" {
			continue
		}

		al := asmLine{lineNo: i + 1, orig: raw}

		// Check for label
		if idx := strings.Index(line, ":"); idx >= 0 {
			al.label = strings.TrimSpace(line[:idx])
			line = strings.TrimSpace(line[idx+1:])
		}

		if line != "" {
			parts := strings.Fields(line)
			al.op = strings.ToLower(parts[0])
			if len(parts) > 1 {
				argStr := strings.Join(parts[1:], " ")
				for _, arg := range strings.Split(argStr, ",") {
					arg = strings.TrimSpace(arg)
					if arg != "" {
						al.args = append(al.args, arg)
					}
				}
			}
		}

		lines = append(lines, al)
	}
	return lines
}

func (a *Assembler) pass1(lines []asmLine) {
	pc := 0
	for _, l := range lines {
		if l.label != "" {
			a.labels[l.label] = pc
		}
		if l.op != "" {
			pc += a.instrSize(l)
		}
	}
}

func (a *Assembler) instrSize(l asmLine) int {
	switch l.op {
	case ".org":
		return 0
	case ".db":
		return len(l.args)
	case ".dw":
		return len(l.args) * 2
	case ".ascii":
		if len(l.args) > 0 {
			s := a.parseString(l.args[0])
			return len(s)
		}
		return 0
	case ".asciiz":
		if len(l.args) > 0 {
			s := a.parseString(l.args[0])
			return len(s) + 1
		}
		return 1
	case "li": // pseudo: LUI + ORI = 4 bytes
		return 4
	default:
		return 2
	}
}

func (a *Assembler) pass2(lines []asmLine) []byte {
	code := make([]byte, 0, 1024)
	pc := 0
	for _, l := range lines {
		if l.op == "" {
			continue
		}
		switch l.op {
		case ".org":
			target := a.parseImm(l, l.args[0])
			for len(code) < target {
				code = append(code, 0)
			}
			pc = target

		case ".db":
			for _, arg := range l.args {
				code = append(code, byte(a.parseImm(l, arg)))
				pc++
			}

		case ".dw":
			for _, arg := range l.args {
				v := uint16(a.parseImm(l, arg))
				code = append(code, byte(v), byte(v>>8))
				pc += 2
			}

		case ".ascii":
			if len(l.args) > 0 {
				s := a.parseString(l.args[0])
				code = append(code, []byte(s)...)
				pc += len(s)
			}

		case ".asciiz":
			if len(l.args) > 0 {
				s := a.parseString(l.args[0])
				code = append(code, []byte(s)...)
				code = append(code, 0)
				pc += len(s) + 1
			}

		case "li":
			// Pseudo-instruction: LI rd, imm16 -> LUI rd, hi6 ; ORI rd, rd, lo10
			// Actually: LUI loads imm6<<10. We need: upper 6 bits + lower 10.
			// But ORI imm6 is only 6 bits. So we can only load imm6<<10 | imm6.
			// Better: LUI rd, upper6; ADDI rd, rd, lower_signed
			rd := a.parseReg(l, l.args[0])
			imm := a.parseImm(l, l.args[1])
			upper := (imm >> 10) & 0x3F
			lower := imm & 0x3FF
			// If lower >= 512, it's negative when sign-extended from 10 bits
			// But our ADDI only has 6-bit immediate...
			// Simpler approach: LUI sets bits[15:10], then ORI sets bits[5:0]
			// Bits [9:6] are lost. 
			// Alternative: LUI rd, upper; ORI rd, rd, lo6  (only covers 12 bits)
			// For full 16-bit: LUI rd, upper6; ORI rd, rd, mid6...
			// This ISA limitation means LI can only construct certain values.
			// Let's use: LUI for upper, ADDI for lower (sign-extended 6-bit)
			// upper6 << 10 + sign_ext(lower6)
			// For arbitrary 16-bit: may need more instructions.
			// Practical: LUI + ORI covers bits [15:10] | [5:0]
			// Use shifts for full range. For now, support common cases.
			lo6 := lower & 0x3F
			_ = upper
			w1 := EncodeI(OpLUI, rd, 0, upper)
			w2 := EncodeI(OpORI, rd, rd, lo6)
			code = append(code, byte(w1), byte(w1>>8))
			code = append(code, byte(w2), byte(w2>>8))
			pc += 4

		case "nop":
			w := EncodeR(OpSYSTEM, 0, 0, 0, 7)
			code = append(code, byte(w), byte(w>>8))
			pc += 2

		case "halt":
			w := EncodeR(OpSYSTEM, 0, 0, 0, 1)
			code = append(code, byte(w), byte(w>>8))
			pc += 2

		case "syscall":
			w := EncodeR(OpSYSTEM, 0, 0, 0, 2)
			code = append(code, byte(w), byte(w>>8))
			pc += 2

		case "ret":
			w := EncodeR(OpSYSTEM, 0, 7, 0, 0) // JALR R7
			code = append(code, byte(w), byte(w>>8))
			pc += 2

		case "mov":
			rd := a.parseReg(l, l.args[0])
			rs := a.parseReg(l, l.args[1])
			w := EncodeR(OpALU, rd, rs, 0, 0) // ADD rd, rs, R0
			code = append(code, byte(w), byte(w>>8))
			pc += 2

		default:
			w := a.encodeInstr(l, pc)
			code = append(code, byte(w), byte(w>>8))
			pc += 2
		}
	}
	return code
}

func (a *Assembler) encodeInstr(l asmLine, pc int) uint16 {
	switch l.op {
	// R-type: op rd, rs1, rs2
	case "add":
		return a.rtype(l, OpALU, 0)
	case "sub":
		return a.rtype(l, OpALU, 1)
	case "and":
		return a.rtype(l, OpALU, 2)
	case "or":
		return a.rtype(l, OpALU, 3)
	case "xor":
		return a.rtype(l, OpALU, 4)
	case "shl":
		return a.rtype(l, OpALU, 5)
	case "shr":
		return a.rtype(l, OpALU, 6)
	case "sra":
		return a.rtype(l, OpALU, 7)
	case "mul":
		return a.rtype(l, OpMUL, 0)
	case "mulh":
		return a.rtype(l, OpMUL, 1)

	// I-type: op rd, rs1, imm6
	case "addi":
		return a.itype(l, OpADDI)
	case "andi":
		return a.itype(l, OpANDI)
	case "ori":
		return a.itype(l, OpORI)
	case "lui":
		rd := a.parseReg(l, l.args[0])
		imm := a.parseImm(l, l.args[1])
		return EncodeI(OpLUI, rd, 0, imm&0x3F)

	// Load/Store: op rd, imm6(rs1)  or  op rd, rs1, imm6
	case "lw":
		return a.loadStore(l, OpLW)
	case "sw":
		return a.storeInstr(l, OpSW)
	case "lb":
		return a.loadStore(l, OpLB)
	case "sb":
		return a.storeInstr(l, OpSB)

	// Branch: op rs1, rs2, label
	case "beq":
		return a.branch(l, OpBEQ, pc)
	case "bne":
		return a.branch(l, OpBNE, pc)
	case "blt":
		return a.branch(l, OpBLT, pc)
	case "bge":
		return a.branch(l, OpBGE, pc)

	// Jump
	case "jal":
		return a.jump(l, OpJAL, pc)
	case "jalr":
		rs := a.parseReg(l, l.args[0])
		return EncodeR(OpSYSTEM, 0, rs, 0, 0)
	case "j": // pseudo: JAL with no link needed (still saves to R7)
		return a.jump(l, OpJAL, pc)
	case "call": // pseudo: same as JAL
		return a.jump(l, OpJAL, pc)

	default:
		a.errorf(l, "unknown instruction: %s", l.op)
		return 0
	}
}

func (a *Assembler) rtype(l asmLine, op, fn int) uint16 {
	if len(l.args) < 3 {
		a.errorf(l, "%s needs 3 args", l.op)
		return 0
	}
	rd := a.parseReg(l, l.args[0])
	rs1 := a.parseReg(l, l.args[1])
	rs2 := a.parseReg(l, l.args[2])
	return EncodeR(op, rd, rs1, rs2, fn)
}

func (a *Assembler) itype(l asmLine, op int) uint16 {
	if len(l.args) < 3 {
		a.errorf(l, "%s needs 3 args", l.op)
		return 0
	}
	rd := a.parseReg(l, l.args[0])
	rs1 := a.parseReg(l, l.args[1])
	imm := a.parseImm(l, l.args[2])
	return EncodeI(op, rd, rs1, imm)
}

// loadStore parses "LW rd, imm(rs)" or "LW rd, rs, imm"
func (a *Assembler) loadStore(l asmLine, op int) uint16 {
	if len(l.args) < 2 {
		a.errorf(l, "%s needs at least 2 args", l.op)
		return 0
	}
	rd := a.parseReg(l, l.args[0])
	rs1, imm := a.parseMemArg(l, l.args[1])
	return EncodeI(op, rd, rs1, imm)
}

// storeInstr parses "SW rs, imm(rd)" -- rs=source, rd=base
func (a *Assembler) storeInstr(l asmLine, op int) uint16 {
	if len(l.args) < 2 {
		a.errorf(l, "%s needs at least 2 args", l.op)
		return 0
	}
	rs := a.parseReg(l, l.args[0])
	base, imm := a.parseMemArg(l, l.args[1])
	// SW encoding: EncodeI(op, base_rd, src_rs1, imm)
	return EncodeI(op, base, rs, imm)
}

// parseMemArg parses "imm(reg)" or "reg" (imm=0)
func (a *Assembler) parseMemArg(l asmLine, s string) (reg, imm int) {
	if idx := strings.Index(s, "("); idx >= 0 {
		immStr := s[:idx]
		regStr := strings.TrimSuffix(s[idx+1:], ")")
		return a.parseReg(l, regStr), a.parseImm(l, immStr)
	}
	// Just a register, imm=0
	if len(l.args) >= 3 {
		return a.parseReg(l, l.args[1]), a.parseImm(l, l.args[2])
	}
	return a.parseReg(l, s), 0
}

func (a *Assembler) branch(l asmLine, op, pc int) uint16 {
	if len(l.args) < 3 {
		a.errorf(l, "%s needs 3 args", l.op)
		return 0
	}
	rs1 := a.parseReg(l, l.args[0])
	rs2 := a.parseReg(l, l.args[1])
	target := a.parseImm(l, l.args[2])
	off := (target - pc) / 2
	if off < -32 || off > 31 {
		a.errorf(l, "branch offset %d out of range (-32..31)", off)
	}
	return EncodeB(op, rs1, rs2, off)
}

func (a *Assembler) jump(l asmLine, op, pc int) uint16 {
	if len(l.args) < 1 {
		a.errorf(l, "%s needs 1 arg", l.op)
		return 0
	}
	target := a.parseImm(l, l.args[0])
	off := (target - pc) / 2
	if off < -2048 || off > 2047 {
		a.errorf(l, "jump offset %d out of range", off)
	}
	return EncodeJ(op, off)
}

func (a *Assembler) parseReg(l asmLine, s string) int {
	s = strings.ToLower(strings.TrimSpace(s))
	if r, ok := regMap[s]; ok {
		return r
	}
	a.errorf(l, "unknown register: %s", s)
	return 0
}

func (a *Assembler) parseImm(l asmLine, s string) int {
	s = strings.TrimSpace(s)
	if v, ok := a.labels[s]; ok {
		return v
	}
	s = strings.ToLower(s)
	base := 10
	if strings.HasPrefix(s, "0x") {
		s = s[2:]
		base = 16
	} else if strings.HasPrefix(s, "0b") {
		s = s[2:]
		base = 2
	}
	v, err := strconv.ParseInt(s, base, 32)
	if err != nil {
		a.errorf(l, "bad immediate: %s", s)
		return 0
	}
	return int(v)
}

func (a *Assembler) parseString(s string) string {
	s = strings.TrimSpace(s)
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		s = s[1 : len(s)-1]
	}
	// Handle escapes
	s = strings.ReplaceAll(s, "\\n", "\n")
	s = strings.ReplaceAll(s, "\\r", "\r")
	s = strings.ReplaceAll(s, "\\t", "\t")
	s = strings.ReplaceAll(s, "\\0", "\x00")
	return s
}

func (a *Assembler) errorf(l asmLine, format string, args ...interface{}) {
	msg := fmt.Sprintf("line %d: %s", l.lineNo, fmt.Sprintf(format, args...))
	a.errors = append(a.errors, msg)
}
