package appleii

import (
	"encoding/json"
	"go6502/cpu"
	"go6502/zhuos/zp"
	"io"
	"log"
	"os"
	"time"

	"github.com/rivo/tview"
)

type AppleII struct {
	cpu *cpu.CPU
	app *tview.Application
	in  *tview.Table
	Mem *cpu.SimpleMem
	log *log.Logger

	RefreshMs int    `json:"refresh_ms"`
	Log       string `json:"log"`
	ROM       string `json:"rom"`
}

func New() *AppleII {
	return &AppleII{}
}

func (a *AppleII) Init(in io.Reader) (err error) {
	d := json.NewDecoder(in)
	err = d.Decode(a)

	a.Mem = &cpu.SimpleMem{}

	fd, err := os.Open(a.ROM)
	if err != nil {
		return
	}
	zps := &zp.ZhuProg{}
	err = zp.Decode(fd, zps)
	if err != nil {
		return
	}

	for i, hdr := range zps.Headers {
		copy(a.Mem[hdr.ProgOffset:], zps.Progs[i])
	}

	if a.RefreshMs <= 10 {
		a.RefreshMs = 1000 / 60
	}

	if a.Log != "" {
		fd, err := os.OpenFile("a2.log", os.O_CREATE|os.O_WRONLY|os.O_SYNC, 0700)
		if err != nil {
			return err
		}
		a.log = log.New(fd, "[CPU]", log.LstdFlags)
	}
	a.cpu = cpu.New(a.log)

	a.in = tview.NewTable()
	a.in.SetBorder(false)

	cell := tview.NewTableCell(" ")
	a.in.SetCell(23, 40, cell)

	for r := 0; r < 24; r++ {
		for c := 0; c < 40; c++ {
			cell = tview.NewTableCell("#")
			cell.SetSelectable(false)
			a.in.SetCell(r, c, cell)
		}
	}

	return
}

func (a *AppleII) Run() {

	a.fullScreen()

	go func() {
		tick := time.NewTicker(time.Duration(a.RefreshMs) * time.Millisecond)
		for {
			<-tick.C
			//a.cpu.NMI <- true
			a.app.Draw()
		}
	}()

	app := tview.NewApplication()
	app.EnableMouse(false)

	a.app = app
	app.SetInputCapture(a.uiKeyPress)

	a.cpu.DurPerCycles = time.Millisecond
	a.cpu.ResetF(a)

	go a.cpu.Run(a)
	if err := app.SetRoot(a.in, true).SetFocus(a.in).Run(); err != nil {
		log.Fatal(err)
	}
}

func (a *AppleII) fullScreen() {
	counter := uint8(0)
	for i := 0x400; i < 0x800; i++ {
		a.WriteByte(uint16(i), counter+'0')
		counter++
		if counter == 40 {
			counter = 0
		}

		if i&0xff > 0xf7 {
			a.WriteByte(uint16(i), '!')
			continue
		}
	}
}

func (a *AppleII) ReadByte(pc uint16) (b uint8) {
	b = a.Mem.ReadByte(pc)
	if a.log != nil {
		a.log.Printf("RB 0x%04X -> %x", pc, b)
	}
	return
}

func (a *AppleII) ReadWord(pc uint16) (n uint16) {
	n = a.Mem.ReadWord(pc)
	if a.log != nil {
		a.log.Printf("RW 0x%04X -> %x", pc, n)
	}
	return
}

func (a *AppleII) WriteByte(pc uint16, b uint8) {
	switch pc {
	case ClearKeyStrobe:
		// reset keyboard
		a.cpu.IRQ = false
		a.Mem[KeyboardData] &= 0x7f // clear higher bit
	}

	if pc >= 0x400 && pc <= 0x7ff {
		// must be printable
		pb := b
		if b < 0x20 || b > 0x7f {
			pb = ' '
		}
		a.pcToScreen(pc, pb)
	}

	if a.log != nil {
		a.log.Printf("WB 0x%04X -> %x", pc, b)
	}
	a.Mem.WriteByte(pc, b)
}

func (a *AppleII) WriteWord(pc uint16, n uint16) {
	if a.log != nil {
		a.log.Printf("WW 0x%04X -> %x", pc, n)
	}
	a.Mem.WriteWord(pc, n)
}

var screenRowTable = [...]uint16{
	0: 0,
	1: 8,
	2: 15,
	3: 1,
	4: 9,
	5: 16,
	6: 255,
}

func (a AppleII) pcToScreen(pc uint16, pb byte) {

	low := pc & 0xff
	row := 0

	if low >= 0x80 {
		low -= 0x80
		row += 1
	}

	ll := low / 40
	if ll >= 3 {
		return
	}

	row += int(ll)
	row = int(screenRowTable[row] + (pc >> 8) - 4)
	col := int(pc) % 40

	c := a.in.GetCell(row, col)
	c.SetText(string(pb))
}
