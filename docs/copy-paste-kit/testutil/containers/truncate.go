// internal/testutil/containers/truncate.go
package containers

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Truncate truncates the specified tables with CASCADE.
// Use when transaction isolation isn't suitable.
//
// Example usage:
//
//	func TestMultipleUsers(t *testing.T) {
//	    pool := containers.NewPostgres(t)
//	    containers.Migrate(t, pool)
//
//	    t.Cleanup(func() {
//	        containers.Truncate(t, pool, "users", "audit_events")
//	    })
//
//	    // ... test that commits data
//	}
func Truncate(t testing.TB, pool *pgxpool.Pool, tables ...string) {
	t.Helper()
	ctx := context.Background()

	if len(tables) == 0 {
		return
	}

	// Use quoted identifiers to prevent SQL injection
	quotedTables := make([]string, len(tables))
	for i, table := range tables {
		quotedTables[i] = fmt.Sprintf("%q", table)
	}

	query := fmt.Sprintf("TRUNCATE TABLE %s CASCADE", strings.Join(quotedTables, ", "))
	if _, err := pool.Exec(ctx, query); err != nil {
		t.Fatalf("failed to truncate tables %v: %v", tables, err)
	}
}

// TruncateContext is like Truncate but uses the provided context.
func TruncateContext(t testing.TB, ctx context.Context, pool *pgxpool.Pool, tables ...string) {
	t.Helper()

	if len(tables) == 0 {
		return
	}

	quotedTables := make([]string, len(tables))
	for i, table := range tables {
		quotedTables[i] = fmt.Sprintf("%q", table)
	}

	query := fmt.Sprintf("TRUNCATE TABLE %s CASCADE", strings.Join(quotedTables, ", "))
	if _, err := pool.Exec(ctx, query); err != nil {
		t.Fatalf("failed to truncate tables %v: %v", tables, err)
	}
}
