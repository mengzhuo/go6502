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

	//log := log.New(os.Stderr, "cpu", log.LstdFlags)
	c := New(nil)
	c.PC = 0x400
	err = c.Run(m)
	if err == nil {
		t.Fatal(err, c)
	}
	if c.PC != 0x3474 {
		t.Fatal(err, c)
	}
}

func TestHelloWorld(t *testing.T) {
	f, err := os.Open("../lisa/testdata/hello_world.out")
	if err != nil {
		t.Skip("no raw data", err)
	}
	defer f.Close()
	data, err := ioutil.ReadAll(f)
	if err != nil {
		t.Fatal(err)
	}
	m := &SimpleMem{}
	copy(m[0x400:], data)
	//log := log.New(os.Stderr, "cpu", log.LstdFlags)
	c := New(nil)
	c.PC = 0x400
	err = c.Run(m)
	t.Logf("%18q", string(m[0x42:0x42+18]))
}
