package ratelimiter

import (
	"context"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// Option is rate limiter option.
type Option func(*options)

// options defines the configuration for rate limiter middleware
type options struct {
	// Rate is the number of requests allowed per second
	rate float64

	// Burst is the maximum number of requests allowed in a burst
	burst int

	// KeyFunc is a function to extract the key for rate limiting
	// Default: uses IP address
	keyFunc func(*http.Request) string

	// ErrorHandler defines a function which is executed when rate limit is exceeded
	// Optional. Default value returns 429 Too Many Requests
	errorHandler func(http.ResponseWriter, *http.Request)
}

// WithRate sets the rate limit (requests per second)
func WithRate(r float64) Option {
	return func(o *options) {
		o.rate = r
	}
}

// WithBurst sets the burst size
func WithBurst(b int) Option {
	return func(o *options) {
		o.burst = b
	}
}

// WithKeyFunc sets the key extraction function
func WithKeyFunc(f func(*http.Request) string) Option {
	return func(o *options) {
		o.keyFunc = f
	}
}

// WithErrorHandler sets the error handler
func WithErrorHandler(h func(http.ResponseWriter, *http.Request)) Option {
	return func(o *options) {
		o.errorHandler = h
	}
}

// limiterEntry holds a rate limiter with its last access time
type limiterEntry struct {
	limiter    *rate.Limiter
	lastAccess time.Time
}

// rateLimiter holds the rate limiters for each key
type rateLimiter struct {
	limiters      map[string]*limiterEntry
	mu            sync.RWMutex
	rate          rate.Limit
	burst         int
	cleanupCancel context.CancelFunc
	cleanupDone   chan struct{}
}

// newRateLimiter creates a new rate limiter
func newRateLimiter(r float64, burst int) *rateLimiter {
	return &rateLimiter{
		limiters:    make(map[string]*limiterEntry),
		rate:        rate.Limit(r),
		burst:       burst,
		cleanupDone: make(chan struct{}),
	}
}

// getLimiter returns the rate limiter for the given key
func (rl *rateLimiter) getLimiter(key string) *rate.Limiter {
	now := time.Now()

	rl.mu.RLock()
	entry, exists := rl.limiters[key]
	rl.mu.RUnlock()

	if exists {
		// Update last access time
		rl.mu.Lock()
		entry.lastAccess = now
		rl.mu.Unlock()
		return entry.limiter
	}

	rl.mu.Lock()
	// Double-check after acquiring write lock
	entry, exists = rl.limiters[key]
	if !exists {
		entry = &limiterEntry{
			limiter:    rate.NewLimiter(rl.rate, rl.burst),
			lastAccess: now,
		}
		rl.limiters[key] = entry
	} else {
		entry.lastAccess = now
	}
	rl.mu.Unlock()

	return entry.limiter
}

// cleanup removes old limiters periodically
func (rl *rateLimiter) cleanup(interval time.Duration, maxAge time.Duration) {
	ctx, cancel := context.WithCancel(context.Background())
	rl.cleanupCancel = cancel

	ticker := time.NewTicker(interval)
	go func() {
		defer ticker.Stop()
		defer close(rl.cleanupDone)

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				rl.mu.Lock()
				now := time.Now()
				// Remove limiters that haven't been accessed recently
				for key, entry := range rl.limiters {
					if now.Sub(entry.lastAccess) > maxAge {
						delete(rl.limiters, key)
					}
				}
				rl.mu.Unlock()
			}
		}
	}()
}

// Stop stops the cleanup goroutine and cleans up resources
func (rl *rateLimiter) Stop() {
	if rl.cleanupCancel != nil {
		rl.cleanupCancel()
		<-rl.cleanupDone // Wait for cleanup to finish
	}
}

// extractIP safely extracts the real IP address from the request
func extractIP(r *http.Request) string {
	// First try RemoteAddr as it's most reliable
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil && net.ParseIP(ip) != nil {
		return ip
	}

	// Only use proxy headers if RemoteAddr fails and they contain valid IPs
	// Check X-Forwarded-For (can contain multiple IPs, use first valid one)
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		ips := strings.Split(forwarded, ",")
		for _, ip := range ips {
			ip = strings.TrimSpace(ip)
			if parsedIP := net.ParseIP(ip); parsedIP != nil && !parsedIP.IsLoopback() {
				return ip
			}
		}
	}

	// Check X-Real-IP as fallback
	if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
		if parsedIP := net.ParseIP(realIP); parsedIP != nil && !parsedIP.IsLoopback() {
			return realIP
		}
	}

	// Fallback to RemoteAddr without validation
	return r.RemoteAddr
}

// New returns a rate limiter middleware with optional configuration
func New(opts ...Option) func(http.Handler) http.Handler {
	o := &options{
		rate:  10,  // 10 requests per second
		burst: 20,  // Allow burst of 20 requests
		keyFunc: extractIP, // Use secure IP extraction
	}

	for _, opt := range opts {
		opt(o)
	}

	limiter := newRateLimiter(o.rate, o.burst)

	// Start cleanup goroutine to remove old limiters
	// Clean up limiters that haven't been used for 10 minutes every 5 minutes
	limiter.cleanup(5*time.Minute, 10*time.Minute)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get key for rate limiting
			key := o.keyFunc(r)

			// Get limiter for this key
			l := limiter.getLimiter(key)

			// Check if request is allowed
			if !l.Allow() {
				if o.errorHandler != nil {
					o.errorHandler(w, r)
					return
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)
				w.Write([]byte(`{"error":"rate limit exceeded"}`))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
