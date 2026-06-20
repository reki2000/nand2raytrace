package cpu

import (
	"fmt"
	"nand16/internal/memory"
	"nand16/internal/op"
)

// CPU is the NAND-16 processor.
type CPU struct {
	PC   uint16
	Regs [8]uint16 // R0 is always 0
	Halt bool

	mem    *memory.Memory
	cycles int
}

// NewCPU creates a CPU connected to the given memory.
func NewCPU(mem *memory.Memory) *CPU {
	return &CPU{mem: mem}
}

// Step executes one instruction. Returns false if halted.
func (c *CPU) Step() bool {
	if c.Halt {
		return false
	}
	c.Regs[0] = 0 // enforce R0=0

	// Fetch
	instr := c.mem.Read16(c.PC)
	d := op.Decode(instr)

	// Read registers
	rs1Val := c.Regs[d.Rs1]
	rs2Val := c.Regs[d.Rs2]
	rdVal := c.Regs[d.Rd]

	nextPC := c.PC + 2
	writeReg := false
	writeDst := d.Rd
	var writeVal uint16

	switch d.Op {
	case op.OpALU:
		result := c.aluExec(d.Func, rs1Val, rs2Val)
		writeReg = true
		writeVal = result

	case op.OpMUL:
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

	case op.OpADDI:
		writeReg = true
		writeVal = rs1Val + uint16(int16(d.Imm6))

	case op.OpANDI:
		writeReg = true
		writeVal = rs1Val & uint16(d.Imm6&0xFFFF)

	case op.OpORI:
		writeReg = true
		writeVal = rs1Val | uint16(d.Imm6&0xFFFF)

	case op.OpLUI:
		// Load upper: rd = imm6 << 10 (puts 6 bits in [15:10])
		writeReg = true
		writeVal = uint16((d.Imm6 & 0x3F) << 10)

	case op.OpLW:
		addr := rs1Val + uint16(int16(d.Imm6))
		writeReg = true
		writeVal = c.mem.Read16(addr)

	case op.OpSW:
		// SW rs1, imm6(rd): mem[rd + imm6] = rs1
		addr := rdVal + uint16(int16(d.Imm6))
		c.mem.Write16(addr, rs1Val)

	case op.OpLB:
		addr := rs1Val + uint16(int16(d.Imm6))
		writeReg = true
		writeVal = uint16(c.mem.Read8(addr))

	case op.OpSB:
		addr := rdVal + uint16(int16(d.Imm6))
		c.mem.Write8(addr, byte(rs1Val))

	case op.OpBEQ:
		// B-type: rs1 is in [11:9], rs2 in [8:6], off6 in [5:0]
		r1 := c.Regs[d.Rd]  // [11:9] is actually rs1 for B-type
		r2 := c.Regs[d.Rs1] // [8:6] is rs2 for B-type
		if r1 == r2 {
			nextPC = uint16(int(c.PC) + d.Imm6*2)
		}

	case op.OpBNE:
		r1 := c.Regs[d.Rd]
		r2 := c.Regs[d.Rs1]
		if r1 != r2 {
			nextPC = uint16(int(c.PC) + d.Imm6*2)
		}

	case op.OpBLT:
		r1 := c.Regs[d.Rd]
		r2 := c.Regs[d.Rs1]
		if int16(r1) < int16(r2) {
			nextPC = uint16(int(c.PC) + d.Imm6*2)
		}

	case op.OpBGE:
		r1 := c.Regs[d.Rd]
		r2 := c.Regs[d.Rs1]
		if int16(r1) >= int16(r2) {
			nextPC = uint16(int(c.PC) + d.Imm6*2)
		}

	case op.OpJAL:
		c.Regs[7] = nextPC // link register = R7
		nextPC = uint16(int(c.PC) + d.Off12*2)

	case op.OpSYSTEM:
		switch d.Func {
		case 0: // JALR rs1
			c.Regs[7] = nextPC
			nextPC = rs1Val
		case 1: // HALT
			c.Halt = true
		default: // SYSCALL (func=2) and NOP (func=7): no hardware effect.
			// I/O is performed by the OS in software via MMIO, not by a trap.
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
