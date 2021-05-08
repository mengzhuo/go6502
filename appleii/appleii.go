package appleii

import (
	"encoding/json"
	"go6502/cpu"
	"io"
	"log"
	"strings"
	"time"

	"github.com/rivo/tview"
)

type AppleII struct {
	cpu *cpu.CPU `json:-`
	app *tview.Application
	in  *tview.TextView `json:-`
	Mem *cpu.SimpleMem  `json:-`

	RefreshMs int `json:"refresh_ms"`
}

func New() *AppleII {
	return &AppleII{}
}

func (a *AppleII) Init(in io.Reader) (err error) {
	d := json.NewDecoder(in)
	err = d.Decode(a)
	a.Mem = &cpu.SimpleMem{}
	a.cpu = cpu.New()
	a.in = tview.NewTextView()

	a.in.SetScrollable(false)
	a.in.SetBorder(true)
	a.in.SetMaxLines(24)

	if a.RefreshMs <= 10 {
		a.RefreshMs = 1000 / 60
	}

	return
}

func (a *AppleII) draw40Cols() {
	var bd strings.Builder
	// First section
	bd.Write(a.Mem[0x400:0x427])
	bd.Write(a.Mem[0x480:0x4a7])
	bd.Write(a.Mem[0x500:0x527])
	bd.Write(a.Mem[0x580:0x5a7])
	bd.Write(a.Mem[0x600:0x627])
	bd.Write(a.Mem[0x680:0x6a7])
	bd.Write(a.Mem[0x700:0x727])
	bd.Write(a.Mem[0x780:0x7a7])
	// MIDDLE section
	bd.Write(a.Mem[0x428:0x44f])
	bd.Write(a.Mem[0x4a8:0x4fc])
	bd.Write(a.Mem[0x528:0x54f])
	bd.Write(a.Mem[0x5a8:0x5fc])
	bd.Write(a.Mem[0x628:0x64f])
	bd.Write(a.Mem[0x6a8:0x6fc])
	bd.Write(a.Mem[0x728:0x74f])
	bd.Write(a.Mem[0x7a8:0x7fc])
	// Bottom Section
	bd.Write(a.Mem[0x450:0x477])
	bd.Write(a.Mem[0x4d0:0x4f7])
	bd.Write(a.Mem[0x550:0x577])
	bd.Write(a.Mem[0x5d0:0x5f7])
	bd.Write(a.Mem[0x650:0x677])
	bd.Write(a.Mem[0x6d0:0x6f7])
	bd.Write(a.Mem[0x750:0x777])
	bd.Write(a.Mem[0x7d0:0x7f7])
	a.in.SetText(bd.String())
}

const logo = `  _____  _ _   _    ___  ___ 
 |_  / || | | | |  / _ \/ __|
  / /| __ | |_| |  | (_) \__ \
 /___|_||_|\___/   \___/|___/`

func (a *AppleII) Run() {

	go func() {
		tick := time.NewTicker(time.Duration(a.RefreshMs) * time.Millisecond)
		for {
			<-tick.C
			a.cpu.NMI <- true
			a.draw40Cols()
			a.app.Draw()
		}
	}()

	app := tview.NewApplication()
	grid := tview.NewGrid()
	//grid.SetBorder(true)
	grid.SetRows(40)
	grid.SetColumns(80)
	grid.AddItem(a.in, 0, 0, 1, 3, 0, 0, false)

	a.app = app
	app.SetInputCapture(a.uiKeyPress)

	a.cpu.DurPerCycles = time.Millisecond
	a.cpu.ResetF(a)
	go a.cpu.Run(a)
	if err := app.SetRoot(grid, true).SetFocus(grid).Run(); err != nil {
		log.Fatal(err)
	}
}

func (a *AppleII) ReadByte(pc uint16) (b uint8) {
	return a.Mem.ReadByte(pc)
}

func (a *AppleII) ReadWord(pc uint16) (n uint16) {
	return a.Mem.ReadWord(pc)
}

func (a *AppleII) WriteByte(pc uint16, b uint8) {
	switch pc {
	case ClearKeyStrobe:
		// reset keyboard
		a.cpu.IRQ = false
		a.Mem[KeyboardData] &= 0x7f // clear higher bit
	}
	a.Mem.WriteByte(pc, b)
}

func (a *AppleII) WriteWord(pc uint16, n uint16) {
	a.Mem.WriteWord(pc, n)
}
