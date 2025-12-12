// Package patterns provides reusable async job patterns for the worker infrastructure.
// These patterns abstract common async job use cases and provide a consistent API.
//
// Fire-and-Forget Pattern:
//
// The fire-and-forget pattern is used for non-critical background tasks where:
//   - The caller doesn't need to wait for completion
//   - Failures don't affect the caller's operation
//   - Work can be done "best effort" style
//
// Use cases:
//   - Analytics event tracking
//   - Cache warming
//   - Cleanup tasks
//   - Non-critical notifications
//   - Audit logging
//
// Example:
//
//	task, _ := tasks.NewNoteArchiveTask(noteID)
//	patterns.FireAndForget(ctx, client, logger, task)
//	// Returns immediately, caller continues
package patterns

import (
	"context"
	"time"

	"github.com/hibiken/asynq"
	"github.com/iruldev/golang-api-hexagonal/internal/worker"
	"github.com/iruldev/golang-api-hexagonal/internal/worker/tasks"
	"go.uber.org/zap"
)

// FireAndForget enqueues a task asynchronously without waiting for result.
// The caller is isolated from task execution - failures don't propagate.
//
// Default queue: low (non-urgent work shouldn't compete with critical tasks).
// Override via asynq.Queue option if more urgency is needed.
//
// The function returns immediately. Enqueueing happens in a goroutine to ensure
// the caller is never blocked. Any errors are logged but not returned.
//
// Parameters:
//   - ctx: Context for logging and tracing (not used for enqueueing cancellation)
//   - enqueuer: TaskEnqueuer interface for enqueueing tasks (typically *worker.Client)
//   - logger: Logger for recording task info and errors
//   - task: The asynq.Task to enqueue
//   - opts: Optional asynq options (Queue, MaxRetry, etc.) - overrides defaults
//
// Example:
//
//	// Default low queue
//	patterns.FireAndForget(ctx, client, logger, task)
//
//	// Override to default queue if more urgent
//	patterns.FireAndForget(ctx, client, logger, task, asynq.Queue(worker.QueueDefault))
func FireAndForget(
	ctx context.Context,
	enqueuer tasks.TaskEnqueuer,
	logger *zap.Logger,
	task *asynq.Task,
	opts ...asynq.Option,
) {
	// Default to low queue for non-critical work
	defaultOpts := []asynq.Option{asynq.Queue(worker.QueueLow)}
	allOpts := append(defaultOpts, opts...)

	// Enqueue in goroutine with timeout to prevent leaks
	go func() {
		// Use background context with timeout - parent context cancellation shouldn't affect enqueue
		enqueueCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		info, err := enqueuer.Enqueue(enqueueCtx, task, allOpts...)
		if err != nil {
			logger.Error("fire-and-forget enqueue failed",
				zap.Error(err),
				zap.String("task_type", task.Type()),
			)
			return
		}
		logger.Debug("fire-and-forget task enqueued",
			zap.String("task_id", info.ID),
			zap.String("task_type", task.Type()),
			zap.String("queue", info.Queue),
		)
	}()
}
