# Http Mock Client

![GPL-3.0 License](https://img.shields.io/badge/License-GPL%20v3-blue.svg)

A flexible and configurable HTTP mock client for Go applications, designed for testing HTTP interactions and simulating API responses.

## Features

- JSON-based configuration for mock responses
- Response caching with TTL support
- Timeout simulation
- Custom headers and status codes
- Context support for request cancellation
- Thread-safe operations
- Support for different HTTP methods (GET, POST, PUT, DELETE, etc.)
- Configurable global and route-specific cache settings
- Error response simulation
- Generic response handling

## Installation

```bash
go get github.com/VirusHQ/go-http-mock-client
```

## Quick Start

1. Create a configuration file (config.json):

```json
{
    "globalCache": {
        "enabled": true,
        "ttl": 300,
        "maxEntries": 1000
    },
    "routes": {
        "get_users": {
            "path": "/api/users",
            "method": "GET",
            "cache": {
                "enabled": true,
                "ttl": 60
            },
            "expectedResponse": {
                "statusCode": 200,
                "headers": {
                    "Content-Type": "application/json"
                },
                "body": {
                    "users": [
                        {
                            "id": 1,
                            "name": "John Doe"
                        }
                    ]
                }
            }
        }
    }
}
```

2. Use the mock client in your code:

```go
package main

import (
    "log"
    httpmockclient "github.com/VirusHQ/go-http-mock-client"
)

func main() {
    // Create new mock client
    client, err := httpmockclient.NewClient("config.json")
    if err != nil {
        log.Fatal(err)
    }

    // Execute a mock request
    resp, err := client.Execute("get_users")
    if err != nil {
        log.Fatal(err)
    }

    log.Printf("Status: %d, Body: %s", resp.StatusCode, string(resp.Body))
}
```

## Configuration

### Global Cache Settings

```json
{
    "globalCache": {
        "enabled": true,
        "ttl": 300,
        "maxEntries": 1000
    }
}
```

### Route Configuration

Each route can be configured with (can use default_config.json):
- Path
- HTTP method
- Headers
- Query parameters
- Expected response
- Cache settings
- Timeout

Example:
```json
{
    "routes": {
        "create_user": {
            "path": "/api/users",
            "method": "POST",
            "headers": {
                "Content-Type": "application/json"
            },
            "body": {
                "name": "John Doe"
            },
            "expectedResponse": {
                "statusCode": 201,
                "body": {
                    "id": 1,
                    "name": "John Doe"
                }
            }
        }
    }
}
```

## Advanced Usage

### Using Context

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

resp, err := client.ExecuteWithContext(ctx, "get_users")
if err != nil {
    log.Fatal(err)
}
```

### Cache Invalidation

```go
// Invalidate specific route
client.InvalidateCacheForRoute("get_users")

// Invalidate all cache
client.InvalidateCache()
```

## Error Handling

The library returns clear error messages for:
- Configuration loading errors
- Route not found
- Context cancellation
- Timeout simulation
- Invalid response bodies

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## Testing

Run the tests with:

```bash
go test ./...
```

## License

This project is licensed under the **GNU General Public License v3.0**.  

You are free to use, modify, and distribute this software as long as you:
- Keep it **open-source** under the same GPL-3.0 license.
- Do **not** repackage or rebrand it as proprietary software.

Read the full license [here](LICENSE).

## Support

For support, please open an issue in the GitHub repository.

## Acknowledgments

- Thanks to the Go community for inspiration and best practices

