package requestid

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

// RequestIDOption is request ID option.
type Option func(*options)

// options defines the configuration for RequestID middleware
type options struct {
	// Generator is a function to generate request ID
	// Default: UUID v4
	generator func() string

	// RequestIDHeader is the header name for request ID
	// Default: X-Request-ID
	requestIDHeader string

	// ContextKey is the key used to store request ID in context
	// Default: requestID
	contextKey string
}

// WithGenerator sets the ID generator function
func WithGenerator(f func() string) Option {
	return func(o *options) {
		o.generator = f
	}
}

// WithRequestIDHeader sets the request ID header name
func WithRequestIDHeader(header string) Option {
	return func(o *options) {
		o.requestIDHeader = header
	}
}

// WithRequestIDContextKey sets the context key for storing request ID
func WithRequestIDContextKey(key string) Option {
	return func(o *options) {
		o.contextKey = key
	}
}

// RequestID returns a RequestID middleware with optional configuration
func New(opts ...Option) func(http.Handler) http.Handler {
	o := &options{
		generator: func() string {
			return uuid.New().String()
		},
		requestIDHeader: "X-Request-ID",
		contextKey:      "requestID",
	}

	for _, opt := range opts {
		opt(o)
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check if request ID already exists
			requestID := r.Header.Get(o.requestIDHeader)
			if requestID == "" {
				requestID = o.generator()
			}

			// Set request ID in response header
			w.Header().Set(o.requestIDHeader, requestID)

			// Store request ID in context
			ctx := context.WithValue(r.Context(), contextKey(o.contextKey), requestID)
			r = r.WithContext(ctx)

			next.ServeHTTP(w, r)
		})
	}
}

// contextKey is the type used for context keys
type contextKey string
