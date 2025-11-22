package vm

import (
	"errors"
	"fmt"
)

var ErrOutOfMemory = errors.New("out out memory")
var ErrInvalidHandle = errors.New("invalid handle")

type HeapHandle uint32

type HeapBlock struct {
	handle HeapHandle
	offset uint32
	size   uint32
	free   bool
}

type Heap struct {
	data     []byte
	cap      uint32
	blocks   []*HeapBlock
	blockmap map[HeapHandle]*HeapBlock
	next     HeapHandle
	limit    float64
}

func (h *Heap) Alloc(size int) (HeapHandle, error) {
	if size <= 0 {
		return 0, fmt.Errorf("size must be greater than 0")
	}
	sz := uint32(size)

	for i, block := range h.blocks {
		if block.free && block.size >= sz {
			return h.splitAndAlloc(i, sz), nil
		}
	}

	return 0, ErrOutOfMemory
}

func (h *Heap) splitAndAlloc(idx int, size uint32) HeapHandle {
	block := h.blocks[idx]
	if block.size == size {
		// exact fit
		block.free = false
		block.handle = h.next
		h.blockmap[h.next] = block
		h.next++
		return block.handle
	}

	// split
	allocated := &HeapBlock{
		handle: h.next,
		offset: block.offset,
		size:   size,
		free:   false,
	}
	rest := &HeapBlock{
		handle: 0,
		offset: block.offset + size,
		size:   block.size - size,
		free:   true,
	}

	blocks := make([]*HeapBlock, 0)
	blocks = append(blocks, h.blocks[:idx]...)
	blocks = append(blocks, allocated)
	blocks = append(blocks, rest)
	blocks = append(blocks, h.blocks[idx+1:]...)
	h.blocks = blocks

	h.blockmap[h.next] = allocated
	h.next++
	return allocated.handle
}

func (h *Heap) Free(handle HeapHandle) error {
	block, ok := h.blockmap[handle]
	if !ok {
		return ErrInvalidHandle
	}
	if block.free {
		return fmt.Errorf("double free")
	}
	block.free = true
	delete(h.blockmap, handle)
	block.handle = 0
	// TODO: Coalesce adjacent blocks
	return nil
}

func NewHeap(cap int) *Heap {
	h := &Heap{
		data:     make([]byte, cap),
		cap:      uint32(cap),
		blocks:   make([]*HeapBlock, 0),
		blockmap: make(map[HeapHandle]*HeapBlock),
		next:     1,
		limit:    0.6,
	}

	h.blocks = append(h.blocks, &HeapBlock{handle: 0, offset: 0, size: uint32(cap), free: true})
	return h
}
