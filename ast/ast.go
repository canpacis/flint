package ast

type Program struct {
	Module *ModStmt
	Links  []*LinkStmt
	Types  []*TypeStmt
	Consts []*ConstStmt
}

func NewProgram(mod *ModStmt, links []*LinkStmt, types []*TypeStmt, consts []*ConstStmt) *Program {
	return &Program{
		Module: mod,
		Links:  links,
		Types:  types,
		Consts: consts,
	}
}

type Location int

type Node interface {
	Location() Location
}

type Stmt interface {
	Node
	stmt()
}

type ModStmt struct {
	loc  Location
	Name *StringLiteral
}

func (s *ModStmt) Location() Location {
	return s.loc
}

func Mod(name string) *ModStmt {
	return &ModStmt{Name: String(name)}
}

type ConstStmt struct {
	loc     Location
	Name    *Identifier // for fn consts
	Index   *IntLiteral
	Type    *Identifier
	Literal Literal
}

func (s *ConstStmt) Location() Location {
	return s.loc
}

func Const(idx int, typ string, lit Literal) *ConstStmt {
	return &ConstStmt{
		Index:   Int(idx),
		Type:    Ident(typ),
		Literal: lit,
	}
}

func FnConst(name string, idx int, typ string, lit Literal) *ConstStmt {
	return &ConstStmt{
		Index:   Int(idx),
		Name:    Ident(name),
		Type:    Ident(typ),
		Literal: lit,
	}
}

type LinkStmt struct {
	loc   Location
	Index *IntLiteral
	Mod   *StringLiteral
}

func (s *LinkStmt) Location() Location {
	return s.loc
}

func Link(idx int, mod string) *LinkStmt {
	return &LinkStmt{
		Index: Int(idx),
		Mod:   String(mod),
	}
}

type TypeField struct {
	loc  Location
	Name *StringLiteral
	Src  *IntLiteral // used in mod imports
	Type *IntLiteral
}

func (s *TypeField) Location() Location {
	return s.loc
}

type TypeStmt struct {
	loc    Location
	Name   *Identifier
	Index  *IntLiteral
	Fields []TypeField
}

func (s *TypeStmt) Location() Location {
	return s.loc
}

type OpStmt interface {
	Stmt
	opstmt()
}

type Op struct {
	loc      Location
	Name     *Identifier
	Operands []*IntLiteral
}

func (s *Op) Location() Location {
	return s.loc
}

func NewOp(name string, operands ...int) *Op {
	ops := make([]*IntLiteral, len(operands))

	for i, operand := range operands {
		ops[i] = Int(operand)
	}

	return &Op{
		Name:     Ident(name),
		Operands: ops,
	}
}

type Label struct {
	loc   Location
	Index *IntLiteral
	Ops   []OpStmt
}

func (s *Label) Location() Location {
	return s.loc
}

func NewLabel(idx int, ops ...OpStmt) *Label {
	return &Label{
		Index: Int(idx),
		Ops:   ops,
	}
}

func (*Op) opstmt()    {}
func (*Label) opstmt() {}

func (*ModStmt) stmt()   {}
func (*ConstStmt) stmt() {}
func (*LinkStmt) stmt()  {}
func (*TypeStmt) stmt()  {}
func (*Op) stmt()        {}
func (*Label) stmt()     {}

type Identifier struct {
	Value string
}

func Ident(v string) *Identifier {
	return &Identifier{Value: v}
}

type Literal interface {
	Node
	Value() any
}

type StringLiteral struct {
	loc    Location
	String string
}

func (l *StringLiteral) Location() Location {
	return l.loc
}

func (l *StringLiteral) Value() any {
	return l.String
}

func String(v string) *StringLiteral {
	return &StringLiteral{String: v}
}

type BoolLiteral struct {
	loc  Location
	Bool bool
}

func (l *BoolLiteral) Location() Location {
	return l.loc
}

func (l *BoolLiteral) Value() any {
	return l.Bool
}

func Bool(v bool) *BoolLiteral {
	return &BoolLiteral{Bool: v}
}

type IntLiteral struct {
	loc Location
	Int int
}

func (l *IntLiteral) Location() Location {
	return l.loc
}

func (l *IntLiteral) Value() any {
	return l.Int
}

func Int(v int) *IntLiteral {
	return &IntLiteral{Int: v}
}

type FloatLiteral struct {
	loc   Location
	Float float64
}

func (l *FloatLiteral) Location() Location {
	return l.loc
}

func (l *FloatLiteral) Value() any {
	return l.Float
}

func Float(v float64) *FloatLiteral {
	return &FloatLiteral{Float: v}
}

type DataLiteral struct {
	loc  Location
	Data []Literal
}

func (l *DataLiteral) Location() Location {
	return l.loc
}

func (l *DataLiteral) Value() any {
	return l.Data
}

func Data(v ...Literal) *DataLiteral {
	return &DataLiteral{Data: v}
}

type FnLiteral struct {
	loc Location
	Ops []OpStmt
}

func (l *FnLiteral) Location() Location {
	return l.loc
}

func (l *FnLiteral) Value() any {
	return l.Ops
}

func Fn(ops ...OpStmt) *FnLiteral {
	return &FnLiteral{Ops: ops}
}
