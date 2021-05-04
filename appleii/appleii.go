package appleii

import (
	"encoding/json"
	"go6502/cpu"
	"io"
	"io/ioutil"
	"log"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type AppleII struct {
	cpu *cpu.CPU `json:-`
	app *tview.Application
	in  *tview.TextView `json:-`
	mem *cpu.SimpleMem  `json:-`

	RefreshMs int `json:"refresh_ms"`
}

func New() *AppleII {
	return &AppleII{}
}

func (a *AppleII) Init(in io.Reader) (err error) {
	d := json.NewDecoder(in)
	err = d.Decode(a)
	a.mem = &cpu.SimpleMem{}
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
	bd.Write(a.mem[0x400:0x427])
	bd.Write(a.mem[0x480:0x4a7])
	bd.Write(a.mem[0x500:0x527])
	bd.Write(a.mem[0x580:0x5a7])
	bd.Write(a.mem[0x600:0x627])
	bd.Write(a.mem[0x680:0x6a7])
	bd.Write(a.mem[0x700:0x727])
	bd.Write(a.mem[0x780:0x7a7])
	// MIDDLE section
	bd.Write(a.mem[0x428:0x44f])
	bd.Write(a.mem[0x4a8:0x4fc])
	bd.Write(a.mem[0x528:0x54f])
	bd.Write(a.mem[0x5a8:0x5fc])
	bd.Write(a.mem[0x628:0x64f])
	bd.Write(a.mem[0x6a8:0x6fc])
	bd.Write(a.mem[0x728:0x74f])
	bd.Write(a.mem[0x7a8:0x7fc])
	// Bottom Section
	bd.Write(a.mem[0x450:0x477])
	bd.Write(a.mem[0x4d0:0x4f7])
	bd.Write(a.mem[0x550:0x577])
	bd.Write(a.mem[0x5d0:0x5f7])
	bd.Write(a.mem[0x650:0x677])
	bd.Write(a.mem[0x6d0:0x6f7])
	bd.Write(a.mem[0x750:0x777])
	bd.Write(a.mem[0x7d0:0x7f7])
	a.in.SetText(bd.String())
}

const logo = `  _____  _ _   _    ___  ___ 
 |_  / || | | | |  / _ \/ __|
  / /| __ | |_| |  | (_) \__ \
 /___|_||_|\___/   \___/|___/`

func (a *AppleII) Run() {
	tick := time.NewTicker(time.Duration(a.RefreshMs) * time.Millisecond)

	go func() {
		for {
			<-tick.C
			a.draw40Cols()
			a.app.Draw()
		}
	}()

	data, err := ioutil.ReadFile("lisa/testdata/simple_kbd.out")
	if err != nil {
		log.Fatal(err)
	}
	copy(a.mem[0xD000:], data)
	copy(a.mem[0xe000:], logo)

	a.mem.WriteWord(0xfffe, 0xD011) // irq
	a.mem.WriteWord(0xfffc, 0xD000) // start at

	app := tview.NewApplication()
	grid := tview.NewGrid()
	//grid.SetBorder(true)
	grid.SetRows(40)
	grid.SetColumns(80)
	grid.AddItem(a.in, 0, 0, 1, 3, 0, 0, false)
	a.app = app
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		a.cpu.IRQ = true
		a.mem[0xc010] = 1
		var b byte
		switch event.Key() {
		case tcell.KeyEnter:
			b = '\n'
		case tcell.KeyBackspace:

		default:
			b = byte(event.Rune())
		}
		a.mem[0xc000] = b
		return event
	})

	a.cpu.DurPerCycles = time.Millisecond
	a.cpu.ResetF(a)
	go a.cpu.Run(a)
	if err := app.SetRoot(grid, true).SetFocus(grid).Run(); err != nil {
		log.Fatal(err)
	}
}

func (a *AppleII) ReadByte(pc uint16) (b uint8) {
	b = a.mem.ReadByte(pc)
	switch pc {
	case 0xc010:
		// reset keyboard
		a.mem.WriteByte(pc, 0)
	}
	return
}

func (a *AppleII) ReadWord(pc uint16) (n uint16) {
	return a.mem.ReadWord(pc)
}

func (a *AppleII) WriteByte(pc uint16, b uint8) {
	switch pc {
	// we can't find release
	case 0x42:
		a.cpu.IRQ = false
		a.mem[0xc010] = 0
	default:
		a.mem.WriteByte(pc, b)
	}
}

func (a *AppleII) WriteWord(pc uint16, n uint16) {
	a.mem.WriteWord(pc, n)
}
