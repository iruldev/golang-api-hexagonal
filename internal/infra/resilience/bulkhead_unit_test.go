package resilience

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"testing"
)

// Unit tests for bulkhead basic logic, construction, and configuration.

func TestNewBulkhead(t *testing.T) {
	// Note: MaxConcurrent and MaxWaiting behavior is verified through
	// behavioral tests (TestBulkhead_Do_RejectedWhenFull, TestBulkhead_Do_WaitsForSlot, etc.)
	// This test focuses on basic construction and initial state.
	tests := []struct {
		name  string
		bName string
		cfg   BulkheadConfig
	}{
		{
			name:  "creates bulkhead with config values",
			bName: "test-bulkhead",
			cfg:   BulkheadConfig{MaxConcurrent: 5, MaxWaiting: 10},
		},
		{
			name:  "creates bulkhead with default-like values",
			bName: "default-bulkhead",
			cfg:   BulkheadConfig{MaxConcurrent: 10, MaxWaiting: 100},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := NewBulkhead(tt.bName, tt.cfg)
			if b.Name() != tt.bName {
				t.Errorf("Name() = %v, want %v", b.Name(), tt.bName)
			}
			if b.ActiveCount() != 0 {
				t.Errorf("ActiveCount() = %v, want 0", b.ActiveCount())
			}
			if b.WaitingCount() != 0 {
				t.Errorf("WaitingCount() = %v, want 0", b.WaitingCount())
			}
		})
	}
}

func TestBulkhead_Do_Success(t *testing.T) {
	b := NewBulkhead("test", BulkheadConfig{MaxConcurrent: 5, MaxWaiting: 10})

	called := false
	err := b.Do(context.Background(), func(ctx context.Context) error {
		called = true
		return nil
	})

	if err != nil {
		t.Errorf("Do() error = %v, want nil", err)
	}
	if !called {
		t.Error("Do() did not call the function")
	}
}

func TestBulkhead_Do_Error(t *testing.T) {
	b := NewBulkhead("test", BulkheadConfig{MaxConcurrent: 5, MaxWaiting: 10})
	expectedErr := errors.New("test error")

	err := b.Do(context.Background(), func(ctx context.Context) error {
		return expectedErr
	})

	if !errors.Is(err, expectedErr) {
		t.Errorf("Do() error = %v, want %v", err, expectedErr)
	}
}

func TestBulkhead_WithOptions(t *testing.T) {
	metrics := NoopBulkheadMetrics()

	t.Run("WithBulkheadMetrics", func(t *testing.T) {
		b := NewBulkhead("test", BulkheadConfig{MaxConcurrent: 5, MaxWaiting: 10},
			WithBulkheadMetrics(metrics))

		err := b.Do(context.Background(), func(ctx context.Context) error {
			return nil
		})

		if err != nil {
			t.Errorf("Do() error = %v, want nil", err)
		}
	})

	t.Run("WithBulkheadLogger", func(t *testing.T) {
		b := NewBulkhead("test", BulkheadConfig{MaxConcurrent: 5, MaxWaiting: 10},
			WithBulkheadLogger(nil)) // Should use default

		err := b.Do(context.Background(), func(ctx context.Context) error {
			return nil
		})

		if err != nil {
			t.Errorf("Do() error = %v, want nil", err)
		}
	})

	t.Run("WithNilMetrics", func(t *testing.T) {
		b := NewBulkhead("test", BulkheadConfig{MaxConcurrent: 5, MaxWaiting: 10},
			WithBulkheadMetrics(nil)) // Should be noop

		err := b.Do(context.Background(), func(ctx context.Context) error {
			return nil
		})

		if err != nil {
			t.Errorf("Do() error = %v, want nil", err)
		}
	})
}

func TestBulkheadPresets(t *testing.T) {
	cfg := BulkheadConfig{MaxConcurrent: 10, MaxWaiting: 100}
	presets := NewBulkheadPresets(cfg)

	t.Run("ForDatabase returns bulkhead", func(t *testing.T) {
		b := presets.ForDatabase()
		if b.Name() != "database" {
			t.Errorf("ForDatabase().Name() = %v, want database", b.Name())
		}
	})

	t.Run("ForExternalAPI returns bulkhead", func(t *testing.T) {
		b := presets.ForExternalAPI()
		if b.Name() != "external_api" {
			t.Errorf("ForExternalAPI().Name() = %v, want external_api", b.Name())
		}
	})

	t.Run("Default returns bulkhead", func(t *testing.T) {
		b := presets.Default()
		if b.Name() != "default" {
			t.Errorf("Default().Name() = %v, want default", b.Name())
		}
	})

	t.Run("ForOperation creates custom bulkhead", func(t *testing.T) {
		b := presets.ForOperation("custom", 3, 5)
		if b.Name() != "custom" {
			t.Errorf("ForOperation().Name() = %v, want custom", b.Name())
		}
	})
}

