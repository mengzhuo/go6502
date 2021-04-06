package cpu

import "fmt"

type SimpleMem [65536]byte

func (s *SimpleMem) ReadByte(pc uint16) (b uint8) {
	if debug {
		fmt.Printf("RB 0x%04X->0x%02X\n", pc, s[pc])
	}
	return s[pc]
}

func (s *SimpleMem) ReadWord(pc uint16) (n uint16) {
	n = uint16(s[pc])
	n |= uint16(s[pc+1]) << 8
	if debug {
		fmt.Printf("RW 0x%04X->0x%04X\n", pc, n)
	}
	return
}

func (s *SimpleMem) WriteByte(pc uint16, b uint8) {
	if debug {
		fmt.Printf("WB 0x%04X<-0x%02X\n", pc, b)
	}
	s[pc] = b
}

func (s *SimpleMem) WriteWord(pc uint16, n uint16) {
	s[pc] = uint8(n)
	s[pc+1] = uint8(n >> 8)
	if debug {
		fmt.Printf("WW 0x%04X<-0x%04X\n", pc, n)
	}
}
