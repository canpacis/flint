package common

import (
	"encoding/binary"
)

type Link string

func (l Link) MarshalBinary() ([]byte, error) {
	buf := make([]byte, len(l)+2)
	binary.LittleEndian.PutUint16(buf[0:2], uint16(len(l)))
	copy(buf[2:], l)
	return buf, nil
}

func (l *Link) UnmarshalBinary(data []byte) error {
	length := int(binary.LittleEndian.Uint16(data[0:2]))
	*l = Link(data[2 : 2+length])
	return nil
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

func (m *Module) putHeader(b []byte) (n int) {
	binary.LittleEndian.PutUint32(b[n:], uint32(m.Version))
	n += 4
	binary.LittleEndian.PutUint32(b[n:], uint32(m.Len()))
	n += 4
	length := len(m.Name)
	binary.LittleEndian.PutUint16(b[n:], uint16(length))
	n += 2
	n += copy(b[n:], []byte(m.Name))
	return
}

func (m *Module) readHeader(b []byte) (n int) {
	m.Version = Version(binary.LittleEndian.Uint32(b[n:]))
	n += 4
	n += 4                                           // skip module length
	length := int(binary.LittleEndian.Uint16(b[n:])) // name length
	n += 2
	name := make([]byte, length)
	n += copy(name, b[n:n+length])
	m.Name = string(name)
	return
}

func (m *Module) putPool(b []byte, pool *Pool) int {
	off := 0
	binary.LittleEndian.PutUint32(b[off:], uint32(pool.Len()))
	off += 4
	off += copy(b[off:], pool.Bytes())
	return off
}

func (m *Module) readPool(b []byte, pool *Pool) int {
	off := 0
	length := int(binary.LittleEndian.Uint32(b[off:]))
	off += 4
	off += copy(pool.data[:], b[off:off+length])
	pool.pointer = length
	return off
}

func (m *Module) MarshalBinary() ([]byte, error) {
	buf := make([]byte, m.Len())
	off := 0
	off += m.putHeader(buf[off:])
	off += m.putPool(buf[off:], m.Links)
	off += m.putPool(buf[off:], m.Types)
	off += m.putPool(buf[off:], m.Consts)
	return buf, nil
}

func (m *Module) UnmarshalBinary(data []byte) error {
	off := 0
	off += m.readHeader(data[off:])
	off += m.readPool(data[off:], m.Links)
	off += m.readPool(data[off:], m.Types)
	off += m.readPool(data[off:], m.Consts)
	return nil
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
