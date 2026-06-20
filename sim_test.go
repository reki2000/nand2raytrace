package nand16

import "testing"

func TestNAND(t *testing.T) {
	m := NewModule("test")
	a, b := NewWire(), NewWire()
	out := m.NAND(a, b)
	sim := NewSimulator(m)

	tests := []struct {
		a, b bool
		want bool
	}{
		{false, false, true},
		{false, true, true},
		{true, false, true},
		{true, true, false},
	}
	for _, tt := range tests {
		a.Val = tt.a
		b.Val = tt.b
		sim.Settle()
		if out.Val != tt.want {
			t.Errorf("NAND(%v,%v)=%v, want %v", tt.a, tt.b, out.Val, tt.want)
		}
	}
}

func TestNOT(t *testing.T) {
	m := NewModule("test")
	a := NewWire()
	// NOT = NAND(a, a)
	out := m.NAND(a, a)
	sim := NewSimulator(m)

	a.Val = false
	sim.Settle()
	if out.Val != true {
		t.Error("NOT(0) should be 1")
	}
	a.Val = true
	sim.Settle()
	if out.Val != false {
		t.Error("NOT(1) should be 0")
	}
}

func TestAND(t *testing.T) {
	m := NewModule("test")
	a, b := NewWire(), NewWire()
	// AND = NOT(NAND(a,b))
	nab := m.NAND(a, b)
	out := m.NAND(nab, nab) // NOT
	sim := NewSimulator(m)

	tests := []struct {
		a, b bool
		want bool
	}{
		{false, false, false},
		{false, true, false},
		{true, false, false},
		{true, true, true},
	}
	for _, tt := range tests {
		a.Val = tt.a
		b.Val = tt.b
		sim.Settle()
		if out.Val != tt.want {
			t.Errorf("AND(%v,%v)=%v, want %v", tt.a, tt.b, out.Val, tt.want)
		}
	}
}

func TestOR(t *testing.T) {
	m := NewModule("test")
	a, b := NewWire(), NewWire()
	// OR = NAND(NOT(a), NOT(b))
	na := m.NAND(a, a)
	nb := m.NAND(b, b)
	out := m.NAND(na, nb)
	sim := NewSimulator(m)

	tests := []struct {
		a, b bool
		want bool
	}{
		{false, false, false},
		{false, true, true},
		{true, false, true},
		{true, true, true},
	}
	for _, tt := range tests {
		a.Val = tt.a
		b.Val = tt.b
		sim.Settle()
		if out.Val != tt.want {
			t.Errorf("OR(%v,%v)=%v, want %v", tt.a, tt.b, out.Val, tt.want)
		}
	}
}

func TestDFF(t *testing.T) {
	m := NewModule("test")
	d := NewWire()
	q, qn := m.DFF(d)
	sim := NewSimulator(m)

	// Initially Q=0
	if q.Val != false {
		t.Error("DFF initial Q should be 0")
	}
	if qn.Val != true {
		t.Error("DFF initial Qn should be 1")
	}

	// Set D=1, cycle
	d.Val = true
	sim.Cycle()
	if q.Val != true {
		t.Error("After D=1 cycle, Q should be 1")
	}
	if qn.Val != false {
		t.Error("After D=1 cycle, Qn should be 0")
	}

	// Set D=0, cycle
	d.Val = false
	sim.Cycle()
	if q.Val != false {
		t.Error("After D=0 cycle, Q should be 0")
	}
}

func TestBus(t *testing.T) {
	b := NewBus(16)
	b.SetVal(0xCAFE)
	if got := b.GetVal(); got != 0xCAFE {
		t.Errorf("Bus.GetVal()=0x%X, want 0xCAFE", got)
	}

	b.SetVal(0xFFFF)
	if got := b.GetSigned(); got != -1 {
		t.Errorf("Bus.GetSigned()=%d, want -1", got)
	}

	b.SetVal(0x8000)
	if got := b.GetSigned(); got != -32768 {
		t.Errorf("Bus.GetSigned()=%d, want -32768", got)
	}
}
