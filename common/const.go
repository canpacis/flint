package common

import (
	"encoding/binary"
	"fmt"
	"io"
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

func (c *CompiledFn) WriteTo(w io.Writer) (n int64, err error) {
	if err := binary.Write(w, binary.LittleEndian, uint32(len(c.name))); err != nil {
		return n, err
	} else {
		n += 4
	}
	if m, err := w.Write([]byte(c.name)); err != nil {
		return n, err
	} else {
		n += int64(m)
	}

	if err := binary.Write(w, binary.LittleEndian, uint32(c.locals)); err != nil {
		return n, err
	} else {
		n += 4
	}
	if err := binary.Write(w, binary.LittleEndian, uint32(len(c.instructions))); err != nil {
		return n, err
	} else {
		n += 4
	}

	if m, err := w.Write(c.instructions); err != nil {
		return n, err
	} else {
		n += int64(m)
	}
	return
}

func (c *CompiledFn) ReadFrom(r io.Reader) (n int64, err error) {
	var length uint32
	if err := binary.Read(r, binary.LittleEndian, &length); err != nil {
		return n, err
	} else {
		n += 4
	}

	buf := make([]byte, length)
	if m, err := r.Read(buf); err != nil {
		return n, err
	} else {
		n += int64(m)
		c.name = string(buf)
	}

	var locals uint32
	if err := binary.Read(r, binary.LittleEndian, &locals); err != nil {
		return n, err
	} else {
		n += 4
		c.locals = int(locals)
	}

	if err := binary.Read(r, binary.LittleEndian, &length); err != nil {
		return n, err
	} else {
		n += 4
	}

	c.instructions = make(Instructions, length)
	if m, err := r.Read(c.instructions); err != nil {
		return n, err
	} else {
		n += int64(m)
	}
	return
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

func (c *Const) WriteTo(w io.Writer) (n int64, err error) {
	if _, err := w.Write([]byte{byte(c.Type)}); err != nil {
		return n, err
	} else {
		n++
	}

	switch c.Type {
	case TrueConst, FalseConst:
	case U8Const, U16Const, U32Const, U64Const, I8Const, I16Const, I32Const, I64Const, F32Const, F64Const, RefConst:
		size := int64(0)
		switch c.Type {
		case U8Const, I8Const:
			size = 1
		case U16Const, I16Const:
			size = 2
		case U32Const, I32Const, F32Const, RefConst:
			size = 4
		case U64Const, I64Const, F64Const:
			size = 8
		}

		if err := binary.Write(w, binary.LittleEndian, c.Value); err != nil {
			return n, err
		}
		n += size
	case StrConst, DataConst:
		var data []byte

		switch c.Type {
		case StrConst:
			data = []byte(c.Value.(string))
		case DataConst:
			data = c.Value.([]byte)
		}

		if err := binary.Write(w, binary.LittleEndian, uint32(len(data))); err != nil {
			return n, err
		} else {
			n += 4
		}
		if m, err := w.Write(data); err != nil {
			return n, err
		} else {
			n += int64(m)
		}
	case FnConst:
		data := c.Value.(*CompiledFn)
		if m, err := data.WriteTo(w); err != nil {
			return n, err
		} else {
			n += m
		}
	default:
		return n, fmt.Errorf("failed to encode const: invalid type %d", c.Type)
	}

	return
}

func (c *Const) ReadFrom(r io.Reader) (n int64, err error) {
	typ := make([]byte, 1)
	if _, err := r.Read(typ); err != nil {
		return n, err
	} else {
		c.Type = ConstType(typ[0])
	}

	switch c.Type {
	case TrueConst, FalseConst:
	case U8Const:
		var value uint8
		if err := binary.Read(r, binary.LittleEndian, &value); err != nil {
			return n, err
		} else {
			c.Value = value
		}
		n += 1
	case I8Const:
		var value int8
		if err := binary.Read(r, binary.LittleEndian, &value); err != nil {
			return n, err
		} else {
			c.Value = value
		}
		n += 1
	case U16Const:
		var value uint16
		if err := binary.Read(r, binary.LittleEndian, &value); err != nil {
			return n, err
		} else {
			c.Value = value
		}
		n += 2
	case I16Const:
		var value int16
		if err := binary.Read(r, binary.LittleEndian, &value); err != nil {
			return n, err
		} else {
			c.Value = value
		}
		n += 2
	case U32Const, RefConst:
		var value uint32
		if err := binary.Read(r, binary.LittleEndian, &value); err != nil {
			return n, err
		} else {
			c.Value = value
		}
		n += 4
	case I32Const:
		var value int32
		if err := binary.Read(r, binary.LittleEndian, &value); err != nil {
			return n, err
		} else {
			c.Value = value
		}
		n += 4
	case F32Const:
		var value float32
		if err := binary.Read(r, binary.LittleEndian, &value); err != nil {
			return n, err
		} else {
			c.Value = value
		}
		n += 4
	case U64Const:
		var value uint64
		if err := binary.Read(r, binary.LittleEndian, &value); err != nil {
			return n, err
		} else {
			c.Value = value
		}
		n += 8
	case I64Const:
		var value int64
		if err := binary.Read(r, binary.LittleEndian, &value); err != nil {
			return n, err
		} else {
			c.Value = value
		}
		n += 8
	case F64Const:
		var value float64
		if err := binary.Read(r, binary.LittleEndian, &value); err != nil {
			return n, err
		} else {
			c.Value = value
		}
		n += 8
	case StrConst, DataConst:
		var length uint32

		if err := binary.Read(r, binary.LittleEndian, &length); err != nil {
			return n, err
		} else {
			n += 4
		}

		buf := make([]byte, length)
		if m, err := r.Read(buf); err != nil {
			return n, err
		} else {
			n += int64(m)
		}

		switch c.Type {
		case StrConst:
			c.Value = string(buf)
		case DataConst:
			c.Value = buf
		}
	case FnConst:
		fn := &CompiledFn{}
		if m, err := fn.ReadFrom(r); err != nil {
			return n, err
		} else {
			n += m
		}
		c.Value = fn
	default:
		return n, fmt.Errorf("failed to decode const: invalid type %d", c.Type)
	}

	return
}

func (c *Const) String() string {
	return fmt.Sprintf("<%s %v>", constmap[c.Type], c.Value)
}

func NewConst(typ ConstType, value any) *Const {
	return &Const{Type: typ, Value: value}
}
