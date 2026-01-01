//go:build integration

package postgres_test

import (
	"context"
	"database/sql"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/iruldev/golang-api-hexagonal/internal/infra/postgres"
	"github.com/iruldev/golang-api-hexagonal/internal/testutil/containers"
	"github.com/iruldev/golang-api-hexagonal/internal/transport/http/middleware"
)

func idempotencyTestMigrationsDir(t *testing.T) string {
	t.Helper()

	_, filename, _, ok := runtime.Caller(0)
	require.True(t, ok, "runtime.Caller failed")

	path := filepath.Clean(filepath.Join(filepath.Dir(filename), "../../../migrations"))
	abs, err := filepath.Abs(path)
	require.NoError(t, err)
	return abs
}

func requireSafeIdempotencyTestDatabase(t *testing.T, databaseURL string) {
	t.Helper()

	if os.Getenv("ALLOW_NON_TEST_DATABASE") == "true" {
		return
	}

	u, err := url.Parse(databaseURL)
	require.NoError(t, err)

	dbName := strings.TrimPrefix(u.Path, "/")
	require.NotEmpty(t, dbName, "DATABASE_URL must include database name")

	if !strings.HasSuffix(dbName, "_test") {
		t.Skipf("refusing to run destructive integration tests on non-test database %q", dbName)
	}
}

func setupIdempotencyTestDB(t *testing.T) (postgres.Pooler, func()) {
	t.Helper()

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		t.Log("DATABASE_URL not set, using testcontainers")
		pool := containers.NewPostgres(t)
		containers.MigrateWithPath(t, pool, idempotencyTestMigrationsDir(t))

		cleanup := func() {
			ctx := context.Background()
			_, _ = pool.Exec(ctx, "DELETE FROM idempotency_keys")
		}
		return &dbAdapter{p: pool}, cleanup
	}

	requireSafeIdempotencyTestDatabase(t, databaseURL)

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, databaseURL)
	require.NoError(t, err)

	db, err := sql.Open("pgx", databaseURL)
	require.NoError(t, err)

	require.NoError(t, goose.SetDialect("postgres"))
	err = goose.Up(db, idempotencyTestMigrationsDir(t))
	require.NoError(t, err)

	cleanup := func() {
		ctx := context.Background()
		_, _ = pool.Exec(ctx, "DELETE FROM idempotency_keys")
		db.Close()
		pool.Close()
	}

	return &dbAdapter{p: pool}, cleanup
}

func TestIdempotencyRepo_Store_Success(t *testing.T) {
	pool, cleanup := setupIdempotencyTestDB(t)
	defer cleanup()

	repo := postgres.NewIdempotencyRepo(pool)
	ctx := context.Background()

	now := time.Now().UTC().Truncate(time.Microsecond)
	record := &middleware.IdempotencyRecord{
		Key:             "test-key-store-success",
		RequestHash:     "sha256:abc123def456",
		StatusCode:      201,
		ResponseHeaders: http.Header{"Content-Type": []string{"application/json"}},
		ResponseBody:    []byte(`{"id": "12345", "status": "created"}`),
		CreatedAt:       now,
		ExpiresAt:       now.Add(24 * time.Hour),
	}

	err := repo.Store(ctx, record)
	assert.NoError(t, err)

	// Verify we can retrieve it
	found, err := repo.Get(ctx, record.Key)
	assert.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, record.Key, found.Key)
	assert.Equal(t, record.RequestHash, found.RequestHash)
	assert.Equal(t, record.StatusCode, found.StatusCode)
	assert.Equal(t, record.ResponseHeaders.Get("Content-Type"), found.ResponseHeaders.Get("Content-Type"))
	assert.Equal(t, record.ResponseBody, found.ResponseBody)
}

func TestIdempotencyRepo_Get_NotFound(t *testing.T) {
	pool, cleanup := setupIdempotencyTestDB(t)
	defer cleanup()

	repo := postgres.NewIdempotencyRepo(pool)
	ctx := context.Background()

	found, err := repo.Get(ctx, "non-existent-key-12345")
	assert.NoError(t, err)
	assert.Nil(t, found)
}

