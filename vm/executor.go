package vm

import (
	"errors"
	"fmt"
	"slices"

	"github.com/canpacis/flint/common"
)

var ErrUnknownOpCode = errors.New("unknown opcode")
var ErrUnsupportedOp = errors.New("unsupported op")
var ErrMissingConst = errors.New("missing constant index")
var ErrFailedToGetModule = errors.New("failed to get module")
var ErrFailedToLoadLink = errors.New("failed to load link")
var ErrDivideByZero = errors.New("divide by zero")
var ErrIncorrectNumberOfArgs = errors.New("function is called with incorrect number of arguments")

type Executor struct {
	vm     *VM
	stack  *Stack[*common.Const]
	frames *Stack[*Frame]
	links  map[int]*common.Module
	paused bool
	done   bool
}

func (e *Executor) Trap(reason string) {
	idx := e.vm.builtins.Get("panic")
	if idx < 0 {
		panic("Cannot trap, builtin panic is not provided")
	}
	if err := e.stack.Push(common.NewConst(common.StrConst, reason)); err != nil {
		panic(fmt.Sprintf("Trap failed: %s", err))
	}
	if err := e.ExecuteLoad(common.OpLoadBuiltin, []int{idx}); err != nil {
		panic(fmt.Sprintf("Trap failed: %s", err))
	}
	if err := e.ExecuteCall(common.OpCall, []int{1}); err != nil {
		panic(fmt.Sprintf("Trap failed: %s", err))
	}
}

func (e *Executor) Running() bool {
	return !e.vm.halted && !e.done && !e.paused
}

func (e *Executor) Run() {
	for e.Running() {
		frame, err := e.frames.Top()
		if err != nil {
			e.Trap(fmt.Errorf("failed to get frame: %w", err).Error())
			continue
		}
		code, operands, err := frame.Fetch()
		if err != nil {
			e.Trap(err.Error())
			continue
		}
		if err := e.Execute(code, operands); err != nil {
			e.Trap(fmt.Errorf("failed to execute op %s: %w", code, err).Error())
			continue
		}
	}
}

func (e *Executor) Execute(code common.OpCode, operands []int) error {
	switch code {
	case common.OpNoop:
		return nil
	case common.OpLoadConst, common.OpLoadModConst, common.OpLoadBuiltin,
		common.OpLoadLocal, common.OpLoadI32, common.OpLoadI64,
		common.OpLoadU32, common.OpLoadU64:
		return e.ExecuteLoad(code, operands)
	case common.OpAddI64, common.OpSubI64, common.OpMulI64, common.OpDivI64,
		common.OpModI64, common.OpMaskAnd, common.OpMaskOr, common.OpShiftRight,
		common.OpShiftLeft, common.OpAnd, common.OpOr:
		return e.ExecuteBinary(code, operands)
	case common.OpCall:
		return e.ExecuteCall(code, operands)
	case common.OpReturn, common.OpReturnValue:
		return e.ExecuteReturn(code)
	case common.OpPop, common.OpSwap, common.OpMaskNot:
		return e.ExecuteMutation(code, operands)
	case common.OpJmp, common.OpJmpz, common.OpJmpt, common.OpJmpn, common.OpJmpp:
		return e.ExecuteJump(code, operands)
	case common.OpYield:
		e.pause()
		return nil
	case common.OpTrap:
		constant, err := e.stack.Pop()
		if err == nil {
			str, err := GetString(constant)
			if err == nil {
				e.vm.panic(str)
			}
		}
		e.vm.halt()
		return nil
	case common.OpHalt:
		e.vm.halt()
		return nil
	default:
		return fmt.Errorf("%w: %s", ErrUnknownOpCode, code)
	}
}

func (e *Executor) ExecuteCall(code common.OpCode, operands []int) error {
	constant, err := e.stack.Pop()
	if err != nil {
		return fmt.Errorf("cannot get function constant: %w", err)
	}

	fn, err := GetFn(constant)
	if err != nil {
		return err
	}
	argsize := operands[0]
	if fn.Locals() != argsize {
		return fmt.Errorf("%w: expected %d got %d", ErrIncorrectNumberOfArgs, fn.Locals(), argsize)
	}

	base := e.stack.Len()
	current, err := e.frames.Top()
	if err != nil {
		return fmt.Errorf("cannot get current frame: %w", err)
	}

	frame := NewFrame(fn, current.mod, base-argsize)
	if err := e.frames.Push(frame); err != nil {
		return fmt.Errorf("cannot push new frame: %w", err)
	}

	builtin, ok := fn.(*BuiltinFn)
	if ok {
		args := make([]*common.Const, argsize)
		for i := range argsize {
			constant, err := e.stack.Get(frame.bp + i)
			if err != nil {
				return fmt.Errorf("cannot get argument constant: %w", err)
			}
			args[i] = constant
		}

		value, err := builtin.Fn(args...)
		if err != nil {
			return fmt.Errorf("builtin call failed: %w", err)
		}
		if value != nil {
			return e.stack.Push(value)
		}
	}
	return nil
}

