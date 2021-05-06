// zp is file format of Zhu os Program file format
package zp

// Magic: ZP
// Program Headers count: uint8

// headers:
// Magic: ZH (Zhu os Header)
// Size of this Program: uint16
// Program Offset: uint16
// File Offset: uint16

const (
	ZPMag uint16 = uint16('Z') | uint16('P')<<8
	ZHMag        = uint16('Z') | uint16('H')<<8
)

const ZhuProgSize = 3

type ZhuProg struct {
	Magic   uint16
	HdrNum  uint8
	Headers []*ZHeader
	Progs   [][]byte
}

const ZHeaderSize = 8

type ZHeader struct {
	Magic      uint16
	ProgSize   uint16
	ProgOffset uint16
	FileOffset uint16
}
