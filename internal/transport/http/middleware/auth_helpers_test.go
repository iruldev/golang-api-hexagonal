//go:build !integration

package middleware

import (
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"

	"github.com/iruldev/golang-api-hexagonal/internal/transport/http/ctxutil"
)

// testSecret is a 32-byte secret for HS256 signing in tests.
var testSecret = []byte("test-secret-key-32-bytes-long!!")

// fixedTime is the fixed time used in all tests for deterministic behavior.
var fixedTime = time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)

// nowFunc returns a function that always returns fixedTime.
func nowFunc() func() time.Time {
	return func() time.Time { return fixedTime }
}

// noopLogger returns a logger that discards all output.
func noopLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

// testJWTAuthConfig creates a JWTAuthConfig for testing with defaults.
func testJWTAuthConfig() JWTAuthConfig {
	return JWTAuthConfig{
		Secret: testSecret,
		Logger: noopLogger(),
		Now:    nowFunc(),
	}
}

// testJWTAuthConfigWith creates a JWTAuthConfig with custom now function.
func testJWTAuthConfigWith(now func() time.Time) JWTAuthConfig {
	return JWTAuthConfig{
		Secret: testSecret,
		Logger: noopLogger(),
		Now:    now,
	}
}

// generateValidToken creates a valid JWT token for testing.
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
