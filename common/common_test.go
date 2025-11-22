package common_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/canpacis/flint/common"
	"github.com/stretchr/testify/assert"
)

func TestOpCodes(t *testing.T) {
	assert := assert.New(t)

	type OpCodeTest struct {
		Code     common.OpCode
		Operands []int
		Expected []byte
	}

	makeOpTests := []OpCodeTest{
		{common.OpNoop, []int{}, []byte{byte(common.OpNoop)}},
		{common.OpLoadConst, []int{256}, []byte{byte(common.OpLoadConst), 0, 1, 0, 0}},
		{common.OpLoadModConst, []int{42, 256}, []byte{byte(common.OpLoadModConst), 42, 0, 0, 0, 0, 1, 0, 0}},
		{common.OpLoadLocal, []int{256}, []byte{byte(common.OpLoadLocal), 0, 1, 0, 0}},
		// {common.OpLoadGlobal, []int{256}, []byte{3, 0, 1, 0, 0}},
		{common.OpLoadBuiltin, []int{256}, []byte{byte(common.OpLoadBuiltin), 0, 1}},
		{common.OpLoadI32, []int{256}, []byte{byte(common.OpLoadI32), 0, 1, 0, 0}},
		{common.OpLoadI64, []int{256}, []byte{byte(common.OpLoadI64), 0, 1, 0, 0, 0, 0, 0, 0}},
		{common.OpLoadU32, []int{256}, []byte{byte(common.OpLoadU32), 0, 1, 0, 0}},
		{common.OpLoadU64, []int{256}, []byte{byte(common.OpLoadU64), 0, 1, 0, 0, 0, 0, 0, 0}},
		{common.OpSetLocal, []int{256}, []byte{byte(common.OpSetLocal), 0, 1, 0, 0}},
		// {common.OpSetGlobal, []int{256}, []byte{13, 0, 1, 0, 0}},
		{common.OpAlloc, []int{256}, []byte{byte(common.OpAlloc), 0, 1, 0, 0}},
		{common.OpRealloc, []int{256, 256}, []byte{byte(common.OpRealloc), 0, 1, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0}},
		{common.OpFree, []int{256}, []byte{byte(common.OpFree), 0, 1, 0, 0, 0, 0, 0, 0}},
		{common.OpNew, []int{256}, []byte{byte(common.OpNew), 0, 1, 0, 0}},
		{common.OpNewMod, []int{256, 256}, []byte{byte(common.OpNewMod), 0, 1, 0, 0, 0, 1, 0, 0}},
		{common.OpNewBuiltin, []int{256}, []byte{byte(common.OpNewBuiltin), 0, 1}},
		{common.OpPop, []int{}, []byte{byte(common.OpPop)}},
		{common.OpSwap, []int{}, []byte{byte(common.OpSwap)}},
		{common.OpCall, []int{256}, []byte{byte(common.OpCall), 0, 1}},
		{common.OpReturn, []int{}, []byte{byte(common.OpReturn)}},
		{common.OpAddU64, []int{}, []byte{byte(common.OpAddU64)}},
		{common.OpAddI64, []int{}, []byte{byte(common.OpAddI64)}},
		{common.OpSubU64, []int{}, []byte{byte(common.OpSubU64)}},
		{common.OpSubI64, []int{}, []byte{byte(common.OpSubI64)}},
		{common.OpMulU64, []int{}, []byte{byte(common.OpMulU64)}},
		{common.OpMulI64, []int{}, []byte{byte(common.OpMulI64)}},
		{common.OpDivU64, []int{}, []byte{byte(common.OpDivU64)}},
		{common.OpDivI64, []int{}, []byte{byte(common.OpDivI64)}},
		{common.OpDivF64, []int{}, []byte{byte(common.OpDivF64)}},
		{common.OpModI64, []int{}, []byte{byte(common.OpModI64)}},
		{common.OpModU64, []int{}, []byte{byte(common.OpModU64)}},
		{common.OpAnd, []int{}, []byte{byte(common.OpAnd)}},
		{common.OpOr, []int{}, []byte{byte(common.OpOr)}},
		{common.OpMaskAnd, []int{}, []byte{byte(common.OpMaskAnd)}},
		{common.OpMaskOr, []int{}, []byte{byte(common.OpMaskOr)}},
		{common.OpMaskXor, []int{}, []byte{byte(common.OpMaskXor)}},
		{common.OpMaskNot, []int{}, []byte{byte(common.OpMaskNot)}},
		{common.OpShiftRight, []int{}, []byte{byte(common.OpShiftRight)}},
		{common.OpShiftLeft, []int{}, []byte{byte(common.OpShiftLeft)}},
		{common.OpJmp, []int{0}, []byte{byte(common.OpJmp), 0, 0}},
		{common.OpJmpz, []int{0}, []byte{byte(common.OpJmpz), 0, 0}},
		{common.OpJmpt, []int{0}, []byte{byte(common.OpJmpt), 0, 0}},
		{common.OpJmpn, []int{0}, []byte{byte(common.OpJmpn), 0, 0}},
		{common.OpJmpp, []int{0}, []byte{byte(common.OpJmpp), 0, 0}},
		{common.OpTrap, []int{}, []byte{byte(common.OpTrap)}},
		{common.OpHalt, []int{}, []byte{byte(common.OpHalt)}},
	}

	for i, test := range makeOpTests {
		set := common.NewOp(test.Code, test.Operands...)
		assert.Equalf(test.Expected, []byte(set), "Test case %d", i)
	}

	type ReadOperandTest struct {
		Definition       common.OpDefinition
		Set              common.Instructions
		ExpectedOperands []int
		ExpectedOffset   int
	}

	for i, makeOpTest := range makeOpTests {
		code, _ := common.LookupOp(byte(makeOpTest.Code))
		test := ReadOperandTest{
			Definition:       code,
			Set:              makeOpTest.Expected[1:],
			ExpectedOperands: makeOpTest.Operands,
			ExpectedOffset:   len(makeOpTest.Expected) - 1,
		}

		operands, offset := common.ReadOperands(test.Definition, test.Set)
		assert.Equalf(test.ExpectedOperands, operands, "Operands: Test case %d", i)
		assert.Equalf(test.ExpectedOffset, offset, "Offset: Test case %d", i)
	}
}

