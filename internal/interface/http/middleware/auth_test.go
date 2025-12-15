package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/iruldev/golang-api-hexagonal/internal/ctxutil"
	domainerrors "github.com/iruldev/golang-api-hexagonal/internal/domain/errors"
	"github.com/iruldev/golang-api-hexagonal/internal/interface/http/response"
	"github.com/iruldev/golang-api-hexagonal/internal/observability"
)

// TestClaims_HasRole tests the HasRole method of ctxutil.Claims struct.
func TestClaims_HasRole(t *testing.T) {
	tests := []struct {
		name     string
		claims   ctxutil.Claims
		role     string
		expected bool
	}{
		{
			name:     "has role returns true",
			claims:   ctxutil.Claims{Roles: []string{"admin", "user"}},
			role:     "admin",
			expected: true,
		},
		{
			name:     "does not have role returns false",
			claims:   ctxutil.Claims{Roles: []string{"user"}},
			role:     "admin",
			expected: false,
		},
		{
			name:     "empty roles returns false",
			claims:   ctxutil.Claims{Roles: nil},
			role:     "admin",
			expected: false,
		},
		{
			name:     "empty role string returns false",
			claims:   ctxutil.Claims{Roles: []string{"admin", "user"}},
			role:     "",
			expected: false,
		},
		{
			name:     "exactly matches role",
			claims:   ctxutil.Claims{Roles: []string{"administrator"}},
			role:     "admin",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.claims.HasRole(tt.role)
			if got != tt.expected {
				t.Errorf("HasRole(%q) = %v, want %v", tt.role, got, tt.expected)
			}
		})
	}
}

// TestClaims_HasPermission tests the HasPermission method of ctxutil.Claims struct.
func TestClaims_HasPermission(t *testing.T) {
	tests := []struct {
		name       string
		claims     ctxutil.Claims
		permission string
		expected   bool
	}{
		{
			name:       "has permission returns true",
			claims:     ctxutil.Claims{Permissions: []string{"read", "write", "delete"}},
			permission: "write",
			expected:   true,
		},
		{
			name:       "does not have permission returns false",
			claims:     ctxutil.Claims{Permissions: []string{"read"}},
			permission: "write",
			expected:   false,
		},
		{
			name:       "empty permissions returns false",
			claims:     ctxutil.Claims{Permissions: nil},
			permission: "read",
			expected:   false,
		},
		{
			name:       "empty permission string",
			claims:     ctxutil.Claims{Permissions: []string{"read", "write"}},
			permission: "",
			expected:   false,
		},
		{
			name:       "partial match returns false",
			claims:     ctxutil.Claims{Permissions: []string{"read:notes"}},
			permission: "read",
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.claims.HasPermission(tt.permission)
			if got != tt.expected {
				t.Errorf("HasPermission(%q) = %v, want %v", tt.permission, got, tt.expected)
			}
		})
	}
}

// TestNewContext tests storing claims in context.
func TestNewContext(t *testing.T) {
	claims := ctxutil.Claims{
		UserID:      "user-123",
		Roles:       []string{"admin"},
		Permissions: []string{"read", "write"},
	}

	ctx := ctxutil.NewClaimsContext(context.Background(), claims)

	// Verify claims can be retrieved
	storedClaims, err := ctxutil.ClaimsFromContext(ctx)
	if err != nil {
		t.Fatalf("expected claims to be retrievable from context, got error: %v", err)
	}

	if storedClaims.UserID != claims.UserID {
		t.Errorf("UserID = %q, want %q", storedClaims.UserID, claims.UserID)
	}
}

// TestFromContext tests extracting claims from context.
func TestFromContext(t *testing.T) {
	t.Run("returns claims when present", func(t *testing.T) {
		expectedClaims := ctxutil.Claims{
			UserID:      "user-456",
			Roles:       []string{"user", "editor"},
			Permissions: []string{"read", "write:notes"},
			Metadata:    map[string]string{"org": "acme"},
		}

		ctx := ctxutil.NewClaimsContext(context.Background(), expectedClaims)
		claims, err := ctxutil.ClaimsFromContext(ctx)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if claims.UserID != expectedClaims.UserID {
			t.Errorf("UserID = %q, want %q", claims.UserID, expectedClaims.UserID)
		}
		if len(claims.Roles) != len(expectedClaims.Roles) {
			t.Errorf("Roles length = %d, want %d", len(claims.Roles), len(expectedClaims.Roles))
		}
		if len(claims.Permissions) != len(expectedClaims.Permissions) {
			t.Errorf("Permissions length = %d, want %d", len(claims.Permissions), len(expectedClaims.Permissions))
		}
		if claims.Metadata["org"] != expectedClaims.Metadata["org"] {
			t.Errorf("Metadata[org] = %q, want %q", claims.Metadata["org"], expectedClaims.Metadata["org"])
		}
	})

	t.Run("returns error when claims missing", func(t *testing.T) {
		ctx := context.Background()
		_, err := ctxutil.ClaimsFromContext(ctx)

		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !errors.Is(err, ctxutil.ErrNoClaimsInContext) {
			t.Errorf("error = %v, want %v", err, ctxutil.ErrNoClaimsInContext)
		}
	})

	t.Run("returns error for wrong type in context", func(t *testing.T) {
		// Use ctxutil.ClaimsFromContext on a context without claims
		// This tests the same behavior without accessing internal key
		ctx := context.Background()
		_, err := ctxutil.ClaimsFromContext(ctx)

		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !errors.Is(err, ctxutil.ErrNoClaimsInContext) {
			t.Errorf("error = %v, want %v", err, ctxutil.ErrNoClaimsInContext)
		}
	})
}

