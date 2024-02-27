package shardmap

import (
	"encoding/binary"
	"fmt"
	"sync"
)

const SHARD_DEFAULT = 32

type Stringer interface {
	fmt.Stringer
	comparable
}

type KeyType[K any] interface {
	string | NumTypes
}

type KVShardBlock[K comparable, V any] struct {
	items map[K]V
	mu    sync.Mutex
}

func createMap[K comparable, V any](f ShardingFunc[K], opts ...Option[K, V]) *Map[K, V] {
	resp := &Map[K, V]{
		shardNum:     SHARD_DEFAULT,
		shardingFunc: f,
	}

	for _, opt := range opts {
		opt(resp)
	}

	// Initialize shards
	resp.shards = make([]*KVShardBlock[K, V], resp.shardNum)
	for i := range resp.shards {
		resp.shards[i] = &KVShardBlock[K, V]{items: make(map[K]V)}
	}

	return resp
}

// Create a ShardMap with string as the key.
func New[V any](opts ...Option[string, V]) *Map[string, V] {
	return createMap[string, V](strFnv32, opts...)
}

func NewNum[K NumTypes, V any](opts ...Option[K, V]) *Map[K, V] {
	return createMap[K, V](numFnv32[K], opts...)
}

// Create a ShardMap where the key can be any struct implementing the String() method.
func NewStringer[K Stringer, V any](opts ...Option[K, V]) *Map[K, V] {
	return createMap[K, V](stringerFnv32, opts...)
}

type ShardingFunc[K comparable] func(key K) uint32
type Map[K comparable, V any] struct {
	shardNum     uint32
	shardingFunc ShardingFunc[K]
	shards       []*KVShardBlock[K, V]

	length int
}

func (sm *Map[K, V]) Length() int {
	total := 0
	for i := range sm.shards {
		total += len(sm.shards[i].items)
	}
	return total
}

// Load returns the value stored the map for a key or nil, if no value is present.
func (sm *Map[K, V]) Get(key K) (value V, ok bool) {
	index := sm.shardingFunc(key) % sm.shardNum
	shard := sm.shards[index]

	shard.mu.Lock()
	defer shard.mu.Unlock()
	val, ok := shard.items[key]
	return val, ok
}

// Set the value for a key.
func (sm *Map[K, V]) Set(key K, value V) {
	index := sm.shardingFunc(key) % sm.shardNum
	shard := sm.shards[index]

	shard.mu.Lock()
	defer shard.mu.Unlock()
	shard.items[key] = value
	sm.length++
}

func (sm *Map[K, V]) Remove(key K) {
	index := sm.shardingFunc(key) % sm.shardNum
	shard := sm.shards[index]

	shard.mu.Lock()
	defer shard.mu.Unlock()
	delete(shard.items, key)
}

// recommend number.
const FNV_BASIS = uint32(2166136261)

// FNV-1a algorithm
func fnv32[T string | []byte](key T) uint32 {
	const FNV_PRIME = uint32(16777619)
	nhash := FNV_BASIS
	for i := 0; i < len(key); i++ {
		nhash ^= uint32(key[i])
		nhash *= FNV_PRIME
	}
	return nhash
}

// Support someone who implement fmt.Stringer
func stringerFnv32[K fmt.Stringer](key K) uint32 {
	return fnv32(key.String())
}

func strFnv32(key string) uint32 {
	return fnv32(key)
}

type NumTypes interface {
	int | int8 | int16 | int32 | int64 | uint | uint8 | uint16 | uint32 | uint64
}

func numFnv32[K NumTypes](key K) uint32 {
	return fnv32(numToBytes[K](key))
}

func numToBytes[T NumTypes](value T) []byte {
	switch any(value).(type) {
	case int8, uint8:
		return []byte{byte(value)}
	case int16, uint16:
		buf := make([]byte, 2)
		binary.BigEndian.PutUint16(buf, uint16(value))
		return buf
	case int32, uint32:
		buf := make([]byte, 4)
		binary.BigEndian.PutUint32(buf, uint32(value))
		return buf
	case int, uint, int64, uint64:
		buf := make([]byte, 8)
		binary.BigEndian.PutUint64(buf, uint64(value))
		return buf
	}
	return nil
}
