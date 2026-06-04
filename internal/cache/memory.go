package cache

import (
	"sync"

	"github.com/Bupyc/link-storage-service/internal/model"
)

type MemoryCache struct {
	mu   sync.RWMutex
	data map[string]model.Link
}

func NewMemoryCache() *MemoryCache {
	return &MemoryCache{
		data: make(map[string]model.Link),
	}
}

func (c *MemoryCache) Get(shortCode string) (model.Link, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	link, ok := c.data[shortCode]
	return link, ok
}

func (c *MemoryCache) Set(link model.Link) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data[link.ID] = link
}

func (c *MemoryCache) Delete(shortCode string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.data, shortCode)
}
