package lisa

import (
	"encoding/binary"
	ehex "encoding/hex"
	"fmt"
	"go6502/ins"
	"io"
	"log"
)

//go:generate stringer --type=SType
type SType uint8

const (
	Default SType = 1 << iota
	Constants
	RawData
	Jump
	Branch
)

type Symbol struct {
	Stmt    *Stmt
	Type    SType
	Addr    uint16
	Size    uint8
	Operand int32
}

func (s *Symbol) String() string {
	if isRelative(s.Stmt.Mnemonic) {
		return fmt.Sprintf("%s T:%s A:%04X O:%d", s.Stmt, s.Type.String(), s.Addr, s.Operand)
	} else {
		return fmt.Sprintf("%s T:%s A:%04X O:0x%04X", s.Stmt, s.Type.String(), s.Addr, uint(s.Operand))
	}
}

func (s *Symbol) Data() (d []byte, err error) {
	op := s.Stmt.Mnemonic
	d = make([]byte, s.Size)
	var sum int
	switch op {
	case HEX:
		for i := 0; i < len(s.Stmt.Expr); i += 2 {
			n, err := ehex.Decode(d[sum:], s.Stmt.Expr[i].Value)
			if err != nil {
				return nil, fmt.Errorf("invalid data%s[%d] error:%s", s, s.Size, err)
			}
			sum += n
		}

		if sum != int(s.Size) {
			err = fmt.Errorf("invalid data%s [%d!=%d]", s, s.Size, sum)
			return
		}
	case STR:
		d[0] = uint8(s.Size)
		sum += 1
		fallthrough
	case ASC:
		for i := 0; i < len(s.Stmt.Expr); i += 2 {
			sum += copy(d[sum:], s.Stmt.Expr[i].Value)
		}
		if sum != int(s.Size) {
			err = fmt.Errorf("invalid data[%d!=%d]%s", s.Size, sum, s)
		}

	case ADR:
		if s.Size != 2 || s.Operand > 0xffff || s.Operand < 0 {
			err = fmt.Errorf("invalid data%s[%d]", s, s.Size)
			return
		}
		binary.LittleEndian.PutUint16(d, uint16(s.Operand))
	default:
		ins := ins.GetNameTable(op.String(), s.Stmt.Mode.String())
		d[0] = ins.Op
		switch ins.Bytes {
		case 1:
		case 2:
			d[1] = uint8(s.Operand)
		case 3:
			binary.LittleEndian.PutUint16(d[1:], uint16(s.Operand))
		default:

			err = fmt.Errorf("invalid ins:%s", s.Stmt)
		}
	}
	return
}

type Compiler struct {
	stmt    []*Stmt
	defined map[string]*Stmt
	Symbols []*Symbol

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

func (c *Compiler) preprocess(s *Stmt) {
	if s.Mnemonic == 0 {
		return
	}

	if s.Label != "" && !isLocalLabel(s.Label) {
		if _, ok := c.defined[s.Label]; ok {
			c.errorf("duplicate %s ", s.Label)
			return
		}
		c.defined[s.Label] = s
	}

	if isRelative(s.Mnemonic) {
		s.Mode = ins.Relative
	}
}

func (c *Compiler) expandConst() (change bool) {
	for _, s := range c.stmt {
		ne := Expression{}
		for _, t := range s.Expr {
			if t.Type != TLabel {
				ne = append(ne, t)
				continue
			}
			ns, ok := c.defined[string(t.Value)]
			if !ok {
				// might be in link time
				ne = append(ne, t)
				continue
			}
			switch ns.Mnemonic {
			case EPZ:
				switch s.Mode {
				case ins.Absolute:
					s.Mode = ins.ZeroPage
				case ins.AbsoluteX:
					s.Mode = ins.ZeroPageX
				case ins.AbsoluteY:
					s.Mode = ins.ZeroPageY
				}
				fallthrough
			case EQU:
				change = true
				ne = append(ne, ns.Expr...)
			default:
				ne = append(ne, t)
			}
		}
		s.Expr = ne
	}
	return
}

func isLocalLabel(s string) bool {
	if s == "" {
		return false
	}
	return s[0] == '^'
}

func Compile(sl []*Stmt, of io.Writer, zobj, debug bool) (err error) {

	rl := []*Stmt{}
	for _, s := range sl {
		if s.Mnemonic != 0 {
			rl = append(rl, s)
		}
	}

	c := &Compiler{stmt: rl, defined: map[string]*Stmt{}}
	for _, s := range c.stmt {
		c.preprocess(s)
	}

	// Find any available value at compile time
	for c.expandConst() {
	}
	c.initSymbol()
	c.determineType()
	c.determineSize()

	// split into blocks
	if debug {
		for _, sym := range c.Symbols {
			fmt.Println(sym)
		}
	}

	if c.errCount > 0 {
		return fmt.Errorf("compile failed")
	}
	err = NewObjFile(c.Symbols).WriteTo(of)
	if err != nil {
		log.Fatalf("compile failed:%s", err)
	}
	return
}

func (c *Compiler) initSymbol() {
	for _, stmt := range c.stmt {
		c.Symbols = append(c.Symbols, &Symbol{
			Stmt: stmt,
		})
	}
}

func (c *Compiler) determineType() {
	for _, sym := range c.Symbols {
		switch sym.Stmt.Mnemonic {
		case EQU, EPZ:
			sym.Type = Constants
			continue
		case JSR, JMP:
			sym.Type = Jump
			continue
		}

		if isRawData(sym.Stmt.Mnemonic) {
			sym.Type = RawData
			continue
		}
		if sym.Stmt.Mode == ins.Relative {
			sym.Type = Branch
			continue
		}

		sym.Type = Default
	}
}

func (c *Compiler) determineSize() {

	for _, sym := range c.Symbols {
		if IsNonAddress(sym.Stmt.Mnemonic) {
			continue
		}

		if isRawData(sym.Stmt.Mnemonic) {
			for _, t := range sym.Stmt.Expr {
				if t.Type == TOperator {
					continue
				}
				ld := uint8(len(t.Value))
				switch sym.Stmt.Mnemonic {
				case STR, ASC:
					sym.Size += ld
				case HEX:
					if ld%2 != 0 {
						ld += 1
					}
					sym.Size += ld / 2
				case ADR:
					sym.Size += 2
				default:
					c.errorf("%d unsupported raw data %s", sym.Stmt.Line, sym.Stmt)
				}
			}
			if sym.Stmt.Mnemonic == STR {
				sym.Size += 1
			}

			continue
		}

		switch sym.Stmt.Mnemonic {
		case LSR, ASL:
			sym.Stmt.Mode = ins.Accumulator
		}

		i := ins.GetNameTable(sym.Stmt.Mnemonic.String(), sym.Stmt.Mode.String())
		if i.Bytes < 1 || i.Bytes > 3 {
			c.errorf("unsupported bytes length:%d sym:%s mod:%s", i.Bytes, sym.Stmt.Mnemonic, sym.Stmt.Mode)
		}
		sym.Size = i.Bytes
	}
}
