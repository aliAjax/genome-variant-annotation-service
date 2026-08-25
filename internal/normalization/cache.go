package normalization

import "sync"

type MemoryCache struct {
	mu      sync.RWMutex
	values  map[string]Result
	maximum int
	keys    []string
}

func NewMemoryCache(maximum int) *MemoryCache {
	return &MemoryCache{maximum: maximum, keys: []string{}}
}
func (c *MemoryCache) Get(key string) (Result, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	v, ok := c.values[key]
	return v, ok
}
func (c *MemoryCache) Put(key string, value Result) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, ok := c.values[key]; !ok {
		c.keys = append(c.keys, key)
	}
	c.values[key] = value
	for c.maximum > 0 && len(c.keys) > c.maximum {
		old := c.keys[0]
		c.keys = c.keys[1:]
		delete(c.values, old)
	}
}
func (c *MemoryCache) Size() int { c.mu.RLock(); defer c.mu.RUnlock(); return len(c.values) }
