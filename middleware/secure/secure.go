package secure

import (
	"net/http"
	"strconv"
)

// Option is secure option.
type Option func(*options)

// options defines the configuration for secure middleware
type options struct {
	// XSSProtection provides protection against cross-site scripting attack (XSS)
	// by setting the `X-XSS-Protection` header.
	// Default: "1; mode=block"
	xssProtection string

	// ContentTypeNosniff provides protection against overriding Content-Type
	// header by setting the `X-Content-Type-Options` header.
	// Default: "nosniff"
	contentTypeNosniff string

	// XFrameOptions can be used to indicate whether or not a browser should
	// be allowed to render a page in a <frame>, <iframe> or <object>.
	// Default: "SAMEORIGIN"
	xFrameOptions string

	// HSTSMaxAge sets the `Strict-Transport-Security` header to indicate how
	// long (in seconds) browsers should remember that this site is only to
	// be accessed using HTTPS.
	// Default: 0 (disabled)
	hstsMaxAge int

	// HSTSExcludeSubdomains won't include subdomains tag in the `Strict-Transport-Security`
	// header, excluding all subdomains from security policy.
	// Default: false
	hstsExcludeSubdomains bool

	// ContentSecurityPolicy sets the `Content-Security-Policy` header providing
	// security against cross-site scripting (XSS), clickjacking and other code
	// injection attacks.
	// Default: ""
	contentSecurityPolicy string

	// CSPReportOnly would use the `Content-Security-Policy-Report-Only` header instead
	// of the `Content-Security-Policy` header. This allows iterative updates of the
	// content security policy by only reporting the violations that would have occurred
	// instead of blocking the resource.
	// Default: false
	cspReportOnly bool

	// ReferrerPolicy sets the `Referrer-Policy` header providing security against
	// leaking potentially sensitive request paths to third parties.
	// Default: ""
	referrerPolicy string

	// PermissionsPolicy sets the `Permissions-Policy` header providing security
	// against using browser features in documents or iframes.
	// Default: ""
	permissionsPolicy string
}

// WithXSSProtection sets the X-XSS-Protection header
func WithXSSProtection(value string) Option {
	return func(o *options) {
		o.xssProtection = value
	}
}

// WithContentTypeNosniff sets the X-Content-Type-Options header
func WithContentTypeNosniff(value string) Option {
	return func(o *options) {
		o.contentTypeNosniff = value
	}
}

// WithXFrameOptions sets the X-Frame-Options header
func WithXFrameOptions(value string) Option {
	return func(o *options) {
		o.xFrameOptions = value
	}
}

// WithHSTSMaxAge sets the Strict-Transport-Security header max-age
func WithHSTSMaxAge(maxAge int) Option {
	return func(o *options) {
		o.hstsMaxAge = maxAge
	}
}

// WithHSTSExcludeSubdomains excludes subdomains from HSTS
func WithHSTSExcludeSubdomains(exclude bool) Option {
	return func(o *options) {
		o.hstsExcludeSubdomains = exclude
	}
}

// WithContentSecurityPolicy sets the Content-Security-Policy header
func WithContentSecurityPolicy(policy string) Option {
	return func(o *options) {
		o.contentSecurityPolicy = policy
	}
}

// WithCSPReportOnly enables CSP report-only mode
func WithCSPReportOnly(reportOnly bool) Option {
	return func(o *options) {
		o.cspReportOnly = reportOnly
	}
}

// WithReferrerPolicy sets the Referrer-Policy header
func WithReferrerPolicy(policy string) Option {
	return func(o *options) {
		o.referrerPolicy = policy
	}
}

// WithPermissionsPolicy sets the Permissions-Policy header
func WithPermissionsPolicy(policy string) Option {
	return func(o *options) {
		o.permissionsPolicy = policy
	}
}

// New returns a middleware that sets security headers
func New(opts ...Option) func(http.Handler) http.Handler {
	o := &options{
		xssProtection:      "1; mode=block",
		contentTypeNosniff: "nosniff",
		xFrameOptions:      "SAMEORIGIN",
		hstsMaxAge:         0,
	}

	for _, opt := range opts {
		opt(o)
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// X-XSS-Protection
			if o.xssProtection != "" {
				w.Header().Set("X-XSS-Protection", o.xssProtection)
			}

			// X-Content-Type-Options
			if o.contentTypeNosniff != "" {
				w.Header().Set("X-Content-Type-Options", o.contentTypeNosniff)
			}

			// X-Frame-Options
			if o.xFrameOptions != "" {
				w.Header().Set("X-Frame-Options", o.xFrameOptions)
			}

			// Strict-Transport-Security
			if o.hstsMaxAge > 0 {
				hstsValue := "max-age=" + strconv.Itoa(o.hstsMaxAge)
				if !o.hstsExcludeSubdomains {
					hstsValue += "; includeSubDomains"
				}
				w.Header().Set("Strict-Transport-Security", hstsValue)
			}

			// Content-Security-Policy
			if o.contentSecurityPolicy != "" {
				if o.cspReportOnly {
					w.Header().Set("Content-Security-Policy-Report-Only", o.contentSecurityPolicy)
				} else {
					w.Header().Set("Content-Security-Policy", o.contentSecurityPolicy)
				}
			}

			// Referrer-Policy
			if o.referrerPolicy != "" {
				w.Header().Set("Referrer-Policy", o.referrerPolicy)
			}

			// Permissions-Policy
			if o.permissionsPolicy != "" {
				w.Header().Set("Permissions-Policy", o.permissionsPolicy)
			}

			next.ServeHTTP(w, r)
		})
	}
}