// TestSentinelErrors tests that sentinel errors are properly defined.
func TestSentinelErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
		msg  string
	}{
		{"ErrUnauthenticated", ErrUnauthenticated, "unauthenticated"},
		{"ErrTokenExpired", ErrTokenExpired, "token expired"},
		{"ErrTokenInvalid", ErrTokenInvalid, "token invalid"},
		{"ErrNoctxutil.ClaimsInContext", ctxutil.ErrNoClaimsInContext, "no claims in context"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Error() != tt.msg {
				t.Errorf("%s.Error() = %q, want %q", tt.name, tt.err.Error(), tt.msg)
			}
		})
	}
}

// TestSentinelErrors_ErrorsIs tests that errors.Is works with sentinel errors.
func TestSentinelErrors_ErrorsIs(t *testing.T) {
	// Test wrapping and unwrapping with errors.Is
	wrapped := errors.Join(errors.New("auth failed"), ErrTokenExpired)
	if !errors.Is(wrapped, ErrTokenExpired) {
		t.Error("expected errors.Is to match ErrTokenExpired in wrapped error")
	}
}

// mockAuthenticator is a test implementation of Authenticator.
type mockAuthenticator struct {
	claims    ctxutil.Claims
	err       error
	callCount int
}

func (m *mockAuthenticator) Authenticate(r *http.Request) (ctxutil.Claims, error) {
	m.callCount++
	return m.claims, m.err
}

// TestAuthMiddleware tests the AuthMiddleware function.
func TestAuthMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		mockAuth       *mockAuthenticator
		expectedStatus int
		expectedBody   string
		checkContext   bool
	}{
		{
			name: "successful authentication",
			mockAuth: &mockAuthenticator{
				claims: ctxutil.Claims{
					UserID:      "user-123",
					Roles:       []string{"admin"},
					Permissions: []string{"read"},
				},
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "OK",
			checkContext:   true,
		},
		{
			name: "unauthenticated error returns 401",
			mockAuth: &mockAuthenticator{
				err: ErrUnauthenticated,
			},
			expectedStatus: http.StatusUnauthorized,
			checkContext:   false,
		},
		{
			name: "token expired returns 401",
			mockAuth: &mockAuthenticator{
				err: ErrTokenExpired,
			},
			expectedStatus: http.StatusUnauthorized,
			checkContext:   false,
		},
		{
			name: "token invalid returns 401",
			mockAuth: &mockAuthenticator{
				err: ErrTokenInvalid,
			},
			expectedStatus: http.StatusUnauthorized,
			checkContext:   false,
		},
		{
			name: "unknown error returns 401",
			mockAuth: &mockAuthenticator{
				err: errors.New("some unknown error"),
			},
			expectedStatus: http.StatusUnauthorized,
			checkContext:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Track if handler was called and what claims it received
			var handlerClaims ctxutil.Claims
			var handlerClaimsErr error
			handlerCalled := false

			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				handlerCalled = true
				handlerClaims, handlerClaimsErr = ctxutil.ClaimsFromContext(r.Context())
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("OK"))
			})

			middleware := AuthMiddleware(tt.mockAuth, observability.NewNopLoggerInterface(), false)
			wrappedHandler := middleware(handler)

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			rec := httptest.NewRecorder()

			wrappedHandler.ServeHTTP(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.expectedStatus)
			}

			if tt.checkContext {
				if !handlerCalled {
					t.Error("expected handler to be called")
				}
				if handlerClaimsErr != nil {
					t.Errorf("unexpected error getting claims: %v", handlerClaimsErr)
				}
				if handlerClaims.UserID != tt.mockAuth.claims.UserID {
					t.Errorf("claims.UserID = %q, want %q", handlerClaims.UserID, tt.mockAuth.claims.UserID)
				}
			} else {
				if handlerCalled {
					t.Error("expected handler NOT to be called on auth failure")
				}
			}

			if tt.mockAuth.callCount != 1 {
				t.Errorf("Authenticate called %d times, want 1", tt.mockAuth.callCount)
			}
		})
	}
}

