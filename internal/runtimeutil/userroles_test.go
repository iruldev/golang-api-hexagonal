package runtimeutil

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewInMemoryUserRoleStore(t *testing.T) {
	t.Parallel()

	t.Run("creates empty store", func(t *testing.T) {
		// Arrange & Act
		store := NewInMemoryUserRoleStore()

		// Assert
		require.NotNil(t, store)
		require.NotNil(t, store.users)
		assert.Empty(t, store.users)
	})

	t.Run("creates store with initial roles", func(t *testing.T) {
		// Arrange
		initialRoles := map[string][]string{
			"user-1": {"admin", "user"},
			"user-2": {"user"},
		}

		// Act
		store := NewInMemoryUserRoleStore(WithInitialUserRoles(initialRoles))

		// Assert
		require.NotNil(t, store)
		assert.Len(t, store.users, 2)
		assert.Equal(t, []string{"admin", "user"}, store.users["user-1"].Roles)
		assert.Equal(t, []string{"user"}, store.users["user-2"].Roles)
	})
}

func TestInMemoryUserRoleStore_GetUserRoles(t *testing.T) {
	t.Parallel()

	t.Run("returns roles for existing user", func(t *testing.T) {
		// Arrange
		store := NewInMemoryUserRoleStore(WithInitialUserRoles(map[string][]string{
			"user-1": {"admin", "user"},
		}))
		ctx := context.Background()

		// Act
		roles, err := store.GetUserRoles(ctx, "user-1")

		// Assert
		require.NoError(t, err)
		require.NotNil(t, roles)
		assert.Equal(t, "user-1", roles.UserID)
		assert.Equal(t, []string{"admin", "user"}, roles.Roles)
	})

	t.Run("returns empty roles for non-existent user", func(t *testing.T) {
		// Arrange
		store := NewInMemoryUserRoleStore()
		ctx := context.Background()

		// Act
		roles, err := store.GetUserRoles(ctx, "unknown-user")

		// Assert
		require.NoError(t, err)
		require.NotNil(t, roles)
		assert.Equal(t, "unknown-user", roles.UserID)
		assert.Empty(t, roles.Roles)
	})

	t.Run("returns error on cancelled context", func(t *testing.T) {
		// Arrange
		store := NewInMemoryUserRoleStore()
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		// Act
		roles, err := store.GetUserRoles(ctx, "user-1")

		// Assert
		require.Error(t, err)
		assert.Nil(t, roles)
		assert.ErrorIs(t, err, context.Canceled)
	})

	t.Run("returns copy of roles to prevent mutation", func(t *testing.T) {
		// Arrange
		store := NewInMemoryUserRoleStore(WithInitialUserRoles(map[string][]string{
			"user-1": {"admin", "user"},
		}))
		ctx := context.Background()

		// Act
		roles, err := store.GetUserRoles(ctx, "user-1")
		require.NoError(t, err)

		// Mutate the returned slice
		roles.Roles[0] = "modified"

		// Get again and verify original is unchanged
		roles2, err := store.GetUserRoles(ctx, "user-1")
		require.NoError(t, err)

		// Assert
		assert.Equal(t, "admin", roles2.Roles[0])
	})
}

