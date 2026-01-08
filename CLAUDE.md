# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Ares Contrib is a collection of middleware packages for the Ares web framework (github.com/xushuhui/ares). This repository extends Ares with commonly-needed middleware while keeping the core framework lightweight.

## Architecture

### Middleware Structure

Each middleware is in its own package under `middleware/`:
- `cors/` - Cross-Origin Resource Sharing
- `jwt/` - JWT authentication
- `requestid/` - Request ID generation and tracking
- `bodylimit/` - Request body size limiting
- `ratelimiter/` - Rate limiting per IP/key
- `gzip/` - Response compression
- `secure/` - Security headers

All middleware follow the standard Go HTTP middleware pattern: `func(http.Handler) http.Handler`

### Middleware Design Pattern

Each middleware package follows this structure:
- `New()` function that accepts functional options
- `Option` type for configuration: `type Option func(*options)`
- Private `options` struct holding configuration
- `With*()` functions returning `Option` for each configurable field

Example from `cors/cors.go`:
```go
func New(opts ...Option) func(http.Handler) http.Handler {
    o := &options{/* defaults */}
    for _, opt := range opts {
        opt(o)
    }
    return func(next http.Handler) http.Handler { /* implementation */ }
}
```

### JWT Middleware Context Storage

The JWT middleware stores claims in request context using a typed `contextKey` to avoid collisions:
- Default context key is `"user"`
- Use `GetClaims(ctx)` to retrieve claims with default key
- Use `GetClaimsWithKey(ctx, key)` for custom keys
- See `middleware/jwt/jwt.go:140-153`

### Rate Limiter Implementation

Uses `golang.org/x/time/rate` with per-key token buckets:
- Stores limiters in sync.Map for concurrent access
- Automatic cleanup of old limiters via background goroutine
- Default key extraction is by IP address
- Custom key functions supported (e.g., by user ID, API key)

## Development Commands

### Running Tests
```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run tests for specific middleware
go test ./middleware/cors/
go test ./middleware/jwt/

# Run specific test
go test -v -run TestJWT ./middleware/jwt/
```

### Building
```bash
# Build all packages
go build ./...

# Build example
cd examples/basic && go build
```

### Running Example
```bash
cd examples/basic
go run main.go
# Server starts on :8080
```

The example demonstrates all middleware in action with endpoints at:
- `/health` - Basic health check
- `/api/v1/*` - API v1 routes with auth middleware
- `/api/v2/*` - API v2 routes (versioning example)
- `/admin/*` - Admin routes with additional middleware

## Dependencies

- `github.com/golang-jwt/jwt/v5` - JWT implementation
- `github.com/google/uuid` - UUID generation for request IDs
- `golang.org/x/time/rate` - Token bucket rate limiting
- `github.com/xushuhui/ares` - Core Ares framework (imported by examples)

## Key Implementation Details

### CORS Origin Handling
When `allowedOrigins` is `["*"]`, the middleware sets `Access-Control-Allow-Origin: *`. For specific origins, it validates the request origin and echoes it back if allowed, adding `Vary: Origin` header. Credentials are only allowed with specific origins, not wildcards (per CORS spec).

### Gzip Compression
Uses writer pooling (`sync.Pool`) for performance. Automatically skips compression for:
- Responses smaller than `MinLength` (default 1KB)
- Already compressed formats (images, archives, media)
- Paths in `ExcludedPaths`
- Extensions in `ExcludedExtensions`

### Body Limit
Uses `http.MaxBytesReader` which is the standard library's efficient way to limit request body size. Prevents memory exhaustion attacks by rejecting oversized requests before reading the entire body.

### Secure Headers
Applies defense-in-depth security headers. HSTS is disabled by default (MaxAge=0) since it requires HTTPS. CSP can be set to report-only mode for testing before enforcement.
