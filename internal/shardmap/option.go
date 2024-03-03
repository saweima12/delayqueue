package shardmap

type Option[K comparable, V any] func(*Map[K, V])

// Set the custom sharding function.
func WithCustomShardingFunc[K comparable, V any](f ShardingFunc[K]) Option[K, V] {
	return func(m *Map[K, V]) {
		m.shardingFunc = f
	}
}

// Set the sharding number.
func WithShardNum[K comparable, V any](num uint32) Option[K, V] {
	return func(m *Map[K, V]) {
		m.shardNum = num
	}
}
