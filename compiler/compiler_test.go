package compiler_test

import (
	"bytes"
	"testing"

	"github.com/canpacis/flint/ast"
	"github.com/canpacis/flint/common"
	"github.com/canpacis/flint/compiler"
	"github.com/stretchr/testify/assert"
)

func TestCompileBlock(t *testing.T) {
	assert := assert.New(t)

	type CompileBlockTest struct {
		Resolver map[string]*ast.Program
		Builtins map[int]int
		Program  *ast.Program
		Block    []ast.OpStmt
		Expected []byte
	}

	mod := ast.Mod("main")
	tests := []CompileBlockTest{
		{
			make(map[string]*ast.Program),
			make(map[int]int),
			ast.NewProgram(mod, nil, nil, nil),
			[]ast.OpStmt{},
			[]byte{},
		},
		{
			map[string]*ast.Program{
				"io": ast.NewProgram(ast.Mod("io"), nil, nil, []*ast.ConstStmt{
					ast.Const(0, "u32", ast.Int(0)),      // Size 5, Index 0
					ast.Const(1, "bool", ast.Bool(true)), // Size 1, Index 5
				}),
				"std": ast.NewProgram(ast.Mod("std"), nil, nil, []*ast.ConstStmt{
					ast.Const(0, "i64", ast.Int(1)), // Size 9, Index 0
				}),
			},
			map[int]int{0: 0, 1: 1, 2: 5, 3: 42}, // builtin indicies
			ast.NewProgram(mod,
				[]*ast.LinkStmt{
					ast.Link(0, "io"),
					ast.Link(1, "std"),
				}, nil, []*ast.ConstStmt{
					ast.Const(0, "i64", ast.Int(0)),          // Size 9, Index 0
					ast.Const(1, "u32", ast.Int(0)),          // Size 5, Index 9
					ast.Const(2, "str", ast.String("Hello")), // Size 10, Index 14
				}),
			[]ast.OpStmt{
				ast.NewOp("load.const", 2),
				ast.NewOp("load.const", 1),
				ast.NewOp("load.const", 0),
				// ast.NewOp("load.modconst", 0, 1),
				// ast.NewOp("load.modconst", 0, 0),
				// ast.NewOp("load.modconst", 1, 0),
				ast.NewOp("load.i32", 256),
				ast.NewOp("load.i64", 256),
				ast.NewOp("load.u32", 256),
				ast.NewOp("load.u64", 256),
				ast.NewOp("load.builtin", 0),
				ast.NewOp("load.builtin", 1),
				ast.NewOp("load.builtin", 2),
				ast.NewOp("load.builtin", 3),
				ast.NewOp("alloc", 256),
				ast.NewOp("realloc", 256, 256),
				ast.NewOp("free", 256),
				ast.NewOp("new", 0),
				ast.NewOp("new.mod", 0, 0),
				ast.NewOp("new.builtin", 0),
			},
			[]byte{
				byte(common.OpLoadConst), 14, 0, 0, 0, // load.const 0, 2
				byte(common.OpLoadConst), 9, 0, 0, 0, // load.const 0, 1
				byte(common.OpLoadConst), 0, 0, 0, 0, // load.const 0, 0
				// byte(common.OpLoadModConst), 0, 0, 0, 0, 5, 0, 0, 0, // load.const 0, 1
				// byte(common.OpLoadModConst), 0, 0, 0, 0, 0, 0, 0, 0, // load.const 0, 0
				// byte(common.OpLoadModConst), 30, 0, 0, 0, 0, 0, 0, 0, // load.const 1, 0
				byte(common.OpLoadI32), 0, 1, 0, 0, // load.i32 256
				byte(common.OpLoadI64), 0, 1, 0, 0, 0, 0, 0, 0, // load.i64 256
				byte(common.OpLoadU32), 0, 1, 0, 0, // load.u32 256
				byte(common.OpLoadU64), 0, 1, 0, 0, 0, 0, 0, 0, // load.u64 256
				byte(common.OpLoadBuiltin), 0, 0, // load.builtin 0
				byte(common.OpLoadBuiltin), 1, 0, // load.builtin 1
				byte(common.OpLoadBuiltin), 5, 0, // load.builtin 2
				byte(common.OpLoadBuiltin), 42, 0, // load.builtin 3
				byte(common.OpAlloc), 0, 1, 0, 0, // alloc 256
				byte(common.OpRealloc), 0, 1, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, // alloc 256, 256
				byte(common.OpFree), 0, 1, 0, 0, 0, 0, 0, 0, // free 256
				byte(common.OpNew), 0, 0, 0, 0, // new 0
				byte(common.OpNewMod), 0, 0, 0, 0, 0, 0, 0, 0, // new.mod 0, 0
				byte(common.OpNewBuiltin), 0, 0, // new.builtin 0
			},
		},
		{
			make(map[string]*ast.Program),
			make(map[int]int),
			ast.NewProgram(mod, nil, nil, nil),
			[]ast.OpStmt{
				ast.NewOp("call", 256),
				ast.NewOp("return"),
				ast.NewOp("add.i64"),
				ast.NewOp("sub.i64"),
				ast.NewOp("mul.i64"),
				ast.NewOp("div.i64"),
				ast.NewOp("mod.i64"),
				ast.NewOp("and"),
				ast.NewOp("or"),
				ast.NewOp("mask.and"),
				ast.NewOp("mask.or"),
				ast.NewOp("mask.xor"),
				ast.NewOp("mask.not"),
				ast.NewOp("shift.right"),
				ast.NewOp("shift.left"),
				ast.NewOp("yield"),
				ast.NewOp("trap"),
				ast.NewOp("halt"),
			},
			[]byte{
				byte(common.OpCall), 0, 1,
				byte(common.OpReturn),
				byte(common.OpAddI64),
				byte(common.OpSubI64),
				byte(common.OpMulI64),
				byte(common.OpDivI64),
				byte(common.OpModI64),
				byte(common.OpAnd),
				byte(common.OpOr),
				byte(common.OpMaskAnd),
				byte(common.OpMaskOr),
				byte(common.OpMaskXor),
				byte(common.OpMaskNot),
				byte(common.OpShiftRight),
				byte(common.OpShiftLeft),
				byte(common.OpYield),
				byte(common.OpTrap),
				byte(common.OpHalt),
			},
		},
		{
			make(map[string]*ast.Program),
			make(map[int]int),
			ast.NewProgram(mod, nil, nil, nil),
			[]ast.OpStmt{
				ast.NewOp("jmp", 0),  // Size 3
				ast.NewOp("jmpz", 1), // Size 3
				ast.NewOp("noop"),    // Buffer +1
				ast.NewOp("noop"),    // Buffer +1
				ast.NewLabel(1, ast.NewOp("noop"), ast.NewOp("noop"), ast.NewOp("noop")), // Size 3
				ast.NewLabel(0, ast.NewOp("noop"), ast.NewOp("noop")),                    // Size 2
			},
			[]byte{
				byte(common.OpJmp), 8, 0,
				byte(common.OpJmpz), 2, 0,
				byte(common.OpNoop), byte(common.OpNoop), // Buffer
				byte(common.OpNoop), byte(common.OpNoop), byte(common.OpNoop), // Label 1
				byte(common.OpNoop), byte(common.OpNoop), // Label 0
			},
		},
	}

	version := common.NewVersion(0, 0, 1)

	for i, test := range tests {
		c := compiler.NewIRCompiler(version)
		c.Init(test.Program, test.Resolver, test.Builtins)

		assert.NoError(c.Compile(), "Compile: Test case %d", i)

		set, err := c.CompileBlock(test.Block)
		assert.NoError(err, "Test case %d", i)
		if set != nil {
			assert.Equalf(test.Expected, []byte(set), "Test case %d", i)
		}
	}
}

