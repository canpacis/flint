package common

import (
	"encoding/binary"
	"fmt"
)

type OpCode byte

const (
	OpNoop = OpCode(iota)
	OpLoadConst
	OpLoadModConst
	OpLoadLocal
	// OpLoadGlobal
	OpLoadBuiltin
	OpLoadI32
	OpLoadI64
	OpLoadU32
	OpLoadU64
	OpSetLocal
	// OpSetGlobal
	OpAlloc
	OpRealloc
	OpFree
	OpNew
	OpNewMod
	OpNewBuiltin
	OpPop
	OpSwap
	OpCall
	OpReturn
	OpReturnValue

	// Arithmetic
	OpAddU64
	OpAddI64
	OpSubU64
	OpSubI64
	OpMulU64
	OpMulI64
	OpDivU64
	OpDivI64
	OpDivF64
	OpModU64
	OpModI64

	// Logic
	OpAnd
	OpOr
	OpMaskAnd
	OpMaskOr
	OpMaskXor
	OpMaskNot
	OpShiftRight
	OpShiftLeft

	// Control Flow
	OpJmp
	OpJmpz
	OpJmpt
	OpJmpn
	OpJmpp

	OpYield
	OpTrap
	OpHalt
)

func (c OpCode) String() string {
	def, ok := ops[c]
	if !ok {
		return ""
	}
	return def.Name
}

type OpDefinition struct {
	Name          string
	OperandWidths []int
}

func (def OpDefinition) Width() int {
	width := 1
	for _, w := range def.OperandWidths {
		width += w
	}
	return width
}

var ops = map[OpCode]OpDefinition{
	OpNoop:         {"noop", []int{}},
	OpLoadConst:    {"load.const", []int{4}},
	OpLoadModConst: {"load.modconst", []int{4, 4}},
	OpLoadLocal:    {"load.local", []int{4}},
	// OpLoadGlobal:   {"load.global", []int{4}},
	OpLoadBuiltin: {"load.builtin", []int{2}},
	OpLoadI32:     {"load.i32", []int{4}},
	OpLoadI64:     {"load.i64", []int{8}},
	OpLoadU32:     {"load.u32", []int{4}},
	OpLoadU64:     {"load.u64", []int{8}},
	OpSetLocal:    {"set.local", []int{4}},
	// OpSetGlobal:    {"set.global", []int{4}},
	OpAlloc:       {"alloc", []int{4}},
	OpRealloc:     {"realloc", []int{8, 4}},
	OpFree:        {"free", []int{8}},
	OpNew:         {"new", []int{4}},
	OpNewMod:      {"new.mod", []int{4, 4}},
	OpNewBuiltin:  {"new.builtin", []int{2}},
	OpPop:         {"pop", []int{}},
	OpSwap:        {"swap", []int{}},
	OpCall:        {"call", []int{2}},
	OpReturn:      {"return", []int{}},
	OpReturnValue: {"return.value", []int{}},
	OpAddU64:      {"add.u64", []int{}},
	OpAddI64:      {"add.i64", []int{}},
	OpSubU64:      {"sub.u64", []int{}},
	OpSubI64:      {"sub.i64", []int{}},
	OpMulU64:      {"mul.u64", []int{}},
	OpMulI64:      {"mul.i64", []int{}},
	OpDivU64:      {"div.u64", []int{}},
	OpDivI64:      {"div.i64", []int{}},
	OpDivF64:      {"div.f64", []int{}},
	OpModU64:      {"mod.u64", []int{}},
	OpModI64:      {"mod.i64", []int{}},
	OpAnd:         {"and", []int{}},
	OpOr:          {"or", []int{}},
	OpMaskAnd:     {"mask.and", []int{}},
	OpMaskOr:      {"mask.or", []int{}},
	OpMaskXor:     {"mask.xor", []int{}},
	OpMaskNot:     {"mask.not", []int{}},
	OpShiftRight:  {"shift.right", []int{}},
	OpShiftLeft:   {"shift.left", []int{}},
	OpJmp:         {"jmp", []int{2}},
	OpJmpz:        {"jmpz", []int{2}},
	OpJmpt:        {"jmpt", []int{2}},
	OpJmpn:        {"jmpn", []int{2}},
	OpJmpp:        {"jmpp", []int{2}},
	OpYield:       {"yield", []int{}},
	OpTrap:        {"trap", []int{}},
	OpHalt:        {"halt", []int{}},
}

func LookupOp(code byte) (OpDefinition, error) {
	def, ok := ops[OpCode(code)]
	if !ok {
		return ops[OpNoop], fmt.Errorf("undefined opcode %d", code)
	}
	return def, nil
}

func FindOpCode(name string) (OpCode, error) {
	for code, def := range ops {
		if def.Name == name {
			return code, nil
		}
	}
	return OpNoop, fmt.Errorf("undefined op %s", name)
}

func NewOp(code OpCode, operands ...int) Instructions {
	def, ok := ops[code]
	if !ok {
		panic(fmt.Sprintf("invalid op code %d", code))
	}
	if len(def.OperandWidths) != len(operands) {
		panic(
			fmt.Sprintf(
				"invalid number of operands for op %s, expected %d found %d",
				code, len(def.OperandWidths), len(operands),
			),
		)
	}

	length := 1
	for _, w := range def.OperandWidths {
		length += w
	}
	buf := make([]byte, length)
	buf[0] = byte(code)

	off := 1
	for i, width := range def.OperandWidths {
		switch width {
		case 1:
			buf[off] = byte(operands[i])
		case 2:
			binary.LittleEndian.PutUint16(buf[off:], uint16(operands[i]))
		case 4:
			binary.LittleEndian.PutUint32(buf[off:], uint32(operands[i]))
		case 8:
			binary.LittleEndian.PutUint64(buf[off:], uint64(operands[i]))
		default:
			panic(fmt.Sprintf("invalid operand width %d", width))
		}
		off += width
	}

	return buf
}

func ReadOperands(def OpDefinition, set Instructions) ([]int, int) {
	operands := make([]int, len(def.OperandWidths))
	off := 0
	for i, width := range def.OperandWidths {
		switch width {
		case 1:
			operands[i] = int(set[off])
		case 2:
			operands[i] = int(binary.LittleEndian.Uint16(set[off:]))
		case 4:
			operands[i] = int(binary.LittleEndian.Uint32(set[off:]))
		case 8:
			operands[i] = int(binary.LittleEndian.Uint64(set[off:]))
		default:
			panic(fmt.Sprintf("invalid operand width %d", width))
		}
		off += width
	}
	return operands, off
}

type Instructions []byte
