package nand16

// mul16 is a combinational 16x16 signed multiplier producing the low and high
// 16-bit halves of the 32-bit product.
//
// It first forms the unsigned 32-bit product with a shift-and-add array, then
// applies the two's-complement sign correction: for signed operands,
//
//	a_s * b_s = a_u * b_u - a15*(b_u<<16) - b15*(a_u<<16)   (mod 2^32)
//
// The correction terms only affect the high word, so the low half equals the
// unsigned low half and the high half is corrected with two subtractions.
func mul16(m *Module, a, b Bus) (low, high Bus) {
	zero := NewWire()

	acc := make(Bus, 32)
	for i := range acc {
		acc[i] = zero
	}
	for j := 0; j < 16; j++ {
		// Partial product (a AND b[j]) shifted left by j.
		pp := make(Bus, 32)
		for i := range pp {
			pp[i] = zero
		}
		for k := 0; k < 16; k++ {
			pp[j+k] = AND(m, a[k], b[j])
		}
		acc = add32(m, acc, pp)
	}

	low = acc[0:16]
	highRaw := acc[16:32]

	// Sign correction on the high word.
	bMasked := maskBus(m, b, a[15]) // a15 ? b : 0
	aMasked := maskBus(m, a, b[15]) // b15 ? a : 0
	d1, _ := Subtractor16(m, highRaw, bMasked)
	high, _ = Subtractor16(m, d1, aMasked)
	return low, high
}

// add32 adds two 32-bit buses (ripple carry across two 16-bit adders).
func add32(m *Module, a, b Bus) Bus {
	lo, carry := Adder16(m, a[0:16], b[0:16], NewWire())
	hi, _ := Adder16(m, a[16:32], b[16:32], carry)
	out := make(Bus, 32)
	copy(out[0:16], lo)
	copy(out[16:32], hi)
	return out
}

// maskBus returns a AND en (broadcast), i.e. a when en=1 else all zeros.
func maskBus(m *Module, a Bus, en *Wire) Bus {
	out := make(Bus, len(a))
	for i := range a {
		out[i] = AND(m, a[i], en)
	}
	return out
}
