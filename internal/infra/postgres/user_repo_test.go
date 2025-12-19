//go:build integration

package postgres_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/iruldev/golang-api-hexagonal/internal/domain"
	"github.com/iruldev/golang-api-hexagonal/internal/infra/postgres"
)

func migrationsDir(t *testing.T) string {
	t.Helper()

	_, filename, _, ok := runtime.Caller(0)
	require.True(t, ok, "runtime.Caller failed")

	// This file is: internal/infra/postgres/user_repo_test.go
	// Migrations live in: migrations/
	path := filepath.Clean(filepath.Join(filepath.Dir(filename), "../../../migrations"))
	abs, err := filepath.Abs(path)
	require.NoError(t, err)
	return abs
}

func requireSafeTestDatabase(t *testing.T, databaseURL string) {
	t.Helper()

	// Hard safety guard: integration tests run destructive cleanup (DELETE FROM users).
	// Refuse to run unless the DB name ends with "_test", unless explicitly overridden.
	if os.Getenv("ALLOW_NON_TEST_DATABASE") == "true" {
		return
	}

	u, err := url.Parse(databaseURL)
	require.NoError(t, err)

	dbName := strings.TrimPrefix(u.Path, "/")
	require.NotEmpty(t, dbName, "DATABASE_URL must include database name")

	if !strings.HasSuffix(dbName, "_test") {
		t.Skip(fmt.Sprintf("refusing to run destructive integration tests on non-test database %q (set DB name suffix _test or ALLOW_NON_TEST_DATABASE=true)", dbName))
	}
}

func setupTestDB(t *testing.T) (*pgxpool.Pool, func()) {
	t.Helper()
	ctx := context.Background()

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		t.Skip("DATABASE_URL not set (run `make infra-up` and set DATABASE_URL to a dedicated test DB)")
	}
	requireSafeTestDatabase(t, databaseURL)

	pool, err := pgxpool.New(ctx, databaseURL)
	require.NoError(t, err)

	// Run migrations using goose
	// Open a standard database/sql connection for goose
	db, err := sql.Open("pgx", databaseURL)
	require.NoError(t, err)

	require.NoError(t, goose.SetDialect("postgres"))
	err = goose.Up(db, migrationsDir(t))
	require.NoError(t, err)

	cleanup := func() {
		// Clean up test data - order matters for foreign keys.
		// Note: We explicitly delete from audit_events here even though this is user_repo_test.
		// This coupling is necessary because audit_events may reference users (though not via FK, logic might vary)
		// and we want a clean state for integration tests.
		_, _ = pool.Exec(ctx, "DELETE FROM audit_events")
		_, _ = pool.Exec(ctx, "DELETE FROM users")
		db.Close()
		pool.Close()
	}

	return pool, cleanup
}

func TestUserRepo_Create(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	repo := postgres.NewUserRepo()
	querier := postgres.NewPoolQuerier(pool)

	// Generate a valid UUID v7
	id, err := uuid.NewV7()
	require.NoError(t, err)

	now := time.Now().UTC().Truncate(time.Microsecond)
	user := &domain.User{
		ID:        domain.ID(id.String()),
		Email:     "test@example.com",
		FirstName: "John",
		LastName:  "Doe",
		CreatedAt: now,
		UpdatedAt: now,
	}

	err = repo.Create(ctx, querier, user)
	assert.NoError(t, err)

	// Verify stored
	found, err := repo.GetByID(ctx, querier, user.ID)
	assert.NoError(t, err)
	assert.Equal(t, user.Email, found.Email)
	assert.Equal(t, user.FirstName, found.FirstName)
	assert.Equal(t, user.LastName, found.LastName)
}

func TestUserRepo_Create_DuplicateEmail(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	repo := postgres.NewUserRepo()
	querier := postgres.NewPoolQuerier(pool)

	// Create first user
	id1, err := uuid.NewV7()
	require.NoError(t, err)

	now := time.Now().UTC().Truncate(time.Microsecond)
	user1 := &domain.User{
		ID:        domain.ID(id1.String()),
		Email:     "duplicate@example.com",
		FirstName: "John",
		LastName:  "Doe",
		CreatedAt: now,
		UpdatedAt: now,
	}

	err = repo.Create(ctx, querier, user1)
	require.NoError(t, err)

	// Try to create second user with same email
	id2, err := uuid.NewV7()
	require.NoError(t, err)

	user2 := &domain.User{
		ID:        domain.ID(id2.String()),
		Email:     "duplicate@example.com", // Same email
		FirstName: "Jane",
		LastName:  "Smith",
		CreatedAt: now,
		UpdatedAt: now,
	}

	err = repo.Create(ctx, querier, user2)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrEmailAlreadyExists), "expected ErrEmailAlreadyExists, got: %v", err)
}

