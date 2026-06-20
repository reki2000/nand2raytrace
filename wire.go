package nand16

// Wire represents a single digital signal line.
type Wire struct {
	Val bool
}

// NewWire creates a single wire.
func NewWire() *Wire {
	return &Wire{}
}

// Bus is a group of wires.
type Bus []*Wire

// NewBus creates n wires.
func NewBus(n int) Bus {
	b := make(Bus, n)
	for i := range b {
		b[i] = NewWire()
	}
	return b
}

// SetVal writes an integer value to the bus (LSB-first).
func (b Bus) SetVal(v int) {
	for i := range b {
		b[i].Val = (v>>i)&1 == 1
	}
}

// GetVal reads the bus as an unsigned integer.
func (b Bus) GetVal() int {
	v := 0
	for i := range b {
		if b[i].Val {
			v |= 1 << i
		}
	}
	return v
}

// GetSigned reads the bus as a signed integer (two's complement).
func (b Bus) GetSigned() int {
	v := b.GetVal()
	n := len(b)
	if n > 0 && v >= (1<<(n-1)) {
		v -= 1 << n
	}
	return v
}
