package wrapper

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// mockRows implements pgx.Rows for testing
type mockRows struct {
	closed bool
}

func (m *mockRows) Close()                                       { m.closed = true }
func (m *mockRows) Err() error                                   { return nil }
func (m *mockRows) CommandTag() pgconn.CommandTag                { return pgconn.CommandTag{} }
func (m *mockRows) FieldDescriptions() []pgconn.FieldDescription { return nil }
func (m *mockRows) Next() bool                                   { return false }
func (m *mockRows) Scan(dest ...any) error                       { return nil }
func (m *mockRows) Values() ([]any, error)                       { return nil, nil }
func (m *mockRows) RawValues() [][]byte                          { return nil }
func (m *mockRows) Conn() *pgx.Conn                              { return nil }

// mockRow implements pgx.Row for testing
type mockRow struct {
	err error
}

func (m *mockRow) Scan(dest ...any) error { return m.err }

// mockQuerier implements Querier for testing
type mockQuerier struct {
	queryFunc    func(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	queryRowFunc func(ctx context.Context, sql string, args ...any) pgx.Row
	execFunc     func(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

func (m *mockQuerier) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	if m.queryFunc != nil {
		return m.queryFunc(ctx, sql, args...)
	}
	return &mockRows{}, nil
}

func (m *mockQuerier) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	if m.queryRowFunc != nil {
		return m.queryRowFunc(ctx, sql, args...)
	}
	return &mockRow{}
}

func (m *mockQuerier) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	if m.execFunc != nil {
		return m.execFunc(ctx, sql, args...)
	}
	return pgconn.CommandTag{}, nil
}

