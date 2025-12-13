package middleware_test

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/iruldev/golang-api-hexagonal/internal/interface/http/middleware"
	"github.com/iruldev/golang-api-hexagonal/internal/runtimeutil"
)

// ExampleRateLimitMiddleware demonstrates basic rate limiting setup.
func ExampleRateLimitMiddleware() {
	// Create rate limiter: 100 requests per minute per IP
	limiter := middleware.NewInMemoryRateLimiter(
		middleware.WithDefaultRate(runtimeutil.NewRate(100, time.Minute)),
	)
	defer limiter.Stop()

	// Create router with rate limiting
	r := chi.NewRouter()
	r.Use(middleware.RateLimitMiddleware(limiter))

	r.Get("/api/data", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	fmt.Println("Rate limiter configured: 100 req/min")
	// Output: Rate limiter configured: 100 req/min
}

// ExampleRateLimitMiddleware_perEndpoint demonstrates per-endpoint rate limits.
func ExampleRateLimitMiddleware_perEndpoint() {
	// Create separate limiters for different endpoints
	strictLimiter := middleware.NewInMemoryRateLimiter(
		middleware.WithDefaultRate(runtimeutil.NewRate(10, time.Minute)), // 10/min for sensitive endpoints
	)
	defer strictLimiter.Stop()

	normalLimiter := middleware.NewInMemoryRateLimiter(
		middleware.WithDefaultRate(runtimeutil.NewRate(100, time.Minute)), // 100/min for normal endpoints
	)
	defer normalLimiter.Stop()

	r := chi.NewRouter()

	// Public endpoints with normal rate limit
	r.Group(func(r chi.Router) {
		r.Use(middleware.RateLimitMiddleware(normalLimiter))
		r.Get("/api/v1/notes", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("notes list"))
		})
	})

	// Sensitive endpoints with strict rate limit
	r.Group(func(r chi.Router) {
		r.Use(middleware.RateLimitMiddleware(strictLimiter))
		r.Post("/api/v1/auth/login", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("login"))
		})
	})

	fmt.Println("Per-endpoint rate limits configured")
	// Output: Per-endpoint rate limits configured
}

// ExampleRateLimitMiddleware_withUserID demonstrates user-based rate limiting.
func ExampleRateLimitMiddleware_withUserID() {
	limiter := middleware.NewInMemoryRateLimiter(
		middleware.WithDefaultRate(runtimeutil.NewRate(50, time.Minute)),
	)
	defer limiter.Stop()

	r := chi.NewRouter()

	// Public routes: rate limit by IP
	r.Group(func(r chi.Router) {
		r.Use(middleware.RateLimitMiddleware(limiter))
		r.Get("/api/v1/public", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("public data"))
		})
	})

	// Protected routes: rate limit by UserID (after auth)
	r.Group(func(r chi.Router) {
		// Auth middleware first, then rate limit by user
		r.Use(middleware.RateLimitMiddleware(limiter,
			middleware.WithKeyExtractor(middleware.UserIDKeyExtractor),
		))
		r.Get("/api/v1/me", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("user data"))
		})
	})

	fmt.Println("User-based rate limiting configured")
	// Output: User-based rate limiting configured
}

// ExampleRateLimitMiddleware_withAuthMiddleware demonstrates combining auth and rate limiting.
func ExampleRateLimitMiddleware_withAuthMiddleware() {
	// Create JWT authenticator (mock for example)
	jwtAuth := &rateLimitMockAuth{}

	// Create rate limiter
	limiter := middleware.NewInMemoryRateLimiter(
		middleware.WithDefaultRate(runtimeutil.NewRate(100, time.Minute)),
	)
	defer limiter.Stop()

	r := chi.NewRouter()

	// Protected route with auth + rate limiting
	r.Group(func(r chi.Router) {
		// 1. First authenticate
		r.Use(middleware.AuthMiddleware(jwtAuth))
		// 2. Then rate limit by user ID
		r.Use(middleware.RateLimitMiddleware(limiter,
			middleware.WithKeyExtractor(middleware.UserIDKeyExtractor),
		))

		r.Get("/api/v1/notes", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("protected notes"))
		})
	})

	fmt.Println("Auth + Rate limiting chain configured")
	// Output: Auth + Rate limiting chain configured
}

// ExampleRateLimitMiddleware_customKeyExtractor demonstrates a custom key extractor.
func ExampleRateLimitMiddleware_customKeyExtractor() {
	limiter := middleware.NewInMemoryRateLimiter(
		middleware.WithDefaultRate(runtimeutil.NewRate(100, time.Minute)),
	)
	defer limiter.Stop()

	// Custom extractor: rate limit by API key
	apiKeyExtractor := func(r *http.Request) string {
		apiKey := r.Header.Get("X-API-Key")
		if apiKey == "" {
			return middleware.IPKeyExtractor(r) // Fallback to IP
		}
		return "apikey:" + apiKey
	}

	r := chi.NewRouter()
	r.Use(middleware.RateLimitMiddleware(limiter,
		middleware.WithKeyExtractor(apiKeyExtractor),
	))

	r.Get("/api/v1/data", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("data"))
	})

	fmt.Println("Custom API key rate limiting configured")
	// Output: Custom API key rate limiting configured
}

// rateLimitMockAuth is a mock authenticator for rate limit examples.
type rateLimitMockAuth struct{}

func (m *rateLimitMockAuth) Authenticate(r *http.Request) (middleware.Claims, error) {
	return middleware.Claims{
		UserID: "test-user",
		Roles:  []string{"user"},
	}, nil
}
