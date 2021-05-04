package main

import (
	"flag"
	"go6502/appleii"
	"io"
	"log"
	"strings"
)

var conf = flag.String("c", "", "config json file")

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
	com.Run()
}
