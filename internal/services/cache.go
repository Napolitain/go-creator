package services

import (
	"time"

	"gocreator/internal/interfaces"

	"github.com/patrickmn/go-cache"
)

// CacheService manages in-memory caching for API calls and computations
type CacheService struct {
	cache *cache.Cache
}

// NewCacheService creates a new cache service
// defaultExpiration: default expiration time for cache entries
// cleanupInterval: interval for cleaning up expired entries
func NewCacheService(defaultExpiration, cleanupInterval time.Duration) interfaces.CacheService {
	return &CacheService{
		cache: cache.New(defaultExpiration, cleanupInterval),
	}
}

// Get retrieves a value from the cache
func (c *CacheService) Get(key string) (interface{}, bool) {
	return c.cache.Get(key)
}

// Set stores a value in the cache with default expiration
func (c *CacheService) Set(key string, value interface{}) {
	c.cache.Set(key, value, cache.DefaultExpiration)
}

// Delete removes a value from the cache
func (c *CacheService) Delete(key string) {
	c.cache.Delete(key)
}

// Clear removes all items from the cache
func (c *CacheService) Clear() {
	c.cache.Flush()
}
