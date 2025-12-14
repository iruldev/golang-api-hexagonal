package worker

import (
	"context"
	"errors"
	"testing"

	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestLoggingMiddleware(t *testing.T) {
	// Arrange
	logger := zap.NewNop()
	middleware := LoggingMiddleware(logger)

	// Assert
	assert.NotNil(t, middleware)
}

func TestRecoveryMiddleware(t *testing.T) {
	// Arrange
	logger := zap.NewNop()
	middleware := RecoveryMiddleware(logger)

	// Assert
	assert.NotNil(t, middleware)
}

func TestRecoveryMiddleware_RecoversPanic(t *testing.T) {
	// Arrange
	logger := zap.NewNop()
	middleware := RecoveryMiddleware(logger)

	panicHandler := asynq.HandlerFunc(func(ctx context.Context, task *asynq.Task) error {
		panic("test panic")
	})

	wrappedHandler := middleware(panicHandler)
	task := asynq.NewTask("test:panic", nil)

	// Act
	err := wrappedHandler.ProcessTask(context.Background(), task)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "panic recovered")
}

func TestLoggingMiddleware_LogsSuccess(t *testing.T) {
	// Arrange
	logger := zap.NewNop()
	middleware := LoggingMiddleware(logger)

	successHandler := asynq.HandlerFunc(func(ctx context.Context, task *asynq.Task) error {
		return nil
	})

	wrappedHandler := middleware(successHandler)
	task := asynq.NewTask("test:success", nil)

	// Act
	err := wrappedHandler.ProcessTask(context.Background(), task)

	// Assert
	assert.NoError(t, err)
}

func TestLoggingMiddleware_LogsFailure(t *testing.T) {
	// Arrange
	logger := zap.NewNop()
	middleware := LoggingMiddleware(logger)

	expectedErr := errors.New("task failed")
	failHandler := asynq.HandlerFunc(func(ctx context.Context, task *asynq.Task) error {
		return expectedErr
	})

	wrappedHandler := middleware(failHandler)
	task := asynq.NewTask("test:fail", nil)

	// Act
	err := wrappedHandler.ProcessTask(context.Background(), task)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
}

func TestTracingMiddleware(t *testing.T) {
	// Arrange
	middleware := TracingMiddleware()

	// Assert
	assert.NotNil(t, middleware)
}

func TestTracingMiddleware_CreatesSpan(t *testing.T) {
	// Arrange
	middleware := TracingMiddleware()

	successHandler := asynq.HandlerFunc(func(ctx context.Context, task *asynq.Task) error {
		return nil
	})

	wrappedHandler := middleware(successHandler)
	task := asynq.NewTask("test:traced", nil)

	// Act
	err := wrappedHandler.ProcessTask(context.Background(), task)

	// Assert
	assert.NoError(t, err)
}