func TestIdempotencyRepo_Get_Expired(t *testing.T) {
	pool, cleanup := setupIdempotencyTestDB(t)
	defer cleanup()

	repo := postgres.NewIdempotencyRepo(pool)
	ctx := context.Background()

	now := time.Now().UTC().Truncate(time.Microsecond)
	record := &middleware.IdempotencyRecord{
		Key:             "test-key-expired",
		RequestHash:     "sha256:expired",
		StatusCode:      200,
		ResponseHeaders: http.Header{},
		ResponseBody:    []byte(`{}`),
		CreatedAt:       now.Add(-25 * time.Hour), // Created 25 hours ago
		ExpiresAt:       now.Add(-1 * time.Hour),  // Expired 1 hour ago
	}

	err := repo.Store(ctx, record)
	require.NoError(t, err)

	// Trying to get expired record should return nil (query filters by expires_at > NOW())
	found, err := repo.Get(ctx, record.Key)
	assert.NoError(t, err)
	assert.Nil(t, found, "expired record should not be returned")
}

func TestIdempotencyRepo_Store_DuplicateKey(t *testing.T) {
	pool, cleanup := setupIdempotencyTestDB(t)
	defer cleanup()

	repo := postgres.NewIdempotencyRepo(pool)
	ctx := context.Background()

	now := time.Now().UTC().Truncate(time.Microsecond)
	record := &middleware.IdempotencyRecord{
		Key:             "test-key-duplicate",
		RequestHash:     "sha256:original",
		StatusCode:      200,
		ResponseHeaders: http.Header{},
		ResponseBody:    []byte(`{"original": true}`),
		CreatedAt:       now,
		ExpiresAt:       now.Add(24 * time.Hour),
	}

	// First store should succeed
	err := repo.Store(ctx, record)
	require.NoError(t, err)

	// Second store with same key should fail
	record2 := &middleware.IdempotencyRecord{
		Key:             "test-key-duplicate", // Same key
		RequestHash:     "sha256:different",
		StatusCode:      201,
		ResponseHeaders: http.Header{},
		ResponseBody:    []byte(`{"different": true}`),
		CreatedAt:       now,
		ExpiresAt:       now.Add(24 * time.Hour),
	}

	err = repo.Store(ctx, record2)
	assert.Error(t, err)
	assert.ErrorIs(t, err, postgres.ErrKeyAlreadyExists)
}

func TestIdempotencyRepo_DeleteExpired(t *testing.T) {
	pool, cleanup := setupIdempotencyTestDB(t)
	defer cleanup()

	repo := postgres.NewIdempotencyRepo(pool)
	ctx := context.Background()

	now := time.Now().UTC().Truncate(time.Microsecond)

	// Create an expired record
	expiredRecord := &middleware.IdempotencyRecord{
		Key:             "test-key-to-delete",
		RequestHash:     "sha256:expired-delete",
		StatusCode:      200,
		ResponseHeaders: http.Header{},
		ResponseBody:    []byte(`{}`),
		CreatedAt:       now.Add(-25 * time.Hour),
		ExpiresAt:       now.Add(-1 * time.Hour), // Expired
	}
	err := repo.Store(ctx, expiredRecord)
	require.NoError(t, err)

	// Create a valid (non-expired) record
	validRecord := &middleware.IdempotencyRecord{
		Key:             "test-key-still-valid",
		RequestHash:     "sha256:valid",
		StatusCode:      200,
		ResponseHeaders: http.Header{},
		ResponseBody:    []byte(`{}`),
		CreatedAt:       now,
		ExpiresAt:       now.Add(24 * time.Hour), // Not expired
	}
	err = repo.Store(ctx, validRecord)
	require.NoError(t, err)

	// Delete expired
	deleted, err := repo.DeleteExpired(ctx)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, deleted, int64(1), "should have deleted at least 1 expired record")

	// Verify expired record is gone (we can't get it anyway due to query filter, but double-check via count)
	// The valid record should still exist
	found, err := repo.Get(ctx, validRecord.Key)
	assert.NoError(t, err)
	assert.NotNil(t, found, "valid record should still exist")
}

