package nand16

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"os"
)

const (
	// Memory map
	RAMSize  = 0xF000 // 60KB
	FBBase   = 0xF000 // Framebuffer: 64x32, 1 byte/pixel, 2KB
	FBSize   = 0x0800
	UARTBase = 0xF800
	UARTData = 0xF800
	UARTStat = 0xF802 // bit0=rx ready, bit1=tx ready (always 1)
	TimerBase = 0xF810
	TimerLo  = 0xF810 // cycle counter low 16
	TimerHi  = 0xF812 // cycle counter high 16
	ROMBase  = 0xFE00
	ROMSize  = 0x0200 // 512 bytes
	MemSize  = 0x10000 // 64KB total

	FBWidth  = 64
	FBHeight = 32
)

// System is the complete SoC.
type System struct {
	CPU   *CPU
	Mem   *Memory
	FB    []byte // framebuffer mirror
	Stdin io.Reader
	Stdout io.Writer

	uartRxBuf []byte
	cycles    uint32
}

// NewSystem creates the full system.
func NewSystem() *System {
	mi := NewMemoryInterface()
	mem := NewMemory(MemSize, mi)
	cpu := NewCPU(mem)

	sys := &System{
		CPU:    cpu,
		Mem:    mem,
		FB:     make([]byte, FBSize),
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
	}

	// MMIO handlers
	mem.OnRead = sys.mmioRead
	mem.OnWrite = sys.mmioWrite
	cpu.syscall = sys.handleSyscall

	// Boot from ROM: set PC to ROMBase
	cpu.PC = ROMBase

	return sys
}

func (s *System) mmioRead(addr uint16) uint16 {
	switch {
	case addr >= FBBase && addr < FBBase+FBSize:
		off := addr - FBBase
		return uint16(s.FB[off])

	case addr == UARTData:
		if len(s.uartRxBuf) > 0 {
			ch := s.uartRxBuf[0]
			s.uartRxBuf = s.uartRxBuf[1:]
			return uint16(ch)
		}
		return 0

	case addr == UARTStat:
		stat := uint16(0x02) // tx always ready
		if len(s.uartRxBuf) > 0 {
			stat |= 0x01 // rx ready
		}
		return stat

	case addr == TimerLo:
		return uint16(s.cycles)
	case addr == TimerHi:
		return uint16(s.cycles >> 16)
	}
	return 0
}

func (s *System) mmioWrite(addr uint16, val uint16) {
	switch {
	case addr >= FBBase && addr < FBBase+FBSize:
		off := addr - FBBase
		s.FB[off] = byte(val)

	case addr == UARTData:
		if s.Stdout != nil {
			s.Stdout.Write([]byte{byte(val)})
		}
	}
}

// SYSCALL convention: R1 = syscall number, R2-R4 = args, R1 = return value
func (s *System) handleSyscall(cpu *CPU) {
	switch cpu.Regs[1] {
	case 1: // putchar(R2)
		if s.Stdout != nil {
			s.Stdout.Write([]byte{byte(cpu.Regs[2])})
		}
	case 2: // getchar() -> R1
		buf := make([]byte, 1)
		if s.Stdin != nil {
			n, _ := s.Stdin.Read(buf)
			if n > 0 {
				cpu.Regs[1] = uint16(buf[0])
			} else {
				cpu.Regs[1] = 0xFFFF
			}
		}
	case 3: // putpixel(R2=x, R3=y, R4=color)
		x := cpu.Regs[2]
		y := cpu.Regs[3]
		c := cpu.Regs[4]
		if x < FBWidth && y < FBHeight {
			s.FB[uint16(y)*FBWidth+x] = byte(c)
		}
	case 4: // halt
		cpu.Halt = true
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

// RenderPNG returns the framebuffer as a grayscale PNG image (scaled up).
func (s *System) RenderPNG(scale int) []byte {
	return s.renderPNG(scale, false)
}

// RenderPNGColor returns the framebuffer as RGB332-decoded color PNG (scaled up).
func (s *System) RenderPNGColor(scale int) []byte {
	return s.renderPNG(scale, true)
}

func (s *System) renderPNG(scale int, colorMode bool) []byte {
	if scale < 1 {
		scale = 1
	}
	w := FBWidth * scale
	h := FBHeight * scale

	rgba := make([]byte, w*h*4)
	for y := 0; y < FBHeight; y++ {
		for x := 0; x < FBWidth; x++ {
			v := s.FB[y*FBWidth+x]
			var r, g, b byte
			if colorMode {
				// RGB332: RRRGGGBB
				r = (v >> 5) * 36        // 0-7 → 0-252
				g = ((v >> 2) & 7) * 36  // 0-7 → 0-252
				b = (v & 3) * 85         // 0-3 → 0-255
			} else {
				r, g, b = v, v, v
			}
			for sy := 0; sy < scale; sy++ {
				for sx := 0; sx < scale; sx++ {
					off := ((y*scale+sy)*w + x*scale + sx) * 4
					rgba[off+0] = r
					rgba[off+1] = g
					rgba[off+2] = b
					rgba[off+3] = 255
				}
			}
		}
	}

	img := &pngImage{w: w, h: h, rgba: rgba}
	var buf bytes.Buffer
	png.Encode(&buf, img)
	return buf.Bytes()
}

// RenderPNGRGB555 reads the framebuffer as 16-bit RGB555 pixels from raw memory.
// Each pixel is 2 bytes (little-endian) at FBBase + (y*FBWidth+x)*2.
// Format: 0RRRRRGGGGGBBBBB
func (s *System) RenderPNGRGB555(scale int) []byte {
	if scale < 1 {
		scale = 1
	}
	w := FBWidth * scale
	h := FBHeight * scale

	rgba := make([]byte, w*h*4)
	for y := 0; y < FBHeight; y++ {
		for x := 0; x < FBWidth; x++ {
			addr := FBBase + uint16((y*FBWidth+x)*2)
			lo := s.Mem.Data[addr]
			hi := s.Mem.Data[addr+1]
			pixel := uint16(lo) | uint16(hi)<<8

			r5 := (pixel >> 10) & 0x1F
			g5 := (pixel >> 5) & 0x1F
			b5 := pixel & 0x1F

			r8 := byte(r5*255/31)
			g8 := byte(g5*255/31)
			b8 := byte(b5*255/31)

			for sy := 0; sy < scale; sy++ {
				for sx := 0; sx < scale; sx++ {
					off := ((y*scale+sy)*w + x*scale + sx) * 4
					rgba[off+0] = r8
					rgba[off+1] = g8
					rgba[off+2] = b8
					rgba[off+3] = 255
				}
			}
		}
	}

	img := &pngImage{w: w, h: h, rgba: rgba}
	var buf bytes.Buffer
	png.Encode(&buf, img)
	return buf.Bytes()
}

// pngImage implements image.Image for PNG encoding.
type pngImage struct {
	w, h int
	rgba []byte
}

func (p *pngImage) ColorModel() color.Model { return color.RGBAModel }
func (p *pngImage) Bounds() image.Rectangle { return image.Rect(0, 0, p.w, p.h) }
func (p *pngImage) At(x, y int) color.Color {
	off := (y*p.w + x) * 4
	return color.RGBA{p.rgba[off], p.rgba[off+1], p.rgba[off+2], p.rgba[off+3]}
}

// FeedUART provides input bytes for the UART.
func (s *System) FeedUART(data []byte) {
	s.uartRxBuf = append(s.uartRxBuf, data...)
}
