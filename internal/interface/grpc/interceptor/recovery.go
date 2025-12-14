// Package interceptor provides gRPC interceptors for logging, recovery, and metrics.
package interceptor

import (
	"context"
	"fmt"
	"runtime/debug"

	"github.com/iruldev/golang-api-hexagonal/internal/observability"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// RecoveryInterceptor recovers from panics in gRPC handlers and returns an INTERNAL error.
func RecoveryInterceptor(logger observability.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp interface{}, err error) {
		defer func() {
			if r := recover(); r != nil {
				stack := string(debug.Stack())
				logger.Error("gRPC panic recovered",
					observability.String("method", info.FullMethod),
					observability.String("panic", fmt.Sprintf("%v", r)),
					observability.String("stack", stack),
				)
				err = status.Errorf(codes.Internal, "internal server error")
			}
		}()

		return handler(ctx, req)
	}
}
