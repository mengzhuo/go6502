// Asm parser for 6502/Apple ][
package lisa

import (
	"bufio"
	"bytes"
	"fmt"
	"go6502/ins"
	"os"
	"strings"
)

var letters []byte
var lan []byte
var hex []byte
var mnemonicMap map[string]Mnemonic

func init() {
	for i := byte('0'); i <= '9'; i++ {
		lan = append(lan, i)
		hex = append(hex, i)
	}
	for i := byte('a'); i <= 'z'; i++ {
		if i <= 'f' {
			hex = append(hex, i)
		}
		letters = append(letters, i)
		lan = append(lan, i)
	}
	for i := byte('A'); i <= 'Z'; i++ {
		if i <= 'F' {
			hex = append(hex, i)
		}
		lan = append(lan, i)
		letters = append(letters, i)
	}
	mnemonicMap = map[string]Mnemonic{}
	for i := ADC; i <= USR; i++ {
		mnemonicMap[i.String()] = i
	}
}

type Stmt struct {
	Line     int
	Label    string
	Mnemonic Mnemonic
	Oper     string
	Comment  string
	Expr     *Term
	Mode     ins.Mode
}

func semantic(il []*Stmt) (err error) {
	// Label Table
	lt := map[string]*Stmt{}
	for _, s := range il {
		if s.Label != "" {
			if _, ok := lt[s.Label]; ok {
				err = fmt.Errorf("duplicate label %s", s.Label)
				return
			}
			lt[s.Label] = s
		}
	}
	return
}

func (l *Stmt) String() string {
	return fmt.Sprintf("[%4d] L:%-8s I:%s O:%s %s", l.Line, l.Label, l.Mnemonic, l.Oper, l.Comment)
}

func parse(il []*Stmt) (err error) {

	for _, s := range il {
		if s.Oper == "" {
			continue
		}
		s.Expr, err = getTermFromExpr([]byte(s.Oper))
		if err != nil {
			return
		}
		err = syntaxCheck(s.Expr)
		if err != nil {
			return
		}
	}
	return
}

func Parse(r *os.File) (il []*Stmt, err error) {
	sc := bufio.NewScanner(r)
	for ln := 1; sc.Scan(); ln++ {
		t := sc.Bytes()
		if len(t) == 0 {
			// skip empty line
			continue
		}

		if len(bytes.TrimSpace(t)) == 0 {
			continue
		}

		st := &Stmt{}
		if err = lexing(st, t); err != nil {
			return
		}
		st.Line = ln
		il = append(il, st)
	}
	err = parse(il)
	return
}

func lexing(l *Stmt, t []byte) (err error) {

	for len(t) > 0 && err == nil {
		c := t[0]
		switch c {
		case ' ', '\t':
			if l.Label == "" {
				l.Label = "\t"
			}
			t = t[1:]
		case ';', '*':
			l.Comment = string(t)
			t = t[:0]
			continue
		default:
			i := bytes.IndexAny(t, "\t \n\r")
			if i == -1 {
				i = len(t)
			}
			var w []byte
			w, t = t[:i], t[i:]
			if l.Label == "" {
				err = l.processLabel(w)
				continue
			}
			if l.Mnemonic == 0 {
				err = l.handleMnemonic(w)
				continue
			}
			if l.Oper == "" {
				l.Oper = string(w)
			}
		}
	}

	if l.Label == "\t" {
		l.Label = ""
	}

	return
}

func (l *Stmt) processLabel(b []byte) (err error) {
	l.Label = strings.ToUpper(string(b))
	return
}

func (l *Stmt) handleMnemonic(b []byte) (err error) {
	if len(b) < 2 || len(b) > 3 {
		return fmt.Errorf("invalid mnemonic %s", string(b))
	}
	b = bytes.ToUpper(b)
	// trim
	if b[0] == '.' {
		b = b[1:]
	}
	m, ok := mnemonicMap[string(b)]
	if !ok {
		return fmt.Errorf("invalid mnemonic %s", string(b))
	}
	l.Mnemonic = m
	return
}
