package interceptor

import (
	"context"
	"time"

	"github.com/iruldev/golang-api-hexagonal/internal/observability"

	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

// LoggingInterceptor logs gRPC requests with structured fields.
func LoggingInterceptor(logger observability.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		start := time.Now()

		resp, err := handler(ctx, req)

		code := status.Code(err)
		duration := time.Since(start)

		// Get request ID from context if available
		requestID := requestIDFromContext(ctx)

		fields := []observability.Field{
			observability.String("method", info.FullMethod),
			observability.String("status", code.String()),
			observability.Duration("duration", duration),
		}

		if requestID != "" {
			fields = append(fields, observability.String("request_id", requestID))
		}

		if err != nil {
			fields = append(fields, observability.Err(err))
			logger.Warn("gRPC request failed", fields...)
		} else {
			logger.Info("gRPC request", fields...)
		}

		return resp, err
	}
}
