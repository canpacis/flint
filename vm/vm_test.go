package vm_test

import (
	"testing"

	"github.com/canpacis/flint/common"
	"github.com/canpacis/flint/vm"
	"github.com/stretchr/testify/assert"
)

func TestFrame(t *testing.T) {
	assert := assert.New(t)

	type FrameTest struct {
		Instructions     common.Instructions
		ExpectedOpCode   common.OpCode
		ExpectedOperands []int
		ExpectedError    error
	}

	tests := []FrameTest{
		{
			common.Instructions{},
			common.OpNoop,
			[]int{},
			vm.ErrOpFetchFailed,
		},
		{
			common.Instructions{byte(common.OpLoadBuiltin)},
			common.OpNoop,
			[]int{},
			vm.ErrOpFetchFailed,
		},
		{
			common.Instructions(common.NewOp(common.OpLoadConst, 256)),
			common.OpLoadConst,
			[]int{256},
			nil,
		},
		{
			common.Instructions(common.NewOp(common.OpLoadModConst, 64, 256)),
			common.OpLoadModConst,
			[]int{64, 256},
			nil,
		},
		{
			common.Instructions(common.NewOp(common.OpLoadBuiltin, 256)),
			common.OpLoadBuiltin,
			[]int{256},
			nil,
		},
		{
			common.Instructions(common.NewOp(common.OpLoadU32, 0)),
			common.OpLoadU32,
			[]int{0},
			nil,
		},
		{
			common.Instructions(common.NewOp(common.OpLoadU64, 0)),
			common.OpLoadU64,
			[]int{0},
			nil,
		},
	}

	for i, test := range tests {
		frame := vm.NewFrame(common.NewCompiledFn("test", 0, test.Instructions), nil, 0)
		code, operands, err := frame.Fetch()
		if test.ExpectedError != nil {
			assert.ErrorIsf(err, test.ExpectedError, "Test case %d", i)
		} else {
			assert.NoError(err, "Error: Test case %d", i)
			assert.Equalf(test.ExpectedOpCode, code, "OpCode: Test case %d", i)
			assert.Equalf(test.ExpectedOperands, operands, "Operands: Test case %d", i)
		}
	}
}

func TestLoadOps(t *testing.T) {
	assert := assert.New(t)

	type LoadOpTest struct {
		OpCode        common.OpCode
		Operands      []int
		ExpectedType  common.ConstType
		ExpectedValue any
		ExpectedError error
	}

	tests := []LoadOpTest{
		{common.OpNoop, []int{}, common.InvalidConstType, 0, vm.ErrUnknownOpCode},
		{common.OpLoadConst, []int{0}, common.I64Const, int64(1), nil},
		// {common.OpLoadModConst, []int{0, 0}, common.I64Const, int64(0), vm.ErrMissingConst},
		{common.OpLoadBuiltin, []int{0}, common.I64Const, int64(1), nil},
		{common.OpLoadBuiltin, []int{1}, common.InvalidConstType, 0, vm.ErrMissingConst},
		{common.OpLoadI32, []int{256}, common.I32Const, int32(256), nil},
		{common.OpLoadI64, []int{256}, common.I64Const, int64(256), nil},
		{common.OpLoadU32, []int{256}, common.U32Const, uint32(256), nil},
		{common.OpLoadU64, []int{256}, common.U64Const, uint64(256), nil},
	}

	var DefaultBuiltins = vm.NewBuiltins()
	DefaultBuiltins.Register("0", common.NewConst(common.I64Const, int64(1)))
	machine := vm.NewVM(DefaultBuiltins)
	mod := common.NewModule("main", 0)
	_, err := mod.Consts.Write(common.NewConst(common.I64Const, int64(1)), 0)
	assert.NoError(err)
	machine.Init(common.NewArchive())
	executor := vm.NewExecutor(machine)
	assert.NoError(
		executor.Frames().Push(vm.NewFrame(common.NewCompiledFn("main", 0, common.Instructions{}), mod, 0)),
	)

	for i, test := range tests {
		err := executor.ExecuteLoad(test.OpCode, test.Operands)
		if test.ExpectedError != nil {
			assert.ErrorIsf(err, test.ExpectedError, "Test case %d", i)
		} else {
			assert.NoErrorf(err, "Load: Test case %d", i)

			constant, err := executor.Stack().Top()
			assert.NoErrorf(err, "Stack: Test case %d", i)
			assert.Equalf(test.ExpectedType, constant.Type, "Type: Test case %d", i)
			assert.Equalf(test.ExpectedValue, constant.Value, "Value: Test case %d", i)
		}
	}
}

