//go:build !integration

package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"

	"github.com/iruldev/golang-api-hexagonal/internal/transport/http/ctxutil"
)

// =============================================================================
// Story 2.2: JWT Claims Validation Tests
// =============================================================================

// Claims validation tests - issuer, audience, missing exp, and clock skew.

// TestJWTAuth_MissingExp tests AC #1: token without exp claim returns 401
func TestJWTAuth_MissingExp(t *testing.T) {
	handlerCalled := false
	handler := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		handlerCalled = true
	})

	// Create token WITHOUT exp claim
	claims := &ctxutil.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:  "user-123",
			IssuedAt: jwt.NewNumericDate(fixedTime),
			// No ExpiresAt!
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString(testSecret)

	middleware := JWTAuth(testJWTAuthConfig())
	wrapped := middleware(handler)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	rr := httptest.NewRecorder()

	wrapped.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.False(t, handlerCalled, "handler should not be called for token without exp")
}

// TestJWTAuth_WrongIssuer tests AC #2: token with wrong issuer returns 401
func TestJWTAuth_WrongIssuer(t *testing.T) {
	handlerCalled := false
	handler := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		handlerCalled = true
	})

	// Create token with wrong issuer
	claims := &ctxutil.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "user-123",
			Issuer:    "wrong-issuer",
			ExpiresAt: jwt.NewNumericDate(fixedTime.Add(1 * time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString(testSecret)

	// Configure middleware to expect specific issuer
	cfg := testJWTAuthConfig()
	cfg.Issuer = "expected-issuer"
	middleware := JWTAuth(cfg)
	wrapped := middleware(handler)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	rr := httptest.NewRecorder()

	wrapped.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.False(t, handlerCalled, "handler should not be called for wrong issuer")
}

// TestJWTAuth_WrongAudience tests AC #2: token with wrong audience returns 401
func TestJWTAuth_WrongAudience(t *testing.T) {
	handlerCalled := false
	handler := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		handlerCalled = true
	})

	// Create token with wrong audience
	claims := &ctxutil.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "user-123",
			Audience:  jwt.ClaimStrings{"wrong-audience"},
			ExpiresAt: jwt.NewNumericDate(fixedTime.Add(1 * time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString(testSecret)

	// Configure middleware to expect specific audience
	cfg := testJWTAuthConfig()
	cfg.Audience = "expected-audience"
	middleware := JWTAuth(cfg)
	wrapped := middleware(handler)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	rr := httptest.NewRecorder()

	wrapped.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.False(t, handlerCalled, "handler should not be called for wrong audience")
}

// TestJWTAuth_ClockSkew tests AC #3: expired token within skew tolerance passes
func TestJWTAuth_ClockSkew(t *testing.T) {
	t.Run("expired within skew passes", func(t *testing.T) {
		handlerCalled := false
		handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			handlerCalled = true
			w.WriteHeader(http.StatusOK)
		})

		// Create token expired 10 seconds ago
		claims := &ctxutil.Claims{
			RegisteredClaims: jwt.RegisteredClaims{
				Subject:   "user-123",
				ExpiresAt: jwt.NewNumericDate(fixedTime.Add(-10 * time.Second)),
			},
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, _ := token.SignedString(testSecret)

		// Configure middleware with 30s clock skew
		cfg := testJWTAuthConfig()
		cfg.ClockSkew = 30 * time.Second
		middleware := JWTAuth(cfg)
		wrapped := middleware(handler)

		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+tokenString)
		rr := httptest.NewRecorder()

		wrapped.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.True(t, handlerCalled, "handler should be called for token expired within skew")
	})

	t.Run("expired beyond skew fails", func(t *testing.T) {
		handlerCalled := false
		handler := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
			handlerCalled = true
		})

		// Create token expired 60 seconds ago
		claims := &ctxutil.Claims{
			RegisteredClaims: jwt.RegisteredClaims{
				Subject:   "user-123",
				ExpiresAt: jwt.NewNumericDate(fixedTime.Add(-60 * time.Second)),
			},
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, _ := token.SignedString(testSecret)

		// Configure middleware with 30s clock skew
		cfg := testJWTAuthConfig()
		cfg.ClockSkew = 30 * time.Second
		middleware := JWTAuth(cfg)
		wrapped := middleware(handler)

		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+tokenString)
		rr := httptest.NewRecorder()

		wrapped.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
		assert.False(t, handlerCalled, "handler should not be called for token expired beyond skew")
	})
}

// TestJWTAuth_CorrectIssuerPasses verifies correct issuer allows access
func TestJWTAuth_CorrectIssuerPasses(t *testing.T) {
	handlerCalled := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	})

	claims := &ctxutil.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "user-123",
			Issuer:    "expected-issuer",
			ExpiresAt: jwt.NewNumericDate(fixedTime.Add(1 * time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString(testSecret)

	cfg := testJWTAuthConfig()
	cfg.Issuer = "expected-issuer"
	middleware := JWTAuth(cfg)
	wrapped := middleware(handler)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	rr := httptest.NewRecorder()

	wrapped.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.True(t, handlerCalled)
}
