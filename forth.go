package nand16

import (
	"fmt"
	"strconv"
	"strings"
)

/*
Register convention:
  R6 (SP) = data stack pointer (grows down)
  R5      = return stack pointer (grows down)
  R4      = TOS (top of data stack, cached)
  R7      = link register
  R1-R3   = scratch

Word calling convention:
  Caller:  JAL target  (sets R7 = return addr)
  Callee:  prologue saves R7 to return stack
           epilogue restores R7, JALR R7
  This allows free use of JAL for internal jumps within words.
*/

type ForthCompiler struct {
	code      []uint16
	words     map[string]int
	here      int
	baseAddr  int
	ctrlStack []ctrlEntry
	compiling bool
	curWord   string
	errors    []string
	// Runtime subroutine addresses
	udivAddr int // unsigned 16-bit division
}

type ctrlEntry struct {
	kind string
	addr int
}

func NewForthCompiler() *ForthCompiler {
	return &ForthCompiler{words: make(map[string]int)}
}

func (fc *ForthCompiler) Compile(source string, startAddr int) ([]byte, error) {
	fc.here = startAddr
	fc.baseAddr = startAddr
	fc.code = nil
	fc.errors = nil

	fc.emitBootstrap()

	tokens := fc.tokenize(source)
	for i := 0; i < len(tokens); i++ {
		tok := strings.ToLower(tokens[i])
		switch {
		case tok == ":":
			i++
			if i >= len(tokens) {
				fc.errorf("expected word name after :")
				break
			}
			// Emit skip jump over word body (for main code flow)
			fc.ctrlStack = append(fc.ctrlStack, ctrlEntry{"wordskip", fc.here})
			fc.emit(EncodeJ(OpJAL, 0)) // placeholder: JAL skip (12-bit range)

			fc.curWord = strings.ToLower(tokens[i])
			fc.words[fc.curWord] = fc.here
			fc.compiling = true
			// Prologue: save R7 to return stack
			fc.emit(EncodeI(OpADDI, 5, 5, -2))
			fc.emit(EncodeI(OpSW, 5, 7, 0))

		case tok == ";":
			// Epilogue: restore R7, return
			fc.emit(EncodeI(OpLW, 7, 5, 0))
			fc.emit(EncodeI(OpADDI, 5, 5, 2))
			fc.emit(EncodeR(OpSYSTEM, 0, 7, 0, 0))
			fc.compiling = false
			// Patch skip jump
			fc.patchCtrl("wordskip")

		case tok == "if":
			fc.emitPop(1)
			// BNE R1, R0, 2: if true, skip JAL → enter IF body
			fc.emit(EncodeB(OpBNE, 1, 0, 2))
			fc.ctrlStack = append(fc.ctrlStack, ctrlEntry{"if", fc.here})
			fc.emit(EncodeJ(OpJAL, 0)) // placeholder: jump to ELSE/THEN

		case tok == "else":
			elseJump := fc.here
			fc.emit(EncodeJ(OpJAL, 0)) // JAL forward (12-bit range)
			fc.patchCtrl("if")
			fc.ctrlStack = append(fc.ctrlStack, ctrlEntry{"else", elseJump})

		case tok == "then":
			fc.patchCtrl("if", "else")

		case tok == "begin":
			fc.ctrlStack = append(fc.ctrlStack, ctrlEntry{"begin", fc.here})

		case tok == "until":
			fc.emitPop(1)
			begin := fc.popCtrl("begin")
			off := (begin.addr - fc.here - 2) / 2 // -2 for the BNE we'll emit first
			// BNE R1, R0, 2: if true (non-zero), skip JAL → exit loop
			fc.emit(EncodeB(OpBNE, 1, 0, 2))
			fc.emit(EncodeJ(OpJAL, off)) // loop back

		case tok == "again":
			begin := fc.popCtrl("begin")
			off := (begin.addr - fc.here) / 2
			fc.emit(EncodeJ(OpJAL, off))

		case tok == "while":
			fc.emitPop(1)
			fc.emit(EncodeB(OpBNE, 1, 0, 2)) // if true, skip JAL
			fc.ctrlStack = append(fc.ctrlStack, ctrlEntry{"while", fc.here})
			fc.emit(EncodeJ(OpJAL, 0)) // placeholder: forward to after REPEAT

		case tok == "repeat":
			whileEntry := fc.popCtrl("while")
			beginEntry := fc.popCtrl("begin")
			off := (beginEntry.addr - fc.here) / 2
			fc.emit(EncodeJ(OpJAL, off))
			fc.patchBranch(whileEntry.addr, fc.here)

		case tok == "do":
			fc.emitInstr(EncodeI(OpADDI, 5, 5, -2))
			fc.emitInstr(EncodeI(OpSW, 5, 4, 0))
			fc.emitPop(4)
			fc.emitInstr(EncodeI(OpADDI, 5, 5, -2))
			fc.emitInstr(EncodeI(OpSW, 5, 4, 0))
			fc.emitPop(4)
			fc.ctrlStack = append(fc.ctrlStack, ctrlEntry{"do", fc.here})

		case tok == "loop":
			fc.emitInstr(EncodeI(OpLW, 1, 5, 2))
			fc.emitInstr(EncodeI(OpADDI, 1, 1, 1))
			fc.emitInstr(EncodeI(OpSW, 5, 1, 2))
			fc.emitInstr(EncodeI(OpLW, 2, 5, 0))
			doEntry := fc.popCtrl("do")
			// BGE R1, R2, 2: if index >= limit, skip JAL → exit
			fc.emit(EncodeB(OpBGE, 1, 2, 2))
			off := (doEntry.addr - fc.here) / 2
			fc.emit(EncodeJ(OpJAL, off)) // loop back
			fc.emitInstr(EncodeI(OpADDI, 5, 5, 4))

		case tok == "i":
			fc.emitPush(4)
			fc.emitInstr(EncodeI(OpLW, 4, 5, 2))

		case tok == "j":
			fc.emitPush(4)
			fc.emitInstr(EncodeI(OpLW, 4, 5, 6))

		default:
			fc.compileToken(tok)
		}
	}

	fc.emit(EncodeR(OpSYSTEM, 0, 0, 0, 1)) // HALT
	if len(fc.errors) > 0 {
		return nil, fmt.Errorf("forth: %s", strings.Join(fc.errors, "; "))
	}
	return fc.toBytes(), nil
}

