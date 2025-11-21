package vm

import (
	"errors"
	"fmt"

	"github.com/canpacis/flint/common"
)

var ErrConstTypeInvalid = errors.New("constant type is invalid")

func GetString(c *common.Const) (string, error) {
	v, ok := c.Value.(string)
	if !ok || c.Type != common.StrConst {
		return "", fmt.Errorf("%w: expected string found %s", ErrConstTypeInvalid, c.Type)
	}
	return v, nil
}

func GetData(c *common.Const) ([]byte, error) {
	v, ok := c.Value.([]byte)
	if !ok || c.Type != common.DataConst {
		return nil, fmt.Errorf("%w: expected data found %s", ErrConstTypeInvalid, c.Type)
	}
	return v, nil
}

func GetI64(c *common.Const) (int64, error) {
	n, ok := c.Value.(int64)
	if !ok || c.Type != common.I64Const {
		return 0, fmt.Errorf("%w: expected i64 found %s", ErrConstTypeInvalid, c.Type)
	}
	return n, nil
}

func GetU64(c *common.Const) (uint64, error) {
	n, ok := c.Value.(uint64)
	if !ok || c.Type != common.U64Const {
		return 0, fmt.Errorf("%w: expected u64 found %s", ErrConstTypeInvalid, c.Type)
	}
	return n, nil
}

func GetF64(c *common.Const) (float64, error) {
	n, ok := c.Value.(float64)
	if !ok || c.Type != common.F64Const {
		return 0, fmt.Errorf("%w: expected f64 found %s", ErrConstTypeInvalid, c.Type)
	}
	return n, nil
}

func GetBool(c *common.Const) (bool, error) {
	v, ok := c.Value.(bool)
	if !ok || (c.Type != common.TrueConst && c.Type != common.FalseConst) {
		return false, fmt.Errorf("%w: expected bool found %s", ErrConstTypeInvalid, c.Type)
	}
	return v, nil
}

func GetFn(c *common.Const) (common.Fn, error) {
	v, ok := c.Value.(common.Fn)
	if !ok || c.Type != common.FnConst {
		return nil, fmt.Errorf("%w: expected fn found %s", ErrConstTypeInvalid, c.Type)
	}
	return v, nil
}

func Binary[T any](stack *Stack[*common.Const], fn func(*common.Const) (T, error)) (T, T, error) {
	var zero T

	rconst, err := stack.Pop()
	if err != nil {
		return zero, zero, err
	}
	lconst, err := stack.Pop()
	if err != nil {
		return zero, zero, err
	}

	right, err := fn(rconst)
	if err != nil {
		return zero, zero, err
	}
	left, err := fn(lconst)
	if err != nil {
		return zero, zero, err
	}
	return left, right, nil
}
