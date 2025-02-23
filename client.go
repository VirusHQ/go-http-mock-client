package httpmockclient

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

// Response wraps http.Response with additional functionality
type Response struct {
	StatusCode  int
	Headers     http.Header
	Body        []byte
	ParsedBody  interface{}
	CachedAt    *time.Time
	ContentType string
}

// Client represents our HTTP client wrapper
type Client struct {
	configPath string
	client     *http.Client
	config     *Config
	cache      *Cache
}

// NewClient creates a new instance of our HTTP client
func NewClient(configPath string) (*Client, error) {
	client := &Client{
		configPath: configPath,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		cache: &Cache{
			entries: make(map[string]CacheEntry),
		},
	}

	if err := client.loadConfig(); err != nil {
		return nil, err
	}

	// Set up cache cleanup routine
	go client.startCacheCleanup()

	return client, nil
}

// loadConfig reads and parses the JSON configuration file
func (c *Client) loadConfig() error {
	data, err := os.ReadFile(c.configPath)
	if err != nil {
		return fmt.Errorf("error reading config file: %w", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("error parsing JSON config: %w", err)
	}

	// Initialize global defaults if not present
	if config.GlobalDefaults == nil {
		config.GlobalDefaults = DefaultResponses
	}

	// Set default cache configuration if not present
	if config.GlobalCache.TTL == 0 {
		config.GlobalCache.TTL = 300 // 5 minutes default
	}

	c.config = &config
	c.cache.maxEntries = config.GlobalCache.MaxEntries
	return nil
}

// getCacheKey generates a unique cache key for a request
func (c *Client) getCacheKey(routeName string, route RouteConfig) string {
	key := fmt.Sprintf("%s_%s_%s", routeName, route.Method, route.Path)

	if len(route.QueryParams) > 0 {
		queryKey, _ := json.Marshal(route.QueryParams)
		key += string(queryKey)
	}

	if route.Body != nil && route.Method != "GET" {
		bodyKey, _ := json.Marshal(route.Body)
		key += string(bodyKey)
	}

	return key
}

// startCacheCleanup starts a background routine to clean expired cache entries
func (c *Client) startCacheCleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	for range ticker.C {
		c.cache.cleanExpiredEntries()
	}
}

// InvalidateCache removes all entries from the cache
func (c *Client) InvalidateCache() {
	c.cache.mutex.Lock()
	defer c.cache.mutex.Unlock()
	c.cache.entries = make(map[string]CacheEntry)
}

// InvalidateCacheForRoute removes cache entries for a specific route
func (c *Client) InvalidateCacheForRoute(routeName string) {
	c.cache.mutex.Lock()
	defer c.cache.mutex.Unlock()

	for key := range c.cache.entries {
		if strings.HasPrefix(key, routeName+"_") {
			delete(c.cache.entries, key)
		}
	}
}

// Execute performs a mock HTTP request based on the route name
func (c *Client) Execute(routeName string) (*Response, error) {
	return c.ExecuteWithContext(context.Background(), routeName)
}

// ExecuteWithContext performs a mock HTTP request based on the route configuration
func (c *Client) ExecuteWithContext(ctx context.Context, routeName string) (*Response, error) {
	route, exists := c.config.Routes[routeName]
	if !exists {
		return nil, fmt.Errorf("route %s not found in configuration", routeName)
	}

	// Check if caching is enabled for this route
	cacheConfig := route.Cache
	if !cacheConfig.Enabled {
		cacheConfig = c.config.GlobalCache
	}

	// Try to get cached response if caching is enabled
	if cacheConfig.Enabled && route.Method == "GET" {
		cacheKey := c.getCacheKey(routeName, route)
		if cachedResp, found := c.cache.get(cacheKey); found {
			return cachedResp, nil
		}
	}

	// Get the expected response from configuration
	var responseConfig ResponseConfig
	if route.ExpectedResp != nil {
		responseConfig = *route.ExpectedResp
	} else {
		// Try to find a matching default response
		statusCode := 200 // Default success
		if defaultResp, exists := route.DefaultResps[fmt.Sprintf("%d", statusCode)]; exists {
			responseConfig = defaultResp
		} else if defaultResp, exists := c.config.GlobalDefaults[fmt.Sprintf("%d", statusCode)]; exists {
			responseConfig = defaultResp
		} else {
			// Use success default response
			responseConfig = c.config.GlobalDefaults["2xx"]
		}
	}

	// Create response headers
	headers := make(http.Header)
	for key, value := range responseConfig.Headers {
		headers.Set(key, value)
	}

	// Marshal response body
	var bodyBytes []byte
	var err error
	if responseConfig.Body != nil {
		bodyBytes, err = json.Marshal(responseConfig.Body)
		if err != nil {
			return nil, fmt.Errorf("error marshaling response body: %w", err)
		}
	}

	response := &Response{
		StatusCode:  responseConfig.StatusCode,
		Headers:     headers,
		Body:        bodyBytes,
		ParsedBody:  responseConfig.Body,
		ContentType: headers.Get("Content-Type"),
	}

	// Cache the response if appropriate
	if cacheConfig.Enabled && route.Method == "GET" && response.StatusCode >= 200 && response.StatusCode < 300 {
		cacheKey := c.getCacheKey(routeName, route)
		ttl := time.Duration(cacheConfig.TTL) * time.Second
		if ttl == 0 {
			ttl = 5 * time.Minute // default TTL
		}
		now := time.Now()
		response.CachedAt = &now
		c.cache.set(cacheKey, response, ttl)
	}

	// Simulate timeout if specified
	if route.Timeout > 0 {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(time.Duration(route.Timeout) * time.Second):
			// Continue with response
		}
	}

	return response, nil
}
