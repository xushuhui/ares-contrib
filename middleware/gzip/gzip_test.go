package gzip

import (
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGzip(t *testing.T) {
	middleware := New()

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(strings.Repeat("test data ", 200))) // > 1KB
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}

	if rr.Header().Get("Content-Encoding") != "gzip" {
		t.Error("Expected Content-Encoding: gzip")
	}

	// Decompress and verify
	gr, err := gzip.NewReader(rr.Body)
	if err != nil {
		t.Fatalf("Failed to create gzip reader: %v", err)
	}
	defer gr.Close()

	decompressed, err := io.ReadAll(gr)
	if err != nil {
		t.Fatalf("Failed to decompress: %v", err)
	}

	expected := strings.Repeat("test data ", 200)
	if string(decompressed) != expected {
		t.Error("Decompressed content doesn't match original")
	}
}

func TestGzipNoAcceptEncoding(t *testing.T) {
	middleware := New()

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("test data"))
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	// No Accept-Encoding header
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Header().Get("Content-Encoding") == "gzip" {
		t.Error("Should not compress when client doesn't accept gzip")
	}

	if rr.Body.String() != "test data" {
		t.Error("Content should not be compressed")
	}
}

func TestGzipMinLength(t *testing.T) {
	middleware := New(WithMinLength(1000))

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("small data")) // < 1000 bytes
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Header().Get("Content-Encoding") == "gzip" {
		t.Error("Should not compress data smaller than min length")
	}
}

func TestGzipExcludedExtensions(t *testing.T) {
	middleware := New()

	tests := []struct {
		path           string
		shouldCompress bool
	}{
		{"/test.html", true},
		{"/test.png", false},
		{"/test.jpg", false},
		{"/test.zip", false},
		{"/test.mp4", false},
		{"/api/data", true},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte(strings.Repeat("test ", 300))) // > 1KB
			}))

			req := httptest.NewRequest("GET", tt.path, nil)
			req.Header.Set("Accept-Encoding", "gzip")
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			hasGzip := rr.Header().Get("Content-Encoding") == "gzip"
			if hasGzip != tt.shouldCompress {
				t.Errorf("Path %s: expected compress=%v, got compress=%v", tt.path, tt.shouldCompress, hasGzip)
			}
		})
	}
}

func TestGzipExcludedPaths(t *testing.T) {
	middleware := New(WithExcludedPaths([]string{"/api/stream", "/ws"}))

	tests := []struct {
		path           string
		shouldCompress bool
	}{
		{"/api/data", true},
		{"/api/stream", false},
		{"/api/stream/video", false},
		{"/ws", false},
		{"/ws/connect", false},
		{"/other", true},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte(strings.Repeat("test ", 300))) // > 1KB
			}))

			req := httptest.NewRequest("GET", tt.path, nil)
			req.Header.Set("Accept-Encoding", "gzip")
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			hasGzip := rr.Header().Get("Content-Encoding") == "gzip"
			if hasGzip != tt.shouldCompress {
				t.Errorf("Path %s: expected compress=%v, got compress=%v", tt.path, tt.shouldCompress, hasGzip)
			}
		})
	}
}

func TestGzipLevel(t *testing.T) {
	tests := []struct {
		level int
		name  string
	}{
		{gzip.BestSpeed, "BestSpeed"},
		{gzip.BestCompression, "BestCompression"},
		{gzip.DefaultCompression, "DefaultCompression"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middleware := New(WithLevel(tt.level))

			handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte(strings.Repeat("test data ", 200)))
			}))

			req := httptest.NewRequest("GET", "/test", nil)
			req.Header.Set("Accept-Encoding", "gzip")
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			if rr.Header().Get("Content-Encoding") != "gzip" {
				t.Error("Expected gzip compression")
			}

			// Verify it's valid gzip
			_, err := gzip.NewReader(rr.Body)
			if err != nil {
				t.Errorf("Failed to create gzip reader: %v", err)
			}
		})
	}
}

