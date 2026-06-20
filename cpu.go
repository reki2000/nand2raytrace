package nand16

import "fmt"

// CPU is the NAND-16 processor.
type CPU struct {
	PC   uint16
	Regs [8]uint16 // R0 is always 0
	Halt bool

	mem     *Memory
	mi      *MemoryInterface
	cycles  int
	syscall func(cpu *CPU) // SYSCALL handler
}

// NewCPU creates a CPU connected to the given memory.
func NewCPU(mem *Memory) *CPU {
	return &CPU{mem: mem, mi: mem.MI}
}

// Step executes one instruction. Returns false if halted.
func (c *CPU) Step() bool {
	if c.Halt {
		return false
	}
	c.Regs[0] = 0 // enforce R0=0

	// Fetch
	instr := c.memRead16(c.PC)
	d := Decode(instr)

	// Read registers
	rs1Val := c.Regs[d.Rs1]
	rs2Val := c.Regs[d.Rs2]
	rdVal := c.Regs[d.Rd]

	nextPC := c.PC + 2
	writeReg := false
	writeDst := d.Rd
	var writeVal uint16

	switch d.Op {
	case OpALU:
		result := c.aluExec(d.Func, rs1Val, rs2Val)
		writeReg = true
		writeVal = result

	case OpMUL:
		// R-type: rd = rs1 * rs2 (signed)
		// func=0: rd = low16(product), func=1: rd = high16(product)
		a := int32(int16(rs1Val))
		b := int32(int16(rs2Val))
		product := uint32(a * b)
		writeReg = true
		if d.Func == 0 {
			writeVal = uint16(product)
		} else {
			writeVal = uint16(product >> 16)
		}

	case OpADDI:
		writeReg = true
		writeVal = rs1Val + uint16(int16(d.Imm6))

	case OpANDI:
		writeReg = true
		writeVal = rs1Val & uint16(d.Imm6&0xFFFF)

	case OpORI:
		writeReg = true
		writeVal = rs1Val | uint16(d.Imm6&0xFFFF)

	case OpLUI:
		// Load upper: rd = imm6 << 10 (puts 6 bits in [15:10])
		writeReg = true
		writeVal = uint16((d.Imm6 & 0x3F) << 10)

	case OpLW:
		addr := rs1Val + uint16(int16(d.Imm6))
		writeReg = true
		writeVal = c.memRead16(addr)

	case OpSW:
		// SW rs1, imm6(rd): mem[rd + imm6] = rs1
		addr := rdVal + uint16(int16(d.Imm6))
		c.memWrite16(addr, rs1Val)

	case OpLB:
		addr := rs1Val + uint16(int16(d.Imm6))
		writeReg = true
		writeVal = uint16(c.memRead8(addr))

	case OpSB:
		addr := rdVal + uint16(int16(d.Imm6))
		c.memWrite8(addr, byte(rs1Val))

	case OpBEQ:
		// B-type: rs1 is in [11:9], rs2 in [8:6], off6 in [5:0]
		r1 := c.Regs[d.Rd]   // [11:9] is actually rs1 for B-type
		r2 := c.Regs[d.Rs1]  // [8:6] is rs2 for B-type
		if r1 == r2 {
			nextPC = uint16(int(c.PC) + d.Imm6*2)
		}

	case OpBNE:
		r1 := c.Regs[d.Rd]
		r2 := c.Regs[d.Rs1]
		if r1 != r2 {
			nextPC = uint16(int(c.PC) + d.Imm6*2)
		}

	case OpBLT:
		r1 := c.Regs[d.Rd]
		r2 := c.Regs[d.Rs1]
		if int16(r1) < int16(r2) {
			nextPC = uint16(int(c.PC) + d.Imm6*2)
		}

	case OpBGE:
		r1 := c.Regs[d.Rd]
		r2 := c.Regs[d.Rs1]
		if int16(r1) >= int16(r2) {
			nextPC = uint16(int(c.PC) + d.Imm6*2)
		}

	case OpJAL:
		c.Regs[7] = nextPC // link register = R7
		nextPC = uint16(int(c.PC) + d.Off12*2)

	case OpSYSTEM:
		switch d.Func {
		case 0: // JALR rs1
			c.Regs[7] = nextPC
			nextPC = rs1Val
		case 1: // HALT
			c.Halt = true
		case 2: // SYSCALL
			if c.syscall != nil {
				c.syscall(c)
			}
		default: // NOP
		}
	}

	if writeReg && writeDst != 0 {
		c.Regs[writeDst] = writeVal
	}
	c.Regs[0] = 0
	c.PC = nextPC
	c.cycles++
	return !c.Halt
}

// aluExec performs ALU operation (matching gate-level ALU op encoding)
func (c *CPU) aluExec(fn int, a, b uint16) uint16 {
	switch fn {
	case 0: // ADD
		return a + b
	case 1: // SUB
		return a - b
	case 2: // AND
		return a & b
	case 3: // OR
		return a | b
	case 4: // XOR
		return a ^ b
	case 5: // SHL
		return a << (b & 0xF)
	case 6: // SHR
		return a >> (b & 0xF)
	case 7: // SRA
		return uint16(int16(a) >> (b & 0xF))
	}
	return 0
}

func (c *CPU) memRead16(addr uint16) uint16 {
	if addr >= 0xF000 && c.mem.OnRead != nil {
		return c.mem.OnRead(addr)
	}
	if int(addr)+1 < len(c.mem.Data) {
		return uint16(c.mem.Data[addr]) | uint16(c.mem.Data[addr+1])<<8
	}
	return 0
}

func (c *CPU) memWrite16(addr uint16, val uint16) {
	if int(addr)+1 < len(c.mem.Data) {
		c.mem.Data[addr] = byte(val)
		c.mem.Data[addr+1] = byte(val >> 8)
	}
	if addr >= 0xF000 && c.mem.OnWrite != nil {
		c.mem.OnWrite(addr, val)
	}
}

func (c *CPU) memRead8(addr uint16) byte {
	if addr >= 0xF000 && c.mem.OnRead != nil {
		return byte(c.mem.OnRead(addr))
	}
	if int(addr) < len(c.mem.Data) {
		return c.mem.Data[addr]
	}
	return 0
}

func (c *CPU) memWrite8(addr uint16, val byte) {
	if int(addr) < len(c.mem.Data) {
		c.mem.Data[addr] = val
	}
	if addr >= 0xF000 && c.mem.OnWrite != nil {
		c.mem.OnWrite(addr, uint16(val))
	}
}

// Run executes until halt or maxCycles.
func (c *CPU) Run(maxCycles int) int {
	for i := 0; i < maxCycles; i++ {
		if !c.Step() {
			return c.cycles
		}
	}
	return c.cycles
}

// DumpRegs returns a string of register values.
func (c *CPU) DumpRegs() string {
	s := fmt.Sprintf("PC=%04X ", c.PC)
	for i := 0; i < 8; i++ {
		s += fmt.Sprintf("R%d=%04X ", i, c.Regs[i])
	}
	return s
}
