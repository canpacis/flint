package common

import (
	"encoding/binary"
	"io"
)

type Link string

func (l *Link) WriteTo(w io.Writer) (n int64, err error) {
	if err := binary.Write(w, binary.LittleEndian, uint16(len(*l))); err != nil {
		return n, err
	} else {
		n += 4
	}
	if m, err := w.Write([]byte(*l)); err != nil {
		return n, err
	} else {
		n += int64(m)
	}
	return
}

func (l *Link) ReadFrom(r io.Reader) (n int64, err error) {
	var length uint16
	if err := binary.Read(r, binary.LittleEndian, &length); err != nil {
		return n, err
	}
	buf := make([]byte, length)
	if m, err := r.Read(buf); err != nil {
		return n, err
	} else {
		n += int64(m)
	}
	*l = Link(buf)
	return
}

func NewLink(name string) *Link {
	link := Link(name)
	return &link
}

type Module struct {
	Name    string
	Version Version
	Links   *Pool
	Types   *Pool
	Consts  *Pool
}

func (m *Module) headerSize() int {
	return 4 /* version */ + 4 /* mod length */ + 2 /* name length */ + len(m.Name)
}

func (mod *Module) writeHeader(w io.Writer) (n int64, err error) {
	if err := binary.Write(w, binary.LittleEndian, uint32(mod.Version)); err != nil {
		return n, err
	} else {
		n += 4
	}
	if err := binary.Write(w, binary.LittleEndian, uint32(mod.Len())); err != nil {
		return n, err
	} else {
		n += 4
	}
	if err := binary.Write(w, binary.LittleEndian, uint32(len(mod.Name))); err != nil {
		return n, err
	} else {
		n += 4
	}
	if m, err := w.Write([]byte(mod.Name)); err != nil {
		return n, err
	} else {
		n += int64(m)
	}
	return
}

func (mod *Module) readHeader(r io.Reader) (n int64, err error) {
	if err := binary.Read(r, binary.LittleEndian, &mod.Version); err != nil {
		return n, err
	} else {
		n += 4
	}
	var length uint32
	// Skip mod length
	if err := binary.Read(r, binary.LittleEndian, &length); err != nil {
		return n, err
	} else {
		n += 4
	}
	if err := binary.Read(r, binary.LittleEndian, &length); err != nil {
		return n, err
	} else {
		n += 4
	}
	buf := make([]byte, length)
	if m, err := r.Read(buf); err != nil {
		return n, err
	} else {
		n += int64(m)
		mod.Name = string(buf)
	}
	return
}

func (mod *Module) WriteTo(w io.Writer) (n int64, err error) {
	if m, err := mod.writeHeader(w); err != nil {
		return n, err
	} else {
		n += m
	}
	if m, err := mod.Links.WriteTo(w); err != nil {
		return n, err
	} else {
		n += m
	}
	if m, err := mod.Types.WriteTo(w); err != nil {
		return n, err
	} else {
		n += m
	}
	if m, err := mod.Consts.WriteTo(w); err != nil {
		return n, err
	} else {
		n += m
	}
	return
}

func (mod *Module) ReadFrom(r io.Reader) (n int64, err error) {
	if m, err := mod.readHeader(r); err != nil {
		return n, err
	} else {
		n += m
	}
	if m, err := mod.Links.ReadFrom(r); err != nil {
		return n, err
	} else {
		n += m
	}
	if m, err := mod.Types.ReadFrom(r); err != nil {
		return n, err
	} else {
		n += m
	}
	if m, err := mod.Consts.ReadFrom(r); err != nil {
		return n, err
	} else {
		n += m
	}
	return
}

func (m *Module) Len() int {
	return m.headerSize() +
		m.Links.Len() + 4 /* length size */ +
		m.Types.Len() + 4 /* length size */ +
		m.Consts.Len() + 4 /* length size */
}

func NewModule(name string, version Version) *Module {
	return &Module{
		Name:    name,
		Version: version,
		Links:   NewPool(),
		Types:   NewPool(),
		Consts:  NewPool(),
	}
}