func TestCompile(t *testing.T) {
	assert := assert.New(t)

	program := ast.NewProgram(
		ast.Mod("main"),
		[]*ast.LinkStmt{
			ast.Link(0, "io"),
			ast.Link(1, "std"),
		},
		nil,
		[]*ast.ConstStmt{
			ast.Const(0, "i64", ast.Int(0)),          // Size 9, Index 0
			ast.Const(1, "u32", ast.Int(0)),          // Size 5, Index 9
			ast.Const(2, "str", ast.String("Hello")), // Size 10, Index 14
			ast.FnConst("main", compiler.POOL_WRITE_LIMIT, "fn", ast.Fn(
				ast.NewOp("load.const", 2),
				ast.NewOp("load.const", 1),
				ast.NewOp("load.const", 0),
				ast.NewOp("load.modconst", 0, 1),
				ast.NewOp("load.modconst", 0, 0),
				ast.NewOp("load.modconst", 1, 0),
				ast.NewOp("load.i32", 256),
				ast.NewOp("load.i64", 256),
				ast.NewOp("load.u32", 256),
				ast.NewOp("load.u64", 256),
				ast.NewOp("load.builtin", 0), // Panic
			)),
		},
	)

	c := compiler.NewIRCompiler(common.NewVersion(0, 0, 1))

	io := ast.NewProgram(ast.Mod("io"), nil, nil, []*ast.ConstStmt{
		ast.Const(0, "u32", ast.Int(0)),      // Size 5, Index 0
		ast.Const(1, "bool", ast.Bool(true)), // Size 1, Index 5
	})
	std := ast.NewProgram(ast.Mod("std"), nil, nil, []*ast.ConstStmt{
		ast.Const(0, "i64", ast.Int(1)), // Size 9, Index 0
	})

	c.Init(program, map[string]*ast.Program{"io": io, "std": std}, map[int]int{0: 0})

	assert.NoError(c.Compile())
	buf := new(bytes.Buffer)
	_, err := c.WriteTo(buf)
	assert.NoError(err)
	assert.NotEqual(0, buf.Len())
}
