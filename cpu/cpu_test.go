package cpu

import (
	"archive/zip"
	"io/ioutil"
	"testing"
)

func TestBinFunction(t *testing.T) {
	f, err := zip.OpenReader("testdata/6502_functional_test.bin.zip")
	if err != nil {
		t.Skip("no raw data", err)
	}
	defer f.Close()

	zf, err := f.File[0].Open()
	data, err := ioutil.ReadAll(zf)
	if err != nil {
		t.Fatal(err)
	}
	m := &SimpleMem{}
	n := copy(m[:], data)
	if err != nil || n != 0xffff {
		t.Fatal(err, n)
	}
	c := New()
	c.PC = 0x400
	err = c.Run(m)
	if err != nil {
		t.Fatal(err, n)
	}
}
