package middleware_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/iruldev/golang-api-hexagonal/internal/domain/errors"
	"github.com/iruldev/golang-api-hexagonal/internal/interface/http/middleware"
	"github.com/iruldev/golang-api-hexagonal/internal/interface/http/response"
	"github.com/iruldev/golang-api-hexagonal/internal/observability"
)

// TestAuthMiddleware_IntegrationWithJWT verifies that the AuthMiddleware
// correctly handles errors returned by the real JWTAuthenticator.
// This ensures that error wrapping/unwrapping (e.g. from jwt-go) works as expected
// and triggers the correct sentinel error paths in the middleware.
func TestAuthMiddleware_IntegrationWithJWT(t *testing.T) {
	// 1. Setup real JWT Authenticator
	secret := []byte("secret-key-must-be-32-bytes-long!")
	auth, err := middleware.NewJWTAuthenticator(secret)
	if err != nil {
		t.Fatalf("Failed to create JWT authenticator: %v", err)
	}

	// 2. Setup Middleware with NopLogger
	mw := middleware.AuthMiddleware(auth, observability.NewNopLoggerInterface(), false)

	// 3. Setup Handler
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	t.Run("Valid Token", func(t *testing.T) {
		// Generate valid token
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"sub": "user-123",
		})
		tokenString, _ := token.SignedString(secret)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer "+tokenString)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("Expected 200 OK, got %d", rec.Code)
		}
	})

	t.Run("Expired Token", func(t *testing.T) {
		// Generate expired token
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"sub": "user-123",
			"exp": time.Now().Add(-1 * time.Hour).Unix(),
		})
		tokenString, _ := token.SignedString(secret)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer "+tokenString)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		// Expect 401
		if rec.Code != http.StatusUnauthorized {
			t.Errorf("Expected 401 Unauthorized, got %d", rec.Code)
		}

		// Verify JSON response structure and trace_id
		resp := parseErrorResponse(t, rec.Body.Bytes())
		if resp.Error.Code != errors.CodeTokenExpired {
			t.Errorf("Expected error code %q, got %q", errors.CodeTokenExpired, resp.Error.Code)
		}
		if resp.Meta.TraceID == "" {
			t.Error("Expected trace_id in meta field, got empty string")
		}
	})

	t.Run("Invalid Signature", func(t *testing.T) {
		// Generate token with different secret
		wrongSecret := []byte("wrong-secret-key-32-bytes-long!!!")
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"sub": "user-123",
		})
		tokenString, _ := token.SignedString(wrongSecret)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer "+tokenString)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		// Expect 401
		if rec.Code != http.StatusUnauthorized {
			t.Errorf("Expected 401 Unauthorized, got %d", rec.Code)
		}

		resp := parseErrorResponse(t, rec.Body.Bytes())
		if resp.Error.Code != errors.CodeTokenInvalid {
			t.Errorf("Expected error code %q, got %q", errors.CodeTokenInvalid, resp.Error.Code)
		}
		if resp.Meta.TraceID == "" {
			t.Error("Expected trace_id in meta field, got empty string")
		}
	})

	t.Run("Malformed Token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer invalid-token-format")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Errorf("Expected 401 Unauthorized, got %d", rec.Code)
		}

		resp := parseErrorResponse(t, rec.Body.Bytes())
		// Could be invalid or unauthorized depending on exact parsing path, but definitely 401
		// Current impl maps any jwt parsing error to TokenInvalid if not expired
		if resp.Error.Code != errors.CodeTokenInvalid {
			t.Errorf("Expected error code %q, got %q", errors.CodeTokenInvalid, resp.Error.Code)
		}
		if resp.Meta.TraceID == "" {
			t.Error("Expected trace_id in meta field, got empty string")
		}
	})
}

func parseErrorResponse(t *testing.T, body []byte) response.TestEnvelopeResponse {
	t.Helper()
	var env response.TestEnvelopeResponse
	if err := json.Unmarshal(body, &env); err != nil {
		t.Fatalf("Failed to parse error response: %v\nBody: %s", err, string(body))
	}
	return env
}
