package cpu

import (
	"errors"
	"fmt"
	"go6502/ins"
	"log"
	"time"
)

const (
	debug = false
	pss   = "CZIDB-VN"

	targetCycle = 98321791
	targetPC    = 0x406
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
	DurPerCycles time.Duration

	NMI   chan bool // edge trigger
	Reset chan bool // edge trigger
	IRQ   uint8     // level trigger zero means no irq
	Mem   Mem

	log *log.Logger

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

func New(log *log.Logger) (c *CPU) {
	c = &CPU{
		NMI:   make(chan bool, 1024),
		Reset: make(chan bool),
		log:   log,
	}
	// pagetable.com/?p=410
	c.irqVec = 0xfffe
	c.nmiVec = 0xfffa
	return
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

func (c *CPU) ResetF(m Mem) {
	c.PC = m.ReadWord(0xfffc)
	c.sp = 0xfd
	c.PS = 0
}

func (c *CPU) Run(m Mem) (err error) {

	c.Mem = m
	var (
		cycles time.Duration
		i      ins.Ins
		prev   uint16
	)

	for {
		// check interupt
		select {
		case <-c.NMI:
			c.pushWordToStack(c.PC)
			c.pushByteToStack(c.PS | FlagUnused)
			c.PC = c.Mem.ReadWord(c.nmiVec)
			c.PS |= FlagIRQDisable
		case <-c.Reset:
			c.ResetF(m)
		default:
		}

		if c.IRQ != 0 && c.PS&FlagIRQDisable == 0 {
			c.pushWordToStack(c.PC)
			c.pushByteToStack(c.PS | FlagUnused)
			c.PC = c.Mem.ReadWord(c.irqVec)
			c.PS |= FlagIRQDisable
		}

		prev = c.PC
		//  Fetch Instruction
		c.IR = m.ReadByte(c.PC)
		// Decode
		i = ins.Table[c.IR]
		if i.Cycles == 0 {
			return c.Fault("invalid op code")
		}

		if c.log != nil {
			c.log.Println(c, i)
		}

		// Execute
		c.PC += uint16(i.Bytes)
		cycles, err = c.insHandler(i)
		if err != nil {
			return
		}
		cycles += i.Cycles
		c.totalCycles += cycles
		time.Sleep(cycles * c.DurPerCycles)

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
