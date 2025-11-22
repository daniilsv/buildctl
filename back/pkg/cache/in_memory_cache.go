package cache

import (
	"sync"
	"time"
)

// InMemoryCache реализует Cache с хранением в памяти
type InMemoryCache[T any] struct {
	cache sync.Map
	ttl   time.Duration
}

var _ Cache[any] = (*InMemoryCache[any])(nil)

// NewInMemoryCache создает новый экземпляр кеша в памяти
func NewInMemoryCache[T any](ttl time.Duration) Cache[T] {
	newCache := &InMemoryCache[T]{
		ttl: ttl,
	}
	// Запускаем очистку устаревших записей
	go newCache.startCleanup()
	return newCache
}

func (c *InMemoryCache[T]) Get(key string) (*T, bool) {
	if value, ok := c.cache.Load(key); ok {
		return (value.(*T)), true
	}
	return nil, false
}

func (c *InMemoryCache[T]) Set(key string, value *T) {
	c.cache.Store(key, value)
}

func (c *InMemoryCache[T]) Delete(key string) {
	c.cache.Delete(key)
}

func (c *InMemoryCache[T]) startCleanup() {
	ticker := time.NewTicker(c.ttl / 2)
	for range ticker.C {
		c.cache.Range(func(key, value any) bool {
			c.cache.Delete(key)
			return true
		})
	}
}
