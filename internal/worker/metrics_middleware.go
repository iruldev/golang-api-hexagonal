package worker

import (
	"context"
	"time"

	"github.com/hibiken/asynq"
	"github.com/iruldev/golang-api-hexagonal/internal/observability"
)

// MetricsMiddleware records Prometheus metrics for task execution.
// Note: Queue name is always "default" because asynq does not expose
// the queue name in the handler context. This is an accepted limitation.
// Future: Consider extracting queue from task options if needed.
func MetricsMiddleware() asynq.MiddlewareFunc {
	return func(next asynq.Handler) asynq.Handler {
		return asynq.HandlerFunc(func(ctx context.Context, t *asynq.Task) error {
			start := time.Now()

			err := next.ProcessTask(ctx, t)

			duration := time.Since(start).Seconds()
			taskType := t.Type()
			// Queue name not available in asynq handler context - using constant for all jobs
			// See: https://github.com/hibiken/asynq/issues - no GetQueue in handler context
			queue := QueueDefault
			status := "success"
			if err != nil {
				status = "failed"
			}

			observability.JobProcessedTotal.WithLabelValues(taskType, queue, status).Inc()
			observability.JobDurationSeconds.WithLabelValues(taskType, queue).Observe(duration)

			return err
		})
	}
}
