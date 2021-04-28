package cpu

import (
	"errors"
	"fmt"
	"go6502/ins"
	"time"
)

const (
	debug = true
	pss   = "CZIDB-VN"

	targetCycle = 98321791
	targetPC    = 0x3469
)

const (
	FlagCarry uint8 = 1 << iota
	FlagZero
	FlagIRQDisable
	FlagDecimalMode
	FlagBreak
	FlagUnused
	FlagOverflow
	FlagNegtive
)

type CPU struct {
	totalCycles  time.Duration
	durPerCycles time.Duration

	NMI   chan bool // edge trigger
	Reset chan bool // edge trigger
	IRQ   bool      // level trigger
	Mem   Mem

	irqVec uint16
	nmiVec uint16

	PC uint16
	AC uint8
	RX uint8
	RY uint8
	PS uint8
	IR uint8 // instruction register
	sp uint8
}

func New() (c *CPU) {
	c = &CPU{
		NMI:   make(chan bool),
		Reset: make(chan bool),
	}
	// pagetable.com/?p=410
	c.irqVec = 0xfffe
	c.nmiVec = 0xfffa
	c.reset()
	return
}

func (c *CPU) reset() {
	c.PC = 0xfffc
	c.sp = 0xfd
}

func (c *CPU) SP() uint16 {
	return uint16(c.sp) | uint16(0x100)
}

func (c *CPU) String() string {
	var t []byte
	for i := 0; i < 8; i++ {
		if c.PS&(1<<i) != 0 {
			t = append(t, pss[i])
		}
	}
	return fmt.Sprintf("[%d]A:%02X X:%02X Y:%02X SP:%04X PC:%04X PS:%s IR:%02X",
		c.totalCycles,
		c.AC, c.RX, c.RY,
		c.SP(), c.PC, t, c.IR)
}

func (c *CPU) Run(m Mem) (err error) {

	c.Mem = m
	var (
		cycles time.Duration
		i      ins.Ins
		prev   uint16
	)

	for {

		select {
		case <-c.NMI:
			c.pushWordToStack(c.PC)
			c.pushByteToStack(c.PS | FlagUnused)
			c.PC = c.Mem.ReadWord(c.nmiVec)
			c.PS |= FlagIRQDisable
		case <-c.Reset:
			c.reset()
		default:
		}

		if c.IRQ && c.PS&FlagIRQDisable == 0 {
			c.pushWordToStack(c.PC)
			c.pushByteToStack(c.PS | FlagUnused)
			c.PC = c.Mem.ReadWord(c.irqVec)
			c.PS |= FlagIRQDisable
		}

		prev = c.PC
		c.IR = m.ReadByte(c.PC)
		i = ins.Table[c.IR]
		if i.Cycles == 0 {
			return c.Fault("invalid op code")
		}
		if debug {
			fmt.Println(c, i)
		}
		c.PC += uint16(i.Bytes)

		cycles, err = c.insHandler(i)
		if err != nil {
			return
		}
		cycles += i.Cycles
		c.totalCycles += cycles
		// time.Sleep(cycles * c.durPerCycles)

		if debug {
			if c.totalCycles > targetCycle {
				break
			}

			if c.PC == targetPC {
				break
			}
		}

		// the instruction didn't change PC
		if c.PC == prev {
			return c.Fault("infinite loop")
		}

	}
	return
}

func (c *CPU) Fault(s string) error {
	return errors.New("FAULT:" + s)
}

func (c *CPU) setZN(x uint8) {
	if x == 0 {
		c.PS |= FlagZero
	} else {
		c.PS &= ^FlagZero
	}
	if x&FlagNegtive == 0 {
		c.PS &= ^FlagNegtive
	} else {
		c.PS |= FlagNegtive
	}
}

// simple version of {bytes,bufio}.Reader
type Mem interface {
	ReadByte(pc uint16) (b uint8)
	ReadWord(pc uint16) (n uint16)
	WriteByte(pc uint16, b uint8)
	WriteWord(pc uint16, n uint16)
}
