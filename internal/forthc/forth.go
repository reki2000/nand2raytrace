package forthc

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
  Caller:  jal target  (sets R7 = return addr)
  Callee:  prologue saves R7 to return stack
           epilogue restores R7, ret (jalr R7)
  This allows free use of jal for internal jumps within words.

Output:
  Compile emits NAND-16 assembly text (.s). All control flow uses labels that
  the assembler (internal/asmc) resolves; constants are emitted as the `li`
  pseudo-instruction and lowered by the assembler.
*/

type ForthCompiler struct {
	out       strings.Builder
	words     map[string]string // forth word name -> asm label
	ctrlStack []ctrlEntry
	errors    []string
	labelN    int
	udivLabel string // runtime unsigned-division subroutine label
}

type ctrlEntry struct {
	kind string
	a    string // associated label (forward target or loop top)
}

func NewForthCompiler() *ForthCompiler {
	return &ForthCompiler{words: make(map[string]string)}
}

// Compile translates Forth source into NAND-16 assembly text.
func (fc *ForthCompiler) Compile(source string) (string, error) {
	fc.out.Reset()
	fc.words = make(map[string]string)
	fc.ctrlStack = nil
	fc.errors = nil
	fc.labelN = 0

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
			// Jump over the word body (inline definition in main flow).
			after := fc.newLabel("word_end")
			fc.pushCtrl("wordskip", after)
			fc.ins("jal %s", after)

			name := strings.ToLower(tokens[i])
			wl := fc.newLabel("word")
			fc.words[name] = wl
			fc.label(wl)
			// Prologue: save R7 to return stack.
			fc.ins("addi r5, r5, -2")
			fc.ins("sw r7, 0(r5)")

		case tok == ";":
			// Epilogue: restore R7, return.
			fc.ins("lw r7, 0(r5)")
			fc.ins("addi r5, r5, 2")
			fc.ins("ret")
			fc.defineCtrl("wordskip")

		case tok == "if":
			fc.emitPop(1)
			body := fc.newLabel("if_body")
			ffalse := fc.newLabel("if_false")
			fc.ins("bne r1, r0, %s", body) // true → enter body
			fc.ins("jal %s", ffalse)       // false → jump to else/then
			fc.label(body)
			fc.pushCtrl("if", ffalse)

		case tok == "else":
			end := fc.newLabel("if_end")
			fc.ins("jal %s", end) // end of true body → skip else body
			fc.defineCtrl("if")   // if's false target lands here
			fc.pushCtrl("else", end)

		case tok == "then":
			fc.defineCtrl("if", "else")

		case tok == "begin":
			top := fc.newLabel("begin")
			fc.label(top)
			fc.pushCtrl("begin", top)

		case tok == "until":
			fc.emitPop(1)
			begin := fc.popCtrl("begin")
			exit := fc.newLabel("until_exit")
			fc.ins("bne r1, r0, %s", exit) // true → exit loop
			fc.ins("jal %s", begin.a)      // false → loop back
			fc.label(exit)

		case tok == "again":
			begin := fc.popCtrl("begin")
			fc.ins("jal %s", begin.a)

		case tok == "while":
			fc.emitPop(1)
			cont := fc.newLabel("while_cont")
			exit := fc.newLabel("while_exit")
			fc.ins("bne r1, r0, %s", cont) // true → continue loop body
			fc.ins("jal %s", exit)         // false → exit to after repeat
			fc.label(cont)
			fc.pushCtrl("while", exit)

		case tok == "repeat":
			while := fc.popCtrl("while")
			begin := fc.popCtrl("begin")
			fc.ins("jal %s", begin.a) // loop back
			fc.label(while.a)         // while's exit target

		case tok == "do":
			fc.ins("addi r5, r5, -2")
			fc.ins("sw r4, 0(r5)") // push limit
			fc.emitPop(4)
			fc.ins("addi r5, r5, -2")
			fc.ins("sw r4, 0(r5)") // push index
			fc.emitPop(4)
			top := fc.newLabel("do")
			fc.label(top)
			fc.pushCtrl("do", top)

		case tok == "loop":
			fc.ins("lw r1, 2(r5)") // index
			fc.ins("addi r1, r1, 1")
			fc.ins("sw r1, 2(r5)")
			fc.ins("lw r2, 0(r5)") // limit
			do := fc.popCtrl("do")
			exit := fc.newLabel("loop_exit")
			fc.ins("bge r1, r2, %s", exit) // index >= limit → exit
			fc.ins("jal %s", do.a)         // loop back
			fc.label(exit)
			fc.ins("addi r5, r5, 4") // pop index+limit

		case tok == "i":
			fc.emitPush(4)
			fc.ins("lw r4, 2(r5)")

		case tok == "j":
			fc.emitPush(4)
			fc.ins("lw r4, 6(r5)")

		default:
			fc.compileToken(tok)
		}
	}

	fc.ins("halt")
	if len(fc.errors) > 0 {
		return "", fmt.Errorf("forth: %s", strings.Join(fc.errors, "; "))
	}
	return fc.out.String(), nil
}

