package main

import (
	"flag"
	"fmt"
	"go6502/appleii"
	"go6502/zhuos/zp"
	"io"
	"log"
	"os"
	"strings"
)

var (
	conf = flag.String("c", "", "config json file")
	rom  = flag.String("r", "", "ROM file path(ZhuOS only)")
)

const defConf = `{"refresh_ms":0, "log":"a2.log"}`

func main() {
	flag.Parse()
	var rd io.Reader
	if *conf == "" {
		rd = strings.NewReader(defConf)
	}

	com := appleii.New()
	if err := com.Init(rd); err != nil {
		log.Fatal(err)
	}

	if *rom != "" {
		fd, err := os.Open(*rom)
		if err != nil {
			log.Fatal(err)
		}
		zf := &zp.ZFile{}
		err = zp.Decode(fd, zf)
		if err != nil {
			log.Fatal(err)
		}
		for _, prog := range zf.L {
			fmt.Printf("%04x=>%x\n", prog.Offset, prog.Prog)
			copy(com.Mem[prog.Offset:], prog.Prog)
		}
	}

	com.Run()
}
