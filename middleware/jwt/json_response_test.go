package jwt

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// ErrorResponse represents the JSON error response structure
type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// TestJSONResponses verifies that all error responses are in JSON format
func TestJSONResponses(t *testing.T) {
	secret := []byte("test-secret")

	tests := []struct {
		name           string
		token          string
		expectedStatus int
		expectedMsg    string
	}{
		{
			name:           "Missing token returns JSON error",
			token:          "",
			expectedStatus: http.StatusUnauthorized,
			expectedMsg:    ErrMissingJwtToken.Error(),
		},
		{
			name:           "Invalid token returns JSON error",
			token:          "invalid.token.here",
			expectedStatus: http.StatusUnauthorized,
			expectedMsg:    ErrTokenInvalid.Error(),
		},
		{
			name:           "Malformed token returns JSON error",
			token:          "Bearer malformed",
			expectedStatus: http.StatusUnauthorized,
			expectedMsg:    ErrTokenInvalid.Error(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create handler with JWT middleware
			handler := New(secret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			// Create request
			req := httptest.NewRequest("GET", "/test", nil)
			if tt.token != "" && tt.token != "Bearer malformed" {
				req.Header.Set("Authorization", "Bearer "+tt.token)
			} else if tt.token == "Bearer malformed" {
				req.Header.Set("Authorization", tt.token)
			}

			// Record response
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			// Check status code
			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			// Check Content-Type header
			contentType := rr.Header().Get("Content-Type")
			if contentType != "application/json" {
				t.Errorf("Expected Content-Type 'application/json', got '%s'", contentType)
			}

			// Check response body is valid JSON
			var response ErrorResponse
			if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
				t.Errorf("Failed to decode JSON response: %v", err)
			}

			// Check code field
			if response.Code != tt.expectedStatus {
				t.Errorf("Expected code %d, got %d", tt.expectedStatus, response.Code)
			}

			// Check message field
			if response.Message != tt.expectedMsg {
				t.Errorf("Expected message '%s', got '%s'", tt.expectedMsg, response.Message)
			}
		})
	}
}

// TestJSONResponseWithExpiredToken tests expired token returns JSON
func TestJSONResponseWithExpiredToken(t *testing.T) {
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

	// Check status
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", rr.Code)
	}

	// Check JSON response
	var response ErrorResponse
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Errorf("Failed to decode JSON response: %v", err)
	}

	if response.Code != http.StatusUnauthorized {
		t.Errorf("Expected code 401, got %d", response.Code)
	}

	if response.Message != ErrTokenExpired.Error() {
		t.Errorf("Expected message '%s', got '%s'", ErrTokenExpired.Error(), response.Message)
	}
}

// TestJSONResponseWithWrongSigningMethod tests wrong signing method returns JSON
func TestJSONResponseWithWrongSigningMethod(t *testing.T) {
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

	// Check status
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", rr.Code)
	}

	// Check JSON response
	var response ErrorResponse
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Errorf("Failed to decode JSON response: %v", err)
	}

	if response.Code != http.StatusUnauthorized {
		t.Errorf("Expected code 401, got %d", response.Code)
	}

	if response.Message != ErrUnSupportSigningMethod.Error() {
		t.Errorf("Expected message '%s', got '%s'", ErrUnSupportSigningMethod.Error(), response.Message)
	}
}
