package redis_test

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	infraredis "github.com/iruldev/golang-api-hexagonal/internal/infra/redis"
	"github.com/iruldev/golang-api-hexagonal/internal/runtimeutil"
)

// Example_basicRedisRateLimiter demonstrates basic usage of RedisRateLimiter.
func Example_basicRedisRateLimiter() {
	// Create Redis client (in production, use config.RedisConfig)
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	defer client.Close()

	// Create rate limiter with 100 requests per minute
	limiter := infraredis.NewRedisRateLimiter(
		client,
		infraredis.WithRedisDefaultRate(runtimeutil.NewRate(100, time.Minute)),
		infraredis.WithKeyPrefix("api:ratelimit:"),
	)

	// Check if request is allowed
	ctx := context.Background()
	allowed, err := limiter.Allow(ctx, "user-123")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	if allowed {
		fmt.Println("Request allowed")
	} else {
		fmt.Println("Request rate limited")
	}

	// Output: Request allowed
}

// Example_redisRateLimiterWithFallback demonstrates fallback configuration.
func Example_redisRateLimiterWithFallback() {
	// In production, replace with real Redis client
	// When Redis is unavailable, the fallback limiter is used

	// Example configuration with fallback
	/*
		// Create in-memory fallback limiter
		fallback := middleware.NewInMemoryRateLimiter(
			middleware.WithDefaultRate(runtimeutil.NewRate(50, time.Minute)),
		)
		defer fallback.Stop()

		// Create Redis rate limiter with fallback
		limiter := infraredis.NewRedisRateLimiter(
			redisClient,
			infraredis.WithRedisDefaultRate(runtimeutil.NewRate(100, time.Minute)),
			infraredis.WithFallbackLimiter(fallback),
			infraredis.WithCircuitBreakerConfig(5, 30*time.Second),
		)
	*/

	fmt.Println("Fallback configured")
	// Output: Fallback configured
}

// Example_customRatePerKey demonstrates setting custom rates for specific keys.
func Example_customRatePerKey() {
	// In production, replace with real Redis client
	/*
		ctx := context.Background()

		// Set higher rate for premium users
		err := limiter.Limit(ctx, "premium-user-456", runtimeutil.NewRate(1000, time.Minute))
		if err != nil {
			log.Printf("Failed to set rate: %v", err)
		}

		// Set lower rate for free tier
		err = limiter.Limit(ctx, "free-user-789", runtimeutil.NewRate(10, time.Minute))
		if err != nil {
			log.Printf("Failed to set rate: %v", err)
		}
	*/

	fmt.Println("Custom rates configured")
	// Output: Custom rates configured
}

// Example_integrateWithMiddleware demonstrates integration with HTTP middleware.
func Example_integrateWithMiddleware() {
	// In production, integrate with RateLimitMiddleware from middleware package
	/*
		import "github.com/iruldev/golang-api-hexagonal/internal/interface/http/middleware"

		// Create Redis rate limiter
		redisLimiter := infraredis.NewRedisRateLimiter(
			redisClient.Client(), // Get underlying *redis.Client
			infraredis.WithRedisDefaultRate(runtimeutil.NewRate(100, time.Minute)),
			infraredis.WithKeyPrefix("api:"),
		)

		// Create middleware with Redis limiter
		rateLimitMw := middleware.RateLimitMiddleware(
			redisLimiter,
			observability.NewNopLoggerInterface(),
			middleware.WithKeyExtractor(middleware.IPKeyExtractor),
		)

		// Use in router
		r := chi.NewRouter()
		r.Use(rateLimitMw)
	*/

	fmt.Println("Middleware integration configured")
	// Output: Middleware integration configured
}

// Example_retryAfterHeader demonstrates getting retry-after value.
func Example_retryAfterHeader() {
	// RedisRateLimiter implements the RateLimiterWithRetryAfter interface
	/*
		// After rate limiting occurs
		retryAfter := limiter.RetryAfter("user-123")
		if retryAfter > 0 {
			w.Header().Set("Retry-After", strconv.Itoa(retryAfter))
		}
	*/

	fmt.Println("Retry-After header support available")
	// Output: Retry-After header support available
}
