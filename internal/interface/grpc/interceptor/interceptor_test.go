package interceptor

import (
	"context"
	"testing"

	"github.com/iruldev/golang-api-hexagonal/internal/observability"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type mockUnaryServerInfo struct {
	FullMethod string
}

func TestRecoveryInterceptor(t *testing.T) {
	logger := observability.NewNopLoggerInterface()

	tests := []struct {
		name        string
		handler     grpc.UnaryHandler
		wantErr     bool
		wantErrCode codes.Code
	}{
		{
			name: "no panic",
			handler: func(ctx context.Context, req interface{}) (interface{}, error) {
				return "success", nil
			},
			wantErr: false,
		},
		{
			name: "handler returns error",
			handler: func(ctx context.Context, req interface{}) (interface{}, error) {
				return nil, status.Error(codes.InvalidArgument, "bad request")
			},
			wantErr:     true,
			wantErrCode: codes.InvalidArgument,
		},
		{
			name: "panic recovered",
			handler: func(ctx context.Context, req interface{}) (interface{}, error) {
				panic("test panic")
			},
			wantErr:     true,
			wantErrCode: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			interceptor := RecoveryInterceptor(logger)
			require.NotNil(t, interceptor)

			info := &grpc.UnaryServerInfo{FullMethod: "/test.Service/Method"}

			resp, err := interceptor(context.Background(), nil, info, tt.handler)

			if tt.wantErr {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, tt.wantErrCode, st.Code())
			} else {
				require.NoError(t, err)
				assert.Equal(t, "success", resp)
			}
		})
	}
}

func TestLoggingInterceptor(t *testing.T) {
	logger := observability.NewNopLoggerInterface()

	tests := []struct {
		name    string
		handler grpc.UnaryHandler
		wantErr bool
	}{
		{
			name: "successful request",
			handler: func(ctx context.Context, req interface{}) (interface{}, error) {
				return "response", nil
			},
			wantErr: false,
		},
		{
			name: "failed request",
			handler: func(ctx context.Context, req interface{}) (interface{}, error) {
				return nil, status.Error(codes.NotFound, "not found")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			interceptor := LoggingInterceptor(logger)
			require.NotNil(t, interceptor)

			info := &grpc.UnaryServerInfo{FullMethod: "/test.Service/Method"}

			resp, err := interceptor(context.Background(), nil, info, tt.handler)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, "response", resp)
			}
		})
	}
}

func TestRequestIDInterceptor(t *testing.T) {
	tests := []struct {
		name          string
		wantRequestID bool
	}{
		{
			name:          "generates request ID when not present",
			wantRequestID: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			interceptor := RequestIDInterceptor()
			require.NotNil(t, interceptor)

			var capturedRequestID string
			handler := func(ctx context.Context, req interface{}) (interface{}, error) {
				capturedRequestID = requestIDFromContext(ctx)
				return "response", nil
			}

			info := &grpc.UnaryServerInfo{FullMethod: "/test.Service/Method"}

			resp, err := interceptor(context.Background(), nil, info, handler)

			require.NoError(t, err)
			assert.Equal(t, "response", resp)

			if tt.wantRequestID {
				assert.NotEmpty(t, capturedRequestID)
				// UUID format check
				assert.Len(t, capturedRequestID, 36)
			}
		})
	}
}

func TestMetricsInterceptor(t *testing.T) {
	tests := []struct {
		name    string
		handler grpc.UnaryHandler
		wantErr bool
	}{
		{
			name: "records metrics for successful request",
			handler: func(ctx context.Context, req interface{}) (interface{}, error) {
				return "response", nil
			},
			wantErr: false,
		},
		{
			name: "records metrics for failed request",
			handler: func(ctx context.Context, req interface{}) (interface{}, error) {
				return nil, status.Error(codes.Internal, "internal error")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			interceptor := MetricsInterceptor()
			require.NotNil(t, interceptor)

			info := &grpc.UnaryServerInfo{FullMethod: "/test.Service/Method"}

			resp, err := interceptor(context.Background(), nil, info, tt.handler)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, "response", resp)
			}
			// Note: Actual metric values would need to be verified via Prometheus registry
		})
	}
}

func TestRequestIDFromContext(t *testing.T) {
	t.Run("returns empty string when not set", func(t *testing.T) {
		id := requestIDFromContext(context.Background())
		assert.Empty(t, id)
	})

	t.Run("returns request ID when set", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), requestIDKey{}, "test-request-id")
		id := requestIDFromContext(ctx)
		assert.Equal(t, "test-request-id", id)
	})
}
