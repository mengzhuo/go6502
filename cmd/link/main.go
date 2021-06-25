package main

import (
	"flag"
	"fmt"
	"go6502/lisa"
	"go6502/zhuos/zp"
	"log"
	"os"
)

const conf = `{"text":4096, "bss":40960, "data":45056}`

var debugList = flag.Bool("L", true, "list symbols")

func main() {
	fl := os.Args
	if len(fl) < 3 {
		log.Fatal("link: out [obj file list...]")
	}
	l, err := lisa.NewLink(fl[2:], conf)
	if err != nil {
		log.Fatal(err)
	}
	l.Load()
	l.Resolve()
	for _, blk := range l.Block {
		for _, sym := range blk.List {
			if sym.Type == lisa.Constants {
				continue
			}
			if *debugList {
				fmt.Println(sym)
			}
		}
	}
	zf := &zp.ZFile{}

	for _, blk := range l.Block {
		if blk.Size == 0 {
			continue
		}
		zp := &zp.ZProg{
			Offset: blk.Origin.Addr,
		}
		for _, sym := range blk.List {
			if lisa.IsNonAddress(sym.Stmt.Mnemonic) {
				continue
			}
			data, err := sym.Data()
			if err != nil {
				l.Errorf("encode:%s", err)
				continue
			}
			zp.Prog = append(zp.Prog, data...)
		}
		zf.L = append(zf.L, zp)
	}
	of, err := os.OpenFile(fl[1], os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatal(err)
	}
	defer of.Close()
	err = zp.Encode(of, zf)
	if err != nil {
		log.Fatal(err)
	}

	if l.ErrCount > 0 {
		log.Fatal("link failed")
	}
}
