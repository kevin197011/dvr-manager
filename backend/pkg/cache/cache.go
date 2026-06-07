package cache

import (
	"log"

	"dvr-vod-system/internal/repository"
)

// Cache URL 缓存接口（局号 → DVR 真实 URL）
type Cache interface {
	Set(key, value string)
	Get(key string) (string, bool)
}

// sqliteCache 基于 SQLite 的持久化缓存，带 TTL
type sqliteCache struct {
	repo    repository.RecordingCacheRepository
	ttlDays int
}

// NewSQLiteCache 创建 SQLite 录像 URL 缓存；ttlDays 为条目保留天数
func NewSQLiteCache(repo repository.RecordingCacheRepository, ttlDays int) Cache {
	if ttlDays <= 0 {
		ttlDays = 30
	}
	return &sqliteCache{
		repo:    repo,
		ttlDays: ttlDays,
	}
}

// Set 设置缓存
func (c *sqliteCache) Set(key, value string) {
	if err := c.repo.Set(key, value, c.ttlDays); err != nil {
		log.Printf("[WARN] recording cache set failed - record_id: %s, error: %v", key, err)
	}
}

// Get 获取缓存（过期条目视为未命中）
func (c *sqliteCache) Get(key string) (string, bool) {
	return c.repo.Get(key)
}
