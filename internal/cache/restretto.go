package cache

import (
	"errors"
	"time"

	"github.com/dgraph-io/ristretto/v2"
)

// RistrettoCache 是基于 Ristretto 实现的缓存结构体
type RistrettoCache[K ristretto.Key, V any] struct {
	cache *ristretto.Cache[K, V]
}

// NewRistrettoCache 创建一个新的 RistrettoCache 实例
func NewRistrettoCache[K ristretto.Key, V any](
	maxCost int64,
	numCounters int64,
	bufferItems int64,
) (*RistrettoCache[K, V], error) {
	cache, err := ristretto.NewCache(&ristretto.Config[K, V]{
		NumCounters: numCounters,
		MaxCost:     maxCost,
		BufferItems: bufferItems,
	})
	if err != nil {
		return nil, err
	}
	return &RistrettoCache[K, V]{cache: cache}, nil
}

// Set 将键值对存入缓存
func (r *RistrettoCache[K, V]) Set(key K, value V, cost int64) bool {
	return r.cache.Set(key, value, cost)
}

// SetWithTTL 将键值对存入缓存，并设置过期时间
func (r *RistrettoCache[K, V]) SetWithTTL(key K, value V, cost int64, ttl time.Duration) bool {
	return r.cache.SetWithTTL(key, value, cost, ttl)
}

// Get 从缓存中获取值，如果不存在则返回错误
func (r *RistrettoCache[K, V]) Get(key K) (V, error) {
	value, found := r.cache.Get(key)
	if !found {
		var zeroValue V
		return zeroValue, errors.New("key not found")
	}

	return value, nil
}

// Delete 从缓存中删除指定的键
func (r *RistrettoCache[K, V]) Delete(key K) {
	r.cache.Del(key)
}

// GetOrSet 尝试从缓存中获取值，如果不存在则调用提供的函数获取值并存入缓存
func (r *RistrettoCache[K, V]) GetOrSet(key K, cost int64, fn func() (V, error)) (V, error) {
	value, found := r.cache.Get(key)
	if found {
		return value, nil
	}

	value, err := fn()
	if err != nil {
		return value, err
	}

	r.cache.Set(key, value, cost)
	return value, nil
}
