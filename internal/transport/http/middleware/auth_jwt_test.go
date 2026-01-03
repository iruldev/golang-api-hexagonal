//go:build !integration

package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/iruldev/golang-api-hexagonal/internal/transport/http/contract"
	"github.com/iruldev/golang-api-hexagonal/internal/transport/http/ctxutil"
)

// JWT token validation tests - token format, parsing, algorithm, and signature validation.

// TestJWTAuth_MissingHeader tests AC #1: missing Authorization header returns 401
func TestJWTAuth_MissingHeader(t *testing.T) {
	handlerCalled := false
	handler := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		handlerCalled = true
	})

	middleware := JWTAuth(testJWTAuthConfig())
	wrapped := middleware(handler)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	rr := httptest.NewRecorder()

	wrapped.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.False(t, handlerCalled, "handler should not be called when Authorization header is missing")

	// Verify RFC 7807 response
	var problem testProblemDetail
	err := json.Unmarshal(rr.Body.Bytes(), &problem)
	require.NoError(t, err)
	assert.Equal(t, contract.CodeAuthExpiredToken, problem.Code) // Story 2.3: New taxonomy
	assert.Equal(t, 401, problem.Status)
}

// TestJWTAuth_MalformedToken tests AC #2: malformed token returns 401
func TestJWTAuth_MalformedToken(t *testing.T) {
	handlerCalled := false
	handler := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		handlerCalled = true
	})

	middleware := JWTAuth(testJWTAuthConfig())
	wrapped := middleware(handler)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer not-a-valid-jwt")
	rr := httptest.NewRecorder()

	wrapped.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.False(t, handlerCalled, "handler should not be called for malformed token")

	var problem testProblemDetail
	err := json.Unmarshal(rr.Body.Bytes(), &problem)
	require.NoError(t, err)
	assert.Equal(t, contract.CodeAuthExpiredToken, problem.Code) // Story 2.3: New taxonomy
}

// TestJWTAuth_InvalidSignature tests AC #2: wrong signature returns 401
func TestJWTAuth_InvalidSignature(t *testing.T) {
	handlerCalled := false
	handler := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		handlerCalled = true
	})

	// Create token with different secret
	wrongSecret := []byte("wrong-secret-key-32-bytes-long!!")
	claims := &ctxutil.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "user-123",
			ExpiresAt: jwt.NewNumericDate(fixedTime.Add(1 * time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString(wrongSecret)

	middleware := JWTAuth(testJWTAuthConfig())
	wrapped := middleware(handler)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	rr := httptest.NewRecorder()

	wrapped.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.False(t, handlerCalled, "handler should not be called for invalid signature")
}

// TestJWTAuth_ExpiredToken tests AC #3, #4: expired token returns 401 using injected time
func TestJWTAuth_ExpiredToken(t *testing.T) {
	handlerCalled := false
	handler := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		handlerCalled = true
	})

	// Token expired 1 hour ago
	tokenString := generateValidToken(t, -1*time.Hour)

	middleware := JWTAuth(testJWTAuthConfig())
	wrapped := middleware(handler)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	rr := httptest.NewRecorder()

	wrapped.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.False(t, handlerCalled, "handler should not be called for expired token")
}

// TestJWTAuth_ValidToken tests AC #5: valid token passes and claims are in context
func TestJWTAuth_ValidToken(t *testing.T) {
	var gotClaims *ctxutil.Claims
	handlerCalled := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		gotClaims = ctxutil.GetClaims(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	// Token expires in 1 hour (valid)
	tokenString := generateValidToken(t, 1*time.Hour)

	middleware := JWTAuth(testJWTAuthConfig())
	wrapped := middleware(handler)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	rr := httptest.NewRecorder()

	wrapped.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.True(t, handlerCalled, "handler should be called for valid token")
	require.NotNil(t, gotClaims, "claims should be in context")
	assert.Equal(t, "user-123", gotClaims.Subject)
}

func TestJWTAuth_BearerCaseInsensitive(t *testing.T) {
	testCases := []string{"bearer", "Bearer", "BEARER", "BeArEr"}

	for _, bearerCase := range testCases {
		t.Run(bearerCase, func(t *testing.T) {
			handlerCalled := false
			handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				handlerCalled = true
				w.WriteHeader(http.StatusOK)
			})

			tokenString := generateValidToken(t, 1*time.Hour)

			middleware := JWTAuth(testJWTAuthConfig())
			wrapped := middleware(handler)

			req := httptest.NewRequest(http.MethodGet, "/protected", nil)
			req.Header.Set("Authorization", bearerCase+" "+tokenString)
			rr := httptest.NewRecorder()

			wrapped.ServeHTTP(rr, req)

			assert.Equal(t, http.StatusOK, rr.Code, "should accept %s prefix", bearerCase)
			assert.True(t, handlerCalled)
		})
	}
}

func TestJWTAuth_InvalidAuthScheme(t *testing.T) {
	handlerCalled := false
	handler := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		handlerCalled = true
	})

	tokenString := generateValidToken(t, 1*time.Hour)

	middleware := JWTAuth(testJWTAuthConfig())
	wrapped := middleware(handler)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Basic "+tokenString) // Wrong scheme
	rr := httptest.NewRecorder()

	wrapped.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.False(t, handlerCalled)
}

