package lisa

import (
	"fmt"
	"go6502/ins"
)

func Disasm(d []byte) (sl []*Stmt, err error) {
	for len(d) > 0 {
		op := d[0]
		in := ins.Table[op]
		if in.Bytes == 0 {
			err = fmt.Errorf("Invalid op")
			return
		}
		s := &Stmt{Line: 0, Mnemonic: mnemonicMap[in.Name.String()]}

		switch in.Mode {
		case ins.Relative:
			s.Oper = fmt.Sprintf("*!%+d", int8(d[1])+2)
		case ins.ZeroPage, ins.Immediate:
			s.Oper = fmt.Sprintf("$%02X", d[1])
		case ins.ZeroPageX:
			s.Oper = fmt.Sprintf("$%02X,X", d[1])
		case ins.ZeroPageY:
			s.Oper = fmt.Sprintf("$%02X,Y", d[1])
		case ins.Absolute:
			s.Oper = fmt.Sprintf("$%02X%02X", d[2], d[1])
		case ins.AbsoluteX:
			s.Oper = fmt.Sprintf("$%02X%02X,X", d[2], d[1])
		case ins.AbsoluteY:
			s.Oper = fmt.Sprintf("$%02X%02X,Y", d[2], d[1])
		case ins.Indirect:
			s.Oper = fmt.Sprintf("($%02X%02X)", d[2], d[1])
		case ins.IndirectX:
			s.Oper = fmt.Sprintf("($%02X),X", d[1])
		case ins.IndirectY:
			s.Oper = fmt.Sprintf("($%02X),Y", d[1])
		default:
		}
		d = d[in.Bytes:]
		sl = append(sl, s)
	}

	return
}
