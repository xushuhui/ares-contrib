package requestid

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequestID(t *testing.T) {
	middleware := New()

	var capturedID string
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Context().Value(contextKey("requestID"))
		if id != nil {
			capturedID = id.(string)
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	// Check response header
	responseID := rr.Header().Get("X-Request-ID")
	if responseID == "" {
		t.Error("Expected X-Request-ID header in response")
	}

	// Check context value
	if capturedID == "" {
		t.Error("Expected request ID in context")
	}

	if responseID != capturedID {
		t.Error("Response header and context ID should match")
	}
}

func TestRequestIDReuseExisting(t *testing.T) {
	middleware := New()

	existingID := "existing-request-id-123"

	var capturedID string
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Context().Value(contextKey("requestID"))
		if id != nil {
			capturedID = id.(string)
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Request-ID", existingID)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	responseID := rr.Header().Get("X-Request-ID")
	if responseID != existingID {
		t.Errorf("Expected to reuse existing ID %s, got %s", existingID, responseID)
	}

	if capturedID != existingID {
		t.Errorf("Expected context ID %s, got %s", existingID, capturedID)
	}
}

func TestRequestIDWithCustomGenerator(t *testing.T) {
	customID := "custom-id-12345"
	middleware := New(WithGenerator(func() string {
		return customID
	}))

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	responseID := rr.Header().Get("X-Request-ID")
	if responseID != customID {
		t.Errorf("Expected custom ID %s, got %s", customID, responseID)
	}
}

func TestRequestIDWithCustomHeader(t *testing.T) {
	middleware := New(WithRequestIDHeader("X-Trace-ID"))

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Header().Get("X-Trace-ID") == "" {
		t.Error("Expected X-Trace-ID header")
	}

	if rr.Header().Get("X-Request-ID") != "" {
		t.Error("Should not have X-Request-ID header")
	}
}

func TestRequestIDWithCustomContextKey(t *testing.T) {
	middleware := New(WithRequestIDContextKey("traceID"))

	var capturedID string
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Context().Value(contextKey("traceID"))
		if id != nil {
			capturedID = id.(string)
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if capturedID == "" {
		t.Error("Expected ID in context with custom key")
	}
}

func TestRequestIDMultipleRequests(t *testing.T) {
	middleware := New()

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	ids := make(map[string]bool)

	// Make multiple requests
	for i := 0; i < 10; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		id := rr.Header().Get("X-Request-ID")
		if id == "" {
			t.Error("Expected request ID")
		}

		if ids[id] {
			t.Errorf("Duplicate request ID: %s", id)
		}
		ids[id] = true
	}

	if len(ids) != 10 {
		t.Errorf("Expected 10 unique IDs, got %d", len(ids))
	}
}
