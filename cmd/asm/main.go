package main

import (
	"flag"
	"fmt"
	"go6502/a65"
	"os"
	"path/filepath"
)

var (
	in  = flag.String("i", "", "input a65 file")
	out = flag.String("o", "", "output object file")
)

func main() {
	flag.Parse()
	if *in == "" {
		flag.Usage()
		return
	}

	if filepath.Ext(*in) != ".a65" {
		flag.Usage()
		return
	}

	inf := *in
	of := *out
	if of == "" {
		of = inf[:len(inf)-3] + "o"
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

	ol, err := a65.Parse(ipf)
	if err != nil {
		return err
	}
	err = encObj(of, ol)
	return
}

func encObj(of string, ol []*a65.Stmt) (err error) {
	for i := range ol {
		fmt.Println(ol[i])
	}
	return
}