func TestBulkheadPresets_WithOptions(t *testing.T) {
	cfg := BulkheadConfig{MaxConcurrent: 5, MaxWaiting: 10}
	metrics := NoopBulkheadMetrics()
	presets := NewBulkheadPresets(cfg, WithBulkheadMetrics(metrics))

	// All preset bulkheads should have metrics
	err := presets.ForDatabase().Do(context.Background(), func(ctx context.Context) error {
		return nil
	})
	if err != nil {
		t.Errorf("ForDatabase().Do() error = %v, want nil", err)
	}

	err = presets.ForExternalAPI().Do(context.Background(), func(ctx context.Context) error {
		return nil
	})
	if err != nil {
		t.Errorf("ForExternalAPI().Do() error = %v, want nil", err)
	}
}

func TestDoWithBulkhead(t *testing.T) {
	b := NewBulkhead("test", BulkheadConfig{MaxConcurrent: 5, MaxWaiting: 10})

	t.Run("returns result on success", func(t *testing.T) {
		result, err := DoWithBulkhead(b, context.Background(), func(ctx context.Context) (string, error) {
			return "hello", nil
		})

		if err != nil {
			t.Errorf("DoWithBulkhead() error = %v, want nil", err)
		}
		if result != "hello" {
			t.Errorf("DoWithBulkhead() result = %v, want hello", result)
		}
	})

	t.Run("returns error on failure", func(t *testing.T) {
		expectedErr := errors.New("test error")
		result, err := DoWithBulkhead(b, context.Background(), func(ctx context.Context) (string, error) {
			return "", expectedErr
		})

		if !errors.Is(err, expectedErr) {
			t.Errorf("DoWithBulkhead() error = %v, want %v", err, expectedErr)
		}
		if result != "" {
			t.Errorf("DoWithBulkhead() result = %v, want empty string", result)
		}
	})
}

func TestBulkheadMetrics_Operations(t *testing.T) {
	metrics := NewBulkheadMetrics(nil)

	t.Run("SetActive", func(t *testing.T) {
		metrics.SetActive("test", 5)
		// No panic = success (gauge is internal)
	})

	t.Run("SetWaiting", func(t *testing.T) {
		metrics.SetWaiting("test", 3)
		// No panic = success
	})

	t.Run("RecordOperation", func(t *testing.T) {
		metrics.RecordOperation("test", "success")
		metrics.RecordOperation("test", "rejected")
		metrics.RecordOperation("test", "error")
		// No panic = success
	})

	t.Run("RecordWaitDuration", func(t *testing.T) {
		metrics.RecordWaitDuration("test", 0.05)
		// No panic = success
	})

	t.Run("Reset", func(t *testing.T) {
		metrics.Reset()
		// No panic = success
	})
}

func TestNoopBulkheadMetrics(t *testing.T) {
	metrics := NoopBulkheadMetrics()
	if metrics == nil {
		t.Error("NoopBulkheadMetrics() returned nil")
	}
}

// TestNewBulkhead_InvalidConfig verifies that NewBulkhead panics on invalid configuration.
func TestNewBulkhead_InvalidConfig(t *testing.T) {
	t.Run("MaxConcurrent zero panics", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected panic for MaxConcurrent=0 but none occurred")
			}
		}()
		NewBulkhead("test", BulkheadConfig{MaxConcurrent: 0, MaxWaiting: 10})
	})

	t.Run("MaxConcurrent negative panics", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected panic for MaxConcurrent=-1 but none occurred")
			}
		}()
		NewBulkhead("test", BulkheadConfig{MaxConcurrent: -1, MaxWaiting: 10})
	})

	t.Run("MaxWaiting negative panics", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected panic for MaxWaiting=-1 but none occurred")
			}
		}()
		NewBulkhead("test", BulkheadConfig{MaxConcurrent: 5, MaxWaiting: -1})
	})

	t.Run("MaxWaiting zero is valid", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Unexpected panic for MaxWaiting=0: %v", r)
			}
		}()
		b := NewBulkhead("test", BulkheadConfig{MaxConcurrent: 5, MaxWaiting: 0})
		if b == nil {
			t.Error("Expected valid bulkhead, got nil")
		}
	})
}

// TestBulkhead_WithNonNilLogger verifies that a custom non-nil logger is used.
func TestBulkhead_WithNonNilLogger(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	customLogger := slog.New(handler)

	b := NewBulkhead("test-logger", BulkheadConfig{MaxConcurrent: 5, MaxWaiting: 10},
		WithBulkheadLogger(customLogger))

	err := b.Do(context.Background(), func(ctx context.Context) error {
		return nil
	})

	if err != nil {
		t.Errorf("Do() error = %v, want nil", err)
	}

	// Verify custom logger was used (should have log output)
	if buf.Len() == 0 {
		t.Error("Custom logger was not used - no log output")
	}

	if !bytes.Contains(buf.Bytes(), []byte("test-logger")) {
		t.Error("Log output doesn't contain bulkhead name")
	}
}
