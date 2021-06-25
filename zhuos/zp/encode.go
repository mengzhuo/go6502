package zp

import (
	"encoding/gob"
	"io"
)

func Encode(wr io.WriteSeeker, f *ZFile) (err error) {
	enc := gob.NewEncoder(wr)
	err = enc.Encode(f)
	return
}
