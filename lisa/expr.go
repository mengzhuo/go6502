package lisa

import (
	"bytes"
	"fmt"
	"go6502/ins"
	"strconv"
	"strings"
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
	TLSLabel
	TGTLabel
	TRaw
)

type Expression []*Term

func (e Expression) String() string {
	if len(e) == 0 {
		return "<EMPTY>"
	}
	var buf strings.Builder
	for _, t := range e {
		buf.WriteString(t.String())
		buf.WriteByte(' ')
	}
	return buf.String()
}

// RIGHT -> LEFT
type Term struct {
	Type  TermType
	Value []byte
}

func (t *Term) Uint16() (u uint16, err error) {

	var u64 uint64
	switch t.Type {
	case THex:
		u64, err = strconv.ParseUint(string(t.Value), 16, 16)
	case TBinary:
		u64, err = strconv.ParseUint(string(t.Value), 2, 16)
	case TDecimal:
		u64, err = strconv.ParseUint(string(t.Value), 10, 16)
	default:
		err = fmt.Errorf("unsupported type:%s", t.Type)
		return
	}
	u = uint16(u64)
	return
}

func (t *Term) String() string {
	if t.Type == TOperator {
		return string(t.Value)
	}
	return fmt.Sprintf("<%s %s>", t.Type, string(t.Value))
}

// Tokenize
func parseOperand(b []byte) (mode ins.Mode, ae Expression, order byte, err error) {
	const (
		direct       uint16 = 1
		indirectFlag uint16 = 4
	)
	if len(b) == 0 {
		// at least is Implied
		mode = ins.Implied
		return
	}

	if len(b) == 1 && (b[0] == 'A' || b[0] == 'a') {
		mode = ins.Accumulator
		return
	}

	// default Absolute
	mode = ins.Absolute

	t := &Term{}
	indirect := direct // (:2 ):4

	for col := 0; len(b) > 0; col += 1 {
		c := b[0]
		b = b[1:]
		switch c {
		case ' ':
		case ',':
			b = bytes.TrimSpace(b)
			if len(b) == 1 {
				switch b[0] {
				case 'Y', 'X':
					indirect |= uint16(b[0]) << 8
					b = b[1:]
				default:
					err = fmt.Errorf("Col:%d should end with X, Y, got=%q", col, string(b))
				}
				continue
			}
			if len(b) == 2 && string(b) == "X)" {
				indirect <<= 1
				indirect |= uint16('X') << 8
				b = b[2:]
				continue
			}

			err = fmt.Errorf("Col:%d should end with X, X) or Y, got=%q", col, string(b))
			return
		case OpApostrophe, OpQuote:
			// looking for pair
			i := bytes.IndexByte(b, c)
			if i == -1 {
				err = fmt.Errorf("can't find pair for=%s %s", string(c), b)
				return
			}
			if i <= 1 {
				continue
			}
			t.Type = TRaw
			t.Value = append(t.Value, b[:i]...)
			b = b[i+1:]

		case OpLeftBracket:
			// should be first
			if t.Type == 0 {
				indirect <<= 1
			}
		case OpRightBracket:
			// should be  closed
			if t.Type == 0 {
				err = fmt.Errorf("Col:%d invalid right bracket", col)
				return
			}
			if indirect != 2 {
				err = fmt.Errorf("Col:%d no matched bracket", col)
				return
			}
			indirect <<= 1
		case OpLowOrder:
			if t.Type != 0 {
				err = fmt.Errorf("order should be the first byte of expr")
				return
			}
			if order != 0 {
				err = fmt.Errorf("duplicate order")
				return
			}
			order = c

		case OpAdd, OpMinus, OpOr, OpXor, OpAnd, OpDivision, OpAsterisk:
			if len(t.Value) == 0 && c == OpMinus {
				t.Value = append(t.Value, c)
				continue
			}

			if len(t.Value) == 0 && t.Type == 0 {
				if c == OpAsterisk {
					t.Value = append(t.Value, c)
					t.Type = TCurrentLine
					mode = ins.Relative
					continue
				}

				if c == OpDivision {
					if order != 0 {
						err = fmt.Errorf("duplicate order")
						return
					}
					order = OpDivision
					continue
				}
			}
			ae = append(ae, t)
			ae = append(ae, &Term{Type: TOperator, Value: []byte{c}})
			t = &Term{}
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
		case OpLSLabel:
			if t.Type != 0 {
				err = fmt.Errorf("duplicate type")
				return
			}
			t.Type = TLSLabel
		case OpGTLabel:
			if t.Type != 0 {
				err = fmt.Errorf("duplicate type")
				return
			}
			t.Type = TGTLabel
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
	ae = append(ae, t)

	switch indirect {
	case uint16('X')<<8 | indirectFlag:
		mode = ins.IndirectX
	case uint16('Y')<<8 | indirectFlag:
		mode = ins.IndirectY
	case indirectFlag:
		mode = ins.Indirect
	case direct | uint16('X')<<8:
		mode <<= 1
	case direct | uint16('Y')<<8:
		mode <<= 2
	case direct:
		switch order {
		case '#', '/':
			mode = ins.Immediate
		case 0:
		default:
			err = fmt.Errorf("invalid order %s", string(order))
		}
	case 2:
		err = fmt.Errorf("bracket not closed")
	}

	return
}

func syntaxCheck(expr Expression) (err error) {

	for i, e := range expr {
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
			if i >= len(expr) {
				return fmt.Errorf("expect valid term for op %s", e.Value)
			}
			next := expr[i+1]
			if next.Type == TOperator || next.Type == 0 {
				return fmt.Errorf("expect valid term for op %s", e.Value)
			}
		}
	}
	return
}
