package gzip

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
	"sync"
)

// GzipOption is gzip option.
type Option func(*options)

// options defines the configuration for gzip middleware
type options struct {
	// Level is the gzip compression level
	// Valid values: -1 (default), 0 (no compression), 1 (best speed) to 9 (best compression)
	level int

	// MinLength is the minimum response size to compress
	// Responses smaller than this will not be compressed
	minLength int

	// ExcludedExtensions is a list of file extensions to exclude from compression
	excludedExtensions []string

	// ExcludedPaths is a list of URL paths to exclude from compression
	excludedPaths []string
}

// WithLevel sets the compression level
func WithLevel(level int) Option {
	return func(o *options) {
		o.level = level
	}
}

// WithMinLength sets the minimum response size to compress
func WithMinLength(length int) Option {
	return func(o *options) {
		o.minLength = length
	}
}

// WithExcludedExtensions sets the file extensions to exclude
func WithExcludedExtensions(extensions []string) Option {
	return func(o *options) {
		o.excludedExtensions = extensions
	}
}

// WithExcludedPaths sets the URL paths to exclude
func WithExcludedPaths(paths []string) Option {
	return func(o *options) {
		o.excludedPaths = paths
	}
}

// gzipResponseWriter wraps http.ResponseWriter to compress response
type gzipResponseWriter struct {
	http.ResponseWriter
	writer         *gzip.Writer
	wroteHeader    bool
	headersSent    bool
	minLength      int
	buffer         []byte
	shouldCompress *bool  // Use pointer to track uninitialized state
}

// gzipWriterPool is a pool of gzip writers
var gzipWriterPool = sync.Pool{
	New: func() interface{} {
		w, _ := gzip.NewWriterLevel(io.Discard, gzip.DefaultCompression)
		return w
	},
}

// newGzipResponseWriter creates a new gzip response writer
func newGzipResponseWriter(w http.ResponseWriter, level, minLength int) *gzipResponseWriter {
	gw := gzipWriterPool.Get().(*gzip.Writer)
	gw.Reset(w)

	return &gzipResponseWriter{
		ResponseWriter: w,
		writer:         gw,
		minLength:      minLength,
		buffer:         make([]byte, 0, minLength),
		shouldCompress: nil,  // Uninitialized - will decide later
	}
}

// WriteHeader implements http.ResponseWriter
func (w *gzipResponseWriter) WriteHeader(code int) {
	if w.wroteHeader {
		return
	}
	w.wroteHeader = true

	// Don't compress if status code indicates no body
	if code == http.StatusNoContent || code == http.StatusNotModified {
		compress := false
		w.shouldCompress = &compress
	}

	// If compression decision is not made yet, decide based on buffered content
	if w.shouldCompress == nil {
		compress := len(w.buffer) >= w.minLength
		w.shouldCompress = &compress
	}

	// Set Content-Encoding header if compressing
	if *w.shouldCompress {
		w.Header().Set("Content-Encoding", "gzip")
		w.Header().Del("Content-Length")
		w.Header().Add("Vary", "Accept-Encoding")
	}

	w.headersSent = true
	w.ResponseWriter.WriteHeader(code)
}

// Write implements http.ResponseWriter
func (w *gzipResponseWriter) Write(b []byte) (int, error) {
	// If headers haven't been sent yet, decide on compression
	if !w.headersSent {
		// Buffer data until we can make a decision or reach minimum length
		if w.shouldCompress == nil && len(w.buffer)+len(b) < w.minLength {
			w.buffer = append(w.buffer, b...)
			return len(b), nil
		}

		// Make compression decision if not already made
		if w.shouldCompress == nil {
			totalLength := len(w.buffer) + len(b)
			compress := totalLength >= w.minLength
			w.shouldCompress = &compress
		}

		// Write headers
		if !w.wroteHeader {
			w.WriteHeader(http.StatusOK)
		}
	}

	// If not compressing, write directly
	if !*w.shouldCompress {
		// Flush buffer first if we have any
		if len(w.buffer) > 0 {
			if _, err := w.ResponseWriter.Write(w.buffer); err != nil {
				return 0, err
			}
			w.buffer = nil
		}
		return w.ResponseWriter.Write(b)
	}

	// Compressing - flush buffer through gzip writer first
	if len(w.buffer) > 0 {
		if _, err := w.writer.Write(w.buffer); err != nil {
			return 0, err
		}
		w.buffer = nil
	}

	return w.writer.Write(b)
}

// Close closes the gzip writer and returns it to the pool
func (w *gzipResponseWriter) Close() error {
	// If we still have buffered data and no decision was made, make one now
	if w.shouldCompress == nil && len(w.buffer) > 0 {
		compress := len(w.buffer) >= w.minLength
		w.shouldCompress = &compress

		// Write headers if not already written
		if !w.wroteHeader {
			w.WriteHeader(http.StatusOK)
		}
	}

	// Write any remaining buffered data
	if len(w.buffer) > 0 {
		if *w.shouldCompress {
			// Write through gzip writer
			if _, err := w.writer.Write(w.buffer); err != nil {
				return err
			}
		} else {
			// Write directly to response writer
			if _, err := w.ResponseWriter.Write(w.buffer); err != nil {
				return err
			}
		}
		w.buffer = nil
	}

	// Close gzip writer only if we used compression
	if w.shouldCompress != nil && *w.shouldCompress {
		if err := w.writer.Close(); err != nil {
			return err
		}
	}

	gzipWriterPool.Put(w.writer)
	return nil
}

// Gzip returns a gzip middleware with optional configuration
func New(opts ...Option) func(http.Handler) http.Handler {
	o := &options{
		level:     gzip.DefaultCompression,
		minLength: 1024, // 1KB
		excludedExtensions: []string{
			".png", ".jpg", ".jpeg", ".gif", ".webp", ".svg",
			".zip", ".gz", ".tar", ".rar", ".7z",
			".mp4", ".avi", ".mov", ".mp3", ".wav",
			".pdf",
		},
	}

	for _, opt := range opts {
		opt(o)
	}

	// Validate level
	if o.level < gzip.HuffmanOnly || o.level > gzip.BestCompression {
		o.level = gzip.DefaultCompression
	}
	if o.minLength <= 0 {
		o.minLength = 1024
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check if client accepts gzip
			if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
				next.ServeHTTP(w, r)
				return
			}

			// Check if path is excluded
			for _, path := range o.excludedPaths {
				if strings.HasPrefix(r.URL.Path, path) {
					next.ServeHTTP(w, r)
					return
				}
			}

			// Check if extension is excluded
			for _, ext := range o.excludedExtensions {
				if strings.HasSuffix(r.URL.Path, ext) {
					next.ServeHTTP(w, r)
					return
				}
			}

			// Create gzip response writer
			gzw := newGzipResponseWriter(w, o.level, o.minLength)
			defer gzw.Close()

			next.ServeHTTP(gzw, r)
		})
	}
}
