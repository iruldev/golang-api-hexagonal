//go:build integration
// +build integration

package postgres_test

import (
	"context"
	"errors"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/iruldev/golang-api-hexagonal/internal/domain"
	"github.com/iruldev/golang-api-hexagonal/internal/infra/postgres"
)

// TestCaseInsensitiveEmail_CITEXT verifies that the email uniqueness constraint
// is case-insensitive after the CITEXT migration (Story 5.2).
//
// This test requires:
// 1. A running PostgreSQL instance with DATABASE_URL environment variable set
// 2. The migration 20251226084756_add_citext_email.sql to be applied
//
// Run with: go test -tags=integration -v ./internal/infra/postgres/... -run TestCaseInsensitiveEmail_CITEXT
func TestCaseInsensitiveEmail_CITEXT(t *testing.T) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		t.Skip("DATABASE_URL not set, skipping integration test")
	}

	// Safety check to prevent running on non-test databases
	u, err := url.Parse(dbURL)
	require.NoError(t, err)
	if !strings.HasSuffix(strings.TrimPrefix(u.Path, "/"), "_test") && os.Getenv("ALLOW_NON_TEST_DATABASE") != "true" {
		t.Skip("Skipping destructive test on non-test database (suffix _test required)")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Connect directly with pgx for raw queries
	conn, err := pgx.Connect(ctx, dbURL)
	require.NoError(t, err, "Failed to connect to database")
	defer conn.Close(ctx)

	// Setup: Generate unique test emails to avoid conflicts with other tests
	testID := uuid.New().String()[:8]
	email1 := "TestUser_" + testID + "@Example.com" // Mixed case
	email2 := "testuser_" + testID + "@example.com" // Lowercase version

	// Cleanup: Remove test users after test
	defer func() {
		_, _ = conn.Exec(ctx, "DELETE FROM users WHERE email ILIKE $1", "%"+testID+"%")
	}()

	// Test 1: Insert first user with mixed-case email
	userID1 := uuid.New()
	now := time.Now().UTC()
	_, err = conn.Exec(ctx,
		`INSERT INTO users (id, email, first_name, last_name, created_at, updated_at) 
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		userID1, email1, "Test", "User", now, now,
	)
	require.NoError(t, err, "Should be able to insert first user")

	// Test 2: Attempt to insert duplicate with different case should fail
	const errCodeUniqueViolation = "23505"
	userID2 := uuid.New()
	_, err = conn.Exec(ctx,
		`INSERT INTO users (id, email, first_name, last_name, created_at, updated_at) 
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		userID2, email2, "Duplicate", "User", now, now,
	)

	require.Error(t, err, "Should fail to insert duplicate case-insensitive email")

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		assert.Equal(t, errCodeUniqueViolation, pgErr.Code, "Expected unique violation error (23505)")
	} else {
		assert.Fail(t, "Expected pgconn.PgError, got different error type: "+err.Error())
	}

	// Test 3: Verify case preservation in stored email
	var storedEmail string
	err = conn.QueryRow(ctx, "SELECT email FROM users WHERE id = $1", userID1).Scan(&storedEmail)
	require.NoError(t, err)
	assert.Equal(t, email1, storedEmail, "Email should preserve original case (CITEXT feature)")
}

// TestCITEXTExtensionExists verifies that the citext extension is installed.
func TestCITEXTExtensionExists(t *testing.T) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		t.Skip("DATABASE_URL not set, skipping integration test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, err := pgx.Connect(ctx, dbURL)
	require.NoError(t, err, "Failed to connect to database")
	defer conn.Close(ctx)

	// Check if citext extension exists
	var extName string
	err = conn.QueryRow(ctx,
		"SELECT extname FROM pg_extension WHERE extname = 'citext'",
	).Scan(&extName)

	require.NoError(t, err, "citext extension should exist")
	assert.Equal(t, "citext", extName)
}

// Ensure postgres package is imported for NewPool
var _ = postgres.PoolConfig{}
var _ = domain.User{}
