package main

import (
	"fmt"
	"go6502/a65"
	"io/ioutil"
	"log"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("objdump binary")
	}
	d, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}

	sl, err := a65.Disasm(d)
	if err != nil {
		log.Fatal(err)
	}

	for _, s := range sl {
		fmt.Printf("0x%04x %v\n", s.Offset, s)
	}
}
