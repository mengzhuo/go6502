package cpu

import (
	"errors"
	"go6502/ins"
	"time"
)

type CPU struct {
	totalCycles  uint64
	durPerCycles time.Duration

	NMI chan bool // edge trigger
	IRQ bool      // level trigger

	irqVec uint16
	nmiVec uint16

	AC uint8
	RX uint8
	RY uint8
	PS uint8
	PC uint16
	SP uint16
}

func (c *CPU) Run(m Mem) (err error) {

	var (
		op     uint8
		cycles uint8
		ins    ins.Ins
	)

	for {
		op, cycles, err = m.ReadByte(c.PC)
		if err != nil {
			return
		}
		ins = ins.Table[op]
		if ins.Cycles == 0 {
			return c.Fault("invalid Op code")
		}
		cycles += ins.Cycles
		cycles += insHandler[op](ins, m)
		time.Sleep(time.Duration(cycles) * c.durPerCycles)

		select {
		case <-c.NMI:
			c.PC = c.nmiVec
			continue
		default:
		}

		if c.IRQ && c.PS&FlagDisableIRQ == 0 {
			c.PC = c.irqVec
			c.PS |= FlagDisableIRQ
		}
	}

}

func (c *CPU) Fault(s string) error {
	return errors.New("FAULT:" + s)
}

// simple version of {bytes,bufio}.Reader
type Mem interface {
	ReadByte(pc uint16) (b uint8, cycles int64, err error)
	ReadWord(pc uint16) (n uint16, cycles int64, err error)
	WriteByte(b uint8, pc uint16) (cycles int64, err error)
	WriteWord(n uint16, pc uint16) (cycles int64, err error)
}
