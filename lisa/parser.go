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
	// build label table
	lt := map[string]*Stmt{}
	el := []error{}
	for _, s := range il {
		if len(el) > 10 {
			break
		}
		if s.Label != "" {
			if _, ok := lt[s.Label]; ok {
				el = append(el, fmt.Errorf("duplicate regular label %s", s.Label))
				continue
			}
			lt[s.Label] = s
		}
	}

	err = elToError(el)
	if err != nil {
		return
	}
end:
	// find Label
	for i, s := range il {
		for t := s.Expr; t != nil; t = t.next {
			if len(el) > 10 {
				el = append(el, fmt.Errorf("too many errors"))
				break end
			}
			switch t.Type {
			case TLabel:
				ls, ok := lt[string(t.Value)]
				if !ok {
					el = append(el, fmt.Errorf("can't find regular label:%s", string(t.Value)))
					continue
				}
				t.label = ls

			case TLSLabel:
				tl := il[:i]
				for j := len(tl) - 1; j >= 0; j-- {
					jt := tl[j]
					if jt.Label == string(t.Value) {
						t.label = jt
						break
					}
				}
				if t.label == nil {
					el = append(el, fmt.Errorf("can't find local label:%s", string(t.Value)))
				}

			case TGTLabel:
				for _, jt := range il[i:] {
					if jt.Label == string(t.Value) {
						t.label = jt
						break
					}
				}
				if t.label == nil {
					el = append(el, fmt.Errorf("can't find local label:%s", string(t.Value)))
				}
			}
		}
	}

	err = elToError(el)
	return
}

func (l *Stmt) String() string {
	return fmt.Sprintf("[%4d] L:%-8s I:%s O:%s M:%s %s", l.Line, l.Label, l.Mnemonic, l.Oper, l.Mode, l.Comment)
}

func elToError(el []error) (err error) {
	if len(el) == 0 {
		return nil
	}
	buf := []string{}
	for _, e := range el {
		buf = append(buf, e.Error())
	}
	return fmt.Errorf(strings.Join(buf, "\n"))
}

func parse(il []*Stmt) (err error) {

	el := []error{}

	for _, s := range il {
		if s.Oper == "" {
			continue
		}

		s.Mode, s.Expr, err = parseOperand([]byte(s.Oper))
		if err != nil {
			el = append(el, fmt.Errorf("Line:%d %s", s.Line, err))
			if len(el) > 10 {
				el = append(el, fmt.Errorf("too many error"))
				break
			}
		}
		err = syntaxCheck(s.Expr)
		if err != nil {
			el = append(el, fmt.Errorf("Line:%d %s", s.Line, err))
			if len(el) > 10 {
				break
			}
		}
	}
	return elToError(el)
}

func Parse(r *os.File) (il []*Stmt, err error) {
	sc := bufio.NewScanner(r)
	for ln := 1; sc.Scan(); ln++ {
		t := sc.Bytes()
		if len(t) == 0 || len(bytes.TrimSpace(t)) == 0 {
			// skip empty line
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
	if err != nil {
		return
	}
	err = semantic(il)
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
	// local label
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