func TestInMemoryUserRoleStore_SetUserRoles(t *testing.T) {
	t.Parallel()

	t.Run("sets roles for new user", func(t *testing.T) {
		// Arrange
		store := NewInMemoryUserRoleStore()
		ctx := context.Background()

		// Act
		roles, err := store.SetUserRoles(ctx, "user-1", []string{"admin", "user"})

		// Assert
		require.NoError(t, err)
		require.NotNil(t, roles)
		assert.Equal(t, "user-1", roles.UserID)
		assert.Equal(t, []string{"admin", "user"}, roles.Roles)
		assert.False(t, roles.UpdatedAt.IsZero())
	})

	t.Run("replaces roles for existing user", func(t *testing.T) {
		// Arrange
		store := NewInMemoryUserRoleStore(WithInitialUserRoles(map[string][]string{
			"user-1": {"admin", "user"},
		}))
		ctx := context.Background()

		// Act
		roles, err := store.SetUserRoles(ctx, "user-1", []string{"service"})

		// Assert
		require.NoError(t, err)
		require.NotNil(t, roles)
		assert.Equal(t, []string{"service"}, roles.Roles)
	})

	t.Run("deduplicates roles", func(t *testing.T) {
		// Arrange
		store := NewInMemoryUserRoleStore()
		ctx := context.Background()

		// Act
		roles, err := store.SetUserRoles(ctx, "user-1", []string{"admin", "user", "admin", "user"})

		// Assert
		require.NoError(t, err)
		require.NotNil(t, roles)
		assert.Len(t, roles.Roles, 2)
		assert.Contains(t, roles.Roles, "admin")
		assert.Contains(t, roles.Roles, "user")
	})

	t.Run("returns error for invalid role", func(t *testing.T) {
		// Arrange
		store := NewInMemoryUserRoleStore()
		ctx := context.Background()

		// Act
		roles, err := store.SetUserRoles(ctx, "user-1", []string{"admin", "invalid_role"})

		// Assert
		require.Error(t, err)
		assert.Nil(t, roles)
		assert.ErrorIs(t, err, ErrInvalidRole)
	})

	t.Run("returns error on cancelled context", func(t *testing.T) {
		// Arrange
		store := NewInMemoryUserRoleStore()
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		// Act
		roles, err := store.SetUserRoles(ctx, "user-1", []string{"admin"})

		// Assert
		require.Error(t, err)
		assert.Nil(t, roles)
		assert.ErrorIs(t, err, context.Canceled)
	})

	t.Run("allows empty roles list", func(t *testing.T) {
		// Arrange
		store := NewInMemoryUserRoleStore(WithInitialUserRoles(map[string][]string{
			"user-1": {"admin", "user"},
		}))
		ctx := context.Background()

		// Act
		roles, err := store.SetUserRoles(ctx, "user-1", []string{})

		// Assert
		require.NoError(t, err)
		require.NotNil(t, roles)
		assert.Empty(t, roles.Roles)
	})
}

func TestInMemoryUserRoleStore_AddUserRole(t *testing.T) {
	t.Parallel()

	t.Run("adds role to new user", func(t *testing.T) {
		// Arrange
		store := NewInMemoryUserRoleStore()
		ctx := context.Background()

		// Act
		roles, err := store.AddUserRole(ctx, "user-1", "admin")

		// Assert
		require.NoError(t, err)
		require.NotNil(t, roles)
		assert.Equal(t, "user-1", roles.UserID)
		assert.Equal(t, []string{"admin"}, roles.Roles)
	})

	t.Run("adds role to existing user", func(t *testing.T) {
		// Arrange
		store := NewInMemoryUserRoleStore(WithInitialUserRoles(map[string][]string{
			"user-1": {"user"},
		}))
		ctx := context.Background()

		// Act
		roles, err := store.AddUserRole(ctx, "user-1", "admin")

		// Assert
		require.NoError(t, err)
		require.NotNil(t, roles)
		assert.Contains(t, roles.Roles, "user")
		assert.Contains(t, roles.Roles, "admin")
	})

	t.Run("is idempotent for existing role", func(t *testing.T) {
		// Arrange
		store := NewInMemoryUserRoleStore(WithInitialUserRoles(map[string][]string{
			"user-1": {"admin", "user"},
		}))
		ctx := context.Background()

		// Act
		roles, err := store.AddUserRole(ctx, "user-1", "admin")

		// Assert
		require.NoError(t, err)
		require.NotNil(t, roles)
		assert.Len(t, roles.Roles, 2)
		assert.Equal(t, []string{"admin", "user"}, roles.Roles)
	})

	t.Run("returns error for invalid role", func(t *testing.T) {
		// Arrange
		store := NewInMemoryUserRoleStore()
		ctx := context.Background()

		// Act
		roles, err := store.AddUserRole(ctx, "user-1", "invalid_role")

		// Assert
		require.Error(t, err)
		assert.Nil(t, roles)
		assert.ErrorIs(t, err, ErrInvalidRole)
	})

	t.Run("returns error on cancelled context", func(t *testing.T) {
		// Arrange
		store := NewInMemoryUserRoleStore()
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		// Act
		roles, err := store.AddUserRole(ctx, "user-1", "admin")

		// Assert
		require.Error(t, err)
		assert.Nil(t, roles)
		assert.ErrorIs(t, err, context.Canceled)
	})
}

