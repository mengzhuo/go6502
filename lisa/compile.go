package lisa

import (
	"go6502/ins"
	"io"
)

type Compiler struct {
	OL     []*Obj //object list
	Origin uint16
}

type Obj struct {
	stmt *Stmt
	ins  ins.Ins
	op   uint16
	done bool
}

func Compile(sl []*Stmt, of io.Writer) (err error) {

	c := &Compiler{}
	// clear all comment , implied
	for _, s := range sl {
		if s.Mnemonic == 0 {
			continue
		}

		obj := &Obj{stmt: s}
		c.OL = append(c.OL, obj)

		i := ins.GetNameTable(obj.stmt.Mnemonic.String(), ins.Implied)
		if i.Mode == ins.Implied && i.Bytes == 1 {
			obj.ins = i
			continue
		}
		c.evalue(obj)
	}
	return
}

func (c *Compiler) evalue(obj *Obj) (err error) {
	if obj.done {
		return
	}

	s := obj.stmt
	if s.Expr == nil {
		return s.NE("expecting operand, got null")
	}

	if s.Expr.next == nil {
		switch s.Expr.Type {
		case THex, TBinary, TDecimal:
			obj.op, err = s.Expr.Uint16()
			return
		case TOperator:
			return s.NE("only operator left")
		case TGTLabel:
		}
	}
	return
}
