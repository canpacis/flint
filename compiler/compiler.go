package compiler

import (
	"fmt"
	"hash/fnv"
	"io"

	"github.com/canpacis/flint/ast"
	"github.com/canpacis/flint/common"
)

type IRCompiler struct {
	version  common.Version
	archive  *common.Archive
	module   *common.Module
	program  *ast.Program
	resolver map[string]*ast.Program
	links    map[int]*common.Module
	builtins map[int]int
}

func (c *IRCompiler) getConstant(stmt *ast.ConstStmt) (*common.Const, error) {
	typ := common.LookupConstType(stmt.Type.Value)
	switch typ {
	case common.StrConst:
		return common.NewConst(typ, stmt.Literal.Value()), nil
	case common.TrueConst, common.FalseConst:
		return common.NewConst(typ, 0), nil
	case common.U8Const:
		return common.NewConst(typ, uint8(stmt.Literal.Value().(int))), nil
	case common.U16Const:
		return common.NewConst(typ, uint16(stmt.Literal.Value().(int))), nil
	case common.U32Const:
		return common.NewConst(typ, uint32(stmt.Literal.Value().(int))), nil
	case common.U64Const:
		return common.NewConst(typ, uint64(stmt.Literal.Value().(int))), nil
	case common.I8Const:
		return common.NewConst(typ, int8(stmt.Literal.Value().(int))), nil
	case common.I16Const:
		return common.NewConst(typ, int16(stmt.Literal.Value().(int))), nil
	case common.I32Const:
		return common.NewConst(typ, int32(stmt.Literal.Value().(int))), nil
	case common.I64Const:
		return common.NewConst(typ, int64(stmt.Literal.Value().(int))), nil
	case common.F32Const:
		return common.NewConst(typ, float32(stmt.Literal.Value().(float64))), nil
	case common.F64Const:
		return common.NewConst(typ, stmt.Literal.Value()), nil
	case common.DataConst:
		return nil, nil
	case common.FnConst:
		stmts := stmt.Literal.Value().([]ast.OpStmt)
		set, err := c.CompileBlock(stmts)
		if err != nil {
			return nil, err
		}
		name := c.module.Name
		if stmt.Name == nil {
			name += ".anonymous"
		} else {
			name += "." + stmt.Name.Value
		}
		return common.NewConst(typ, common.NewCompiledFn(name, 0, set)), nil
	default:
		return nil, fmt.Errorf("invalid const type %s", stmt.Type.Value)
	}
}

func (c *IRCompiler) Compile() error {
	for _, stmt := range c.program.Links {
		idx := stmt.Index.Int

		program, ok := c.resolver[stmt.Mod.String]
		if !ok {
			return fmt.Errorf("failed to write link: cannot resolve link module %s", stmt.Mod.String)
		}
		link := NewIRCompiler(c.version)
		link.Init(program, c.resolver, c.builtins)

		if err := link.Compile(); err != nil {
			return fmt.Errorf("failed to write link: %w", err)
		}

		if _, err := c.module.Links.Write(common.Link(stmt.Mod.String), idx); err != nil {
			return fmt.Errorf("failed to write link: %w", err)
		}

		hash := hash(stmt.Mod.String)

		// Archive may already have the link written
		if !c.archive.Modules.Has(hash) {
			if _, err := c.archive.Modules.Write(link.module, hash); err != nil {
				return fmt.Errorf("failed to write link: %w", err)
			}
		}

		// write to cache
		c.links[hash] = link.module
	}

	for _, stmt := range c.program.Types {
		idx := stmt.Index.Int
		// TODO: Create the actual type
		typ := common.NewType()
		if _, err := c.module.Types.Write(typ, idx); err != nil {
			return fmt.Errorf("failed to write type: %w", err)
		}
	}

	for _, stmt := range c.program.Consts {
		idx := stmt.Index.Int
		constant, err := c.getConstant(stmt)
		if err != nil {
			return fmt.Errorf("failed to write const: %w", err)
		}
		if _, err := c.module.Consts.Write(constant, idx); err != nil {
			return fmt.Errorf("failed to write const: %w", err)
		}
	}

	return nil
}