func (fc *ForthCompiler) compileToken(tok string) {
	if v, err := strconv.ParseInt(tok, 0, 32); err == nil {
		fc.emitPushConst(int(v))
		return
	}

	switch tok {
	case "+":
		fc.emitBinOp(0)
	case "-":
		fc.emitBinOp(1)
	case "and":
		fc.emitBinOp(2)
	case "or":
		fc.emitBinOp(3)
	case "xor":
		fc.emitBinOp(4)
	case "lshift":
		fc.emitBinOp(5)
	case "rshift":
		fc.emitBinOp(6)

	case "*":
		fc.emitPop(1)
		fc.emit(EncodeR(OpMUL, 4, 1, 4, 0))

	case "f*":
		// Fixed-point 8.8 multiply: ( a b -- (a*b)>>8 )
		fc.emitPop(1)                              // R1 = a, R4 = b
		fc.emit(EncodeR(OpMUL, 2, 1, 4, 0))       // R2 = low16(a*b)
		fc.emit(EncodeR(OpMUL, 3, 1, 4, 1))       // R3 = high16(a*b)
		fc.emit(EncodeI(OpADDI, 1, 0, 8))         // R1 = 8
		fc.emit(EncodeR(OpALU, 2, 2, 1, 6))       // R2 = low >> 8 (logical)
		fc.emit(EncodeR(OpALU, 3, 3, 1, 5))       // R3 = high << 8
		fc.emit(EncodeR(OpALU, 4, 2, 3, 3))       // R4 = R2 | R3

	case "/":
		// Signed integer division: ( a b -- a/b )
		fc.emitPop(1)                              // R1 = divisor, R4 = dividend
		fc.emit(EncodeI(OpADDI, 2, 0, 0))         // R2 = 0 (initial remainder)
		fc.emitCallRuntime(fc.udivAddr)

	case "mod":
		fc.emitPop(1)
		fc.emit(EncodeI(OpADDI, 2, 0, 0))
		fc.emitCallRuntime(fc.udivAddr)
		fc.emit(EncodeR(OpALU, 4, 2, 0, 0))       // R4 = remainder (R2)

	case "f/":
		// Fixed-point SIGNED divide: ( a b -- (a*256)/b )
		// Handles negative dividend (divisor assumed positive)
		fc.emitPop(1) // R1 = divisor, R4 = dividend (a)

		// Save sign of a, make a positive
		fc.emit(EncodeR(OpALU, 3, 0, 0, 0)) // R3 = 0 (sign flag)
		posLabel := fc.here
		fc.emit(EncodeB(OpBGE, 4, 0, 0)) // if R4 >= 0, skip negate
		fc.emit(EncodeR(OpALU, 4, 0, 4, 1)) // R4 = -R4
		fc.emit(EncodeI(OpADDI, 3, 0, -1))  // R3 = -1 (was negative)
		fc.patchBranch(posLabel, fc.here)

		// Save sign flag to return stack (udiv clobbers R3)
		fc.emit(EncodeI(OpADDI, 5, 5, -2)) // RSP -= 2
		fc.emit(EncodeI(OpSW, 5, 3, 0))    // mem[RSP] = sign flag

		// Prepare extended dividend: R2 = |a| >> 8, R4 = |a| << 8
		fc.emit(EncodeI(OpADDI, 3, 0, 8))         // R3 = 8
		fc.emit(EncodeR(OpALU, 2, 4, 3, 6))       // R2 = |a| >> 8 (logical OK, a is positive)
		fc.emit(EncodeR(OpALU, 4, 4, 3, 5))       // R4 = |a| << 8
		fc.emitCallRuntime(fc.udivAddr)            // R4 = quotient

		// Restore sign flag and negate if needed
		fc.emit(EncodeI(OpLW, 3, 5, 0))    // R3 = sign flag
		fc.emit(EncodeI(OpADDI, 5, 5, 2))  // RSP += 2
		negLabel := fc.here
		fc.emit(EncodeB(OpBEQ, 3, 0, 0))   // if R3 == 0, skip negate
		fc.emit(EncodeR(OpALU, 4, 0, 4, 1)) // R4 = -R4
		fc.patchBranch(negLabel, fc.here)

	case "*/":
		// ( a b c -- a*b/c ) with 32-bit intermediate
		fc.emitPop(1) // R1 = c (divisor)
		fc.emitPop(2) // R2 = b, R4 = a
		fc.emit(EncodeR(OpMUL, 3, 4, 2, 1)) // R3 = high16(a*b)
		fc.emit(EncodeR(OpMUL, 4, 4, 2, 0)) // R4 = low16(a*b)
		fc.emit(EncodeR(OpALU, 2, 3, 0, 0)) // R2 = R3 (initial remainder = high word)
		fc.emitCallRuntime(fc.udivAddr)

	case "negate":
		fc.emit(EncodeR(OpALU, 4, 0, 4, 1))

	case "abs":
		fc.emit(EncodeB(OpBGE, 4, 0, 2))
		fc.emit(EncodeR(OpALU, 4, 0, 4, 1))

	case "dup":
		fc.emitPush(4)

	case "drop":
		fc.emitPop(4)

	case "swap":
		fc.emitInstr(EncodeI(OpLW, 1, 6, 0))
		fc.emitInstr(EncodeI(OpSW, 6, 4, 0))
		fc.emitInstr(EncodeR(OpALU, 4, 1, 0, 0))

	case "over":
		fc.emitPush(4)
		fc.emitInstr(EncodeI(OpLW, 4, 6, 2))

	case "rot":
		fc.emitInstr(EncodeI(OpLW, 1, 6, 0))
		fc.emitInstr(EncodeI(OpLW, 2, 6, 2))
		fc.emitInstr(EncodeI(OpSW, 6, 1, 2))
		fc.emitInstr(EncodeI(OpSW, 6, 4, 0))
		fc.emitInstr(EncodeR(OpALU, 4, 2, 0, 0))

	case "nip":
		fc.emitInstr(EncodeI(OpADDI, 6, 6, 2))

	case "2dup":
		fc.emitInstr(EncodeI(OpLW, 1, 6, 0))     // R1 = second
		fc.emitInstr(EncodeI(OpADDI, 6, 6, -2))   // sp -= 2
		fc.emitInstr(EncodeI(OpSW, 6, 1, 0))      // mem[sp] = second
		fc.emitInstr(EncodeI(OpADDI, 6, 6, -2))   // sp -= 2
		fc.emitInstr(EncodeI(OpSW, 6, 4, 0))      // mem[sp] = TOS

	case "=":
		fc.emitCompare(OpBEQ) // equal

	case "<>":
		fc.emitCompare(OpBNE) // not equal

	case "<":
		// ( a b -- flag ) flag = a < b
		// After pop: R1=b (was TOS), R4=a (from stack)
		// Test: R4 < R1 (a < b)
		fc.emitPop(1)
		fc.emit(EncodeB(OpBLT, 4, 1, 3)) // if a < b, jump to true
		fc.emit(EncodeI(OpADDI, 4, 0, 0))
		fc.emit(EncodeB(OpBEQ, 0, 0, 2))
		fc.emit(EncodeI(OpADDI, 4, 0, -1))

	case ">":
		fc.emitPop(1)
		fc.emit(EncodeB(OpBLT, 1, 4, 3)) // if b < a, i.e. a > b
		fc.emit(EncodeI(OpADDI, 4, 0, 0))
		fc.emit(EncodeB(OpBEQ, 0, 0, 2))
		fc.emit(EncodeI(OpADDI, 4, 0, -1))

	case "0=":
		fc.emit(EncodeB(OpBEQ, 4, 0, 3))
		fc.emit(EncodeI(OpADDI, 4, 0, 0))
		fc.emit(EncodeB(OpBEQ, 0, 0, 2))
		fc.emit(EncodeI(OpADDI, 4, 0, -1))

	case "0<":
		fc.emit(EncodeB(OpBLT, 4, 0, 3))
		fc.emit(EncodeI(OpADDI, 4, 0, 0))
		fc.emit(EncodeB(OpBEQ, 0, 0, 2))
		fc.emit(EncodeI(OpADDI, 4, 0, -1))

	case "0>":
		fc.emit(EncodeB(OpBLT, 0, 4, 3)) // 0 < R4 means R4 > 0
		fc.emit(EncodeI(OpADDI, 4, 0, 0))
		fc.emit(EncodeB(OpBEQ, 0, 0, 2))
		fc.emit(EncodeI(OpADDI, 4, 0, -1))

	case "max":
		fc.emitPop(1) // R1=b, R4=a
		fc.emit(EncodeB(OpBLT, 1, 4, 2)) // if b < a, keep a
		fc.emit(EncodeR(OpALU, 4, 1, 0, 0)) // else R4=b

	case "min":
		fc.emitPop(1)
		fc.emit(EncodeB(OpBLT, 4, 1, 2)) // if a < b, keep a
		fc.emit(EncodeR(OpALU, 4, 1, 0, 0)) // else R4=b

	case "@":
		fc.emitInstr(EncodeI(OpLW, 4, 4, 0))

	case "!":
		fc.emitPop(1)
		fc.emitInstr(EncodeI(OpSW, 1, 4, 0)) // mem[R1(addr)] = R4(val)
		fc.emitPop(4)

	case "c@":
		fc.emitInstr(EncodeI(OpLB, 4, 4, 0))

	case "c!":
		fc.emitPop(1)
		fc.emitInstr(EncodeI(OpSB, 1, 4, 0)) // mem[R1(addr)] = byte(R4)
		fc.emitPop(4)

	case "emit":
		fc.emitInstr(EncodeR(OpALU, 2, 4, 0, 0))
		fc.emitInstr(EncodeI(OpADDI, 1, 0, 1))
		fc.emit(EncodeR(OpSYSTEM, 0, 0, 0, 2))
		fc.emitPop(4)

	case "pixel":
		// ( color y x -- ) 8-bit write
		fc.emitInstr(EncodeR(OpALU, 2, 4, 0, 0))
		fc.emitPop(4)
		fc.emitInstr(EncodeR(OpALU, 3, 4, 0, 0))
		fc.emitPop(4)
		fc.emitLoadConst(1, 6)
		fc.emitInstr(EncodeR(OpALU, 3, 3, 1, 5))
		fc.emitLoadConst(1, 0xF000)
		fc.emitInstr(EncodeR(OpALU, 1, 1, 3, 0))
		fc.emitInstr(EncodeR(OpALU, 1, 1, 2, 0))
		fc.emitInstr(EncodeI(OpSB, 1, 4, 0))
		fc.emitPop(4)

	case "pixel16":
		// ( color16 y x -- ) 16-bit RGB555 write
		fc.emitInstr(EncodeR(OpALU, 2, 4, 0, 0)) // R2 = x
		fc.emitPop(4)
		fc.emitInstr(EncodeR(OpALU, 3, 4, 0, 0)) // R3 = y
		fc.emitPop(4)                               // TOS = color16
		fc.emitLoadConst(1, 7)
		fc.emitInstr(EncodeR(OpALU, 3, 3, 1, 5)) // R3 = y*128
		fc.emitInstr(EncodeR(OpALU, 2, 2, 2, 0)) // R2 = x*2
		fc.emitLoadConst(1, 0xF000)
		fc.emitInstr(EncodeR(OpALU, 1, 1, 3, 0))
		fc.emitInstr(EncodeR(OpALU, 1, 1, 2, 0))
		fc.emitInstr(EncodeI(OpSW, 1, 4, 0))     // 16-bit store
		fc.emitPop(4)

	case "fb-addr":
		fc.emitPop(1) // R1=x, TOS=y
		fc.emitLoadConst(2, 6)
		fc.emitInstr(EncodeR(OpALU, 4, 4, 2, 5)) // TOS = y<<6
		fc.emitLoadConst(2, 0xF000)
		fc.emitInstr(EncodeR(OpALU, 4, 4, 2, 0))
		fc.emitInstr(EncodeR(OpALU, 4, 4, 1, 0))

	case "halt":
		fc.emit(EncodeR(OpSYSTEM, 0, 0, 0, 1))

	default:
		if addr, ok := fc.words[tok]; ok {
			fc.emitCall(addr)
		} else {
			fc.errorf("unknown word: %s", tok)
		}
	}
}

