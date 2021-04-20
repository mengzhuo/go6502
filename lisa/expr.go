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
	Type     TermType
	Stmt     *Stmt
	next     *Term
	Value    []byte
	Oprand   uint16
	Resolved bool
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
		return
	}
	u = uint16(u64)
	if err == nil {
		t.Resolved = true
		t.Oprand = u
	}
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
		zeroPage     uint16 = 1
		indirectFlag uint16 = 4
	)

	root := &Term{}
	t := root
	indirect := zeroPage // (:2 ):4

	for col := 0; len(b) > 0; col += 1 {
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
					err = fmt.Errorf("Col:%d should end with X, Y, got=%s", col, string(b))
				}
				continue
			}
			if len(b) == 2 && string(b) == "X)" {
				indirect <<= 1
				indirect |= uint16('X') << 8
				b = b[2:]
				continue
			}

			err = fmt.Errorf("Col:%d should end with X, X) or Y, got=%s", col, string(b))
			return
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

	switch indirect {
	case uint16('X')<<8 | indirectFlag:
		mode = ins.IndirectX
	case uint16('Y')<<8 | indirectFlag:
		mode = ins.IndirectY
	case indirectFlag:
		mode = ins.Indirect
	case zeroPage | uint16('X')<<8:
		mode = ins.ZeroPageX
	case zeroPage | uint16('Y')<<8:
		mode = ins.ZeroPageY
	case 2:
		err = fmt.Errorf("bracket not closed")
	}

	for e := root; e != nil; e = e.next {
		ae = append(ae, e)
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
