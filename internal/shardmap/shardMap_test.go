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
	nm := shardmap.New(
		shardmap.WithShardNum[string, *TestValueItem](64),
	)
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
	nm32 := shardmap.NewNum[uint32, int]()
	nm16 := shardmap.NewNum[uint16, int]()
	nm8 := shardmap.NewNum[uint8, int]()
	nm64 := shardmap.NewNum[uint64, int]()
	nm32.Set(8, 10)

	nm8.Set(1, 20)
	nm16.Set(1, 20)
	nm32.Set(1, 20)
	nm64.Set(1, 20)

	val, ok := nm32.Get(1)
	if val != 20 || !ok {
		t.Fatalf("The val must be 10.")
		t.Fail()
	}
	nm32.Remove(1)
	val, ok = nm32.Get(1)
	if val != 0 || ok {
		t.Fatalf("The val must be 0")
		t.Fail()
	}

}

func TestShardMap(t *testing.T) {
	nm := shardmap.NewStringer[*TestKeyItem, *TestValueItem]()

	item := &TestKeyItem{Key: "100"}
	item2 := &TestKeyItem{Key: "200"}
	item3 := &TestKeyItem{Key: "300"}

	nm.Set(item, &TestValueItem{Value: "Woo"})
	nm.Set(item2, &TestValueItem{Value: "Woo"})
	nm.Set(item3, &TestValueItem{Value: "Woo"})
	nm.Set(&TestKeyItem{Key: "400"}, &TestValueItem{Value: "Woo"})

	val, ok := nm.Get(item)

	for item := range nm.Iter() {
		fmt.Println(item)
	}

	fmt.Println(val, ok, nm.Length())

}

const FNV_BASIS = uint32(2166136261)

func TestCustomShardingFunc(t *testing.T) {
	nm := shardmap.New(shardmap.WithCustomShardingFunc[string, int](func(key string) uint32 {
		const FNV_PRIME = uint32(16777619)
		nhash := FNV_BASIS
		for i := 0; i < len(key); i++ {
			nhash ^= uint32(key[i])
			nhash *= FNV_PRIME
		}
		return nhash
	}))

	nm.Set("hi", 10)
	if val, ok := nm.Get("hi"); val != 10 {
		fmt.Println(ok)
	}
}