// emitCompare: comparison using SUB then branch
func (fc *ForthCompiler) emitCompare(branchOp int) {
	fc.emitPop(1) // R1 = old TOS, R4 = second
	fc.emit(EncodeR(OpALU, 4, 4, 1, 1)) // R4 = R4 - R1
	fc.emit(EncodeB(branchOp, 4, 0, 3))
	fc.emit(EncodeI(OpADDI, 4, 0, 0))
	fc.emit(EncodeB(OpBEQ, 0, 0, 2))
	fc.emit(EncodeI(OpADDI, 4, 0, -1))
}

func (fc *ForthCompiler) emit(instr uint16) {
	fc.code = append(fc.code, instr)
	fc.here += 2
}

func (fc *ForthCompiler) emitInstr(instr uint16) { fc.emit(instr) }

func (fc *ForthCompiler) emitPush(srcReg int) {
	fc.emit(EncodeI(OpADDI, 6, 6, -2))
	fc.emit(EncodeI(OpSW, 6, 4, 0))
	if srcReg != 4 {
		fc.emit(EncodeR(OpALU, 4, srcReg, 0, 0))
	}
}

func (fc *ForthCompiler) emitPop(dstReg int) {
	if dstReg != 4 {
		fc.emit(EncodeR(OpALU, dstReg, 4, 0, 0))
	}
	fc.emit(EncodeI(OpLW, 4, 6, 0))
	fc.emit(EncodeI(OpADDI, 6, 6, 2))
}

