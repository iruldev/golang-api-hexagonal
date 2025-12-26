package postgres

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

// mockPool implements the Pooler interface for testing
type mockPool struct {
	pingErr     error
	closeCalled bool
}

func (m *mockPool) Ping(_ context.Context) error {
	return m.pingErr
}

func (m *mockPool) Close() {
	m.closeCalled = true
}

func (m *mockPool) Pool() *pgxpool.Pool {
	return nil
}

// TestResilientPool_PingDoesNotClosePool verifies that Ping() no longer closes
// the pool on failure, solving the stale reference bug.
func TestResilientPool_PingDoesNotClosePool(t *testing.T) {
	t.Run("ping failure preserves pool connection", func(t *testing.T) {
		mock := &mockPool{pingErr: errors.New("connection lost")}
		db := &ResilientPool{
			log: slog.Default(),
			// In this test path, we pre-set the pool, so creator isn't called
			// unless we start with nil pool.
			poolCreator: func(ctx context.Context, dsn string) (Pooler, error) {
				return mock, nil
			},
		}

		// Pre-condition: Pool is already established
		db.pool = mock

		// Execute Ping which fails
		err := db.Ping(context.Background())

		// Assertions
		if err == nil {
			t.Fatal("expected ping error")
		}

		if mock.closeCalled {
			t.Error("CRITICAL: Pool.Close() was called on Ping failure! (Destructive reset still active)")
		}

		db.mu.RLock()
		currentPool := db.pool
		db.mu.RUnlock()

		if currentPool == nil {
			t.Error("CRITICAL: db.pool was set to nil on Ping failure! (Destructive reset still active)")
		}
	})
}

// TestResilientPool_PoolGetter verifies the Pool() getter logic.
func TestResilientPool_PoolGetter(t *testing.T) {
	t.Run("Pool() returns current pool safely", func(t *testing.T) {
		mock := &mockPool{}
		db := &ResilientPool{
			log: slog.Default(),
		}
		db.pool = mock // Set directly

		p := db.Pool()
		if p != nil {
			t.Errorf("expected Pool() to return nil (from mock)")
		}
	})

	t.Run("Pool() returns nil if not connected", func(t *testing.T) {
		db := &ResilientPool{
			log: slog.Default(),
		}
		if db.Pool() != nil {
			t.Errorf("expected nil pool when not connected")
		}
	})

	t.Run("Concurrent access is safe", func(t *testing.T) {
		mock := &mockPool{}
		db := &ResilientPool{
			log: slog.Default(),
		}
		db.pool = mock

		var wg sync.WaitGroup
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_ = db.Pool()
			}()
		}
		wg.Wait()
	})
}

// TestResilientPool_LazyConnection checks the lazy connection logic
func TestResilientPool_LazyConnection(t *testing.T) {
	t.Run("Ping creates pool if nil", func(t *testing.T) {
		mock := &mockPool{}
		creatorCalled := false

		db := &ResilientPool{
			dsn: "postgres://mock",
			log: slog.Default(),
			poolCreator: func(ctx context.Context, dsn string) (Pooler, error) {
				creatorCalled = true
				if dsn != "postgres://mock" {
					t.Errorf("unexpected dsn: %s", dsn)
				}
				return mock, nil
			},
		}

		// Initial state
		if db.pool != nil {
			t.Fatal("pool should be nil initially")
		}

		// First Ping should create pool
		err := db.Ping(context.Background())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !creatorCalled {
			t.Error("poolCreator should have been called")
		}

		if db.pool != mock {
			t.Error("pool should have been set to mock")
		}
	})
}
