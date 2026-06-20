package nand16

// Register16 is a 16-bit register with load enable.
// When load=1 at clock edge, D is captured into Q.
type Register16 struct {
	D    Bus
	Load *Wire
	Q    Bus
	ffs  [16]*FlipFlop
	mux  Bus // internal: muxed input to FFs
}

// NewRegister16 creates a 16-bit loadable register.
func NewRegister16(m *Module, d Bus, load *Wire) Bus {
	q := NewBus(16)
	for i := 0; i < 16; i++ {
		muxOut := Mux2(m, load, q[i], d[i])
		m.DFFTo(muxOut, q[i])
	}
	return q
}

// RegisterFile: 8 x 16-bit registers with 2 read ports + 1 write port.
// R0 is hardwired to 0.
type RegisterFile struct {
	m       *Module
	regs    [8]Bus // register outputs
	RdData1 Bus    // read port 1 output
	RdData2 Bus    // read port 2 output
}

// NewRegisterFile creates the register file.
// rs1, rs2: 3-bit read address buses
// rd: 3-bit write address bus
// wData: 16-bit write data bus
// wEn: write enable
func NewRegisterFile(m *Module, rs1, rs2, rd Bus, wData Bus, wEn *Wire) *RegisterFile {
	rf := &RegisterFile{m: m}

	// Decode write address to 8 one-hot signals
	wSel := decode3to8(m, rd, wEn)

	// Create 8 registers. R0 stays 0 (wSel[0] forced to 0).
	zero := NewWire() // constant 0
	wSel[0] = zero    // R0 never written

	for i := 0; i < 8; i++ {
		rf.regs[i] = NewRegister16(m, wData, wSel[i])
	}
	// Force R0 outputs to 0
	rf.regs[0] = NewBusConst(16, 0)

	// Read muxes
	rf.RdData1 = Mux8Bus(m, rs1,
		rf.regs[0], rf.regs[1], rf.regs[2], rf.regs[3],
		rf.regs[4], rf.regs[5], rf.regs[6], rf.regs[7])
	rf.RdData2 = Mux8Bus(m, rs2,
		rf.regs[0], rf.regs[1], rf.regs[2], rf.regs[3],
		rf.regs[4], rf.regs[5], rf.regs[6], rf.regs[7])

	return rf
}

// decode3to8 generates 8 one-hot signals from a 3-bit address + enable.
func decode3to8(m *Module, addr Bus, en *Wire) [8]*Wire {
	var out [8]*Wire
	n0 := NOT(m, addr[0])
	n1 := NOT(m, addr[1])
	n2 := NOT(m, addr[2])
	out[0] = AND3(m, n2, n1, n0)
	out[1] = AND3(m, n2, n1, addr[0])
	out[2] = AND3(m, n2, addr[1], n0)
	out[3] = AND3(m, n2, addr[1], addr[0])
	out[4] = AND3(m, addr[2], n1, n0)
	out[5] = AND3(m, addr[2], n1, addr[0])
	out[6] = AND3(m, addr[2], addr[1], n0)
	out[7] = AND3(m, addr[2], addr[1], addr[0])
	// AND with enable
	for i := range out {
		out[i] = AND(m, out[i], en)
	}
	return out
}

// Counter16: loadable counter with increment and hold.
// mode: 0=hold, 1=increment, load overrides all.
type Counter16 struct {
	Q Bus
}

// NewCounter16 creates a loadable 16-bit counter.
// When load=1, loads from d. When inc=1 (and load=0), increments.
// Otherwise holds.
func NewCounter16(m *Module, d Bus, load, inc *Wire) Bus {
	q := NewBus(16)
	incremented, _ := Adder16(m, q, NewBusConst(16, 0), inc)
	// mux: load ? d : incremented (when inc=0, incremented=q)
	muxed := make(Bus, 16)
	Mux2Bus(m, load, incremented, d, muxed)

	for i := 0; i < 16; i++ {
		m.DFFTo(muxed[i], q[i])
	}
	return q
}

// MemoryInterface provides the gate-level side of the memory bus.
// The actual storage is a Go []byte managed by the simulator.
type MemoryInterface struct {
	Addr     Bus    // 16-bit address output (from CPU)
	WData    Bus    // 16-bit write data output (from CPU)
	RData    Bus    // 16-bit read data input (from memory)
	MemRead  *Wire  // read enable
	MemWrite *Wire  // write enable
	ByteMode *Wire  // 1=byte access, 0=word access
}

// NewMemoryInterface creates the memory bus signals.
func NewMemoryInterface() *MemoryInterface {
	return &MemoryInterface{
		Addr:     NewBus(16),
		WData:    NewBus(16),
		RData:    NewBus(16),
		MemRead:  NewWire(),
		MemWrite: NewWire(),
		ByteMode: NewWire(),
	}
}

// Memory is the Go-native memory backend (not gate-level).
type Memory struct {
	Data []byte
	MI   *MemoryInterface
	// MMIO handlers
	OnRead  func(addr uint16) uint16
	OnWrite func(addr uint16, val uint16)
}

// NewMemory creates a memory backend with the given size.
func NewMemory(size int, mi *MemoryInterface) *Memory {
	return &Memory{
		Data: make([]byte, size),
		MI:   mi,
	}
}

// Eval reads/writes memory based on control signals.
// Called after gate evaluation, before FF tick.
func (mem *Memory) Eval() {
	addr := uint16(mem.MI.Addr.GetVal())

	if mem.MI.MemRead.Val {
		var val uint16
		if addr >= 0xF000 && mem.OnRead != nil {
			val = mem.OnRead(addr)
		} else if mem.MI.ByteMode.Val {
			if int(addr) < len(mem.Data) {
				val = uint16(mem.Data[addr])
			}
		} else {
			if int(addr)+1 < len(mem.Data) {
				val = uint16(mem.Data[addr]) | uint16(mem.Data[addr+1])<<8
			}
		}
		mem.MI.RData.SetVal(int(val))
	}

	if mem.MI.MemWrite.Val {
		val := uint16(mem.MI.WData.GetVal())
		if addr >= 0xF000 && mem.OnWrite != nil {
			mem.OnWrite(addr, val)
		} else if mem.MI.ByteMode.Val {
			if int(addr) < len(mem.Data) {
				mem.Data[addr] = byte(val)
			}
		} else {
			if int(addr)+1 < len(mem.Data) {
				mem.Data[addr] = byte(val)
				mem.Data[addr+1] = byte(val >> 8)
			}
		}
	}
}

// LoadBinary loads a binary image at the given address.
func (mem *Memory) LoadBinary(addr int, data []byte) {
	copy(mem.Data[addr:], data)
}
