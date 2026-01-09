# JWT Middleware

JWT middleware for Ares web framework that provides both token validation and token generation capabilities.

## Features

- **Token Validation**: Validates JWT tokens in incoming HTTP requests
- **Token Generation**: Generate signed JWT tokens with the same configuration
- **Custom Claims Support**: Works with both `jwt.MapClaims` and custom claim types
- **Flexible Signing Methods**: Supports all JWT signing methods (HS256, HS512, etc.)
- **Context Storage**: Automatically stores validated claims in request context
- **Custom Context Keys**: Configure custom keys for storing claims in context
- **JSON Error Responses**: Returns standardized JSON error responses for authentication failures
- **Comprehensive Error Handling**: Specific error types for different validation failures

## Installation

```bash
go get github.com/xushuhui/ares-contrib/middleware/jwt
```

## Usage

### Basic Usage

```go
package main

import (
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/xushuhui/ares"
	"github.com/xushuhui/ares-contrib/middleware/jwt"
)

func main() {
	secret := []byte("your-secret-key")

	// Create Ares app
	app := ares.Default()

	// Add JWT middleware
	app.Use(jwt.New(secret))

	// Protected route
	app.GET("/protected", func(ctx *ares.Context) error {
		claims, ok := jwt.GetClaims(ctx.Request().Context())
		if !ok {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{
				"error": "no claims found",
			})
		}

		return ctx.JSON(http.StatusOK, map[string]interface{}{
			"message": "access granted",
			"claims":  claims,
		})
	})

	app.Run(":8080")
}
```

### Token Generation

```go
package main

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/xushuhui/ares-contrib/middleware/jwt"
)

func generateToken() {
	secret := []byte("your-secret-key")

	// Method 1: Using GenerateTokenWithDefaultClaims (simpler)
	claims := map[string]interface{}{
		"user_id": "123",
		"email":   "user@example.com",
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	}

	token, err := jwt.GenerateTokenWithDefaultClaims(secret, claims)
	if err != nil {
		panic(err)
	}

	println("Generated token:", token)

	// Method 2: Using GenerateToken with custom claims
	type CustomClaims struct {
		UserID   string `json:"user_id"`
		Username string `json:"username"`
		jwt.RegisteredClaims
	}

	customClaims := CustomClaims{
		UserID:   "123",
		Username: "john_doe",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token, err = jwt.GenerateToken(secret, customClaims)
	if err != nil {
		panic(err)
	}

	println("Generated token with custom claims:", token)
}
```

## Configuration Options

### WithSigningMethod

Configure the expected signing method for token validation:

```go
app.Use(jwt.New(secret, jwt.WithSigningMethod(jwt.SigningMethodHS512)))
```

### WithClaims

Provide a function to create custom claim instances for parsing:

```go
type CustomClaims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

app.Use(jwt.New(secret, jwt.WithClaims(func() jwt.Claims {
	return &CustomClaims{}
})))
```

### WithContextKey

Use a custom key for storing claims in request context:

```go
app.Use(jwt.New(secret, jwt.WithContextKey("custom_user")))

// Retrieve with custom key
claims, ok := jwt.GetClaimsWithKey(ctx.Request().Context(), "custom_user")
```

## API Reference

### Middleware Functions

#### `New(signingKey []byte, opts ...Option) func(http.Handler) http.Handler`

Creates a JWT middleware that validates incoming tokens.

**Parameters:**
- `signingKey`: The secret key used to verify token signatures
- `opts`: Optional configuration options

**Returns:** Standard HTTP middleware function

#### `GetClaims(ctx context.Context) (jwt.Claims, bool)`

Extracts JWT claims from request context using default key ("user").

#### `GetClaimsWithKey(ctx context.Context, key string) (jwt.Claims, bool)`

Extracts JWT claims from request context using custom key.

### Token Generation Functions

#### `GenerateToken(signingKey []byte, claims jwt.Claims, opts ...Option) (string, error)`

Creates a signed JWT token with the given claims.

**Parameters:**
- `signingKey`: The secret key used to sign the token
- `claims`: The JWT claims to include in the token
- `opts`: Optional configuration (same as middleware options)

**Returns:**
- `string`: The signed JWT token
- `error`: Error if token generation fails

#### `GenerateTokenWithDefaultClaims(signingKey []byte, claims map[string]interface{}, opts ...Option) (string, error)`

Convenience function that creates a token with `jwt.MapClaims`.

**Parameters:**
- `signingKey`: The secret key used to sign the token
- `claims`: Map of claim key-value pairs
- `opts`: Optional configuration options

**Returns:**
- `string`: The signed JWT token
- `error`: Error if token generation fails

## Error Handling

The middleware returns appropriate HTTP status codes and JSON error responses:

- `401 Unauthorized`: Missing, invalid, expired, or malformed tokens
- `401 Unauthorized`: Wrong signing method

Example error response:
```json
{
	"code": 401,
	"message": "JWT token has expired"
}
```

## Complete Example

```go
package main

import (
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/xushuhui/ares"
	"github.com/xushuhui/ares-contrib/middleware/jwt"
)

func main() {
	secret := []byte("your-secret-key")

	app := ares.Default()

	// Login endpoint - generates tokens
	app.POST("/login", func(ctx *ares.Context) error {
		// In real app, validate credentials here
		userID := "123"

		claims := map[string]interface{}{
			"user_id": userID,
			"exp":     time.Now().Add(time.Hour * 24).Unix(),
		}

		token, err := jwt.GenerateTokenWithDefaultClaims(secret, claims)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{
				"error": "failed to generate token",
			})
		}

		return ctx.JSON(http.StatusOK, map[string]string{
			"token": token,
		})
	})

	// Protected API group
	api := app.Group("/api")
	api.Use(jwt.New(secret))

	api.GET("/profile", func(ctx *ares.Context) error {
		claims, ok := jwt.GetClaims(ctx.Request().Context())
		if !ok {
			return ctx.JSON(http.StatusUnauthorized, map[string]string{
				"error": "invalid token",
			})
		}

		// Extract user ID from claims
		mapClaims, ok := claims.(jwt.MapClaims)
		if !ok {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{
				"error": "invalid claims format",
			})
		}

		userID := mapClaims["user_id"]

		return ctx.JSON(http.StatusOK, map[string]interface{}{
			"user_id": userID,
			"message": "profile data",
		})
	})

	app.Run(":8080")
}
```

## Testing

Run tests:

```bash
go test ./middleware/jwt/ -v
```

Run specific test:

```bash
go test -v -run TestGenerateToken ./middleware/jwt/
```

## Security Considerations

- Keep your signing key secret and secure
- Use strong signing keys (at least 32 bytes for HS256)
- Set appropriate token expiration times
- Use HTTPS in production
- Consider using asymmetric signing methods (RS256) for distributed systems
- Validate all claims before trusting them
