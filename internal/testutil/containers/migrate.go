package containers

import (
	"database/sql"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

// Migrate applies goose migrations to the database.
// Uses migrations from the project's migrations/ directory.
//
// Example usage:
//
//	func TestUserRepo(t *testing.T) {
//	    pool := containers.NewPostgres(t)
//	    containers.Migrate(t, pool)
//	    // Tables now exist, ready for testing
//	}
func Migrate(t testing.TB, pool *pgxpool.Pool) {
	t.Helper()

	// Use stdlib to get a *sql.DB from pgxpool
	db := stdlib.OpenDBFromPool(pool)
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("failed to close sql.DB: %v", err)
		}
	}()

	// Set goose dialect
	if err := goose.SetDialect("postgres"); err != nil {
		t.Fatalf("failed to set goose dialect: %v", err)
	}

	// Run migrations
	if err := goose.Up(db, "migrations"); err != nil {
		t.Fatalf("goose up failed: %v", err)
	}
}

// MigrateWithPath applies goose migrations from a custom path.
// Useful when running tests from different working directories.
func MigrateWithPath(t testing.TB, pool *pgxpool.Pool, migrationsPath string) {
	t.Helper()

	db := stdlib.OpenDBFromPool(pool)
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("failed to close sql.DB: %v", err)
		}
	}()

	if err := goose.SetDialect("postgres"); err != nil {
		t.Fatalf("failed to set goose dialect: %v", err)
	}

	if err := goose.Up(db, migrationsPath); err != nil {
		t.Fatalf("goose up failed: %v", err)
	}
}

// Ensure sql import is used.
var _ sql.DB
