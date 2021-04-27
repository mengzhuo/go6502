package lisa

import (
	"fmt"
	"go6502/ins"
	"io"
	"log"
)

type Symbol struct {
	Name    string
	Prev    *Symbol
	Next    *Symbol
	Stmt    *Stmt
	Operand uint16
	Address uint16
	Done    bool
}

func (s *Symbol) String() string {
	return fmt.Sprintf("%s  O:0x%04X A:0x%04X", s.Stmt, s.Operand, s.Address)
}

type Compiler struct {
	Symbols     map[string]*Symbol
	target      []*Symbol
	PC          uint16
	errCount    int
	warnAsError bool
}

func (c *Compiler) errorf(f string, args ...interface{}) {
	c.errCount++
	if c.errCount > 10 {
		log.Fatal("too many errors")
	}
	log.Printf(f, args...)
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
	if s.Mnemonic == ORG {
		return
	}
	if s.Label == "" {
		c.errorf(s.NE("Pseudo require label").Error())
		return
	}

	c.Symbols[s.Label] = sym

	switch s.Mnemonic {
	case STR, HEX, OBJ:
		c.errorf("unsupported mnemonic (yet)", s.Mnemonic)
	}
}

func isLocalLabel(s string) bool {
	if s == "" {
		return false
	}
	return s[0] == '^'
}

