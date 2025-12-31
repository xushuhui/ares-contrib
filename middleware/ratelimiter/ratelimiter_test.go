package ratelimiter

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRateLimiter(t *testing.T) {
	// Create middleware with low limits for testing
	middleware := New(
		WithRate(2),   // 2 requests per second
		WithBurst(2),  // Allow burst of 2
	)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// First 2 requests should succeed (burst)
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.1:1234"
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("Request %d: Expected status 200, got %d", i+1, rr.Code)
		}
	}

	// Third request should be rate limited
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.1:1234"
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusTooManyRequests {
		t.Errorf("Expected status 429, got %d", rr.Code)
	}
}

func TestRateLimiterDifferentIPs(t *testing.T) {
	middleware := New(
		WithRate(1),
		WithBurst(1),
	)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Requests from different IPs should not affect each other
	ips := []string{"192.168.1.1:1234", "192.168.1.2:1234", "192.168.1.3:1234"}

	for _, ip := range ips {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = ip
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("Request from %s: Expected status 200, got %d", ip, rr.Code)
		}
	}
}

func TestRateLimiterWithCustomKeyFunc(t *testing.T) {
	// Use user ID from header as key
	middleware := New(
		WithRate(1),
		WithBurst(1),
		WithKeyFunc(func(r *http.Request) string {
			return r.Header.Get("X-User-ID")
		}),
	)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// First request for user1 should succeed
	req1 := httptest.NewRequest("GET", "/test", nil)
	req1.Header.Set("X-User-ID", "user1")
	rr1 := httptest.NewRecorder()
	handler.ServeHTTP(rr1, req1)

	if rr1.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr1.Code)
	}

	// Second request for user1 should be rate limited
	req2 := httptest.NewRequest("GET", "/test", nil)
	req2.Header.Set("X-User-ID", "user1")
	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, req2)

	if rr2.Code != http.StatusTooManyRequests {
		t.Errorf("Expected status 429, got %d", rr2.Code)
	}

	// Request for user2 should succeed (different key)
	req3 := httptest.NewRequest("GET", "/test", nil)
	req3.Header.Set("X-User-ID", "user2")
	rr3 := httptest.NewRecorder()
	handler.ServeHTTP(rr3, req3)

	if rr3.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr3.Code)
	}
}

func TestRateLimiterWithCustomErrorHandler(t *testing.T) {
	customErrorCalled := false

	middleware := New(
		WithRate(1),
		WithBurst(1),
		WithErrorHandler(func(w http.ResponseWriter, r *http.Request) {
			customErrorCalled = true
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte("Custom rate limit error"))
		}),
	)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// First request succeeds
	req1 := httptest.NewRequest("GET", "/test", nil)
	req1.RemoteAddr = "192.168.1.1:1234"
	rr1 := httptest.NewRecorder()
	handler.ServeHTTP(rr1, req1)

	// Second request triggers custom error handler
	req2 := httptest.NewRequest("GET", "/test", nil)
	req2.RemoteAddr = "192.168.1.1:1234"
	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, req2)

	if !customErrorCalled {
		t.Error("Expected custom error handler to be called")
	}

	if rr2.Code != http.StatusServiceUnavailable {
		t.Errorf("Expected status 503, got %d", rr2.Code)
	}

	if rr2.Body.String() != "Custom rate limit error" {
		t.Errorf("Expected custom error message, got %s", rr2.Body.String())
	}
}

func TestRateLimiterRecovery(t *testing.T) {
	middleware := New(
		WithRate(2),
		WithBurst(2),
	)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Use up the burst
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.1:1234"
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
	}

	// Next request should be rate limited
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.1:1234"
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusTooManyRequests {
		t.Errorf("Expected status 429, got %d", rr.Code)
	}

	// Wait for rate limiter to recover (500ms at 2 req/s = 1 token)
	time.Sleep(500 * time.Millisecond)

	// Should succeed now
	req2 := httptest.NewRequest("GET", "/test", nil)
	req2.RemoteAddr = "192.168.1.1:1234"
	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, req2)

	if rr2.Code != http.StatusOK {
		t.Errorf("Expected status 200 after recovery, got %d", rr2.Code)
	}
}

func TestRateLimiterXRealIP(t *testing.T) {
	middleware := New(
		WithRate(1),
		WithBurst(1),
	)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// First request with X-Real-IP
	req1 := httptest.NewRequest("GET", "/test", nil)
	req1.Header.Set("X-Real-IP", "10.0.0.1")
	rr1 := httptest.NewRecorder()
	handler.ServeHTTP(rr1, req1)

	if rr1.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr1.Code)
	}

	// Second request with same X-Real-IP should be rate limited
	req2 := httptest.NewRequest("GET", "/test", nil)
	req2.Header.Set("X-Real-IP", "10.0.0.1")
	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, req2)

	if rr2.Code != http.StatusTooManyRequests {
		t.Errorf("Expected status 429, got %d", rr2.Code)
	}
}