func TestGzipVaryHeader(t *testing.T) {
	middleware := New()

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(strings.Repeat("test ", 300)))
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if !strings.Contains(rr.Header().Get("Vary"), "Accept-Encoding") {
		t.Error("Expected Vary header to include Accept-Encoding")
	}
}

func TestGzipNoContentLength(t *testing.T) {
	middleware := New()

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "1000")
		w.Write([]byte(strings.Repeat("test ", 300)))
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Header().Get("Content-Length") != "" {
		t.Error("Content-Length should be removed when using gzip")
	}
}

// TestGzipWithCustomExcludedExtensions tests custom excluded extensions
func TestGzipWithCustomExcludedExtensions(t *testing.T) {
	middleware := New(WithExcludedExtensions([]string{".html", ".txt"}))

	tests := []struct {
		path           string
		shouldCompress bool
	}{
		{"/test.html", false},
		{"/test.txt", false},
		{"/test.json", true},
		{"/api/data", true},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte(strings.Repeat("test ", 300)))
			}))

			req := httptest.NewRequest("GET", tt.path, nil)
			req.Header.Set("Accept-Encoding", "gzip")
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			hasGzip := rr.Header().Get("Content-Encoding") == "gzip"
			if hasGzip != tt.shouldCompress {
				t.Errorf("Path %s: expected compress=%v, got compress=%v", tt.path, tt.shouldCompress, hasGzip)
			}
		})
	}
}

// TestGzipSmallResponse tests that small responses are not compressed
func TestGzipSmallResponse(t *testing.T) {
	middleware := New()

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("small")) // < default min length (1024)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Header().Get("Content-Encoding") == "gzip" {
		t.Error("Small responses should not be compressed")
	}

	if rr.Body.String() != "small" {
		t.Errorf("Expected 'small', got '%s'", rr.Body.String())
	}
}

// TestGzipContentTypeTests tests compression based on content type
func TestGzipContentTypeTests(t *testing.T) {
	middleware := New()

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(strings.Repeat("test ", 300)))
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	// HTML should be compressed by default
	if rr.Header().Get("Content-Encoding") != "gzip" {
		t.Error("HTML content should be compressed")
	}
}

// TestGzipHEADRequest tests that HEAD requests are handled correctly
func TestGzipHEADRequest(t *testing.T) {
	middleware := New()

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		// HEAD requests should not write body
	}))

	req := httptest.NewRequest("HEAD", "/test", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	// HEAD request should not have body but should be handled correctly
	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}

	// HEAD request with no body won't have Content-Encoding header
	// because the gzip writer is only created when content is written
}

// TestGzipMultipleAcceptEncoding tests various Accept-Encoding formats
func TestGzipMultipleAcceptEncoding(t *testing.T) {
	tests := []struct {
		acceptEncoding  string
		shouldCompress  bool
	}{
		{"gzip", true},
		{"gzip, deflate", true},
		{"deflate, gzip", true},
		{"deflate", false},
		{"*", false},
		{"identity", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.acceptEncoding, func(t *testing.T) {
			middleware := New()

			handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte(strings.Repeat("test ", 300)))
			}))

			req := httptest.NewRequest("GET", "/test", nil)
			if tt.acceptEncoding != "" {
				req.Header.Set("Accept-Encoding", tt.acceptEncoding)
			}
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			hasGzip := rr.Header().Get("Content-Encoding") == "gzip"
			if hasGzip != tt.shouldCompress {
				t.Errorf("Accept-Encoding '%s': expected compress=%v, got compress=%v",
					tt.acceptEncoding, tt.shouldCompress, hasGzip)
			}
		})
	}
}

// TestGzipWriterPool tests that writers are properly pooled
func TestGzipWriterPool(t *testing.T) {
	middleware := New()

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(strings.Repeat("test data ", 200)))
	}))

	// Make multiple requests to test pooling
	for i := 0; i < 10; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Accept-Encoding", "gzip")
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		if rr.Header().Get("Content-Encoding") != "gzip" {
			t.Errorf("Request %d: Expected gzip encoding", i)
		}
	}
}