func TestJWTAuth_EmptyBearerToken(t *testing.T) {
	handlerCalled := false
	handler := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		handlerCalled = true
	})

	middleware := JWTAuth(testJWTAuthConfig())
	wrapped := middleware(handler)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer ") // Empty token
	rr := httptest.NewRecorder()

	wrapped.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.False(t, handlerCalled)
}

func TestJWTAuth_WrongAlgorithm(t *testing.T) {
	handlerCalled := false
	handler := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		handlerCalled = true
	})

	// Create token with HS384 (not allowed)
	claims := &ctxutil.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "user-123",
			ExpiresAt: jwt.NewNumericDate(fixedTime.Add(1 * time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS384, claims)
	tokenString, _ := token.SignedString(testSecret)

	middleware := JWTAuth(testJWTAuthConfig())
	wrapped := middleware(handler)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	rr := httptest.NewRecorder()

	wrapped.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.False(t, handlerCalled, "handler should not be called for wrong algorithm")
}

func TestJWTAuth_AlgNone(t *testing.T) {
	handlerCalled := false
	handler := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		handlerCalled = true
	})

	// Create token with "none" algorithm (critical vulnerability check)
	token := jwt.New(jwt.SigningMethodNone)
	claims := token.Claims.(jwt.MapClaims)
	claims["sub"] = "user-123"
	claims["exp"] = float64(fixedTime.Add(1 * time.Hour).Unix())

	// Unsigned token
	tokenString, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
	require.NoError(t, err)

	middleware := JWTAuth(testJWTAuthConfig())
	wrapped := middleware(handler)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	rr := httptest.NewRecorder()

	wrapped.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.False(t, handlerCalled, "handler should not be called for alg:none")
}

func TestJWTAuth_TimeInjection(t *testing.T) {
	// Create a token that expires at a specific time
	expTime := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)
	claims := &ctxutil.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "user-123",
			ExpiresAt: jwt.NewNumericDate(expTime),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString(testSecret)

	// Test 1: Before expiry - should pass
	t.Run("before expiry", func(t *testing.T) {
		handlerCalled := false
		handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			handlerCalled = true
			w.WriteHeader(http.StatusOK)
		})

		beforeExp := func() time.Time { return expTime.Add(-1 * time.Hour) }
		middleware := JWTAuth(testJWTAuthConfigWith(beforeExp))
		wrapped := middleware(handler)

		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+tokenString)
		rr := httptest.NewRecorder()

		wrapped.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.True(t, handlerCalled)
	})

	// Test 2: After expiry - should fail
	t.Run("after expiry", func(t *testing.T) {
		handlerCalled := false
		handler := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
			handlerCalled = true
		})

		afterExp := func() time.Time { return expTime.Add(1 * time.Hour) }
		middleware := JWTAuth(testJWTAuthConfigWith(afterExp))
		wrapped := middleware(handler)

		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+tokenString)
		rr := httptest.NewRecorder()

		wrapped.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
		assert.False(t, handlerCalled)
	})
}

func TestJWTAuth_GetClaimsFromContext(t *testing.T) {
	// Create token with specific claims
	claims := &ctxutil.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "test-issuer",
			Subject:   "user-456",
			Audience:  jwt.ClaimStrings{"api-client"},
			ExpiresAt: jwt.NewNumericDate(fixedTime.Add(1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(fixedTime),
			ID:        "jti-789",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString(testSecret)

	var gotClaims *ctxutil.Claims
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotClaims = ctxutil.GetClaims(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	middleware := JWTAuth(testJWTAuthConfig())
	wrapped := middleware(handler)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	rr := httptest.NewRecorder()

	wrapped.ServeHTTP(rr, req)

	require.NotNil(t, gotClaims)
	assert.Equal(t, "test-issuer", gotClaims.Issuer)
	assert.Equal(t, "user-456", gotClaims.Subject)
	assert.Contains(t, gotClaims.Audience, "api-client")
	assert.Equal(t, "jti-789", gotClaims.ID)
}

// TestJWTAuth_EnforcesHS256_ConstantTime validates AC #3:
// Ensures we are STRICTLY configured to use HS256.
// HS256 in golang-jwt uses hmac.Equal, which is a constant-time comparison.
// This test ensures we don't accidentally enable alg:none or RSA (which might not be constant-time here).
func TestJWTAuth_EnforcesHS256_ConstantTime(t *testing.T) {
	// 1. Verify AllowedAlgorithm constant
	assert.Equal(t, "HS256", AllowedAlgorithm, "Must use HS256 for constant-time HMAC")

	// 2. Verify behavior: Reject valid JWT signed with different alg (e.g., HS384)
	// even if the secret is correct. This proves we strictly enforce the algorithm
	// associated with constant-time check.
	handlerCalled := false
	handler := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		handlerCalled = true
	})

	// Create valid token but with HS512 (arguably also constant time, but we want STRICT whitelist)
	claims := &ctxutil.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "user-123",
			ExpiresAt: jwt.NewNumericDate(fixedTime.Add(1 * time.Hour)),
		},
	}
	// HS512 is safe/constant-time usually, but our requirement is strict usage of configured alg.
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	tokenString, _ := token.SignedString(testSecret)

	middleware := JWTAuth(testJWTAuthConfig())
	wrapped := middleware(handler)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	rr := httptest.NewRecorder()

	wrapped.ServeHTTP(rr, req)

	// Should be rejected because we only allow HS256
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.False(t, handlerCalled)
}
