package pqueue

import (
	"container/heap"
)

type PriorityQueue[T any] struct {
	h *baseHeap[T]
}

func NewPriorityQueue[T any](f LessFunc[T], size int, data ...T) *PriorityQueue[T] {
	h := data

	if h == nil {
		h = make([]T, 0, size)
	}

	return &PriorityQueue[T]{
		h: NewHeap[T](f, h),
	}
}

func (pq *PriorityQueue[T]) Init() *PriorityQueue[T] {
	heap.Init(pq.h)
	return pq
}

func (pq *PriorityQueue[T]) Push(item T) {
	heap.Push(pq.h, item)
}

func (pq *PriorityQueue[T]) Pop() T {
	return heap.Pop(pq.h).(T)
}

func (pq *PriorityQueue[T]) Peek() T {
	return pq.h.Peek().(T)
}

func (pq *PriorityQueue[T]) Len() int {
	return pq.h.Len()
}
