package jwt

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestNew(t *testing.T) {
	secret := []byte("test-secret")

	// Create a valid token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": "123",
		"exp":     time.Now().Add(time.Hour).Unix(),
	})
	tokenString, err := token.SignedString(secret)
	if err != nil {
		t.Fatalf("Failed to create token: %v", err)
	}

	tests := []struct {
		name           string
		token          string
		expectedStatus int
		checkClaims    bool
	}{
		{
			name:           "Valid token",
			token:          tokenString,
			expectedStatus: http.StatusOK,
			checkClaims:    true,
		},
		{
			name:           "Missing token",
			token:          "",
			expectedStatus: http.StatusUnauthorized,
			checkClaims:    false,
		},
		{
			name:           "Invalid token",
			token:          "invalid.token.here",
			expectedStatus: http.StatusUnauthorized,
			checkClaims:    false,
		},
		{
			name:           "Malformed token",
			token:          "Bearer malformed",
			expectedStatus: http.StatusUnauthorized,
			checkClaims:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test handler
			handler := New(secret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.checkClaims {
					claims, ok := GetClaims(r.Context())
					if !ok {
						t.Error("Expected claims in context")
					}
					if claims == nil {
						t.Error("Claims should not be nil")
					}
				}
				w.WriteHeader(http.StatusOK)
			}))

			// Create request
			req := httptest.NewRequest("GET", "/test", nil)
			if tt.token != "" {
				req.Header.Set("Authorization", "Bearer "+tt.token)
			}

			// Record response
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			// Check status code
			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rr.Code)
			}
		})
	}
}

func TestJWTWithCustomClaims(t *testing.T) {
	secret := []byte("test-secret")

	type CustomClaims struct {
		UserID string `json:"user_id"`
		jwt.RegisteredClaims
	}

	// Create token with custom claims
	claims := CustomClaims{
		UserID: "123",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(secret)
	if err != nil {
		t.Fatalf("Failed to create token: %v", err)
	}

	// Create middleware with custom claims
	handler := New(secret, WithClaims(func() jwt.Claims {
		return &CustomClaims{}
	}))(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := GetClaims(r.Context())
		if !ok {
			t.Error("Expected claims in context")
		}
		customClaims, ok := claims.(*CustomClaims)
		if !ok {
			t.Error("Expected CustomClaims type")
		}
		if customClaims.UserID != "123" {
			t.Errorf("Expected UserID 123, got %s", customClaims.UserID)
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}
}

func TestJWTExpiredToken(t *testing.T) {
	secret := []byte("test-secret")

	// Create expired token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": "123",
		"exp":     time.Now().Add(-time.Hour).Unix(), // Expired 1 hour ago
	})
	tokenString, err := token.SignedString(secret)
	if err != nil {
		t.Fatalf("Failed to create token: %v", err)
	}

	handler := New(secret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401 for expired token, got %d", rr.Code)
	}
}

func TestJWTWrongSigningMethod(t *testing.T) {
	secret := []byte("test-secret")

	// Create token with HS512
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, jwt.MapClaims{
		"user_id": "123",
		"exp":     time.Now().Add(time.Hour).Unix(),
	})
	tokenString, err := token.SignedString(secret)
	if err != nil {
		t.Fatalf("Failed to create token: %v", err)
	}

	// Middleware expects HS256 (default)
	handler := New(secret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401 for wrong signing method, got %d", rr.Code)
	}
}

func TestGetClaims(t *testing.T) {
	claims := jwt.MapClaims{"user_id": "123"}
	ctx := context.WithValue(context.Background(), contextKey("user"), claims)

	retrievedClaims, ok := GetClaims(ctx)
	if !ok {
		t.Error("Expected to retrieve claims")
	}
	if retrievedClaims == nil {
		t.Error("Claims should not be nil")
	}
}

func TestGetClaimsWithKey(t *testing.T) {
	claims := jwt.MapClaims{"user_id": "123"}
	ctx := context.WithValue(context.Background(), contextKey("custom"), claims)

	retrievedClaims, ok := GetClaimsWithKey(ctx, "custom")
	if !ok {
		t.Error("Expected to retrieve claims with custom key")
	}
	if retrievedClaims == nil {
		t.Error("Claims should not be nil")
	}
}

func TestJWTWithContextKey(t *testing.T) {
	secret := []byte("test-secret")

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": "123",
		"exp":     time.Now().Add(time.Hour).Unix(),
	})
	tokenString, err := token.SignedString(secret)
	if err != nil {
		t.Fatalf("Failed to create token: %v", err)
	}

	handler := New(secret, WithContextKey("custom"))(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := GetClaimsWithKey(r.Context(), "custom")
		if !ok {
			t.Error("Expected claims with custom key")
		}
		if claims == nil {
			t.Error("Claims should not be nil")
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}
}

