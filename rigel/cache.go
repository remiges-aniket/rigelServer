package rigel

import (
	"sync"
)

type InMemoryCache struct {
	data map[string]string
	mu   sync.RWMutex
}

func NewInMemoryCache() *InMemoryCache {
	return &InMemoryCache{
		data: make(map[string]string),
	}
}

func (c *InMemoryCache) Get(key string) (value string, found bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	value, found = c.data[key]
	return
}

func (c *InMemoryCache) Set(key string, value string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data[key] = value
}

func (c *InMemoryCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.data, key)
}
