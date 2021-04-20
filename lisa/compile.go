package lisa

import (
	"fmt"
	"go6502/ins"
	"io"
	"strings"
)

type Prog struct {
	Stmt *Stmt // origin statement

	// for final linkage
	Done    bool
	Ins     ins.Ins
	Operand uint16
	Addr    uint16
}

type Compiler struct {
	Labels  map[string]*Prog //symbol table
	Symbols []*Prog
	Origin  uint16
}

func (c *Compiler) evalueExpr(ex Expression) (err error) {
	for i := len(ex) - 1; i >= 0; i++ {
		e := ex[i]
		switch e.Type {
		case THex, TDecimal, TBinary:
		}
	}
	return
}

func (c *Compiler) init(sl []*Stmt) (err error) {
	// clear all comment, build symbol table, guess prog first time
	for _, s := range sl {
		if s.Mnemonic == 0 {
			continue
		}

		p := &Prog{
			Stmt: s,
		}
		c.Symbols = append(c.Symbols, p)
		if s.Label != "" && !strings.HasPrefix(s.Label, "^") {
			c.Labels[s.Label] = p
		}
	}
	return
}

func Compile(sl []*Stmt, of io.Writer) (err error) {

	c := &Compiler{
		Labels: make(map[string]*Prog),
	}

	err = c.init(sl)
	if err != nil {
		return
	}
	resolved := false
	limit := 10
	i := 0
	for ; !resolved && i < limit; i++ {
		resolved, err = c.resovleEqualLabel()
		if err != nil {
			return
		}
	}
	if i == limit {
		err = fmt.Errorf("over resolve limit")
		return
	}
	err = c.Evalue()
	if err != nil {
		return
	}
	fmt.Println("Finaly step")
	for _, p := range c.Symbols {
		fmt.Println(p.Stmt)
	}
	if err = c.encode(); err != nil {
		return
	}
	err = c.link(of)
	return
}

func (c *Compiler) encode() (err error) {
	// first we have to look for ORG
	for _, p := range c.Symbols {
		if p.Stmt.Mnemonic != ORG {
			continue
		}
		c.Origin = p.Operand
		break
	}
	for _, p := range c.Symbols {
		switch p.Stmt.Mnemonic {
		case STR, ASC, EQU, EPZ:
			continue
		}

		i := ins.GetNameTable(p.Stmt.Mnemonic.String(), ins.Implied)
		if i.Bytes == 1 {
			if p.Stmt.Oper != "" {
				err = fmt.Errorf("implied instruction has operand")
				return
			}
			p.Ins = i
			p.Done = true
			p.Addr = c.Origin
			c.Origin += uint16(i.Bytes)
			continue
		}
	}
	return
}

func (c *Compiler) link(of io.Writer) (err error) {
	// XXX sort obj by addr?
	for _, p := range c.Symbols {
		if !p.Done {
			continue
		}
		_, err = of.Write([]byte{p.Ins.Op})
		if err != nil {
			return
		}
		switch p.Ins.Bytes {
		case 1:
		case 2:
			_, err = of.Write([]byte{uint8(0xff & p.Operand)})
		case 3:
			buf := []byte{byte(p.Operand & 0xff), byte(p.Operand >> 8)}
			_, err = of.Write(buf)
		default:
			err = p.Stmt.NE("unsupported bytes length:%d", p.Ins.Bytes)
		}
		if err != nil {
			return
		}
	}
	return
}

// resovleLabel XXX should use toplogical sort first
func (c *Compiler) resovleEqualLabel() (resolved bool, err error) {
	resolved = true
	for _, p := range c.Symbols {
		var ne Expression
		for _, t := range p.Stmt.Expr {
			if t.Type != TLabel {
				ne = append(ne, t)
				continue
			}
			tls, ok := c.Labels[string(t.Value)]
			if !ok {
				err = p.Stmt.NE("can't find label :%s", string(t.Value))
				return
			}
			switch tls.Stmt.Mnemonic {
			case EQU, EPZ:
				resolved = false
				ne = append(ne, tls.Stmt.Expr...)
			default:
				ne = append(ne, t)
			}
		}
		p.Stmt.Expr = ne
	}
	return
}

func (c *Compiler) evalue(p *Prog) (err error) {
	ex := p.Stmt.Expr
	switch len(ex) {
	case 0:
		return
	case 1:
		p.Operand, err = ex[0].Uint16()
		if err != nil {
			return p.Stmt.NE(err.Error(), ex)
		}
		return
	case 2:
		err = p.Stmt.NE("invalid expression")
		return
	default:
		if (len(ex)-1)%2 != 0 {
			err = p.Stmt.NE("invalid expression")
			return
		}
	}
	l := len(ex) - 1
	p.Operand, err = ex[l].Uint16()
	if err != nil {
		return p.Stmt.NE(err.Error(), ex)
	}
	ex = ex[:l]

	// from right to left!!!
	for i := len(ex) - 1; i >= 0; i -= 2 {
		xt, operator := ex[i-1], ex[i]
		if operator.Type != TOperator {
			err = p.Stmt.NE("require operator got:%s", operator)
			return
		}
		var x uint16
		x, err = xt.Uint16()
		if err != nil {
			return
		}

		switch operator.Value[0] {
		case '+':
			p.Operand = x + p.Operand
		case '-':
			p.Operand = x - p.Operand
		case '*':
			p.Operand = x * p.Operand
		case '/':
			p.Operand = x / p.Operand
		case '&':
			p.Operand = x & p.Operand
		case '|':
			p.Operand = x | p.Operand
		case '^':
			p.Operand = x ^ p.Operand
		default:
			err = p.Stmt.NE("unsuppored operator:%s", operator)
			return
		}
	}
	return
}

func (c *Compiler) Evalue() (err error) {
	el := []error{}
	for _, p := range c.Symbols {
		if len(el) > 10 {
			break
		}
		err = c.evalue(p)
		if err != nil {
			el = append(el, err)
		}
	}
	return elToError(el)
}