func (fc *ForthCompiler) emitPushConst(val int) {
	fc.emitPush(4)
	fc.emitLoadConstToR4(val)
}

func (fc *ForthCompiler) emitLoadConstToR4(val int) {
	fc.emitLoadConst(4, val&0xFFFF)
}

func (fc *ForthCompiler) emitLoadConst(rd, val int) {
	val = val & 0xFFFF
	signed := int16(uint16(val))

	// Simple: fits in imm6 (-32..31) → 1 instruction
	if signed >= -32 && signed <= 31 {
		fc.emit(EncodeI(OpADDI, rd, 0, int(signed)))
		return
	}

	// Medium: two ADDI (-63..62) → 2 instructions
	if signed >= 32 && signed <= 62 {
		fc.emit(EncodeI(OpADDI, rd, 0, 31))
		fc.emit(EncodeI(OpADDI, rd, rd, int(signed-31)))
		return
	}
	if signed >= -63 && signed <= -33 {
		fc.emit(EncodeI(OpADDI, rd, 0, -32))
		fc.emit(EncodeI(OpADDI, rd, rd, int(signed+32)))
		return
	}

	// LUI shortcut: if lower part is 0, LUI alone suffices (1 instruction)
	upper := (val >> 10) & 0x3F
	lower := val - (upper << 10)
	if upper != 0 && lower == 0 {
		fc.emit(EncodeI(OpLUI, rd, 0, upper))
		return
	}

	// General: shift-based. val = (hi << 8) + lo
	// Typically 3-8 instructions, much better than LUI+chain for negative values
	hi := (val >> 8) & 0xFF
	lo := val & 0xFF
	hiSigned := int(int8(byte(hi)))

	fc.emitLoadSmall(rd, hiSigned)
	fc.emit(EncodeI(OpADDI, 3, 0, 8))
	fc.emit(EncodeR(OpALU, rd, rd, 3, 5)) // rd <<= 8
	fc.emitAdditiveChain(rd, int(lo))
}

