//go:build !integration

package middleware

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/iruldev/golang-api-hexagonal/internal/app"
	"github.com/iruldev/golang-api-hexagonal/internal/transport/http/contract"
	"github.com/iruldev/golang-api-hexagonal/internal/transport/http/ctxutil"
)

// testSecret is a 32-byte secret for HS256 signing in tests
var testSecret = []byte("test-secret-key-32-bytes-long!!")

// fixedTime is the fixed time used in all tests for deterministic behavior
var fixedTime = time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)

// nowFunc returns a function that always returns fixedTime
func nowFunc() func() time.Time {
	return func() time.Time { return fixedTime }
}

// noopLogger returns a logger that discards all output
func noopLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

// testJWTAuthConfig creates a JWTAuthConfig for testing with defaults
func testJWTAuthConfig() JWTAuthConfig {
	return JWTAuthConfig{
		Secret: testSecret,
		Logger: noopLogger(),
		Now:    nowFunc(),
	}
}

// testJWTAuthConfigWith creates a JWTAuthConfig with custom now function
func testJWTAuthConfigWith(now func() time.Time) JWTAuthConfig {
	return JWTAuthConfig{
		Secret: testSecret,
		Logger: noopLogger(),
		Now:    now,
	}
}

// generateValidToken creates a valid JWT token for testing
func generateValidToken(t *testing.T, expOffset time.Duration) string {
	t.Helper()
	claims := &ctxutil.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "user-123",
			ExpiresAt: jwt.NewNumericDate(fixedTime.Add(expOffset)),
			IssuedAt:  jwt.NewNumericDate(fixedTime),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(testSecret)
	require.NoError(t, err)
	return tokenString
}

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
	var problem contract.ProblemDetail
	err := json.Unmarshal(rr.Body.Bytes(), &problem)
	require.NoError(t, err)
	assert.Equal(t, "UNAUTHORIZED", problem.Code)
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

	var problem contract.ProblemDetail
	err := json.Unmarshal(rr.Body.Bytes(), &problem)
	require.NoError(t, err)
	assert.Equal(t, "UNAUTHORIZED", problem.Code)
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

// TestJWTAuth_SetsAuthContext verifies AuthContext is populated for app layer usage via bridge.
func TestJWTAuth_SetsAuthContext(t *testing.T) {
	var gotAuth *app.AuthContext
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = app.GetAuthContext(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	claims := &ctxutil.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "user-999",
			ExpiresAt: jwt.NewNumericDate(fixedTime.Add(1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(fixedTime),
		},
		Role: "admin",
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(testSecret)
	require.NoError(t, err)

	middleware := JWTAuth(testJWTAuthConfig())
	wrapped := middleware(AuthContextBridge(handler))

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	rr := httptest.NewRecorder()

	wrapped.ServeHTTP(rr, req)

	require.NotNil(t, gotAuth)
	assert.Equal(t, "user-999", gotAuth.SubjectID)
	assert.Equal(t, "admin", gotAuth.Role)
}

// TestJWTAuth_NormalizesRoleCase verifies normalization path.
func TestJWTAuth_NormalizesRoleCase(t *testing.T) {
	var gotAuth *app.AuthContext
	var validated bool
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = app.GetAuthContext(r.Context())
		validated = isClaimsValidated(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	claims := &ctxutil.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "user-abc",
			ExpiresAt: jwt.NewNumericDate(fixedTime.Add(1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(fixedTime),
		},
		Role: "Admin", // mixed case
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(testSecret)
	require.NoError(t, err)

	middleware := JWTAuth(testJWTAuthConfig())
	wrapped := middleware(AuthContextBridge(handler))

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	rr := httptest.NewRecorder()

	wrapped.ServeHTTP(rr, req)

	require.NotNil(t, gotAuth)
	assert.Equal(t, "user-abc", gotAuth.SubjectID)
	assert.Equal(t, "admin", gotAuth.Role, "role should be normalized to lower-case")
	assert.True(t, validated, "validated marker should be set in request context")
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

// =============================================================================
// Story 2.2: JWT Claims Validation Tests
// =============================================================================

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