func TestBinaryOps(t *testing.T) {
	assert := assert.New(t)

	type BinaryOpTest struct {
		Left          *common.Const
		Right         *common.Const
		OpCode        common.OpCode
		ExpectedType  common.ConstType
		ExpectedValue any
		ExpectedError error
	}

	tests := []BinaryOpTest{
		{
			common.NewConst(common.I32Const, int32(5)), common.NewConst(common.I32Const, int32(7)),
			common.OpAddI64, common.InvalidConstType, 0, vm.ErrConstTypeInvalid,
		},
		{
			common.NewConst(common.I64Const, int64(5)), common.NewConst(common.I64Const, int64(7)),
			common.OpAddI64, common.I64Const, int64(12), nil,
		},
		{
			common.NewConst(common.I64Const, int64(5)), common.NewConst(common.I64Const, int64(7)),
			common.OpSubI64, common.I64Const, int64(-2), nil,
		},
		{
			common.NewConst(common.I64Const, int64(5)), common.NewConst(common.I64Const, int64(7)),
			common.OpMulI64, common.I64Const, int64(35), nil,
		},
		{
			common.NewConst(common.I64Const, int64(35)), common.NewConst(common.I64Const, int64(7)),
			common.OpDivI64, common.I64Const, int64(5), nil,
		},
		{
			common.NewConst(common.I64Const, int64(35)), common.NewConst(common.I64Const, int64(0)),
			common.OpDivI64, common.InvalidConstType, 0, vm.ErrDivideByZero,
		},
		{
			common.NewConst(common.I64Const, int64(37)), common.NewConst(common.I64Const, int64(7)),
			common.OpModI64, common.I64Const, int64(2), nil,
		},
		{
			common.NewConst(common.I64Const, int64(37)), common.NewConst(common.I64Const, int64(0)),
			common.OpModI64, common.InvalidConstType, 0, vm.ErrDivideByZero,
		},
	}

	var DefaultBuiltins = vm.NewBuiltins()
	machine := vm.NewVM(DefaultBuiltins)
	machine.Init(common.NewArchive())
	executor := vm.NewExecutor(machine)

	for i, test := range tests {
		stack := executor.Stack()
		stack.Push(test.Left)
		stack.Push(test.Right)

		err := executor.ExecuteBinary(test.OpCode, []int{})
		if test.ExpectedError != nil {
			assert.ErrorIsf(err, test.ExpectedError, "Test case %d", i)
		} else {
			assert.NoErrorf(err, "Test case %d", i)
			constant, err := stack.Pop()
			assert.NoErrorf(err, "Stack: Test case %d", i)
			assert.Equalf(test.ExpectedType, constant.Type, "Type: Test case %d", i)
			assert.Equalf(test.ExpectedValue, constant.Value, "Value: Test case %d", i)
		}
	}
}

func Add() *common.Const {
	var set common.Instructions
	set = append(set, common.NewOp(common.OpLoadLocal, 0)...)
	set = append(set, common.NewOp(common.OpLoadLocal, 1)...)
	set = append(set, common.NewOp(common.OpAddI64)...)
	set = append(set, common.NewOp(common.OpReturnValue)...)
	return common.NewConst(common.FnConst, common.NewCompiledFn("add", 2, set))
}

func Builtin() *common.Const {
	fn := vm.NewBuiltinFn("builtin", 0, common.InvalidConstType, func(c ...*common.Const) (*common.Const, error) {
		return nil, nil
	})
	return common.NewConst(common.FnConst, fn)
}

func TestCall(t *testing.T) {
	assert := assert.New(t)

	type CallTest struct {
		Fn            *common.Const
		Args          []*common.Const
		ExpectedError error
	}

	tests := []CallTest{
		{common.NewConst(common.I64Const, 0), []*common.Const{}, vm.ErrConstTypeInvalid},
		{Add(), []*common.Const{}, vm.ErrIncorrectNumberOfArgs},
		{Add(), []*common.Const{common.NewConst(common.I64Const, 5), common.NewConst(common.I64Const, 10)}, nil},
		{Builtin(), []*common.Const{}, nil},
	}

	var DefaultBuiltins = vm.NewBuiltins()
	machine := vm.NewVM(DefaultBuiltins)
	mod := common.NewModule("main", 0)
	machine.Init(common.NewArchive())
	executor := vm.NewExecutor(machine)
	assert.NoError(
		executor.Frames().Push(vm.NewFrame(common.NewCompiledFn("main", 0, common.Instructions{}), mod, 0)),
	)

	for i, test := range tests {
		stack := executor.Stack()
		for j, arg := range test.Args {
			assert.NoErrorf(stack.Push(arg), "Push Arg: Test case %d arg %d", i, j)
		}
		assert.NoErrorf(stack.Push(test.Fn), "Push Fn: Test case %d", i)
		err := executor.ExecuteCall(common.OpCall, []int{len(test.Args)})
		if test.ExpectedError != nil {
			assert.ErrorIsf(err, test.ExpectedError, "Test case %d", i)
		} else {
			assert.NoErrorf(err, "Test case %d", i)
			frame, err := executor.Frames().Top()
			assert.NoErrorf(err, "Frame: Test case %d", i)
			fn, err := vm.GetFn(test.Fn)
			assert.NoErrorf(err, "Get Fn: Test case %d", i)
			assert.Equalf(fn.Name(), frame.String(), "Test case %d", i)
		}
	}
}

