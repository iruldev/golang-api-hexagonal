// Package http provides HTTP transport layer components.
package http

import (
	"log/slog"
	stdhttp "net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/iruldev/golang-api-hexagonal/internal/shared/metrics"
	"github.com/iruldev/golang-api-hexagonal/internal/transport/http/middleware"
)

// UserRoutes defines the interface for user-related HTTP handlers.
// This interface breaks the import cycle between http and handler packages.
type UserRoutes interface {
	CreateUser(w stdhttp.ResponseWriter, r *stdhttp.Request)
	GetUser(w stdhttp.ResponseWriter, r *stdhttp.Request)
	ListUsers(w stdhttp.ResponseWriter, r *stdhttp.Request)
}

// JWTConfig holds JWT authentication configuration for the router.
type JWTConfig struct {
	// Enabled controls whether JWT authentication is applied to protected routes.
	// When false, protected routes are accessible without authentication.
	Enabled bool
	// Secret is the key used for JWT signature validation (HS256).
	// Required when Enabled is true.
	Secret []byte
	// Now provides the current time for token validation.
	// Inject for deterministic testing; defaults to time.Now if nil.
	Now func() time.Time
	// Issuer is the expected issuer claim. If set, tokens must have matching iss claim.
	Issuer string
	// Audience is the expected audience claim. If set, tokens must have matching aud claim.
	Audience string
	// ClockSkew is the tolerance for expired tokens (e.g., 30s).
	// Default: 0 (no tolerance).
	ClockSkew time.Duration
}

// RateLimitConfig holds rate limiting configuration for the router.
type RateLimitConfig struct {
	// RequestsPerSecond is the number of requests allowed per second.
	// Default: 100.
	RequestsPerSecond int
	// TrustProxy enables trusting X-Forwarded-For/X-Real-IP headers for client IP.
	// Default: false.
	TrustProxy bool
}

// BasePath is the versioned base path for the API.
const BasePath = "/api/v1"

// NewRouter creates a new chi router with the provided handlers and logger.
//
// Middleware ordering:
//  1. RequestID: Generate/passthrough request ID FIRST
//  2. Tracing: Create spans and propagate trace context
//  3. Metrics: Record request counts and durations
//  4. Logging: Needs requestId and traceId in context
//  5. RealIP: Extract real IP from headers
//  6. BodyLimiter: Enforce request body size limits
//  7. Recoverer: Panic recovery
//
// JWT middleware is applied per-route group (protected endpoints only).
func NewRouter(
	logger *slog.Logger,
	tracingEnabled bool,
	metricsReg *prometheus.Registry,
	httpMetrics metrics.HTTPMetrics,
	healthHandler, readyHandler stdhttp.Handler,
	userHandler UserRoutes,
	maxRequestSize int64,
	jwtConfig JWTConfig,
	rateLimitConfig RateLimitConfig,
) chi.Router {
	r := chi.NewRouter()

	// Global middleware stack (order matters!)
	// SecureHeaders FIRST - ensures security headers on ALL responses including errors
	r.Use(middleware.SecureHeaders)
	r.Use(middleware.RequestID)

	// Story 2.6: Only trust proxy headers (X-Forwarded-For, X-Real-IP) when explicitly configured.
	// This prevents IP spoofing when not behind a trusted proxy.
	// MUST be before Logger and Metrics so they see the real IP.
	if rateLimitConfig.TrustProxy {
		r.Use(chiMiddleware.RealIP)
	}

	if tracingEnabled {
		r.Use(middleware.Tracing)
	}
	r.Use(middleware.Metrics(httpMetrics))
	r.Use(middleware.RequestLogger(logger))
	r.Use(middleware.BodyLimiter(maxRequestSize))
	r.Use(chiMiddleware.Recoverer)

	// NOTE: /metrics endpoint moved to internal router (Story 2.5b)

	// Health check endpoints (no auth required)
	r.Get("/health", healthHandler.ServeHTTP)
	r.Get("/ready", readyHandler.ServeHTTP)

	// API v1 routes (protected when JWT is enabled)
	if userHandler != nil {
		r.Route(BasePath, func(r chi.Router) {
			// Apply JWT auth middleware if enabled
			if jwtConfig.Enabled {
				now := jwtConfig.Now
				if now == nil {
					now = time.Now
				}
				r.Use(middleware.JWTAuth(middleware.JWTAuthConfig{
					Secret:    jwtConfig.Secret,
					Logger:    logger,
					Now:       now,
					Issuer:    jwtConfig.Issuer,
					Audience:  jwtConfig.Audience,
					ClockSkew: jwtConfig.ClockSkew,
				}))
				r.Use(middleware.AuthContextBridge)
			}

			// Apply rate limiting after JWT auth so claims are available for per-user limiting
			r.Use(middleware.RateLimiter(middleware.RateLimitConfig{
				RequestsPerSecond: rateLimitConfig.RequestsPerSecond,
				TrustProxy:        rateLimitConfig.TrustProxy,
			}))

			r.Post("/users", userHandler.CreateUser)
			r.Get("/users/{id}", userHandler.GetUser)
			r.Get("/users", userHandler.ListUsers)
		})
	}

	return r
}

// NewInternalRouter creates a router for internal endpoints like /metrics.
// This router should be bound to INTERNAL_PORT for security isolation.
// Story 2.5b: Separated from public router to protect internal metrics.
func NewInternalRouter(
	logger *slog.Logger,
	metricsReg *prometheus.Registry,
	httpMetrics metrics.HTTPMetrics,
) *chi.Mux {
	r := chi.NewRouter()

	// Recovery middleware to prevent panics from crashing internal server
	r.Use(chiMiddleware.Recoverer)
	// Apply metrics middleware to track requests to internal endpoints
	r.Use(middleware.Metrics(httpMetrics))
	// Internal router logs are useful for debugging scraping issues
	r.Use(middleware.RequestLogger(logger))

	// Metrics endpoint (internal only)
	r.Handle("/metrics", promhttp.HandlerFor(metricsReg, promhttp.HandlerOpts{}))

	logger.Debug("internal router configured", slog.String("endpoints", "/metrics"))

	return r
}
