# Ares Contrib

Extended middleware and utilities for the Ares web framework.

This package contains additional middleware that extends Ares functionality but is kept separate from the core framework to maintain its lightweight nature.

## Installation

```bash
go get github.com/xushuhui/ares-contrib
```

## Available Middleware

### CORS

Cross-Origin Resource Sharing middleware.

**Features:**
- Configurable allowed origins, methods, and headers
- Support for credentials
- Preflight request handling
- Max age configuration

**Usage:**

```go
import (
    "github.com/xushuhui/ares"
    "github.com/xushuhui/ares-contrib/middleware"
)

func main() {
    app := ares.New()

    // Simple usage with default config
    app.Use(middleware.CORS(middleware.DefaultCORSOptions))

    // Custom configuration
    app.Use(middleware.CORS(middleware.CORSOptions{
        AllowedOrigins:   []string{"https://example.com"},
        AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE"},
        AllowedHeaders:   []string{"Authorization", "Content-Type"},
        AllowCredentials: true,
        MaxAge:           3600,
    }))

    app.Run(":8080")
}
```

### Request ID

Generates unique request IDs for tracking and logging.

**Features:**
- UUID v4 generation by default
- Custom ID generator support
- Reuses existing request ID from header
- Stores ID in context for access in handlers

**Usage:**

```go
import (
    "github.com/xushuhui/ares"
    "github.com/xushuhui/ares-contrib/middleware"
)

func main() {
    app := ares.New()

    // Simple usage with default config
    app.Use(middleware.RequestID())

    // Custom configuration
    app.Use(middleware.RequestIDWithConfig(middleware.RequestIDConfig{
        Generator: func() string {
            return "custom-id-" + uuid.New().String()
        },
        RequestIDHeader: "X-Request-ID",
        ContextKey:      "requestID",
    }))

    app.Run(":8080")
}
```

### Body Limit

Limits the maximum request body size.

**Features:**
- Configurable size limit
- Uses http.MaxBytesReader for efficient limiting
- Prevents memory exhaustion attacks

**Usage:**

```go
import (
    "github.com/xushuhui/ares"
    "github.com/xushuhui/ares-contrib/middleware"
)

func main() {
    app := ares.New()

    // Limit request body to 10MB
    app.Use(middleware.BodyLimit(10 * 1024 * 1024))

    app.Run(":8080")
}
```

### JWT Authentication

JWT middleware for token-based authentication.

**Features:**
- Flexible key management with jwt.Keyfunc
- Configurable signing methods
- Custom claims support
- Detailed error classification (expired, invalid, malformed)
- Context-based claims storage

**Usage:**

```go
import (
    "github.com/golang-jwt/jwt/v5"
    "github.com/xushuhui/ares"
    "github.com/xushuhui/ares-contrib/middleware"
)

func main() {
    app := ares.New()

    // Simple usage with signing key only
    app.Use(middleware.JWT([]byte("your-secret-key")))

    // With custom signing method
    app.Use(middleware.JWT(
        []byte("your-secret-key"),
        middleware.WithSigningMethod(jwt.SigningMethodHS256),
    ))

    // With custom claims
    app.Use(middleware.JWT(
        []byte("your-secret-key"),
        middleware.WithClaims(func() jwt.Claims {
            return &jwt.RegisteredClaims{}
        }),
        middleware.WithContextKey("user"),
    ))

    // Access claims in handler
    app.GET("/protected", func(ctx *ares.Context) error {
        claims, ok := middleware.GetClaims(ctx.Request.Context())
        if !ok {
            return ctx.JSON(401, map[string]string{"error": "no claims"})
        }
        return ctx.JSON(200, claims)
    })

    app.Run(":8080")
}
```

**Options:**
- `WithSigningMethod(method)` - Set JWT signing method (default: HS256)
- `WithClaims(func)` - Use custom claims struct
- `WithContextKey(key)` - Set custom context key for storing claims (default: "user")

### Rate Limiter

Rate limiting middleware to prevent API abuse.

**Features:**
- Per-IP rate limiting
- Configurable rate and burst
- Custom key extraction
- Automatic cleanup of old limiters

**Usage:**

