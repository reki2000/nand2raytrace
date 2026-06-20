package asmc

import "nand16/internal/op"

// Signed offset ranges (in words) for the branch and jump immediate fields.
const (
	branchMin = -32 // 6-bit signed
	branchMax = 31
	jumpMin   = -2048 // 12-bit signed
	jumpMax   = 2047
)

// wordOffset returns the signed distance, in instruction words, from byte
// address `from` to byte address `to`. Instructions are 2 bytes wide.
func wordOffset(from, to int) int { return (to - from) / 2 }

// branchInRange reports whether a word offset fits the 6-bit branch field.
func branchInRange(off int) bool { return off >= branchMin && off <= branchMax }

// jumpInRange reports whether a word offset fits the 12-bit jump field.
func jumpInRange(off int) bool { return off >= jumpMin && off <= jumpMax }

// appendWordLE appends a 16-bit word to b in little-endian byte order.
func appendWordLE(b []byte, w uint16) []byte { return append(b, byte(w), byte(w>>8)) }

// longCallWords is the fixed instruction count of the long-call trampoline
// emitted by longCall. It is constant so branch relaxation has a deterministic
// size for an out-of-range `call`.
const longCallWords = 17

// longCall lowers an out-of-JAL-range `call label` into a trampoline that loads
// the 16-bit target address and jumps via JALR. It clobbers only R3 (which the
// callee may clobber anyway, and which is never a live argument across a call),
// shifting by repeated self-add so no second scratch register is needed.
func longCall(addr int) []uint16 {
	v := addr & 0xFFFF
	nib := [4]int{(v >> 12) & 0xF, (v >> 8) & 0xF, (v >> 4) & 0xF, v & 0xF}
	out := make([]uint16, 0, longCallWords)
	out = append(out, op.EncodeI(op.OpADDI, 3, 0, nib[0])) // r3 = top nibble
	for _, n := range nib[1:] {
		for k := 0; k < 4; k++ {
			out = append(out, op.EncodeR(op.OpALU, 3, 3, 3, 0)) // r3 += r3 (<<1)
		}
		out = append(out, op.EncodeI(op.OpORI, 3, 3, n)) // r3 |= nibble
	}
	out = append(out, op.EncodeR(op.OpSYSTEM, 0, 3, 0, 0)) // jalr r3
	return out
}

// expandLI lowers the "load 16-bit immediate" pseudo-instruction (li rd, imm)
// into a concrete instruction sequence. It chooses the shortest form:
//   - 1 ADDI when the value fits the signed 6-bit immediate (-32..31)
//   - 2 ADDI for values just outside that range (-63..62)
//   - 1 LUI when only the upper bits are set (imm == upper<<10)
//   - otherwise a shift-based sequence: load the high byte, shift left 8, then
//     add the low byte. The general case uses R3 as scratch for the shift
//     amount, so `li r3, <large>` is not supported.
//
// The same expansion is used for instruction sizing (pass 1) and emission
// (pass 2) so label addresses stay consistent.
func expandLI(rd, val int) []uint16 {
	val = val & 0xFFFF
	signed := int16(uint16(val))
	var out []uint16

	// Fits in imm6 (-32..31): 1 instruction.
	if signed >= -32 && signed <= 31 {
		return append(out, op.EncodeI(op.OpADDI, rd, 0, int(signed)))
	}
	// Just outside imm6 (-63..62): 2 ADDI.
	if signed >= 32 && signed <= 62 {
		out = append(out, op.EncodeI(op.OpADDI, rd, 0, 31))
		out = append(out, op.EncodeI(op.OpADDI, rd, rd, int(signed-31)))
		return out
	}
	if signed >= -63 && signed <= -33 {
		out = append(out, op.EncodeI(op.OpADDI, rd, 0, -32))
		out = append(out, op.EncodeI(op.OpADDI, rd, rd, int(signed+32)))
		return out
	}
	// LUI shortcut: only upper bits set.
	upper := (val >> 10) & 0x3F
	lower := val - (upper << 10)
	if upper != 0 && lower == 0 {
		return append(out, op.EncodeI(op.OpLUI, rd, 0, upper))
	}
	// General: val = (hi << 8) + lo.
	hi := (val >> 8) & 0xFF
	lo := val & 0xFF
	hiSigned := int(int8(byte(hi)))

	out = append(out, expandLoadSmall(rd, hiSigned)...)
	out = append(out, op.EncodeI(op.OpADDI, 3, 0, 8))
	out = append(out, op.EncodeR(op.OpALU, rd, rd, 3, 5)) // rd <<= 8
	out = append(out, expandAdditiveChain(rd, int(lo))...)
	return out
}

// expandLoadSmall loads a value in -128..127 into rd using minimal ADDI.
func expandLoadSmall(rd, val int) []uint16 {
	var out []uint16
	switch {
	case val >= -32 && val <= 31:
		out = append(out, op.EncodeI(op.OpADDI, rd, 0, val))
	case val >= 32 && val <= 62:
		out = append(out, op.EncodeI(op.OpADDI, rd, 0, 31))
		out = append(out, op.EncodeI(op.OpADDI, rd, rd, val-31))
	case val >= -63 && val <= -33:
		out = append(out, op.EncodeI(op.OpADDI, rd, 0, -32))
		out = append(out, op.EncodeI(op.OpADDI, rd, rd, val+32))
	case val >= 63 && val <= 127:
		out = append(out, op.EncodeI(op.OpADDI, rd, 0, 31))
		out = append(out, op.EncodeI(op.OpADDI, rd, rd, 31))
		out = append(out, op.EncodeI(op.OpADDI, rd, rd, val-62))
	default: // -128..-64
		out = append(out, op.EncodeI(op.OpADDI, rd, 0, -32))
		out = append(out, op.EncodeI(op.OpADDI, rd, rd, -32))
		out = append(out, op.EncodeI(op.OpADDI, rd, rd, val+64))
	}
	return out
}

// expandAdditiveChain adds a non-negative value to rd using an ADDI chain.
func expandAdditiveChain(rd, val int) []uint16 {
	var out []uint16
	for val > 31 {
		out = append(out, op.EncodeI(op.OpADDI, rd, rd, 31))
		val -= 31
	}
	if val > 0 {
		out = append(out, op.EncodeI(op.OpADDI, rd, rd, val))
	}
	return out
}
