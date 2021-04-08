package ins

import (
	"fmt"
	"sync"
	"time"
)

// http://obelisk.me.uk/6502/instructions.html

//go:generate stringer -type=Mode
type Mode uint8

const (
	Implied Mode = 1 + iota
	Accumulator
	Immediate
	ZeroPage
	ZeroPageX
	ZeroPageY
	Relative
	Absolute
	AbsoluteX
	AbsoluteY
	Indirect
	IndirectX
	IndirectY
)

//go:generate stringer -type=Name
type Name uint8

const (
	ADC Name = 1 + iota
	AND
	ASL
	BCC
	BCS
	BEQ
	BIT
	BMI
	BNE
	BPL
	BRK
	BVC
	BVS
	CLC
	CLD
	CLI
	CLV
	CMP
	CPX
	CPY
	DEC
	DEX
	DEY
	EOR
	INC
	INX
	INY
	JMP
	JSP
	JSR
	LDA
	LDX
	LDY
	LSR
	NOP
	ORA
	PHA
	PHP
	PLA
	PLP
	ROL
	ROR
	RTI
	RTS
	SBC
	SEC
	SED
	SEI
	STA
	STX
	STY
	TAX
	TAY
	TSX
	TXA
	TXS
	TYA
)

type Flag uint8

//go:generate stringer --type=PagePolicy
type PagePolicy uint8

const (
	None   PagePolicy = 0
	Cross             = 1
	Branch            = 2
)

//go:generate python3 gen.py
type Ins struct {
	Cycles time.Duration // base cycle count
	Name   Name
	Op     uint8
	Mode   Mode
	Affect Flag
	Bytes  uint8
	Page   PagePolicy
}

func (i Ins) String() string {
	if i.Mode == Implied {
		return fmt.Sprintf("%s", i.Name)
	}
	return fmt.Sprintf("%s_%s", i.Name, i.Mode)
}

var (
	Table        = [256]Ins{}
	nameCache    = map[string][]Ins{}
	nameCacheGen = &sync.Once{}
)

func GetNameTable(name string, mode Mode) Ins {
	nameCacheGen.Do(func() {
		for idx := range Table {
			i := Table[idx]
			nameCache[i.Name.String()] = append(nameCache[i.Name.String()], i)
		}
	})

	for _, i := range nameCache[name] {
		if i.Mode == mode {
			return i
		}
	}
	return Ins{}
}
