// Asm parser for 6502/Apple ][
package a65

import (
	"bufio"
	"fmt"
	"go6502/ins"
	"os"
	"strings"
)

type Line struct {
	No      int
	Label   string
	Ins     string
	Oper    string
	Comment string
	ins     ins.Ins
}

func (l *Line) String() string {
	return fmt.Sprintf("[%4d] L:%s I:%s O:%s %s;%s", l.No, l.Label, l.Ins, l.Oper, l.Comment, l.ins)
}

func Parse(r *os.File) (il []*Line, err error) {
	sc := bufio.NewScanner(r)
	for ln := 1; sc.Scan(); ln++ {
		t := sc.Bytes()
		if len(t) == 0 {
			// skip empty line
			continue
		}
		var l *Line
		l, err = parseline(t, ln)
		if err != nil {
			return
		}
		il = append(il, l)
	}
	err = sc.Err()
	if err != nil {
		return
	}

	for _, l := range il {
		err = l.findIns()
		if err != nil {
			return
		}
	}
	return
}

func (l *Line) findIns() (err error) {
	if l.Ins == "" {
		return
	}
	m := ins.Implied
	if l.Oper != "" {
	}

	l.ins = ins.GetNameTable(l.Ins, m)
	return
}

func parseline(t []byte, ln int) (l *Line, err error) {
	label := true
	l = &Line{No: ln}
	switch t[0] {
	case ';', '*':
		l.Comment = string(t)
		return
	case '\t':
		label = false
	}

	fs := strings.FieldsFunc(string(t), func(r rune) bool { return r == '\t' })
	last := len(fs) - 1
	if fs[last][0] == ';' {
		l.Comment = fs[last]
		fs = fs[:last]
	}
	if len(fs) == 0 {
		return
	}

	if label {
		l.Label = fs[0]
		fs = fs[1:]
	}
	if len(fs) != 1 {
		return
	}
	op := strings.Fields(fs[0])
	switch len(op) {
	default:
		err = fmt.Errorf("Syntax error:%s", string(t))
	case 1:
		l.Ins = strings.ToUpper(op[0])
	case 2:
		l.Ins = strings.ToUpper(op[0])
		l.Oper = op[1]
	}
	return
}
