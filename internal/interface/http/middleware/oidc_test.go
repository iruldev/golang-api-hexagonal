package middleware

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/golang-jwt/jwt/v5"
	"github.com/iruldev/golang-api-hexagonal/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockKeySet implements oidc.KeySet for testing.
type MockKeySet struct {
	PublicKey *rsa.PublicKey
}

func (k *MockKeySet) VerifySignature(ctx context.Context, rawToken string) ([]byte, error) {
	// Simple verification using the public key
	// In reality, we'd parse the token and verify signature.
	// Here we reuse jwt-go to verify implementation correctness.
	token, err := jwt.Parse(rawToken, func(token *jwt.Token) (interface{}, error) {
		return k.PublicKey, nil
	})
	if err != nil {
		return nil, err
	}

	// Convert claims to JSON bytes as expected by go-oidc
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return json.Marshal(claims)
	}

	return nil, assert.AnError
}

func TestOIDCAuthenticator_Authenticate(t *testing.T) {
	// 1. Setup Test Keys
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	// 2. Setup Mock KeySet and Verifier
	keySet := &MockKeySet{PublicKey: &privateKey.PublicKey}

	cfg := config.OIDCConfig{
		Enabled:   true,
		IssuerURL: "https://test-issuer.com",
		ClientID:  "test-client",
	}

	// Ensure checking issuer matches
	oidcConfig := &oidc.Config{
		ClientID:          cfg.ClientID,
		SkipClientIDCheck: false,
		SkipExpiryCheck:   false,
		SkipIssuerCheck:   false,
	}
	verifier := oidc.NewVerifier(cfg.IssuerURL, keySet, oidcConfig)

	auth := NewOIDCAuthenticator(cfg, verifier)

	t.Run("Success", func(t *testing.T) {
		// Create Token
		claims := jwt.MapClaims{
			"iss": cfg.IssuerURL,
			"aud": cfg.ClientID,
			"sub": "user-123",
			"exp": time.Now().Add(time.Hour).Unix(),
			"realm_access": map[string]interface{}{
				"roles": []string{"admin", "editor"},
			},
		}
		token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
		signedToken, err := token.SignedString(privateKey)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer "+signedToken)

		gotClaims, err := auth.Authenticate(req)

		require.NoError(t, err)
		assert.Equal(t, "user-123", gotClaims.UserID)
		assert.Contains(t, gotClaims.Roles, "admin")
		assert.Contains(t, gotClaims.Roles, "editor")
	})

	t.Run("No Token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		// No Authorization header

		_, err = auth.Authenticate(req)
		assert.ErrorIs(t, err, ErrUnauthenticated)
	})

	t.Run("Invalid Token Format", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer invalid-token")

		_, err = auth.Authenticate(req)
		assert.ErrorIs(t, err, ErrTokenInvalid) // or unauthenticated depending on implementation
	})

	t.Run("Expired Token", func(t *testing.T) {
		claims := jwt.MapClaims{
			"iss": cfg.IssuerURL,
			"aud": cfg.ClientID,
			"sub": "user-123",
			"exp": time.Now().Add(-time.Hour).Unix(), // Expired
		}
		token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
		signedToken, err := token.SignedString(privateKey)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer "+signedToken)

		_, err = auth.Authenticate(req)
		// go-oidc returns specific errors, we should map them or checking error string
		assert.Error(t, err)
		// Ideally we verify it maps to ErrTokenExpired if we implement that mapping
	})

	t.Run("Wrong Issuer", func(t *testing.T) {
		claims := jwt.MapClaims{
			"iss": "https://attacker.com",
			"aud": cfg.ClientID,
			"sub": "user-123",
			"exp": time.Now().Add(time.Hour).Unix(),
		}
		token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
		signedToken, err := token.SignedString(privateKey)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer "+signedToken)

		_, err = auth.Authenticate(req)
		assert.Error(t, err)
	})

	t.Run("Custom Roles Claim", func(t *testing.T) {
		cfgWithRole := cfg
		cfgWithRole.RolesClaim = "custom_roles"
		authWithRole := NewOIDCAuthenticator(cfgWithRole, verifier)

		claims := jwt.MapClaims{
			"iss":          cfg.IssuerURL,
			"aud":          cfg.ClientID,
			"sub":          "user-custom",
			"exp":          time.Now().Add(time.Hour).Unix(),
			"custom_roles": []string{"custom-admin"},
		}
		token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
		signedToken, err := token.SignedString(privateKey)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer "+signedToken)

		gotClaims, err := authWithRole.Authenticate(req)
		require.NoError(t, err)
		assert.Equal(t, "user-custom", gotClaims.UserID)
		assert.Contains(t, gotClaims.Roles, "custom-admin")
	})

	t.Run("Roles as Space-Separated String", func(t *testing.T) {
		claims := jwt.MapClaims{
			"iss": cfg.IssuerURL,
			"aud": cfg.ClientID,
			"sub": "user-string",
			"exp": time.Now().Add(time.Hour).Unix(),
			"realm_access": map[string]interface{}{
				"roles": "admin editor subscriber",
			},
		}
		token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
		signedToken, err := token.SignedString(privateKey)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer "+signedToken)

		gotClaims, err := auth.Authenticate(req)
		require.NoError(t, err)
		assert.Len(t, gotClaims.Roles, 3)
		assert.Contains(t, gotClaims.Roles, "subscriber")
	})

	t.Run("Audience Validation", func(t *testing.T) {
		cfgAud := cfg
		cfgAud.Audience = []string{"my-api", "other-api"}

		// Create a verifier that skips ClientID check, mimicking main.go logic when Audience is set
		oidcConfigAud := &oidc.Config{
			ClientID:          cfg.ClientID,
			SkipClientIDCheck: true,
			SkipExpiryCheck:   false,
			SkipIssuerCheck:   false,
		}
		verifierAud := oidc.NewVerifier(cfgAud.IssuerURL, keySet, oidcConfigAud)

		authAud := NewOIDCAuthenticator(cfgAud, verifierAud)

		// Case 1: Valid Audience
		claims := jwt.MapClaims{
			"iss": cfg.IssuerURL,
			"aud": []string{"my-api"}, // Matches one
			"sub": "user-aud",
			"exp": time.Now().Add(time.Hour).Unix(),
		}
		token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
		signedToken, err := token.SignedString(privateKey)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer "+signedToken)
		_, err = authAud.Authenticate(req)
		require.NoError(t, err)

		// Case 2: Invalid Audience
		claimsInvalid := jwt.MapClaims{
			"iss": cfg.IssuerURL,
			"aud": []string{"wrong-api"},
			"sub": "user-aud",
			"exp": time.Now().Add(time.Hour).Unix(),
		}
		tokenInvalid := jwt.NewWithClaims(jwt.SigningMethodRS256, claimsInvalid)
		signedTokenInvalid, err := tokenInvalid.SignedString(privateKey)
		require.NoError(t, err)

		reqInvalid := httptest.NewRequest(http.MethodGet, "/", nil)
		reqInvalid.Header.Set("Authorization", "Bearer "+signedTokenInvalid)
		_, err = authAud.Authenticate(reqInvalid)
		assert.ErrorIs(t, err, ErrTokenInvalid)
	})
}
