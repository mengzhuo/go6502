package main

import (
	"flag"
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

const defConf = `{"refresh_ms":0}`

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
		zf := &zp.ZhuProg{}
		err = zp.Decode(fd, zf)
		if err != nil {
			log.Fatal(err)
		}
		for i, hdr := range zf.Headers {
			copy(com.Mem[hdr.ProgOffset:], zf.Progs[i])
		}
	}

	com.Run()
}
