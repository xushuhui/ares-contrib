package main

import (
	"net/http"

	"github.com/xushuhui/ares"
	"github.com/xushuhui/ares-contrib/middleware/cors"
	"github.com/xushuhui/ares-contrib/middleware/gzip"
	"github.com/xushuhui/ares-contrib/middleware/ratelimiter"
	"github.com/xushuhui/ares-contrib/middleware/requestid"
	"github.com/xushuhui/ares-contrib/middleware/secure"
)

type User struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Age  int    `json:"age"`
}

type CreateUserRequest struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func main() {
	// Create Ares instance with default middleware (logger + recovery)
	app := ares.Default()

	// Add contrib middleware
	app.Use(requestid.New())
	app.Use(cors.New(
		cors.WithAllowedOrigins([]string{"*"}),
		cors.WithAllowedMethods([]string{"GET", "POST", "PUT", "DELETE"}),
	))
	app.Use(gzip.New(gzip.WithLevel(5)))
	app.Use(secure.New(
		secure.WithXFrameOptions("DENY"),
		secure.WithContentSecurityPolicy("default-src 'self'"),
	))
	app.Use(ratelimiter.New(
		ratelimiter.WithRate(100),
		ratelimiter.WithBurst(200),
	))

	// Basic routes
	app.GET("/health", HealthCheck)
	app.GET("/panic", PanicHandler)

	// API v1 group with middleware
	v1 := app.Group("/api/v1")
	v1.Use(AuthMiddleware) // Group-specific middleware
	v1.GET("/status", GetStatus)
	v1.GET("/users/{id}", GetUser)
	v1.POST("/users", CreateUser)
	v1.GET("/profile", GetProfile) // Demonstrates context storage

	// API v2 group (nested group example)
	v2 := app.Group("/api/v2")
	v2.GET("/status", GetStatusV2)

	// Admin group with additional middleware
	admin := app.Group("/admin")
	admin.Use(AdminMiddleware)
	admin.GET("/stats", GetAdminStats)

	// Start server
	app.Logger().Info("starting example server on :8080")
	app.Logger().Info("try these endpoints:")
	app.Logger().Info("  GET  http://localhost:8080/health")
	app.Logger().Info("  GET  http://localhost:8080/api/v1/status")
	app.Logger().Info("  GET  http://localhost:8080/api/v1/users/123")
	app.Logger().Info("  POST http://localhost:8080/api/v1/users")
	app.Logger().Info("  GET  http://localhost:8080/api/v1/profile")
	app.Logger().Info("  GET  http://localhost:8080/api/v2/status")
	app.Logger().Info("  GET  http://localhost:8080/admin/stats")

	if err := app.Run(":8080"); err != nil {
		app.Logger().Error("server error", "error", err)
	}
}

// AuthMiddleware demonstrates middleware that uses context storage
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// In a real app, you would validate JWT token here
		// For demo purposes, we'll just set some user data
		w.Header().Set("X-Auth", "authenticated")
		next.ServeHTTP(w, r)
	})
}

// AdminMiddleware demonstrates additional authorization
func AdminMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Admin", "true")
		next.ServeHTTP(w, r)
	})
}

// HealthCheck handler - basic health check
func HealthCheck(ctx *ares.Context) error {
	return ctx.JSON(http.StatusOK, map[string]string{
		"status": "ok",
	})
}

// GetStatus handler - demonstrates API versioning
func GetStatus(ctx *ares.Context) error {
	return ctx.JSON(http.StatusOK, map[string]any{
		"version": "1.0.0",
		"status":  "running",
		"message": "API v1 is operational",
	})
}

// GetStatusV2 handler - demonstrates API versioning
func GetStatusV2(ctx *ares.Context) error {
	return ctx.JSON(http.StatusOK, map[string]any{
		"version":  "2.0.0",
		"status":   "running",
		"message":  "API v2 is operational",
		"features": []string{"enhanced", "faster", "better"},
	})
}

// GetUser handler - demonstrates URL parameters
func GetUser(ctx *ares.Context) error {
	id := ctx.Param("id")
	ctx.Logger().Info("fetching user", "id", id)

	// Simulate user data
	user := User{
		ID:   id,
		Name: "John Doe",
		Age:  30,
	}

	return ctx.JSON(http.StatusOK, user)
}

// CreateUser handler - demonstrates request binding
func CreateUser(ctx *ares.Context) error {
	var req CreateUserRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{
			"error": "invalid request body",
		})
	}

	ctx.Logger().Info("creating user", "name", req.Name, "age", req.Age)

	// Simulate user creation
	user := User{
		ID:   "123",
		Name: req.Name,
		Age:  req.Age,
	}

	return ctx.JSON(http.StatusCreated, map[string]any{
		"message": "user created",
		"user":    user,
	})
}

// GetProfile handler - demonstrates context key-value storage
func GetProfile(ctx *ares.Context) error {
	// Store some values in context
	ctx.Set("user_id", 123)
	ctx.Set("username", "john_doe")
	ctx.Set("is_admin", false)
	ctx.Set("email", "john@example.com")

	// Retrieve values from context
	userID := ctx.GetInt("user_id")
	username := ctx.GetString("username")
	isAdmin := ctx.GetBool("is_admin")
	email := ctx.GetString("email")

	return ctx.JSON(http.StatusOK, map[string]any{
		"user_id":  userID,
		"username": username,
		"is_admin": isAdmin,
		"email":    email,
		"message":  "Profile data retrieved from context storage",
	})
}

// GetAdminStats handler - demonstrates admin-only endpoint
func GetAdminStats(ctx *ares.Context) error {
	return ctx.JSON(http.StatusOK, map[string]any{
		"total_users":    1000,
		"active_users":   850,
		"total_requests": 50000,
		"uptime":         "99.9%",
	})
}

// PanicHandler - demonstrates panic recovery
func PanicHandler(ctx *ares.Context) error {
	panic("intentional panic for testing recovery middleware")
}
