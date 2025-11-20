package common

import (
	"encoding/binary"
	"fmt"
	"math"
)

type ConstType byte

const (
	InvalidConstType = ConstType(iota)
	StrConst
	TrueConst
	FalseConst
	U8Const
	U16Const
	U32Const
	U64Const
	I8Const
	I16Const
	I32Const
	I64Const
	F32Const
	F64Const
	RefConst
	DataConst
	FnConst
)

var constmap = map[ConstType]string{
	StrConst:   "str",
	TrueConst:  "bool",
	FalseConst: "bool",
	U8Const:    "u8",
	U16Const:   "u16",
	U32Const:   "u32",
	U64Const:   "u64",
	I8Const:    "i8",
	I16Const:   "i16",
	I32Const:   "i32",
	I64Const:   "i64",
	F32Const:   "f32",
	F64Const:   "f64",
	RefConst:   "ref",
	DataConst:  "data",
	FnConst:    "fn",
}

func (t ConstType) String() string {
	return constmap[t]
}

func LookupConstType(name string) ConstType {
	for typ, n := range constmap {
		if name == n {
			return typ
		}
	}
	return InvalidConstType
}

type Fn interface {
	Name() string
	Locals() int
	Instructions() Instructions
}

type CompiledFn struct {
	name         string
	locals       int
	instructions Instructions
}

func (c *CompiledFn) Name() string {
	return c.name
}

func (c *CompiledFn) Locals() int {
	return c.locals
}

func (c *CompiledFn) Instructions() Instructions {
	return c.instructions
}

func (c *CompiledFn) Len() int {
	return 4 /* name length */ + len(c.name) + 4 /* local count */ + 4 /* length of instructions */ + len(c.instructions)
}

func (c *CompiledFn) MarshalBinary() ([]byte, error) {
	buf := make([]byte, c.Len())
	off := 0
	binary.LittleEndian.PutUint32(buf[off:], uint32(len(c.name)))
	off += 4
	off += copy(buf[off:], c.name)
	binary.LittleEndian.PutUint32(buf[off:], uint32(c.locals))
	off += 4
	binary.LittleEndian.PutUint32(buf[off:], uint32(len(c.instructions)))
	off += 4
	off += copy(buf[off:], c.instructions)
	return buf, nil
}

func (c *CompiledFn) UnmarshalBinary(b []byte) error {
	off := 0
	length := int(binary.LittleEndian.Uint32(b[off:]))
	off += 4
	buf := make([]byte, length)
	off += copy(buf, b[off:off+length])
	c.name = string(buf)
	c.locals = int(binary.LittleEndian.Uint32(b[off:]))
	off += 4
	length = int(binary.LittleEndian.Uint32(b[off:]))
	off += 4
	c.instructions = make(Instructions, length)
	off += copy(c.instructions, b[off:off+length])
	return nil
}

func NewCompiledFn(name string, locals int, set Instructions) *CompiledFn {
	return &CompiledFn{
		name:         name,
		locals:       locals,
		instructions: set,
	}
}

type Const struct {
	Type  ConstType
	Value any
}

func (c *Const) raw() []byte {
	switch c.Type {
	case StrConst:
		return []byte(c.Value.(string))
	case DataConst:
		return c.Value.([]byte)
	case FnConst:
		raw, _ := c.Value.(*CompiledFn).MarshalBinary()
		return raw
	default:
		panic(fmt.Sprintf("cannot read raw value of const type %d", c.Type))
	}
}

