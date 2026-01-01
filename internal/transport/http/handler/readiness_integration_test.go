//go:build integration

package handler_test

import (
	"context"
	"encoding/json"
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

	// 2. Setup the ReadinessHandler with the real DatabaseHealthChecker
	dbChecker := postgres.NewDatabaseHealthChecker(pool)
	readinessHandler := handler.NewReadinessHandler(2*time.Second, dbChecker)

	t.Run("Healthy Database", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
		rec := httptest.NewRecorder()

		readinessHandler.ServeHTTP(rec, req)

		// Assert 200 OK
		assert.Equal(t, http.StatusOK, rec.Code)

		// Assert Response Body
		var resp handler.ReadinessResponse
		err := json.NewDecoder(rec.Body).Decode(&resp)
		require.NoError(t, err)

		assert.Equal(t, "healthy", resp.Status)
		assert.Equal(t, "healthy", resp.Checks["database"].Status)
		assert.GreaterOrEqual(t, resp.Checks["database"].LatencyMs, int64(0))
		assert.Empty(t, resp.Checks["database"].Error)
	})

	t.Run("Unhealthy Database (Closed Pool)", func(t *testing.T) {
		// Create a separate pool for this test so we can close it without affecting others
		deadPool, err := pgxpool.New(ctx, dbURL)
		require.NoError(t, err)
		deadPool.Close() // Immediately close it to simulate failure

		deadChecker := postgres.NewDatabaseHealthChecker(deadPool)
		deadHandler := handler.NewReadinessHandler(2*time.Second, deadChecker)

		req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
		rec := httptest.NewRecorder()

		deadHandler.ServeHTTP(rec, req)

		// Assert 503 Service Unavailable
		assert.Equal(t, http.StatusServiceUnavailable, rec.Code)

		// Assert Response Body
		var resp handler.ReadinessResponse
		err = json.NewDecoder(rec.Body).Decode(&resp)
		require.NoError(t, err)

		assert.Equal(t, "unhealthy", resp.Status)
		assert.Equal(t, "unhealthy", resp.Checks["database"].Status)
		assert.NotEmpty(t, resp.Checks["database"].Error)
		assert.Contains(t, resp.Checks["database"].Error, "closed") // pgx returns "closed pool" or similar
	})
}
