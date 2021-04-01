package cpu

import (
	"go6502/ins"
	"time"
)

func (c *CPU) insHandler(i ins.Ins) (cycles time.Duration) {

	var addr uint16
	var oper uint8

	switch i.Name {
	default:
		panic("invalid name")
	case ins.NOP:
		return
	case ins.LDA:
		c.AC, cycles = c.getOper(i.Mode)
		c.setZN(c.AC)
	case ins.LDX:
		c.RX, cycles = c.getOper(i.Mode)
		c.setZN(c.RX)
	case ins.LDY:
		c.RY, cycles = c.getOper(i.Mode)
		c.setZN(c.RY)
	case ins.STA:
		addr, cycles = c.getAddr(i.Mode)
		c.Mem.WriteByte(addr, c.AC)
	case ins.STX:
		addr, cycles = c.getAddr(i.Mode)
		c.Mem.WriteByte(addr, c.RX)
	case ins.STY:
		addr, cycles = c.getAddr(i.Mode)
		c.Mem.WriteByte(addr, c.RY)

	// Stack Operations
	case ins.TSX:
		c.RX = uint8(c.SP & 0xff)
		c.setZN(c.RX)
	case ins.TXS:
		c.SP = uint16(c.RX)
	case ins.PHA:
		c.pushByteToStack(c.AC)
	case ins.PHP:
		c.pushByteToStack(c.PS | FlagBreak | FlagUnused)
	case ins.PLA:
		c.AC = c.popByteFromStack()
		c.setZN(c.AC)
	case ins.PLP:
		c.PS = c.popByteFromStack()
		c.PS &= ^(FlagBreak | FlagUnused)

	// Register Transfer
	case ins.TAX:
		c.RX = c.AC
		c.setZN(c.RX)
	case ins.TAY:
		c.RY = c.AC
		c.setZN(c.RY)
	case ins.TXA:
		c.AC = c.RX
		c.setZN(c.AC)
	case ins.TYA:
		c.AC = c.RY
		c.setZN(c.AC)

	// Logical
	case ins.AND:
		oper, cycles = c.getOper(i.Mode)
		c.AC &= oper
		c.setZN(c.AC)
	case ins.ORA:
		oper, cycles = c.getOper(i.Mode)
		c.AC |= oper
		c.setZN(c.AC)
	case ins.EOR:
		oper, cycles = c.getOper(i.Mode)
		c.AC ^= oper
		c.setZN(c.AC)
	case ins.BIT:
		oper, cycles = c.getOper(i.Mode)
		value := c.AC & oper
		c.setZN(value)

		if value&FlagOverflow != 0 {
			c.PS |= FlagOverflow
		} else {
			c.PS &= ^FlagOverflow
		}

	// Arithmetic
	case ins.ADC, ins.SBC:
		oper, cycles = c.getOper(i.Mode)
		if i.Name == ins.SBC {
			oper ^= 0
		}

		if c.PS&FlagDecimalMode != 0 {
			panic("no decimal mode")
			return
		}

		sameSigned := (c.AC^oper)&FlagNegtive == 0
		sum := uint16(c.AC)
		sum += uint16(oper)
		if c.PS&FlagCarry != 0 {
			sum += 1
		}
		c.AC = uint8(sum & 0xff)
		c.setZN(c.AC)
		if sum > 0xff {
			c.PS |= FlagCarry
		} else {
			c.PS &= ^FlagCarry
		}

		if sameSigned && ((c.AC^oper)&FlagNegtive != 0) {
			c.PS |= FlagOverflow
		} else {
			c.PS &= ^FlagOverflow
		}
	case ins.CMP:
		oper, cycles = c.getOper(i.Mode)
		if c.AC > oper {
			c.PS |= FlagCarry
		} else {
			c.PS &= ^FlagCarry
		}
		v := c.AC - oper
		c.setZN(v)
	case ins.CPX:
		oper, cycles = c.getOper(i.Mode)
		if c.RX > oper {
			c.PS |= FlagCarry
		} else {
			c.PS &= ^FlagCarry
		}
		v := c.RX - oper
		c.setZN(v)
	case ins.CPY:
		oper, cycles = c.getOper(i.Mode)
		if c.RY > oper {
			c.PS |= FlagCarry
		} else {
			c.PS &= ^FlagCarry
		}
		v := c.RY - oper
		c.setZN(v)

	// Increments & Decrements
	case ins.INC, ins.DEC:
		addr, cycles = c.getAddr(i.Mode)
		oper := c.Mem.ReadByte(addr)
		if i.Name == ins.DEC {
			oper -= 1
		} else {
			oper += 1
		}
		c.setZN(oper)
		c.Mem.WriteByte(addr, oper)
	case ins.INX:
		c.RX++
		c.setZN(c.RX)
	case ins.INY:
		c.RY++
		c.setZN(c.RY)
	case ins.DEX:
		c.RX--
		c.setZN(c.RX)
	case ins.DEY:
		c.RY--
		c.setZN(c.RY)

	// Shifts
	case ins.ASL:
	case ins.LSR:
	case ins.ROL:
	case ins.ROR:

	// Jumps & Calls
	case ins.JMP:
	case ins.JSR:
	case ins.RTS:

	// Branches
	case ins.BCC:
	case ins.BCS:
	case ins.BEQ:
	case ins.BMI:
	case ins.BNE:
	case ins.BPL:
	case ins.BVC:
	case ins.BVS:

	// Status Flag Changes
	case ins.CLC:
		c.PS &= ^FlagCarry
	case ins.CLD:
		c.PS &= ^FlagDecimalMode
	case ins.CLI:
		c.PS &= ^FlagIRQDisable
	case ins.CLV:
		c.PS &= ^FlagOverflow
	case ins.SEC:
		c.PS |= FlagCarry
	case ins.SED:
		c.PS |= FlagDecimalMode
	case ins.SEI:
		c.PS |= FlagIRQDisable

	// System Functions
	case ins.BRK:
	case ins.RTI:
	}
	return
}

