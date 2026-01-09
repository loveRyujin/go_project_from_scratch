package geecache

import "container/list"

type Cache struct {
	maxBytes  int64
	usedBytes int64
	dl        *list.List
	cache     map[string]*list.Element
	OnEvicted func(key string, value Value)
}

type entry struct {
	key   string
	value Value
}

// Value use Len to count how many bytes it takes
type Value interface {
	Len() int
}

func New(maxBytes int64, onEvicted func(string, Value)) *Cache {
	return &Cache{
		maxBytes:  maxBytes,
		dl:        list.New(),
		cache:     make(map[string]*list.Element),
		OnEvicted: onEvicted,
	}
}

// Get look ups a key's value
func (c *Cache) Get(key string) (Value, bool) {
	ele, exist := c.cache[key]
	if !exist {
		return nil, false
	}

	c.dl.MoveToFront(ele)
	kv := ele.Value.(*entry)
	return kv.value, true
}

// RemoveOldest removes the oldest item
func (c *Cache) RemoveOldest() {
	ele := c.dl.Back()
	if ele != nil {
		c.dl.Remove(ele)
		kv := ele.Value.(*entry)
		delete(c.cache, kv.key)
		c.usedBytes -= int64(len(kv.key)) + int64(kv.value.Len())
		if c.OnEvicted != nil {
			c.OnEvicted(kv.key, kv.value)
		}
	}
}

// Add adds a value to the cache.
func (c *Cache) Add(key string, value Value) {
	if ele, ok := c.cache[key]; ok {
		c.dl.MoveToFront(ele)
		kv := ele.Value.(*entry)
		c.usedBytes += int64(value.Len()) - int64(kv.value.Len())
		kv.value = value
	} else {
		ele := c.dl.PushFront(&entry{key, value})
		c.cache[key] = ele
		c.usedBytes += int64(len(key)) + int64(value.Len())
	}
	for c.maxBytes != 0 && c.maxBytes < c.usedBytes {
		c.RemoveOldest()
	}
}
