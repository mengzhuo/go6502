package main

import (
	"flag"
	"fmt"
	"go6502/lisa"
	"log"
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
		log.Println("input file not found")
		flag.Usage()
		return
	}

	ext := filepath.Ext(*in)

	inf := *in
	of := *out
	if of == "" {
		of = inf[:len(inf)-len(ext)] + ".out"
	}
	err := load(*in, of)
	if err != nil {
		log.Fatal(err)
	}

}

func load(in, of string) (err error) {
	ipf, err := os.Open(in)
	if err != nil {
		return
	}
	defer ipf.Close()

	ol, err := lisa.Parse(ipf)
	if err != nil {
		return err
	}

	if *debugIns {
		for i := range ol {
			if ol[i].Mnemonic == 0 && ol[i].Comment != "" {
				continue
			}
			fmt.Println(ol[i])
		}
	}

	outf, err := os.OpenFile(of, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0700)
	if err != nil {
		return
	}
	defer outf.Close()
	err = lisa.Compile(ol, outf)
	return
}
