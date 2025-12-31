# Ares Basic Example

This example demonstrates the core features of the Ares web framework, including:

- Default middleware (logger + recovery) via `ares.Default()`
- Contrib middleware (CORS, Gzip, Rate Limiter, Request ID, Secure Headers)
- Route groups with middleware
- Context key-value storage
- URL parameters and request binding
- API versioning
- Error handling and panic recovery

## Running the Example

```bash
cd contrib/examples/basic
go run main.go
```

The server will start on `http://localhost:8080`.

## Available Endpoints

### Basic Endpoints

**Health Check**
```bash
curl http://localhost:8080/health
```

**Panic Test** (demonstrates recovery middleware)
```bash
curl http://localhost:8080/panic
```

### API v1 Endpoints

**Status**
```bash
curl http://localhost:8080/api/v1/status
```

**Get User** (demonstrates URL parameters)
```bash
curl http://localhost:8080/api/v1/users/123
```

**Create User** (demonstrates request binding)
```bash
curl -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{"name":"Alice","age":25}'
```

**Get Profile** (demonstrates context storage)
```bash
curl http://localhost:8080/api/v1/profile
```

### API v2 Endpoints

**Status** (demonstrates API versioning)
```bash
curl http://localhost:8080/api/v2/status
```

### Admin Endpoints

**Admin Stats** (demonstrates group-specific middleware)
```bash
curl http://localhost:8080/admin/stats
```

## Features Demonstrated

### 1. Default Middleware

```go
// Create app with default middleware (logger + recovery)
app := ares.Default()
```

The `Default()` method automatically includes:
- **Logger middleware**: Logs all HTTP requests
- **Recovery middleware**: Recovers from panics

### 2. Contrib Middleware

```go
// Request ID
app.Use(requestid.New())

// CORS
app.Use(cors.New(
    cors.WithAllowedOrigins([]string{"*"}),
    cors.WithAllowedMethods([]string{"GET", "POST", "PUT", "DELETE"}),
))

// Gzip compression
app.Use(gzip.New(gzip.WithLevel(5)))

// Security headers
app.Use(secure.New(
    secure.WithXFrameOptions("DENY"),
    secure.WithContentSecurityPolicy("default-src 'self'"),
))

// Rate limiting
app.Use(ratelimiter.New(
    ratelimiter.WithRate(100),
    ratelimiter.WithBurst(200),
))
```

### 3. Route Groups

```go
// API v1 group with middleware
v1 := app.Group("/api/v1")
v1.Use(AuthMiddleware)
v1.GET("/status", GetStatus)
v1.GET("/users/{id}", GetUser)

// API v2 group
v2 := app.Group("/api/v2")
v2.GET("/status", GetStatusV2)

// Admin group with additional middleware
admin := app.Group("/admin")
admin.Use(AdminMiddleware)
admin.GET("/stats", GetAdminStats)
```

### 4. Context Key-Value Storage

```go
func GetProfile(ctx *ares.Context) error {
    // Store values
    ctx.Set("user_id", 123)
    ctx.Set("username", "john_doe")

    // Retrieve values
    userID := ctx.GetInt("user_id")
    username := ctx.GetString("username")

    return ctx.JSON(200, map[string]any{
        "user_id": userID,
        "username": username,
    })
}
```

### 5. URL Parameters

```go
func GetUser(ctx *ares.Context) error {
    id := ctx.Param("id")
    // Use the id...
}
```

### 6. Request Binding

```go
func CreateUser(ctx *ares.Context) error {
    var req CreateUserRequest
    if err := ctx.Bind(&req); err != nil {
        return ctx.JSON(400, map[string]string{"error": "invalid request"})
    }
    // Use the request data...
}
```

## Testing Middleware

### CORS Headers

```bash
curl -H "Origin: http://example.com" \
     -H "Access-Control-Request-Method: POST" \
     -X OPTIONS http://localhost:8080/api/v1/users -v
```

### Gzip Compression

```bash
curl -H "Accept-Encoding: gzip" http://localhost:8080/api/v1/status -v
```

### Security Headers

```bash
curl http://localhost:8080/health -v | grep -E "X-Frame-Options|X-XSS-Protection|X-Content-Type-Options"
```

### Rate Limiting

```bash
# Send multiple requests quickly to trigger rate limiting
for i in {1..150}; do
  curl http://localhost:8080/health
done
```

## Project Structure

```
basic/
├── main.go       # Main application with all handlers
├── go.mod        # Go module definition
└── README.md     # This file
```

## Next Steps

- Explore the [main Ares documentation](../../../README.md)
- Check out [contrib middleware documentation](../../README.md)
- Try modifying the example to add your own endpoints
- Experiment with different middleware configurations
