package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Test helper - generate test JWT
func generateTestJWT(t *testing.T, claims jwt.MapClaims, secret []byte) string {
	t.Helper()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(secret)
	if err != nil {
		t.Fatalf("failed to sign test token: %v", err)
	}
	return tokenString
}

// Generate JWT with "none" algorithm (for security testing)
func generateNoneAlgorithmJWT(t *testing.T, claims jwt.MapClaims) string {
	t.Helper()
	token := jwt.NewWithClaims(jwt.SigningMethodNone, claims)
	tokenString, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
	if err != nil {
		t.Fatalf("failed to sign none algorithm token: %v", err)
	}
	return tokenString
}

func TestNewJWTAuthenticator(t *testing.T) {
	secret := []byte("test-secret-key-at-least-32-bytes!!")

	t.Run("creates authenticator with valid secret", func(t *testing.T) {
		auth, err := NewJWTAuthenticator(secret)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if auth == nil {
			t.Fatal("expected non-nil authenticator")
		}
		if string(auth.config.SecretKey) != string(secret) {
			t.Error("secret key not set correctly")
		}
	})

	t.Run("returns error for short secret key", func(t *testing.T) {
		shortSecret := []byte("short")
		auth, err := NewJWTAuthenticator(shortSecret)
		if err != ErrSecretKeyTooShort {
			t.Errorf("expected ErrSecretKeyTooShort, got %v", err)
		}
		if auth != nil {
			t.Error("expected nil authenticator for short key")
		}
	})

	t.Run("returns error for nil secret key", func(t *testing.T) {
		auth, err := NewJWTAuthenticator(nil)
		if err != ErrSecretKeyTooShort {
			t.Errorf("expected ErrSecretKeyTooShort, got %v", err)
		}
		if auth != nil {
			t.Error("expected nil authenticator for nil key")
		}
	})

	t.Run("returns error for empty secret key", func(t *testing.T) {
		auth, err := NewJWTAuthenticator([]byte{})
		if err != ErrSecretKeyTooShort {
			t.Errorf("expected ErrSecretKeyTooShort, got %v", err)
		}
		if auth != nil {
			t.Error("expected nil authenticator for empty key")
		}
	})

	t.Run("accepts exactly 32 byte secret", func(t *testing.T) {
		exact32 := []byte("12345678901234567890123456789012") // exactly 32 bytes
		auth, err := NewJWTAuthenticator(exact32)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if auth == nil {
			t.Fatal("expected non-nil authenticator")
		}
	})

	t.Run("applies WithIssuer option", func(t *testing.T) {
		auth, err := NewJWTAuthenticator(secret, WithIssuer("test-issuer"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if auth.config.Issuer != "test-issuer" {
			t.Errorf("expected issuer 'test-issuer', got '%s'", auth.config.Issuer)
		}
	})

	t.Run("applies WithAudience option", func(t *testing.T) {
		auth, err := NewJWTAuthenticator(secret, WithAudience("test-audience"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if auth.config.Audience != "test-audience" {
			t.Errorf("expected audience 'test-audience', got '%s'", auth.config.Audience)
		}
	})

	t.Run("applies multiple options", func(t *testing.T) {
		auth, err := NewJWTAuthenticator(secret,
			WithIssuer("my-issuer"),
			WithAudience("my-audience"),
		)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if auth.config.Issuer != "my-issuer" {
			t.Errorf("expected issuer 'my-issuer', got '%s'", auth.config.Issuer)
		}
		if auth.config.Audience != "my-audience" {
			t.Errorf("expected audience 'my-audience', got '%s'", auth.config.Audience)
		}
	})
}

func TestJWTAuthenticator_Authenticate(t *testing.T) {
	secret := []byte("test-secret-key-at-least-32-bytes!!")
	auth, err := NewJWTAuthenticator(secret)
	if err != nil {
		t.Fatalf("failed to create authenticator: %v", err)
	}

	tests := []struct {
		name       string
		authHeader string
		wantErr    error
		wantUserID string
		wantRoles  []string
		wantPerms  []string
	}{
		{
			name: "valid token with all claims",
			authHeader: "Bearer " + generateTestJWT(t, jwt.MapClaims{
				"sub":         "user-123",
				"roles":       []interface{}{"admin", "user"},
				"permissions": []interface{}{"read", "write"},
				"exp":         time.Now().Add(time.Hour).Unix(),
			}, secret),
			wantErr:    nil,
			wantUserID: "user-123",
			wantRoles:  []string{"admin", "user"},
			wantPerms:  []string{"read", "write"},
		},
		{
			name: "valid token with only sub claim",
			authHeader: "Bearer " + generateTestJWT(t, jwt.MapClaims{
				"sub": "user-456",
				"exp": time.Now().Add(time.Hour).Unix(),
			}, secret),
			wantErr:    nil,
			wantUserID: "user-456",
			wantRoles:  nil,
			wantPerms:  nil,
		},
		{
			name:       "missing authorization header",
			authHeader: "",
			wantErr:    ErrUnauthenticated,
		},
		{
			name:       "missing Bearer prefix",
			authHeader: "Basic abc123",
			wantErr:    ErrUnauthenticated,
		},
		{
			name:       "Bearer prefix only (empty token)",
			authHeader: "Bearer ",
			wantErr:    ErrUnauthenticated,
		},
		{
			name: "expired token",
			authHeader: "Bearer " + generateTestJWT(t, jwt.MapClaims{
				"sub": "user-789",
				"exp": time.Now().Add(-time.Hour).Unix(),
			}, secret),
			wantErr: ErrTokenExpired,
		},
		{
			name: "invalid signature (wrong secret)",
			authHeader: "Bearer " + generateTestJWT(t, jwt.MapClaims{
				"sub": "user-123",
				"exp": time.Now().Add(time.Hour).Unix(),
			}, []byte("wrong-secret-key-totally-different")),
			wantErr: ErrTokenInvalid,
		},
		{
			name:       "malformed token (not JWT format)",
			authHeader: "Bearer not.a.valid.jwt.token.at.all",
			wantErr:    ErrTokenInvalid,
		},
		{
			name:       "completely invalid token",
			authHeader: "Bearer garbage",
			wantErr:    ErrTokenInvalid,
		},
		{
			name: "none algorithm token (security)",
			authHeader: "Bearer " + generateNoneAlgorithmJWT(t, jwt.MapClaims{
				"sub": "attacker",
				"exp": time.Now().Add(time.Hour).Unix(),
			}),
			wantErr: ErrTokenInvalid,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			// Act
			claims, err := auth.Authenticate(req)

			// Assert
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("expected error %v, got %v", tt.wantErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if claims.UserID != tt.wantUserID {
				t.Errorf("expected UserID '%s', got '%s'", tt.wantUserID, claims.UserID)
			}

			if len(claims.Roles) != len(tt.wantRoles) {
				t.Errorf("expected %d roles, got %d", len(tt.wantRoles), len(claims.Roles))
			} else {
				for i, role := range tt.wantRoles {
					if claims.Roles[i] != role {
						t.Errorf("expected role[%d] '%s', got '%s'", i, role, claims.Roles[i])
					}
				}
			}

			if len(claims.Permissions) != len(tt.wantPerms) {
				t.Errorf("expected %d permissions, got %d", len(tt.wantPerms), len(claims.Permissions))
			} else {
				for i, perm := range tt.wantPerms {
					if claims.Permissions[i] != perm {
						t.Errorf("expected permission[%d] '%s', got '%s'", i, perm, claims.Permissions[i])
					}
				}
			}
		})
	}
}

func TestJWTAuthenticator_IssuerValidation(t *testing.T) {
	secret := []byte("test-secret-key-at-least-32-bytes!!")
	auth, err := NewJWTAuthenticator(secret, WithIssuer("my-app"))
	if err != nil {
		t.Fatalf("failed to create authenticator: %v", err)
	}

	t.Run("valid issuer", func(t *testing.T) {
		token := generateTestJWT(t, jwt.MapClaims{
			"sub": "user-1",
			"iss": "my-app",
			"exp": time.Now().Add(time.Hour).Unix(),
		}, secret)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		claims, err := auth.Authenticate(req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if claims.UserID != "user-1" {
			t.Errorf("expected UserID 'user-1', got '%s'", claims.UserID)
		}
	})

	t.Run("invalid issuer", func(t *testing.T) {
		token := generateTestJWT(t, jwt.MapClaims{
			"sub": "user-1",
			"iss": "wrong-issuer",
			"exp": time.Now().Add(time.Hour).Unix(),
		}, secret)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		_, err := auth.Authenticate(req)
		if !errors.Is(err, ErrTokenInvalid) {
			t.Errorf("expected ErrTokenInvalid, got %v", err)
		}
	})
}

func TestJWTAuthenticator_AudienceValidation(t *testing.T) {
	secret := []byte("test-secret-key-at-least-32-bytes!!")
	auth, err := NewJWTAuthenticator(secret, WithAudience("my-api"))
	if err != nil {
		t.Fatalf("failed to create authenticator: %v", err)
	}

	t.Run("valid audience", func(t *testing.T) {
		token := generateTestJWT(t, jwt.MapClaims{
			"sub": "user-1",
			"aud": "my-api",
			"exp": time.Now().Add(time.Hour).Unix(),
		}, secret)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		claims, err := auth.Authenticate(req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if claims.UserID != "user-1" {
			t.Errorf("expected UserID 'user-1', got '%s'", claims.UserID)
		}
	})

	t.Run("invalid audience", func(t *testing.T) {
		token := generateTestJWT(t, jwt.MapClaims{
			"sub": "user-1",
			"aud": "wrong-audience",
			"exp": time.Now().Add(time.Hour).Unix(),
		}, secret)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		_, err := auth.Authenticate(req)
		if !errors.Is(err, ErrTokenInvalid) {
			t.Errorf("expected ErrTokenInvalid, got %v", err)
		}
	})
}

func TestMapJWTClaims(t *testing.T) {
	tests := []struct {
		name         string
		jwtClaims    jwt.MapClaims
		wantID       string
		wantRoles    []string
		wantPerms    []string
		wantMetadata map[string]string
	}{
		{
			name: "all claims present",
			jwtClaims: jwt.MapClaims{
				"sub":         "user-abc",
				"roles":       []interface{}{"admin", "editor"},
				"permissions": []interface{}{"create", "delete"},
			},
			wantID:       "user-abc",
			wantRoles:    []string{"admin", "editor"},
			wantPerms:    []string{"create", "delete"},
			wantMetadata: map[string]string{},
		},
		{
			name: "all claims with metadata",
			jwtClaims: jwt.MapClaims{
				"sub":         "user-meta",
				"roles":       []interface{}{"user"},
				"permissions": []interface{}{"read"},
				"metadata": map[string]interface{}{
					"org":    "acme",
					"tenant": "prod",
				},
			},
			wantID:       "user-meta",
			wantRoles:    []string{"user"},
			wantPerms:    []string{"read"},
			wantMetadata: map[string]string{"org": "acme", "tenant": "prod"},
		},
		{
			name: "only sub claim",
			jwtClaims: jwt.MapClaims{
				"sub": "user-xyz",
			},
			wantID:       "user-xyz",
			wantRoles:    nil,
			wantPerms:    nil,
			wantMetadata: map[string]string{},
		},
		{
			name:         "empty claims",
			jwtClaims:    jwt.MapClaims{},
			wantID:       "",
			wantRoles:    nil,
			wantPerms:    nil,
			wantMetadata: map[string]string{},
		},
		{
			name: "non-string sub claim ignored",
			jwtClaims: jwt.MapClaims{
				"sub": 12345,
			},
			wantID:       "",
			wantRoles:    nil,
			wantPerms:    nil,
			wantMetadata: map[string]string{},
		},
		{
			name: "roles with non-string elements skipped",
			jwtClaims: jwt.MapClaims{
				"sub":   "user-1",
				"roles": []interface{}{"admin", 123, "user"},
			},
			wantID:       "user-1",
			wantRoles:    []string{"admin", "user"},
			wantPerms:    nil,
			wantMetadata: map[string]string{},
		},
		{
			name: "metadata with non-string values skipped",
			jwtClaims: jwt.MapClaims{
				"sub": "user-2",
				"metadata": map[string]interface{}{
					"valid":   "value",
					"invalid": 12345,
				},
			},
			wantID:       "user-2",
			wantRoles:    nil,
			wantPerms:    nil,
			wantMetadata: map[string]string{"valid": "value"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims := mapJWTClaims(tt.jwtClaims)

			if claims.UserID != tt.wantID {
				t.Errorf("expected UserID '%s', got '%s'", tt.wantID, claims.UserID)
			}

			if len(claims.Roles) != len(tt.wantRoles) {
				t.Errorf("expected %d roles, got %d", len(tt.wantRoles), len(claims.Roles))
			}

			if len(claims.Permissions) != len(tt.wantPerms) {
				t.Errorf("expected %d permissions, got %d", len(tt.wantPerms), len(claims.Permissions))
			}

			// Verify Metadata is initialized
			if claims.Metadata == nil {
				t.Error("expected Metadata to be initialized, got nil")
			}

			// Verify Metadata contents
			if len(claims.Metadata) != len(tt.wantMetadata) {
				t.Errorf("expected %d metadata entries, got %d", len(tt.wantMetadata), len(claims.Metadata))
			}
			for k, v := range tt.wantMetadata {
				if claims.Metadata[k] != v {
					t.Errorf("expected Metadata[%q] = %q, got %q", k, v, claims.Metadata[k])
				}
			}
		})
	}
}

func TestJWTAuthenticator_ImplementsAuthenticator(t *testing.T) {
	// Compile-time check that JWTAuthenticator implements Authenticator
	var _ Authenticator = (*JWTAuthenticator)(nil)
}
