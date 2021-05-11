package appleii

import "github.com/gdamore/tcell/v2"

const (
	KeyboardData   = 0xc000
	ClearKeyStrobe = 0xc010
)

func (a *AppleII) uiKeyPress(event *tcell.EventKey) *tcell.EventKey {
	a.cpu.IRQ = true
	var b byte
	switch event.Key() {
	default:
		b = byte(event.Rune())
	}
	a.Mem[KeyboardData] = b | 0x80 // apple ii use higher
	return event
}
