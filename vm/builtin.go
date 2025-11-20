package vm

import (
	"fmt"

	"github.com/canpacis/flint/common"
)

type BuiltinFn struct {
	name    string
	locals  int
	returns common.ConstType
	Fn      func(...*common.Const) (*common.Const, error)
}

func (f *BuiltinFn) Name() string {
	return f.name
}

func (f *BuiltinFn) Locals() int {
	return f.locals
}

func (f *BuiltinFn) Instructions() common.Instructions {
	var set common.Instructions
	if f.returns == common.InvalidConstType {
		set = append(set, common.NewOp(common.OpReturn)...)
	} else {
		set = append(set, common.NewOp(common.OpReturnValue)...)
	}
	return set
}

func NewBuiltinFn(
	name string,
	locals int,
	returns common.ConstType,
	fn func(...*common.Const,
	) (*common.Const, error)) *BuiltinFn {
	return &BuiltinFn{
		name:    name,
		locals:  locals,
		returns: returns,
		Fn:      fn,
	}
}

type Builtins struct {
	indicies []*common.Const
	names    map[string]int
	pointer  int
}

func (b *Builtins) Len() int {
	return b.pointer
}

func (b *Builtins) Register(name string, c *common.Const) {
	pointer := b.pointer
	b.indicies[pointer] = c
	b.names[name] = pointer
	b.pointer++
}

func (b *Builtins) Get(name string) int {
	idx, ok := b.names[name]
	if !ok {
		return -1
	}
	return idx
}

func (b *Builtins) Map() map[int]int {
	m := make(map[int]int, b.Len())
	for i := range b.indicies {
		m[i] = i
	}
	return m
}

func NewBuiltins() *Builtins {
	return &Builtins{
		indicies: make([]*common.Const, 1024),
		names:    make(map[string]int),
	}
}

func CreatePanic() *common.Const {
	var set common.Instructions
	set = append(set, common.NewOp(common.OpTrap)...)
	return common.NewConst(common.FnConst, common.NewCompiledFn("panic", 1, set))
}

type SyscallOp int

const (
	SyscallRead = SyscallOp(iota)
	SyscallWrite
)

func CreateSyscall(processor Processor) *common.Const {
	fn := NewBuiltinFn("syscall", 3, common.I64Const, func(args ...*common.Const) (*common.Const, error) {
		zero := common.NewConst(common.I64Const, 0)

		op, err := GetInt64(args[0])
		if err != nil {
			return zero, err
		}
		fd, err := GetInt64(args[1])
		if err != nil {
			return zero, err
		}
		data, err := GetData(args[2])
		if err != nil {
			return zero, err
		}

		switch SyscallOp(op) {
		case SyscallRead:
			return zero, nil
		case SyscallWrite:
			proccess := processor.Process()
			w, err := proccess.WriteDescriptors.Get(int(fd))
			if err != nil {
				return zero, fmt.Errorf("invalid syscall descriptor %d: %w", fd, err)
			}
			n, err := w.Write(data)
			if err != nil {
				return zero, fmt.Errorf("failed to write to buffer: %w", err)
			}
			return common.NewConst(common.I64Const, int64(n)), nil
		default:
			return zero, fmt.Errorf("invalid op argument for syscall %d", op)
		}
	})
	return common.NewConst(common.FnConst, fn)
}

func DefaultBuiltins(processor Processor) *Builtins {
	builtins := NewBuiltins()
	builtins.Register("panic", CreatePanic())
	builtins.Register("syscall", CreateSyscall(processor))
	return builtins
}
