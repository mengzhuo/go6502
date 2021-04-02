package cpu

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestBinFunction(t *testing.T) {
	f, err := os.Open("testdata/6502_functional_test.bin")
	if err != nil {
		t.Skip("no raw data", err)
	}
	defer f.Close()

	data, err := ioutil.ReadAll(f)
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
		t.Fatal(err, c)
	}
}
