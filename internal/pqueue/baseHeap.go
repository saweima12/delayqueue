package pqueue

type LessFunc[T any] func(i, j T) bool

type baseHeap[T any] struct {
	lessFunc LessFunc[T]
	list     []T
}

func NewHeap[T any](compareFunc LessFunc[T], data []T) *baseHeap[T] {
	res := &baseHeap[T]{
		lessFunc: compareFunc,
		list:     data,
	}
	return res
}

// Len is the number of elements in the collection.
func (pr *baseHeap[T]) Len() int {
	return len(pr.list)
}

// Less reports whether the element with index i
// must sort before the element with index j.
//
// If both Less(i, j) and Less(j, i) are false,
// then the elements at index i and j are considered equal.
// Sort may place equal elements in any order in the final result,
// while Stable preserves the original input order of equal elements.
//
// Less must describe a transitive ordering:
//   - if both Less(i, j) and Less(j, k) are true, then Less(i, k) must be true as well.
//   - if both Less(i, j) and Less(j, k) are false, then Less(i, k) must be false as well.
//
// Note that floating-point comparison (the < operator on float32 or float64 values)
// is not a transitive ordering when not-a-number (NaN) values are involved.
// See Float64Slice.Less for a correct implementation for floating-point values.
func (bh *baseHeap[T]) Less(i int, j int) bool {
	return bh.lessFunc(bh.list[i], bh.list[j])
}

// Swap swaps the elements with indexes i and j.
func (bh *baseHeap[T]) Swap(i int, j int) {
	bh.list[i], bh.list[j] = bh.list[j], bh.list[i]
}

func (bh *baseHeap[T]) Push(x any) {
	bh.list = append(bh.list, x.(T))
}

func (bh *baseHeap[T]) Pop() any {
	n := len(bh.list)
	rtn := bh.list[n-1] // get tail element.
	bh.list = bh.list[0 : n-1]
	return rtn
}

func (bh *baseHeap[T]) Peek() any {
	return bh.list[0]
}
