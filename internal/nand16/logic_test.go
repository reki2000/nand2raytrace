package nand16

import (
	"math/rand"
	"testing"
)

func TestXOR(t *testing.T) {
	m := NewModule("test")
	a, b := NewWire(), NewWire()
	out := XOR(m, a, b)
	sim := NewSimulator(m)

	for _, tt := range []struct{ a, b, want bool }{
		{false, false, false}, {false, true, true},
		{true, false, true}, {true, true, false},
	} {
		a.Val = tt.a
		b.Val = tt.b
		sim.Settle()
		if out.Val != tt.want {
			t.Errorf("XOR(%v,%v)=%v, want %v", tt.a, tt.b, out.Val, tt.want)
		}
	}
}

func TestMux2(t *testing.T) {
	m := NewModule("test")
	sel, a, b := NewWire(), NewWire(), NewWire()
	out := Mux2(m, sel, a, b)
	sim := NewSimulator(m)

	a.Val = true
	b.Val = false
	sel.Val = false
	sim.Settle()
	if out.Val != true {
		t.Error("Mux2(sel=0) should select a")
	}
	sel.Val = true
	sim.Settle()
	if out.Val != false {
		t.Error("Mux2(sel=1) should select b")
	}
}

func TestFullAdder(t *testing.T) {
	m := NewModule("test")
	a, b, cin := NewWire(), NewWire(), NewWire()
	sum, cout := FullAdder(m, a, b, cin)
	sim := NewSimulator(m)

	for i := 0; i < 8; i++ {
		av := (i >> 0) & 1
		bv := (i >> 1) & 1
		cv := (i >> 2) & 1
		a.Val = av == 1
		b.Val = bv == 1
		cin.Val = cv == 1
		sim.Settle()
		s := av + bv + cv
		wantSum := (s & 1) == 1
		wantCout := (s >> 1) == 1
		if sum.Val != wantSum || cout.Val != wantCout {
			t.Errorf("FA(%d,%d,%d): sum=%v cout=%v, want sum=%v cout=%v",
				av, bv, cv, sum.Val, cout.Val, wantSum, wantCout)
		}
	}
}

func TestAdder16(t *testing.T) {
	m := NewModule("test")
	a, b := NewBus(16), NewBus(16)
	cin := NewWire()
	sum, _ := Adder16(m, a, b, cin)
	sim := NewSimulator(m)

	rng := rand.New(rand.NewSource(42))
	for i := 0; i < 100; i++ {
		av := rng.Intn(65536)
		bv := rng.Intn(65536)
		a.SetVal(av)
		b.SetVal(bv)
		cin.Val = false
		sim.Settle()
		want := (av + bv) & 0xFFFF
		got := sum.GetVal()
		if got != want {
			t.Errorf("Add16(%d, %d) = %d, want %d", av, bv, got, want)
		}
	}
}

func TestSubtractor16(t *testing.T) {
	m := NewModule("test")
	a, b := NewBus(16), NewBus(16)
	diff, _ := Subtractor16(m, a, b)
	sim := NewSimulator(m)

	rng := rand.New(rand.NewSource(43))
	for i := 0; i < 100; i++ {
		av := rng.Intn(65536)
		bv := rng.Intn(65536)
		a.SetVal(av)
		b.SetVal(bv)
		sim.Settle()
		want := (av - bv + 65536) & 0xFFFF
		got := diff.GetVal()
		if got != want {
			t.Errorf("Sub16(%d, %d) = %d, want %d", av, bv, got, want)
		}
	}
}

func TestShiftLeft16(t *testing.T) {
	m := NewModule("test")
	a := NewBus(16)
	amt := NewBus(4)
	out := ShiftLeft16(m, a, amt)
	sim := NewSimulator(m)

	tests := []struct{ val, shift, want int }{
		{1, 0, 1}, {1, 1, 2}, {1, 4, 16}, {1, 15, 0x8000},
		{0xFF, 8, 0xFF00}, {0xFFFF, 1, 0xFFFE},
	}
	for _, tt := range tests {
		a.SetVal(tt.val)
		amt.SetVal(tt.shift)
		sim.Settle()
		got := out.GetVal()
		want := (tt.val << tt.shift) & 0xFFFF
		if got != want {
			t.Errorf("SHL(%d, %d) = %d, want %d", tt.val, tt.shift, got, want)
		}
	}
}

func TestShiftRight16(t *testing.T) {
	m := NewModule("test")
	a := NewBus(16)
	amt := NewBus(4)
	out := ShiftRight16(m, a, amt)
	sim := NewSimulator(m)

	tests := []struct{ val, shift, want int }{
		{0x8000, 1, 0x4000}, {0xFF00, 8, 0xFF}, {1, 1, 0},
	}
	for _, tt := range tests {
		a.SetVal(tt.val)
		amt.SetVal(tt.shift)
		sim.Settle()
		got := out.GetVal()
		if got != tt.want {
			t.Errorf("SHR(%X, %d) = %X, want %X", tt.val, tt.shift, got, tt.want)
		}
	}
}

func TestShiftRightArith16(t *testing.T) {
	m := NewModule("test")
	a := NewBus(16)
	amt := NewBus(4)
	out := ShiftRightArith16(m, a, amt)
	sim := NewSimulator(m)

	// -2 >> 1 should be -1
	a.SetVal(0xFFFE) // -2
	amt.SetVal(1)
	sim.Settle()
	if got := out.GetSigned(); got != -1 {
		t.Errorf("SRA(-2, 1) = %d, want -1", got)
	}

	// -256 >> 4 = -16
	a.SetVal(uint16ToInt(-256))
	amt.SetVal(4)
	sim.Settle()
	if got := out.GetSigned(); got != -16 {
		t.Errorf("SRA(-256, 4) = %d, want -16", got)
	}
}

func TestZero16(t *testing.T) {
	m := NewModule("test")
	a := NewBus(16)
	z := Zero16(m, a)
	sim := NewSimulator(m)

	a.SetVal(0)
	sim.Settle()
	if !z.Val {
		t.Error("Zero16(0) should be true")
	}
	a.SetVal(1)
	sim.Settle()
	if z.Val {
		t.Error("Zero16(1) should be false")
	}
}

func TestComparator16Signed(t *testing.T) {
	m := NewModule("test")
	a, b := NewBus(16), NewBus(16)
	eq, lt := Comparator16Signed(m, a, b)
	sim := NewSimulator(m)

	tests := []struct {
		a, b          int
		wantEq, wantLt bool
	}{
		{0, 0, true, false},
		{1, 2, false, true},
		{2, 1, false, false},
		{-1, 0, false, true},
		{0, -1, false, false},
		{-100, -50, false, true},
	}
	for _, tt := range tests {
		a.SetVal(uint16ToInt(tt.a))
		b.SetVal(uint16ToInt(tt.b))
		sim.Settle()
		if eq.Val != tt.wantEq || lt.Val != tt.wantLt {
			t.Errorf("Cmp(%d,%d): eq=%v lt=%v, want eq=%v lt=%v",
				tt.a, tt.b, eq.Val, lt.Val, tt.wantEq, tt.wantLt)
		}
	}
}

func uint16ToInt(v int) int {
	return v & 0xFFFF
}
