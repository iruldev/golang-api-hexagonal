// Package main is the entry point for the backend service.
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/iruldev/golang-api-hexagonal/internal/app"
	"github.com/iruldev/golang-api-hexagonal/internal/config"
	"github.com/iruldev/golang-api-hexagonal/internal/infra/postgres"
	"github.com/iruldev/golang-api-hexagonal/internal/infra/redis"
	"github.com/iruldev/golang-api-hexagonal/internal/interface/graphql"
	grpcserver "github.com/iruldev/golang-api-hexagonal/internal/interface/grpc"
	"github.com/iruldev/golang-api-hexagonal/internal/interface/grpc/interceptor"
	grpcnote "github.com/iruldev/golang-api-hexagonal/internal/interface/grpc/note"
	httpx "github.com/iruldev/golang-api-hexagonal/internal/interface/http"
	"github.com/iruldev/golang-api-hexagonal/internal/observability"
	noteuc "github.com/iruldev/golang-api-hexagonal/internal/usecase/note"
	notev1 "github.com/iruldev/golang-api-hexagonal/proto/note/v1"
)

func main() {
	// Load and validate configuration (Epic 2: Configuration & Environment)
	cfg, err := config.Load()
	if err != nil {
		// Exit code 1 with clear error message (Story 2.5)
		log.Fatalf("Configuration error: %v", err)
	}

	// Initialize logger (Story 5.6 / 5.7)
	zapLogger, err := observability.NewLogger(&cfg.Log, cfg.App.Env)
	if err != nil {
		log.Fatalf("Logger initialization error: %v", err)
	}
	logger := observability.NewZapLogger(zapLogger)
	defer logger.Sync()

	// Initialize database connection pool (Story 4.1)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	pool, err := postgres.NewPool(ctx, cfg)
	cancel()

	var dbChecker *postgres.PoolHealthChecker
	if err != nil {
		// Log warning but don't fail - database may be optional for some routes
		logger.Warn("Database connection failed", observability.Err(err))
	} else {
		defer pool.Close()
		logger.Info("Database connected",
			observability.String("user", cfg.Database.User),
			observability.String("host", cfg.Database.Host),
			observability.Int("port", cfg.Database.Port),
			observability.String("name", cfg.Database.Name),
			observability.Int("max_conns", cfg.Database.MaxOpenConns))
		// Create DB health checker for readiness probe (Story 4.7)
		dbChecker = postgres.NewPoolHealthChecker(pool)
	}

	// Initialize Redis connection pool (Story 8.1)
	var redisChecker *redis.Client
	redisClient, err := redis.NewClient(cfg.Redis)
	if err != nil {
		// Log warning but don't fail - Redis may be optional for some routes
		logger.Warn("Redis connection failed", observability.Err(err))
	} else {
		defer redisClient.Close()
		logger.Info("Redis connected",
			observability.String("host", cfg.Redis.Host),
			observability.Int("port", cfg.Redis.Port),
			observability.Int("pool_size", cfg.Redis.PoolSize))
		// Use Redis client as health checker for readiness probe
		redisChecker = redisClient
	}

	// Use typed config instead of raw os.Getenv
	port := fmt.Sprintf("%d", cfg.App.HTTPPort)

	// Create chi router with versioned API routes (Story 3.1)
	router := httpx.NewRouter(httpx.RouterDeps{
		Config:       cfg,
		DBChecker:    dbChecker,
		RedisChecker: redisChecker,
	})

	server := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	// Start HTTP server in goroutine
	go func() {
		logger.Info("HTTP server starting", observability.String("port", port))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("HTTP server error", observability.Err(err))
			os.Exit(1)
		}
	}()

	// Initialize Note module (Common for REST, gRPC, GraphQL)
	var noteUsecase *noteuc.Usecase
	if pool != nil {
		noteRepo := postgres.NewNoteRepository(pool)
		noteUsecase = noteuc.NewUsecase(noteRepo)
	}

	// Initialize GraphQL handler (Story 12.3)
	// We register this on the main router to leverage existing middleware (logging, auth, etc)
	if noteUsecase != nil {
		srv := handler.NewDefaultServer(graphql.NewExecutableSchema(graphql.Config{
			Resolvers: &graphql.Resolver{
				NoteUsecase: noteUsecase,
			},
		}))
		router.Handle("/query", srv)
		logger.Info("GraphQL handler registered at /query")

		// Initialize GraphQL Playground (Story 12.4)
		// CRITICAL: Only enable in development/local for security - playground exposes schema introspection
		if cfg.App.IsDevelopment() {
			router.Handle("/playground", playground.Handler("GraphQL playground", "/query"))
			logger.Info("GraphQL Playground enabled at /playground (dev mode)")
		} else {
			logger.Info("GraphQL Playground disabled (non-dev mode)", observability.String("env", cfg.App.Env))
		}
	} else {
		logger.Warn("GraphQL handler not registered - database unavailable")
	}

	// Start gRPC server if enabled (Story 12.1)
	var grpcSrv *grpcserver.Server
	if cfg.GRPC.Enabled {
		grpcSrv = grpcserver.NewServer(
			&cfg.GRPC,
			logger,
			grpcserver.WithUnaryInterceptors(
				interceptor.RequestIDInterceptor(),      // First: propagate request ID to all downstream
				interceptor.MetricsInterceptor(),        // Second: capture metrics including panic errors
				interceptor.LoggingInterceptor(logger),  // Third: log request including panic errors
				interceptor.RecoveryInterceptor(logger), // Last: catch panics and return INTERNAL
			),
		)

		// Register gRPC services (Story 12.2)
		if noteUsecase != nil {
			noteHandler := grpcnote.NewHandler(noteUsecase)
			notev1.RegisterNoteServiceServer(grpcSrv.GRPCServer(), noteHandler)
			logger.Info("gRPC NoteService registered")
		} else {
			logger.Warn("gRPC NoteService not registered - database unavailable")
		}

		go func() {
			if err := grpcSrv.Start(context.Background()); err != nil {
				logger.Error("gRPC server error", observability.Err(err))
				os.Exit(1)
			}
		}()
	}

	// Wait for shutdown signal (Story 1.4 graceful shutdown)
	done := make(chan error, 1)
	go app.GracefulShutdown(server, done)

	if err := <-done; err != nil {
		logger.Error("HTTP shutdown error", observability.Err(err))
	}

	// Shutdown gRPC server if running (Story 12.1 - respects global shutdown timeout)
	if grpcSrv != nil {
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
		if err := grpcSrv.Shutdown(shutdownCtx); err != nil {
			logger.Error("gRPC shutdown error", observability.Err(err))
		}
		shutdownCancel()
	}

	// Shutdown tracer provider to flush remaining spans (Story 3.5)
	if httpx.TracerShutdown != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := httpx.TracerShutdown(ctx); err != nil {
			logger.Error("Tracer shutdown error", observability.Err(err))
		}
	}

	logger.Info("Server shutdown complete")
	os.Exit(0)
}
