// zp is file format of ZHUOS Program file format
package zp

type ZFile struct {
	L []*ZProg
}

type ZProg struct {
	Offset uint16
	Prog   []byte
}
