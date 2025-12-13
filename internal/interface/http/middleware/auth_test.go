package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestClaims_HasRole tests the HasRole method of Claims struct.
func TestClaims_HasRole(t *testing.T) {
	tests := []struct {
		name     string
		claims   Claims
		role     string
		expected bool
	}{
		{
			name:     "has role returns true",
			claims:   Claims{Roles: []string{"admin", "user"}},
			role:     "admin",
			expected: true,
		},
		{
			name:     "does not have role returns false",
			claims:   Claims{Roles: []string{"user"}},
			role:     "admin",
			expected: false,
		},
		{
			name:     "empty roles returns false",
			claims:   Claims{Roles: nil},
			role:     "admin",
			expected: false,
		},
		{
			name:     "empty role string returns false",
			claims:   Claims{Roles: []string{"admin", "user"}},
			role:     "",
			expected: false,
		},
		{
			name:     "exactly matches role",
			claims:   Claims{Roles: []string{"administrator"}},
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

// TestClaims_HasPermission tests the HasPermission method of Claims struct.
func TestClaims_HasPermission(t *testing.T) {
	tests := []struct {
		name       string
		claims     Claims
		permission string
		expected   bool
	}{
		{
			name:       "has permission returns true",
			claims:     Claims{Permissions: []string{"read", "write", "delete"}},
			permission: "write",
			expected:   true,
		},
		{
			name:       "does not have permission returns false",
			claims:     Claims{Permissions: []string{"read"}},
			permission: "write",
			expected:   false,
		},
		{
			name:       "empty permissions returns false",
			claims:     Claims{Permissions: nil},
			permission: "read",
			expected:   false,
		},
		{
			name:       "empty permission string",
			claims:     Claims{Permissions: []string{"read", "write"}},
			permission: "",
			expected:   false,
		},
		{
			name:       "partial match returns false",
			claims:     Claims{Permissions: []string{"read:notes"}},
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
	claims := Claims{
		UserID:      "user-123",
		Roles:       []string{"admin"},
		Permissions: []string{"read", "write"},
	}

	ctx := NewContext(context.Background(), claims)

	// Verify claims are stored
	value := ctx.Value(claimsKey)
	if value == nil {
		t.Fatal("expected claims to be stored in context")
	}

	storedClaims, ok := value.(Claims)
	if !ok {
		t.Fatal("expected value to be Claims type")
	}

	if storedClaims.UserID != claims.UserID {
		t.Errorf("UserID = %q, want %q", storedClaims.UserID, claims.UserID)
	}
}

// TestFromContext tests extracting claims from context.
func TestFromContext(t *testing.T) {
	t.Run("returns claims when present", func(t *testing.T) {
		expectedClaims := Claims{
			UserID:      "user-456",
			Roles:       []string{"user", "editor"},
			Permissions: []string{"read", "write:notes"},
			Metadata:    map[string]string{"org": "acme"},
		}

		ctx := NewContext(context.Background(), expectedClaims)
		claims, err := FromContext(ctx)

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
		_, err := FromContext(ctx)

		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !errors.Is(err, ErrNoClaimsInContext) {
			t.Errorf("error = %v, want %v", err, ErrNoClaimsInContext)
		}
	})

	t.Run("returns error for wrong type in context", func(t *testing.T) {
		// Store something other than Claims
		ctx := context.WithValue(context.Background(), claimsKey, "not-claims")
		_, err := FromContext(ctx)

		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !errors.Is(err, ErrNoClaimsInContext) {
			t.Errorf("error = %v, want %v", err, ErrNoClaimsInContext)
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
		{"ErrNoClaimsInContext", ErrNoClaimsInContext, "no claims in context"},
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
	claims    Claims
	err       error
	callCount int
}

func (m *mockAuthenticator) Authenticate(r *http.Request) (Claims, error) {
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
				claims: Claims{
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
			var handlerClaims Claims
			var handlerClaimsErr error
			handlerCalled := false

			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				handlerCalled = true
				handlerClaims, handlerClaimsErr = FromContext(r.Context())
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("OK"))
			})

			middleware := AuthMiddleware(tt.mockAuth)
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

// TestAuthMiddleware_ErrorResponseBody tests error response bodies.
func TestAuthMiddleware_ErrorResponseBody(t *testing.T) {
	tests := []struct {
		name         string
		err          error
		expectedCode string
	}{
		{
			name:         "token expired has specific error code",
			err:          ErrTokenExpired,
			expectedCode: "ERR_TOKEN_EXPIRED",
		},
		{
			name:         "token invalid has specific error code",
			err:          ErrTokenInvalid,
			expectedCode: "ERR_TOKEN_INVALID",
		},
		{
			name:         "unauthenticated has generic error code",
			err:          ErrUnauthenticated,
			expectedCode: "ERR_UNAUTHORIZED",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			auth := &mockAuthenticator{err: tt.err}
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				t.Error("handler should not be called")
			})

			middleware := AuthMiddleware(auth)
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

			// Check response body contains expected error code
			body := rec.Body.String()
			if body == "" {
				t.Error("expected non-empty response body")
			}
			// The body should contain the error code (JSON formatted)
			if !strings.Contains(body, tt.expectedCode) {
				t.Errorf("body should contain %q, got %q", tt.expectedCode, body)
			}
		})
	}
}
