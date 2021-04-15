package main

import (
	"flag"
	"fmt"
	"go6502/lisa"
	"os"
	"path/filepath"
)

var (
	in       = flag.String("i", "", "input file")
	out      = flag.String("o", "", "output object file")
	debugIns = flag.Bool("L", false, "print instruction")
)

func main() {
	flag.Parse()
	if *in == "" {
		flag.Usage()
		return
	}

	ext := filepath.Ext(*in)
	if ext != ".s" {
		flag.Usage()
		return
	}

	inf := *in
	of := *out
	if of == "" {
		of = inf[:len(inf)-len(ext)] + "o"
	}
	err := load(*in, of)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func load(in, of string) (err error) {
	ipf, err := os.Open(in)
	if err != nil {
		return
	}

	ol, err := lisa.Parse(ipf)
	if err != nil {
		return err
	}
	err = encObj(of, ol)
	return
}

func encObj(of string, ol []*lisa.Stmt) (err error) {
	for i := range ol {
		if ol[i].Mnemonic == 0 && ol[i].Comment != "" {
			continue
		}
		if *debugIns {
			fmt.Println(ol[i])
		}
	}
	return
}
