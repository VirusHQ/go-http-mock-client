package httpmockclient

import (
	"sync"
	"time"
)

// CacheEntry represents a cached response
type CacheEntry struct {
	Response  *Response
	CreatedAt time.Time
	ExpiresAt time.Time
}

// Cache represents the in-memory cache
type Cache struct {
	entries    map[string]CacheEntry
	mutex      sync.RWMutex
	maxEntries int
}

// get retrieves a cached response if available
func (c *Cache) get(key string) (*Response, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	entry, exists := c.entries[key]
	if !exists {
		return nil, false
	}

	if time.Now().After(entry.ExpiresAt) {
		return nil, false
	}

	return entry.Response, true
}

// set adds or updates a cache entry
func (c *Cache) set(key string, response *Response, ttl time.Duration) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.maxEntries > 0 && len(c.entries) >= c.maxEntries {
		var oldestKey string
		var oldestTime time.Time

		for k, v := range c.entries {
			if oldestKey == "" || v.CreatedAt.Before(oldestTime) {
				oldestKey = k
				oldestTime = v.CreatedAt
			}
		}

		delete(c.entries, oldestKey)
	}

	now := time.Now()
	c.entries[key] = CacheEntry{
		Response:  response,
		CreatedAt: now,
		ExpiresAt: now.Add(ttl),
	}
}

// cleanExpiredEntries removes expired entries from the cache
func (c *Cache) cleanExpiredEntries() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	now := time.Now()
	for key, entry := range c.entries {
		if now.After(entry.ExpiresAt) {
			delete(c.entries, key)
		}
	}
}
