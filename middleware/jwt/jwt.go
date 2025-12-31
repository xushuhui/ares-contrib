package jwt

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

const (
	// bearerWord the bearer key word for authorization
	bearerWord string = "Bearer"

	// authorizationKey holds the key used to store the JWT Token in the request header.
	authorizationKey string = "Authorization"
)

var (
	ErrMissingJwtToken        = errors.New("JWT token is missing")
	ErrMissingKeyFunc         = errors.New("keyFunc is missing")
	ErrTokenInvalid           = errors.New("token is invalid")
	ErrTokenExpired           = errors.New("JWT token has expired")
	ErrTokenParseFail         = errors.New("fail to parse JWT token")
	ErrUnSupportSigningMethod = errors.New("wrong signing method")
)

// Option is jwt option.
type Option func(*options)

// options holds JWT middleware configuration
type options struct {
	signingKey    []byte
	signingMethod jwt.SigningMethod
	claims        func() jwt.Claims
	contextKey    string
}

// WithSigningMethod with signing method option.
func WithSigningMethod(method jwt.SigningMethod) Option {
	return func(o *options) {
		o.signingMethod = method
	}
}

// WithClaims with custom claim
// f needs to return a new jwt.Claims object each time to avoid concurrent write problems
func WithClaims(f func() jwt.Claims) Option {
	return func(o *options) {
		o.claims = f
	}
}

// WithContextKey with custom context key for storing claims
func WithContextKey(key string) Option {
	return func(o *options) {
		o.contextKey = key
	}
}

// New returns a JWT middleware with signing key and optional configuration
func New(signingKey []byte, opts ...Option) func(http.Handler) http.Handler {
	o := &options{
		signingKey:    signingKey,
		signingMethod: jwt.SigningMethodHS256,
		contextKey:    "user",
	}
	for _, opt := range opts {
		opt(o)
	}

	// Validate signing key
	if o.signingKey == nil {
		panic("jwt middleware requires signing key")
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract token from Authorization header
			auths := strings.SplitN(r.Header.Get(authorizationKey), " ", 2)
			if len(auths) != 2 || !strings.EqualFold(auths[0], bearerWord) {
				http.Error(w, ErrMissingJwtToken.Error(), http.StatusUnauthorized)
				return
			}
			jwtToken := auths[1]

			// Parse token
			var (
				tokenInfo *jwt.Token
				err       error
			)

			// Create keyFunc
			keyFunc := func(token *jwt.Token) (interface{}, error) {
				return o.signingKey, nil
			}

			if o.claims != nil {
				tokenInfo, err = jwt.ParseWithClaims(jwtToken, o.claims(), keyFunc)
			} else {
				tokenInfo, err = jwt.Parse(jwtToken, keyFunc)
			}

			if err != nil {
				// Classify error types
				if errors.Is(err, jwt.ErrTokenMalformed) || errors.Is(err, jwt.ErrTokenUnverifiable) {
					http.Error(w, ErrTokenInvalid.Error(), http.StatusUnauthorized)
					return
				}
				if errors.Is(err, jwt.ErrTokenNotValidYet) || errors.Is(err, jwt.ErrTokenExpired) {
					http.Error(w, ErrTokenExpired.Error(), http.StatusUnauthorized)
					return
				}
				http.Error(w, ErrTokenParseFail.Error(), http.StatusUnauthorized)
				return
			}

			// Validate token
			if !tokenInfo.Valid {
				http.Error(w, ErrTokenInvalid.Error(), http.StatusUnauthorized)
				return
			}

			// Verify signing method
			if tokenInfo.Method != o.signingMethod {
				http.Error(w, ErrUnSupportSigningMethod.Error(), http.StatusUnauthorized)
				return
			}

			// Store claims in context
			ctx := context.WithValue(r.Context(), contextKey(o.contextKey), tokenInfo.Claims)
			r = r.WithContext(ctx)

			next.ServeHTTP(w, r)
		})
	}
}

// contextKey is the type used for context keys
type contextKey string

// GetClaims extracts JWT claims from context
func GetClaims(ctx context.Context) (jwt.Claims, bool) {
	claims, ok := ctx.Value(contextKey("user")).(jwt.Claims)
	return claims, ok
}

// GetClaimsWithKey extracts JWT claims from context with custom key
func GetClaimsWithKey(ctx context.Context, key string) (jwt.Claims, bool) {
	claims, ok := ctx.Value(contextKey(key)).(jwt.Claims)
	return claims, ok
}
