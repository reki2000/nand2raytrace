package op

import "testing"

// TestDecodeFields verifies that Decode extracts each bit field from the
// correct position of a 16-bit instruction word.
func TestDecodeFields(t *testing.T) {
	// op=0xA rd=5 rs1=3 rs2=2 func=1
	// 1010 101 011 010 001 = 0xAAD1
	instr := uint16(0xA<<12 | 5<<9 | 3<<6 | 2<<3 | 1)
	if instr != 0xAAD1 {
		t.Fatalf("test setup wrong: instr=0x%04X want 0xAAD1", instr)
	}
	d := Decode(instr)
	if d.Op != 0xA {
		t.Errorf("Op = %d, want 10", d.Op)
	}
	if d.Rd != 5 {
		t.Errorf("Rd = %d, want 5", d.Rd)
	}
	if d.Rs1 != 3 {
		t.Errorf("Rs1 = %d, want 3", d.Rs1)
	}
	if d.Rs2 != 2 {
		t.Errorf("Rs2 = %d, want 2", d.Rs2)
	}
	if d.Func != 1 {
		t.Errorf("Func = %d, want 1", d.Func)
	}
	if d.Raw != instr {
		t.Errorf("Raw = 0x%04X, want 0x%04X", d.Raw, instr)
	}
}

// TestDecodeImm6SignExtend covers the imm6 sign-extension boundaries.
func TestDecodeImm6SignExtend(t *testing.T) {
	tests := []struct {
		field uint16 // low 6 bits
		want  int
	}{
		{0x00, 0},
		{0x01, 1},
		{0x1F, 31},   // largest positive
		{0x20, -32},  // smallest negative (sign bit set)
		{0x21, -31},
		{0x3F, -1},   // all ones
	}
	for _, tt := range tests {
		// Imm6 must ignore the upper bits; pack noise into [15:6].
		d := Decode(0xFFC0 | tt.field)
		if d.Imm6 != tt.want {
			t.Errorf("Decode(field=0x%02X).Imm6 = %d, want %d", tt.field, d.Imm6, tt.want)
		}
	}
}

// TestDecodeOff12SignExtend covers the off12 sign-extension boundaries.
func TestDecodeOff12SignExtend(t *testing.T) {
	tests := []struct {
		field uint16 // low 12 bits
		want  int
	}{
		{0x000, 0},
		{0x001, 1},
		{0x7FF, 2047},  // largest positive
		{0x800, -2048}, // smallest negative
		{0x801, -2047},
		{0xFFF, -1},    // all ones
	}
	for _, tt := range tests {
		// Off12 must ignore the opcode bits [15:12].
		d := Decode(0xF000 | tt.field)
		if d.Off12 != tt.want {
			t.Errorf("Decode(field=0x%03X).Off12 = %d, want %d", tt.field, d.Off12, tt.want)
		}
	}
}

// TestEncodeRRoundTrip checks that EncodeR places fields where Decode reads them.
func TestEncodeRRoundTrip(t *testing.T) {
	for op := 0; op < 16; op++ {
		for rd := 0; rd < 8; rd++ {
			for rs1 := 0; rs1 < 8; rs1++ {
				for rs2 := 0; rs2 < 8; rs2++ {
					for fn := 0; fn < 8; fn++ {
						w := EncodeR(op, rd, rs1, rs2, fn)
						d := Decode(w)
						if d.Op != op || d.Rd != rd || d.Rs1 != rs1 || d.Rs2 != rs2 || d.Func != fn {
							t.Fatalf("EncodeR(%d,%d,%d,%d,%d) -> 0x%04X decoded as Op=%d Rd=%d Rs1=%d Rs2=%d Func=%d",
								op, rd, rs1, rs2, fn, w, d.Op, d.Rd, d.Rs1, d.Rs2, d.Func)
						}
					}
				}
			}
		}
	}
}

// TestEncodeIRoundTrip checks I-type encode then decode recovers signed imm6.
func TestEncodeIRoundTrip(t *testing.T) {
	for op := 0; op < 16; op++ {
		for rd := 0; rd < 8; rd++ {
			for rs1 := 0; rs1 < 8; rs1++ {
				for imm := -32; imm <= 31; imm++ {
					w := EncodeI(op, rd, rs1, imm)
					d := Decode(w)
					if d.Op != op || d.Rd != rd || d.Rs1 != rs1 || d.Imm6 != imm {
						t.Fatalf("EncodeI(%d,%d,%d,%d) -> 0x%04X decoded Op=%d Rd=%d Rs1=%d Imm6=%d",
							op, rd, rs1, imm, w, d.Op, d.Rd, d.Rs1, d.Imm6)
					}
				}
			}
		}
	}
}

// TestEncodeBRoundTrip checks B-type field layout: rs1->[11:9], rs2->[8:6],
// off6->[5:0] (sign-extended via the Imm6 field on decode).
func TestEncodeBRoundTrip(t *testing.T) {
	for op := 0; op < 16; op++ {
		for rs1 := 0; rs1 < 8; rs1++ {
			for rs2 := 0; rs2 < 8; rs2++ {
				for off := -32; off <= 31; off++ {
					w := EncodeB(op, rs1, rs2, off)
					d := Decode(w)
					// B-type reuses the Rd/Rs1/Imm6 decode positions.
					if d.Op != op || d.Rd != rs1 || d.Rs1 != rs2 || d.Imm6 != off {
						t.Fatalf("EncodeB(%d,%d,%d,%d) -> 0x%04X decoded Op=%d [11:9]=%d [8:6]=%d Imm6=%d",
							op, rs1, rs2, off, w, d.Op, d.Rd, d.Rs1, d.Imm6)
					}
				}
			}
		}
	}
}

// TestEncodeJRoundTrip checks J-type encode then decode recovers signed off12.
func TestEncodeJRoundTrip(t *testing.T) {
	for op := 0; op < 16; op++ {
		for off := -2048; off <= 2047; off++ {
			w := EncodeJ(op, off)
			d := Decode(w)
			if d.Op != op || d.Off12 != off {
				t.Fatalf("EncodeJ(%d,%d) -> 0x%04X decoded Op=%d Off12=%d",
					op, off, w, d.Op, d.Off12)
			}
		}
	}
}