func (e *Executor) ExecuteReturn(code common.OpCode) error {
	frame, err := e.frames.Pop()
	if err != nil {
		return err
	}

	if e.frames.Len() == 0 {
		// TODO: Executor is done, move the return value to the main thread
		e.finish()
		return nil
	}

	locals := frame.fn.Locals()
	var returns = new(common.Const)

	if code == common.OpReturnValue {
		var err error
		returns, err = e.stack.Pop()
		if err != nil {
			return err
		}
	}
	for range locals {
		_, err := e.stack.Pop()
		if err != nil {
			return err
		}
	}

	if returns != nil {
		return e.stack.Push(returns)
	}
	return nil
}

func (e *Executor) ExecuteLoad(code common.OpCode, operands []int) error {
	switch code {
	case common.OpLoadConst:
		mod, err := e.Context()
		if err != nil {
			return err
		}
		constant := new(common.Const)
		if err := mod.Consts.Get(operands[0], constant); err != nil {
			return err
		}
		return e.stack.Push(constant)
	case common.OpLoadModConst:
		mod, err := e.LoadLink(operands[0])
		if err != nil {
			return err
		}
		constant := new(common.Const)
		if err := mod.Consts.Get(operands[1], constant); err != nil {
			return err
		}
		return e.stack.Push(constant)
	case common.OpLoadBuiltin:
		if operands[0] >= e.vm.builtins.Len() {
			return fmt.Errorf("%w: no such builtin %d", ErrMissingConst, operands[0])
		}
		constant := e.vm.builtins.indicies[operands[0]]
		return e.stack.Push(constant)
	case common.OpLoadLocal:
		frame, err := e.frames.Top()
		if err != nil {
			return err
		}
		offset := operands[0]
		constant, err := e.stack.Get(frame.bp + offset)
		if err != nil {
			return err
		}
		return e.stack.Push(constant)
	case common.OpLoadI32:
		return e.stack.Push(common.NewConst(common.I32Const, int32(operands[0])))
	case common.OpLoadI64:
		return e.stack.Push(common.NewConst(common.I64Const, int64(operands[0])))
	case common.OpLoadU32:
		return e.stack.Push(common.NewConst(common.U32Const, uint32(operands[0])))
	case common.OpLoadU64:
		return e.stack.Push(common.NewConst(common.U64Const, uint64(operands[0])))
	default:
		return fmt.Errorf("%w: %s", ErrUnknownOpCode, code)
	}
}

