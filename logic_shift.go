package nand16

// ShiftLeft16: shifts a left by amount (4-bit), fills with 0
func ShiftLeft16(m *Module, a Bus, amount Bus) Bus {
	// 4-stage barrel shifter: shift by 1, 2, 4, 8
	zero := NewWire() // constant 0
	cur := make(Bus, 16)
	copy(cur, a)

	// Shift by 1
	s1 := make(Bus, 16)
	s1[0] = zero
	for i := 1; i < 16; i++ {
		s1[i] = cur[i-1]
	}
	next := make(Bus, 16)
	Mux2Bus(m, amount[0], cur, s1, next)
	cur = next

	// Shift by 2
	s2 := make(Bus, 16)
	s2[0] = zero
	s2[1] = zero
	for i := 2; i < 16; i++ {
		s2[i] = cur[i-2]
	}
	next = make(Bus, 16)
	Mux2Bus(m, amount[1], cur, s2, next)
	cur = next

	// Shift by 4
	s4 := make(Bus, 16)
	for i := 0; i < 4; i++ {
		s4[i] = zero
	}
	for i := 4; i < 16; i++ {
		s4[i] = cur[i-4]
	}
	next = make(Bus, 16)
	Mux2Bus(m, amount[2], cur, s4, next)
	cur = next

	// Shift by 8
	s8 := make(Bus, 16)
	for i := 0; i < 8; i++ {
		s8[i] = zero
	}
	for i := 8; i < 16; i++ {
		s8[i] = cur[i-8]
	}
	next = make(Bus, 16)
	Mux2Bus(m, amount[3], cur, s8, next)
	return next
}

// ShiftRight16: logical right shift (fills with 0)
func ShiftRight16(m *Module, a Bus, amount Bus) Bus {
	zero := NewWire()
	cur := make(Bus, 16)
	copy(cur, a)

	// Shift by 1
	s1 := make(Bus, 16)
	for i := 0; i < 15; i++ {
		s1[i] = cur[i+1]
	}
	s1[15] = zero
	next := make(Bus, 16)
	Mux2Bus(m, amount[0], cur, s1, next)
	cur = next

	// Shift by 2
	s2 := make(Bus, 16)
	for i := 0; i < 14; i++ {
		s2[i] = cur[i+2]
	}
	s2[14] = zero
	s2[15] = zero
	next = make(Bus, 16)
	Mux2Bus(m, amount[1], cur, s2, next)
	cur = next

	// Shift by 4
	s4 := make(Bus, 16)
	for i := 0; i < 12; i++ {
		s4[i] = cur[i+4]
	}
	for i := 12; i < 16; i++ {
		s4[i] = zero
	}
	next = make(Bus, 16)
	Mux2Bus(m, amount[2], cur, s4, next)
	cur = next

	// Shift by 8
	s8 := make(Bus, 16)
	for i := 0; i < 8; i++ {
		s8[i] = cur[i+8]
	}
	for i := 8; i < 16; i++ {
		s8[i] = zero
	}
	next = make(Bus, 16)
	Mux2Bus(m, amount[3], cur, s8, next)
	return next
}

// ShiftRightArith16: arithmetic right shift (fills with sign bit)
func ShiftRightArith16(m *Module, a Bus, amount Bus) Bus {
	sign := a[15]
	cur := make(Bus, 16)
	copy(cur, a)

	// Shift by 1
	s1 := make(Bus, 16)
	for i := 0; i < 15; i++ {
		s1[i] = cur[i+1]
	}
	s1[15] = sign
	next := make(Bus, 16)
	Mux2Bus(m, amount[0], cur, s1, next)
	cur = next

	// Shift by 2
	s2 := make(Bus, 16)
	for i := 0; i < 14; i++ {
		s2[i] = cur[i+2]
	}
	s2[14] = sign
	s2[15] = sign
	next = make(Bus, 16)
	Mux2Bus(m, amount[1], cur, s2, next)
	cur = next

	// Shift by 4
	s4 := make(Bus, 16)
	for i := 0; i < 12; i++ {
		s4[i] = cur[i+4]
	}
	for i := 12; i < 16; i++ {
		s4[i] = sign
	}
	next = make(Bus, 16)
	Mux2Bus(m, amount[2], cur, s4, next)
	cur = next

	// Shift by 8
	s8 := make(Bus, 16)
	for i := 0; i < 8; i++ {
		s8[i] = cur[i+8]
	}
	for i := 8; i < 16; i++ {
		s8[i] = sign
	}
	next = make(Bus, 16)
	Mux2Bus(m, amount[3], cur, s8, next)
	return next
}