func TestQuery_NoDeadline_AddsDefaultTimeout(t *testing.T) {
	t.Parallel()

	mock := &mockQuerier{
		queryFunc: func(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
			// Verify timeout was applied
			deadline, ok := ctx.Deadline()
			if !ok {
				t.Error("expected deadline to be set")
			}
			// Should be approximately 30s from now (with some tolerance)
			expected := time.Now().Add(DefaultQueryTimeout)
			if deadline.Before(time.Now()) || deadline.After(expected.Add(time.Second)) {
				t.Errorf("deadline %v not within expected range", deadline)
			}
			return &mockRows{}, nil
		},
	}

	ctx := context.Background() // no deadline
	_, err := Query(ctx, mock, "SELECT 1")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestQuery_WithDeadline_PreservesExisting(t *testing.T) {
	t.Parallel()

	existingDeadline := time.Now().Add(5 * time.Second)

	mock := &mockQuerier{
		queryFunc: func(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
			deadline, ok := ctx.Deadline()
			if !ok {
				t.Error("expected deadline to be preserved")
			}
			// Should be the existing deadline, not a new one
			if !deadline.Equal(existingDeadline) {
				t.Errorf("deadline %v should equal existing %v", deadline, existingDeadline)
			}
			return &mockRows{}, nil
		},
	}

	ctx, cancel := context.WithDeadline(context.Background(), existingDeadline)
	defer cancel()

	_, err := Query(ctx, mock, "SELECT 1")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestQuery_CancelledContext_ReturnsImmediately(t *testing.T) {
	t.Parallel()

	mock := &mockQuerier{
		queryFunc: func(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
			t.Error("query should not be called with cancelled context")
			return nil, nil
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	_, err := Query(ctx, mock, "SELECT 1")
	if err == nil {
		t.Error("expected error for cancelled context")
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

func TestQuery_DeadlineExceeded_ReturnsImmediately(t *testing.T) {
	t.Parallel()

	mock := &mockQuerier{
		queryFunc: func(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
			t.Error("query should not be called with exceeded deadline")
			return nil, nil
		},
	}

	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
	defer cancel()

	_, err := Query(ctx, mock, "SELECT 1")
	if err == nil {
		t.Error("expected error for exceeded deadline")
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("expected context.DeadlineExceeded, got %v", err)
	}
}

func TestQueryRow_NoDeadline_AddsDefaultTimeout(t *testing.T) {
	t.Parallel()

	mock := &mockQuerier{
		queryRowFunc: func(ctx context.Context, sql string, args ...any) pgx.Row {
			_, ok := ctx.Deadline()
			if !ok {
				t.Error("expected deadline to be set")
			}
			return &mockRow{}
		},
	}

	ctx := context.Background()
	row := QueryRow(ctx, mock, "SELECT 1")
	if row == nil {
		t.Error("expected row to be returned")
	}
}

func TestQueryRowWithCancel_NoDeadline_ReturnsCancelFunc(t *testing.T) {
	t.Parallel()

	mock := &mockQuerier{
		queryRowFunc: func(ctx context.Context, sql string, args ...any) pgx.Row {
			_, ok := ctx.Deadline()
			if !ok {
				t.Error("expected deadline to be set")
			}
			return &mockRow{}
		},
	}

	ctx := context.Background()
	row, cancel := QueryRowWithCancel(ctx, mock, "SELECT 1")
	defer cancel()

	if row == nil {
		t.Error("expected row to be returned")
	}
}

func TestQueryRowWithCancel_WithDeadline_NoopCancel(t *testing.T) {
	t.Parallel()

	existingDeadline := time.Now().Add(5 * time.Second)

	mock := &mockQuerier{
		queryRowFunc: func(ctx context.Context, sql string, args ...any) pgx.Row {
			deadline, ok := ctx.Deadline()
			if !ok {
				t.Error("expected deadline to be preserved")
			}
			if !deadline.Equal(existingDeadline) {
				t.Errorf("deadline %v should equal existing %v", deadline, existingDeadline)
			}
			return &mockRow{}
		},
	}

	ctx, ctxCancel := context.WithDeadline(context.Background(), existingDeadline)
	defer ctxCancel()

	row, cancel := QueryRowWithCancel(ctx, mock, "SELECT 1")
	cancel() // should be noop

	if row == nil {
		t.Error("expected row to be returned")
	}
}

func TestExec_NoDeadline_AddsDefaultTimeout(t *testing.T) {
	t.Parallel()

	mock := &mockQuerier{
		execFunc: func(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
			_, ok := ctx.Deadline()
			if !ok {
				t.Error("expected deadline to be set")
			}
			return pgconn.CommandTag{}, nil
		},
	}

	ctx := context.Background()
	_, err := Exec(ctx, mock, "UPDATE users SET name = $1", "test")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestExec_WithDeadline_PreservesExisting(t *testing.T) {
	t.Parallel()

	existingDeadline := time.Now().Add(5 * time.Second)

	mock := &mockQuerier{
		execFunc: func(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
			deadline, ok := ctx.Deadline()
			if !ok {
				t.Error("expected deadline to be preserved")
			}
			if !deadline.Equal(existingDeadline) {
				t.Errorf("deadline %v should equal existing %v", deadline, existingDeadline)
			}
			return pgconn.CommandTag{}, nil
		},
	}

	ctx, cancel := context.WithDeadline(context.Background(), existingDeadline)
	defer cancel()

	_, err := Exec(ctx, mock, "UPDATE users SET name = $1", "test")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestExec_CancelledContext_ReturnsImmediately(t *testing.T) {
	t.Parallel()

	mock := &mockQuerier{
		execFunc: func(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
			t.Error("exec should not be called with cancelled context")
			return pgconn.CommandTag{}, nil
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := Exec(ctx, mock, "UPDATE users SET name = $1", "test")
	if err == nil {
		t.Error("expected error for cancelled context")
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

func TestExec_DeadlineExceeded_ReturnsImmediately(t *testing.T) {
	t.Parallel()

	mock := &mockQuerier{
		execFunc: func(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
			t.Error("exec should not be called with exceeded deadline")
			return pgconn.CommandTag{}, nil
		},
	}

	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
	defer cancel()

	_, err := Exec(ctx, mock, "UPDATE users SET name = $1", "test")
	if err == nil {
		t.Error("expected error for exceeded deadline")
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("expected context.DeadlineExceeded, got %v", err)
	}
}