func TestInMemoryUserRoleStore_RemoveUserRole(t *testing.T) {
	t.Parallel()

	t.Run("removes role from existing user", func(t *testing.T) {
		// Arrange
		store := NewInMemoryUserRoleStore(WithInitialUserRoles(map[string][]string{
			"user-1": {"admin", "user"},
		}))
		ctx := context.Background()

		// Act
		roles, err := store.RemoveUserRole(ctx, "user-1", "admin")

		// Assert
		require.NoError(t, err)
		require.NotNil(t, roles)
		assert.Equal(t, []string{"user"}, roles.Roles)
	})

	t.Run("is idempotent for non-existent role", func(t *testing.T) {
		// Arrange
		store := NewInMemoryUserRoleStore(WithInitialUserRoles(map[string][]string{
			"user-1": {"user"},
		}))
		ctx := context.Background()

		// Act
		roles, err := store.RemoveUserRole(ctx, "user-1", "admin")

		// Assert
		require.NoError(t, err)
		require.NotNil(t, roles)
		assert.Equal(t, []string{"user"}, roles.Roles)
	})

	t.Run("is idempotent for non-existent user", func(t *testing.T) {
		// Arrange
		store := NewInMemoryUserRoleStore()
		ctx := context.Background()

		// Act
		roles, err := store.RemoveUserRole(ctx, "unknown-user", "admin")

		// Assert
		require.NoError(t, err)
		require.NotNil(t, roles)
		assert.Empty(t, roles.Roles)
	})

	t.Run("returns error on cancelled context", func(t *testing.T) {
		// Arrange
		store := NewInMemoryUserRoleStore()
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		// Act
		roles, err := store.RemoveUserRole(ctx, "user-1", "admin")

		// Assert
		require.Error(t, err)
		assert.Nil(t, roles)
		assert.ErrorIs(t, err, context.Canceled)
	})

	t.Run("returns error for invalid role", func(t *testing.T) {
		// Arrange
		store := NewInMemoryUserRoleStore(WithInitialUserRoles(map[string][]string{
			"user-1": {"admin", "user"},
		}))
		ctx := context.Background()

		// Act
		roles, err := store.RemoveUserRole(ctx, "user-1", "invalid_role")

		// Assert
		require.Error(t, err)
		assert.Nil(t, roles)
		assert.ErrorIs(t, err, ErrInvalidRole)
	})
}

func TestInMemoryUserRoleStore_ThreadSafety(t *testing.T) {
	t.Parallel()

	// Arrange
	store := NewInMemoryUserRoleStore()
	ctx := context.Background()
	userID := "concurrent-user"

	const goroutines = 10
	const operations = 100

	var wg sync.WaitGroup
	wg.Add(goroutines * 3) // 3 types of operations

	// Act - concurrent SetUserRoles
	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < operations; j++ {
				roles := []string{"user"}
				if j%2 == 0 {
					roles = append(roles, "admin")
				}
				_, err := store.SetUserRoles(ctx, userID, roles)
				assert.NoError(t, err)
			}
		}(i)
	}

	// Act - concurrent AddUserRole
	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < operations; j++ {
				_, err := store.AddUserRole(ctx, userID, "service")
				assert.NoError(t, err)
			}
		}(i)
	}

	// Act - concurrent GetUserRoles
	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < operations; j++ {
				_, err := store.GetUserRoles(ctx, userID)
				assert.NoError(t, err)
			}
		}(i)
	}

	wg.Wait()

	// Assert - no panics occurred and data is accessible
	roles, err := store.GetUserRoles(ctx, userID)
	require.NoError(t, err)
	require.NotNil(t, roles)
}

func TestIsValidRole(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		role     string
		expected bool
	}{
		{"valid admin", "admin", true},
		{"valid service", "service", true},
		{"valid user", "user", true},
		{"invalid role", "invalid_role", false},
		{"empty role", "", false},
		{"uppercase admin", "ADMIN", false},
		{"superuser", "superuser", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Act
			result := IsValidRole(tt.role)

			// Assert
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestWithInitialUserRoles_SetsTimestamp(t *testing.T) {
	// NOTE: DO NOT use t.Parallel() here - this test modifies the package-level
	// variable userRolesTimeNow, which would race with other parallel tests.

	// Arrange
	fixedTime := time.Date(2025, 12, 14, 23, 0, 0, 0, time.UTC)
	originalTimeNow := userRolesTimeNow
	userRolesTimeNow = func() time.Time { return fixedTime }
	defer func() { userRolesTimeNow = originalTimeNow }()

	// Act
	store := NewInMemoryUserRoleStore(WithInitialUserRoles(map[string][]string{
		"user-1": {"admin"},
	}))

	// Assert
	require.Contains(t, store.users, "user-1")
	assert.Equal(t, fixedTime, store.users["user-1"].UpdatedAt)
}
