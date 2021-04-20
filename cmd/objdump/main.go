package main

import (
	"flag"
	"fmt"
	"go6502/lisa"
	"io/ioutil"
	"log"
)

var (
	file  = flag.String("b", "", "binary file")
	fromI = flag.Int("from", 0, "offset from")
	toI   = flag.Int("to", 0, "offset to")
)

func main() {
	flag.Parse()
	if *file == "" {
		log.Println(*file, *fromI, *toI)
		flag.Usage()
		return
	}

	d, err := ioutil.ReadFile(*file)
	if err != nil {
		log.Fatal(err)
	}

	from := *fromI
	if from > len(d) {
		from = len(d)
	}
	to := *toI
	if to < from {
		to = from
	}

	sl, err := lisa.Disasm(d[from:to])
	if err != nil {
		log.Fatal(err)
	}

	for _, s := range sl {
		fmt.Println(s.Mnemonic, s.Oper)
	}
}