func (c *Const) MarshalBinary() ([]byte, error) {
	switch c.Type {
	case TrueConst, FalseConst:
		return []byte{byte(c.Type)}, nil
	case U8Const:
		return []byte{byte(c.Type), c.Value.(uint8)}, nil
	case U16Const:
		buf := make([]byte, 3)
		buf[0] = byte(c.Type)
		binary.LittleEndian.PutUint16(buf[1:], c.Value.(uint16))
		return buf, nil
	case U32Const:
		buf := make([]byte, 5)
		buf[0] = byte(c.Type)
		binary.LittleEndian.PutUint32(buf[1:], c.Value.(uint32))
		return buf, nil
	case U64Const:
		buf := make([]byte, 9)
		buf[0] = byte(c.Type)
		binary.LittleEndian.PutUint64(buf[1:], c.Value.(uint64))
		return buf, nil
	case I8Const:
		return []byte{byte(c.Type), uint8(c.Value.(int8))}, nil
	case I16Const:
		buf := make([]byte, 3)
		buf[0] = byte(c.Type)
		binary.LittleEndian.PutUint16(buf[1:], uint16(c.Value.(int16)))
		return buf, nil
	case I32Const:
		buf := make([]byte, 5)
		buf[0] = byte(c.Type)
		binary.LittleEndian.PutUint32(buf[1:], uint32(c.Value.(int32)))
		return buf, nil
	case I64Const:
		buf := make([]byte, 9)
		buf[0] = byte(c.Type)
		binary.LittleEndian.PutUint64(buf[1:], uint64(c.Value.(int64)))
		return buf, nil
	case F32Const:
		buf := make([]byte, 1 /*type size*/ +4 /* f32 size*/)
		buf[0] = byte(c.Type)
		binary.LittleEndian.PutUint32(buf[1:], math.Float32bits(c.Value.(float32)))
		return buf, nil
	case F64Const:
		buf := make([]byte, 1 /*type size*/ +8 /* f64 size*/)
		buf[0] = byte(c.Type)
		binary.LittleEndian.PutUint64(buf[1:], math.Float64bits(c.Value.(float64)))
		return buf, nil
	case RefConst:
		buf := make([]byte, 1 /*type size*/ +4 /* u32 size*/)
		buf[0] = byte(c.Type)
		binary.LittleEndian.PutUint32(buf[1:], c.Value.(uint32))
		return buf, nil
	case StrConst, DataConst, FnConst:
		data := c.raw()
		length := len(data)
		buf := make([]byte, length+1 /* type size */ +4 /* length size */)
		off := 0
		buf[off] = byte(c.Type)
		off++
		binary.LittleEndian.PutUint32(buf[off:], uint32(length))
		off += 4
		copy(buf[off:], data)
		return buf, nil
	default:
		return nil, fmt.Errorf("failed to encode const: invalid type %d", c.Type)
	}
}

func (c *Const) UnmarshalBinary(data []byte) error {
	off := 0
	c.Type = ConstType(data[off])
	off++

	switch c.Type {
	case TrueConst, FalseConst:
	case U8Const:
		c.Value = uint8(data[off])
	case U16Const:
		c.Value = binary.LittleEndian.Uint16(data[off:])
	case U32Const:
		c.Value = binary.LittleEndian.Uint32(data[off:])
	case U64Const:
		c.Value = binary.LittleEndian.Uint64(data[off:])
	case I8Const:
		c.Value = int8(uint8(data[off]))
	case I16Const:
		c.Value = int16(binary.LittleEndian.Uint16(data[off:]))
	case I32Const:
		c.Value = int32(binary.LittleEndian.Uint32(data[off:]))
	case I64Const:
		c.Value = int64(binary.LittleEndian.Uint64(data[off:]))
	case F32Const:
		c.Value = math.Float32frombits(binary.LittleEndian.Uint32(data[off:]))
	case F64Const:
		c.Value = math.Float64frombits(binary.LittleEndian.Uint64(data[off:]))
	case RefConst:
		c.Value = binary.LittleEndian.Uint32(data[off:])
	case StrConst, DataConst, FnConst:
		length := int(binary.LittleEndian.Uint32(data[off:]))
		off += 4
		switch c.Type {
		case StrConst:
			c.Value = string(data[off : off+length])
		case DataConst:
			c.Value = data[off : off+length]
		case FnConst:
			fn := &CompiledFn{}
			if err := fn.UnmarshalBinary(data[off : off+length]); err != nil {
				return err
			}
			c.Value = fn
		}
	default:
		return fmt.Errorf("failed to decode const: invalid type %d", c.Type)
	}
	return nil
}

func (c *Const) String() string {
	return fmt.Sprintf("<%s %v>", constmap[c.Type], c.Value)
}

func NewConst(typ ConstType, value any) *Const {
	return &Const{Type: typ, Value: value}
}
