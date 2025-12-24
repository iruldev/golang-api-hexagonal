//go:build integration

package main

import (
	"context"
	"database/sql"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMigrationIdempotency verifies that running goose migrations multiple times
// does not cause errors (idempotency requirement from Story 1.5, FR5).
//
// This test requires a running PostgreSQL database.
// Run with: go test -tags=integration ./cmd/api/... -run TestMigrationIdempotency
func TestMigrationIdempotency(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		t.Skip("DATABASE_URL not set, skipping migration idempotency test")
	}

	ctx := context.Background()

	// Connect to database with pgxpool for checking results
	pool, err := pgxpool.New(ctx, dbURL)
	require.NoError(t, err, "Failed to connect to database")
	defer pool.Close()

	// Connect to database with stdlib for goose
	db, err := sql.Open("pgx", dbURL)
	require.NoError(t, err, "Failed to open stdlib connection")
	defer db.Close()

	// Set goose dialect
	require.NoError(t, goose.SetDialect("postgres"))

	migrationsDir := "../../migrations"

	// First migration run
	err = goose.Up(db, migrationsDir)
	require.NoError(t, err, "First goose up should succeed")

	// Second migration run - this is the idempotency test
	err = goose.Up(db, migrationsDir)
	assert.NoError(t, err, "Second goose up should succeed (idempotency)")

	// Verify goose_db_version table has correct count
	// Dynamic check: count .sql files in migrations dir to ensure test doesn't break on new migrations
	files, err := os.ReadDir(migrationsDir)
	require.NoError(t, err)
	expectedCount := 0
	for _, f := range files {
		if !f.IsDir() && len(f.Name()) > 4 && f.Name()[len(f.Name())-4:] == ".sql" {
			expectedCount++
		}
	}

	var migrationCount int
	err = pool.QueryRow(ctx, "SELECT COUNT(*) FROM goose_db_version WHERE version_id > 0").Scan(&migrationCount)
	require.NoError(t, err)
	assert.Equal(t, expectedCount, migrationCount, "Database migration count should match number of migration files")

	// Verify schema_info table exists and has data
	var schemaVersion string
	err = pool.QueryRow(ctx, "SELECT version FROM schema_info LIMIT 1").Scan(&schemaVersion)
	require.NoError(t, err, "schema_info table should exist with data")
	assert.Equal(t, "0.0.1", schemaVersion)
}

// TestMigrationTablesExist verifies that both schema_info and goose_db_version
// tables are created after running migrations.
func TestMigrationTablesExist(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		t.Skip("DATABASE_URL not set, skipping migration tables test")
	}

	ctx := context.Background()

	pool, err := pgxpool.New(ctx, dbURL)
	require.NoError(t, err)
	defer pool.Close()

	// Check schema_info exists
	var schemaInfoExists bool
	err = pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_name = 'schema_info'
		)
	`).Scan(&schemaInfoExists)
	require.NoError(t, err)
	assert.True(t, schemaInfoExists, "schema_info table should exist")

	// Check goose_db_version exists
	var gooseTableExists bool
	err = pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_name = 'goose_db_version'
		)
	`).Scan(&gooseTableExists)
	require.NoError(t, err)
	assert.True(t, gooseTableExists, "goose_db_version table should exist")
}
