package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"go6502/ins"
	"io"
	"log"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("objdump binary")
	}
	f, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	l, err := disasm(f)
	if err != nil && err != io.EOF {
		log.Fatal(err)
	}
	for _, o := range l {
		if o.Ins.Mode != ins.Implied {
			fmt.Printf("0x%04x %v 0x%04x\n", o.Offset, o.Ins, o.Oper)
			continue
		}
		fmt.Printf("0x%04x %v\n", o.Offset, o.Ins)
	}
}

type Obj struct {
	Offset int
	Ins    ins.Ins
	Oper   uint16
}

func disasm(f *os.File) (l []*Obj, err error) {
	b := bufio.NewReader(f)
	offset := 0
	for {
		op, err := b.ReadByte()
		if err != nil {
			return l, err
		}
		i := ins.Table[op]
		if i.Cycles == 0 {
			obj := &Obj{
				Ins:    ins.Table[ins.NOPImplied],
				Offset: offset,
			}
			l = append(l, obj)
			offset += 1
			continue
		}

		tmp := make([]byte, i.Bytes-1)
		n, err := b.Read(tmp)
		if n != int(i.Bytes-1) || err != nil {
			return l, fmt.Errorf("read %d err=%v", n, err)
		}
		obj := &Obj{
			Ins:    i,
			Offset: offset,
		}

		switch i.Bytes {
		case 2:
			obj.Oper = uint16(tmp[0])
		case 3:
			obj.Oper = binary.LittleEndian.Uint16(tmp)
		}
		l = append(l, obj)
		offset += int(i.Bytes)
	}
}
