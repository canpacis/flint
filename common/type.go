package common

import "io"

type TypeField struct {
	Name string
	Type int
}

type Type struct {
	Name   string
	Fields []TypeField
}

func (t *Type) WriteTo(w io.Writer) (n int64, err error) {
	m, err := w.Write([]byte{42})
	return int64(m), err
}

func (t *Type) ReadFrom(r io.Reader) (n int64, err error) {
	buf := make([]byte, 1)
	m, err := r.Read(buf)
	return int64(m), err
}

func NewType() *Type {
	return &Type{}
}
