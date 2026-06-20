package nand16

// MemoryBus carries the gate-level signals that connect a CPU to memory.
// It is the hardware-side view of the memory port: a backing store (the
// functional model in package memory) drives RData and consumes WData
// according to these control signals.
type MemoryBus struct {
	Addr     Bus   // 16-bit address output (from CPU)
	WData    Bus   // 16-bit write data output (from CPU)
	RData    Bus   // 16-bit read data input (from memory)
	MemRead  *Wire // read enable
	MemWrite *Wire // write enable
	ByteMode *Wire // 1=byte access, 0=word access
}

// NewMemoryBus creates the memory bus signals.
func NewMemoryBus() *MemoryBus {
	return &MemoryBus{
		Addr:     NewBus(16),
		WData:    NewBus(16),
		RData:    NewBus(16),
		MemRead:  NewWire(),
		MemWrite: NewWire(),
		ByteMode: NewWire(),
	}
}
