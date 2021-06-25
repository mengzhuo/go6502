package lisa

import (
	"strings"
	"testing"
)

var good = [...]string{
	`8000`,
	`$5-$3+$2`,
	`#LABEL`,
	`#$FF`,
	`#!6253`,
	`#%1011001`,
	`#'A'`,
	`#"Z"+$1`,
	`LBL+$3`,
	`HERE-THERE`,
	`*+!10`,
	`"Z"+$1`,
	`$FF`,
	`!10`,
	`!-936`,
	`LABEL/2*X^$FFFF&$10FF|1`,
	`LBL-$FF+!10-%1010011`,
	`/LABEL`,
	`/$FF`,
	`/!6253`,
	`/%101001100`,
	`/LBL+$4`,
	`/$F88F`,
}

var bad = map[string]string{
	`/LBL+`:   "expect valid term",
	`/LBL+$`:  "empty hex",
	`/LBL+$G`: "invalid hex",
}

func TestTokenizeExpr(t *testing.T) {
	for i, g := range good {
		c := []string{}
		_, e, order, err := parseOperand([]byte(g))
		c = append(c, "ORDER:"+string(order))
		for term := e; term != nil; term = term.next {
			c = append(c, term.String())
		}
		t.Log(i, g, strings.Join(c, " <-> "))
		if err != nil {
			t.Errorf("%s:%v", g, err)
		}
	}
}

func TestSyntaxExpr(t *testing.T) {
	for k, v := range bad {
		_, e, _, err := parseOperand([]byte(k))
		if err != nil || e == nil {
			t.Errorf("%s:%v", k, err)
		}
		err = syntaxCheck(e)
		if err == nil {
			t.Errorf("expecting %s err=%s got nil", k, v)
			continue
		}
		if !strings.Contains(err.Error(), v) {
			t.Errorf("expecting %s err=%s, got=%v", k, v, err)
		}
	}
}

func TestTermUint16(t *testing.T) {

	for k, v := range map[string]uint16{
		"$16":   0x16,
		"!16":   16,
		"%0101": 5,
		"$42":   0x42,
		"$FF":   0xff,
		"$FFFF": 0xFFff,
	} {
		_, e, _, err := parseOperand([]byte(k))
		if err != nil {
			t.Error(err)
		}
		u, err := e.Uint16()
		if err != nil {
			t.Error(err)
		}
		if u != v {
			t.Errorf("%s => 0x%x, expect 0x%x", k, u, v)
		}
	}
}
