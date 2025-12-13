package middleware

import (
	"context"
	"errors"
	"net/http"
	"os"
	"strings"
)

// DefaultAPIKeyHeader is the default header name for API key authentication.
const DefaultAPIKeyHeader = "X-API-Key"

// ErrValidatorRequired indicates that a KeyValidator is required.
var ErrValidatorRequired = errors.New("validator is required")

// KeyInfo contains information about a validated API key.
// This struct is returned by KeyValidator implementations after successful validation.
type KeyInfo struct {
	// ServiceID is the unique identifier for the service/client.
	ServiceID string

	// Roles are the roles assigned to this API key (e.g., "service", "admin").
	Roles []string

	// Permissions are specific permissions granted to this key.
	Permissions []string

	// Metadata contains additional key-value pairs associated with the key.
	Metadata map[string]string
}

// KeyValidator validates API keys and returns service information.
// Implementations can use environment variables, database, or external services.
//
// Example implementations:
//   - EnvKeyValidator: Validates against environment variables
//   - MapKeyValidator: Validates against in-memory map (for testing)
//   - DBKeyValidator: Validates against database (custom implementation)
type KeyValidator interface {
	// Validate checks if the API key is valid and returns service info.
	// Returns nil KeyInfo and error if key is invalid.
	Validate(ctx context.Context, key string) (*KeyInfo, error)
}

// APIKeyConfig holds API key authenticator configuration.
type APIKeyConfig struct {
	// HeaderName is the HTTP header to read the API key from.
	// Default: "X-API-Key"
	HeaderName string

	// Validator is the key validation implementation.
	Validator KeyValidator
}

// APIKeyOption configures the API key authenticator.
type APIKeyOption func(*APIKeyConfig)

// WithHeaderName sets a custom header name for the API key.
//
// Example:
//
//	apiAuth, err := middleware.NewAPIKeyAuthenticator(validator,
//	    middleware.WithHeaderName("Authorization"),
//	)
func WithHeaderName(name string) APIKeyOption {
	return func(c *APIKeyConfig) {
		c.HeaderName = name
	}
}

// APIKeyAuthenticator validates API keys from request headers.
// It implements the Authenticator interface.
//
// The authenticator extracts the API key from a configurable header
// (default: X-API-Key) and validates it using the provided KeyValidator.
//
// Example usage:
//
//	// Create validator from environment variables
//	validator := middleware.NewEnvKeyValidator("API_KEYS")
//
//	// Create authenticator
//	apiAuth, err := middleware.NewAPIKeyAuthenticator(validator)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Use with AuthMiddleware
//	router.Group(func(r chi.Router) {
//	    r.Use(middleware.AuthMiddleware(apiAuth))
//	    r.Get("/api/v1/internal", internalHandler)
//	})
type APIKeyAuthenticator struct {
	config APIKeyConfig
}

// NewAPIKeyAuthenticator creates a new API key authenticator.
//
// The validator is required and cannot be nil.
// Returns ErrValidatorRequired if validator is nil.
// Options can be used to configure the header name.
//
// Example:
//
//	validator := middleware.NewEnvKeyValidator("API_KEYS")
//	apiAuth, err := middleware.NewAPIKeyAuthenticator(validator,
//	    middleware.WithHeaderName("X-Custom-Key"),
//	)
//	if err != nil {
//	    log.Fatal(err)
//	}
func NewAPIKeyAuthenticator(validator KeyValidator, opts ...APIKeyOption) (*APIKeyAuthenticator, error) {
	if validator == nil {
		return nil, ErrValidatorRequired
	}

	config := APIKeyConfig{
		HeaderName: DefaultAPIKeyHeader,
		Validator:  validator,
	}

	for _, opt := range opts {
		opt(&config)
	}

	return &APIKeyAuthenticator{config: config}, nil
}

