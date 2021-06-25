package lisa

import (
	"encoding/json"
	"go6502/ins"
	"log"
	"os"
)

const (
	textSeg = "_TEXT"
	dataSeg = "_DATA"
	bssSeg  = "_BSS"
)

func NewLink(fl []string, cfg string) (l *Link, err error) {
	l = &Link{
		SymList: make(map[string]*Symbol),
	}
	err = json.Unmarshal([]byte(cfg), l)
	l.fileList = fl
	return
}

type Link struct {
	ErrCount int                `json:"-"`
	out      string             `json:"-"`
	fileList []string           `json:"-"`
	SymList  map[string]*Symbol `json:"-"`
	Block    []*Block           `json:"-"`

	Text uint16 `json:"text"`
	BSS  uint16 `json:"bss"`
	Data uint16 `json:"data"`
}

func (l *Link) Errorf(f string, args ...interface{}) {
	if l.ErrCount > 10 {
		log.Fatal("too many error")
	}
	l.ErrCount++
	log.Printf(f, args)
}

func (l *Link) Load() {
	for _, f := range l.fileList {
		fd, err := os.Open(f)
		if err != nil {
			l.Errorf("load:%s", err)
			continue
		}
		dec := json.NewDecoder(fd)
		of := &ObjFile{}
		err = dec.Decode(of)
		fd.Close()
		if err != nil {
			l.Errorf("parse:%s", err)
			continue
		}
		for _, blk := range of.Block {
			l.Block = append(l.Block, blk)
			for _, sym := range blk.List {

				lb := sym.Stmt.Label
				if lb == "" || isLocalLabel(lb) {
					continue
				}
				ns, ok := l.SymList[lb]
				if ok {
					if ns.Stmt.Expr.String() == sym.Stmt.Expr.String() {
						continue
					}
					l.Errorf("conflict label:%s %s ||%s", lb, sym.Stmt, ns.Stmt)
				}
				l.SymList[lb] = sym
			}
		}
	}
	return
}

func (l *Link) evalueTerm(blk *Block, sym *Symbol, t *Term) (o int32, err error) {
	switch t.Type {
	case TBinary, TDecimal, THex:
		var ui uint16
		ui, err = t.Uint16()
		o = int32(ui)
	case TLabel:
		n := string(t.Value)
		s, ok := l.SymList[n]
		if !ok {
			l.Errorf("can't find symbol:%s", n)
			return
		}
		if s.Type == Constants {
			o, err = l.evalueExpr(blk, s)
		} else {
			o = int32(s.Addr)
		}
	case TCurrentLine:
		o = int32(sym.Addr)
	case TLSLabel:
		sum := int8(0)
		var i int
		for i := range blk.List {
			if blk.List[i] == sym {
				break
			}
		}
		for ; i >= 0; i-- {
			label := blk.List[i].Stmt.Label
			sum -= int8(blk.List[i].Size)
			if isLocalLabel(label) && label[1:] == string(t.Value) {
				o = int32(sum)
				return
			}
		}
		l.Errorf("can't find LSLabel for:%s", string(t.Value))
	case TGTLabel:
		sum := int8(0)
		var i int
		for i := range blk.List {
			if blk.List[i] == sym {
				break
			}
		}
		for ; i < len(blk.List); i++ {
			label := blk.List[i].Stmt.Label
			sum += int8(blk.List[i].Size)
			if isLocalLabel(label) && label[1:] == string(t.Value) {
				o = int32(sum)
				return
			}
		}
		l.Errorf("can't find GTLabel for:%s", string(t.Value))
	case TRaw:
		o = int32(sym.Addr)
	default:
		l.Errorf("unsupported type:%s", t.Type)
	}
	return
}

func (l *Link) evalueExpr(blk *Block, sym *Symbol) (o int32, err error) {

	ex := sym.Stmt.Expr
	if len(ex) == 0 {
		return
	}

	el := len(ex) - 1
	o, err = l.evalueTerm(blk, sym, ex[el])
	if err != nil {
		return
	}
	ex = ex[:el]

	// NOTE from right to left!!!
	for i := len(ex) - 1; i >= 0; i -= 2 {
		xt, operator := ex[i-1], ex[i]
		if operator.Type != TOperator {
			err = sym.Stmt.NE("require operator got:%s", operator)
			return
		}

		var x int32
		x, err = l.evalueTerm(blk, sym, xt)
		if err != nil {
			return
		}

		switch operator.Value[0] {
		case '+':
			o = x + o
		case '-':
			o = x - o
		case '*':
			o = x * o
		case '/':
			o = x / o
		case '&':
			o = int32(uint16(x) & uint16(o))
		case '|':
			o = int32(uint16(x) | uint16(o))
		case '^':
			o = int32(uint16(x) ^ uint16(o))
		default:
			err = sym.Stmt.NE("unsuppored operator:%s", operator)
			return
		}
	}
	return
}

func (l *Link) Resolve() {
	for _, blk := range l.Block {
		es := ""
		if blk.Origin == nil {
			es = textSeg // default text
			l.set(es, blk)
			continue
		}

		expr := blk.Origin.Stmt.Expr
		if len(expr) == 1 {
			es := string(expr[0].Value)
			switch es {
			case textSeg, dataSeg, bssSeg:
				l.set(es, blk)
				continue
			}
		}
		o, err := l.evalueExpr(blk, blk.Origin)
		if err != nil || o < 0 {
			l.Errorf("origin resolve Failed:%s", err)
		}
		blk.Origin.Addr = uint16(o)
		l.set("", blk)
	}

	for _, blk := range l.Block {
		for _, sym := range blk.List {
			if IsNonAddress(sym.Stmt.Mnemonic) {
				continue
			}
			var err error
			sym.Operand, err = l.evalueExpr(blk, sym)
			if err != nil {
				l.Errorf("evalue :%s", err)
			}
			if sym.Stmt.Mode == ins.Relative {
				sym.Operand = sym.Operand - int32(sym.Addr) - 2
			}
		}
	}
}

func (l *Link) set(seg string, blk *Block) {
	if blk.Origin == nil {
		blk.Origin = &Symbol{}
	}
	switch seg {
	case textSeg:
		blk.Origin.Addr = l.Text
		l.Text += blk.Size
	case dataSeg:
		blk.Origin.Addr = l.Data
		l.Data += blk.Size
	case bssSeg:
		blk.Origin.Addr = l.BSS
		l.BSS += blk.Size
	default:
	}
	org := uint16(blk.Origin.Addr)

	// TODO check overlay

	for _, sym := range blk.List {
		if IsNonAddress(sym.Stmt.Mnemonic) {
			continue
		}
		sym.Addr = org
		org += uint16(sym.Size)
	}
}
