package lisa

import "go6502/ins"

type ObjFile struct {
	Origin uint16
	tb     map[string]*Obj
}

type Obj struct {
	s *Stmt

	ins   ins.Ins
	final bool
}

func Compile(sl []*Stmt) (of *ObjFile, err error) {
	of = &ObjFile{}
	// Evalution
	for ev := 1; ev > 0; {
		ev = 0
	}
	return
}

func (of *ObjFile) findConst(sl []*Stmt) (err error) {
	el := []error{}

	for _, s := range sl {
		if len(el) > 10 {
			return elToError(el)
		}
		switch s.Mnemonic {
		case ORG:
			if of.Origin != 0 {
				el = append(el, s.NE("duplicate ORG"))
				continue
			}
		}
	}

	return elToError(el)
}
