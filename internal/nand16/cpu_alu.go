package nand16

// ALU16 performs one of 8 operations selected by op[2:0].
// op: 000=ADD, 001=SUB, 010=AND, 011=OR, 100=XOR, 101=SHL, 110=SHR, 111=SRA
// Returns result bus, zero flag, negative flag.
func ALU16(m *Module, a, b Bus, op Bus) (result Bus, zero, neg *Wire) {
	// Compute all operations in parallel, then mux the result
	addResult, _ := AddSub16(m, a, b, NewWire()) // ADD: sub=0
	subResult, _ := AddSub16(m, a, b, constWire(true))  // SUB: sub=1
	andResult := busAND(m, a, b)
	orResult := busOR(m, a, b)
	xorResult := busXOR(m, a, b)
	shlResult := ShiftLeft16(m, a, b[0:4])  // shift amount = b[3:0]
	shrResult := ShiftRight16(m, a, b[0:4])
	sraResult := ShiftRightArith16(m, a, b[0:4])

	// 8-to-1 mux per bit
	result = Mux8Bus(m, op,
		addResult, subResult, andResult, orResult,
		xorResult, shlResult, shrResult, sraResult)

	zero = Zero16(m, result)
	neg = result[15] // sign bit
	return
}

func constWire(v bool) *Wire {
	w := NewWire()
	w.Val = v
	return w
}

func busAND(m *Module, a, b Bus) Bus {
	out := make(Bus, len(a))
	for i := range a {
		out[i] = AND(m, a[i], b[i])
	}
	return out
}

func busOR(m *Module, a, b Bus) Bus {
	out := make(Bus, len(a))
	for i := range a {
		out[i] = OR(m, a[i], b[i])
	}
	return out
}

func busXOR(m *Module, a, b Bus) Bus {
	out := make(Bus, len(a))
	for i := range a {
		out[i] = XOR(m, a[i], b[i])
	}
	return out
}
