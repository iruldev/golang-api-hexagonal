package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/iruldev/golang-api-hexagonal/internal/ctxutil"
)

// MinSecretKeyLength is the minimum recommended length for HMAC-SHA256 secrets.
const MinSecretKeyLength = 32

// ErrSecretKeyTooShort indicates the secret key is shorter than MinSecretKeyLength.
var ErrSecretKeyTooShort = errors.New("secret key must be at least 32 bytes for HMAC-SHA256")

// JWTConfig holds JWT authenticator configuration.
type JWTConfig struct {
	// SecretKey is the HMAC secret used to verify token signatures.
	// Should be at least 32 bytes for HS256.
	SecretKey []byte

	// Issuer optionally validates the "iss" claim.
	Issuer string

	// Audience optionally validates the "aud" claim.
	Audience string
}

// JWTOption configures the JWT authenticator.
type JWTOption func(*JWTConfig)

// WithIssuer validates the token issuer claim.
func WithIssuer(issuer string) JWTOption {
	return func(c *JWTConfig) {
		c.Issuer = issuer
	}
}

// WithAudience validates the token audience claim.
func WithAudience(audience string) JWTOption {
	return func(c *JWTConfig) {
		c.Audience = audience
	}
}

// JWTAuthenticator validates JWT tokens from the Authorization header.
// It implements the Authenticator interface.
//
// The authenticator expects tokens in the format: "Bearer <token>"
// and validates them using HMAC-SHA256 (HS256) signing method.
//
// Example usage:
//
//	// Create authenticator with secret key
//	jwtAuth := middleware.NewJWTAuthenticator(
//	    []byte(os.Getenv("JWT_SECRET")),
//	    middleware.WithIssuer("my-app"),
//	)
//
//	// Use with AuthMiddleware
//	router.Group(func(r chi.Router) {
//	    r.Use(middleware.AuthMiddleware(jwtAuth))
//	    r.Get("/api/v1/protected", protectedHandler)
//	})
type JWTAuthenticator struct {
	config JWTConfig
	// Cached parser options to avoid re-allocation on every request
	parserOptions []jwt.ParserOption
}

// NewJWTAuthenticator creates a new JWT authenticator with the given secret key.
//
// The secret key must be at least 32 bytes for HMAC-SHA256 security.
// Returns ErrSecretKeyTooShort if the key is too short.
// Options can be used to configure issuer and audience validation.
//
// Example:
//
//	jwtAuth, err := middleware.NewJWTAuthenticator(
//	    []byte("your-secret-key-at-least-32-bytes!!"),
//	    middleware.WithIssuer("my-app"),
//	    middleware.WithAudience("my-api"),
//	)
//	if err != nil {
//	    log.Fatal(err)
//	}
func NewJWTAuthenticator(secretKey []byte, opts ...JWTOption) (*JWTAuthenticator, error) {
	if len(secretKey) < MinSecretKeyLength {
		return nil, ErrSecretKeyTooShort
	}
	config := JWTConfig{SecretKey: secretKey}
	for _, opt := range opts {
		opt(&config)
	}

	auth := &JWTAuthenticator{config: config}
	auth.parserOptions = auth.buildParserOptions()

	return auth, nil
}

// Authenticate implements the Authenticator interface.
// It extracts and validates a JWT from the Authorization header.
//
// Returns:
//   - ctxutil.Claims: The extracted user claims if authentication succeeds
//   - ErrUnauthenticated: If Authorization header is missing or empty
//   - ErrTokenInvalid: If token format, signature, or algorithm is invalid
//   - ErrTokenExpired: If token has expired
func (a *JWTAuthenticator) Authenticate(r *http.Request) (ctxutil.Claims, error) {
	// 1. Extract Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return ctxutil.Claims{}, ErrUnauthenticated
	}

	// 2. Strip "Bearer " prefix
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return ctxutil.Claims{}, ErrUnauthenticated
	}
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")

	// Empty token after stripping prefix
	if tokenString == "" {
		return ctxutil.Claims{}, ErrUnauthenticated
	}

	// 3. Parse and validate token
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		// Validate algorithm - only allow HMAC (reject "none" and others)
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrTokenInvalid
		}
		return a.config.SecretKey, nil
	}, a.parserOptions...)

	if err != nil {
		// Map jwt library errors to our sentinel errors
		if errors.Is(err, jwt.ErrTokenExpired) {
			return ctxutil.Claims{}, ErrTokenExpired
		}
		// Wrap the original error to preserve context for logging
		return ctxutil.Claims{}, fmt.Errorf("%w: %v", ErrTokenInvalid, err)
	}

	// Defensive check: in practice jwt.Parse returns error for invalid tokens,
	// but we keep this as a safety net for edge cases or library changes.
	if !token.Valid {
		return ctxutil.Claims{}, ErrTokenInvalid
	}

	// 4. Extract claims
	mapClaims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return ctxutil.Claims{}, ErrTokenInvalid
	}

	return mapJWTClaims(mapClaims), nil
}

// buildParserOptions constructs parser options based on config.
func (a *JWTAuthenticator) buildParserOptions() []jwt.ParserOption {
	opts := []jwt.ParserOption{
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
	}

	if a.config.Issuer != "" {
		opts = append(opts, jwt.WithIssuer(a.config.Issuer))
	}
	if a.config.Audience != "" {
		opts = append(opts, jwt.WithAudience(a.config.Audience))
	}

	return opts
}

// mapJWTClaims converts JWT claims to ctxutil.Claims.
func mapJWTClaims(jwtClaims jwt.MapClaims) ctxutil.Claims {
	c := ctxutil.Claims{
		Metadata: make(map[string]string),
	}

	// Map standard claims
	if sub, ok := jwtClaims["sub"].(string); ok {
		c.UserID = sub
	}

	// Map roles (array)
	if roles, ok := jwtClaims["roles"].([]interface{}); ok {
		for _, r := range roles {
			if role, ok := r.(string); ok {
				c.Roles = append(c.Roles, role)
			}
		}
	}

	// Map permissions (array)
	if perms, ok := jwtClaims["permissions"].([]interface{}); ok {
		for _, p := range perms {
			if perm, ok := p.(string); ok {
				c.Permissions = append(c.Permissions, perm)
			}
		}
	}

	// Map metadata (map)
	if meta, ok := jwtClaims["metadata"].(map[string]interface{}); ok {
		for k, v := range meta {
			if val, ok := v.(string); ok {
				c.Metadata[k] = val
			}
		}
	}

	return c
}
