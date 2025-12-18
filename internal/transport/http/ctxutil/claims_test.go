//go:build !integration

package ctxutil

import (
	"context"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetAndGetClaims(t *testing.T) {
	ctx := context.Background()

	// Initially no claims
	assert.Nil(t, GetClaims(ctx))

	// Create claims
	claims := &Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "user-123",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	// Set claims
	ctxWithClaims := SetClaims(ctx, claims)

	// Original context still has no claims
	assert.Nil(t, GetClaims(ctx))

	// New context has claims
	gotClaims := GetClaims(ctxWithClaims)
	require.NotNil(t, gotClaims)
	assert.Equal(t, "user-123", gotClaims.Subject)
}

func TestGetClaims_NilContext(t *testing.T) {
	// GetClaims should handle context without claims gracefully
	ctx := context.Background()
	claims := GetClaims(ctx)
	assert.Nil(t, claims)
}

func TestGetClaims_WrongType(t *testing.T) {
	// If someone stores a wrong type at the same key (shouldn't happen),
	// GetClaims should return nil
	ctx := context.WithValue(context.Background(), claimsKey{}, "not-a-claims")
	claims := GetClaims(ctx)
	assert.Nil(t, claims)
}

func TestClaims_RegisteredClaimsFields(t *testing.T) {
	now := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	exp := now.Add(1 * time.Hour)

	claims := &Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "test-issuer",
			Subject:   "user-456",
			Audience:  jwt.ClaimStrings{"api-client"},
			ExpiresAt: jwt.NewNumericDate(exp),
			NotBefore: jwt.NewNumericDate(now),
			IssuedAt:  jwt.NewNumericDate(now),
			ID:        "jti-123",
		},
	}

	ctx := SetClaims(context.Background(), claims)
	got := GetClaims(ctx)

	require.NotNil(t, got)
	assert.Equal(t, "test-issuer", got.Issuer)
	assert.Equal(t, "user-456", got.Subject)
	assert.Contains(t, got.Audience, "api-client")
	assert.Equal(t, exp.Unix(), got.ExpiresAt.Unix())
	assert.Equal(t, now.Unix(), got.NotBefore.Unix())
	assert.Equal(t, now.Unix(), got.IssuedAt.Unix())
	assert.Equal(t, "jti-123", got.ID)
}
