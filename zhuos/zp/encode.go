package zp

import (
	"encoding/binary"
	"fmt"
	"io"
)

func Encode(wr io.WriteSeeker, f *ZhuProg) (err error) {
	if len(f.Progs) != len(f.Headers) {
		return fmt.Errorf("prog not match header")
	}

	f.HdrNum = uint8(len(f.Headers))
	f.Magic = ZPMag

	err = binary.Write(wr, en, f.Magic)
	if err != nil {
		return
	}
	err = binary.Write(wr, en, f.HdrNum)
	if err != nil {
		return
	}
	hdrTotal := ZhuProgSize + ZHeaderSize*len(f.Headers)
	fileOffset := uint16(hdrTotal)
	for i, hdr := range f.Headers {
		hdr.ProgSize = uint16(len(f.Progs[i]))
		hdr.FileOffset = fileOffset
		fileOffset += hdr.ProgSize
	}

	for _, hdr := range f.Headers {
		fmt.Printf("prog:0x%x fo:0x%x\n", hdr.ProgSize, hdr.FileOffset)
	}

	for i, hdr := range f.Headers {
		ho := int64(ZhuProgSize + i*ZHeaderSize)

		_, err = wr.Seek(ho, io.SeekStart)
		if err != nil {
			return
		}

		err = binary.Write(wr, en, hdr)
		if err != nil {
			return
		}

		_, err = wr.Seek(int64(hdr.FileOffset), io.SeekStart)
		if err != nil {
			return
		}
		var n int
		n, err = wr.Write(f.Progs[i])
		if err != nil || n != int(hdr.ProgSize) {
			return
		}
	}
	return
}