func TestConstants(t *testing.T) {
	assert := assert.New(t)

	type ConstantEncodeTest struct {
		Input    *common.Const
		Expected []byte
	}

	encodeTests := []ConstantEncodeTest{
		{common.NewConst(common.StrConst, ""), []byte{byte(common.StrConst), 0, 0, 0, 0}},
		{common.NewConst(common.StrConst, "A"), []byte{byte(common.StrConst), 1, 0, 0, 0, 65}},
		{common.NewConst(common.TrueConst, 0), []byte{byte(common.TrueConst)}},
		{common.NewConst(common.FalseConst, 0), []byte{byte(common.FalseConst)}},
		{common.NewConst(common.U8Const, uint8(255)), []byte{byte(common.U8Const), 255}},
		{common.NewConst(common.U16Const, uint16(256)), []byte{byte(common.U16Const), 0, 1}},
		{common.NewConst(common.U32Const, uint32(256)), []byte{byte(common.U32Const), 0, 1, 0, 0}},
		{common.NewConst(common.U64Const, uint64(256)), []byte{byte(common.U64Const), 0, 1, 0, 0, 0, 0, 0, 0}},
		{common.NewConst(common.I8Const, int8(-128)), []byte{byte(common.I8Const), 128}},
		{common.NewConst(common.I16Const, int16(256)), []byte{byte(common.I16Const), 0, 1}},
		{common.NewConst(common.I32Const, int32(256)), []byte{byte(common.I32Const), 0, 1, 0, 0}},
		{common.NewConst(common.I64Const, int64(256)), []byte{byte(common.I64Const), 0, 1, 0, 0, 0, 0, 0, 0}},
		{common.NewConst(common.RefConst, uint32(256)), []byte{byte(common.RefConst), 0, 1, 0, 0}},
		{common.NewConst(common.FnConst, common.NewCompiledFn("A", 2, common.NewOp(common.OpNoop))), []byte{byte(common.FnConst), 1, 0, 0, 0, 65, 2, 0, 0, 0, 1, 0, 0, 0, 0}},
	}

	for i, test := range encodeTests {
		buf := new(bytes.Buffer)
		_, err := test.Input.WriteTo(buf)
		assert.NoError(err, "Test case %d", i)
		assert.Equalf(test.Expected, buf.Bytes(), "Test case %d", i)
	}

	type ConstantDecodeTest struct {
		Input         []byte
		ExpectedType  common.ConstType
		ExpectedValue any
	}

	decodeTests := []ConstantDecodeTest{
		{[]byte{byte(common.StrConst), 0, 0, 0, 0}, common.StrConst, ""},
		{[]byte{byte(common.StrConst), 1, 0, 0, 0, 65}, common.StrConst, "A"},
		{[]byte{byte(common.TrueConst)}, common.TrueConst, nil},
		{[]byte{byte(common.FalseConst)}, common.FalseConst, nil},
		{[]byte{byte(common.U8Const), 255}, common.U8Const, uint8(255)},
		{[]byte{byte(common.U16Const), 0, 1}, common.U16Const, uint16(256)},
		{[]byte{byte(common.U32Const), 0, 1, 0, 0}, common.U32Const, uint32(256)},
		{[]byte{byte(common.U64Const), 0, 1, 0, 0, 0, 0, 0, 0}, common.U64Const, uint64(256)},
		{[]byte{byte(common.I8Const), 128}, common.I8Const, int8(-128)},
		{[]byte{byte(common.I16Const), 0, 1}, common.I16Const, int16(256)},
		{[]byte{byte(common.I32Const), 0, 1, 0, 0}, common.I32Const, int32(256)},
		{[]byte{byte(common.I64Const), 0, 1, 0, 0, 0, 0, 0, 0}, common.I64Const, int64(256)},
		{[]byte{byte(common.RefConst), 0, 1, 0, 0, 0, 0, 0, 0}, common.RefConst, uint64(256)},
	}

	for i, test := range decodeTests {
		constant := new(common.Const)
		_, err := constant.ReadFrom(bytes.NewBuffer(test.Input))
		assert.NoError(err, "Test case %d", i)
		assert.Equalf(test.ExpectedType, constant.Type, "Test case %d", i)
		assert.Equalf(test.ExpectedValue, constant.Value, "Test case %d", i)
	}
}

