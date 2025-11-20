package common

import (
	"encoding"
	"fmt"
)

const POOL_SIZE = 4096

type Pool struct {
	data     [POOL_SIZE]byte
	indicies map[int]int
	pointer  int
}

func (p *Pool) Write(m encoding.BinaryMarshaler, idx int) (int, error) {
	if p.Has(idx) {
		return 0, fmt.Errorf("failed to add to pool: index %d already exists", idx)
	}
	pointer := p.pointer
	data, err := m.MarshalBinary()
	if err != nil {
		return 0, fmt.Errorf("failed to add to pool: %w", err)
	}
	n := copy(p.data[p.pointer:], data)
	p.pointer += n
	p.indicies[idx] = pointer
	return pointer, nil
}

func (p *Pool) Has(idx int) bool {
	_, ok := p.indicies[idx]
	return ok
}

func (p *Pool) Get(idx int) int {
	pointer, ok := p.indicies[idx]
	if !ok {
		return -1
	}
	return pointer
}

func (p *Pool) Read(u encoding.BinaryUnmarshaler, pointer int) error {
	if err := u.UnmarshalBinary(p.data[pointer:]); err != nil {
		return fmt.Errorf("failed to read pool value: %w", err)
	}
	return nil
}

func (p *Pool) Len() int {
	return p.pointer
}

func (p *Pool) Bytes() []byte {
	return p.data[:p.pointer]
}

func NewPool() *Pool {
	return &Pool{
		indicies: make(map[int]int),
	}
}
