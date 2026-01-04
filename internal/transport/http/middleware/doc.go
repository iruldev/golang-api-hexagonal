// Package middleware provides HTTP middleware for the transport layer.
//
// This package contains reusable middleware components for the Chi router
// that implement cross-cutting concerns like authentication, rate limiting,
// idempotency, logging, and graceful shutdown.
//
// # Middleware Ordering
//
// Middleware should be applied in this specific order (outermost to innermost execution):
//
//  1. RequestID      - Assigns unique request ID for tracing
//  2. Logger         - Logs request/response with timing
//  3. Recoverer      - Catches panics and returns 500 response
//  4. RateLimit      - Enforces rate limits per IP/endpoint
//  5. Auth           - Validates JWT tokens
//  6. Idempotency    - Handles duplicate POST requests (POST only)
//
// # Chi Router Integration
//
// Apply middleware using Chi's Use method:
//
//	r := chi.NewRouter()
//	r.Use(middleware.RequestID)
//	r.Use(middleware.Logger(logger))
//	r.Use(middleware.Recoverer(logger))
//	r.Use(middleware.RateLimit(limiter))
//
//	r.Route("/api/v1", func(r chi.Router) {
//	    r.Use(middleware.JWTAuth(authConfig))
//	    r.Post("/", middleware.Idempotency(store)(handler))
//	})
//
// # Available Middleware
//
// Authentication:
//   - JWTAuth: Validates JWT tokens from Authorization header (HS256 only)
//   - AuthContextBridge: Bridges auth context between middleware and handlers
//
// Rate Limiting:
//   - RateLimit: Token bucket rate limiting with RFC 7807 responses
//   - Includes X-RateLimit-* headers and Retry-After on 429
//
// Idempotency:
//   - Idempotency: Caches POST responses by Idempotency-Key header
//   - Prevents duplicate side effects on retry
//
// Observability:
//   - RequestID: Generates unique request IDs (X-Request-ID header)
//   - Logger: Structured logging with request/response timing
//   - Metrics: Prometheus metrics for HTTP requests
//   - Tracing: OpenTelemetry distributed tracing spans
//
// Resilience:
//   - Shutdown: Graceful shutdown with request draining
//   - Recoverer: Panic recovery with RFC 7807 error response
//   - ResilienceWrapper: Composed circuit breaker, retry, timeout, and bulkhead patterns
//
// Security:
//   - Security: OWASP-recommended security headers
//   - BodyLimiter: Request body size limits
//
// # Error Responses
//
// All middleware use RFC 7807 Problem Details format for error responses
// via the contract package. Example 401 response:
//
//	{
//	    "type": "https://api.example.com/problems/unauthorized",
//	    "title": "Unauthorized",
//	    "status": 401,
//	    "code": "ERR_AUTH_UNAUTHORIZED",
//	    "request_id": "req_abc123"
//	}
//
// # Configuration
//
// Most middleware accept configuration structs. Example:
//
//	cfg := middleware.JWTAuthConfig{
//	    Secret:    []byte("secret-key"),
//	    Logger:    slog.Default(),
//	    Issuer:    "my-app",
//	    Audience:  "my-api",
//	    ClockSkew: 5 * time.Second,
//	}
//	r.Use(middleware.JWTAuth(cfg))
//
// # See Also
//
//   - Chi router documentation: https://github.com/go-chi/chi
//   - ADR-002: Layer Boundary Enforcement
//   - contract package: RFC 7807 error responses
package middleware
