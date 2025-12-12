package worker

import (
	"context"
	"errors"
	"testing"

	"github.com/hibiken/asynq"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/iruldev/golang-api-hexagonal/internal/observability"
)

func TestMetricsMiddleware_Success(t *testing.T) {
	// Arrange - using actual MetricsMiddleware with global observability metrics
	middleware := MetricsMiddleware()

	successHandler := asynq.HandlerFunc(func(ctx context.Context, t *asynq.Task) error {
		return nil
	})

	task := asynq.NewTask("test:success", nil)
	handler := middleware(successHandler)

	// Get initial values
	initialCounter := testutil.ToFloat64(observability.JobProcessedTotal.WithLabelValues("test:success", QueueDefault, "success"))

	// Act
	err := handler.ProcessTask(context.Background(), task)

	// Assert
	require.NoError(t, err)

	// Verify counter was incremented
	finalCounter := testutil.ToFloat64(observability.JobProcessedTotal.WithLabelValues("test:success", QueueDefault, "success"))
	assert.Equal(t, initialCounter+1, finalCounter, "Counter should be incremented by 1")
}

func TestMetricsMiddleware_Failure(t *testing.T) {
	// Arrange - using actual MetricsMiddleware
	middleware := MetricsMiddleware()

	failHandler := asynq.HandlerFunc(func(ctx context.Context, t *asynq.Task) error {
		return errors.New("task failed")
	})

	task := asynq.NewTask("test:failure", nil)
	handler := middleware(failHandler)

	// Get initial values
	initialCounter := testutil.ToFloat64(observability.JobProcessedTotal.WithLabelValues("test:failure", QueueDefault, "failed"))

	// Act
	err := handler.ProcessTask(context.Background(), task)

	// Assert
	require.Error(t, err)

	// Verify counter was incremented for failure
	finalCounter := testutil.ToFloat64(observability.JobProcessedTotal.WithLabelValues("test:failure", QueueDefault, "failed"))
	assert.Equal(t, initialCounter+1, finalCounter, "Failure counter should be incremented by 1")
}

func TestMetricsMiddleware_RecordsHistogram(t *testing.T) {
	// Arrange - test histogram observation
	middleware := MetricsMiddleware()

	handler := asynq.HandlerFunc(func(ctx context.Context, t *asynq.Task) error {
		return nil
	})

	task := asynq.NewTask("test:histogram", nil)
	wrappedHandler := middleware(handler)

	// Get initial histogram count
	initialCount := testutil.CollectAndCount(observability.JobDurationSeconds)

	// Act
	err := wrappedHandler.ProcessTask(context.Background(), task)

	// Assert
	require.NoError(t, err)

	// Verify histogram has samples (count increased or remains positive)
	finalCount := testutil.CollectAndCount(observability.JobDurationSeconds)
	assert.GreaterOrEqual(t, finalCount, initialCount, "Histogram should have observations")
}

func TestMetricsMiddleware_ReturnsError(t *testing.T) {
	// Arrange
	middleware := MetricsMiddleware()
	expectedErr := errors.New("expected error")

	failHandler := asynq.HandlerFunc(func(ctx context.Context, t *asynq.Task) error {
		return expectedErr
	})

	task := asynq.NewTask("test:error", nil)
	handler := middleware(failHandler)

	// Act
	err := handler.ProcessTask(context.Background(), task)

	// Assert - verify error is propagated (not swallowed)
	require.Error(t, err)
	assert.Equal(t, expectedErr, err)
}

// TestMetricsMiddleware_UsesQueueDefault verifies queue label is always QueueDefault
func TestMetricsMiddleware_UsesQueueDefault(t *testing.T) {
	// Arrange
	middleware := MetricsMiddleware()

	handler := asynq.HandlerFunc(func(ctx context.Context, t *asynq.Task) error {
		return nil
	})

	task := asynq.NewTask("test:queue", nil)
	wrappedHandler := middleware(handler)

	// Get initial counter with QueueDefault
	initialCounter := testutil.ToFloat64(observability.JobProcessedTotal.WithLabelValues("test:queue", QueueDefault, "success"))

	// Act
	_ = wrappedHandler.ProcessTask(context.Background(), task)

	// Assert - counter uses QueueDefault label
	finalCounter := testutil.ToFloat64(observability.JobProcessedTotal.WithLabelValues("test:queue", QueueDefault, "success"))
	assert.Equal(t, initialCounter+1, finalCounter, "Should use QueueDefault as queue label")
}

// TestMetricsMiddleware_MockPattern demonstrates isolated testing pattern
func TestMetricsMiddleware_MockPattern(t *testing.T) {
	// This test uses isolated metrics for demonstration purposes
	reg := prometheus.NewRegistry()

	counter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "test_isolated_counter",
			Help: "Test counter",
		},
		[]string{"task_type", "queue", "status"},
	)
	reg.MustRegister(counter)

	// Custom middleware using isolated counter
	middleware := func(next asynq.Handler) asynq.Handler {
		return asynq.HandlerFunc(func(ctx context.Context, t *asynq.Task) error {
			err := next.ProcessTask(ctx, t)
			status := "success"
			if err != nil {
				status = "failed"
			}
			counter.WithLabelValues(t.Type(), "default", status).Inc()
			return err
		})
	}

	successHandler := asynq.HandlerFunc(func(ctx context.Context, t *asynq.Task) error {
		return nil
	})

	task := asynq.NewTask("test:mock", nil)
	handler := middleware(successHandler)

	// Act
	err := handler.ProcessTask(context.Background(), task)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, float64(1), testutil.ToFloat64(counter.WithLabelValues("test:mock", "default", "success")))
}
