package bodylimit

import (
	"net/http"
)

// Option is body limit option.
type Option func(*options)

// options defines the configuration for BodyLimit middleware
type options struct {
	// Limit is the maximum allowed size for a request body in bytes
	limit int64
}

// WithLimit sets the body size limit
func WithLimit(limit int64) Option {
	return func(o *options) {
		o.limit = limit
	}
}

// New returns a BodyLimit middleware with the specified limit
func New(limit int64, opts ...Option) func(http.Handler) http.Handler {
	o := &options{
		limit: limit,
	}

	for _, opt := range opts {
		opt(o)
	}

	if o.limit <= 0 {
		panic("body limit must be greater than 0")
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Limit request body size
			r.Body = http.MaxBytesReader(w, r.Body, o.limit)

			next.ServeHTTP(w, r)
		})
	}
}
