package cpu

import (
	"fmt"
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
		if c.AC >= oper {
			c.PS |= FlagCarry
		} else {
			c.PS &= ^FlagCarry
		}
		v := c.AC - oper
		c.setZN(v)
	case ins.CPX:
		oper, cycles = c.getOper(i.Mode)
		if c.RX >= oper {
			c.PS |= FlagCarry
		} else {
			c.PS &= ^FlagCarry
		}
		v := c.RX - oper
		c.setZN(v)
	case ins.CPY:
		oper, cycles = c.getOper(i.Mode)
		if c.RY >= oper {
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
		if i.Mode == ins.Accumulator {
			if c.AC&(1<<7) != 0 {
				c.PS |= FlagCarry
			} else {
				c.PS &= ^FlagCarry
			}
			c.AC <<= 1
			c.setZN(c.AC)
			return
		}
		addr, cycles = c.getAddr(i.Mode)
		oper = c.Mem.ReadByte(addr)
		if oper&(1<<7) != 0 {
			c.PS |= FlagCarry
		} else {
			c.PS &= ^FlagCarry
		}
		oper <<= 1
		c.setZN(oper)
		c.Mem.WriteByte(addr, oper)
	case ins.LSR:
		if i.Mode == ins.Accumulator {
			if c.AC&1 != 0 {
				c.PS |= FlagCarry
			} else {
				c.PS &= ^FlagCarry
			}
			c.AC >>= 1
			c.setZN(c.AC)
			return
		}
		addr, cycles = c.getAddr(i.Mode)
		oper = c.Mem.ReadByte(addr)
		if oper&1 != 0 {
			c.PS |= FlagCarry
		} else {
			c.PS &= ^FlagCarry
		}
		oper >>= 1
		c.setZN(oper)
		c.Mem.WriteByte(addr, oper)
	case ins.ROL:
		var more uint8
		if i.Mode == ins.Accumulator {
			if c.AC&(1<<7) != 0 {
				c.PS |= FlagCarry
				more = 1
			} else {
				c.PS &= ^FlagCarry
			}
			c.AC <<= 1
			c.AC |= more
			c.setZN(c.AC)
			return
		}
		addr, cycles = c.getAddr(i.Mode)
		oper = c.Mem.ReadByte(addr)
		if oper&(1<<7) != 0 {
			more = 1
			c.PS |= FlagCarry
		} else {
			c.PS &= ^FlagCarry
		}
		oper <<= 1
		oper |= more
		c.setZN(oper)
		c.Mem.WriteByte(addr, oper)
	case ins.ROR:
		var more uint8
		if i.Mode == ins.Accumulator {
			if c.AC&1 != 0 {
				c.PS |= FlagCarry
				more |= 1 << 7
			} else {
				c.PS &= ^FlagCarry
			}
			c.AC >>= 1
			c.AC |= more
			c.setZN(c.AC)
			return
		}
		addr, cycles = c.getAddr(i.Mode)
		oper = c.Mem.ReadByte(addr)
		if oper&1 != 0 {
			c.PS |= FlagCarry
			more |= 1 << 7
		} else {
			c.PS &= ^FlagCarry
		}
		oper >>= 1
		oper |= more
		c.setZN(oper)
		c.Mem.WriteByte(addr, oper)

	// Jumps & Calls
	case ins.JMP:
		addr, cycles = c.getAddr(i.Mode)
		c.PC = addr
	case ins.JSR:
		addr, cycles = c.getAddr(i.Mode)
		c.pushWordToStack(c.PC + 2)
		c.PC = addr
	case ins.RTS:
		addr = c.popWordFromStack()
		c.PC = addr + 1

	// Branches
	case ins.BCC:
		cycles = c.BranchIf(i, FlagCarry, false)
	case ins.BCS:
		cycles = c.BranchIf(i, FlagCarry, true)
	case ins.BEQ:
		cycles = c.BranchIf(i, FlagZero, true)
	case ins.BMI:
		cycles = c.BranchIf(i, FlagNegtive, true)
	case ins.BNE:
		cycles = c.BranchIf(i, FlagZero, false)
	case ins.BPL:
		cycles = c.BranchIf(i, FlagNegtive, false)
	case ins.BVC:
		cycles = c.BranchIf(i, FlagOverflow, false)
	case ins.BVS:
		cycles = c.BranchIf(i, FlagOverflow, true)

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
		c.pushWordToStack(c.PC + 1)
		c.pushByteToStack(c.PS)
		c.PC = c.Mem.ReadWord(c.irqVec)
		c.PS |= FlagBreak | FlagIRQDisable
		fmt.Println("BRK", c)
		panic("ok")
	case ins.RTI:
		c.PS = c.popByteFromStack()
		c.PC = c.popWordFromStack()
	}
	return
}

func (c *CPU) BranchIf(i ins.Ins, flag uint8, set bool) (cycles time.Duration) {
	var ok bool
	if set {
		ok = flag&c.PS != 0
	} else {
		ok = flag&c.PS == 0
	}

	if !ok {
		c.PC += uint16(i.Bytes)
		cycles = 1
		return
	}

	oper := c.Mem.ReadByte(c.PC + 1)
	cycles = 1
	if oper&FlagNegtive == 0 {
		c.PC += uint16(oper) + 2
	} else {
		c.PC -= uint16(0xff^oper) - 1
	}

	old := c.PC
	// cross page
	if (old^c.PC)>>8 != 0 {
		cycles += 1
	}
	return
}

func (c *CPU) getAddr(mode ins.Mode) (addr uint16, cycles time.Duration) {
	m := c.Mem
	switch mode {
	default:
		panic("no handler")
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
	case ins.Indirect:
		addr = m.ReadWord(c.PC + 1)
		addr = m.ReadWord(addr)
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
