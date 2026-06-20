package nand16

import (
	"bytes"
	"testing"
)

func runForth(source string) (*System, string) {
	fc := NewForthCompiler()
	code, err := fc.Compile(source, 0)
	if err != nil {
		return nil, "COMPILE ERROR: " + err.Error()
	}

	sys := NewSystem()
	sys.LoadRAM(0, code)
	sys.CPU.PC = 0
	var out bytes.Buffer
	sys.Stdout = &out
	sys.Run(100000)
	return sys, out.String()
}

func TestForth_Arithmetic(t *testing.T) {
	// 3 + 4 = 7, emit as ASCII
	src := `3 4 + 48 + emit halt`
	// 48 + 7 = 55 = '7'
	_, out := runForth(src)
	if out != "7" {
		t.Errorf("3+4: output=%q, want '7'", out)
	}
}

func TestForth_Subtraction(t *testing.T) {
	src := `10 3 - 48 + emit halt`
	// 10-3=7, +48='7'
	_, out := runForth(src)
	if out != "7" {
		t.Errorf("10-3: output=%q, want '7'", out)
	}
}

func TestForth_Multiply(t *testing.T) {
	src := `3 5 * 48 + emit halt`
	// 15+48=63='?'
	// Actually 15 is too big for single digit. Use: 2 3 * 48 + emit
	src = `2 3 * 48 + emit halt`
	// 6+48=54='6'
	_, out := runForth(src)
	if out != "6" {
		t.Errorf("2*3: output=%q, want '6'", out)
	}
}

func TestForth_WordDef(t *testing.T) {
	src := `
	: double dup + ;
	3 double 48 + emit halt
`
	// 3 double = 6, +48 = '6'
	_, out := runForth(src)
	if out != "6" {
		t.Errorf("double(3): output=%q, want '6'", out)
	}
}

func TestForth_IfElse(t *testing.T) {
	src := `
	: test 0= if 89 emit else 78 emit then ;
	\ Y=89, N=78
	0 test
	1 test
	halt
`
	// 0 → true → 'Y'; 1 → false → 'N'
	_, out := runForth(src)
	if out != "YN" {
		t.Errorf("if/else: output=%q, want 'YN'", out)
	}
}

func TestForth_Loop(t *testing.T) {
	// Count down from 5, emit each digit
	src := `
	: countdown
		begin
			dup 48 + emit
			1 -
			dup 0=
		until
		drop ;
	5 countdown halt
`
	_, out := runForth(src)
	if out != "54321" {
		t.Errorf("countdown: output=%q, want '54321'", out)
	}
}

func TestForth_Pixel(t *testing.T) {
	// Write pixel at (5, 3) with color 200
	src := `200 3 5 pixel halt`
	sys, _ := runForth(src)
	if sys == nil {
		t.Fatal("system is nil")
	}
	idx := 3*FBWidth + 5
	if sys.FB[idx] != 200 {
		t.Errorf("FB[%d]=%d, want 200", idx, sys.FB[idx])
	}
}

func TestForth_Comparison(t *testing.T) {
	src := `
	5 3 < if 89 emit else 78 emit then
	3 5 < if 89 emit else 78 emit then
	halt
`
	// 5 < 3 → false → N; 3 < 5 → true → Y
	_, out := runForth(src)
	if out != "NY" {
		t.Errorf("comparison: output=%q, want 'NY'", out)
	}
}

func TestForth_NestedWords(t *testing.T) {
	src := `
	: square dup * ;
	: cube dup square * ;
	2 cube
	\ 2^3 = 8
	48 + emit halt
`
	_, out := runForth(src)
	if out != "8" {
		t.Errorf("cube(2): output=%q, want '8'", out)
	}
}

func TestForth_FixedMul(t *testing.T) {
	// F* test: 1.5 * 2.0 = 3.0
	// 1.5 in 8.8 = 384, 2.0 = 512, result = 768 = 3.0
	src := `384 512 f* halt`
	sys, _ := runForth(src)
	if sys == nil {
		t.Fatal("nil system")
	}
	// TOS should be 768 (3.0 in 8.8)
	if sys.CPU.Regs[4] != 768 {
		t.Errorf("f* 1.5*2.0: R4=%d, want 768", sys.CPU.Regs[4])
	}
}

