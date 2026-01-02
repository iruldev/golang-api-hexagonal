//go:build integration

package handler_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/iruldev/golang-api-hexagonal/internal/infra/postgres"
	"github.com/iruldev/golang-api-hexagonal/internal/transport/http/handler"
)

// TestReadinessHandler_Integration_RealDB verifies the readiness probe against a real database.
// This fulfills the requirement to test with a real database connection.
func TestReadinessHandler_Integration_RealDB(t *testing.T) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		t.Skip("DATABASE_URL not set, skipping integration test")
	}

	// Safety check to prevent running on non-test databases
	if !strings.Contains(dbURL, "_test") && os.Getenv("ALLOW_NON_TEST_DATABASE") != "true" {
		t.Skip("Skipping destructive test on non-test database (suffix _test required)")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 1. Connect to the real database
	pool, err := pgxpool.New(ctx, dbURL)
	require.NoError(t, err, "Failed to connect to database")
	defer pool.Close()

	// Verify connection is actually alive first
	err = pool.Ping(ctx)
	require.NoError(t, err, "Database connection is not healthy")

	// 2. Setup HealthCheckRegistry
	registry := handler.NewHealthCheckRegistrySimple()
	registry.AddReadinessCheck("database", postgres.NewDatabaseCheck(pool, 2*time.Second))

	t.Run("Healthy Database", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
		rec := httptest.NewRecorder()

		registry.ReadyHandler()(rec, req)

		// Assert 200 OK
		assert.Equal(t, http.StatusOK, rec.Code)

		// Library returns {} for success
		assert.JSONEq(t, "{}", rec.Body.String())
	})

	t.Run("Unhealthy Database (Closed Pool)", func(t *testing.T) {
		// Create a separate pool for this test so we can close it without affecting others
		deadPool, err := pgxpool.New(ctx, dbURL)
		require.NoError(t, err)
		deadPool.Close() // Immediately close it to simulate failure

		registry := handler.NewHealthCheckRegistrySimple()
		registry.AddReadinessCheck("database", postgres.NewDatabaseCheck(deadPool, 2*time.Second))

		req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
		rec := httptest.NewRecorder()

		registry.ReadyHandler()(rec, req)

		// Assert 503 Service Unavailable
		assert.Equal(t, http.StatusServiceUnavailable, rec.Code)

		// Assert Response Body is empty (library behavior)
		assert.JSONEq(t, "{}", rec.Body.String())
	})
}
