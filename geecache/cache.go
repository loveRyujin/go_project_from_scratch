package geecache

import (
	"sync"

	"github.com/loveRyujin/geecache/lru"
)

type cache struct {
	lru        *lru.Cache
	mu         sync.Mutex
	cacheBytes int64
}

func (c *cache) Add(key string, val ByteView) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.lru == nil {
		c.lru = lru.New(c.cacheBytes, nil)
	}
	c.lru.Add(key, val)
}

func (c *cache) Get(key string) (val ByteView, exist bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.lru == nil {
		return
	}

	v, ok := c.lru.Get(key)
	if !ok {
		return
	}

	return v.(ByteView), ok
}
