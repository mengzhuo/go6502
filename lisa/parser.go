// Asm parser for 6502/Apple ][
package lisa

import (
	"bufio"
	"bytes"
	"fmt"
	"go6502/ins"
	"io"
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
	mnemonicMap["="] = EQU
}

type Stmt struct {
	Line     int
	Label    string
	Mnemonic Mnemonic
	Oper     string
	Comment  string
	Order    byte
	Expr     Expression
	Mode     ins.Mode
}

func (s *Stmt) NE(f string, args ...interface{}) error {
	args = append(args, s.Line)
	if len(args) == 1 {
		return fmt.Errorf("Line:%d "+f, s.Line)
	}
	l := len(args) - 1
	args[0], args[l] = args[l], args[0]
	return fmt.Errorf("Line:%d "+f, args...)
}

func checkLabels(il []*Stmt) (err error) {
	// build label table
	lt := map[string]*Stmt{}
	el := []error{}
	for _, s := range il {
		if len(el) > 10 {
			break
		}
		if s.Label != "" {
			if ps, ok := lt[s.Label]; ok {
				el = append(el, fmt.Errorf("Line %d duplicate regular label %q, previous defined at line %d",
					s.Line, s.Label, ps.Line))
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
		if len(s.Expr) == 0 {
			continue
		}
	next:
		for t := s.Expr[0]; t != nil; t = t.next {
			if len(el) > 10 {
				el = append(el, fmt.Errorf("too many errors"))
				break end
			}
			switch t.Type {
			case TLabel:
				if _, ok := lt[string(t.Value)]; !ok {
					el = append(el, s.NE("can't find regular label:%s", string(t.Value)))
				}

			case TLSLabel:
				tl := il[:i]
				for j := len(tl) - 1; j >= 0; j-- {
					jt := tl[j]
					if strings.HasPrefix(jt.Label, "^") && jt.Label[1:] == string(t.Value) {
						break next
					}
				}
				el = append(el, s.NE("can't find local label:%s", string(t.Value)))

			case TGTLabel:
				for _, jt := range il[i:] {
					if strings.HasPrefix(jt.Label, "^") && jt.Label[1:] == string(t.Value) {
						break next
					}
				}
				el = append(el, s.NE("can't find local label:%s", string(t.Value)))
			}
		}
	}

	err = elToError(el)
	return
}

func semantic(il []*Stmt) (err error) {
	type checker func(il []*Stmt) error
	for _, f := range []checker{
		checkLabels,
	} {
		err = f(il)
		if err != nil {
			return
		}
	}
	return
}

func (l *Stmt) String() string {
	return fmt.Sprintf("[%4d] L:%-8s %s O[%s]%s M:%s %s", l.Line, l.Label, l.Mnemonic,
		string(l.Order), l.Expr, l.Mode, l.Comment)
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

		switch s.Mnemonic {
		case ASC, STR, BYT, DA:
			s.Expr = Expression{&Term{Type: TRaw, Value: []byte(s.Oper)}}
			continue
		}

		s.Mode, s.Expr, s.Order, err = parseOperand([]byte(s.Oper))
		if err != nil {
			el = append(el, fmt.Errorf("Line:%d %s", s.Line, err))
			if len(el) > 10 {
				el = append(el, fmt.Errorf("too many error"))
				break
			}
		}
		if len(s.Expr) == 0 {
			continue
		}
		for _, t := range s.Expr {
			t.Stmt = s
		}
		err = syntaxCheck(s.Expr[0])
		if err != nil {
			el = append(el, fmt.Errorf("Line:%d %s", s.Line, err))
			if len(el) > 10 {
				break
			}
		}
	}
	return elToError(el)
}

func Parse(rd io.Reader) (il []*Stmt, err error) {
	sc := bufio.NewScanner(rd)
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
		case ';':
			l.Comment = string(t)
			t = t[:0]
		case '*':
			if l.Mnemonic == 0 {
				l.Comment = string(t)
				t = t[:0]
				continue
			}
			fallthrough
		default:
			i := bytes.IndexAny(t, "\t \n\r")
			if i == -1 {
				i = len(t)
			}

			var w []byte
			// XXX How to do better ?
			if l.Mnemonic == ASC || l.Mnemonic == STR {
				w, t, err = lookForQuote(t)
				if err != nil {
					return l.NE(err.Error())
				}
			} else {
				w, t = t[:i], t[i:]
			}

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
	if len(b) > 3 {
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

func lookForQuote(t []byte) (w, r []byte, err error) {
	if len(t) < 2 {
		err = fmt.Errorf("quotation must more than 2")
		return
	}
	start := t[0]
	switch start {
	case '\'', '"':
	default:
		err = fmt.Errorf("looking for quota, got: %s ", string(t))
		return
	}
	l := bytes.LastIndexByte(t[1:], start)
	if l == -1 {
		err = fmt.Errorf("can't find matched start: %s", string(start))
		return
	}
	w, r = t[1:l+1], t[l+2:]
	return
}
