package cache

import (
	"context"
	"sync"
	"time"
)

type item struct {
	data      []byte
	expiresAt time.Time
	forever   bool
}

// MemoryCache implementa a um [Cache] em memória. Deve ser usado apenas
// em desenvolvimento e testes.
type MemoryCache struct {
	mu    sync.RWMutex
	items map[string]item
}

func NewMemoryCache() *MemoryCache {
	c := &MemoryCache{
		items: make(map[string]item),
	}
	go c.gc()
	return c
}

func (c *MemoryCache) Get(ctx context.Context, key string) ([]byte, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	it, ok := c.items[key]
	if !ok {
		return nil, ErrCacheMiss
	}

	if !it.forever && time.Now().After(it.expiresAt) {
		return nil, ErrCacheMiss
	}

	return it.data, nil
}

func (c *MemoryCache) Put(ctx context.Context, key string, data []byte, ttl time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items[key] = item{
		data:      data,
		expiresAt: time.Now().Add(ttl),
		forever:   ttl == 0,
	}

	return nil
}

func (c *MemoryCache) Del(ctx context.Context, key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.items, key)
	return nil
}

func (c *MemoryCache) Remember(ctx context.Context, key string, ttl time.Duration, fn func() ([]byte, error)) ([]byte, error) {
	data, err := c.Get(ctx, key)
	if err == nil {
		return data, nil
	}

	data, err = fn()
	if err != nil {
		return nil, err
	}

	_ = c.Put(ctx, key, data, ttl)
	return data, nil
}

func (c *MemoryCache) gc() {
	ticker := time.NewTicker(1 * time.Minute)
	for range ticker.C {
		c.mu.Lock()
		for key, it := range c.items {
			if !it.forever && time.Now().After(it.expiresAt) {
				delete(c.items, key)
			}
		}
		c.mu.Unlock()
	}
}