func TestForth_Division(t *testing.T) {
	src := `42 7 / halt`
	sys, _ := runForth(src)
	if sys == nil {
		t.Fatal("nil system")
	}
	if sys.CPU.Regs[4] != 6 {
		t.Errorf("42/7: R4=%d, want 6", sys.CPU.Regs[4])
	}
}

func TestForth_FixedDiv(t *testing.T) {
	// 3.0 / 2.0 = 1.5
	// 768 / 512 → f/ → (768<<8)/512 = 196608/512 = 384 = 1.5
	src := `768 512 f/ halt`
	sys, _ := runForth(src)
	if sys == nil {
		t.Fatal("nil system")
	}
	if sys.CPU.Regs[4] != 384 {
		t.Errorf("f/ 3.0/2.0: R4=%d, want 384", sys.CPU.Regs[4])
	}
}

// runForthLong runs with a higher instruction limit for the raytracer.
func runForthLong(source string, maxCycles int) (*System, string) {
	fc := NewForthCompiler()
	code, err := fc.Compile(source, 0)
	if err != nil {
		return nil, "COMPILE ERROR: " + err.Error()
	}
	sys := NewSystem()
	sys.LoadRAM(0, code)
	sys.CPU.PC = 0
	var out bytes.Buffer
	sys.Stdout = &out
	sys.Run(maxCycles)
	return sys, out.String()
}

func TestForth_Raytracer(t *testing.T) {
	// Render a sphere with gradient shading, no I/J from inside words
	src := `
\ Helper: compute pixel shade from dx dy
\ ( dx dy -- shade )
: shade
  dup * swap dup * +    \ dist2 = dx*dx + dy*dy
  dup 144 < if          \ inside sphere (r=12)
    144 swap -           \ (144 - dist2)
    dup + dup 255 > if drop 255 then
  else
    drop 0
  then
;

\ Main render loop
32 0 do         \ y = 0..31
  64 0 do       \ x = 0..63
    i 32 -      \ dx = x - 32 (on data stack)
    j 16 -      \ dy = y - 16 (on data stack)
    shade       \ ( dx dy -- shade )
    j i pixel   \ write pixel ( color y x )
  loop
loop
halt
`
	sys, errStr := runForthLong(src, 5000000)
	if sys == nil {
		t.Fatalf("raytracer failed: %s", errStr)
	}

	centerIdx := 16*FBWidth + 32
	if sys.FB[centerIdx] == 0 {
		t.Errorf("center pixel is black, expected bright (got %d)", sys.FB[centerIdx])
	}
	if sys.FB[0] != 0 {
		t.Errorf("corner should be 0, got %d", sys.FB[0])
	}
	t.Logf("Center pixel: %d, Edge pixel (32,4): %d", sys.FB[centerIdx], sys.FB[4*FBWidth+32])
}


func TestForth_DoLoop(t *testing.T) {
	src := `5 0 do i 48 + emit loop halt`
	_, out := runForth(src)
	if out != "01234" {
		t.Errorf("do/loop: output=%q, want '01234'", out)
	}
}

func TestForth_NestedDoLoop(t *testing.T) {
	src := `2 0 do 3 0 do j 48 + emit i 48 + emit 32 emit loop loop halt`
	_, out := runForth(src)
	if out != "00 01 02 10 11 12 " {
		t.Errorf("nested do/loop: got=%q, want='00 01 02 10 11 12 '", out)
	}
}

func TestForth_ShadeWord(t *testing.T) {
	src := `
	: shade
	  dup * swap dup * +
	  dup 144 < if
	    144 swap -
	    dup + dup 255 > if drop 255 then
	  else
	    drop 0
	  then
	;
	0 0 shade halt
`
	sys, _ := runForth(src)
	if sys.CPU.Regs[4] != 255 {
		t.Errorf("shade(0,0): R4=%d, want 255", sys.CPU.Regs[4])
	}
}
