package cors

import (
	"net/http"
	"strconv"
	"strings"
)

// CORSOption is CORS option.
type Option func(*options)

// options defines the configuration for CORS middleware
type options struct {
	// AllowedOrigins is a list of origins a cross-domain request can be executed from
	// Default value is ["*"]
	allowedOrigins []string

	// AllowedMethods is a list of methods the client is allowed to use with cross-domain requests
	// Default value is ["GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"]
	allowedMethods []string

	// AllowedHeaders is list of non simple headers the client is allowed to use with cross-domain requests
	// Default value is []
	allowedHeaders []string

	// ExposedHeaders indicates which headers are safe to expose to the API of a CORS API specification
	// Default value is []
	exposedHeaders []string

	// AllowCredentials indicates whether the request can include user credentials
	// Default value is false
	allowCredentials bool

	// MaxAge indicates how long (in seconds) the results of a preflight request can be cached
	// Default value is 0
	maxAge int
}

// WithAllowedOrigins sets the allowed origins
func WithAllowedOrigins(origins []string) Option {
	return func(o *options) {
		o.allowedOrigins = origins
	}
}

// WithAllowedMethods sets the allowed methods
func WithAllowedMethods(methods []string) Option {
	return func(o *options) {
		o.allowedMethods = methods
	}
}

// WithAllowedHeaders sets the allowed headers
func WithAllowedHeaders(headers []string) Option {
	return func(o *options) {
		o.allowedHeaders = headers
	}
}

// WithExposedHeaders sets the exposed headers
func WithExposedHeaders(headers []string) Option {
	return func(o *options) {
		o.exposedHeaders = headers
	}
}

// WithAllowCredentials sets whether credentials are allowed
func WithAllowCredentials(allow bool) Option {
	return func(o *options) {
		o.allowCredentials = allow
	}
}

// WithMaxAge sets the max age for preflight requests
func WithMaxAge(age int) Option {
	return func(o *options) {
		o.maxAge = age
	}
}

// isOriginAllowed checks if the given origin is in the allowed list
func isOriginAllowed(origin string, allowedOrigins []string) bool {
	for _, allowed := range allowedOrigins {
		if allowed == "*" || allowed == origin {
			return true
		}
	}
	return false
}

// CORS returns a CORS middleware with optional configuration
func New(opts ...Option) func(http.Handler) http.Handler {
	o := &options{
		allowedOrigins: []string{"*"},
		allowedMethods: []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"},
		allowedHeaders: []string{"Accept", "Content-Type", "Content-Length", "Accept-Encoding", "Authorization"},
		maxAge:         3600,
	}

	for _, opt := range opts {
		opt(o)
	}

	allowedMethods := strings.Join(o.allowedMethods, ", ")
	allowedHeaders := strings.Join(o.allowedHeaders, ", ")
	exposedHeaders := strings.Join(o.exposedHeaders, ", ")

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// Determine allowed origin
			var allowedOrigin string
			if len(o.allowedOrigins) == 1 && o.allowedOrigins[0] == "*" {
				allowedOrigin = "*"
			} else if isOriginAllowed(origin, o.allowedOrigins) {
				allowedOrigin = origin
			} else {
				// Origin not allowed, still set other headers but not Access-Control-Allow-Origin
				w.Header().Set("Access-Control-Allow-Methods", allowedMethods)
				w.Header().Set("Access-Control-Allow-Headers", allowedHeaders)

				// Handle preflight requests
				if r.Method == http.MethodOptions {
					w.WriteHeader(http.StatusNoContent)
					return
				}

				next.ServeHTTP(w, r)
				return
			}

			// Set CORS headers
			w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
			w.Header().Set("Access-Control-Allow-Methods", allowedMethods)
			w.Header().Set("Access-Control-Allow-Headers", allowedHeaders)

			// Only add Vary header when not using wildcard
			if allowedOrigin != "*" {
				w.Header().Add("Vary", "Origin")
			}

			if len(exposedHeaders) > 0 {
				w.Header().Set("Access-Control-Expose-Headers", exposedHeaders)
			}

			// Only set credentials header if origin is not wildcard
			if o.allowCredentials && allowedOrigin != "*" {
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}

			if o.maxAge > 0 {
				w.Header().Set("Access-Control-Max-Age", strconv.Itoa(o.maxAge))
			}

			// Handle preflight requests
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