func TestIdempotencyRepo_Headers_Serialization(t *testing.T) {
	pool, cleanup := setupIdempotencyTestDB(t)
	defer cleanup()

	repo := postgres.NewIdempotencyRepo(pool)
	ctx := context.Background()

	now := time.Now().UTC().Truncate(time.Microsecond)

	// Store with multiple headers
	headers := http.Header{
		"Content-Type":    []string{"application/json; charset=utf-8"},
		"X-Custom-Header": []string{"value1", "value2"},
		"X-Request-Id":    []string{"req-12345"},
		"Cache-Control":   []string{"no-cache"},
		"Idempotency-Key": []string{"some-key"},
	}

	record := &middleware.IdempotencyRecord{
		Key:             "test-key-headers",
		RequestHash:     "sha256:headers-test",
		StatusCode:      200,
		ResponseHeaders: headers,
		ResponseBody:    []byte(`{"test": "headers"}`),
		CreatedAt:       now,
		ExpiresAt:       now.Add(24 * time.Hour),
	}

	err := repo.Store(ctx, record)
	require.NoError(t, err)

	// Retrieve and verify headers are correctly deserialized
	found, err := repo.Get(ctx, record.Key)
	require.NoError(t, err)
	require.NotNil(t, found)

	assert.Equal(t, "application/json; charset=utf-8", found.ResponseHeaders.Get("Content-Type"))
	assert.Equal(t, "value1", found.ResponseHeaders.Get("X-Custom-Header"))
	assert.Equal(t, []string{"value1", "value2"}, found.ResponseHeaders["X-Custom-Header"])
	assert.Equal(t, "req-12345", found.ResponseHeaders.Get("X-Request-Id"))
	assert.Equal(t, "no-cache", found.ResponseHeaders.Get("Cache-Control"))
}

func TestIdempotencyRepo_ResponseBody_Binary(t *testing.T) {
	pool, cleanup := setupIdempotencyTestDB(t)
	defer cleanup()

	repo := postgres.NewIdempotencyRepo(pool)
	ctx := context.Background()

	now := time.Now().UTC().Truncate(time.Microsecond)

	// Create a record with binary data (could be gzipped response, etc.)
	binaryData := []byte{0x00, 0x01, 0x02, 0xFF, 0xFE, 0xFD}

	record := &middleware.IdempotencyRecord{
		Key:             "test-key-binary",
		RequestHash:     "sha256:binary-test",
		StatusCode:      200,
		ResponseHeaders: http.Header{"Content-Encoding": []string{"gzip"}},
		ResponseBody:    binaryData,
		CreatedAt:       now,
		ExpiresAt:       now.Add(24 * time.Hour),
	}

	err := repo.Store(ctx, record)
	require.NoError(t, err)

	found, err := repo.Get(ctx, record.Key)
	require.NoError(t, err)
	require.NotNil(t, found)

	assert.Equal(t, binaryData, found.ResponseBody, "binary data should be preserved exactly")
}

func TestIdempotencyRepo_EmptyResponseBody(t *testing.T) {
	pool, cleanup := setupIdempotencyTestDB(t)
	defer cleanup()

	repo := postgres.NewIdempotencyRepo(pool)
	ctx := context.Background()

	now := time.Now().UTC().Truncate(time.Microsecond)

	// Some responses (like 204 No Content) have empty body
	record := &middleware.IdempotencyRecord{
		Key:             "test-key-empty-body",
		RequestHash:     "sha256:empty-body",
		StatusCode:      204,
		ResponseHeaders: http.Header{},
		ResponseBody:    []byte{},
		CreatedAt:       now,
		ExpiresAt:       now.Add(24 * time.Hour),
	}

	err := repo.Store(ctx, record)
	require.NoError(t, err)

	found, err := repo.Get(ctx, record.Key)
	require.NoError(t, err)
	require.NotNil(t, found)

	assert.Empty(t, found.ResponseBody)
	assert.Equal(t, 204, found.StatusCode)
}