func TestModules(t *testing.T) {
	assert := assert.New(t)

	type ModuleTest struct {
		Input func() *common.Module
	}

	tests := []ModuleTest{
		{func() *common.Module {
			mod := common.NewModule("main", common.NewVersion(0, 0, 1))
			return mod
		}},
		{func() *common.Module {
			mod := common.NewModule("io", common.NewVersion(1, 0, 0))
			return mod
		}},
		{func() *common.Module {
			io := common.NewModule("io", common.NewVersion(0, 0, 1))
			std := common.NewModule("std", common.NewVersion(0, 0, 1))

			mod := common.NewModule("main", common.NewVersion(0, 0, 1))
			mod.Links.Set(0, io)
			mod.Links.Set(1, std)
			return mod
		}},
		{func() *common.Module {
			mod := common.NewModule("main", common.NewVersion(0, 0, 1))

			mod.Consts.Set(0, common.NewConst(common.StrConst, "Hello, World!"))
			mod.Consts.Set(1, common.NewConst(common.I64Const, int64(0)))
			return mod
		}},
		{func() *common.Module {
			io := common.NewModule("io", common.NewVersion(0, 0, 1))
			std := common.NewModule("std", common.NewVersion(0, 0, 1))
			mod := common.NewModule("main", common.NewVersion(0, 0, 1))

			mod.Links.Set(0, io)
			mod.Links.Set(1, std)

			mod.Consts.Set(0, common.NewConst(common.StrConst, "Hello, World!"))
			mod.Consts.Set(1, common.NewConst(common.I64Const, int64(0)))
			return mod
		}},
	}

	for i, test := range tests {
		buf := new(bytes.Buffer)
		mod := test.Input()
		_, err := mod.WriteTo(buf)
		assert.NoError(err, "Test case %d", i)

		decoded := common.NewModule("", 0)
		_, err = decoded.ReadFrom(buf)
		assert.NoError(err, "Test case %d", i)

		assert.Equalf(mod.Name, decoded.Name, "Name: Test case %d", i)
		assert.Equalf(mod.Version, decoded.Version, "Version: Test case %d", i)
		assert.Equalf(mod.Links.Bytes(), decoded.Links.Bytes(), "Links: Test case %d", i)
		assert.Equalf(mod.Types.Bytes(), decoded.Types.Bytes(), "Types: Test case %d", i)
		assert.Equalf(mod.Consts.Bytes(), decoded.Consts.Bytes(), "Consts: Test case %d", i)
	}
}