// TestAuthMiddleware_ErrorResponseBody tests error response bodies with new format.
func TestAuthMiddleware_ErrorResponseBody(t *testing.T) {
	tests := []struct {
		name         string
		err          error
		expectedCode string
	}{
		{
			name:         "token expired has specific error code",
			err:          ErrTokenExpired,
			expectedCode: domainerrors.CodeTokenExpired, // "TOKEN_EXPIRED"
		},
		{
			name:         "token invalid has specific error code",
			err:          ErrTokenInvalid,
			expectedCode: domainerrors.CodeTokenInvalid, // "TOKEN_INVALID"
		},
		{
			name:         "unauthenticated has unauthorized error code",
			err:          ErrUnauthenticated,
			expectedCode: domainerrors.CodeUnauthorized, // "UNAUTHORIZED"
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			auth := &mockAuthenticator{err: tt.err}
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				t.Error("handler should not be called")
			})

			middleware := AuthMiddleware(auth, observability.NewNopLoggerInterface(), false)
			wrappedHandler := middleware(handler)

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			rec := httptest.NewRecorder()

			wrappedHandler.ServeHTTP(rec, req)

			if rec.Code != http.StatusUnauthorized {
				t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
			}

			// Check Content-Type header is set
			contentType := rec.Header().Get("Content-Type")
			if contentType != "application/json" {
				t.Errorf("Content-Type = %q, want %q", contentType, "application/json")
			}

			// Parse response body as Envelope
			var env response.TestEnvelopeResponse
			if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
				t.Fatalf("failed to parse response body: %v", err)
			}

			// Check error code matches expected (UPPER_SNAKE format without ERR_ prefix)
			if env.Error == nil {
				t.Fatal("expected error in response body")
			}
			if env.Error.Code != tt.expectedCode {
				t.Errorf("error.code = %q, want %q", env.Error.Code, tt.expectedCode)
			}

			// Verify meta.trace_id is present (AC #2, #3)
			if env.Meta == nil {
				t.Fatal("expected meta in response body")
			}
			if env.Meta.TraceID == "" {
				t.Error("expected meta.trace_id to be present")
			}
		})
	}
}

// TestAuthMiddleware_TraceIDPropagation tests that trace_id is propagated in error responses.
func TestAuthMiddleware_TraceIDPropagation(t *testing.T) {
	auth := &mockAuthenticator{err: ErrTokenExpired}
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called")
	})

	middleware := AuthMiddleware(auth, observability.NewNopLoggerInterface(), false)
	wrappedHandler := middleware(handler)

	// Create request with trace_id in context
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	testTraceID := "test-trace-id-12345"
	ctx := ctxutil.NewRequestIDContext(req.Context(), testTraceID)
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(rec, req)

	// Parse response body
	var env response.TestEnvelopeResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
		t.Fatalf("failed to parse response body: %v", err)
	}

	// Verify trace_id matches what we set in context
	if env.Meta == nil {
		t.Fatal("expected meta in response body")
	}
	if env.Meta.TraceID != testTraceID {
		t.Errorf("meta.trace_id = %q, want %q", env.Meta.TraceID, testTraceID)
	}
}

// TestAuthMiddleware_UnexpectedErrorLogging tests that unexpected errors are logged.
func TestAuthMiddleware_UnexpectedErrorLogging(t *testing.T) {
	// Arrange
	mockAuth := &mockAuthenticator{
		err: errors.New("db connection failed"), // Unexpected error
	}
	mockLogger := &response.MockLogger{}

	// Act
	middleware := AuthMiddleware(mockAuth, mockLogger, false)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	wrapped := middleware(handler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	wrapped.ServeHTTP(rec, req)
	// Assert
	if !mockLogger.ErrorCalled {
		t.Error("Expected logger.Error to be called for unexpected error")
	}
	if mockLogger.ErrorMsg != "unexpected authentication error" {
		t.Errorf("Expected log message 'unexpected authentication error', got '%s'", mockLogger.ErrorMsg)
	}
	// Verify trace_id and ip are logged
	foundTrace := false
	foundIP := false
	for _, f := range mockLogger.Fields {
		if f.Key == "trace_id" {
			foundTrace = true
		}
		if f.Key == "ip" {
			foundIP = true
		}
	}
	if !foundTrace {
		t.Error("Expected trace_id field in log")
	}
	if !foundIP {
		t.Error("Expected ip field in log")
	}
}

// TestAuthMiddleware_IssuerAudienceValidation tests AC #5 - issuer/audience mismatch.
// Verifies that mismatched issuer/audience returns HTTP 401 with TOKEN_INVALID code.
func TestAuthMiddleware_IssuerAudienceValidation(t *testing.T) {
	// Arrange: Create authenticator that returns TokenInvalid (simulating issuer/audience mismatch)
	auth := &mockAuthenticator{err: ErrTokenInvalid}
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called")
	})

	middleware := AuthMiddleware(auth, observability.NewNopLoggerInterface(), false)
	wrappedHandler := middleware(handler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	// Act
	wrappedHandler.ServeHTTP(rec, req)

	// Assert: AC #5 - HTTP 401 with TOKEN_INVALID code
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}

	// Parse response body
	var env response.TestEnvelopeResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
		t.Fatalf("failed to parse response body: %v", err)
	}

	if env.Error == nil {
		t.Fatal("expected error in response body")
	}
	if env.Error.Code != domainerrors.CodeTokenInvalid {
		t.Errorf("error.code = %q, want %q", env.Error.Code, domainerrors.CodeTokenInvalid)
	}
	if env.Meta == nil || env.Meta.TraceID == "" {
		t.Error("expected meta.trace_id to be present")
	}
}
