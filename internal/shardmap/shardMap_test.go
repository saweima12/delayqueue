package shardmap_test

import (
	"fmt"
	"testing"

	"github.com/saweima12/delaywheel/internal/shardmap"
)

type TestKeyItem struct {
	Key string
}

func (te *TestKeyItem) String() string {
	return te.Key
}

type TestValueItem struct {
	Value string
}

func TestSetAndGet(t *testing.T) {
	nm := shardmap.New[*TestValueItem]()
	nm.Set("Item1", &TestValueItem{Value: "Hello"})
	nm.Set("Item2", &TestValueItem{Value: "Hello2"})

	val, ok := nm.Get("Item1")
	if ok {
		t.Log(val, ok)
	}

	if !ok || val.Value != "Hello" {
		t.Error("The value should equals 'hello'")
	}
}

func TestStringer(t *testing.T) {
	nm := shardmap.NewStringer[*TestKeyItem, int]()

	k := &TestKeyItem{Key: "Item"}
	nm.Set(k, 1)

	if nm.Length() != 1 {
		t.Error("The length should be 1")
		return
	}
}

func TestNum(t *testing.T) {
	nm := shardmap.NewNum[uint32, int]()
	nm.Set(8, 10)

	val, ok := nm.Get(8)
	if val != 10 || !ok {
		t.Errorf("The val must be 10.")
		return
	}
}

func TestShardMap(t *testing.T) {
	nm := shardmap.NewStringer[*TestKeyItem, *TestValueItem]()

	item := &TestKeyItem{Key: "100"}

	nm.Set(item, &TestValueItem{Value: "Woo"})
	nm.Set(item, &TestValueItem{Value: "Woo"})
	nm.Set(item, &TestValueItem{Value: "Woo"})
	nm.Set(item, &TestValueItem{Value: "Woo"})
	nm.Set(item, &TestValueItem{Value: "Woo"})
	nm.Set(item, &TestValueItem{Value: "Wow"})
	nm.Set(&TestKeyItem{Key: "300"}, &TestValueItem{Value: "Woo"})

	val, ok := nm.Get(item)

	fmt.Println(val, ok, nm.Length())

}
