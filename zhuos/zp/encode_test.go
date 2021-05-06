package zp

import (
	"io"
	"io/ioutil"
	"testing"
)

func TestEncode(t *testing.T) {
	zf := &ZhuProg{}
	zf.Headers = []*ZHeader{
		{ProgOffset: 0x4000},
		{ProgOffset: 0x8000},
	}

	zf.Progs = [][]byte{
		[]byte("Hello"),
		[]byte("World"),
	}
	tf, err := ioutil.TempFile("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer tf.Close()
	err = Encode(tf, zf)
	if err != nil {
		t.Error(err)
	}
	_, err = tf.Seek(0, io.SeekStart)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(tf.Name())
}
