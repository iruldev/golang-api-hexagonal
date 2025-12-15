package ctxutil

import (
	"context"
	"errors"
	"testing"
)

func TestClaims_HasRole(t *testing.T) {
	tests := []struct {
		name     string
		claims   Claims
		role     string
		expected bool
	}{
		{
			name:     "has role",
			claims:   Claims{Roles: []string{"admin", "user"}},
			role:     "admin",
			expected: true,
		},
		{
			name:     "does not have role",
			claims:   Claims{Roles: []string{"user"}},
			role:     "admin",
			expected: false,
		},
		{
			name:     "empty roles",
			claims:   Claims{Roles: []string{}},
			role:     "admin",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.claims.HasRole(tt.role); got != tt.expected {
				t.Errorf("Claims.HasRole() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestClaims_HasPermission(t *testing.T) {
	tests := []struct {
		name       string
		claims     Claims
		permission string
		expected   bool
	}{
		{
			name:       "has permission",
			claims:     Claims{Permissions: []string{"read", "write"}},
			permission: "read",
			expected:   true,
		},
		{
			name:       "does not have permission",
			claims:     Claims{Permissions: []string{"read"}},
			permission: "delete",
			expected:   false,
		},
		{
			name:       "empty permissions",
			claims:     Claims{Permissions: []string{}},
			permission: "read",
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.claims.HasPermission(tt.permission); got != tt.expected {
				t.Errorf("Claims.HasPermission() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestContextOperations(t *testing.T) {
	t.Run("Claims storage and retrieval", func(t *testing.T) {
		claims := Claims{UserID: "user1"}
		ctx := NewClaimsContext(context.Background(), claims)

		got, err := ClaimsFromContext(ctx)
		if err != nil {
			t.Fatalf("ClaimsFromContext() error = %v", err)
		}
		if got.UserID != claims.UserID {
			t.Errorf("got UserID %q, want %q", got.UserID, claims.UserID)
		}
	})

	t.Run("Claims missing", func(t *testing.T) {
		_, err := ClaimsFromContext(context.Background())
		if !errors.Is(err, ErrNoClaimsInContext) {
			t.Errorf("expected ErrNoClaimsInContext, got %v", err)
		}
	})

	t.Run("RequestID storage and retrieval", func(t *testing.T) {
		reqID := "req-123"
		ctx := NewRequestIDContext(context.Background(), reqID)

		if got := RequestIDFromContext(ctx); got != reqID {
			t.Errorf("RequestIDFromContext() = %q, want %q", got, reqID)
		}
	})

	t.Run("RequestID missing", func(t *testing.T) {
		if got := RequestIDFromContext(context.Background()); got != "" {
			t.Errorf("RequestIDFromContext() = %q, want empty string", got)
		}
	})
}
