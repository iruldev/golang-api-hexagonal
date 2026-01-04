package postgres

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockPooler struct {
	pool *pgxpool.Pool
}

func (m *mockPooler) Ping(ctx context.Context) error { return nil }
func (m *mockPooler) Close()                         {}
func (m *mockPooler) Pool() *pgxpool.Pool            { return m.pool }

// TestDBMetrics_Collect verifies that metrics are collected correctly.
// Note: Since we can't easily mock pgxpool.Stat() without a real pool,
// we mostly verify that the collector registers without error and handles nil pools safely.
func TestDBMetrics_Collect(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	t.Run("HandleNilPool", func(t *testing.T) {
		// Create collector with nil pool
		m := NewDBMetrics(&mockPooler{pool: nil}, logger)

		reg := prometheus.NewRegistry()
		err := reg.Register(m)
		require.NoError(t, err)

		// Collect should not panic
		families, err := reg.Gather()
		require.NoError(t, err)
		assert.Empty(t, families)
	})

	// Note: Testing with a real pool requires integration setup which we skip for unit tests.
	// We rely on the integration/manual verification for actual metric values.
}
