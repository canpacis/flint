package common

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

const POOL_SIZE = 4096

var ErrPoolKeyExists = errors.New("key already exists")

type Pool struct {
	data     [POOL_SIZE]byte
	indicies map[int]int
	wp       int
	rp       int
}

func (p *Pool) Write(b []byte) (int, error) {
	if len(b)+p.wp >= POOL_SIZE {
		return 0, fmt.Errorf("pool overflow")
	}
	n := copy(p.data[p.wp:], b)
	p.wp += n
	return n, nil
}

func (p *Pool) Read(b []byte) (int, error) {
	if p.rp >= POOL_SIZE {
		return 0, io.EOF
	}
	n := copy(b, p.data[p.rp:])
	p.rp += n
	return n, nil
}

func (p *Pool) Set(key int, value io.WriterTo) (int, error) {
	if p.Has(key) {
		return 0, fmt.Errorf("%w: %d", ErrPoolKeyExists, key)
	}
	pointer := p.wp
	if _, err := value.WriteTo(p); err != nil {
		return 0, err
	}
	p.indicies[key] = pointer
	return pointer, nil
}

func (p *Pool) Get(n int, value io.ReaderFrom) error {
	p.rp = n
	_, err := value.ReadFrom(p)
	return err
}

func (p *Pool) Has(key int) bool {
	_, ok := p.indicies[key]
	return ok
}

func (p *Pool) Lookup(key int) int {
	pointer, ok := p.indicies[key]
	if !ok {
		return -1
	}
	return pointer
}

func (p *Pool) Len() int {
	return p.wp
}

func (p *Pool) Bytes() []byte {
	return p.data[:p.wp]
}

func (p *Pool) WriteTo(w io.Writer) (n int64, err error) {
	if err := binary.Write(w, binary.LittleEndian, uint32(p.Len())); err != nil {
		return n, err
	} else {
		n += 4
	}

	if m, err := w.Write(p.Bytes()); err != nil {
		return n, err
	} else {
		n += int64(m)
	}
	return
}

func (p *Pool) ReadFrom(r io.Reader) (n int64, err error) {
	var length uint32
	if err := binary.Read(r, binary.LittleEndian, &length); err != nil {
		return n, err
	} else {
		n += 4
		p.wp = int(length)
	}

	if m, err := r.Read(p.data[:length]); err != nil {
		return n, err
	} else {
		n += int64(m)
	}
	return
}

func NewPool() *Pool {
	return &Pool{
		indicies: make(map[int]int),
	}
}
