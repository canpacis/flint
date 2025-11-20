package vm

import (
	"errors"
	"fmt"

	"github.com/canpacis/flint/common"
)

var ErrOpFetchFailed = errors.New("op fetch failed")

type Frame struct {
	fn  common.Fn
	mod *common.Module
	ip  int
	bp  int
}

func (f *Frame) String() string {
	return f.fn.Name()
}

func (f *Frame) Fetch() (common.OpCode, []int, error) {
	instructions := f.fn.Instructions()
	if f.ip >= len(instructions) {
		return 0, nil, fmt.Errorf("%w: pointer is reading outside of function instructions", ErrOpFetchFailed)
	}
	b := instructions[f.ip]
	def, err := common.LookupOp(b)
	if err != nil {
		return 0, nil, fmt.Errorf("%w: %w", ErrOpFetchFailed, err)
	}
	if f.ip+def.Width() > len(instructions) {
		return 0, nil, fmt.Errorf("%w: pointer is reading outside of function instructions", ErrOpFetchFailed)
	}
	operands, off := common.ReadOperands(def, instructions[f.ip+1:])
	f.ip += off + 1
	return common.OpCode(b), operands, nil
}

func NewFrame(fn common.Fn, mod *common.Module, bp int) *Frame {
	return &Frame{
		fn:  fn,
		mod: mod,
		ip:  0,
		bp:  bp,
	}
}
