package middleware

import (
	"net/http"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/iruldev/golang-api-hexagonal/internal/config"
)

// OIDCAuthenticator implements Authenticator using OpenID Connect.
type OIDCAuthenticator struct {
	verifier *oidc.IDTokenVerifier
	config   config.OIDCConfig
}

// NewOIDCAuthenticator creates a new OIDCAuthenticator with a specific verifier.
// This allows injection of a verifier with a custom KeySet for testing.
func NewOIDCAuthenticator(cfg config.OIDCConfig, verifier *oidc.IDTokenVerifier) *OIDCAuthenticator {
	return &OIDCAuthenticator{
		config:   cfg,
		verifier: verifier,
	}
}

// Authenticate validates the request using OIDC.
func (a *OIDCAuthenticator) Authenticate(r *http.Request) (Claims, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return Claims{}, ErrUnauthenticated
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return Claims{}, ErrUnauthenticated
	}

	tokenStr := parts[1]

	// Verify options with multi-audience support
	// Note: go-oidc Verifier uses the config it was created with.
	// We rely on the verifier injected in NewOIDCAuthenticator being correctly configured.

	idToken, err := a.verifier.Verify(r.Context(), tokenStr)
	if err != nil {
		// Simple string matching for error mapping as go-oidc error types aren't always exposed
		if strings.Contains(err.Error(), "expired") || strings.Contains(err.Error(), "exp not satisfied") {
			return Claims{}, ErrTokenExpired
		}
		return Claims{}, ErrTokenInvalid
	}

	// Manual Audience Validation if multiple audiences configured
	if len(a.config.Audience) > 0 {
		if !containsAny(idToken.Audience, a.config.Audience) {
			return Claims{}, ErrTokenInvalid
		}
	}

	// Extract all claims into a map to support flexible mapping
	var allClaims map[string]interface{}
	if err := idToken.Claims(&allClaims); err != nil {
		return Claims{}, ErrTokenInvalid
	}

	// Extract sub (UserID)
	sub, ok := allClaims["sub"].(string)
	if !ok {
		return Claims{}, ErrTokenInvalid
	}

	// Extract Roles based on configuration
	roles := a.extractRoles(allClaims)

	return Claims{
		UserID: sub,
		Roles:  roles,
	}, nil
}

// extractRoles extracts roles from the claims map based on the configured RolesClaim.
// It supports nested keys via dot notation (e.g., "realm_access.roles").
func (a *OIDCAuthenticator) extractRoles(claims map[string]interface{}) []string {
	claimName := a.config.RolesClaim
	if claimName == "" {
		claimName = "realm_access.roles" // Default to Keycloak standard
	}

	// Traverse nested map if needed
	val, found := getPath(claims, claimName)
	if !found {
		// Fallback to standard "roles" if configured claim not found and default was used
		if claimName == "realm_access.roles" {
			if v, ok := claims["roles"]; ok {
				return toCheck(v)
			}
		}
		return nil
	}

	return toCheck(val)
}

// getPath traverses a map using dot notation
func getPath(data map[string]interface{}, path string) (interface{}, bool) {
	parts := strings.Split(path, ".")
	var current interface{} = data

	for _, part := range parts {
		m, ok := current.(map[string]interface{})
		if !ok {
			return nil, false
		}
		val, ok := m[part]
		if !ok {
			return nil, false
		}
		current = val
	}
	return current, true
}

// toCheck converts an interface value to []string.
// Supports []interface{}, []string, or space-separated string.
func toCheck(val interface{}) []string {
	switch v := val.(type) {
	case []interface{}:
		var roles []string
		for _, item := range v {
			if s, ok := item.(string); ok {
				roles = append(roles, s)
			}
		}
		return roles
	case []string:
		return v
	case string:
		return strings.Fields(v) // Handle space-separated roles
	default:
		return nil
	}
}
