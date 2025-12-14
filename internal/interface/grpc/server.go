package grpc

import (
	"context"
	"fmt"
	"net"

	"github.com/iruldev/golang-api-hexagonal/internal/config"
	"github.com/iruldev/golang-api-hexagonal/internal/observability"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// Server wraps the underlying gRPC server with configuration and lifecycle management.
type Server struct {
	server   *grpc.Server
	cfg      *config.GRPCConfig
	logger   observability.Logger
	listener net.Listener
}

// ServerOption is a functional option for configuring the gRPC server.
type ServerOption func(*serverOptions)

type serverOptions struct {
	unaryInterceptors  []grpc.UnaryServerInterceptor
	streamInterceptors []grpc.StreamServerInterceptor
}

// WithUnaryInterceptors adds unary interceptors to the server.
func WithUnaryInterceptors(interceptors ...grpc.UnaryServerInterceptor) ServerOption {
	return func(o *serverOptions) {
		o.unaryInterceptors = append(o.unaryInterceptors, interceptors...)
	}
}

// WithStreamInterceptors adds stream interceptors to the server.
func WithStreamInterceptors(interceptors ...grpc.StreamServerInterceptor) ServerOption {
	return func(o *serverOptions) {
		o.streamInterceptors = append(o.streamInterceptors, interceptors...)
	}
}

// NewServer creates a new gRPC server with the given configuration.
func NewServer(cfg *config.GRPCConfig, logger observability.Logger, opts ...ServerOption) *Server {
	options := &serverOptions{}
	for _, opt := range opts {
		opt(options)
	}

	// Use StatsHandler for OpenTelemetry instrumentation (new API, interceptors are deprecated)
	serverOpts := []grpc.ServerOption{
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		grpc.ChainUnaryInterceptor(options.unaryInterceptors...),
		grpc.ChainStreamInterceptor(options.streamInterceptors...),
	}

	server := grpc.NewServer(serverOpts...)

	if cfg.ReflectionEnabled {
		reflection.Register(server)
		logger.Info("gRPC reflection enabled")
	}

	return &Server{
		server: server,
		cfg:    cfg,
		logger: logger,
	}
}

// GRPCServer returns the underlying grpc.Server for service registration.
func (s *Server) GRPCServer() *grpc.Server {
	return s.server
}

// Start starts the gRPC server on the configured port.
func (s *Server) Start(ctx context.Context) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.cfg.Port))
	if err != nil {
		return fmt.Errorf("grpc listen on port %d: %w", s.cfg.Port, err)
	}
	s.listener = lis

	s.logger.Info("gRPC server starting",
		observability.Int("port", s.cfg.Port),
		observability.Bool("reflection", s.cfg.ReflectionEnabled))

	// Serve blocks until Stop/GracefulStop is called
	if err := s.server.Serve(lis); err != nil {
		return fmt.Errorf("grpc serve: %w", err)
	}

	return nil
}

// Shutdown gracefully stops the gRPC server, draining existing connections.
func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("gRPC server shutting down gracefully")
	s.server.GracefulStop()
	return nil
}

// Stop immediately stops the gRPC server without waiting for connections to drain.
func (s *Server) Stop() {
	s.logger.Info("gRPC server stopping immediately")
	s.server.Stop()
}
