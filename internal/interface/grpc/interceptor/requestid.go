package interceptor

import (
	"context"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type requestIDKey struct{}

const requestIDMetadataKey = "x-request-id"

// RequestIDInterceptor ensures each request has a unique request ID.
// It propagates existing request IDs from metadata or generates a new one.
func RequestIDInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		var requestID string

		// Try to get existing request ID from metadata
		if md, ok := metadata.FromIncomingContext(ctx); ok {
			if ids := md.Get(requestIDMetadataKey); len(ids) > 0 {
				requestID = ids[0]
			}
		}

		// Generate new request ID if not present
		if requestID == "" {
			requestID = uuid.New().String()
		}

		// Add request ID to context
		ctx = context.WithValue(ctx, requestIDKey{}, requestID)

		// Add request ID to outgoing metadata for downstream services
		ctx = metadata.AppendToOutgoingContext(ctx, requestIDMetadataKey, requestID)

		return handler(ctx, req)
	}
}

// requestIDFromContext extracts the request ID from context.
func requestIDFromContext(ctx context.Context) string {
	if id, ok := ctx.Value(requestIDKey{}).(string); ok {
		return id
	}
	return ""
}
