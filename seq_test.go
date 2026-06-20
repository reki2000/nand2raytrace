package nand16

import "testing"

func TestRegister16(t *testing.T) {
	m := NewModule("test")
	d := NewBus(16)
	load := NewWire()
	q := NewRegister16(m, d, load)
	sim := NewSimulator(m)

	// Initially 0
	sim.Cycle()
	if got := q.GetVal(); got != 0 {
		t.Errorf("Reg initial = %d, want 0", got)
	}

	// Load 0x1234
	d.SetVal(0x1234)
	load.Val = true
	sim.Settle() // propagate mux
	sim.Cycle()  // clock edge
	if got := q.GetVal(); got != 0x1234 {
		t.Errorf("After load 0x1234: got 0x%X", got)
	}

	// Hold (load=0)
	d.SetVal(0xFFFF)
	load.Val = false
	sim.Settle()
	sim.Cycle()
	if got := q.GetVal(); got != 0x1234 {
		t.Errorf("After hold: got 0x%X, want 0x1234", got)
	}
}

func TestRegisterFile(t *testing.T) {
	m := NewModule("test")
	rs1 := NewBus(3)
	rs2 := NewBus(3)
	rd := NewBus(3)
	wData := NewBus(16)
	wEn := NewWire()

	rf := NewRegisterFile(m, rs1, rs2, rd, wData, wEn)
	sim := NewSimulator(m)

	// Write 0x42 to R1
	rd.SetVal(1)
	wData.SetVal(0x42)
	wEn.Val = true
	sim.Settle()
	sim.Cycle()

	// Read R1 from port 1
	wEn.Val = false
	rs1.SetVal(1)
	sim.Settle()
	if got := rf.RdData1.GetVal(); got != 0x42 {
		t.Errorf("R1 read port1 = 0x%X, want 0x42", got)
	}

	// Write 0xBEEF to R5
	rd.SetVal(5)
	wData.SetVal(0xBEEF)
	wEn.Val = true
	sim.Settle()
	sim.Cycle()

	// Read R1 from port1, R5 from port2
	wEn.Val = false
	rs1.SetVal(1)
	rs2.SetVal(5)
	sim.Settle()
	if got := rf.RdData1.GetVal(); got != 0x42 {
		t.Errorf("R1 = 0x%X, want 0x42", got)
	}
	if got := rf.RdData2.GetVal(); got != 0xBEEF {
		t.Errorf("R5 = 0x%X, want 0xBEEF", got)
	}

	// R0 should always be 0
	rs1.SetVal(0)
	sim.Settle()
	if got := rf.RdData1.GetVal(); got != 0 {
		t.Errorf("R0 = 0x%X, want 0", got)
	}

	// Write to R0 should be ignored
	rd.SetVal(0)
	wData.SetVal(0xFFFF)
	wEn.Val = true
	sim.Settle()
	sim.Cycle()
	wEn.Val = false
	rs1.SetVal(0)
	sim.Settle()
	if got := rf.RdData1.GetVal(); got != 0 {
		t.Errorf("R0 after write = 0x%X, want 0", got)
	}
}

func TestCounter16(t *testing.T) {
	m := NewModule("test")
	d := NewBus(16)
	load := NewWire()
	inc := NewWire()
	q := NewCounter16(m, d, load, inc)
	sim := NewSimulator(m)

	// Load 10
	d.SetVal(10)
	load.Val = true
	inc.Val = false
	sim.Settle()
	sim.Cycle()
	load.Val = false

	// Verify loaded
	sim.Settle()
	if got := q.GetVal(); got != 10 {
		t.Errorf("After load: %d, want 10", got)
	}

	// Increment 3 times
	inc.Val = true
	for i := 0; i < 3; i++ {
		sim.Settle()
		sim.Cycle()
	}
	sim.Settle()
	if got := q.GetVal(); got != 13 {
		t.Errorf("After 3 increments: %d, want 13", got)
	}

	// Hold
	inc.Val = false
	sim.Settle()
	sim.Cycle()
	sim.Settle()
	if got := q.GetVal(); got != 13 {
		t.Errorf("After hold: %d, want 13", got)
	}
}

func TestMemory(t *testing.T) {
	mi := NewMemoryInterface()
	mem := NewMemory(65536, mi)

	// Write word 0xABCD at address 0x100
	mi.Addr.SetVal(0x100)
	mi.WData.SetVal(0xABCD)
	mi.MemWrite.Val = true
	mi.ByteMode.Val = false
	mem.Eval()
	mi.MemWrite.Val = false

	// Read word
	mi.MemRead.Val = true
	mem.Eval()
	if got := mi.RData.GetVal(); got != 0xABCD {
		t.Errorf("Word read = 0x%X, want 0xABCD", got)
	}

	// Read byte (low)
	mi.ByteMode.Val = true
	mem.Eval()
	if got := mi.RData.GetVal(); got != 0xCD {
		t.Errorf("Byte read low = 0x%X, want 0xCD", got)
	}

	// Read byte (high)
	mi.Addr.SetVal(0x101)
	mem.Eval()
	if got := mi.RData.GetVal(); got != 0xAB {
		t.Errorf("Byte read high = 0x%X, want 0xAB", got)
	}
}
