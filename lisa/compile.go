package lisa

import (
	"fmt"
	"go6502/ins"
	"io"
	"log"
)

type SType uint8

const (
	literal SType = iota + 1
	pseudo
	origin
)

type Symbol struct {
	Name    string
	Prev    *Symbol
	Next    *Symbol
	Stmt    *Stmt
	Operand uint16
	PC      uint16
}

type Compiler struct {
	Symbols     map[string]*Symbol
	target      []*Symbol
	first       *Symbol
	PC          uint16
	errCount    int
	warnAsError bool
}

func (c *Compiler) errorf(f string, args ...interface{}) {
	c.errCount++
	if c.errCount > 10 {
		log.Fatal("too many errors")
	}
	log.Printf(f, args)
}

func (c *Compiler) warnf(f string, args ...interface{}) {
	if c.warnAsError {
		c.errorf(f, args...)
		return
	}
	log.Printf(f, args)
}

func (c *Compiler) processPseudo(sym *Symbol) {

	s := sym.Stmt
	var err error
	if s.Mnemonic == ORG {
		if _, ok := c.Symbols["_ORG"]; ok {
			c.warnf("duplicate ORG found, replaced original")
		}
		if len(s.Expr) != 1 {
			c.errorf("ORG must be literal")
			return
		}

		if sym.Operand, err = s.Expr[0].Uint16(); err != nil {
			c.errorf("get ORG failed:", err)
			return
		}
		c.Symbols["_ORG"] = sym
		return
	}

	if s.Label == "" {
		c.errorf(s.NE("Pseudo require label").Error())
		return
	}

	c.Symbols[s.Label] = sym

	switch s.Mnemonic {
	case EPZ:
		sym.Stmt.Mode = ModeZeroPage
	case EQU:
		sym.Stmt.Mode = ModeAbsolute
	case STR, ASC, HEX, OBJ:
		c.errorf("unsupported mnemonic (yet)", s.Mnemonic)
	}
}

func isLocalLabel(s string) bool {
	if s == "" {
		return false
	}
	return s[0] == '^'
}

func (c *Compiler) buildSymbolTable(sl []*Stmt) {
	var prev *Symbol
	// clear all comment, build symbol table, guess prog first time
	for _, s := range sl {
		sym := &Symbol{
			Stmt: s,
		}
		if s.Mnemonic == 0 {
			continue
		}

		if s.Label != "" && !isLocalLabel(s.Label) {
			c.Symbols[s.Label] = sym
			sym.Name = s.Label
		}

		if isPseudo(s.Mnemonic) {
			c.processPseudo(sym)
			continue
		}

		if prev != nil {
			prev.Next = sym
			sym.Prev = prev
		}
		prev = sym

		if c.first == nil {
			c.first = sym
		}

		c.target = append(c.target, sym)
	}

	return
}

func Compile(sl []*Stmt, of io.Writer) (err error) {

	c := &Compiler{
		Symbols: make(map[string]*Symbol),
	}
	c.buildSymbolTable(sl)
	c.evalueLabel()
	c.encode()
	if c.errCount > 0 {
		err = fmt.Errorf("compile failed")
	}
	for k, s := range c.target {
		fmt.Println(k, s)
	}
	return
}

func (c *Compiler) encode() (err error) {

	for _, p := range c.target {
		i := ins.GetNameTable(p.Stmt.Mnemonic.String(), p.Stmt.Mode.String())
		if i.Mode == 0 {
			c.errorf("can't find instruction:%v ", p.Stmt)
			continue
		}
		if i.Bytes < 1 || i.Bytes > 3 {
			c.errorf(p.Stmt.NE("unsupported bytes length:%d", i.Bytes).Error())
			continue
		}

		p.PC = c.PC
		c.PC += uint16(i.Bytes)

	}
	return
}

func (c *Compiler) evalueLabel() {
}

func (c *Compiler) evalueExpr(p *Symbol) {
	var err error
	ex := p.Stmt.Expr
	switch len(ex) {
	case 0:
		return
	case 1:
		p.Operand, err = ex[0].Uint16()
		if err != nil {
			c.errorf(err.Error())
		}
		return
	case 2:
		c.errorf("invalid expression")
		return
	default:
		if (len(ex)-1)%2 != 0 {
			c.errorf("invalid expression")
			return
		}
	}

	l := len(ex) - 1
	p.Operand, err = ex[l].Uint16()
	if err != nil {
		c.errorf(err.Error())
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
