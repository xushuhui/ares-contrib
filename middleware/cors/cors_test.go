package cors

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCORS(t *testing.T) {
	middleware := New()

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Error("Expected Access-Control-Allow-Origin: *")
	}

	if rr.Header().Get("Access-Control-Allow-Methods") == "" {
		t.Error("Expected Access-Control-Allow-Methods header")
	}

	if rr.Header().Get("Access-Control-Allow-Headers") == "" {
		t.Error("Expected Access-Control-Allow-Headers header")
	}
}

func TestCORSPreflight(t *testing.T) {
	middleware := New()

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Handler should not be called for OPTIONS request")
	}))

	req := httptest.NewRequest("OPTIONS", "/test", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Errorf("Expected status 204 for OPTIONS, got %d", rr.Code)
	}
}

func TestCORSWithAllowedOrigins(t *testing.T) {
	middleware := New(WithAllowedOrigins([]string{"https://example.com", "https://test.com"}))

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Test with allowed origin
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "https://example.com")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	origin := rr.Header().Get("Access-Control-Allow-Origin")
	if origin != "https://example.com" {
		t.Errorf("Expected 'https://example.com', got '%s'", origin)
	}

	// Should have Vary header for non-wildcard origins
	if rr.Header().Get("Vary") == "" {
		t.Error("Expected Vary header for specific origins")
	}

	// Test with disallowed origin
	req2 := httptest.NewRequest("GET", "/test", nil)
	req2.Header.Set("Origin", "https://malicious.com")
	rr2 := httptest.NewRecorder()

	handler.ServeHTTP(rr2, req2)

	origin2 := rr2.Header().Get("Access-Control-Allow-Origin")
	if origin2 != "" {
		t.Errorf("Expected no origin header for disallowed origin, got '%s'", origin2)
	}
}

func TestCORSWithAllowedMethods(t *testing.T) {
	middleware := New(WithAllowedMethods([]string{"GET", "POST"}))

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	methods := rr.Header().Get("Access-Control-Allow-Methods")
	if methods != "GET, POST" {
		t.Errorf("Expected 'GET, POST', got %s", methods)
	}
}

func TestCORSWithAllowedHeaders(t *testing.T) {
	middleware := New(WithAllowedHeaders([]string{"Authorization", "Content-Type"}))

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	headers := rr.Header().Get("Access-Control-Allow-Headers")
	if headers != "Authorization, Content-Type" {
		t.Errorf("Expected 'Authorization, Content-Type', got %s", headers)
	}
}

func TestCORSWithExposedHeaders(t *testing.T) {
	middleware := New(WithExposedHeaders([]string{"X-Custom-Header", "X-Another-Header"}))

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	exposed := rr.Header().Get("Access-Control-Expose-Headers")
	if exposed != "X-Custom-Header, X-Another-Header" {
		t.Errorf("Expected exposed headers, got %s", exposed)
	}
}

func TestCORSWithAllowCredentials(t *testing.T) {
	middleware := New(
		WithAllowedOrigins([]string{"https://example.com"}), // Need specific origin for credentials
		WithAllowCredentials(true),
	)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "https://example.com") // Set the allowed origin
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Header().Get("Access-Control-Allow-Credentials") != "true" {
		t.Error("Expected Access-Control-Allow-Credentials: true")
	}

	// Should also set the correct origin
	if rr.Header().Get("Access-Control-Allow-Origin") != "https://example.com" {
		t.Error("Expected specific origin when credentials are enabled")
	}

	// Test that credentials are NOT set for wildcard origin
	middleware2 := New(WithAllowCredentials(true)) // Uses default wildcard origin
	handler2 := middleware2(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req2 := httptest.NewRequest("GET", "/test", nil)
	rr2 := httptest.NewRecorder()
	handler2.ServeHTTP(rr2, req2)

	if rr2.Header().Get("Access-Control-Allow-Credentials") == "true" {
		t.Error("Credentials should not be set with wildcard origin")
	}
}

func TestCORSWithMaxAge(t *testing.T) {
	middleware := New(WithMaxAge(7200))

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	maxAge := rr.Header().Get("Access-Control-Max-Age")
	if maxAge == "" {
		t.Error("Expected Access-Control-Max-Age header")
	}
}

func TestCORSMultipleOptions(t *testing.T) {
	middleware := New(
		WithAllowedOrigins([]string{"https://example.com"}),
		WithAllowedMethods([]string{"GET", "POST", "PUT"}),
		WithAllowedHeaders([]string{"Authorization"}),
		WithAllowCredentials(true),
		WithMaxAge(3600),
	)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "https://example.com") // Set the allowed origin
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Header().Get("Access-Control-Allow-Origin") != "https://example.com" {
		t.Errorf("Origin not set correctly, got '%s'", rr.Header().Get("Access-Control-Allow-Origin"))
	}

	if rr.Header().Get("Access-Control-Allow-Methods") != "GET, POST, PUT" {
		t.Errorf("Methods not set correctly, got '%s'", rr.Header().Get("Access-Control-Allow-Methods"))
	}

	if rr.Header().Get("Access-Control-Allow-Credentials") != "true" {
		t.Errorf("Credentials not set correctly, got '%s'", rr.Header().Get("Access-Control-Allow-Credentials"))
	}

	if rr.Header().Get("Access-Control-Allow-Headers") != "Authorization" {
		t.Errorf("Headers not set correctly, got '%s'", rr.Header().Get("Access-Control-Allow-Headers"))
	}

	if rr.Header().Get("Access-Control-Max-Age") != "3600" {
		t.Errorf("Max-Age not set correctly, got '%s'", rr.Header().Get("Access-Control-Max-Age"))
	}

	if rr.Header().Get("Vary") == "" {
		t.Error("Vary header should be set for non-wildcard origins")
	}
}
