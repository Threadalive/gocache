package project

import (
	"github.com/Threadalive/gocache/project/lru"
	"sync"
)

//用来执行并发访问的公共对象
type cache struct {
	mu sync.Mutex
	//引用缓存
	lru        *lru.Cache
	cacheBytes int64
}

func (c *cache) set(key string, value ByteView) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lru == nil {
		c.lru = lru.New(c.cacheBytes, nil)
	}
	c.lru.Set(key, value)
}

func (c *cache) get(key string) (value ByteView, ok bool) {
	c.mu.Lock()

	defer c.mu.Unlock()
	if c.lru == nil {
		return
	}
	if v, ok := c.lru.Get(key); ok {
		return v.(ByteView), ok
	}
	return
}
