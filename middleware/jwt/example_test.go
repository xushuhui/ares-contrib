package jwt

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Example_generateToken demonstrates how to generate JWT tokens
func Example_generateToken() {
	secret := []byte("your-secret-key")

	// Example 1: Generate token with MapClaims (simple use case)
	claims := map[string]interface{}{
		"user_id": "123",
		"email":   "user@example.com",
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	}

	token, err := GenerateTokenWithDefaultClaims(secret, claims)
	if err != nil {
		panic(err)
	}

	fmt.Println("Generated token:", token)
}

// Example_generateTokenWithCustomClaims demonstrates generating tokens with custom claims
func Example_generateTokenWithCustomClaims() {
	secret := []byte("your-secret-key")

	// Define custom claims struct
	type CustomClaims struct {
		UserID   string `json:"user_id"`
		Username string `json:"username"`
		Role     string `json:"role"`
		jwt.RegisteredClaims
	}

	// Create custom claims
	claims := CustomClaims{
		UserID:   "123",
		Username: "john_doe",
		Role:     "admin",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "my-app",
			Subject:   "user-auth",
		},
	}

	// Generate token
	token, err := GenerateToken(secret, claims)
	if err != nil {
		panic(err)
	}

	fmt.Println("Generated token with custom claims:", token)
}

// Example_generateTokenWithCustomSigningMethod demonstrates generating tokens with custom signing method
func Example_generateTokenWithCustomSigningMethod() {
	secret := []byte("your-secret-key")

	claims := jwt.MapClaims{
		"user_id": "123",
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	}

	// Generate token with HS512 signing method
	token, err := GenerateToken(secret, claims, WithSigningMethod(jwt.SigningMethodHS512))
	if err != nil {
		panic(err)
	}

	fmt.Println("Generated token with HS512:", token)
}

// Example_generateAndValidate demonstrates the complete flow: generate token, then validate with middleware
func Example_generateAndValidate() {
	secret := []byte("your-secret-key")

	// Step 1: Generate token (e.g., during login)
	claims := map[string]interface{}{
		"user_id": "123",
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	}

	token, err := GenerateTokenWithDefaultClaims(secret, claims)
	if err != nil {
		panic(err)
	}

	fmt.Println("Generated token for authentication:", token)

	// Step 2: Use the token in Authorization header
	// In a real HTTP request:
	// req.Header.Set("Authorization", "Bearer " + token)

	// Step 3: Middleware will validate the token and extract claims
	// The middleware (New function) will automatically validate the token
	// and make the claims available via GetClaims()
}