func TestUserRepo_GetByID_Success(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	repo := postgres.NewUserRepo()
	querier := postgres.NewPoolQuerier(pool)

	// Create a user first
	id, err := uuid.NewV7()
	require.NoError(t, err)

	now := time.Now().UTC().Truncate(time.Microsecond)
	user := &domain.User{
		ID:        domain.ID(id.String()),
		Email:     "getbyid@example.com",
		FirstName: "Alice",
		LastName:  "Wonder",
		CreatedAt: now,
		UpdatedAt: now,
	}

	err = repo.Create(ctx, querier, user)
	require.NoError(t, err)

	// Get by ID
	found, err := repo.GetByID(ctx, querier, user.ID)
	assert.NoError(t, err)
	assert.NotNil(t, found)
	assert.Equal(t, user.ID, found.ID)
	assert.Equal(t, user.Email, found.Email)
	assert.Equal(t, user.FirstName, found.FirstName)
	assert.Equal(t, user.LastName, found.LastName)
}

func TestUserRepo_GetByID_NotFound(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	repo := postgres.NewUserRepo()
	querier := postgres.NewPoolQuerier(pool)

	// Generate a random UUID that doesn't exist
	nonExistentID, err := uuid.NewV7()
	require.NoError(t, err)

	found, err := repo.GetByID(ctx, querier, domain.ID(nonExistentID.String()))
	assert.Nil(t, found)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrUserNotFound), "expected ErrUserNotFound, got: %v", err)
}

func TestUserRepo_List_WithPagination(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	repo := postgres.NewUserRepo()
	querier := postgres.NewPoolQuerier(pool)

	// Create multiple users
	now := time.Now().UTC().Truncate(time.Microsecond)
	for i := 0; i < 25; i++ {
		id, err := uuid.NewV7()
		require.NoError(t, err)

		user := &domain.User{
			ID:        domain.ID(id.String()),
			Email:     "user" + string(rune('a'+i)) + "@example.com",
			FirstName: "User",
			LastName:  string(rune('A' + i)),
			CreatedAt: now.Add(time.Duration(i) * time.Second),
			UpdatedAt: now.Add(time.Duration(i) * time.Second),
		}
		err = repo.Create(ctx, querier, user)
		require.NoError(t, err)
	}

	// Test first page
	params := domain.ListParams{Page: 1, PageSize: 10}
	users, totalCount, err := repo.List(ctx, querier, params)
	assert.NoError(t, err)
	assert.Equal(t, 25, totalCount)
	assert.Len(t, users, 10)

	// Test second page
	params = domain.ListParams{Page: 2, PageSize: 10}
	users, totalCount, err = repo.List(ctx, querier, params)
	assert.NoError(t, err)
	assert.Equal(t, 25, totalCount)
	assert.Len(t, users, 10)

	// Test third page (partial)
	params = domain.ListParams{Page: 3, PageSize: 10}
	users, totalCount, err = repo.List(ctx, querier, params)
	assert.NoError(t, err)
	assert.Equal(t, 25, totalCount)
	assert.Len(t, users, 5)
}

