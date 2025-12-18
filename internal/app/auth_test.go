//go:build !integration

package app

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRoleConstants(t *testing.T) {
	// Verify role constants are defined correctly
	assert.Equal(t, "admin", RoleAdmin)
	assert.Equal(t, "user", RoleUser)
}

func TestErrNoAuthContext(t *testing.T) {
	// Verify error is defined
	require.NotNil(t, ErrNoAuthContext)
	assert.Equal(t, "no authentication context", ErrNoAuthContext.Error())
}

func TestSetAndGetAuthContext(t *testing.T) {
	ctx := context.Background()

	// Initially no auth context
	assert.Nil(t, GetAuthContext(ctx))

	// Create auth context
	authCtx := &AuthContext{
		SubjectID: "user-123",
		Role:      RoleAdmin,
	}

	// Set auth context
	ctxWithAuth := SetAuthContext(ctx, authCtx)

	// Original context still has no auth context
	assert.Nil(t, GetAuthContext(ctx))

	// New context has auth context
	gotAuth := GetAuthContext(ctxWithAuth)
	require.NotNil(t, gotAuth)
	assert.Equal(t, "user-123", gotAuth.SubjectID)
	assert.Equal(t, RoleAdmin, gotAuth.Role)
}

func TestGetAuthContext_NilContext(t *testing.T) {
	// GetAuthContext should handle context without auth gracefully
	ctx := context.Background()
	authCtx := GetAuthContext(ctx)
	assert.Nil(t, authCtx)
}

func TestGetAuthContext_WrongType(t *testing.T) {
	// If someone stores a wrong type at the same key (shouldn't happen),
	// GetAuthContext should return nil
	ctx := context.WithValue(context.Background(), authContextKey{}, "not-an-auth-context")
	authCtx := GetAuthContext(ctx)
	assert.Nil(t, authCtx)
}

func TestAuthContext_HasRole(t *testing.T) {
	tests := []struct {
		name     string
		authCtx  *AuthContext
		role     string
		expected bool
	}{
		{
			name:     "nil auth context returns false",
			authCtx:  nil,
			role:     RoleAdmin,
			expected: false,
		},
		{
			name: "matching role returns true",
			authCtx: &AuthContext{
				SubjectID: "user-123",
				Role:      RoleAdmin,
			},
			role:     RoleAdmin,
			expected: true,
		},
		{
			name: "non-matching role returns false",
			authCtx: &AuthContext{
				SubjectID: "user-123",
				Role:      RoleUser,
			},
			role:     RoleAdmin,
			expected: false,
		},
		{
			name: "empty role returns false",
			authCtx: &AuthContext{
				SubjectID: "user-123",
				Role:      "",
			},
			role:     RoleAdmin,
			expected: false,
		},
		{
			name: "case sensitive role check",
			authCtx: &AuthContext{
				SubjectID: "user-123",
				Role:      "ADMIN",
			},
			role:     RoleAdmin,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.authCtx.HasRole(tt.role)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAuthContext_IsAdmin(t *testing.T) {
	tests := []struct {
		name     string
		authCtx  *AuthContext
		expected bool
	}{
		{
			name:     "nil auth context returns false",
			authCtx:  nil,
			expected: false,
		},
		{
			name: "admin role returns true",
			authCtx: &AuthContext{
				SubjectID: "user-123",
				Role:      RoleAdmin,
			},
			expected: true,
		},
		{
			name: "user role returns false",
			authCtx: &AuthContext{
				SubjectID: "user-123",
				Role:      RoleUser,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.authCtx.IsAdmin()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAuthContext_IsUser(t *testing.T) {
	tests := []struct {
		name     string
		authCtx  *AuthContext
		expected bool
	}{
		{
			name:     "nil auth context returns false",
			authCtx:  nil,
			expected: false,
		},
		{
			name: "user role returns true",
			authCtx: &AuthContext{
				SubjectID: "user-123",
				Role:      RoleUser,
			},
			expected: true,
		},
		{
			name: "admin role returns false",
			authCtx: &AuthContext{
				SubjectID: "user-123",
				Role:      RoleAdmin,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.authCtx.IsUser()
			assert.Equal(t, tt.expected, result)
		})
	}
}