func (fc *ForthCompiler) compileToken(tok string) {
	if v, err := strconv.ParseInt(tok, 0, 32); err == nil {
		fc.emitPushConst(int(v))
		return
	}

	switch tok {
	case "+":
		fc.emitBinOp("add")
	case "-":
		fc.emitBinOp("sub")
	case "and":
		fc.emitBinOp("and")
	case "or":
		fc.emitBinOp("or")
	case "xor":
		fc.emitBinOp("xor")
	case "lshift":
		fc.emitBinOp("shl")
	case "rshift":
		fc.emitBinOp("shr")

	case "*":
		fc.emitPop(1)
		fc.ins("mul r4, r1, r4")

	case "f*":
		// Fixed-point 8.8 multiply: ( a b -- (a*b)>>8 )
		fc.emitPop(1)             // R1 = a, R4 = b
		fc.ins("mul r2, r1, r4")  // R2 = low16(a*b)
		fc.ins("mulh r3, r1, r4") // R3 = high16(a*b)
		fc.ins("addi r1, r0, 8")  // R1 = 8
		fc.ins("shr r2, r2, r1")  // R2 = low >> 8 (logical)
		fc.ins("shl r3, r3, r1")  // R3 = high << 8
		fc.ins("or r4, r2, r3")   // R4 = R2 | R3

	case "/":
		// Signed integer division: ( a b -- a/b )
		fc.emitPop(1)            // R1 = divisor, R4 = dividend
		fc.ins("addi r2, r0, 0") // R2 = 0 (initial remainder)
		fc.emitCall(fc.udivLabel)

	case "mod":
		fc.emitPop(1)
		fc.ins("addi r2, r0, 0")
		fc.emitCall(fc.udivLabel)
		fc.ins("add r4, r2, r0") // R4 = remainder (R2)

	case "f/":
		// Fixed-point SIGNED divide: ( a b -- (a*256)/b )
		// Handles negative dividend (divisor assumed positive)
		fc.emitPop(1) // R1 = divisor, R4 = dividend (a)

		// Save sign of a, make a positive.
		fc.ins("add r3, r0, r0") // R3 = 0 (sign flag)
		pos := fc.newLabel("fdiv_pos")
		fc.ins("bge r4, r0, %s", pos) // if R4 >= 0, skip negate
		fc.ins("sub r4, r0, r4")      // R4 = -R4
		fc.ins("addi r3, r0, -1")     // R3 = -1 (was negative)
		fc.label(pos)

		// Save sign flag to return stack (udiv clobbers R3).
		fc.ins("addi r5, r5, -2") // RSP -= 2
		fc.ins("sw r3, 0(r5)")    // mem[RSP] = sign flag

		// Prepare extended dividend: R2 = |a| >> 8, R4 = |a| << 8.
		fc.ins("addi r3, r0, 8")  // R3 = 8
		fc.ins("shr r2, r4, r3")  // R2 = |a| >> 8 (logical OK, a is positive)
		fc.ins("shl r4, r4, r3")  // R4 = |a| << 8
		fc.emitCall(fc.udivLabel) // R4 = quotient

		// Restore sign flag and negate if needed.
		fc.ins("lw r3, 0(r5)")   // R3 = sign flag
		fc.ins("addi r5, r5, 2") // RSP += 2
		neg := fc.newLabel("fdiv_neg")
		fc.ins("beq r3, r0, %s", neg) // if R3 == 0, skip negate
		fc.ins("sub r4, r0, r4")      // R4 = -R4
		fc.label(neg)

	case "*/":
		// ( a b c -- a*b/c ) with 32-bit intermediate
		fc.emitPop(1)             // R1 = c (divisor)
		fc.emitPop(2)             // R2 = b, R4 = a
		fc.ins("mulh r3, r4, r2") // R3 = high16(a*b)
		fc.ins("mul r4, r4, r2")  // R4 = low16(a*b)
		fc.ins("add r2, r3, r0")  // R2 = R3 (initial remainder = high word)
		fc.emitCall(fc.udivLabel)

	case "negate":
		fc.ins("sub r4, r0, r4")

	case "abs":
		end := fc.newLabel("abs")
		fc.ins("bge r4, r0, %s", end)
		fc.ins("sub r4, r0, r4")
		fc.label(end)

	case "dup":
		fc.emitPush(4)

	case "drop":
		fc.emitPop(4)

	case "swap":
		fc.ins("lw r1, 0(r6)")
		fc.ins("sw r4, 0(r6)")
		fc.ins("add r4, r1, r0")

	case "over":
		fc.emitPush(4)
		fc.ins("lw r4, 2(r6)")

	case "rot":
		fc.ins("lw r1, 0(r6)")
		fc.ins("lw r2, 2(r6)")
		fc.ins("sw r1, 2(r6)")
		fc.ins("sw r4, 0(r6)")
		fc.ins("add r4, r2, r0")

	case "nip":
		fc.ins("addi r6, r6, 2")

	case "2dup":
		fc.ins("lw r1, 0(r6)")    // R1 = second
		fc.ins("addi r6, r6, -2") // sp -= 2
		fc.ins("sw r1, 0(r6)")    // mem[sp] = second
		fc.ins("addi r6, r6, -2") // sp -= 2
		fc.ins("sw r4, 0(r6)")    // mem[sp] = TOS

	case "=":
		fc.emitCompare("beq") // equal

	case "<>":
		fc.emitCompare("bne") // not equal

	case "<":
		// ( a b -- flag ) flag = a < b. After pop: R1=b, R4=a.
		fc.emitPop(1)
		fc.boolBranch("blt", 4, 1)

	case ">":
		fc.emitPop(1)
		fc.boolBranch("blt", 1, 4) // b < a, i.e. a > b

	case "0=":
		fc.boolBranch("beq", 4, 0)

	case "0<":
		fc.boolBranch("blt", 4, 0)

	case "0>":
		fc.boolBranch("blt", 0, 4) // 0 < R4 means R4 > 0

	case "max":
		fc.emitPop(1) // R1=b, R4=a
		end := fc.newLabel("max")
		fc.ins("blt r1, r4, %s", end) // if b < a, keep a
		fc.ins("add r4, r1, r0")      // else R4=b
		fc.label(end)

	case "min":
		fc.emitPop(1)
		end := fc.newLabel("min")
		fc.ins("blt r4, r1, %s", end) // if a < b, keep a
		fc.ins("add r4, r1, r0")      // else R4=b
		fc.label(end)

	case "@":
		fc.ins("lw r4, 0(r4)")

	case "!":
		fc.emitPop(1)
		fc.ins("sw r4, 0(r1)") // mem[R1(addr)] = R4(val)
		fc.emitPop(4)

	case "c@":
		fc.ins("lb r4, 0(r4)")

	case "c!":
		fc.emitPop(1)
		fc.ins("sb r4, 0(r1)") // mem[R1(addr)] = byte(R4)
		fc.emitPop(4)

	case "emit":
		// Output a character by writing it to the UART data port (MMIO).
		fc.ins("add r2, r4, r0") // R2 = char (TOS)
		fc.emitLI(1, 0xF800)     // UART data register
		fc.ins("sb r2, 0(r1)")
		fc.emitPop(4)

	case "pixel":
		// ( color y x -- ) 8-bit write
		fc.ins("add r2, r4, r0") // R2 = x
		fc.emitPop(4)
		fc.ins("add r3, r4, r0") // R3 = y
		fc.emitPop(4)
		fc.emitLI(1, 6)
		fc.ins("shl r3, r3, r1")
		fc.emitLI(1, 0xF000)
		fc.ins("add r1, r1, r3")
		fc.ins("add r1, r1, r2")
		fc.ins("sb r4, 0(r1)")
		fc.emitPop(4)

	case "pixel16":
		// ( color16 y x -- ) 16-bit RGB555 write
		fc.ins("add r2, r4, r0") // R2 = x
		fc.emitPop(4)
		fc.ins("add r3, r4, r0") // R3 = y
		fc.emitPop(4)            // TOS = color16
		fc.emitLI(1, 7)
		fc.ins("shl r3, r3, r1") // R3 = y*128
		fc.ins("add r2, r2, r2") // R2 = x*2
		fc.emitLI(1, 0xF000)
		fc.ins("add r1, r1, r3")
		fc.ins("add r1, r1, r2")
		fc.ins("sw r4, 0(r1)") // 16-bit store
		fc.emitPop(4)

	case "fb-addr":
		fc.emitPop(1) // R1=x, TOS=y
		fc.emitLI(2, 6)
		fc.ins("shl r4, r4, r2") // TOS = y<<6
		fc.emitLI(2, 0xF000)
		fc.ins("add r4, r4, r2")
		fc.ins("add r4, r4, r1")

	case "halt":
		fc.ins("halt")

	default:
		if label, ok := fc.words[tok]; ok {
			fc.emitCall(label)
		} else {
			fc.errorf("unknown word: %s", tok)
		}
	}
}

