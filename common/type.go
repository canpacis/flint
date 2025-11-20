package common

type Type struct {
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
