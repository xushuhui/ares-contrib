# Ares Contrib

[中文文档](README.zh.md)

Extended middleware collection for the [Ares](https://github.com/xushuhui/ares) web framework - a lightweight, high-performance Go web framework built on chi router.

## 📋 Table of Contents

- [Overview](#overview)
- [Installation](#installation)
- [Available Middleware](#available-middleware)
- [Quick Start](#quick-start)
- [Middleware Examples](#middleware-examples)
- [Best Practices](#best-practices)
- [Testing](#testing)
- [Benchmarks](#benchmarks)
- [Dependencies](#dependencies)
- [Contributing](#contributing)
- [License](#license)

## 🎯 Overview

Ares Contrib provides a collection of production-ready middleware that extends the Ares framework's functionality while keeping the core lightweight. Each middleware is:

- ✅ **Well-tested** with comprehensive test coverage (87%+ overall)
- ✅ **Production-ready** used in real-world applications
- ✅ **Performant** optimized for speed and memory efficiency
- ✅ **Flexible** with functional options pattern
- ✅ **Standard library compliant** following Go best practices

## 📦 Installation

```bash
go get github.com/xushuhui/ares-contrib
```

## 🚀 Available Middleware

### Middleware Overview

| Middleware | Coverage | Description | Status |
|-----------|----------|-------------|--------|
| [RequestID](#request-id) | 100% | Unique request tracking | ✅ Stable |
| [Secure](#secure-headers) | 100% | Security headers protection | ✅ Stable |
| [CORS](#cors) | 96.2% | Cross-origin resource sharing | ✅ Stable |
| [JWT](#jwt-authentication) | 85.7% | Token-based authentication | ✅ Stable |
| [GZIP](#gzip-compression) | 80.9% | Response compression | ✅ Stable |
| [BodyLimit](#body-limit) | 72.7% | Request body size limit | ✅ Stable |
| [RateLimiter](#rate-limiter) | 72.0% | Rate limiting per IP/key | ✅ Stable |

---

## 🔥 Quick Start

```go
package main

import (
    "github.com/xushuhui/ares"
    "github.com/xushuhui/ares-contrib/middleware/cors"
    "github.com/xushuhui/ares-contrib/middleware/gzip"
    "github.com/xushuhui/ares-contrib/middleware/requestid"
    "github.com/xushuhui/ares-contrib/middleware/secure"
    "github.com/xushuhui/ares-contrib/middleware/jwt"
)

func main() {
    app := ares.New()

    // Add middleware
    app.Use(requestid.New())
    app.Use(secure.New())
    app.Use(cors.New(
        cors.WithAllowedOrigins([]string{"https://example.com"}),
        cors.WithAllowCredentials(true),
    ))
    app.Use(gzip.New(gzip.WithLevel(5)))

    // Public routes
    app.POST("/login", loginHandler)

    // Protected routes with JWT
    api := app.Group("/api", jwt.New([]byte("your-secret-key")))
    api.GET("/users", getUsersHandler)
    api.GET("/profile", getProfileHandler)

    app.Run(":8080")
}
```

---

## 📚 Middleware Examples

### Request ID

Generates unique request IDs for distributed tracing and logging.

**Features:**
- UUID v4 generation by default
- Custom generator support
- Reuses existing request ID from header
- Context-based access

**Usage:**

```go
import "github.com/xushuhui/ares-contrib/middleware/requestid"

// Default configuration
app.Use(requestid.New())

// Custom configuration
app.Use(requestid.New(
    requestid.WithGenerator(func() string {
        return "req-" + uuid.New().String()
    }),
    requestid.WithHeader("X-Request-ID"),
    requestid.WithContextKey("request_id"),
))

// Access in handler
app.GET("/test", func(ctx *ares.Context) error {
    reqID := ctx.GetString("request_id")
    ctx.Logger().Info("processing request", "id", reqID)
    return ctx.JSON(200, map[string]string{"request_id": reqID})
})
```

**Response Headers:**
```
X-Request-ID: 550e8400-e29b-41d4-a716-446655440000
```

---

### Secure Headers

Protects against common web vulnerabilities with security headers.

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
import "github.com/xushuhui/ares-contrib/middleware/secure"

// Default configuration
app.Use(secure.New())

// Production-ready configuration
app.Use(secure.New(
    secure.WithXSSProtection("1; mode=block"),
    secure.WithContentTypeNosniff("nosniff"),
    secure.WithXFrameOptions("DENY"),
    secure.WithHSTSMaxAge(31536000),           // 1 year
    secure.WithHSTSIncludeSubdomains(true),
    secure.WithContentSecurityPolicy(
        "default-src 'self'; " +
        "script-src 'self' 'unsafe-inline' 'unsafe-eval'; " +
        "style-src 'self' 'unsafe-inline'; " +
        "img-src 'self' data: https:; " +
        "font-src 'self' data:;",
    ),
    secure.WithReferrerPolicy("strict-origin-when-cross-origin"),
    secure.WithPermissionsPolicy(
        "geolocation=(self), " +
        "microphone=(), " +
        "camera=(), " +
        "payment=()",
    ),
))
```

**Default Headers:**
```
X-XSS-Protection: 1; mode=block
X-Content-Type-Options: nosniff
X-Frame-Options: SAMEORIGIN
```

**Best Practices:**
- Enable HSTS only if you have HTTPS enabled
- Use CSP report-only mode first to test policies
- Regularly update CSP policies as needed

---

### CORS

Cross-Origin Resource Sharing middleware for API access control.

**Features:**
- Configurable allowed origins, methods, headers
- Credentials support
- Preflight request handling (OPTIONS)
- Max age configuration
- Automatic Vary header management

**Usage:**

```go
import "github.com/xushuhui/ares-contrib/middleware/cors"

// Allow all origins (development only!)
app.Use(cors.New())

// Production configuration
app.Use(cors.New(
    cors.WithAllowedOrigins([]string{
        "https://example.com",
        "https://www.example.com",
    }),
    cors.WithAllowedMethods([]string{
        "GET", "POST", "PUT", "DELETE", "OPTIONS",
    }),
    cors.WithAllowedHeaders([]string{
        "Authorization",
        "Content-Type",
        "X-Requested-With",
    }),
    cors.WithExposedHeaders([]string{
        "X-Total-Count",
        "X-Page-Count",
    }),
    cors.WithAllowCredentials(true),
    cors.WithMaxAge(3600), // 1 hour
))

// API endpoints
app.GET("/api/data", handler)
```

**CORS vs Credentials:**
```go
// ❌ WRONG: Cannot use wildcard with credentials
app.Use(cors.New(
    cors.WithAllowCredentials(true),
))

// ✅ CORRECT: Specify origins when using credentials
app.Use(cors.New(
    cors.WithAllowedOrigins([]string{"https://example.com"}),
    cors.WithAllowCredentials(true),
))
```

**Best Practices:**
- Never use wildcard (`*`) with `AllowCredentials: true`
- Be specific about allowed origins in production
- Set appropriate MaxAge to reduce preflight requests

---

### JWT Authentication

JWT-based token authentication middleware.

**Features:**
- Multiple signing algorithms (HS256, HS512, etc.)
- Custom claims support
- Detailed error classification
- JSON error responses
- Context-based claims storage

**Usage:**

```go
import (
    "github.com/golang-jwt/jwt/v5"
    "github.com/xushuhui/ares-contrib/middleware/jwt"
)

// Simple usage
app.Use(jwt.New([]byte("your-secret-key")))

// With custom claims
type CustomClaims struct {
    UserID   string `json:"user_id"`
    Email    string `json:"email"`
    jwt.RegisteredClaims
}

api := app.Group("/api", jwt.New(
    []byte("your-secret-key"),
    jwt.WithSigningMethod(jwt.SigningMethodHS256),
    jwt.WithClaims(func() jwt.Claims {
        return &CustomClaims{}
    }),
    jwt.WithContextKey("user"),
))

// Access claims in handler
api.GET("/profile", func(ctx *ares.Context) error {
    claims, ok := jwt.GetClaims(ctx.Request.Context())
    if !ok {
        return ctx.JSON(401, map[string]string{"error": "unauthorized"})
    }

    customClaims, ok := claims.(*CustomClaims)
    if !ok {
        return ctx.JSON(500, map[string]string{"error": "invalid claims type"})
    }

    return ctx.JSON(200, map[string]interface{}{
        "user_id": customClaims.UserID,
        "email":   customClaims.Email,
    })
})
```

**Creating a Token:**

```go
func generateToken(userID string) (string, error) {
    claims := jwt.MapClaims{
        "user_id": userID,
        "exp":     time.Now().Add(time.Hour * 24).Unix(),
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString([]byte("your-secret-key"))
}
```

**Error Responses:**

```json
// Missing token
{
  "code": 401,
  "message": "JWT token is missing"
}

// Expired token
{
  "code": 401,
  "message": "JWT token has expired"
}

// Invalid token
{
  "code": 401,
  "message": "token is invalid"
}
```

**Best Practices:**
- Store secrets in environment variables
- Use strong, random secrets (256+ bits)
- Implement token refresh mechanism
- Validate token expiration on each request

---

### GZIP Compression

Response compression middleware to reduce bandwidth usage.

**Features:**
- Configurable compression level (1-9)
- Minimum response size threshold
- Exclude specific file extensions
- Exclude specific paths (WebSockets, streams)
- Writer pooling for performance

**Usage:**

```go
import "github.com/xushuhui/ares-contrib/middleware/gzip"

// Default configuration
app.Use(gzip.New())

// Custom configuration
app.Use(gzip.New(
    gzip.WithLevel(5),                        // Compression level (1-9)
    gzip.WithMinLength(1024),                 // Only compress > 1KB
    gzip.WithExcludedExtensions([]string{
        ".png", ".jpg", ".jpeg", ".gif",     // Already compressed
        ".zip", ".gz", ".tar",
        ".pdf", ".mp4", ".mp3",
    }),
    gzip.WithExcludedPaths([]string{
        "/api/stream",                        // WebSocket/streams
        "/ws",
        "/download",
    }),
))
```

**Default Excluded Extensions:**
- Images: `.png`, `.jpg`, `.jpeg`, `.gif`, `.webp`, `.svg`
- Archives: `.zip`, `.gz`, `.tar`, `.rar`, `.7z`
- Media: `.mp4`, `.avi`, `.mov`, `.mp3`, `.wav`
- Documents: `.pdf`

**Performance Tips:**
- Level 5-7 provides good balance
- Don't compress already compressed files (images, videos)
- Exclude streaming endpoints
- Monitor CPU usage vs bandwidth savings

---

### Body Limit

Limits request body size to prevent memory exhaustion attacks.

**Features:**
- Configurable size limit
- Uses `http.MaxBytesReader` for efficiency
- Returns 413 Payload Too Large on overflow

**Usage:**

```go
import "github.com/xushuhui/ares-contrib/middleware/bodylimit"

// Global limit: 10MB
app.Use(bodylimit.New(10 * 1024 * 1024))

// Different limits per route
uploadGroup := app.Group("/upload", bodylimit.New(100 * 1024 * 1024)) // 100MB
uploadGroup.POST("/image", uploadImageHandler)

apiGroup := app.Group("/api", bodylimit.New(1 * 1024 * 1024)) // 1MB
apiGroup.POST("/data", postDataHandler)
```

**Error Response:**
```
HTTP 413 Payload Too Large
```

**Best Practices:**
- Set lower limits for API endpoints
- Set higher limits for file uploads
- Consider using chunked upload for large files

---

### Rate Limiter

Token bucket rate limiter to prevent API abuse.

**Features:**
- Per-IP rate limiting by default
- Custom key extraction (user ID, API key)
- Configurable rate and burst
- Automatic cleanup of old limiters
- Custom error handler support

**Usage:**

```go
import "github.com/xushuhui/ares-contrib/middleware/ratelimiter"

// Default: 10 requests/second, burst of 20
app.Use(ratelimiter.New())

// Custom configuration
app.Use(ratelimiter.New(
    ratelimiter.WithRate(100),              // 100 requests/second
    ratelimiter.WithBurst(200),             // Allow burst of 200
    ratelimiter.WithKeyFunc(func(r *http.Request) string {
        // Rate limit by user ID instead of IP
        userID := r.Header.Get("X-User-ID")
        if userID != "" {
            return "user:" + userID
        }
        return "ip:" + r.RemoteAddr
    }),
    ratelimiter.WithErrorHandler(func(w http.ResponseWriter, r *http.Request) {
        http.Error(w, "Rate limit exceeded. Please try again later.", 429)
    }),
))
```

**Different Limits for Different Routes:**

```go
// Public API: 10 req/s
publicAPI := app.Group("/api/public", ratelimiter.New(
    ratelimiter.WithRate(10),
))

// Authenticated users: 100 req/s
userAPI := app.Group("/api/user", ratelimiter.New(
    ratelimiter.WithRate(100),
    ratelimiter.WithKeyFunc(func(r *http.Request) string {
        return r.Context().Value("user_id").(string)
    }),
))
```

**Best Practices:**
- Use different limits for public vs authenticated users
- Consider burst capacity for user experience
- Monitor and adjust based on usage patterns
- Implement request retry with exponential backoff

---

## 🎯 Best Practices

### Middleware Order

The order of middleware matters! Here's the recommended order:

```go
app := ares.New()

// 1. Request ID (first, for tracing)
app.Use(requestid.New())

// 2. Security headers (early)
app.Use(secure.New())

// 3. Rate limiting (before expensive operations)
app.Use(ratelimiter.New())

// 4. Body limit (before reading body)
app.Use(bodylimit.New(10 * 1024 * 1024))

// 5. CORS (before auth)
app.Use(cors.New())

// 6. Compression (before response)
app.Use(gzip.New())

// 7. Authentication (for protected routes)
api := app.Group("/api", jwt.New(secret))
```

### Performance Tips

1. **Use GZIP for text-based content only**
   ```go
   gzip.WithExcludedExtensions([]string{".png", ".jpg", ".mp4"})
   ```

2. **Set appropriate rate limits**
   ```go
   // Too low: poor UX
   ratelimiter.WithRate(1)  // ❌

   // Too high: no protection
   ratelimiter.WithRate(10000)  // ❌

   // Just right
   ratelimiter.WithRate(100)  // ✅
   ```

3. **Use context instead of global variables**
   ```go
   // ❌ WRONG
   var userID string

   // ✅ CORRECT
   userID := ctx.GetString("user_id")
   ```

### Security Checklist

- [ ] Enable HTTPS in production
- [ ] Set secure cookies
- [ ] Implement rate limiting
- [ ] Use CSP to prevent XSS
- [ ] Enable HSTS with long max-age
- [ ] Validate and sanitize input
- [ ] Keep dependencies updated
- [ ] Log security events
- [ ] Implement authentication
- [ ] Use CORS correctly (no wildcard with credentials)

---

## 🧪 Testing

All middleware has comprehensive test coverage:

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test ./... -cover

# Run tests with coverage report
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

**Test Coverage Summary:**

```
Middleware          Coverage    Tests
----------------------------------------
RequestID           100.0%      6
Secure              100.0%      11
CORS                96.2%       14
JWT                 85.7%       10
GZIP                80.9%       14
BodyLimit           72.7%       8
RateLimiter         72.0%       6
----------------------------------------
TOTAL               ~87%        69
```

---

## 📊 Benchmarks

Run benchmarks to test middleware performance:

```bash
cd middleware/<middleware-name>
go test -bench=. -benchmem
```

Example results (Apple M1 Pro, Go 1.23):

```
BenchmarkRequestID-8       10000000    105 ns/op    0 B/op    0 allocs/op
BenchmarkSecure-8          10000000    120 ns/op    0 B/op    0 allocs/op
BenchmarkCORS-8            5000000     250 ns/op    0 B/op    0 allocs/op
BenchmarkJWT-8            1000000     1200 ns/op   512 B/op  8 allocs/op
BenchmarkGZIP-8           3000000     450 ns/op    128 B/op  2 allocs/op
```

---

## 📦 Dependencies

| Package | Version | Purpose |
|---------|---------|---------|
| [github.com/golang-jwt/jwt/v5](https://github.com/golang-jwt/jwt) | ^5.2.0 | JWT implementation |
| [github.com/google/uuid](https://github.com/google/uuid) | ^1.5.0 | UUID generation |
| [golang.org/x/time/rate](https://golang.org/x/time/rate) | latest | Rate limiting |
| [github.com/xushuhui/ares](https://github.com/xushuhui/ares) | latest | Core framework |

---

## 🤝 Contributing

We welcome contributions! Please follow these steps:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Write tests for your changes
4. Ensure all tests pass (`go test ./...`)
5. Maintain test coverage above 80%
6. Commit your changes (`git commit -m 'Add amazing feature'`)
7. Push to the branch (`git push origin feature/amazing-feature`)
8. Open a Pull Request

**Development Requirements:**
- Go 1.21 or higher
- Follow Go best practices and effective Go guidelines
- Write clear, idiomatic Go code
- Include tests for new features
- Update documentation as needed

---

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

## 🔗 Links

- [Ares Framework](https://github.com/xushuhui/ares) - Core framework
- [Documentation](https://github.com/xushuhui/ares/wiki) - Official docs
- [Examples](./examples/) - Usage examples
- [Issues](https://github.com/xushuhui/ares-contrib/issues) - Bug reports and feature requests

---

## 🌟 Star History

If you find this project useful, please consider giving it a ⭐ star!

---

Made with ❤️ by the Ares community
