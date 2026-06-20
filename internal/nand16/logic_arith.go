package nand16

// HalfAdder: sum = a XOR b, carry = a AND b
func HalfAdder(m *Module, a, b *Wire) (sum, carry *Wire) {
	sum = XOR(m, a, b)
	carry = AND(m, a, b)
	return
}

// FullAdder: sum, carry from a + b + cin
func FullAdder(m *Module, a, b, cin *Wire) (sum, cout *Wire) {
	s1, c1 := HalfAdder(m, a, b)
	sum, c2 := HalfAdder(m, s1, cin)
	cout = OR(m, c1, c2)
	return
}

// Adder16: 16-bit ripple carry adder
func Adder16(m *Module, a, b Bus, cin *Wire) (sum Bus, cout *Wire) {
	sum = NewBus(16)
	carry := cin
	for i := 0; i < 16; i++ {
		sum[i], carry = FullAdder(m, a[i], b[i], carry)
	}
	cout = carry
	return
}

// Subtractor16: a - b = a + ~b + 1
func Subtractor16(m *Module, a, b Bus) (diff Bus, borrow *Wire) {
	nb := make(Bus, 16)
	for i := 0; i < 16; i++ {
		nb[i] = NOT(m, b[i])
	}
	one := NewWire()
	one.Val = true // constant 1
	diff, cout := Adder16(m, a, nb, one)
	borrow = NOT(m, cout)
	return
}

// AddSub16: sub=0 -> add, sub=1 -> subtract. Returns result + cout.
func AddSub16(m *Module, a, b Bus, sub *Wire) (result Bus, cout *Wire) {
	xb := make(Bus, 16)
	for i := 0; i < 16; i++ {
		xb[i] = XOR(m, b[i], sub) // conditionally invert b
	}
	result, cout = Adder16(m, a, xb, sub) // cin=sub acts as +1 for 2's complement
	return
}

// Zero16: returns 1 if all bits are 0
func Zero16(m *Module, a Bus) *Wire {
	// OR-reduce all bits, then NOT
	r := OR(m, a[0], a[1])
	for i := 2; i < 16; i++ {
		r = OR(m, r, a[i])
	}
	return NOT(m, r)
}

// Comparator16Signed: returns (eq, lt) where lt = (a < b) signed
func Comparator16Signed(m *Module, a, b Bus) (eq, lt *Wire) {
	diff, _ := Subtractor16(m, a, b)
	eq = Zero16(m, diff)
	// For signed comparison: lt = sign(diff) XOR overflow
	// Simplified: if same sign, lt = diff[15]; if diff sign, lt = a[15]
	sameSign := XNOR(m, a[15], b[15])
	lt = Mux2(m, sameSign, a[15], diff[15])
	return
}

// Incrementer16: a + 1
func Incrementer16(m *Module, a Bus) (result Bus, cout *Wire) {
	one := NewWire()
	one.Val = true
	result, cout = Adder16(m, a, NewBusConst(16, 0), one)
	return
}

// NewBusConst creates a bus with a constant value (using constant wires).
func NewBusConst(n, val int) Bus {
	b := NewBus(n)
	b.SetVal(val)
	return b
}
