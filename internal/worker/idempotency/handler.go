package idempotency

import (
	"context"
	"time"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"
)

// KeyExtractor is a function that extracts an idempotency key from a task.
// Return empty string to skip idempotency check (task will be processed normally).
type KeyExtractor func(*asynq.Task) string

// HandlerOption is a functional option for configuring IdempotentHandler.
type HandlerOption func(*handlerConfig)

type handlerConfig struct {
	logger   *zap.Logger
	failMode FailMode
}

// WithHandlerLogger sets the logger for logging duplicate detection.
func WithHandlerLogger(logger *zap.Logger) HandlerOption {
	return func(c *handlerConfig) {
		c.logger = logger
	}
}

// WithHandlerFailMode sets the fail mode for the handler.
// Note: This controls behavior when store.Check() returns an error.
// If the store also has a FailMode (e.g., RedisStore), they work together:
// - Store's FailMode: Determines if store.Check() returns error or (true, nil) on Redis failure
// - Handler's FailMode: Determines behavior if store.Check() returns an error
// For most cases, set the same FailMode on both store and handler.
func WithHandlerFailMode(mode FailMode) HandlerOption {
	return func(c *handlerConfig) {
		c.failMode = mode
	}
}

// IdempotentHandler wraps an asynq handler with idempotency checking.
// Duplicate tasks are skipped without error.
//
// Parameters:
//   - store: Idempotency store for checking/storing keys
//   - keyExtractor: Function to extract idempotency key from task (return "" to skip check)
//   - ttl: Time-to-live for idempotency keys
//   - handler: Original task handler
//   - opts: Optional configuration
//
// Behavior:
//   - If keyExtractor returns empty string, task is processed normally (no idempotency)
//   - If key is new, task is processed and completion is tracked
//   - If key exists (duplicate), task is skipped with nil error (no retry)
//   - On Redis error with FailOpen, task is processed anyway
//   - On Redis error with FailClosed, error is returned (task will be retried)
//
// Example:
//
//	handler := idempotency.IdempotentHandler(
//	    store,
//	    func(t *asynq.Task) string {
//	        var p Payload
//	        json.Unmarshal(t.Payload(), &p)
//	        return fmt.Sprintf("%s:%s", t.Type(), p.ID)
//	    },
//	    24*time.Hour,
//	    originalHandler,
//	    idempotency.WithHandlerLogger(logger),
//	)
func IdempotentHandler(
	store Store,
	keyExtractor KeyExtractor,
	ttl time.Duration,
	handler func(context.Context, *asynq.Task) error,
	opts ...HandlerOption,
) func(context.Context, *asynq.Task) error {
	// Validate required parameters
	if store == nil {
		panic("idempotency: store cannot be nil")
	}
	if keyExtractor == nil {
		panic("idempotency: keyExtractor cannot be nil")
	}
	if handler == nil {
		panic("idempotency: handler cannot be nil")
	}

	cfg := &handlerConfig{
		logger:   zap.NewNop(),
		failMode: FailOpen,
	}
	for _, opt := range opts {
		opt(cfg)
	}

	return func(ctx context.Context, t *asynq.Task) error {
		key := keyExtractor(t)
		if key == "" {
			// No idempotency key, process normally
			return handler(ctx, t)
		}

		isNew, err := store.Check(ctx, key, ttl)
		if err != nil {
			// Error during check - behavior depends on fail mode
			if cfg.failMode == FailOpen {
				cfg.logger.Warn("idempotency check failed, processing anyway",
					zap.String("idempotency_key", key),
					zap.String("task_type", t.Type()),
					zap.Error(err),
				)
				return handler(ctx, t)
			}
			cfg.logger.Error("idempotency check failed",
				zap.String("idempotency_key", key),
				zap.String("task_type", t.Type()),
				zap.Error(err),
			)
			return err // Return error to trigger retry
		}

		if !isNew {
			// Duplicate task - skip processing
			cfg.logger.Debug("duplicate task skipped",
				zap.String("idempotency_key", key),
				zap.String("task_type", t.Type()),
			)
			return nil // Skip without error
		}

		// New task - process normally
		return handler(ctx, t)
	}
}