func TestStack(t *testing.T) {
	assert := assert.New(t)

	type StackTest struct {
		Fn    func() error
		Error error
	}

	tests := []StackTest{
		{
			Fn: func() error {
				stack := vm.NewStack[int](2)
				return stack.Push(0)
			},
			Error: nil,
		},
		{
			Fn: func() error {
				stack := vm.NewStack[int](2)
				stack.Push(0)
				_, err := stack.Pop()
				return err
			},
			Error: nil,
		},
		{
			Fn: func() error {
				stack := vm.NewStack[int](2)
				stack.Push(0)
				_, err := stack.Top()
				return err
			},
			Error: nil,
		},
		{
			Fn: func() error {
				stack := vm.NewStack[int](2)
				_, err := stack.Pop()
				return err
			},
			Error: vm.ErrStackUnderflow,
		},
		{
			Fn: func() error {
				stack := vm.NewStack[int](2)
				_, err := stack.Top()
				return err
			},
			Error: vm.ErrStackUnderflow,
		},
		{
			Fn: func() error {
				stack := vm.NewStack[int](0)
				return stack.Push(0)
			},
			Error: vm.ErrStackOverflow,
		},
	}

	for i, test := range tests {
		err := test.Fn()
		if test.Error == nil {
			assert.NoError(err, "Test case %d", i)
		} else {
			assert.ErrorIs(err, test.Error, "Test case %d", i)
		}
	}
}

func SetupMachine(t *testing.T, mod *common.Module, fn *common.Const) *vm.VM {
	assert := assert.New(t)

	builtins := vm.DefaultBuiltins
	vm := vm.NewVM(builtins)
	fnidx, err := mod.Consts.Write(fn, 255)
	assert.NoError(err)

	archive := common.NewArchive()
	modidx, err := archive.Modules.Write(mod, 0)
	assert.NoError(err)
	archive.SetEntry(modidx, fnidx)
	assert.NoError(vm.Init(archive))

	return vm
}

func TestRun(t *testing.T) {
	assert := assert.New(t)

	mod := common.NewModule("main", common.NewVersion(0, 0, 1))
	add, err := mod.Consts.Write(Add(), 0)
	assert.NoError(err)

	var set common.Instructions
	set = append(set, common.NewOp(common.OpLoadI64, 5)...)
	set = append(set, common.NewOp(common.OpLoadI64, 7)...)
	set = append(set, common.NewOp(common.OpLoadConst, add)...)
	set = append(set, common.NewOp(common.OpCall, 2)...)
	set = append(set, common.NewOp(common.OpHalt)...)
	fn := common.NewConst(common.FnConst, common.NewCompiledFn("main", 0, set))

	vm := SetupMachine(t, mod, fn)
	vm.Run()

	assert.Equal(true, vm.Halted(), "VM Halted")
	assert.Equal(false, vm.Paniced(), "VM Paniced")
}

func TestTrap(t *testing.T) {
	assert := assert.New(t)

	mod := common.NewModule("main", common.NewVersion(0, 0, 1))

	var set common.Instructions
	set = append(set, common.NewOp(common.OpTrap)...)
	fn := common.NewConst(common.FnConst, common.NewCompiledFn("main", 0, set))

	vm := SetupMachine(t, mod, fn)
	vm.Run()

	assert.Equal(true, vm.Halted(), "VM Halted")
	assert.Equal(true, vm.Paniced(), "VM Paniced")
}

func TestPanic(t *testing.T) {
	assert := assert.New(t)

	mod := common.NewModule("main", common.NewVersion(0, 0, 1))

	var set common.Instructions
	set = append(set, common.NewOp(common.OpLoadBuiltin, vm.DefaultBuiltins.Get("panic"))...)
	set = append(set, common.NewOp(common.OpCall, 1)...)
	fn := common.NewConst(common.FnConst, common.NewCompiledFn("main", 0, set))

	vm := SetupMachine(t, mod, fn)
	vm.Run()

	assert.Equal(true, vm.Halted(), "VM Halted")
	assert.Equal(true, vm.Paniced(), "VM Paniced")
}
