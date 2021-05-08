package zp

import (
	"encoding/binary"
	"fmt"
	"io"
)

var en = binary.LittleEndian

func Decode(rd io.ReadSeeker, f *ZhuProg) (err error) {
	err = binary.Read(rd, en, &f.Magic)
	if err != nil {
		return
	}
	err = binary.Read(rd, en, &f.HdrNum)
	if err != nil {
		return
	}

	if f.Magic != ZPMag {
		err = fmt.Errorf("invalid Zhu OS program: header %x", f.Magic)
		return
	}

	var rn int

	for i := 0; i < int(f.HdrNum); i++ {
		ho := int64(ZhuProgSize + i*ZHeaderSize)
		_, err = rd.Seek(ho, io.SeekStart)
		if err != nil {
			return
		}

		hdr := &ZHeader{}
		f.Headers = append(f.Headers, hdr)
		binary.Read(rd, en, hdr)

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
		f.Progs = append(f.Progs, prog)
	}
	return
}
