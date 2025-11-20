package common

import "encoding/binary"

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
	if err := a.Modules.Read(mod, int(a.entrymod)); err != nil {
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
	if err := mod.Consts.Read(fn, int(a.entryconst)); err != nil {
		return nil, err
	}
	return fn, nil
}

func (a *Archive) MarshalBinary() ([]byte, error) {
	length := a.Modules.Len()
	buf := make([]byte, length+4 /* entry mod index */ +4 /* entry const index */ +4 /* mod list length */)

	binary.LittleEndian.PutUint32(buf[0:], a.entrymod)
	binary.LittleEndian.PutUint32(buf[4:], a.entryconst)
	binary.LittleEndian.PutUint32(buf[8:], uint32(length))
	copy(buf[12:], a.Modules.Bytes())

	return buf, nil
}

func (a *Archive) UnmarshalBinary(data []byte) error {
	a.entrymod = binary.LittleEndian.Uint32(data[0:])
	a.entryconst = binary.LittleEndian.Uint32(data[4:])
	length := int(binary.LittleEndian.Uint32(data[8:]))
	copy(a.Modules.data[:length], data[12:12+length])
	return nil
}
