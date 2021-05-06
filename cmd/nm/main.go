package main

import (
	"fmt"
	"go6502/zhuos/zp"
	"log"
	"os"
)

const hdrTmp = `|-- PO:0x%04X
|-- PS:0x%04X
+-- FO:0x%04X
`

func main() {
	if len(os.Args) < 2 {
		log.Fatal("nm program")
	}
	fd, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	zf := &zp.ZhuProg{}
	err = zp.Decode(fd, zf)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("HdrNum:%d\n", zf.HdrNum)

	for _, hdr := range zf.Headers {
		fmt.Printf(hdrTmp, hdr.ProgOffset, hdr.ProgSize, hdr.FileOffset)
	}
}
