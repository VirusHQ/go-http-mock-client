package httpmockclient

// Config represents the main configuration structure
type Config struct {
	GlobalDefaults map[string]ResponseConfig `json:"globalDefaults"`
	GlobalCache    CacheConfig               `json:"globalCache"`
	Routes         map[string]RouteConfig    `json:"routes"`
}

// CacheConfig defines caching behavior
type CacheConfig struct {
	Enabled    bool  `json:"enabled"`
	TTL        int64 `json:"ttl"` // in seconds
	MaxEntries int   `json:"maxEntries"`
}

// RouteConfig represents a single route configuration
type RouteConfig struct {
	Path         string                    `json:"path"`
	Method       string                    `json:"method"`
	Headers      map[string]string         `json:"headers"`
	Body         interface{}               `json:"body"`
	QueryParams  map[string]string         `json:"queryParams"`
	Timeout      int                       `json:"timeout"`
	ExpectedResp *ResponseConfig           `json:"expectedResponse"`
	DefaultResps map[string]ResponseConfig `json:"defaultResponses"`
	Cache        CacheConfig               `json:"cache"`
}

// ResponseConfig defines the expected response
type ResponseConfig struct {
	StatusCode int               `json:"statusCode"`
	Headers    map[string]string `json:"headers"`
	Body       interface{}       `json:"body"`
}

// DefaultResponses provides default responses for common HTTP status codes
var DefaultResponses = map[string]ResponseConfig{
	"2xx": {
		StatusCode: 200,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: map[string]interface{}{
			"status": "success",
		},
	},
	"400": {
		StatusCode: 400,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: map[string]interface{}{
			"status":  "error",
			"message": "Bad request",
		},
	},
	"401": {
		StatusCode: 401,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: map[string]interface{}{
			"status":  "error",
			"message": "Unauthorized",
		},
	},
	"403": {
		StatusCode: 403,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: map[string]interface{}{
			"status":  "error",
			"message": "Forbidden",
		},
	},
	"404": {
		StatusCode: 404,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: map[string]interface{}{
			"status":  "error",
			"message": "Not found",
		},
	},
	"5xx": {
		StatusCode: 500,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: map[string]interface{}{
			"status":  "error",
			"message": "Internal server error",
		},
	},
}
