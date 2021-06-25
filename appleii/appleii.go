package appleii

import (
	"encoding/json"
	"go6502/cpu"
	"io"
	"log"
	"os"
	"time"

	"github.com/rivo/tview"
)

type AppleII struct {
	cpu      *cpu.CPU
	app      *tview.Application
	in       *tview.Table
	Mem      *cpu.SimpleMem
	log      *log.Logger
	pixelMap map[uint16]*tview.TableCell

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

	if a.RefreshMs <= 10 {
		a.RefreshMs = 1000 / 60
	}

	if a.Log != "" {
		fd, err := os.OpenFile(a.Log, os.O_CREATE|os.O_WRONLY|os.O_SYNC, 0700)
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
			cell = tview.NewTableCell(" ")
			cell.SetSelectable(false)
			a.in.SetCell(r, c, cell)
		}
	}
	a.initDisplay()

	return
}

func (a *AppleII) Run() {

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

	//a.fullScreen()

	go a.cpu.Run(a)
	if err := app.SetRoot(a.in, true).SetFocus(a.in).Run(); err != nil {
		log.Fatal(err)
	}
}

func (a *AppleII) fullScreen() {
	for _, ba := range baseAddress {
		for j := uint16(0); j < 40; j++ {
			a.WriteByte(ba+j, byte(j)+'0')
		}
	}
}

func (a *AppleII) ReadByte(pc uint16) (b uint8) {
	b = a.Mem.ReadByte(pc)
	switch pc {
	case ClearKeyStrobe:
		// reset keyboard
		a.cpu.IRQ &= keyIRQMask
	}
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

	if pc >= 0x400 && pc <= 0x7ff {
		// must be printable
		if cell := a.pixelMap[pc]; cell != nil {
			cell.SetText(string(b))
		}
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
