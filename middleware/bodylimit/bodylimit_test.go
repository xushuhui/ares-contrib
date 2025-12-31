package bodylimit

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestBodyLimit(t *testing.T) {
	limit := int64(100) // 100 bytes
	middleware := New(limit)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Write(body)
	}))

	// Test with body under limit
	t.Run("Under limit", func(t *testing.T) {
		body := strings.Repeat("a", 50) // 50 bytes
		req := httptest.NewRequest("POST", "/test", strings.NewReader(body))
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", rr.Code)
		}

		if rr.Body.String() != body {
			t.Error("Body content mismatch")
		}
	})

	// Test with body at limit
	t.Run("At limit", func(t *testing.T) {
		body := strings.Repeat("a", 100) // 100 bytes
		req := httptest.NewRequest("POST", "/test", strings.NewReader(body))
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", rr.Code)
		}
	})

	// Test with body over limit
	t.Run("Over limit", func(t *testing.T) {
		body := strings.Repeat("a", 150) // 150 bytes
		req := httptest.NewRequest("POST", "/test", strings.NewReader(body))
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", rr.Code)
		}
	})
}

func TestBodyLimitLargeBody(t *testing.T) {
	limit := int64(1024) // 1KB
	middleware := New(limit)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))

	// 10KB body (over limit)
	body := bytes.Repeat([]byte("a"), 10*1024)
	req := httptest.NewRequest("POST", "/test", bytes.NewReader(body))
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400 for large body, got %d", rr.Code)
	}
}

func TestBodyLimitEmptyBody(t *testing.T) {
	limit := int64(100)
	middleware := New(limit)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if len(body) != 0 {
			t.Error("Expected empty body")
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("POST", "/test", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200 for empty body, got %d", rr.Code)
	}
}

func TestBodyLimitDifferentLimits(t *testing.T) {
	tests := []struct {
		name       string
		limit      int64
		bodySize   int
		expectCode int
	}{
		{"1KB limit, 500B body", 1024, 500, http.StatusOK},
		{"1KB limit, 1KB body", 1024, 1024, http.StatusOK},
		{"1KB limit, 2KB body", 1024, 2048, http.StatusBadRequest},
		{"10MB limit, 5MB body", 10 * 1024 * 1024, 5 * 1024 * 1024, http.StatusOK},
		{"10MB limit, 15MB body", 10 * 1024 * 1024, 15 * 1024 * 1024, http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middleware := New(tt.limit)

			handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_, err := io.ReadAll(r.Body)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				w.WriteHeader(http.StatusOK)
			}))

			body := bytes.Repeat([]byte("a"), tt.bodySize)
			req := httptest.NewRequest("POST", "/test", bytes.NewReader(body))
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			if rr.Code != tt.expectCode {
				t.Errorf("Expected status %d, got %d", tt.expectCode, rr.Code)
			}
		})
	}
}

func TestBodyLimitPanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic for zero limit")
		}
	}()

	New(0)
}

func TestBodyLimitNegativePanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic for negative limit")
		}
	}()

	New(-1)
}

func TestBodyLimitGETRequest(t *testing.T) {
	limit := int64(100)
	middleware := New(limit)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200 for GET request, got %d", rr.Code)
	}
}

func TestBodyLimitMultipleReads(t *testing.T) {
	limit := int64(100)
	middleware := New(limit)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// First read
		buf1 := make([]byte, 50)
		n1, _ := r.Body.Read(buf1)

		// Second read
		buf2 := make([]byte, 50)
		n2, _ := r.Body.Read(buf2)

		total := n1 + n2
		if total > 100 {
			t.Errorf("Read more than limit: %d bytes", total)
		}

		w.WriteHeader(http.StatusOK)
	}))

	body := strings.Repeat("a", 150) // Over limit
	req := httptest.NewRequest("POST", "/test", strings.NewReader(body))
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	// Should still work, but limited
	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}
}
