package asmc

import (
	"fmt"
	"nand16/internal/op"
	"strconv"
	"strings"
)

// Assembler translates NAND-16 assembly to binary.
type Assembler struct {
	labels map[string]int
	errors []string
	base   int // address the emitted code will be loaded at (label origin)
}

// Assemble translates assembly source to binary loaded at address 0.
func Assemble(source string) ([]byte, error) {
	return AssembleAt(source, 0)
}

// AssembleAt translates assembly source to binary, resolving labels and
// PC-relative offsets as if the code were loaded at the given base address.
// The returned bytes are not padded; the caller loads them at base.
func AssembleAt(source string, base int) ([]byte, error) {
	a := &Assembler{labels: make(map[string]int), base: base}
	lines := a.tokenize(source)

	// Relaxation: decide which `call`s need the long-call trampoline.
	a.relaxCalls(lines)

	// Pass 1: collect labels
	a.labels = make(map[string]int)
	a.errors = nil
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

// relaxCalls iteratively marks `call` instructions whose JAL offset falls out of
// range, so they are emitted as the long-call trampoline. Marking only ever
// grows (long calls take more space, which can push later calls out of range),
// so the fixpoint is reached monotonically.
func (a *Assembler) relaxCalls(lines []asmLine) {
	for iter := 0; iter < len(lines)+2; iter++ {
		a.labels = make(map[string]int)
		a.errors = nil
		a.pass1(lines)

		changed := false
		pc := a.base
		for i := range lines {
			l := &lines[i]
			if l.op == "" {
				continue
			}
			if l.op == "call" && !l.callLong && len(l.args) >= 1 {
				target := a.lookupOrParse(l.args[0])
				if !jumpInRange(wordOffset(pc, target)) {
					l.callLong = true
					changed = true
				}
			}
			pc += a.instrSize(*l)
		}
		if !changed {
			break
		}
	}
	a.errors = nil // discard transient relaxation errors
}

// lookupOrParse resolves a label or numeric literal without recording errors;
// unknown symbols resolve to 0 (the real error surfaces in pass2).
func (a *Assembler) lookupOrParse(s string) int {
	s = strings.TrimSpace(s)
	if v, ok := a.labels[s]; ok {
		return v
	}
	t := strings.ToLower(s)
	base := 10
	if strings.HasPrefix(t, "0x") {
		t, base = t[2:], 16
	} else if strings.HasPrefix(t, "0b") {
		t, base = t[2:], 2
	}
	v, err := strconv.ParseInt(t, base, 32)
	if err != nil {
		return 0
	}
	return int(v)
}

type asmLine struct {
	lineNo   int
	label    string
	op       string
	args     []string
	orig     string
	callLong bool // resolved by relaxCalls: emit the long-call trampoline
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
	pc := a.base
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
	case "li": // pseudo: variable-length load of a 16-bit immediate
		if len(l.args) < 2 {
			return 0
		}
		rd := a.parseReg(l, l.args[0])
		imm := a.parseImm(l, l.args[1])
		return len(expandLI(rd, imm)) * 2
	case "call": // pseudo: jal when in range, else the long-call trampoline
		if l.callLong {
			return longCallWords * 2
		}
		return 2
	default:
		return 2
	}
}

func (a *Assembler) pass2(lines []asmLine) []byte {
	code := make([]byte, 0, 1024)
	pc := a.base
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
				code = appendWordLE(code, v)
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
			// Pseudo-instruction: load a 16-bit immediate via expandLI.
			rd := a.parseReg(l, l.args[0])
			imm := a.parseImm(l, l.args[1])
			for _, w := range expandLI(rd, imm) {
				code = appendWordLE(code, w)
				pc += 2
			}

		case "call":
			// Pseudo-instruction: jal when in range, else long-call trampoline.
			target := a.parseImm(l, l.args[0])
			if l.callLong {
				for _, w := range longCall(target) {
					code = appendWordLE(code, w)
					pc += 2
				}
			} else {
				off := wordOffset(pc, target)
				if !jumpInRange(off) {
					a.errorf(l, "call offset %d out of range", off)
				}
				code = appendWordLE(code, op.EncodeJ(op.OpJAL, off))
				pc += 2
			}

		case "nop":
			code = appendWordLE(code, op.EncodeR(op.OpSYSTEM, 0, 0, 0, 7))
			pc += 2

		case "halt":
			code = appendWordLE(code, op.EncodeR(op.OpSYSTEM, 0, 0, 0, 1))
			pc += 2

		case "syscall":
			code = appendWordLE(code, op.EncodeR(op.OpSYSTEM, 0, 0, 0, 2))
			pc += 2

		case "ret":
			code = appendWordLE(code, op.EncodeR(op.OpSYSTEM, 0, 7, 0, 0)) // JALR R7
			pc += 2

		case "mov":
			rd := a.parseReg(l, l.args[0])
			rs := a.parseReg(l, l.args[1])
			code = appendWordLE(code, op.EncodeR(op.OpALU, rd, rs, 0, 0)) // ADD rd, rs, R0
			pc += 2

		default:
			code = appendWordLE(code, a.encodeInstr(l, pc))
			pc += 2
		}
	}
	return code
}