// emitCompare: comparison via SUB then a branch-driven boolean.
func (fc *ForthCompiler) emitCompare(branch string) {
	fc.emitPop(1)            // R1 = old TOS, R4 = second
	fc.ins("sub r4, r4, r1") // R4 = R4 - R1
	fc.boolBranch(branch, 4, 0)
}

// boolBranch sets TOS to -1 (true) when the branch `mnem r<ra>, r<rb>` is taken,
// otherwise 0. Targets are nearby labels, well within branch range.
func (fc *ForthCompiler) boolBranch(mnem string, ra, rb int) {
	t := fc.newLabel("true")
	e := fc.newLabel("cmp_end")
	fc.ins("%s r%d, r%d, %s", mnem, ra, rb, t)
	fc.ins("addi r4, r0, 0")
	fc.ins("beq r0, r0, %s", e) // unconditional skip (no link)
	fc.label(t)
	fc.ins("addi r4, r0, -1")
	fc.label(e)
}

func (fc *ForthCompiler) ins(format string, args ...interface{}) {
	fmt.Fprintf(&fc.out, "\t"+format+"\n", args...)
}

func (fc *ForthCompiler) label(name string) {
	fmt.Fprintf(&fc.out, "%s:\n", name)
}

func (fc *ForthCompiler) newLabel(prefix string) string {
	fc.labelN++
	return fmt.Sprintf(".L%s_%d", prefix, fc.labelN)
}

