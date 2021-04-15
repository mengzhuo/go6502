package lisa

import (
	"bytes"
	"fmt"
	"go6502/ins"
)

type TermType byte

//go:generate stringer --type=TermType --trimprefix=T
const (
	TLabel TermType = iota + 1
	TOperator
	THex
	TDecimal
	TBinary
	TAscii
	TCurrentLine
)

// RIGHT -> LEFT
type Term struct {
	next     *Term
	Type     TermType
	operator byte
	order    byte
	Value    []byte
}

func (t *Term) String() string {
	if t.Type == TOperator {
		return string(t.Value)
	}
	return fmt.Sprintf("[%s]%s %s", string(t.order), t.Type, string(t.Value))
}

// Tokenize
func parseOperand(b []byte) (mode ins.Mode, root *Term, err error) {
	root = &Term{}
	t := root
	indirect := uint16(1) // (:2 ):4
	for len(b) > 0 {
		c := b[0]
		b = b[1:]
		switch c {
		case ',':
			if len(b) == 1 {
				switch b[0] {
				case 'Y', 'X':
					indirect |= uint16(b[0]) << 8
					b = b[1:]
				default:
					err = fmt.Errorf("should end with X, Y, got=%s", string(b))
				}
				continue
			}
			if len(b) == 2 && string(b) == "X)" {
				indirect <<= 1
				indirect |= uint16('X') << 8
				b = b[2:]
				continue
			}

			err = fmt.Errorf("should end with X or X) or Y, got=%s", string(b))
			return
		case OpLeftBracket:
			// should be first
			if t.Type == 0 {
				indirect <<= 1
			}
		case OpRightBracket:
			// should not be first
			if t.Type != 0 {
				indirect <<= 1
			}
		case OpLowOrder:
			if t.order != 0 {
				err = fmt.Errorf("duplicate order")
				return
			}
			t.order = c
		case OpAdd, OpMinus, OpOr, OpXor, OpAnd, OpDivision, OpAsterisk:
			if len(t.Value) == 0 && c == OpMinus {
				t.Value = append(t.Value, c)
				continue
			}

			if len(t.Value) == 0 && t.Type == 0 && c == OpAsterisk {
				t.Value = append(t.Value, c)
				t.Type = TCurrentLine
				continue
			}

			if len(t.Value) == 0 && t.order == 0 {
				t.order = c
				continue
			}

			t.next = &Term{Type: TOperator, Value: []byte{c}, next: &Term{}}
			t = t.next.next
		case OpHex:
			if t.Type != 0 {
				err = fmt.Errorf("duplicate type")
				return
			}
			t.Type = THex
		case OpDecimal:
			if t.Type != 0 {
				err = fmt.Errorf("duplicate type")
				return
			}
			t.Type = TDecimal
		case OpBinary:
			if t.Type != 0 {
				err = fmt.Errorf("duplicate type")
				return
			}
			t.Type = TBinary
		default:
			if t.Type == 0 {
				if c >= '0' && c <= '9' {
					t.Type = THex
				} else {
					t.Type = TLabel
				}
			}
			t.Value = append(t.Value, c)
		}
	}
	switch indirect {
	case uint16('X')<<8 | 4:
		mode = ins.IndirectX
	case uint16('Y')<<8 | 4:
		mode = ins.IndirectY
	case 4:
		mode = ins.Indirect
	case 1 | uint16('X')<<8:
		mode = ins.ZeroPageX
	case 1 | uint16('Y')<<8:
		mode = ins.ZeroPageY
	}

	return
}

func syntaxCheck(root *Term) (err error) {
	for e := root; e != nil; e = e.next {
		switch e.Type {
		case THex:
			if len(e.Value) == 0 {
				return fmt.Errorf("empty hex")
			}
			for _, c := range e.Value {
				if bytes.IndexByte(hex, c) == -1 {
					return fmt.Errorf("invalid hex %x", c)
				}
			}
		case TDecimal:
			if len(e.Value) == 0 {
				return fmt.Errorf("empty decimal")
			}
			for _, c := range e.Value {
				if c < '0' || c > '9' {
					return fmt.Errorf("invalid decimal %x", c)

				}
			}
		case TOperator:
			if e.next == nil || e.next.Type == TOperator || e.next.Type == 0 {
				return fmt.Errorf("expect valid term for op %s", e.Value)
			}
		}
	}
	return
}