```go
import (
    "github.com/xushuhui/ares"
    "github.com/xushuhui/ares-contrib/middleware"
)

func main() {
    app := ares.New()

    // Simple usage with default config (10 req/s, burst 20)
    app.Use(middleware.RateLimiter())

    // Custom configuration
    app.Use(middleware.RateLimiterWithConfig(middleware.RateLimiterConfig{
        Rate:  100,  // 100 requests per second
        Burst: 200,  // Allow burst of 200 requests
        KeyFunc: func(r *http.Request) string {
            // Custom key extraction (e.g., by user ID)
            return r.Header.Get("X-User-ID")
        },
    }))

    app.Run(":8080")
}
```

### Gzip Compression

Response compression middleware to reduce bandwidth usage.

**Features:**
- Configurable compression level
- Minimum response size threshold
- Exclude specific file extensions
- Exclude specific paths
- Writer pooling for performance

**Usage:**

```go
import (
    "github.com/xushuhui/ares"
    "github.com/xushuhui/ares-contrib/middleware"
)

func main() {
    app := ares.New()

    // Simple usage with default config
    app.Use(middleware.Gzip())

    // Custom configuration
    app.Use(middleware.GzipWithConfig(middleware.GzipConfig{
        Level:     5,     // Compression level (1-9)
        MinLength: 1024,  // Only compress responses > 1KB
        ExcludedExtensions: []string{".png", ".jpg", ".gif"},
        ExcludedPaths:      []string{"/api/stream"},
    }))

    app.Run(":8080")
}
```

**Default Excluded Extensions:**
- Images: `.png`, `.jpg`, `.jpeg`, `.gif`, `.webp`, `.svg`
- Archives: `.zip`, `.gz`, `.tar`, `.rar`, `.7z`
- Media: `.mp4`, `.avi`, `.mov`, `.mp3`, `.wav`
- Documents: `.pdf`

### Secure Headers

Security headers middleware to protect against common web vulnerabilities.

**Features:**
- XSS Protection
- Content Type Options (nosniff)
- X-Frame-Options (clickjacking protection)
- HSTS (HTTP Strict Transport Security)
- Content Security Policy
- Referrer Policy
- Permissions Policy

**Usage:**

```go
import (
    "github.com/xushuhui/ares"
    "github.com/xushuhui/ares-contrib/middleware/secure"
)

func main() {
    app := ares.New()

    // Simple usage with default config
    app.Use(secure.New())

    // Custom configuration
    app.Use(secure.New(
        secure.WithXFrameOptions("DENY"),
        secure.WithHSTSMaxAge(31536000),  // 1 year
        secure.WithContentSecurityPolicy("default-src 'self'; script-src 'self' 'unsafe-inline'"),
        secure.WithReferrerPolicy("strict-origin-when-cross-origin"),
        secure.WithPermissionsPolicy("geolocation=(self), microphone=()"),
    ))

    app.Run(":8080")
}
```

**Options:**
- `WithXSSProtection(value)` - Set X-XSS-Protection header (default: "1; mode=block")
- `WithContentTypeNosniff(value)` - Set X-Content-Type-Options header (default: "nosniff")
- `WithXFrameOptions(value)` - Set X-Frame-Options header (default: "SAMEORIGIN")
- `WithHSTSMaxAge(seconds)` - Set HSTS max-age in seconds (default: 0, disabled)
- `WithHSTSExcludeSubdomains(bool)` - Exclude subdomains from HSTS (default: false)
- `WithContentSecurityPolicy(policy)` - Set Content-Security-Policy header
- `WithCSPReportOnly(bool)` - Use CSP report-only mode (default: false)
- `WithReferrerPolicy(policy)` - Set Referrer-Policy header
- `WithPermissionsPolicy(policy)` - Set Permissions-Policy header

**Default Headers:**
- `X-XSS-Protection: 1; mode=block`
- `X-Content-Type-Options: nosniff`
- `X-Frame-Options: SAMEORIGIN`

## Dependencies

- `github.com/golang-jwt/jwt/v5` - JWT implementation
- `github.com/google/uuid` - UUID generation
- `golang.org/x/time/rate` - Rate limiting

## Future Plans

This contrib package will eventually be moved to a separate repository to allow independent versioning and development.

## License

MIT
