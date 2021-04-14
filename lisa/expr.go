package lisa

import (
	"bytes"
	"fmt"
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
func getTermFromExpr(b []byte) (root *Term, err error) {
	root = &Term{}
	t := root
	for len(b) > 0 {
		c := b[0]
		b = b[1:]
		switch c {
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
