package cache

import (
	"sync"
)

// Cache URL 缓存接口
type Cache interface {
	Set(key, value string)
	Get(key string) (string, bool)
}

// memoryCache 内存缓存实现
type memoryCache struct {
	data map[string]string
	mu   sync.RWMutex
}

// NewMemoryCache 创建新的内存缓存
func NewMemoryCache() Cache {
	return &memoryCache{
		data: make(map[string]string),
	}
}

// Set 设置缓存
func (c *memoryCache) Set(key, value string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data[key] = value
}

// Get 获取缓存
func (c *memoryCache) Get(key string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	value, exists := c.data[key]
	return value, exists
}
