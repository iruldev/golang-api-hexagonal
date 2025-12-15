package wrapper

import (
	"context"
	"errors"
	"testing"
	"time"
)

// mockRedisPinger implements RedisPinger for testing
type mockRedisPinger struct {
	pingFunc func(ctx context.Context) error
}

func (m *mockRedisPinger) Ping(ctx context.Context) error {
	if m.pingFunc != nil {
		return m.pingFunc(ctx)
	}
	return nil
}

func TestDoRedis_NormalExecution(t *testing.T) {
	t.Parallel()

	executed := false
	err := DoRedis(context.Background(), func(ctx context.Context) error {
		executed = true
		return nil
	})

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !executed {
		t.Error("function should have been executed")
	}
}

func TestDoRedis_CancelledContext_ReturnsImmediately(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := DoRedis(ctx, func(ctx context.Context) error {
		t.Error("function should not be called with cancelled context")
		return nil
	})

	if err == nil {
		t.Error("expected error for cancelled context")
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

func TestDoRedis_DeadlineExceeded_ReturnsImmediately(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
	defer cancel()

	err := DoRedis(ctx, func(ctx context.Context) error {
		t.Error("function should not be called with exceeded deadline")
		return nil
	})

	if err == nil {
		t.Error("expected error for exceeded deadline")
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("expected context.DeadlineExceeded, got %v", err)
	}
}

func TestDoRedis_PropagatesError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("redis error")
	err := DoRedis(context.Background(), func(ctx context.Context) error {
		return expectedErr
	})

	if err == nil {
		t.Error("expected error to be propagated")
	}
	if !errors.Is(err, expectedErr) {
		t.Errorf("expected %v, got %v", expectedErr, err)
	}
}

func TestDoRedisResult_NormalExecution(t *testing.T) {
	t.Parallel()

	result, err := DoRedisResult(context.Background(), func(ctx context.Context) (string, error) {
		return "value", nil
	})

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result != "value" {
		t.Errorf("expected 'value', got '%s'", result)
	}
}

func TestDoRedisResult_CancelledContext_ReturnsImmediately(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	result, err := DoRedisResult(ctx, func(ctx context.Context) (string, error) {
		t.Error("function should not be called with cancelled context")
		return "should not reach", nil
	})

	if err == nil {
		t.Error("expected error for cancelled context")
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled, got %v", err)
	}
	if result != "" {
		t.Errorf("expected zero value, got '%s'", result)
	}
}

func TestDoRedisResult_DeadlineExceeded_ReturnsImmediately(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
	defer cancel()

	result, err := DoRedisResult(ctx, func(ctx context.Context) (int, error) {
		t.Error("function should not be called with exceeded deadline")
		return 42, nil
	})

	if err == nil {
		t.Error("expected error for exceeded deadline")
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("expected context.DeadlineExceeded, got %v", err)
	}
	if result != 0 {
		t.Errorf("expected zero value, got %d", result)
	}
}

func TestDoRedisResult_PropagatesError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("redis error")
	result, err := DoRedisResult(context.Background(), func(ctx context.Context) (string, error) {
		return "", expectedErr
	})

	if err == nil {
		t.Error("expected error to be propagated")
	}
	if !errors.Is(err, expectedErr) {
		t.Errorf("expected %v, got %v", expectedErr, err)
	}
	if result != "" {
		t.Errorf("expected empty string, got '%s'", result)
	}
}

func TestPingRedis_NormalExecution(t *testing.T) {
	t.Parallel()

	mock := &mockRedisPinger{
		pingFunc: func(ctx context.Context) error {
			return nil
		},
	}

	err := PingRedis(context.Background(), mock)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestPingRedis_CancelledContext_ReturnsImmediately(t *testing.T) {
	t.Parallel()

	mock := &mockRedisPinger{
		pingFunc: func(ctx context.Context) error {
			t.Error("ping should not be called with cancelled context")
			return nil
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := PingRedis(ctx, mock)
	if err == nil {
		t.Error("expected error for cancelled context")
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

func TestPingRedis_DeadlineExceeded_ReturnsImmediately(t *testing.T) {
	t.Parallel()

	mock := &mockRedisPinger{
		pingFunc: func(ctx context.Context) error {
			t.Error("ping should not be called with exceeded deadline")
			return nil
		},
	}

	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
	defer cancel()

	err := PingRedis(ctx, mock)
	if err == nil {
		t.Error("expected error for exceeded deadline")
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("expected context.DeadlineExceeded, got %v", err)
	}
}

func TestPingRedis_PropagatesError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("redis connection error")
	mock := &mockRedisPinger{
		pingFunc: func(ctx context.Context) error {
			return expectedErr
		},
	}

	err := PingRedis(context.Background(), mock)
	if err == nil {
		t.Error("expected error to be propagated")
	}
	if !errors.Is(err, expectedErr) {
		t.Errorf("expected %v, got %v", expectedErr, err)
	}
}

func TestDoRedis_NoDeadline_AddsDefaultTimeout(t *testing.T) {
	t.Parallel()

	executed := false
	err := DoRedis(context.Background(), func(ctx context.Context) error {
		executed = true
		deadline, ok := ctx.Deadline()
		if !ok {
			t.Error("expected deadline to be set")
		}
		expected := time.Now().Add(DefaultRedisTimeout)
		if deadline.Before(time.Now()) || deadline.After(expected.Add(time.Second)) {
			t.Errorf("deadline %v not within expected range", deadline)
		}
		return nil
	})

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !executed {
		t.Error("function should have been executed")
	}
}

func TestDoRedisResult_NoDeadline_AddsDefaultTimeout(t *testing.T) {
	t.Parallel()

	_, err := DoRedisResult(context.Background(), func(ctx context.Context) (string, error) {
		deadline, ok := ctx.Deadline()
		if !ok {
			t.Error("expected deadline to be set")
		}
		expected := time.Now().Add(DefaultRedisTimeout)
		if deadline.Before(time.Now()) || deadline.After(expected.Add(time.Second)) {
			t.Errorf("deadline %v not within expected range", deadline)
		}
		return "value", nil
	})

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestPingRedis_NoDeadline_AddsDefaultTimeout(t *testing.T) {
	t.Parallel()

	mock := &mockRedisPinger{
		pingFunc: func(ctx context.Context) error {
			deadline, ok := ctx.Deadline()
			if !ok {
				t.Error("expected deadline to be set")
			}
			expected := time.Now().Add(DefaultRedisTimeout)
			if deadline.Before(time.Now()) || deadline.After(expected.Add(time.Second)) {
				t.Errorf("deadline %v not within expected range", deadline)
			}
			return nil
		},
	}

	err := PingRedis(context.Background(), mock)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
