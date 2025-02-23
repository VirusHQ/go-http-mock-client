package httpmockclient

import (
	"testing"
	"time"
)

func TestCache_SetAndGet(t *testing.T) {
	cache := &Cache{
		entries: make(map[string]CacheEntry),
	}

	response := &Response{
		StatusCode: 200,
		Body:       []byte(`{"test":"data"}`),
	}

	// Test setting and getting a cache entry
	t.Run("Basic Set and Get", func(t *testing.T) {
		cache.set("test-key", response, 5*time.Second)

		got, exists := cache.get("test-key")
		if !exists {
			t.Error("Expected cache entry to exist")
		}
		if got.StatusCode != response.StatusCode {
			t.Errorf("Expected status code %d, got %d", response.StatusCode, got.StatusCode)
		}
	})

	// Test expired entry
	t.Run("Expired Entry", func(t *testing.T) {
		cache.set("expired-key", response, -1*time.Second)

		_, exists := cache.get("expired-key")
		if exists {
			t.Error("Expected cache entry to be expired")
		}
	})

	// Test max entries limit
	t.Run("Max Entries Limit", func(t *testing.T) {
		cache = &Cache{
			entries:    make(map[string]CacheEntry),
			maxEntries: 2,
		}

		cache.set("key1", response, 5*time.Second)
		time.Sleep(1 * time.Millisecond) // Ensure different CreatedAt times
		cache.set("key2", response, 5*time.Second)
		time.Sleep(1 * time.Millisecond)
		cache.set("key3", response, 5*time.Second)

		if len(cache.entries) > 2 {
			t.Error("Cache exceeded max entries limit")
		}

		// key1 should have been removed as it's the oldest
		_, exists := cache.get("key1")
		if exists {
			t.Error("Expected oldest entry to be removed")
		}
	})
}

func TestCache_CleanExpiredEntries(t *testing.T) {
	cache := &Cache{
		entries: make(map[string]CacheEntry),
	}

	response := &Response{StatusCode: 200}

	// Add mix of expired and valid entries
	cache.set("expired1", response, -2*time.Second)
	cache.set("valid1", response, 5*time.Second)
	cache.set("expired2", response, -1*time.Second)
	cache.set("valid2", response, 5*time.Second)

	cache.cleanExpiredEntries()
	afterCount := len(cache.entries)

	if afterCount != 2 {
		t.Errorf("Expected 2 entries after cleanup, got %d", afterCount)
	}

	// Check that valid entries still exist
	if _, exists := cache.get("valid1"); !exists {
		t.Error("Valid entry 'valid1' was incorrectly removed")
	}
	if _, exists := cache.get("valid2"); !exists {
		t.Error("Valid entry 'valid2' was incorrectly removed")
	}
}