// emitLI loads a 16-bit immediate into rd via the `li` pseudo-instruction.
func (fc *ForthCompiler) emitLI(rd, val int) {
	fc.ins("li r%d, %d", rd, val&0xFFFF)
}

func (fc *ForthCompiler) emitPush(srcReg int) {
	fc.ins("addi r6, r6, -2")
	fc.ins("sw r4, 0(r6)")
	if srcReg != 4 {
		fc.ins("add r4, r%d, r0", srcReg)
	}
}

func (fc *ForthCompiler) emitPop(dstReg int) {
	if dstReg != 4 {
		fc.ins("add r%d, r4, r0", dstReg)
	}
	fc.ins("lw r4, 0(r6)")
	fc.ins("addi r6, r6, 2")
}

func (fc *ForthCompiler) emitPushConst(val int) {
	fc.emitPush(4)
	fc.emitLI(4, val)
}

func (fc *ForthCompiler) emitBinOp(mnem string) {
	fc.emitPop(1)
	fc.ins("%s r4, r4, r1", mnem)
}

// emitCall calls a word or runtime subroutine via the `call` pseudo-instruction.
// The assembler emits a plain jal when in range, or a long-call trampoline
// otherwise. call clobbers R7 (saved by the callee prologue / dead in main flow)
// and, for long calls, R3 (also clobbered by callees, never live across a call).
func (fc *ForthCompiler) emitCall(label string) {
	fc.ins("call %s", label)
}