// emitLoadSmall loads a small value (-128..127) into rd using minimal instructions
func (fc *ForthCompiler) emitLoadSmall(rd, val int) {
	if val >= -32 && val <= 31 {
		fc.emit(EncodeI(OpADDI, rd, 0, val))
	} else if val >= 32 && val <= 62 {
		fc.emit(EncodeI(OpADDI, rd, 0, 31))
		fc.emit(EncodeI(OpADDI, rd, rd, val-31))
	} else if val >= -63 && val <= -33 {
		fc.emit(EncodeI(OpADDI, rd, 0, -32))
		fc.emit(EncodeI(OpADDI, rd, rd, val+32))
	} else if val >= 63 && val <= 127 {
		fc.emit(EncodeI(OpADDI, rd, 0, 31))
		fc.emit(EncodeI(OpADDI, rd, rd, 31))
		fc.emit(EncodeI(OpADDI, rd, rd, val-62))
	} else { // -128..-64
		fc.emit(EncodeI(OpADDI, rd, 0, -32))
		fc.emit(EncodeI(OpADDI, rd, rd, -32))
		fc.emit(EncodeI(OpADDI, rd, rd, val+64))
	}
}

// emitAdditiveChain adds a non-negative value to rd using ADDI chain
func (fc *ForthCompiler) emitAdditiveChain(rd, val int) {
	for val > 31 {
		fc.emit(EncodeI(OpADDI, rd, rd, 31))
		val -= 31
	}
	if val > 0 {
		fc.emit(EncodeI(OpADDI, rd, rd, val))
	}
}

