// Asm parser for 6502/Apple ][
package a65

import (
	"bufio"
	"bytes"
	"fmt"
	"go6502/ins"
	"os"
	"strings"
)

var letters []rune
var lan []rune

func init() {
	for i := '0'; i <= '9'; i++ {
		lan = append(lan, i)
	}
	for i := 'a'; i <= 'z'; i++ {
		letters = append(letters, i)
		lan = append(lan, i)
	}
	for i := 'A'; i <= 'Z'; i++ {
		lan = append(lan, i)
		letters = append(letters, i)
	}
}

type Stmt struct {
	No       int
	Label    string
	Mnemonic string
	Oper     string
	Comment  string
	ins      ins.Ins
	operand  uint16
}

func (l *Stmt) String() string {
	return fmt.Sprintf("[%4d] L:%-8s I:%s O:%s %s;%s", l.No, l.Label, l.Mnemonic, l.Oper, l.Comment, l.ins)
}

func Parse(r *os.File) (il []*Stmt, err error) {
	sc := bufio.NewScanner(r)
	for ln := 1; sc.Scan(); ln++ {
		t := sc.Bytes()
		if len(t) == 0 {
			// skip empty line
			continue
		}
		st := &Stmt{No: ln}
		if err = st.parse(t); err != nil {
			return
		}
		if err = st.checkSyntax(); err != nil {
			return
		}
		il = append(il, st)
	}
	err = sc.Err()
	return
}

func (l *Stmt) findIns() (err error) {
	if l.Mnemonic == "" {
		return
	}

	m := ins.Implied
	if l.Oper == "" {
		l.ins = ins.GetNameTable(l.Mnemonic, m)
		return
	}
	switch l.Oper[0] {
	case '(':

	case '#':
	}

	l.ins = ins.GetNameTable(l.Mnemonic, m)
	return
}

func (l *Stmt) checkSyntax() (err error) {
	if l.Label == "" && l.Mnemonic == "" && l.Comment == "" {
		return fmt.Errorf("statment invalid:%s", l)
	}

	if len(l.Label) > 0 {
		if len(l.Label) > 8 {
			return fmt.Errorf("Label too long >8:%s", l.Label)
		}
		i := strings.IndexRune(string(letters), rune(l.Label[0]))
		if i == -1 {
			return fmt.Errorf("Label not start with char:%s", l.Label)
		}
		for _, c := range l.Label[1:] {
			i := strings.IndexRune(string(lan), c)
			if i == -1 {
				return fmt.Errorf("Label contains invalid char:%s", l.Label)
			}
		}
	}
	if len(l.Mnemonic) > 0 {
		if len(l.Mnemonic) != 3 {
			return fmt.Errorf("Mnemonic invalid:%s", l.Mnemonic)
		}
		for _, c := range l.Mnemonic {
			if strings.IndexRune(string(letters), c) == -1 {
				return fmt.Errorf("Mnemonic invalid", l.Mnemonic)
			}
		}
	}
	return
}

func (l *Stmt) parse(t []byte) (err error) {
	defer func() {
		if l.Label == "\t" {
			l.Label = ""
		}
	}()
	for len(t) > 0 {
		c := t[0]
		switch c {
		case ' ', '\t':
			if l.Label == "" {
				l.Label = "\t"
			}
			t = t[1:]
		case ';', '*':
			l.Comment = string(t)
			return
		default:
			i := bytes.IndexAny(t, "\t \n\r")
			if i == -1 {
				i = len(t)
			}
			var w []byte
			w, t = t[:i], t[i:]
			if l.Label == "" {
				l.Label = string(w)
				continue
			}
			if l.Mnemonic == "" {
				l.Mnemonic = string(w)
				continue
			}
			if l.Oper == "" {
				l.Oper = string(w)
				continue
			}
		}
	}
	return
}
