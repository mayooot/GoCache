package gocache

import (
	"GoCache/gocache/lru"
	"sync"
)

// 实例化lru，封装get和add方法，并添加互斥锁mutex
type cache struct {
	mu         sync.Mutex
	lru        *lru.Cache
	cacheBytes int64
}

// 可以优化为单例初始化.....
func (c *cache) add(key string, value ByteView) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lru == nil {
		// 如果c.lru为nil，再创建lru实例。
		// 延迟初始化：一个对象的延迟初始化意味着该对象的创建将会延迟到第一次使用该对象时，主要用于提高性能，减少程序内存要求。
		c.lru = lru.New(c.cacheBytes, nil)
	}
	c.lru.Add(key, value)
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