func (a *Assembler) encodeInstr(l asmLine, pc int) uint16 {
	switch l.op {
	// R-type: op rd, rs1, rs2
	case "add":
		return a.rtype(l, op.OpALU, 0)
	case "sub":
		return a.rtype(l, op.OpALU, 1)
	case "and":
		return a.rtype(l, op.OpALU, 2)
	case "or":
		return a.rtype(l, op.OpALU, 3)
	case "xor":
		return a.rtype(l, op.OpALU, 4)
	case "shl":
		return a.rtype(l, op.OpALU, 5)
	case "shr":
		return a.rtype(l, op.OpALU, 6)
	case "sra":
		return a.rtype(l, op.OpALU, 7)
	case "mul":
		return a.rtype(l, op.OpMUL, 0)
	case "mulh":
		return a.rtype(l, op.OpMUL, 1)

	// I-type: op rd, rs1, imm6
	case "addi":
		return a.itype(l, op.OpADDI)
	case "andi":
		return a.itype(l, op.OpANDI)
	case "ori":
		return a.itype(l, op.OpORI)
	case "lui":
		rd := a.parseReg(l, l.args[0])
		imm := a.parseImm(l, l.args[1])
		return op.EncodeI(op.OpLUI, rd, 0, imm&0x3F)

	// Load/Store: op rd, imm6(rs1)  or  op rd, rs1, imm6
	case "lw":
		return a.loadStore(l, op.OpLW)
	case "sw":
		return a.storeInstr(l, op.OpSW)
	case "lb":
		return a.loadStore(l, op.OpLB)
	case "sb":
		return a.storeInstr(l, op.OpSB)

	// Branch: op rs1, rs2, label
	case "beq":
		return a.branch(l, op.OpBEQ, pc)
	case "bne":
		return a.branch(l, op.OpBNE, pc)
	case "blt":
		return a.branch(l, op.OpBLT, pc)
	case "bge":
		return a.branch(l, op.OpBGE, pc)

	// Jump
	case "jal":
		return a.jump(l, op.OpJAL, pc)
	case "jalr":
		rs := a.parseReg(l, l.args[0])
		return op.EncodeR(op.OpSYSTEM, 0, rs, 0, 0)
	case "j": // pseudo: JAL with no link needed (still saves to R7)
		return a.jump(l, op.OpJAL, pc)

	default:
		a.errorf(l, "unknown instruction: %s", l.op)
		return 0
	}
}

func (a *Assembler) rtype(l asmLine, opcode, fn int) uint16 {
	if len(l.args) < 3 {
		a.errorf(l, "%s needs 3 args", l.op)
		return 0
	}
	rd := a.parseReg(l, l.args[0])
	rs1 := a.parseReg(l, l.args[1])
	rs2 := a.parseReg(l, l.args[2])
	return op.EncodeR(opcode, rd, rs1, rs2, fn)
}

func (a *Assembler) itype(l asmLine, opcode int) uint16 {
	if len(l.args) < 3 {
		a.errorf(l, "%s needs 3 args", l.op)
		return 0
	}
	rd := a.parseReg(l, l.args[0])
	rs1 := a.parseReg(l, l.args[1])
	imm := a.parseImm(l, l.args[2])
	return op.EncodeI(opcode, rd, rs1, imm)
}

// loadStore parses "LW rd, imm(rs)" or "LW rd, rs, imm"
func (a *Assembler) loadStore(l asmLine, opcode int) uint16 {
	if len(l.args) < 2 {
		a.errorf(l, "%s needs at least 2 args", l.op)
		return 0
	}
	rd := a.parseReg(l, l.args[0])
	rs1, imm := a.parseMemArg(l, l.args[1])
	return op.EncodeI(opcode, rd, rs1, imm)
}

// storeInstr parses "SW rs, imm(rd)" -- rs=source, rd=base
func (a *Assembler) storeInstr(l asmLine, opcode int) uint16 {
	if len(l.args) < 2 {
		a.errorf(l, "%s needs at least 2 args", l.op)
		return 0
	}
	rs := a.parseReg(l, l.args[0])
	base, imm := a.parseMemArg(l, l.args[1])
	// SW encoding: op.EncodeI(opcode, base_rd, src_rs1, imm)
	return op.EncodeI(opcode, base, rs, imm)
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

func (a *Assembler) branch(l asmLine, opcode, pc int) uint16 {
	if len(l.args) < 3 {
		a.errorf(l, "%s needs 3 args", l.op)
		return 0
	}
	rs1 := a.parseReg(l, l.args[0])
	rs2 := a.parseReg(l, l.args[1])
	target := a.parseImm(l, l.args[2])
	off := wordOffset(pc, target)
	if !branchInRange(off) {
		a.errorf(l, "branch offset %d out of range (%d..%d)", off, branchMin, branchMax)
	}
	return op.EncodeB(opcode, rs1, rs2, off)
}

func (a *Assembler) jump(l asmLine, opcode, pc int) uint16 {
	if len(l.args) < 1 {
		a.errorf(l, "%s needs 1 arg", l.op)
		return 0
	}
	target := a.parseImm(l, l.args[0])
	off := wordOffset(pc, target)
	if !jumpInRange(off) {
		a.errorf(l, "jump offset %d out of range", off)
	}
	return op.EncodeJ(opcode, off)
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
