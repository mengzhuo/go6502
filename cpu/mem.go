package cpu

type SimpleMem [65536]byte

func (s *SimpleMem) ReadByte(pc uint16) (b uint8) {
	return s[pc]
}

func (s *SimpleMem) ReadWord(pc uint16) (n uint16) {
	n = uint16(s[pc])
	n |= uint16(s[pc+1]) << 8
	return
}

func (s *SimpleMem) WriteByte(pc uint16, b uint8) {
	s[pc] = b
}

func (s *SimpleMem) WriteWord(pc uint16, n uint16) {
	s[pc] = uint8(n)
	s[pc+1] = uint8(n >> 8)
}
