package vm

import (
	"errors"
	"fmt"
	"strings"
)

var ErrStackOverflow = errors.New("stack overflow")
var ErrStackUnderflow = errors.New("stack underflow")

type Stack[T any] struct {
	data    []T
	pointer int
}

func (s *Stack[T]) Top() (T, error) {
	if s.pointer == 0 {
		var zero T
		return zero, ErrStackUnderflow
	}
	return s.data[s.pointer-1], nil
}

func (s *Stack[T]) Get(n int) (T, error) {
	if n >= s.pointer {
		var zero T
		return zero, ErrStackOverflow
	} else if n < 0 {
		var zero T
		return zero, ErrStackUnderflow
	}
	return s.data[n], nil
}

func (s *Stack[T]) Len() int {
	return s.pointer
}

func (s *Stack[T]) Push(value T) error {
	if s.pointer >= len(s.data) {
		return ErrStackOverflow
	}
	s.data[s.pointer] = value
	s.pointer++
	return nil
}

func (s *Stack[T]) Pop() (T, error) {
	if s.pointer == 0 {
		var zero T
		return zero, ErrStackUnderflow
	}
	s.pointer--
	return s.data[s.pointer], nil
}

func (s *Stack[T]) String() string {
	out := make([]string, s.Len())

	for i, item := range s.data[:s.pointer] {
		out[len(out)-i-1] = fmt.Sprintf("%v", item)
	}

	return strings.Join(out, "\n")
}

func NewStack[T any](size int) *Stack[T] {
	return &Stack[T]{
		data: make([]T, size),
	}
}
