//go:build !integration

package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/iruldev/golang-api-hexagonal/internal/app"
	"github.com/iruldev/golang-api-hexagonal/internal/transport/http/ctxutil"
)

// Middleware tests - context propagation, role normalization, and AuthContextBridge.

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

func TestNormalizeRole(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"mixed case", "Admin", "admin"},
		{"with spaces", "  User  ", "user"},
		{"all caps", "ADMIN", "admin"},
		{"already normal", "user", "user"},
		{"empty", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, NormalizeRole(tt.input))
		})
	}
}
