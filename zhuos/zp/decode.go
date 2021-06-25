package zp

import (
	"encoding/gob"
	"io"
)

func Decode(rd io.ReadSeeker, f *ZFile) (err error) {
	dec := gob.NewDecoder(rd)
	err = dec.Decode(f)
	return
}
