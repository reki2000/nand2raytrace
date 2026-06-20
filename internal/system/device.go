package system

import "io"

// Device is a memory-mapped peripheral occupying the half-open address range
// [Base, Base+Size). Offsets passed to Read/Write are relative to Base.
type Device interface {
	Base() uint16
	Size() uint16
	Read(off uint16) uint16
	Write(off uint16, val uint16)
}

// Framebuffer is a byte-per-pixel display buffer.
type Framebuffer struct {
	base uint16
	data []byte
}

func newFramebuffer(base uint16, size int) *Framebuffer {
	return &Framebuffer{base: base, data: make([]byte, size)}
}

func (f *Framebuffer) Base() uint16 { return f.base }
func (f *Framebuffer) Size() uint16 { return uint16(len(f.data)) }

func (f *Framebuffer) Read(off uint16) uint16     { return uint16(f.data[off]) }
func (f *Framebuffer) Write(off uint16, v uint16) { f.data[off] = byte(v) }

// UART is a byte-oriented serial port. The data register echoes received
// bytes; the status register reports tx/rx readiness.
type UART struct {
	base  uint16
	rxBuf []byte
	out   func() io.Writer // resolved lazily so System.Stdout stays reassignable
}

const (
	uartDataOff = 0 // data register
	uartStatOff = 2 // status register
	uartSize    = 4
)

func (u *UART) Base() uint16 { return u.base }
func (u *UART) Size() uint16 { return uartSize }

func (u *UART) Read(off uint16) uint16 {
	switch off {
	case uartDataOff:
		if len(u.rxBuf) > 0 {
			ch := u.rxBuf[0]
			u.rxBuf = u.rxBuf[1:]
			return uint16(ch)
		}
	case uartStatOff:
		stat := uint16(0x02) // tx always ready
		if len(u.rxBuf) > 0 {
			stat |= 0x01 // rx ready
		}
		return stat
	}
	return 0
}

func (u *UART) Write(off uint16, v uint16) {
	if off == uartDataOff {
		u.Putchar(byte(v))
	}
}

// Putchar emits a single byte to the output writer.
func (u *UART) Putchar(ch byte) {
	if w := u.out(); w != nil {
		w.Write([]byte{ch})
	}
}

// Feed queues bytes to be read back from the data register.
func (u *UART) Feed(data []byte) { u.rxBuf = append(u.rxBuf, data...) }

// Timer exposes a free-running cycle counter as two 16-bit registers.
type Timer struct {
	base   uint16
	cycles func() uint32
}

const (
	timerLoOff = 0
	timerHiOff = 2
	timerSize  = 4
)

func (t *Timer) Base() uint16 { return t.base }
func (t *Timer) Size() uint16 { return timerSize }

func (t *Timer) Read(off uint16) uint16 {
	switch off {
	case timerLoOff:
		return uint16(t.cycles())
	case timerHiOff:
		return uint16(t.cycles() >> 16)
	}
	return 0
}

func (t *Timer) Write(off uint16, v uint16) {} // counter is read-only
