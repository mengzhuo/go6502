package main

import (
	"fmt"
	"go6502/zhuos/zp"
	"log"
	"os"
)

const hdrTmp = `|-- PO:0x%04X
+-- PS:0x%04X
`

func main() {
	if len(os.Args) < 2 {
		log.Fatal("nm program")
	}
	fd, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	zf := &zp.ZFile{}
	err = zp.Decode(fd, zf)
	if err != nil {
		log.Fatal(err)
	}

	for _, hdr := range zf.L {
		fmt.Printf(hdrTmp, hdr.Offset, len(hdr.Prog))
	}
}
