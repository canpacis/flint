package vm

const HEAP_SIZE = 4096

type Heap struct {
	// data [HEAP_SIZE]byte
}

func (h *Heap) Alloc(size int) int {
	return 0
}

func (h *Heap) Free(ptr int) {

}

func NewHeap() *Heap {
	return &Heap{}
}
