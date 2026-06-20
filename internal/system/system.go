package system

import (
	"fmt"
	"io"
	"nand16/internal/cpu"
	"nand16/internal/memory"
	"os"
)

const (
	// Memory map
	RAMSize   = 0xF000 // 60KB
	FBBase    = 0xF000 // Framebuffer: 64x32, 1 byte/pixel, 2KB
	FBSize    = 0x0800
	UARTBase  = 0xF800
	UARTData  = 0xF800
	UARTStat  = 0xF802 // bit0=rx ready, bit1=tx ready (always 1)
	TimerBase = 0xF810
	TimerLo   = 0xF810 // cycle counter low 16
	TimerHi   = 0xF812 // cycle counter high 16
	ROMBase   = 0xFE00
	ROMSize   = 0x0200  // 512 bytes
	MemSize   = 0x10000 // 64KB total

	FBWidth  = 64
	FBHeight = 32
)

// System is the complete SoC.
type System struct {
	CPU    *cpu.CPU
	Mem    *memory.Memory
	FB     []byte // framebuffer mirror (shared with the framebuffer device)
	Stdout io.Writer

	fb      *Framebuffer
	uart    *UART
	devices []Device
	cycles  uint32
}

// NewSystem creates the full system.
func NewSystem() *System {
	mem := memory.NewMemory(MemSize)
	c := cpu.NewCPU(mem)

	sys := &System{
		CPU:    c,
		Mem:    mem,
		Stdout: os.Stdout,
	}

	// Memory-mapped peripherals.
	sys.fb = newFramebuffer(FBBase, FBSize)
	sys.FB = sys.fb.data // share the backing slice
	sys.uart = &UART{base: UARTBase, out: func() io.Writer { return sys.Stdout }}
	timer := &Timer{base: TimerBase, cycles: func() uint32 { return sys.cycles }}
	sys.devices = []Device{sys.fb, sys.uart, timer}

	// MMIO handlers. All I/O is performed by software (the OS) through these
	// memory-mapped devices; there is no magic syscall trap into Go.
	mem.OnRead = sys.mmioRead
	mem.OnWrite = sys.mmioWrite

	// Boot from ROM: set PC to ROMBase
	c.PC = ROMBase

	return sys
}

// deviceAt returns the device mapped at addr, or nil.
func (s *System) deviceAt(addr uint16) Device {
	for _, d := range s.devices {
		if addr >= d.Base() && addr < d.Base()+d.Size() {
			return d
		}
	}
	return nil
}

func (s *System) mmioRead(addr uint16) uint16 {
	if d := s.deviceAt(addr); d != nil {
		return d.Read(addr - d.Base())
	}
	return 0
}

func (s *System) mmioWrite(addr uint16, val uint16) {
	if d := s.deviceAt(addr); d != nil {
		d.Write(addr-d.Base(), val)
	}
}

// LoadROM loads binary data into ROM area.
func (s *System) LoadROM(data []byte) {
	if len(data) > ROMSize {
		data = data[:ROMSize]
	}
	copy(s.Mem.Data[ROMBase:], data)
}

// LoadRAM loads binary data at the given address.
func (s *System) LoadRAM(addr int, data []byte) {
	copy(s.Mem.Data[addr:], data)
}

// Run executes until halt or maxCycles.
func (s *System) Run(maxCycles int) {
	for i := 0; i < maxCycles; i++ {
		if !s.CPU.Step() {
			break
		}
		s.cycles++
	}
}

// DumpFB returns framebuffer as ASCII art (for debugging).
func (s *System) DumpFB() string {
	out := ""
	chars := " .:-=+*#%@"
	for y := 0; y < FBHeight; y++ {
		for x := 0; x < FBWidth; x++ {
			v := s.FB[y*FBWidth+x]
			idx := int(v) * len(chars) / 256
			if idx >= len(chars) {
				idx = len(chars) - 1
			}
			out += string(chars[idx])
		}
		out += "\n"
	}
	return out
}

// DumpFBPGM returns framebuffer as PGM image data.
func (s *System) DumpFBPGM() []byte {
	header := fmt.Sprintf("P5\n%d %d\n255\n", FBWidth, FBHeight)
	data := make([]byte, 0, len(header)+FBWidth*FBHeight)
	data = append(data, []byte(header)...)
	data = append(data, s.FB[:FBWidth*FBHeight]...)
	return data
}

// FeedUART provides input bytes for the UART.
func (s *System) FeedUART(data []byte) {
	s.uart.Feed(data)
}