func TestGenerateToken(t *testing.T) {
	secret := []byte("test-secret")

	tests := []struct {
		name        string
		claims      jwt.Claims
		opts        []Option
		expectError bool
	}{
		{
			name: "Generate token with MapClaims",
			claims: jwt.MapClaims{
				"user_id": "123",
				"exp":     time.Now().Add(time.Hour).Unix(),
			},
			opts:        nil,
			expectError: false,
		},
		{
			name: "Generate token with custom signing method",
			claims: jwt.MapClaims{
				"user_id": "456",
				"exp":     time.Now().Add(time.Hour).Unix(),
			},
			opts:        []Option{WithSigningMethod(jwt.SigningMethodHS512)},
			expectError: false,
		},
		{
			name: "Generate token with RegisteredClaims",
			claims: &jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
				Subject:   "test",
			},
			opts:        nil,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokenString, err := GenerateToken(secret, tt.claims, tt.opts...)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("Expected no error, got %v", err)
			}

			if tokenString == "" {
				t.Error("Token string should not be empty")
			}

			// Verify the token can be parsed and validated
			token, err := jwt.ParseWithClaims(tokenString, tt.claims, func(token *jwt.Token) (interface{}, error) {
				return secret, nil
			})

			if err != nil {
				t.Fatalf("Failed to parse generated token: %v", err)
			}

			if !token.Valid {
				t.Error("Generated token should be valid")
			}
		})
	}
}

func TestGenerateTokenWithNilKey(t *testing.T) {
	claims := jwt.MapClaims{"user_id": "123"}

	_, err := GenerateToken(nil, claims)
	if err == nil {
		t.Error("Expected error for nil signing key")
	}
	if err.Error() != "signing key is nil" {
		t.Errorf("Expected 'signing key is nil' error, got %v", err)
	}
}

func TestGenerateTokenWithDefaultClaims(t *testing.T) {
	secret := []byte("test-secret")

	tests := []struct {
		name        string
		claims      map[string]interface{}
		expectError bool
	}{
		{
			name: "Generate token with simple claims",
			claims: map[string]interface{}{
				"user_id": "123",
				"exp":     time.Now().Add(time.Hour).Unix(),
			},
			expectError: false,
		},
		{
			name: "Generate token with multiple fields",
			claims: map[string]interface{}{
				"user_id":  "456",
				"username": "testuser",
				"role":     "admin",
				"exp":      time.Now().Add(2 * time.Hour).Unix(),
				"iat":      time.Now().Unix(),
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokenString, err := GenerateTokenWithDefaultClaims(secret, tt.claims)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("Expected no error, got %v", err)
			}

			if tokenString == "" {
				t.Error("Token string should not be empty")
			}

			// Verify the token can be parsed
			mapClaims := jwt.MapClaims{}
			token, err := jwt.ParseWithClaims(tokenString, mapClaims, func(token *jwt.Token) (interface{}, error) {
				return secret, nil
			})

			if err != nil {
				t.Fatalf("Failed to parse generated token: %v", err)
			}

			if !token.Valid {
				t.Error("Generated token should be valid")
			}

			// Verify claims are preserved
			for key, expectedValue := range tt.claims {
				actualValue, ok := mapClaims[key]
				if !ok {
					t.Errorf("Expected claim key %s not found", key)
					continue
				}

				// Convert Unix timestamps if necessary
				if key == "exp" || key == "iat" {
					if expectedFloat, ok := expectedValue.(float64); ok {
						expectedValue = int64(expectedFloat)
					}
					if actualFloat, ok := actualValue.(float64); ok {
						actualValue = int64(actualFloat)
					}
				}

				if actualValue != expectedValue {
					t.Errorf("Expected claim value %v for key %s, got %v", expectedValue, key, actualValue)
				}
			}
		})
	}
}

func TestGenerateAndValidateToken(t *testing.T) {
	secret := []byte("test-secret")

	// Generate token using GenerateTokenWithDefaultClaims
	claims := map[string]interface{}{
		"user_id": "123",
		"exp":     time.Now().Add(time.Hour).Unix(),
	}

	tokenString, err := GenerateTokenWithDefaultClaims(secret, claims)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// Validate the token using middleware
	handler := New(secret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		retrievedClaims, ok := GetClaims(r.Context())
		if !ok {
			t.Error("Expected claims in context")
		}
		if retrievedClaims == nil {
			t.Error("Claims should not be nil")
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}
}

func TestGenerateTokenWithCustomClaims(t *testing.T) {
	secret := []byte("test-secret")

	type CustomClaims struct {
		UserID string `json:"user_id"`
		Email  string `json:"email"`
		jwt.RegisteredClaims
	}

	claims := CustomClaims{
		UserID: "123",
		Email:  "test@example.com",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	tokenString, err := GenerateToken(secret, claims)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	if tokenString == "" {
		t.Error("Token string should not be empty")
	}

	// Verify the token can be parsed
	parsedClaims := &CustomClaims{}
	token, err := jwt.ParseWithClaims(tokenString, parsedClaims, func(token *jwt.Token) (interface{}, error) {
		return secret, nil
	})

	if err != nil {
		t.Fatalf("Failed to parse generated token: %v", err)
	}

	if !token.Valid {
		t.Error("Generated token should be valid")
	}

	if parsedClaims.UserID != "123" {
		t.Errorf("Expected UserID 123, got %s", parsedClaims.UserID)
	}

	if parsedClaims.Email != "test@example.com" {
		t.Errorf("Expected Email test@example.com, got %s", parsedClaims.Email)
	}
}

func TestGenerateTokenWithCustomSigningMethod(t *testing.T) {
	secret := []byte("test-secret")

	claims := jwt.MapClaims{
		"user_id": "123",
		"exp":     time.Now().Add(time.Hour).Unix(),
	}

	// Generate token with HS512
	tokenString, err := GenerateToken(secret, claims, WithSigningMethod(jwt.SigningMethodHS512))
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// Verify the token uses HS512
	token, _ := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return secret, nil
	})

	if token.Method != jwt.SigningMethodHS512 {
		t.Errorf("Expected signing method HS512, got %v", token.Method)
	}
}