func (c *CPU) getAddr(mode ins.Mode) (addr uint16, cycles time.Duration) {
	m := c.Mem
	switch mode {
	case ins.Immediate:
		addr = c.PC + 1
	case ins.ZeroPage:
		addr = uint16(m.ReadByte(c.PC + 1))
	case ins.ZeroPageX:
		addr = uint16(m.ReadByte(c.PC + 1))
		addr += uint16(c.RX)
	case ins.ZeroPageY:
		addr = uint16(m.ReadByte(c.PC + 1))
		addr += uint16(c.RY)
	case ins.Absolute:
		addr = m.ReadWord(c.PC + 1)
	case ins.AbsoluteX:
		addr = m.ReadWord(c.PC + 1)
		addrx := uint16(c.RX) + addr
		if addr^addrx>>8 != 0 {
			cycles += 1
		}
		addr = addrx
	case ins.AbsoluteY:
		addr = m.ReadWord(c.PC + 1)
		addry := uint16(c.RY) + addr
		if addr^addry>>8 != 0 {
			cycles += 1
		}
		addr = addry
	case ins.IndirectX:
		zaddr := m.ReadByte(c.PC + 1)
		addr = uint16(c.RX + zaddr)
		// effective addr
		addr = m.ReadWord(addr)
	case ins.IndirectY:
		addr = uint16(m.ReadByte(c.PC + 1))
		// effective addr
		addr = m.ReadWord(addr)
		addry := addr + uint16(c.RY)
		if (addr^addry)>>8 != 0 {
			cycles += 1
		}
		addr = addry
	}
	return
}

func (c *CPU) getOper(mode ins.Mode) (oper uint8, cycles time.Duration) {
	var addr uint16
	addr, cycles = c.getAddr(mode)
	oper = c.Mem.ReadByte(addr)
	return
}

func (c *CPU) pushByteToStack(value uint8) {
	addr := c.SP | 0x100
	c.Mem.WriteByte(addr, value)
	c.SP -= 1
}

func (c *CPU) popByteFromStack() uint8 {
	c.SP += 1
	addr := c.SP | 0x100
	return c.Mem.ReadByte(addr)
}

func (c *CPU) pushWordToStack(v uint16) {
	c.pushByteToStack(uint8(v >> 8))
	c.pushByteToStack(uint8(v & 0xff))
}

func (c *CPU) popWordFromStack() (r uint16) {
	r = c.Mem.ReadWord((c.SP | 0x100) + 1)
	c.SP += 2
	return
}
