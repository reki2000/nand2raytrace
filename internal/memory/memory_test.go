package memory

import "testing"

func TestMemory(t *testing.T) {
	mem := NewMemory(65536)

	// Write word 0xABCD at address 0x100, read it back
	mem.Write16(0x100, 0xABCD)
	if got := mem.Read16(0x100); got != 0xABCD {
		t.Errorf("Word read = 0x%X, want 0xABCD", got)
	}

	// Read low/high bytes
	if got := mem.Read8(0x100); got != 0xCD {
		t.Errorf("Byte read low = 0x%X, want 0xCD", got)
	}
	if got := mem.Read8(0x101); got != 0xAB {
		t.Errorf("Byte read high = 0x%X, want 0xAB", got)
	}

	// Byte write
	mem.Write8(0x200, 0x5A)
	if got := mem.Read8(0x200); got != 0x5A {
		t.Errorf("Byte write/read = 0x%X, want 0x5A", got)
	}
}

func TestMemoryMMIO(t *testing.T) {
	mem := NewMemory(65536)
	var lastWrite struct {
		addr uint16
		val  uint16
	}
	mem.OnWrite = func(addr, val uint16) { lastWrite.addr, lastWrite.val = addr, val }
	mem.OnRead = func(addr uint16) uint16 { return 0x1234 }

	// Writes at or above MMIOBase notify the handler...
	mem.Write16(MMIOBase, 0xBEEF)
	if lastWrite.addr != MMIOBase || lastWrite.val != 0xBEEF {
		t.Errorf("MMIO write = (0x%X,0x%X), want (0x%X,0xBEEF)", lastWrite.addr, lastWrite.val, MMIOBase)
	}
	// ...and reads are served by the handler.
	if got := mem.Read16(MMIOBase); got != 0x1234 {
		t.Errorf("MMIO read = 0x%X, want 0x1234", got)
	}

	// Below MMIOBase the handlers are not consulted.
	mem.Write16(0x10, 0x99)
	if mem.Read16(0x10) != 0x99 {
		t.Errorf("non-MMIO read should hit backing store")
	}
}
