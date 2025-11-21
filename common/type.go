package common

type TypeField struct {
	Name string
	Type int
}

type Type struct {
	Name   string
	Fields []TypeField
}

func (t *Type) MarshalBinary() ([]byte, error) {
	return []byte{42}, nil
}

func (t *Type) UnmarshalBinary(b []byte) error {
	return nil
}

func NewType() *Type {
	return &Type{}
}
