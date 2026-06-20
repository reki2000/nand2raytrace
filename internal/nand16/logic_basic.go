package nand16

// NOT: 1 NAND
func NOT(m *Module, a *Wire) *Wire {
	return m.NAND(a, a)
}

// AND: 2 NAND
func AND(m *Module, a, b *Wire) *Wire {
	return NOT(m, m.NAND(a, b))
}

// OR: 3 NAND
func OR(m *Module, a, b *Wire) *Wire {
	return m.NAND(NOT(m, a), NOT(m, b))
}

// NOR: 4 NAND
func NOR(m *Module, a, b *Wire) *Wire {
	return NOT(m, OR(m, a, b))
}

// XOR: 4 NAND (optimized)
func XOR(m *Module, a, b *Wire) *Wire {
	n := m.NAND(a, b)
	return m.NAND(m.NAND(a, n), m.NAND(n, b))
}

// XNOR: XOR + NOT
func XNOR(m *Module, a, b *Wire) *Wire {
	return NOT(m, XOR(m, a, b))
}

// AND3: 3-input AND
func AND3(m *Module, a, b, c *Wire) *Wire {
	return AND(m, AND(m, a, b), c)
}

// OR3: 3-input OR
func OR3(m *Module, a, b, c *Wire) *Wire {
	return OR(m, OR(m, a, b), c)
}

// Mux2: sel=0 -> a, sel=1 -> b
func Mux2(m *Module, sel, a, b *Wire) *Wire {
	nsel := NOT(m, sel)
	return OR(m, AND(m, nsel, a), AND(m, sel, b))
}

// Mux4: sel[0..1] selects among a,b,c,d
func Mux4(m *Module, sel0, sel1, a, b, c, d *Wire) *Wire {
	lo := Mux2(m, sel0, a, b)
	hi := Mux2(m, sel0, c, d)
	return Mux2(m, sel1, lo, hi)
}

// Mux2Bus: bus-width mux
func Mux2Bus(m *Module, sel *Wire, a, b, out Bus) {
	for i := range out {
		out[i] = Mux2(m, sel, a[i], b[i])
	}
}

// Mux4Bus: bus-width 4-input mux
func Mux4Bus(m *Module, sel0, sel1 *Wire, a, b, c, d, out Bus) {
	for i := range out {
		out[i] = Mux4(m, sel0, sel1, a[i], b[i], c[i], d[i])
	}
}

// Mux8Bus: bus-width 8-input mux using sel[0..2]
func Mux8Bus(m *Module, sel Bus, a, b, c, d, e, f, g, h Bus) Bus {
	n := len(a)
	lo := make(Bus, n)
	hi := make(Bus, n)
	out := make(Bus, n)
	Mux4Bus(m, sel[0], sel[1], a, b, c, d, lo)
	Mux4Bus(m, sel[0], sel[1], e, f, g, h, hi)
	Mux2Bus(m, sel[2], lo, hi, out)
	return out
}

// Buffer: just passes through (0 gates, wire aliasing)
func Buffer(m *Module, a *Wire) *Wire {
	return NOT(m, NOT(m, a))
}