func (fc *ForthCompiler) emitBinOp(fn int) {
	fc.emitPop(1)
	fc.emit(EncodeR(OpALU, 4, 4, 1, fn))
}

func (fc *ForthCompiler) emitCall(target int) {
	off := (target - fc.here) / 2
	if off >= -2048 && off <= 2047 {
		fc.emit(EncodeJ(OpJAL, off))
	} else {
		// Long call: load target address into R2, then JALR R2
		fc.emitLoadConst(2, target)
		fc.emit(EncodeR(OpSYSTEM, 0, 2, 0, 0)) // JALR R2
	}
}

// emitCallRuntime calls a runtime subroutine. Since runtime subs save/restore R7,
// we don't need to handle R7 here. But if called from within a word (where R7
// is already saved in prologue), we just JAL directly.
func (fc *ForthCompiler) emitCallRuntime(target int) {
	off := (target - fc.here) / 2
	if off >= -2048 && off <= 2047 {
		fc.emit(EncodeJ(OpJAL, off))
	} else {
		fc.emitLoadConst(2, target)
		fc.emit(EncodeR(OpSYSTEM, 0, 2, 0, 0))
	}
}

func (fc *ForthCompiler) emitRet() {
	fc.emit(EncodeR(OpSYSTEM, 0, 7, 0, 0))
}