// Authenticate implements the Authenticator interface.
// It extracts and validates an API key from the configured header.
//
// SECURITY WARNING: Never log the API key value. Use key identifiers for debugging.
//
// Returns:
//   - Claims: The extracted service claims if authentication succeeds
//   - ErrUnauthenticated: If API key header is missing or empty
//   - ErrTokenInvalid: If API key is invalid, revoked, or validator returns nil
func (a *APIKeyAuthenticator) Authenticate(r *http.Request) (Claims, error) {
	// 1. Extract API key from header
	// SECURITY: Do not log apiKey value
	apiKey := r.Header.Get(a.config.HeaderName)
	if apiKey == "" {
		return Claims{}, ErrUnauthenticated
	}

	// 2. Validate key using validator
	keyInfo, err := a.config.Validator.Validate(r.Context(), apiKey)
	if err != nil {
		return Claims{}, ErrTokenInvalid
	}

	// 3. Defensive check: validator returned nil KeyInfo without error
	if keyInfo == nil {
		return Claims{}, ErrTokenInvalid
	}

	// 4. Map KeyInfo to Claims
	return mapKeyInfoToClaims(keyInfo), nil
}

// mapKeyInfoToClaims converts KeyInfo to middleware.Claims.
func mapKeyInfoToClaims(info *KeyInfo) Claims {
	roles := info.Roles
	// Default to "service" role if no roles specified
	if len(roles) == 0 {
		roles = []string{"service"}
	}

	// Initialize metadata map consistently with JWT implementation
	metadata := info.Metadata
	if metadata == nil {
		metadata = make(map[string]string)
	}

	return Claims{
		UserID:      info.ServiceID,
		Roles:       roles,
		Permissions: info.Permissions,
		Metadata:    metadata,
	}
}

// =============================================================================
// Built-in Validators
// =============================================================================

// EnvKeyValidator validates API keys from an environment variable.
// Keys are stored as comma-separated "key:service_id" pairs.
//
// Format: "key1:service1,key2:service2"
//
// Example:
//
//	# Set environment variable
//	export API_KEYS="abc123:svc-payments,xyz789:svc-inventory"
//
//	# Create validator
//	validator := middleware.NewEnvKeyValidator("API_KEYS")
type EnvKeyValidator struct {
	keys map[string]string // key -> serviceID
}

// NewEnvKeyValidator creates a validator from an environment variable.
// The environment variable should contain comma-separated "key:service_id" pairs.
//
// Example:
//
//	// Environment: API_KEYS="abc123:svc-payments,xyz789:svc-inventory"
//	validator := middleware.NewEnvKeyValidator("API_KEYS")
func NewEnvKeyValidator(envVar string) *EnvKeyValidator {
	v := &EnvKeyValidator{keys: make(map[string]string)}

	rawKeys := os.Getenv(envVar)
	if rawKeys == "" {
		return v
	}

	for _, pair := range strings.Split(rawKeys, ",") {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}
		parts := strings.SplitN(pair, ":", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			serviceID := strings.TrimSpace(parts[1])
			if key != "" && serviceID != "" {
				v.keys[key] = serviceID
			}
		}
	}

	return v
}

// Validate implements KeyValidator interface.
func (v *EnvKeyValidator) Validate(ctx context.Context, key string) (*KeyInfo, error) {
	serviceID, ok := v.keys[key]
	if !ok {
		return nil, ErrTokenInvalid
	}
	return &KeyInfo{
		ServiceID: serviceID,
		Roles:     []string{"service"},
	}, nil
}

// KeyCount returns the number of keys loaded from the environment variable.
// Useful for debugging and validating that keys were parsed correctly at startup.
func (v *EnvKeyValidator) KeyCount() int {
	return len(v.keys)
}

// MapKeyValidator validates API keys against an in-memory map.
// This is primarily intended for testing purposes.
//
// Example:
//
//	validator := &middleware.MapKeyValidator{
//	    Keys: map[string]*middleware.KeyInfo{
//	        "test-key": {ServiceID: "test-service", Roles: []string{"service"}},
//	    },
//	}
type MapKeyValidator struct {
	Keys map[string]*KeyInfo
}

// Validate implements KeyValidator interface.
func (v *MapKeyValidator) Validate(ctx context.Context, key string) (*KeyInfo, error) {
	info, ok := v.Keys[key]
	if !ok {
		return nil, ErrTokenInvalid
	}
	return info, nil
}
