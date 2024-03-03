package list

import (
	"testing"
)

func TestGenericList(t *testing.T) {
	list := NewGeneric[string]()

	list.PushFront("Hi")
	list.PushBack("John")
	list.PushBack("Lee")

	t.Run("The PopBack result should be `Lee`", func(t *testing.T) {
		val, ok := list.PopBack()
		if !ok || val != "Lee" {
			t.Fail()
		}
	})

	t.Run("The PopFront result should be `Hi`", func(t *testing.T) {
		val, ok := list.PopFront()
		if !ok || val != "Hi" {
			t.Fail()
		}
	})

	list.PushBack("Lee")
	list.PushBack("Say")
	list.PushBack("Hello")

	list.RemoveFirst("Say")

	rtn := list.PopAll()
	answer := []string{"John", "Lee", "Hello"}
	if ok := compareArr(rtn, answer); !ok {
		t.Errorf("The result should be %v, got %v", answer, rtn)
	}

	t.Run("Should get zero value", func(t *testing.T) {
		val, ok := list.PopFront()
		if ok || val != "" {
			t.Fail()
		}

		val, ok = list.PopBack()
		if ok || val != "" {
			t.Fail()
		}
	})

}

func compareArr[T comparable](arr1, arr2 []T) bool {
	if len(arr1) != len(arr2) {
		return false
	}

	for i := range arr1 {
		if arr1[i] != arr2[i] {
			return false
		}
	}

	return true
}
