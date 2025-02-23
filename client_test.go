package httpmockclient

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"
)

func createTestConfig(t *testing.T) string {
	config := Config{
		GlobalDefaults: DefaultResponses,
		GlobalCache: CacheConfig{
			Enabled: true,
			TTL:     300,
		},
		Routes: map[string]RouteConfig{
			"test-route": {
				Path:   "/test",
				Method: "GET",
				ExpectedResp: &ResponseConfig{
					StatusCode: 200,
					Headers: map[string]string{
						"Content-Type": "application/json",
					},
					Body: map[string]interface{}{
						"message": "test response",
					},
				},
			},
			"timeout-route": {
				Path:    "/timeout",
				Method:  "GET",
				Timeout: 1,
			},
		},
	}

	configFile, err := os.CreateTemp("", "test-config-*.json")
	if err != nil {
		t.Fatal(err)
	}

	encoder := json.NewEncoder(configFile)
	if err := encoder.Encode(config); err != nil {
		t.Fatal(err)
	}

	return configFile.Name()
}

func TestNewClient(t *testing.T) {
	configPath := createTestConfig(t)
	defer os.Remove(configPath)

	client, err := NewClient(configPath)
	if err != nil {
		t.Fatalf("Failed to create new client: %v", err)
	}

	if client.config == nil {
		t.Error("Expected config to be loaded")
	}

	if client.cache == nil {
		t.Error("Expected cache to be initialized")
	}
}

func TestClient_Execute(t *testing.T) {
	configPath := createTestConfig(t)
	defer os.Remove(configPath)

	client, err := NewClient(configPath)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("Successful Request", func(t *testing.T) {
		resp, err := client.Execute("test-route")
		if err != nil {
			t.Fatalf("Execute failed: %v", err)
		}

		if resp.StatusCode != 200 {
			t.Errorf("Expected status code 200, got %d", resp.StatusCode)
		}

		var body map[string]interface{}
		if err := json.Unmarshal(resp.Body, &body); err != nil {
			t.Fatal(err)
		}

		if msg, ok := body["message"].(string); !ok || msg != "test response" {
			t.Error("Unexpected response body")
		}
	})

	t.Run("Non-existent Route", func(t *testing.T) {
		_, err := client.Execute("non-existent")
		if err == nil {
			t.Error("Expected error for non-existent route")
		}
	})

	t.Run("Timeout", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
		defer cancel()

		_, err := client.ExecuteWithContext(ctx, "timeout-route")
		if err == nil {
			t.Error("Expected timeout error")
		}
	})
}

func TestClient_Cache(t *testing.T) {
	configPath := createTestConfig(t)
	defer os.Remove(configPath)

	client, err := NewClient(configPath)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("Cache Hit", func(t *testing.T) {
		// First request should cache
		resp1, err := client.Execute("test-route")
		if err != nil {
			t.Fatal(err)
		}

		// Second request should hit cache
		resp2, err := client.Execute("test-route")
		if err != nil {
			t.Fatal(err)
		}

		if resp2.CachedAt == nil {
			t.Error("Expected response to be cached")
		}

		if resp1.StatusCode != resp2.StatusCode {
			t.Error("Cached response doesn't match original")
		}
	})

	t.Run("Cache Invalidation", func(t *testing.T) {
		// Make initial request
		_, err := client.Execute("test-route")
		if err != nil {
			t.Fatal(err)
		}

		// Invalidate cache
		client.InvalidateCache()

		// Verify cache was cleared
		cacheSize := len(client.cache.entries)
		if cacheSize != 0 {
			t.Errorf("Expected cache to be empty, got %d entries", cacheSize)
		}
	})

	t.Run("Route Cache Invalidation", func(t *testing.T) {
		// Make initial request
		_, err := client.Execute("test-route")
		if err != nil {
			t.Fatal(err)
		}

		// Invalidate specific route
		client.InvalidateCacheForRoute("test-route")

		// Verify route was cleared from cache
		key := client.getCacheKey("test-route", client.config.Routes["test-route"])
		if _, exists := client.cache.entries[key]; exists {
			t.Error("Expected route to be cleared from cache")
		}
	})
}
