package appleii

import (
	"github.com/rivo/tview"
)

var baseAddress = [...]uint16{
	0x400,
	0x480,
	0x500,
	0x580,
	0x600,
	0x680,
	0x700,
	0x780,
	0x428,
	0x4A8,
	0x528,
	0x5A8,
	0x628,
	0x6A8,
	0x728,
	0x7A8,
	0x450,
	0x4D0,
	0x550,
	0x5D0,
	0x650,
	0x6D0,
	0x750,
	0x7D0,
}

func (a *AppleII) initDisplay() {
	a.pixelMap = map[uint16]*tview.TableCell{}
	for i, ba := range baseAddress {
		for j := uint16(0); j < 40; j++ {
			a.pixelMap[ba+j] = a.in.GetCell(i, int(j))
		}
	}
}