func (fc *ForthCompiler) emitBootstrap() {
	fc.emitLoadConst(6, 0xEF00) // SP
	fc.emitLoadConst(5, 0xDF00) // RSP
	fc.emit(EncodeI(OpADDI, 4, 0, 0)) // TOS = 0
	// Jump over runtime subroutines
	skipAddr := fc.here
	fc.emit(EncodeJ(OpJAL, 0)) // placeholder

	// === Runtime: _udiv ===
	// Input: R4 = dividend, R1 = divisor, R2 = initial remainder (0 for normal div)
	// Output: R4 = quotient, R2 = remainder
	// Clobbers: R3
	fc.udivAddr = fc.here
	fc.emit(EncodeI(OpADDI, 5, 5, -2))   // rsp -= 2
	fc.emit(EncodeI(OpSW, 5, 7, 0))      // save R7
	fc.emit(EncodeI(OpADDI, 3, 0, 16))   // counter = 16
	// _loop:
	fc.emit(EncodeR(OpALU, 2, 2, 2, 0))  // rem += rem (rem <<= 1)
	fc.emit(EncodeB(OpBGE, 4, 0, 2))     // if dividend >= 0, skip
	fc.emit(EncodeI(OpADDI, 2, 2, 1))    // rem |= 1
	fc.emit(EncodeR(OpALU, 4, 4, 4, 0))  // dividend <<= 1
	fc.emit(EncodeB(OpBLT, 2, 1, 3))     // if rem < div, skip sub
	fc.emit(EncodeR(OpALU, 2, 2, 1, 1))  // rem -= div
	fc.emit(EncodeI(OpADDI, 4, 4, 1))    // quotient |= 1
	// _skip_sub:
	fc.emit(EncodeI(OpADDI, 3, 3, -1))   // counter--
	fc.emit(EncodeB(OpBLT, 0, 3, -8))    // if 0 < counter, loop back
	fc.emit(EncodeI(OpLW, 7, 5, 0))      // restore R7
	fc.emit(EncodeI(OpADDI, 5, 5, 2))    // rsp += 2
	fc.emit(EncodeR(OpSYSTEM, 0, 7, 0, 0)) // JALR R7

	// Patch skip jump
	fc.patchBranch(skipAddr, fc.here)
}

func (fc *ForthCompiler) patchCtrl(kinds ...string) {
	for i := len(fc.ctrlStack) - 1; i >= 0; i-- {
		for _, k := range kinds {
			if fc.ctrlStack[i].kind == k {
				entry := fc.ctrlStack[i]
				fc.ctrlStack = append(fc.ctrlStack[:i], fc.ctrlStack[i+1:]...)
				fc.patchBranch(entry.addr, fc.here)
				return
			}
		}
	}
	fc.errorf("unmatched control structure")
}

func (fc *ForthCompiler) popCtrl(kind string) ctrlEntry {
	for i := len(fc.ctrlStack) - 1; i >= 0; i-- {
		if fc.ctrlStack[i].kind == kind {
			entry := fc.ctrlStack[i]
			fc.ctrlStack = append(fc.ctrlStack[:i], fc.ctrlStack[i+1:]...)
			return entry
		}
	}
	fc.errorf("unmatched: %s", kind)
	return ctrlEntry{}
}

func (fc *ForthCompiler) patchBranch(instrAddr, targetAddr int) {
	codeIdx := (instrAddr - fc.baseAddr) / 2
	if codeIdx < 0 || codeIdx >= len(fc.code) {
		fc.errorf("patch OOB: %d", instrAddr)
		return
	}
	instr := fc.code[codeIdx]
	op := instr >> 12
	off := (targetAddr - instrAddr) / 2
	switch op {
	case OpBEQ, OpBNE, OpBLT, OpBGE:
		if off < -32 || off > 31 {
			fc.errorf("Branch offset overflow at 0x%X → 0x%X (off=%d words)", instrAddr, targetAddr, off)
		}
		fc.code[codeIdx] = (instr & 0xFFC0) | uint16(off&0x3F)
	case OpJAL:
		if off < -2048 || off > 2047 {
			fc.errorf("JAL offset overflow at 0x%X → 0x%X (off=%d words)", instrAddr, targetAddr, off)
		}
		fc.code[codeIdx] = (instr & 0xF000) | uint16(off&0xFFF)
	}
}

func (fc *ForthCompiler) toBytes() []byte {
	out := make([]byte, len(fc.code)*2)
	for i, w := range fc.code {
		out[i*2] = byte(w)
		out[i*2+1] = byte(w >> 8)
	}
	return out
}

func (fc *ForthCompiler) tokenize(source string) []string {
	var tokens []string
	for _, line := range strings.Split(source, "\n") {
		if idx := strings.Index(line, "\\"); idx >= 0 {
			line = line[:idx]
		}
		for {
			start := strings.Index(line, "(")
			end := strings.Index(line, ")")
			if start >= 0 && end > start {
				line = line[:start] + line[end+1:]
			} else {
				break
			}
		}
		for _, tok := range strings.Fields(line) {
			tokens = append(tokens, tok)
		}
	}
	return tokens
}

func (fc *ForthCompiler) errorf(format string, args ...interface{}) {
	fc.errors = append(fc.errors, fmt.Sprintf(format, args...))
}
