package zp

import (
	"encoding/binary"
	"fmt"
	"io"
)

var en = binary.LittleEndian

func Decode(rd io.ReadSeeker, f *ZhuProg) (err error) {
	f = &ZhuProg{}
	err = binary.Read(rd, en, f)
	if err != nil {
		return
	}

	if f.Magic != ZPMag {
		err = fmt.Errorf("invalid Zhu OS program: header %x", f.Magic)
		return
	}

	f.Headers = make([]*ZHeader, f.HdrNum)
	f.Progs = make([][]byte, f.HdrNum)

	var rn int

	for i := range f.Headers {
		ho := int64(3 + i*8)
		_, err = rd.Seek(ho, io.SeekStart)
		if err != nil {
			return
		}

		hdr := &ZHeader{}
		err = binary.Read(rd, en, hdr)
		if err != nil || hdr.Magic != ZHMag {
			return
		}
		f.Headers[i] = hdr
		prog := make([]byte, hdr.ProgSize)

		_, err = rd.Seek(int64(hdr.FileOffset), io.SeekStart)
		if err != nil {
			return
		}

		rn, err = rd.Read(prog)
		if err != nil || rn != int(hdr.ProgSize) {
			err = fmt.Errorf("read prog failed:%s expected = %d, got=%d", err, hdr.ProgSize, rn)
			return
		}
		f.Progs[i] = prog
	}
	return
}
