package shardmap_test

import (
	"math/rand"
	"strconv"
	"sync"
	"testing"

	"github.com/saweima12/delaywheel/internal/shardmap"
)

type ShardMapInterface[K comparable, V any] interface {
	Set(key K, value interface{})
	Get(key K) (interface{}, bool)
}

func benchmarkShardMapStrSetGet(b *testing.B, shardMap ShardMapInterface[string, any]) {
	b.Run("StrSet", func(b *testing.B) {

		for i := 0; i < b.N; i++ {
			k := strconv.Itoa(i)
			shardMap.Set(k, i)
		}
	})

	b.Run("StrGet", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			shardMap.Get(strconv.Itoa(i))
		}
	})
}

func benchmarkShardMapNumSetGet(b *testing.B, shardMap ShardMapInterface[uint32, any]) {
	b.Run("Set", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			shardMap.Set(uint32(i), i)
		}
	})

	b.Run("Get", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			shardMap.Get(uint32(i))
		}
	})
}

func benchmarkSyncMapSetGet(b *testing.B, syncMap *sync.Map) {
	b.Run("Store", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			syncMap.Store(strconv.Itoa(i), i)
		}
	})

	b.Run("Load", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			syncMap.Load(strconv.Itoa(i))
		}
	})
}

func benchmarkShardMapMixedStr(b *testing.B, shardMap ShardMapInterface[string, any]) {
	keys := make([]string, b.N)
	for i := 0; i < b.N; i++ {
		keys[i] = strconv.Itoa(rand.Intn(b.N))
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			op := rand.Intn(10)
			key := keys[rand.Intn(len(keys))]
			if op < 3 {
				shardMap.Get(key)
			} else {
				shardMap.Set(key, rand.Intn(100))
			}
		}
	})
}

func benchmarkShardMapMixedNum(b *testing.B, shardMap ShardMapInterface[uint32, any]) {
	keys := make([]uint32, b.N)
	for i := 0; i < b.N; i++ {
		keys[i] = uint32(i)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			op := rand.Intn(10)
			key := rand.Intn(len(keys))
			if op < 3 {
				shardMap.Get(uint32(key))
			} else {
				shardMap.Set(uint32(key), rand.Intn(100))
			}
		}
	})
}

func benchmarkSyncMapMixed(b *testing.B, syncMap *sync.Map) {
	keys := make([]string, b.N)
	for i := 0; i < b.N; i++ {
		keys[i] = strconv.Itoa(rand.Intn(b.N))
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			op := rand.Intn(10)
			key := keys[rand.Intn(len(keys))]
			if op < 3 {
				syncMap.Load(key)
			} else {
				syncMap.Store(key, rand.Intn(100))
			}
		}
	})
}

func BenchmarkShardMapStr(b *testing.B) {
	shardMap := shardmap.New[any]()
	benchmarkShardMapStrSetGet(b, shardMap)
}

func BenchmarkShardMapNum(b *testing.B) {
	shardMap := shardmap.NewNum[uint32, any]()
	benchmarkShardMapNumSetGet(b, shardMap)
}

func BenchmarkSyncMap(b *testing.B) {
	var syncMap sync.Map
	benchmarkSyncMapSetGet(b, &syncMap)
}

func BenchmarkShardMapMixedStr(b *testing.B) {
	shardMap := shardmap.New[interface{}]()
	benchmarkShardMapMixedStr(b, shardMap)
}

func BenchmarkShardMapMixedNum(b *testing.B) {
	shardMap := shardmap.NewNum[uint32, any]()
	benchmarkShardMapMixedNum(b, shardMap)
}

func BenchmarkSyncMapMixed(b *testing.B) {
	var syncMap sync.Map
	benchmarkSyncMapMixed(b, &syncMap)
}
