package resilience

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/goleak"
)

func TestTimeout_Do_SuccessWithinTimeout(t *testing.T) {
	tests := []struct {
		name          string
		timeout       time.Duration
		operationTime time.Duration
		wantErr       error
	}{
		{
			name:          "completes immediately",
			timeout:       100 * time.Millisecond,
			operationTime: 0,
			wantErr:       nil,
		},
		{
			name:          "completes well before timeout",
			timeout:       100 * time.Millisecond,
			operationTime: 10 * time.Millisecond,
			wantErr:       nil,
		},
		{
			name:          "completes just before timeout",
			timeout:       50 * time.Millisecond,
			operationTime: 20 * time.Millisecond,
			wantErr:       nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			to := NewTimeout("test", tt.timeout)
			ctx := context.Background()

			err := to.Do(ctx, func(ctx context.Context) error {
				if tt.operationTime > 0 {
					select {
					case <-time.After(tt.operationTime):
					case <-ctx.Done():
						return ctx.Err()
					}
				}
				return nil
			})

			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Do() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTimeout_Do_ExceedsTimeout(t *testing.T) {
	to := NewTimeout("test", 10*time.Millisecond)
	ctx := context.Background()

	err := to.Do(ctx, func(ctx context.Context) error {
		select {
		case <-time.After(100 * time.Millisecond):
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	})

	if !errors.Is(err, ErrTimeoutExceeded) {
		t.Errorf("Do() error = %v, want ErrTimeoutExceeded", err)
	}
}

func TestTimeout_Do_OperationReturnsError(t *testing.T) {
	to := NewTimeout("test", 100*time.Millisecond)
	ctx := context.Background()
	testErr := errors.New("operation failed")

	err := to.Do(ctx, func(ctx context.Context) error {
		return testErr
	})

	if !errors.Is(err, testErr) {
		t.Errorf("Do() error = %v, want %v", err, testErr)
	}

	// Should NOT be wrapped as timeout error
	if errors.Is(err, ErrTimeoutExceeded) {
		t.Error("Do() should not wrap non-timeout errors as ErrTimeoutExceeded")
	}
}

func TestTimeout_Do_ParentContextCancellation(t *testing.T) {
	to := NewTimeout("test", 100*time.Millisecond)
	ctx, cancel := context.WithCancel(context.Background())

	// Cancel immediately
	cancel()

	err := to.Do(ctx, func(ctx context.Context) error {
		select {
		case <-time.After(50 * time.Millisecond):
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	})

	// Should return context.Canceled, NOT ErrTimeoutExceeded
	if !errors.Is(err, context.Canceled) {
		t.Errorf("Do() error = %v, want context.Canceled", err)
	}

	if errors.Is(err, ErrTimeoutExceeded) {
		t.Error("Do() should not wrap context.Canceled as ErrTimeoutExceeded")
	}
}

func TestTimeout_Do_ParentContextShorterDeadline(t *testing.T) {
	to := NewTimeout("test", 100*time.Millisecond)

	// Parent context with shorter deadline
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	err := to.Do(ctx, func(ctx context.Context) error {
		select {
		case <-time.After(50 * time.Millisecond):
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	})

	// Parent deadline triggered first, should be RES-003
	if !errors.Is(err, ErrTimeoutExceeded) {
		t.Errorf("Do() error = %v, want ErrTimeoutExceeded", err)
	}
}

func TestTimeout_Name(t *testing.T) {
	to := NewTimeout("database_query", 5*time.Second)

	if got := to.Name(); got != "database_query" {
		t.Errorf("Name() = %v, want database_query", got)
	}
}

func TestTimeout_Duration(t *testing.T) {
	to := NewTimeout("test", 5*time.Second)

	if got := to.Duration(); got != 5*time.Second {
		t.Errorf("Duration() = %v, want 5s", got)
	}
}

func TestTimeoutPresets_ForDatabase(t *testing.T) {
	cfg := TimeoutConfig{
		Default:     30 * time.Second,
		Database:    5 * time.Second,
		ExternalAPI: 10 * time.Second,
	}

	presets := NewTimeoutPresets(cfg)
	dbTimeout := presets.ForDatabase()

	if dbTimeout.Name() != "database" {
		t.Errorf("ForDatabase().Name() = %v, want database", dbTimeout.Name())
	}

	if dbTimeout.Duration() != 5*time.Second {
		t.Errorf("ForDatabase().Duration() = %v, want 5s", dbTimeout.Duration())
	}
}

func TestTimeoutPresets_ForExternalAPI(t *testing.T) {
	cfg := TimeoutConfig{
		Default:     30 * time.Second,
		Database:    5 * time.Second,
		ExternalAPI: 10 * time.Second,
	}

	presets := NewTimeoutPresets(cfg)
	apiTimeout := presets.ForExternalAPI()

	if apiTimeout.Name() != "external_api" {
		t.Errorf("ForExternalAPI().Name() = %v, want external_api", apiTimeout.Name())
	}

	if apiTimeout.Duration() != 10*time.Second {
		t.Errorf("ForExternalAPI().Duration() = %v, want 10s", apiTimeout.Duration())
	}
}

func TestTimeoutPresets_Default(t *testing.T) {
	cfg := TimeoutConfig{
		Default:     30 * time.Second,
		Database:    5 * time.Second,
		ExternalAPI: 10 * time.Second,
	}

	presets := NewTimeoutPresets(cfg)
	defaultTimeout := presets.Default()

	if defaultTimeout.Name() != "default" {
		t.Errorf("Default().Name() = %v, want default", defaultTimeout.Name())
	}

	if defaultTimeout.Duration() != 30*time.Second {
		t.Errorf("Default().Duration() = %v, want 30s", defaultTimeout.Duration())
	}
}

func TestTimeoutPresets_ForOperation(t *testing.T) {
	cfg := TimeoutConfig{
		Default:     30 * time.Second,
		Database:    5 * time.Second,
		ExternalAPI: 10 * time.Second,
	}

	presets := NewTimeoutPresets(cfg)
	customTimeout := presets.ForOperation("custom_op", 15*time.Second)

	if customTimeout.Name() != "custom_op" {
		t.Errorf("ForOperation().Name() = %v, want custom_op", customTimeout.Name())
	}

	if customTimeout.Duration() != 15*time.Second {
		t.Errorf("ForOperation().Duration() = %v, want 15s", customTimeout.Duration())
	}
}

func TestTimeoutPresets_DurationMethods(t *testing.T) {
	cfg := TimeoutConfig{
		Default:     30 * time.Second,
		Database:    5 * time.Second,
		ExternalAPI: 10 * time.Second,
	}

	presets := NewTimeoutPresets(cfg)

	if got := presets.DatabaseDuration(); got != 5*time.Second {
		t.Errorf("DatabaseDuration() = %v, want 5s", got)
	}

	if got := presets.ExternalAPIDuration(); got != 10*time.Second {
		t.Errorf("ExternalAPIDuration() = %v, want 10s", got)
	}

	if got := presets.DefaultDuration(); got != 30*time.Second {
		t.Errorf("DefaultDuration() = %v, want 30s", got)
	}
}

func TestTimeoutPresets_WithOptions(t *testing.T) {
	cfg := TimeoutConfig{
		Default:     30 * time.Second,
		Database:    5 * time.Second,
		ExternalAPI: 10 * time.Second,
	}

	registry := prometheus.NewRegistry()
	metrics := NewTimeoutMetrics(registry)
	logger := slog.Default()

	// Create presets with options
	presets := NewTimeoutPresets(cfg, WithTimeoutMetrics(metrics), WithTimeoutLogger(logger))

	ctx := context.Background()

	// Test that ForDatabase timeout works with metrics
	dbTimeout := presets.ForDatabase()
	err := dbTimeout.Do(ctx, func(ctx context.Context) error {
		return nil
	})
	if err != nil {
		t.Errorf("ForDatabase().Do() unexpected error: %v", err)
	}

	// Test that ForOperation also gets the options
	customTimeout := presets.ForOperation("custom", 1*time.Second)
	err = customTimeout.Do(ctx, func(ctx context.Context) error {
		return nil
	})
	if err != nil {
		t.Errorf("ForOperation().Do() unexpected error: %v", err)
	}

	// Verify metrics were recorded (no panic means success)
}

func TestTimeout_WithMetrics(t *testing.T) {
	registry := prometheus.NewRegistry()
	metrics := NewTimeoutMetrics(registry)
	to := NewTimeout("test", 100*time.Millisecond, WithTimeoutMetrics(metrics))

	ctx := context.Background()

	// Successful operation
	err := to.Do(ctx, func(ctx context.Context) error {
		return nil
	})
	if err != nil {
		t.Errorf("Do() unexpected error: %v", err)
	}

	// Timeout operation
	err = to.Do(ctx, func(ctx context.Context) error {
		select {
		case <-time.After(200 * time.Millisecond):
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	})
	if !errors.Is(err, ErrTimeoutExceeded) {
		t.Errorf("Do() error = %v, want ErrTimeoutExceeded", err)
	}

	// Error operation
	testErr := errors.New("test error")
	err = to.Do(ctx, func(ctx context.Context) error {
		return testErr
	})
	if !errors.Is(err, testErr) {
		t.Errorf("Do() error = %v, want testErr", err)
	}
}

func TestTimeout_WithLogger(t *testing.T) {
	logger := slog.Default()
	to := NewTimeout("test", 100*time.Millisecond, WithTimeoutLogger(logger))

	ctx := context.Background()

	err := to.Do(ctx, func(ctx context.Context) error {
		return nil
	})

	if err != nil {
		t.Errorf("Do() unexpected error: %v", err)
	}
}

func TestTimeout_WithNilLogger(t *testing.T) {
	// Should not panic with nil logger
	to := NewTimeout("test", 100*time.Millisecond, WithTimeoutLogger(nil))

	ctx := context.Background()

	err := to.Do(ctx, func(ctx context.Context) error {
		return nil
	})

	if err != nil {
		t.Errorf("Do() unexpected error: %v", err)
	}
}

func TestTimeout_WithNilMetrics(t *testing.T) {
	// Should not panic with nil metrics
	to := NewTimeout("test", 100*time.Millisecond, WithTimeoutMetrics(nil))

	ctx := context.Background()

	// Test successful operation
	err := to.Do(ctx, func(ctx context.Context) error {
		return nil
	})

	if err != nil {
		t.Errorf("Do() unexpected error: %v", err)
	}

	// Test timeout operation - should still work without metrics
	err = to.Do(ctx, func(ctx context.Context) error {
		select {
		case <-time.After(200 * time.Millisecond):
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	})

	if !errors.Is(err, ErrTimeoutExceeded) {
		t.Errorf("Do() error = %v, want ErrTimeoutExceeded", err)
	}
}

func TestDoWithTimeout(t *testing.T) {
	to := NewTimeout("test", 100*time.Millisecond)
	ctx := context.Background()

	// Successful operation
	result, err := DoWithTimeout[string](to, ctx, func(ctx context.Context) (string, error) {
		return "success", nil
	})

	if err != nil {
		t.Errorf("DoWithTimeout() unexpected error: %v", err)
	}

	if result != "success" {
		t.Errorf("DoWithTimeout() result = %v, want success", result)
	}
}

func TestDoWithTimeout_Timeout(t *testing.T) {
	to := NewTimeout("test", 10*time.Millisecond)
	ctx := context.Background()

	result, err := DoWithTimeout[string](to, ctx, func(ctx context.Context) (string, error) {
		select {
		case <-time.After(100 * time.Millisecond):
			return "should not reach", nil
		case <-ctx.Done():
			return "", ctx.Err()
		}
	})

	if !errors.Is(err, ErrTimeoutExceeded) {
		t.Errorf("DoWithTimeout() error = %v, want ErrTimeoutExceeded", err)
	}

	if result != "" {
		t.Errorf("DoWithTimeout() result = %v, want empty string", result)
	}
}

func TestDoWithTimeout_Error(t *testing.T) {
	to := NewTimeout("test", 100*time.Millisecond)
	ctx := context.Background()
	testErr := errors.New("test error")

	result, err := DoWithTimeout[int](to, ctx, func(ctx context.Context) (int, error) {
		return 0, testErr
	})

	if !errors.Is(err, testErr) {
		t.Errorf("DoWithTimeout() error = %v, want testErr", err)
	}

	if result != 0 {
		t.Errorf("DoWithTimeout() result = %v, want 0", result)
	}
}

func TestTimeoutMetrics_RecordOperation(t *testing.T) {
	registry := prometheus.NewRegistry()
	metrics := NewTimeoutMetrics(registry)

	metrics.RecordOperation("test", "success", 0.5)
	metrics.RecordOperation("test", "timeout", 1.0)
	metrics.RecordOperation("test", "error", 0.1)

	// Just verify no panic - actual metric values can be verified via registry
}

func TestTimeoutMetrics_Reset(t *testing.T) {
	registry := prometheus.NewRegistry()
	metrics := NewTimeoutMetrics(registry)

	metrics.RecordOperation("test", "success", 0.5)
	metrics.Reset()

	// Just verify no panic
}

func TestNoopTimeoutMetrics(t *testing.T) {
	metrics := NoopTimeoutMetrics()

	// Should not panic
	metrics.RecordOperation("test", "success", 0.5)
	metrics.Reset()
}

// TestTimeout_NoGoroutineLeak verifies no goroutine leaks on timeout.
func TestTimeout_NoGoroutineLeak(t *testing.T) {
	defer goleak.VerifyNone(t,
		// Ignore known background goroutines
		goleak.IgnoreTopFunction("go.opencensus.io/stats/view.(*worker).start"),
		goleak.IgnoreTopFunction("internal/poll.runtime_pollWait"),
		// slog handler goroutines
		goleak.IgnoreTopFunction("log/slog.(*defaultHandler).Handle"),
	)

	to := NewTimeout("test", 10*time.Millisecond)
	ctx := context.Background()

	// Run several timeout operations
	for i := 0; i < 5; i++ {
		_ = to.Do(ctx, func(ctx context.Context) error {
			select {
			case <-time.After(100 * time.Millisecond):
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		})
	}

	// Wait for cleanup
	time.Sleep(50 * time.Millisecond)
}

// TestTimeout_NoGoroutineLeak_Success verifies no goroutine leaks on successful operations.
func TestTimeout_NoGoroutineLeak_Success(t *testing.T) {
	defer goleak.VerifyNone(t,
		goleak.IgnoreTopFunction("go.opencensus.io/stats/view.(*worker).start"),
		goleak.IgnoreTopFunction("internal/poll.runtime_pollWait"),
	)

	to := NewTimeout("test", 100*time.Millisecond)
	ctx := context.Background()

	// Run several successful operations
	for i := 0; i < 5; i++ {
		_ = to.Do(ctx, func(ctx context.Context) error {
			return nil
		})
	}
}