func TestUserRepo_List_OrderByCreatedAtDesc(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	repo := postgres.NewUserRepo()
	querier := postgres.NewPoolQuerier(pool)

	now := time.Now().UTC().Truncate(time.Microsecond)

	// Create users with different created_at times
	user1ID, _ := uuid.NewV7()
	user1 := &domain.User{
		ID:        domain.ID(user1ID.String()),
		Email:     "first@example.com",
		FirstName: "First",
		LastName:  "User",
		CreatedAt: now,
		UpdatedAt: now,
	}
	err := repo.Create(ctx, querier, user1)
	require.NoError(t, err)

	user2ID, _ := uuid.NewV7()
	user2 := &domain.User{
		ID:        domain.ID(user2ID.String()),
		Email:     "second@example.com",
		FirstName: "Second",
		LastName:  "User",
		CreatedAt: now.Add(10 * time.Second), // Created later
		UpdatedAt: now.Add(10 * time.Second),
	}
	err = repo.Create(ctx, querier, user2)
	require.NoError(t, err)

	// List should return user2 first (newer created_at)
	params := domain.ListParams{Page: 1, PageSize: 10}
	users, _, err := repo.List(ctx, querier, params)
	assert.NoError(t, err)
	assert.Len(t, users, 2)
	assert.Equal(t, "second@example.com", users[0].Email) // Newest first
	assert.Equal(t, "first@example.com", users[1].Email)
}

func TestUserRepo_List_OrderByIDDescWhenCreatedAtEqual(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	repo := postgres.NewUserRepo()
	querier := postgres.NewPoolQuerier(pool)

	now := time.Now().UTC().Truncate(time.Microsecond)

	// Create two users with the same created_at to verify the id DESC tie-breaker.
	id1 := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	id2 := uuid.MustParse("00000000-0000-0000-0000-000000000002")

	user1 := &domain.User{
		ID:        domain.ID(id1.String()),
		Email:     "tie1@example.com",
		FirstName: "Tie",
		LastName:  "One",
		CreatedAt: now,
		UpdatedAt: now,
	}
	require.NoError(t, repo.Create(ctx, querier, user1))

	user2 := &domain.User{
		ID:        domain.ID(id2.String()),
		Email:     "tie2@example.com",
		FirstName: "Tie",
		LastName:  "Two",
		CreatedAt: now,
		UpdatedAt: now,
	}
	require.NoError(t, repo.Create(ctx, querier, user2))

	params := domain.ListParams{Page: 1, PageSize: 10}
	users, _, err := repo.List(ctx, querier, params)
	require.NoError(t, err)
	require.Len(t, users, 2)

	// Same created_at, so id DESC should put id2 before id1.
	assert.Equal(t, domain.ID(id2.String()), users[0].ID)
	assert.Equal(t, domain.ID(id1.String()), users[1].ID)
}

func TestTxManager_WithTx_Rollback(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	repo := postgres.NewUserRepo()
	querier := postgres.NewPoolQuerier(pool)
	txManager := postgres.NewTxManager(pool)

	// Create a user in a transaction that will fail
	id, err := uuid.NewV7()
	require.NoError(t, err)

	now := time.Now().UTC().Truncate(time.Microsecond)
	user := &domain.User{
		ID:        domain.ID(id.String()),
		Email:     "rollback@example.com",
		FirstName: "Rollback",
		LastName:  "Test",
		CreatedAt: now,
		UpdatedAt: now,
	}

	expectedErr := errors.New("intentional failure")
	err = txManager.WithTx(ctx, func(tx domain.Querier) error {
		if err := repo.Create(ctx, tx, user); err != nil {
			return err
		}
		// Return error to trigger rollback
		return expectedErr
	})

	assert.ErrorIs(t, err, expectedErr)

	// Verify user was NOT created (rollback worked)
	found, err := repo.GetByID(ctx, querier, user.ID)
	assert.Nil(t, found)
	assert.True(t, errors.Is(err, domain.ErrUserNotFound))
}

func TestTxManager_WithTx_Commit(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	repo := postgres.NewUserRepo()
	querier := postgres.NewPoolQuerier(pool)
	txManager := postgres.NewTxManager(pool)

	id, err := uuid.NewV7()
	require.NoError(t, err)

	now := time.Now().UTC().Truncate(time.Microsecond)
	user := &domain.User{
		ID:        domain.ID(id.String()),
		Email:     "commit@example.com",
		FirstName: "Commit",
		LastName:  "Test",
		CreatedAt: now,
		UpdatedAt: now,
	}

	err = txManager.WithTx(ctx, func(tx domain.Querier) error {
		return repo.Create(ctx, tx, user)
	})
	assert.NoError(t, err)

	// Verify user WAS created (commit worked)
	found, err := repo.GetByID(ctx, querier, user.ID)
	assert.NoError(t, err)
	assert.NotNil(t, found)
	assert.Equal(t, user.Email, found.Email)
}