func (e *Executor) ExecuteBinary(code common.OpCode, operands []int) error {
	i64 := []common.OpCode{
		common.OpAddI64, common.OpSubI64, common.OpMulI64, common.OpDivI64, common.OpModI64,
		common.OpMaskAnd, common.OpMaskOr, common.OpShiftRight,
		common.OpShiftLeft,
	}
	u64 := []common.OpCode{common.OpAddU64, common.OpSubU64, common.OpMulU64,
		common.OpDivU64, common.OpModU64,
	}
	f64 := []common.OpCode{common.OpDivF64}
	bl := []common.OpCode{common.OpAnd, common.OpOr}

	if slices.Contains(i64, code) {
		left, right, err := Binary(e.stack, GetI64)
		if err != nil {
			return err
		}

		switch code {
		case common.OpAddI64:
			return e.stack.Push(common.NewConst(common.I64Const, left+right))
		case common.OpSubI64:
			return e.stack.Push(common.NewConst(common.I64Const, left-right))
		case common.OpMulI64:
			return e.stack.Push(common.NewConst(common.I64Const, left*right))
		case common.OpDivI64:
			if right == 0 {
				return ErrDivideByZero
			}
			return e.stack.Push(common.NewConst(common.I64Const, left/right))
		case common.OpModI64:
			if right == 0 {
				return ErrDivideByZero
			}
			return e.stack.Push(common.NewConst(common.I64Const, left%right))
		case common.OpMaskAnd:
			return e.stack.Push(common.NewConst(common.I64Const, left&right))
		case common.OpMaskOr:
			return e.stack.Push(common.NewConst(common.I64Const, left|right))
		case common.OpShiftRight:
			return e.stack.Push(common.NewConst(common.I64Const, left>>right))
		case common.OpShiftLeft:
			return e.stack.Push(common.NewConst(common.I64Const, left<<right))
		default:
			return fmt.Errorf("%w: %s", ErrUnknownOpCode, code)
		}
	} else if slices.Contains(u64, code) {
		left, right, err := Binary(e.stack, GetU64)
		if err != nil {
			return err
		}

		switch code {
		case common.OpAddU64:
			return e.stack.Push(common.NewConst(common.U64Const, left+right))
		case common.OpSubU64:
			return e.stack.Push(common.NewConst(common.U64Const, left-right))
		case common.OpMulU64:
			return e.stack.Push(common.NewConst(common.U64Const, left*right))
		case common.OpDivU64:
			if right == 0 {
				return ErrDivideByZero
			}
			return e.stack.Push(common.NewConst(common.U64Const, left/right))
		case common.OpModU64:
			if right == 0 {
				return ErrDivideByZero
			}
			return e.stack.Push(common.NewConst(common.U64Const, left%right))
		default:
			return fmt.Errorf("%w: %s", ErrUnknownOpCode, code)
		}
	} else if slices.Contains(f64, code) {
		left, right, err := Binary(e.stack, GetF64)
		if err != nil {
			return err
		}

		switch code {
		case common.OpDivF64:
			if right == 0 {
				return ErrDivideByZero
			}
			return e.stack.Push(common.NewConst(common.F64Const, left/right))
		default:
			return fmt.Errorf("%w: %s", ErrUnknownOpCode, code)
		}
	} else if slices.Contains(bl, code) {
		left, right, err := Binary(e.stack, GetBool)
		if err != nil {
			return err
		}

		switch code {
		case common.OpAnd:
			if left && right {
				return e.stack.Push(common.NewConst(common.TrueConst, 0))
			}
			return e.stack.Push(common.NewConst(common.FalseConst, 0))
		case common.OpOr:
			if left || right {
				return e.stack.Push(common.NewConst(common.TrueConst, 0))
			}
			return e.stack.Push(common.NewConst(common.FalseConst, 0))
		default:
			return fmt.Errorf("%w: %s", ErrUnknownOpCode, code)
		}
	} else {
		return fmt.Errorf("%w: %s", ErrUnknownOpCode, code)
	}
}

func (e *Executor) ExecuteMutation(code common.OpCode, operands []int) error {
	switch code {
	case common.OpPop:
		_, err := e.stack.Pop()
		return err
	case common.OpSwap:
		right, err := e.stack.Pop()
		if err != nil {
			return err
		}
		left, err := e.stack.Pop()
		if err != nil {
			return err
		}
		if err := e.stack.Push(right); err != nil {
			return err
		}
		if err := e.stack.Push(left); err != nil {
			return err
		}
		return nil
	case common.OpMaskNot:
		constant, err := e.stack.Pop()
		if err != nil {
			return err
		}
		n, err := GetI64(constant)
		if err != nil {
			return err
		}
		return e.stack.Push(common.NewConst(common.I64Const, ^n))
	default:
		return fmt.Errorf("%w: %s", ErrUnsupportedOp, code)
	}
}

func (e *Executor) ExecuteJump(code common.OpCode, operands []int) error {
	return fmt.Errorf("%w: %s", ErrUnsupportedOp, code)
}

func (e *Executor) Context() (*common.Module, error) {
	frame, err := e.frames.Top()
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrFailedToGetModule, err)
	}
	return frame.mod, nil
}

func (e *Executor) LoadLink(idx int) (*common.Module, error) {
	cached, ok := e.links[idx]
	if ok {
		return cached, nil
	}
	mod := common.NewModule("", 0)
	if err := e.vm.archive.Modules.Get(idx, mod); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrFailedToLoadLink, err)
	}

	e.links[idx] = mod
	return mod, nil
}

func (e *Executor) Stack() *Stack[*common.Const] {
	return e.stack
}

func (e *Executor) Frames() *Stack[*Frame] {
	return e.frames
}

func (e *Executor) pause() {
	e.paused = true
}

func (e *Executor) finish() {
	e.done = true
}

const STACK_SIZE = 4096
const FRAME_SIZE = 4096

func NewExecutor(vm *VM) *Executor {
	return &Executor{
		vm:     vm,
		stack:  NewStack[*common.Const](STACK_SIZE),
		frames: NewStack[*Frame](FRAME_SIZE),
	}
}
