package vm

import (
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

var DefaultBuiltins = NewBuiltins()

func CreatePanic() *common.Const {
	var set common.Instructions
	set = append(set, common.NewOp(common.OpTrap)...)
	return common.NewConst(common.FnConst, common.NewCompiledFn("panic", 1, set))
}

func init() {
	DefaultBuiltins.Register("panic", CreatePanic())
}