func (fc *ForthCompiler) emitBootstrap() {
	fc.emitLI(6, 0xEF00)     // SP
	fc.emitLI(5, 0xDF00)     // RSP
	fc.ins("addi r4, r0, 0") // TOS = 0
	main := fc.newLabel("main")
	fc.ins("jal %s", main) // skip over runtime subroutines

	// === Runtime: _udiv ===
	// Input: R4 = dividend, R1 = divisor, R2 = initial remainder (0 for normal div)
	// Output: R4 = quotient, R2 = remainder. Clobbers: R3.
	fc.udivLabel = fc.newLabel("udiv")
	fc.label(fc.udivLabel)
	fc.ins("addi r5, r5, -2") // rsp -= 2
	fc.ins("sw r7, 0(r5)")    // save R7
	fc.ins("addi r3, r0, 16") // counter = 16
	loop := fc.newLabel("udiv_loop")
	s1 := fc.newLabel("udiv_skip1")
	s2 := fc.newLabel("udiv_skip2")
	fc.label(loop)
	fc.ins("add r2, r2, r2")     // rem <<= 1
	fc.ins("bge r4, r0, %s", s1) // if dividend >= 0, skip
	fc.ins("addi r2, r2, 1")     // rem |= 1
	fc.label(s1)
	fc.ins("add r4, r4, r4")     // dividend <<= 1
	fc.ins("blt r2, r1, %s", s2) // if rem < div, skip sub
	fc.ins("sub r2, r2, r1")     // rem -= div
	fc.ins("addi r4, r4, 1")     // quotient |= 1
	fc.label(s2)
	fc.ins("addi r3, r3, -1")      // counter--
	fc.ins("blt r0, r3, %s", loop) // if 0 < counter, loop back
	fc.ins("lw r7, 0(r5)")         // restore R7
	fc.ins("addi r5, r5, 2")       // rsp += 2
	fc.ins("ret")                  // jalr R7
	fc.label(main)
}

func (fc *ForthCompiler) pushCtrl(kind, label string) {
	fc.ctrlStack = append(fc.ctrlStack, ctrlEntry{kind, label})
}

// defineCtrl pops the nearest control entry matching one of kinds and emits its
// label at the current position.
func (fc *ForthCompiler) defineCtrl(kinds ...string) {
	for i := len(fc.ctrlStack) - 1; i >= 0; i-- {
		for _, k := range kinds {
			if fc.ctrlStack[i].kind == k {
				entry := fc.ctrlStack[i]
				fc.ctrlStack = append(fc.ctrlStack[:i], fc.ctrlStack[i+1:]...)
				fc.label(entry.a)
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
		tokens = append(tokens, strings.Fields(line)...)
	}
	return tokens
}

func (fc *ForthCompiler) errorf(format string, args ...interface{}) {
	fc.errors = append(fc.errors, fmt.Sprintf(format, args...))
}
