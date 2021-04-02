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

	targetCycle = 500000
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
	totalCycles  uint64
	durPerCycles time.Duration

	NMI chan bool // edge trigger
	IRQ bool      // level trigger
	Mem Mem

	irqVec uint16
	nmiVec uint16

	AC uint8
	RX uint8
	RY uint8
	PS uint8
	PC uint16
	SP uint16
}

func New() *CPU {
	return &CPU{
		NMI:    make(chan bool),
		irqVec: 0xFFFE,
		nmiVec: 0xfffd,
		PC:     0xFFFE,
		SP:     0x0000,
	}
}

func (c *CPU) String() string {
	var t []byte
	for i := 0; i < 8; i++ {
		if c.PS&(1<<i) != 0 {
			t = append(t, pss[i])
		}
	}
	return fmt.Sprintf("A:%02X X:%02X Y:%02X SP:%04X PC:%04X PS:%s",
		c.AC, c.RX, c.RY,
		c.SP|0x100, c.PC, t)
}

func (c *CPU) Run(m Mem) (err error) {

	defer func() {
		err := recover()
		if err != nil {
			fmt.Println(err, c.String())
		}
	}()

	c.Mem = m
	var (
		op     uint8
		cycles time.Duration
		i      ins.Ins
		total  time.Duration
		prev   uint16
	)

	for {
		prev = c.PC
		op = m.ReadByte(c.PC)
		i = ins.Table[op]
		if i.Cycles == 0 {
			return c.Fault("invalid Op code")
		}
		if debug {
			fmt.Println(c, i)
		}
		cycles = i.Cycles
		cycles += c.insHandler(i)
		total += cycles
		time.Sleep(cycles * c.durPerCycles)
		if debug && total > targetCycle {
			break
		}

		// the instruction didn't change PC
		switch i.Name {
		case ins.BCC, ins.BCS, ins.BEQ, ins.BMI,
			ins.BNE, ins.BPL, ins.BVC, ins.BVS, ins.JMP,
			ins.JSR, ins.BRK, ins.RTI, ins.RTS:
			if c.PC == prev {
				return c.Fault("infinite loop")
			}
		default:
			c.PC += uint16(i.Bytes)
		}

		select {
		case <-c.NMI:
			c.PC = c.nmiVec
			continue
		default:
		}

		if c.IRQ && c.PS&FlagIRQDisable == 0 {
			c.PC = c.irqVec
			c.PS |= FlagIRQDisable
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
