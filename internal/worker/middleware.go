package worker

import (
	"context"
	"fmt"
	"time"

	"github.com/hibiken/asynq"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// LoggingMiddleware logs task execution with zap structured logger.
func LoggingMiddleware(logger *zap.Logger) asynq.MiddlewareFunc {
	return func(next asynq.Handler) asynq.Handler {
		return asynq.HandlerFunc(func(ctx context.Context, t *asynq.Task) error {
			start := time.Now()
			taskID, _ := asynq.GetTaskID(ctx)
			retryCount, _ := asynq.GetRetryCount(ctx)

			err := next.ProcessTask(ctx, t)

			duration := time.Since(start)
			fields := []zap.Field{
				zap.String("task_type", t.Type()),
				zap.String("task_id", taskID),
				zap.Int("retry_count", retryCount),
				zap.Duration("duration", duration),
				zap.Bool("success", err == nil),
			}

			if err != nil {
				fields = append(fields, zap.Error(err))
				logger.Error("task failed", fields...)
			} else {
				logger.Info("task processed", fields...)
			}

			return err
		})
	}
}

// RecoveryMiddleware recovers from panics in task handlers.
func RecoveryMiddleware(logger *zap.Logger) asynq.MiddlewareFunc {
	return func(next asynq.Handler) asynq.Handler {
		return asynq.HandlerFunc(func(ctx context.Context, t *asynq.Task) (err error) {
			defer func() {
				if r := recover(); r != nil {
					taskID, _ := asynq.GetTaskID(ctx)
					logger.Error("panic in task",
						zap.String("task_type", t.Type()),
						zap.String("task_id", taskID),
						zap.Any("panic", r),
					)
					err = fmt.Errorf("panic recovered: %v", r)
				}
			}()
			return next.ProcessTask(ctx, t)
		})
	}
}

// TracingMiddleware creates OpenTelemetry spans for task execution.
func TracingMiddleware() asynq.MiddlewareFunc {
	tracer := otel.Tracer("asynq")
	return func(next asynq.Handler) asynq.Handler {
		return asynq.HandlerFunc(func(ctx context.Context, t *asynq.Task) error {
			taskID, _ := asynq.GetTaskID(ctx)

			ctx, span := tracer.Start(ctx, "task:"+t.Type(),
				trace.WithSpanKind(trace.SpanKindConsumer),
			)
			span.SetAttributes(
				attribute.String("task.id", taskID),
				attribute.String("task.type", t.Type()),
			)
			defer span.End()

			err := next.ProcessTask(ctx, t)
			if err != nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, err.Error())
			} else {
				span.SetStatus(codes.Ok, "")
			}

			return err
		})
	}
}
