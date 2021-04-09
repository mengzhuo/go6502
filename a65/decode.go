package a65

import (
	"encoding/binary"
	"fmt"
	"go6502/ins"
)

func Disasm(d []byte) (sl []*Stmt, err error) {

	for i := 0; i < len(d); {
		op := ins.Table[d[i]]
		if op.Bytes == 0 {
			err = fmt.Errorf("invalid op at %d", i)
			return
		}

		s := &Stmt{Offset: uint16(i) + 0x400, ins: op}
		switch op.Mode {
		case ins.Immediate, ins.ZeroPage, ins.ZeroPageX, ins.ZeroPageY,
			ins.IndirectX, ins.IndirectY:
			s.u8 = d[i+1]
		case ins.Relative:
			s.s8 = int8(d[i+1])
		case ins.Absolute, ins.AbsoluteX, ins.AbsoluteY, ins.Indirect:
			s.u16 = binary.LittleEndian.Uint16(d[i+1:])
		}
		i += int(op.Bytes)
		sl = append(sl, s)
	}
	return
}
