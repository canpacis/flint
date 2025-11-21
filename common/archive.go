package common

import (
	"encoding/binary"
	"io"
)

type Archive struct {
	Modules    *Pool
	entrymod   uint32
	entryconst uint32
}

func NewArchive() *Archive {
	return &Archive{
		Modules: NewPool(),
	}
}

func (a *Archive) SetEntry(mod, c int) {
	a.entrymod = uint32(mod)
	a.entryconst = uint32(c)
}

func (a *Archive) MainModule() (*Module, error) {
	mod := NewModule("", 0)
	if err := a.Modules.Get(int(a.entrymod), mod); err != nil {
		return nil, err
	}
	return mod, nil
}

func (a *Archive) MainFn() (*Const, error) {
	mod, err := a.MainModule()
	if err != nil {
		return nil, err
	}
	fn := new(Const)
	if err := mod.Consts.Get(int(a.entryconst), fn); err != nil {
		return nil, err
	}
	return fn, nil
}

func (a *Archive) WriteTo(w io.Writer) (n int64, err error) {
	if err := binary.Write(w, binary.LittleEndian, a.entrymod); err != nil {
		return n, err
	} else {
		n += 4
	}
	if err := binary.Write(w, binary.LittleEndian, a.entryconst); err != nil {
		return n, err
	} else {
		n += 4
	}
	if m, err := a.Modules.WriteTo(w); err != nil {
		return n, err
	} else {
		n += m
	}
	return
}

func (a *Archive) ReadFrom(r io.Reader) (n int64, err error) {
	if err := binary.Read(r, binary.LittleEndian, &a.entrymod); err != nil {
		return n, err
	} else {
		n += 4
	}
	if err := binary.Read(r, binary.LittleEndian, &a.entryconst); err != nil {
		return n, err
	} else {
		n += 4
	}
	if m, err := a.Modules.ReadFrom(r); err != nil {
		return n, err
	} else {
		n += m
	}
	return
}
