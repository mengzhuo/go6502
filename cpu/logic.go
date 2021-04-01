package cpu

import (
	"go6502/ins"
	"time"
)

func (c *CPU) insHandler(i ins.Ins) (cycles time.Duration) {

	switch i.Name {
	default:
		panic("invalid name")
	case ins.NOP:
		return
	case ins.LDA:
		c.AC, cycles = c.getOper(i.Mode)
		c.SetZN(c.AC)
	case ins.LDX:
		c.RX, cycles = c.getOper(i.Mode)
		c.SetZN(c.RX)
	case ins.LDY:
		c.RY, cycles = c.getOper(i.Mode)
		c.SetZN(c.RY)
	case ins.STA:
		cycles = c.setByte(i.Mode, c.AC)
	case ins.STX:
		cycles = c.setByte(i.Mode, c.RX)
	case ins.STY:
		cycles = c.setByte(i.Mode, c.RY)
	}
	return
}

func (c *CPU) setByte(mode ins.Mode, value uint8) (cycles time.Duration) {
	var addr uint16
	addr, cycles = c.getAddr(mode)
	c.Mem.WriteByte(addr, value)
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
