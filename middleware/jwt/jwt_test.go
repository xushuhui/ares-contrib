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
