package secure

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSecureDefaults(t *testing.T) {
	middleware := New()

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	// Check default headers
	if rr.Header().Get("X-XSS-Protection") != "1; mode=block" {
		t.Errorf("Expected X-XSS-Protection='1; mode=block', got %s", rr.Header().Get("X-XSS-Protection"))
	}

	if rr.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Errorf("Expected X-Content-Type-Options='nosniff', got %s", rr.Header().Get("X-Content-Type-Options"))
	}

	if rr.Header().Get("X-Frame-Options") != "SAMEORIGIN" {
		t.Errorf("Expected X-Frame-Options='SAMEORIGIN', got %s", rr.Header().Get("X-Frame-Options"))
	}

	// HSTS should not be set by default (hstsMaxAge = 0)
	if rr.Header().Get("Strict-Transport-Security") != "" {
		t.Error("Expected Strict-Transport-Security to not be set by default")
	}
}

func TestSecureXSSProtection(t *testing.T) {
	middleware := New(WithXSSProtection("0"))

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Header().Get("X-XSS-Protection") != "0" {
		t.Errorf("Expected X-XSS-Protection='0', got %s", rr.Header().Get("X-XSS-Protection"))
	}
}

func TestSecureXFrameOptions(t *testing.T) {
	middleware := New(WithXFrameOptions("DENY"))

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Header().Get("X-Frame-Options") != "DENY" {
		t.Errorf("Expected X-Frame-Options='DENY', got %s", rr.Header().Get("X-Frame-Options"))
	}
}

func TestSecureHSTS(t *testing.T) {
	middleware := New(WithHSTSMaxAge(31536000))

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	expected := "max-age=31536000; includeSubDomains"
	if rr.Header().Get("Strict-Transport-Security") != expected {
		t.Errorf("Expected Strict-Transport-Security='%s', got %s", expected, rr.Header().Get("Strict-Transport-Security"))
	}
}

func TestSecureHSTSExcludeSubdomains(t *testing.T) {
	middleware := New(
		WithHSTSMaxAge(31536000),
		WithHSTSExcludeSubdomains(true),
	)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	expected := "max-age=31536000"
	if rr.Header().Get("Strict-Transport-Security") != expected {
		t.Errorf("Expected Strict-Transport-Security='%s', got %s", expected, rr.Header().Get("Strict-Transport-Security"))
	}
}

func TestSecureContentSecurityPolicy(t *testing.T) {
	policy := "default-src 'self'"
	middleware := New(WithContentSecurityPolicy(policy))

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Header().Get("Content-Security-Policy") != policy {
		t.Errorf("Expected Content-Security-Policy='%s', got %s", policy, rr.Header().Get("Content-Security-Policy"))
	}
}

func TestSecureCSPReportOnly(t *testing.T) {
	policy := "default-src 'self'"
	middleware := New(
		WithContentSecurityPolicy(policy),
		WithCSPReportOnly(true),
	)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Header().Get("Content-Security-Policy-Report-Only") != policy {
		t.Errorf("Expected Content-Security-Policy-Report-Only='%s', got %s", policy, rr.Header().Get("Content-Security-Policy-Report-Only"))
	}

	// Regular CSP header should not be set
	if rr.Header().Get("Content-Security-Policy") != "" {
		t.Error("Expected Content-Security-Policy to not be set when report-only is enabled")
	}
}

func TestSecureReferrerPolicy(t *testing.T) {
	policy := "no-referrer"
	middleware := New(WithReferrerPolicy(policy))

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Header().Get("Referrer-Policy") != policy {
		t.Errorf("Expected Referrer-Policy='%s', got %s", policy, rr.Header().Get("Referrer-Policy"))
	}
}

func TestSecurePermissionsPolicy(t *testing.T) {
	policy := "geolocation=(self), microphone=()"
	middleware := New(WithPermissionsPolicy(policy))

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Header().Get("Permissions-Policy") != policy {
		t.Errorf("Expected Permissions-Policy='%s', got %s", policy, rr.Header().Get("Permissions-Policy"))
	}
}

func TestSecureMultipleOptions(t *testing.T) {
	middleware := New(
		WithXSSProtection("1; mode=block"),
		WithXFrameOptions("DENY"),
		WithHSTSMaxAge(31536000),
		WithContentSecurityPolicy("default-src 'self'"),
		WithReferrerPolicy("strict-origin-when-cross-origin"),
	)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	// Check all headers are set
	if rr.Header().Get("X-XSS-Protection") != "1; mode=block" {
		t.Error("Expected X-XSS-Protection to be set")
	}
	if rr.Header().Get("X-Frame-Options") != "DENY" {
		t.Error("Expected X-Frame-Options to be set")
	}
	if rr.Header().Get("Strict-Transport-Security") == "" {
		t.Error("Expected Strict-Transport-Security to be set")
	}
	if rr.Header().Get("Content-Security-Policy") != "default-src 'self'" {
		t.Error("Expected Content-Security-Policy to be set")
	}
	if rr.Header().Get("Referrer-Policy") != "strict-origin-when-cross-origin" {
		t.Error("Expected Referrer-Policy to be set")
	}
}

func TestSecureDisableHeaders(t *testing.T) {
	middleware := New(
		WithXSSProtection(""),
		WithContentTypeNosniff(""),
		WithXFrameOptions(""),
	)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	// Headers should not be set when empty string is provided
	if rr.Header().Get("X-XSS-Protection") != "" {
		t.Error("Expected X-XSS-Protection to not be set")
	}
	if rr.Header().Get("X-Content-Type-Options") != "" {
		t.Error("Expected X-Content-Type-Options to not be set")
	}
	if rr.Header().Get("X-Frame-Options") != "" {
		t.Error("Expected X-Frame-Options to not be set")
	}
}