func (c *IRCompiler) WriteTo(w io.Writer) (int64, error) {
	encoded, err := c.archive.MarshalBinary()
	if err != nil {
		return 0, err
	}
	n, err := w.Write(encoded)
	if err != nil {
		return 0, err
	}
	return int64(n), nil
}

func readOp(stmt *ast.Op) (common.OpCode, []int, error) {
	code, err := common.FindOpCode(stmt.Name.Value)
	if err != nil {
		return 0, nil, err
	}
	operands := make([]int, len(stmt.Operands))
	for i, operand := range stmt.Operands {
		operands[i] = operand.Int
	}
	return code, operands, nil
}

type jump struct {
	code  common.OpCode
	label int
	idx   int
}

func (c *IRCompiler) CompileBlock(ops []ast.OpStmt) (common.Instructions, error) {
	var set common.Instructions

	blocks := map[int]int{}
	jumps := []jump{}

	for _, stmt := range ops {
		switch stmt := stmt.(type) {
		case *ast.Op:
			code, operands, err := readOp(stmt)
			if err != nil {
				return nil, err
			}

			switch code {
			case common.OpLoadConst:
				idx := operands[0]

				if !c.module.Consts.Has(idx) {
					return nil, fmt.Errorf("undefined const index %d", idx)
				}
				operands[0] = c.module.Consts.Get(idx)
			case common.OpLoadModConst:
				modidx := operands[0]
				idx := operands[1]

				link := new(common.Link)
				if !c.module.Links.Has(modidx) {
					return nil, fmt.Errorf("undefined mod index %d", modidx)
				}

				if err := c.module.Links.Read(link, c.module.Links.Get(modidx)); err != nil {
					return nil, err
				}

				hash := hash(string(*link))
				mod, ok := c.links[hash]
				if !ok {
					return nil, fmt.Errorf("found mod index %d but failed to resolve it", modidx)
				}

				if !mod.Consts.Has(idx) {
					return nil, fmt.Errorf("undefined const index %d in mod %d", idx, modidx)
				}
				operands[0] = c.archive.Modules.Get(hash)
				operands[1] = mod.Consts.Get(idx)
			case common.OpLoadBuiltin:
				idx := operands[0]

				pointer, ok := c.builtins[idx]
				if !ok {
					return nil, fmt.Errorf("undefined builtin index %d", idx)
				}
				operands[0] = pointer
			case common.OpJmp, common.OpJmpt, common.OpJmpz, common.OpJmpn, common.OpJmpp:
				jumps = append(jumps, jump{code, operands[0], len(set)})
			}

			set = append(set, common.NewOp(code, operands...)...)
		case *ast.Label:
			block, err := c.CompileBlock(stmt.Ops)
			if err != nil {
				return nil, err
			}
			blocks[stmt.Index.Int] = len(set)
			set = append(set, block...)
		default:
			return nil, fmt.Errorf("unknown op statement type %T", stmt)
		}
	}

	jumpSize := 3
	for _, jump := range jumps {
		resolved, ok := blocks[jump.label]

		if !ok {
			return nil, fmt.Errorf("undefined label index %d", jump.label)
		}

		// Calculate the actual jump by subtracting the jump index and instruction size
		amount := resolved - (jump.idx + jumpSize)
		op := common.NewOp(jump.code, amount)

		// Replace the actual op with the modified one
		for i, b := range op {
			set[jump.idx+i] = b
		}
	}

	return set, nil
}

func (c *IRCompiler) Init(program *ast.Program, resolver map[string]*ast.Program, builtins map[int]int) {
	c.program = program
	c.resolver = resolver
	c.builtins = builtins
	c.links = make(map[int]*common.Module)
	c.archive = common.NewArchive()
	c.module = common.NewModule(program.Module.Name.String, c.version)
}

func NewIRCompiler(version common.Version) *IRCompiler {
	return &IRCompiler{version: version}
}

func hash(str string) int {
	hash := fnv.New64()
	hash.Write([]byte(str))
	return int(hash.Sum64())
}
