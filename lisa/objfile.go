package lisa

import (
	"encoding/json"
	"io"
)

type Block struct {
	Origin *Symbol        `json:"origin"`
	Table  map[string]int `json:"table"` // name -> list index
	List   []*Symbol      `json:"list"`
	Size   uint16         `json:"size"`
}

type ObjFile struct {
	Block []*Block `json:"block"`
}

func NewObjFile(sl []*Symbol) (of *ObjFile) {
	of = &ObjFile{
		Block: []*Block{&Block{Table: make(map[string]int)}},
	}
	blk := of.Block[0]

	if len(sl) > 1 && sl[0].Stmt.Mnemonic == ORG {
		blk.Origin = sl[0]
		sl = sl[1:]
	}

	for _, sym := range sl {
		if sym.Stmt.Mnemonic == ORG {
			blk = &Block{
				Origin: sym,
				List:   []*Symbol{sym},
				Table:  make(map[string]int),
			}
			of.Block = append(of.Block, blk)

			if sym.Stmt.Label != "" {
				blk.Table[sym.Stmt.Label] = 0
			}
			continue
		}
		blk.Size += uint16(sym.Size)
		blk.List = append(blk.List, sym)

		if sym.Stmt.Label != "" {
			blk.Table[sym.Stmt.Label] = len(blk.List) - 1
		}

	}
	return
}

func (of *ObjFile) WriteTo(wt io.Writer) (err error) {
	enc := json.NewEncoder(wt)
	enc.SetIndent("", "  ")
	err = enc.Encode(of)
	return
}