func (c *Compiler) expandConst() (change bool) {
	for _, sym := range c.target {
		ne := Expression{}
		for _, t := range sym.Stmt.Expr {
			if t.Type != TLabel {
				ne = append(ne, t)
				continue
			}
			ns, ok := c.Symbols[string(t.Value)]
			if !ok {
				c.errorf("can't find label:%s", string(t.Value))
				continue
			}
			switch ns.Stmt.Mnemonic {
			case EPZ:
				switch sym.Stmt.Mode {
				case ins.Absolute:
					sym.Stmt.Mode = ins.ZeroPage
				case ins.AbsoluteX:
					sym.Stmt.Mode = ins.ZeroPageX
				case ins.AbsoluteY:
					sym.Stmt.Mode = ins.ZeroPageY
				}
				fallthrough
			case EQU:
				change = true
				ne = append(ne, ns.Stmt.Expr...)
			default:
				ne = append(ne, t)
			}
		}
		sym.Stmt.Expr = ne
	}
	return
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

		if isNonAddress(s.Mnemonic) {
			c.processPseudo(sym)
			continue
		}

		if prev != nil {
			prev.Next = sym
			sym.Prev = prev
		}
		prev = sym

		if isRelative(s.Mnemonic) {
			sym.Stmt.Mode = ins.Relative
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
	for c.expandConst() {
	}
	c.determineAddress()
	c.evalue()
	c.encode(of)
	for k, s := range c.target {
		fmt.Println(k, s)
	}
	if c.errCount > 0 {
		return fmt.Errorf("compile failed")
	}
	return
}

func (c *Compiler) writeRawData(of io.Writer, sym *Symbol) {
	switch sym.Stmt.Mnemonic {
	case ASC:
		d := sym.Stmt.Expr[0].Value
		n, err := of.Write(d)
		if err != nil || n != len(d) {
			c.errorf("%s write error:%d %s", sym.Stmt, n, err)
		}	
	default:
		c.errorf("%d unsupported raw data %s", sym.Stmt.Line, sym.Stmt)
	}
}

func (c *Compiler) encode(of io.Writer) {
	for _, sym := range c.target {
		stm := sym.Stmt
		if isRawData(stm.Mnemonic) {
			c.writeRawData(of, sym)
			continue
		}	
		i := ins.GetNameTable(stm.Mnemonic.String(), stm.Mode.String())
		if i.Mode == 0 {
			c.errorf("can't find instruction:%v ", stm)
			continue
		}
		if i.Bytes < 1 || i.Bytes > 3 {
			c.errorf(stm.NE("unsupported bytes length:%d", i.Bytes).Error())
			continue
		}
		buf := make([]byte, i.Bytes)
		buf[0] = i.Op
		if i.Bytes == 1 {
			continue
		}

		switch i.Mode {
		case ins.Relative:
			// op is absolute address
			ab := int32(sym.Operand)
			ab -= int32(sym.Address)
			if ab < -126 || ab > 128 {
				c.errorf("overflow relative address:%d %d %d", ab, sym.Operand, sym.Address)
				continue
			}
			buf[1] = byte(int8(ab))
		default:
			if i.Bytes == 2 {
				buf[1] = byte(sym.Operand)
			}
			if i.Bytes == 3 {
				buf[1] = byte(0xff & sym.Operand)
				buf[2] = byte(sym.Operand >> 8)
			}
		}
		n, err := of.Write(buf)
		if err != nil || n != len(buf) {
			c.errorf(stm.NE("write data failed:%s %d", err, n).Error())
			continue
		}
	}
}

func (c *Compiler) determineAddress() (err error) {

	for _, p := range c.target {
		if p.Stmt.Mnemonic == ORG {
			if _, ok := c.Symbols["_ORG"]; ok {
				c.warnf("duplicate ORG found, replaced original")
			}
			if len(p.Stmt.Expr) != 1 {
				c.errorf("ORG must be literal")
				return
			}

			if c.PC, err = c.evalueExpr(p); err != nil {
				c.errorf("get ORG failed:", err)
				return
			}
			c.Symbols["_ORG"] = p
			continue
		}
	}

	for _, p := range c.target {

		if isRawData(p.Stmt.Mnemonic) {
			p.Address = c.PC
			c.PC += uint16(len(p.Stmt.Expr[0].Value))
			continue
		}

		i := ins.GetNameTable(p.Stmt.Mnemonic.String(), p.Stmt.Mode.String())
		if i.Mode == 0 {
			c.errorf("can't find instruction:%v ", p.Stmt)
			continue
		}
		if i.Bytes < 1 || i.Bytes > 3 {
			c.errorf(p.Stmt.NE("unsupported bytes length:%d", i.Bytes).Error())
			continue
		}
		p.Address = c.PC
		c.PC += uint16(i.Bytes)
	}
	return
}

func (c *Compiler) evalue() {
	var err error
	// look up for absolute term
	for _, s := range c.target {
		if isRawData(s.Stmt.Mnemonic) {
			continue
		}
		_, err = c.evalueExpr(s)
		if err != nil {
			c.errorf("evalue expression failed:%s", err)
			continue
		}
	}
}

func (c *Compiler) evalueTerm(sym *Symbol, t *Term) (o uint16, err error) {
	switch t.Type {
	case TBinary, TDecimal, THex:
		return t.Uint16()
	case TLabel:
		n := string(t.Value)
		s, ok := c.Symbols[n]
		if !ok {
			err = fmt.Errorf("can't find symbol:%s", n)
			return
		}
		o = s.Address
		return
	case TCurrentLine:
		o = sym.Address
	case TLSLabel:
		for ns := sym.Prev; ns != nil; ns = ns.Prev {
			if ns.Stmt.Label == "" {
				continue
			}
			if ns.Stmt.Label[1:] == string(t.Value) {
				o = ns.Address
				return
			}
		}
		err = fmt.Errorf("can't find LSLabel for:%s", string(t.Value))
	case TGTLabel:
		for ns := sym.Next; ns != nil; ns = ns.Next {
			if ns.Stmt.Label == "" {
				continue
			}
			if ns.Stmt.Label[1:] == string(t.Value) {
				o = ns.Address
				return
			}
		}
		err = fmt.Errorf("can't find GTLabel for:%s", string(t.Value))
	default:
		err = fmt.Errorf("unsupported type:%s", t.Type)
	}
	return
}

func (c *Compiler) evalueExpr(p *Symbol) (o uint16, err error) {

	if p.Done {
		o = p.Operand
		return
	}

	ex := p.Stmt.Expr
	if len(ex) == 0 {
		return
	}

	l := len(ex) - 1
	p.Operand, err = c.evalueTerm(p, ex[l])
	if err != nil {
		return
	}
	ex = ex[:l]
	// NOTE from right to left!!!
	for i := len(ex) - 1; i >= 0; i -= 2 {
		xt, operator := ex[i-1], ex[i]
		if operator.Type != TOperator {
			err = p.Stmt.NE("require operator got:%s", operator)
			return
		}

		var x uint16
		x, err = c.evalueTerm(p, xt)
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
	o = p.Operand
	p.Done = true
	return
}
