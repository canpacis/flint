package vm

import (
	"fmt"

	"github.com/canpacis/flint/common"
)

type VM struct {
	heap     *Heap
	process  *Process
	thread   *Executor
	builtins *Builtins
	archive  *common.Archive
	halted   bool
	paniced  bool
	panicmsg string
}

func (vm *VM) Run() {
	vm.thread.Run()
}

func (vm *VM) Halted() bool {
	return vm.halted
}

func (vm *VM) Paniced() bool {
	return vm.paniced
}

func (vm *VM) PanicMessage() string {
	return vm.panicmsg
}

func (vm *VM) Process() *Process {
	return vm.process
}

func (vm *VM) Init(archive *common.Archive, builtins *Builtins) error {
	vm.archive = archive
	vm.builtins = builtins
	main, err := archive.MainModule()
	if err != nil {
		return fmt.Errorf("failed to find main module in archive: %w", err)
	}
	mainfn, err := archive.MainFn()
	if err != nil {
		return fmt.Errorf("failed to find main function in module: %w", err)
	}
	fn, err := GetFn(mainfn)
	if err != nil {
		return err
	}
	frame := NewFrame(fn, main, 0)
	vm.thread = NewExecutor(vm)
	return vm.thread.frames.Push(frame)
}

func (vm *VM) halt() {
	vm.halted = true
}

func (vm *VM) panic(msg string) {
	vm.panicmsg = msg
	vm.paniced = true
}

func NewVM() *VM {
	return &VM{
		heap:    NewHeap(),
		process: NewProcess(),
	}
}
