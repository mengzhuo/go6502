package lisa

import (
	"fmt"
	"go6502/ins"
	"io"
	"log"
	"strings"
)

type Addr struct {
	Mode ins.Mode
	Pos  int32
}

type Prog struct {
	Stmt    *Stmt // origin statement
	Next    *Prog
	Addr    Addr
	Op      byte
	Operand uint16
	Done    bool
}

type Compiler struct {
	Labels map[string]*Prog //symbol table
	Const  map[string]uint16
	first  *Prog
	Origin uint16
}

func (c *Compiler) init(sl []*Stmt) (err error) {
	// clear all comment, build symbol table, guess prog first time
	var prev *Prog
	for _, s := range sl {
		if s.Mnemonic == 0 {
			continue
		}

		p := &Prog{
			Stmt: s,
		}
		if prev != nil {
			prev.Next = p
		}
		prev = p

		if c.first == nil {
			c.first = p
		}

		if s.Label != "" && !strings.HasPrefix(s.Label, "^") {
			c.Labels[s.Label] = p
		}
		if s.Mode != 0 {
			// if instruction select a mode force prog to use it
			p.Addr.Mode = s.Mode
		}

		i := ins.GetNameTable(p.Stmt.Mnemonic.String(), ins.Implied)
		if i.Bytes == 1 {
			if s.Oper != "" {
				err = fmt.Errorf("implied instruction has operand")
				return
			}
			p.Op = i.Op
			p.Addr.Mode = ins.Implied
			p.Done = true
		}
	}
	for p := c.first; p != nil; p = p.Next {
		log.Println("init:", p)
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

	return
}
