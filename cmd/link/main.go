package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"go6502/zhuos/zp"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

type ZObj struct {
	Offset uint16
	Size   uint16
	Data   []byte
}

type Link struct {
	errCount int
	out      string
	fileList []string
	ol       []*ZObj
}

func main() {
	fl := os.Args
	if len(fl) < 3 {
		log.Fatal("link: out [obj file list...]")
	}
	l := &Link{
		out:      fl[1],
		fileList: fl[2:],
	}
	fmt.Println(l)
	for _, f := range l.fileList {
		l.parse(f)
	}
	if l.errCount > 0 {
		log.Fatal("load failed")
	}
	l.makeout()
	if l.errCount > 0 {
		log.Fatal("load failed")
	}
}

func (l *Link) Errorf(f string, args ...interface{}) {
	if l.errCount > 10 {
		log.Fatal("too many error")
	}
	l.errCount++
	log.Printf(f, args)
}

func (l *Link) parse(f string) {
	fd, err := os.Open(f)
	if err != nil {
		l.Errorf("parse:%s", err)
		return
	}
	defer fd.Close()

	stat, err := fd.Stat()
	if err != nil {
		l.Errorf("parse:%s", err)
		return
	}

	obj := &ZObj{
		Size: uint16(stat.Size() - 2),
	}
	err = binary.Read(fd, binary.LittleEndian, &obj.Offset)
	if err != nil {
		l.Errorf("parse:%s", err)
		return
	}
	buf := bytes.NewBuffer(nil)
	n, err := io.Copy(buf, fd)
	if err != nil || n != int64(obj.Size) {
		l.Errorf("parse: n=%d size=%d err=%s", n, obj.Size, err)
		return
	}
	obj.Data = buf.Bytes()
	l.ol = append(l.ol, obj)
}

func (l *Link) makeout() {
	zf := &zp.ZhuProg{}
	of, err := ioutil.TempFile("", "link-")
	if err != nil {
		l.Errorf("out: %s", err)
		return
	}

	zf.HdrNum = uint8(len(l.ol))

	for _, p := range l.ol {
		hdr := &zp.ZHeader{
			ProgSize:   p.Size,
			ProgOffset: p.Offset,
		}
		zf.Headers = append(zf.Headers, hdr)
		zf.Progs = append(zf.Progs, p.Data)
	}

	err = zp.Encode(of, zf)
	if err != nil {
		l.Errorf("out: %s", err)
		return
	}
	of.Close()
	if !filepath.IsAbs(l.out) {
		l.out, err = filepath.Abs(l.out)
		if err != nil {
			l.Errorf("out:%s", err)
		}
	}
	err = os.Rename(of.Name(), l.out)
	if err != nil {
		l.Errorf("out:%s", err)
	}
}
