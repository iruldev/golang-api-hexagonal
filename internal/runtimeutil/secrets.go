// Package runtimeutil provides runtime utility interfaces for external services.
package runtimeutil

import (
	"context"
	"errors"
	"os"
	"time"
)

// ErrSecretNotFound indicates the secret key was not found.
var ErrSecretNotFound = errors.New("secret: key not found")

// Secret represents a secret value with optional expiration.
type Secret struct {
	// Value is the secret value.
	Value string

	// ExpiresAt is when the secret expires. Zero means no expiration.
	ExpiresAt time.Time
}

// IsExpired returns true if the secret has expired.
func (s Secret) IsExpired() bool {
	if s.ExpiresAt.IsZero() {
		return false
	}
	return time.Now().After(s.ExpiresAt)
}

// SecretProvider defines secret management abstraction for swappable implementations.
// Implement this interface for Vault, AWS Secrets Manager, GCP Secret Manager, or other providers.
//
// Usage Example:
//
//	// Get a secret
//	dbPassword, err := provider.GetSecret(ctx, "DB_PASSWORD")
//	if errors.Is(err, runtimeutil.ErrSecretNotFound) {
//	    log.Fatal("DB_PASSWORD not configured")
//	}
//
//	// Get a secret with TTL (for rotation)
//	secret, err := provider.GetSecretWithTTL(ctx, "API_KEY")
//	if secret.IsExpired() {
//	    // Refresh the secret
//	}
//
// Implementing Vault Provider:
//
//	type VaultProvider struct {
//	    client *vault.Client
//	}
//
//	func (p *VaultProvider) GetSecret(ctx context.Context, key string) (string, error) {
//	    secret, err := p.client.Logical().Read(ctx, "secret/data/"+key)
//	    if err != nil || secret == nil {
//	        return "", runtimeutil.ErrSecretNotFound
//	    }
//	    return secret.Data["value"].(string), nil
//	}
type SecretProvider interface {
	// GetSecret retrieves a secret value by key.
	// Returns ErrSecretNotFound if the key does not exist.
	GetSecret(ctx context.Context, key string) (string, error)

	// GetSecretWithTTL retrieves a secret with its expiration information.
	// Useful for secret rotation scenarios.
	GetSecretWithTTL(ctx context.Context, key string) (Secret, error)
}

// EnvSecretProvider reads secrets from environment variables.
// This is the default implementation for local development.
type EnvSecretProvider struct{}

// NewEnvSecretProvider creates a new EnvSecretProvider.
func NewEnvSecretProvider() SecretProvider {
	return &EnvSecretProvider{}
}

// GetSecret retrieves a secret from environment variables.
func (p *EnvSecretProvider) GetSecret(_ context.Context, key string) (string, error) {
	val := os.Getenv(key)
	if val == "" {
		return "", ErrSecretNotFound
	}
	return val, nil
}

// GetSecretWithTTL retrieves a secret from environment variables.
// Environment secrets do not have TTL, so ExpiresAt is always zero.
func (p *EnvSecretProvider) GetSecretWithTTL(_ context.Context, key string) (Secret, error) {
	val := os.Getenv(key)
	if val == "" {
		return Secret{}, ErrSecretNotFound
	}
	return Secret{
		Value:     val,
		ExpiresAt: time.Time{}, // No expiration for env vars
	}, nil
}