func TestTypes(t *testing.T) {}

func TestVersion(t *testing.T) {
	assert := assert.New(t)

	type VersionTest struct {
		Version  common.Version
		Expected string
	}

	tests := []VersionTest{
		{common.NewVersion(0, 0, 0), "0.0.0"},
		{common.NewVersion(0, 0, 1), "0.0.1"},
		{common.NewVersion(0, 1, 0), "0.1.0"},
		{common.NewVersion(1, 0, 0), "1.0.0"},
		{common.NewVersion(1, 27, 8), "1.27.8"},
		{common.NewVersion(255, 255, 255), "255.255.255"},
	}

	for i, test := range tests {
		version := test.Version.String()
		if version != test.Expected {
			assert.Equalf(test.Expected, version, "Test case %d", i)
		}
	}
}

func TestPool(t *testing.T) {
	assert := assert.New(t)

	type PoolTest struct {
		Key             int
		Value           io.WriterTo
		Empty           io.ReaderFrom
		ExpectedPointer int
		ExpectedError   error
	}

	tests := []PoolTest{
		{0, common.NewConst(common.U8Const, uint8(255)), &common.Const{}, 0, nil},           // Size 2
		{1, common.NewConst(common.I32Const, int32(-255)), &common.Const{}, 2, nil},         // Size 5
		{2, common.NewConst(common.F64Const, float64(3.14159265)), &common.Const{}, 7, nil}, // Size 9
		{3, common.NewConst(common.StrConst, "Hello, World\n"), &common.Const{}, 16, nil},   // Size 18
		{4, common.NewLink("io"), common.NewLink(""), 34, nil},                              // Size 4
		{5, common.NewLink("std"), common.NewLink(""), 38, nil},                             // Size 5
		{0, nil, nil, 0, common.ErrPoolKeyExists},
	}

	pool := common.NewPool()

	for i, test := range tests {
		pointer, err := pool.Set(test.Key, test.Value)
		if test.ExpectedError != nil {
			assert.ErrorIsf(err, test.ExpectedError, "Test case %d", i)
		} else {
			assert.NoErrorf(err, "Test case %d", i)
			assert.Equalf(test.ExpectedPointer, pointer, "Test case %d", i)
			assert.Equalf(test.ExpectedPointer, pool.Lookup(test.Key), "Test case %d", i)
			assert.NoErrorf(pool.Get(pool.Lookup(test.Key), test.Empty), "Test case %d", i)
			assert.Equalf(test.Value, test.Empty, "Test case %d", i)
		}
	}
}
