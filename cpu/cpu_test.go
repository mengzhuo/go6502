package cpu

import (
	"compress/gzip"
	"io/ioutil"
	"os"
	"testing"
)

func TestBinFunction(t *testing.T) {
	f, err := os.Open("testdata/6502_functional_test.bin.gz")
	if err != nil {
		t.Skip("no raw data", err)
	}
	defer f.Close()

	zr, err := gzip.NewReader(f)
	if err != nil {
		t.Fatal("gzip", err)
	}
	defer zr.Close()
	zr.Multistream(false)

	data, err := ioutil.ReadAll(zr)
	if err != nil {
		t.Fatal(err)
	}
	m := &SimpleMem{}
	n := copy(m[:], data)
	if err != nil || n != 0x10000 {
		t.Fatal(err, n)
	}
	c := New()
	c.PC = 0x400
	err = c.Run(m)
	if err == nil {
		t.Fatal(err, c)
	}
	if c.PC != 0x3474 {
		t.Fatal(err, c)
	}
}
