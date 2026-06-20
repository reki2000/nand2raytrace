// Package memory provides the system memory backend: a Go-native byte store
// with memory-mapped I/O dispatch. It is a purely functional model with no
// dependency on the gate-level simulation.
package memory

// Memory is the Go-native memory backend (not gate-level).
type Memory struct {
	Data []byte
	// MMIO handlers
	OnRead  func(addr uint16) uint16
	OnWrite func(addr uint16, val uint16)
}

// MMIOBase is the start of the memory-mapped I/O region. Accesses at or above
// this address are routed to the OnRead/OnWrite handlers (when set).
const MMIOBase = 0xF000

// NewMemory creates a memory backend with the given size.
func NewMemory(size int) *Memory {
	return &Memory{
		Data: make([]byte, size),
	}
}

// Read16 reads a little-endian 16-bit word, dispatching to OnRead for MMIO.
func (mem *Memory) Read16(addr uint16) uint16 {
	if addr >= MMIOBase && mem.OnRead != nil {
		return mem.OnRead(addr)
	}
	if int(addr)+1 < len(mem.Data) {
		return uint16(mem.Data[addr]) | uint16(mem.Data[addr+1])<<8
	}
	return 0
}

// Write16 writes a little-endian 16-bit word, also notifying OnWrite for MMIO.
func (mem *Memory) Write16(addr uint16, val uint16) {
	if int(addr)+1 < len(mem.Data) {
		mem.Data[addr] = byte(val)
		mem.Data[addr+1] = byte(val >> 8)
	}
	if addr >= MMIOBase && mem.OnWrite != nil {
		mem.OnWrite(addr, val)
	}
}

// Read8 reads a single byte, dispatching to OnRead for MMIO.
func (mem *Memory) Read8(addr uint16) byte {
	if addr >= MMIOBase && mem.OnRead != nil {
		return byte(mem.OnRead(addr))
	}
	if int(addr) < len(mem.Data) {
		return mem.Data[addr]
	}
	return 0
}

// Write8 writes a single byte, also notifying OnWrite for MMIO.
func (mem *Memory) Write8(addr uint16, val byte) {
	if int(addr) < len(mem.Data) {
		mem.Data[addr] = val
	}
	if addr >= MMIOBase && mem.OnWrite != nil {
		mem.OnWrite(addr, uint16(val))
	}
}

// LoadBinary loads a binary image at the given address.
func (mem *Memory) LoadBinary(addr int, data []byte) {
	copy(mem.Data[addr:], data)
}
