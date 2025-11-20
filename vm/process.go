package vm

import (
	"io"
	"os"
)

type Processor interface {
	Process() *Process
}

type Process struct {
	ReadDescriptors  *Stack[io.Reader]
	WriteDescriptors *Stack[io.Writer]
}

func NewProcess() *Process {
	p := &Process{
		ReadDescriptors:  NewStack[io.Reader](255),
		WriteDescriptors: NewStack[io.Writer](255),
	}

	p.WriteDescriptors.Push(nil)
	p.WriteDescriptors.Push(os.Stdout)
	p.ReadDescriptors.Push(nil)
	p.ReadDescriptors.Push(os.Stdin)
	return p
}
