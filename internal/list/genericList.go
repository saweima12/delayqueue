package list

import clist "container/list"

type GenericList[T comparable] struct {
	list *clist.List
}

func NewGeneric[T comparable]() *GenericList[T] {
	return &GenericList[T]{
		list: clist.New(),
	}
}

func (gl *GenericList[T]) PushFront(value T) *clist.Element {
	return gl.list.PushFront(value)
}

func (gl *GenericList[T]) PushBack(value T) *clist.Element {
	return gl.list.PushBack(value)
}

func (gl *GenericList[T]) PopAll() []T {

	result := make([]T, 0, gl.list.Len())

	for {
		node := gl.remove(gl.list.Front())
		if node == nil {
			break
		}

		data, ok := node.Value.(T)
		if !ok {
			continue
		}
		result = append(result, data)
	}
	return result
}

func (gl *GenericList[T]) PopFront() (T, bool) {
	var zero T
	node := gl.remove(gl.list.Front())
	if node == nil {
		return zero, false
	}

	result, ok := node.Value.(T)
	if !ok {
		return zero, false
	}
	return result, true
}

func (gl *GenericList[T]) PopBack() (T, bool) {
	var zero T
	node := gl.remove(gl.list.Back())
	if node == nil {
		return zero, false
	}

	result, ok := node.Value.(T)
	if !ok {
		return zero, false
	}
	return result, true
}

func (gl *GenericList[T]) RemoveFirst(t T) {
	node := gl.list.Front()
	for {
		if node == nil {
			return
		}

		val, ok := node.Value.(T)
		if !ok {
			continue
		}

		if val == t {
			gl.list.Remove(node)
		}
		node = node.Next()
	}
}

func (gl *GenericList[T]) remove(node *clist.Element) *clist.Element {
	if node == nil {
		return nil
	}
	gl.list.Remove(node)
	return node
}
