// Package runtimeutil provides runtime utility interfaces for external services.
package runtimeutil

import (
	"context"
	"time"
)

// Rate defines a rate limit configuration.
type Rate struct {
	// Limit is the number of requests allowed within the period.
	Limit int

	// Period is the time window for the rate limit.
	Period time.Duration
}

// NewRate creates a new Rate with the given limit and period.
func NewRate(limit int, period time.Duration) Rate {
	return Rate{Limit: limit, Period: period}
}

// RateLimiter defines rate limiting abstraction for swappable implementations.
// Compatible with middleware usage for HTTP rate limiting.
//
// Usage Example:
//
//	// In middleware
//	allowed, err := limiter.Allow(ctx, userIP)
//	if !allowed {
//	    http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
//	    return
//	}
//
//	// Setting custom rate for a user
//	limiter.Limit(ctx, userID, runtimeutil.NewRate(100, time.Minute))
//
// Implementing Redis RateLimiter:
//
//	type RedisRateLimiter struct {
//	    client *redis.Client
//	    defaultRate Rate
//	}
//
//	func (r *RedisRateLimiter) Allow(ctx context.Context, key string) (bool, error) {
//	    // Use Redis INCR with EXPIRE for sliding window
//	    count, _ := r.client.Incr(ctx, "rl:"+key).Result()
//	    if count == 1 {
//	        r.client.Expire(ctx, "rl:"+key, r.defaultRate.Period)
//	    }
//	    return count <= int64(r.defaultRate.Limit), nil
//	}
type RateLimiter interface {
	// Allow checks if the request should be allowed for the given key.
	// Returns true if allowed, false if rate limited.
	// The key is typically the user ID, IP address, or API key.
	Allow(ctx context.Context, key string) (bool, error)

	// Limit sets the rate limit for the given key.
	// Use this to configure per-user or per-IP limits dynamically.
	Limit(ctx context.Context, key string, rate Rate) error
}

// NopRateLimiter is a no-op rate limiter that always allows requests.
// Use for testing or when rate limiting is disabled.
type NopRateLimiter struct{}

// NewNopRateLimiter creates a new NopRateLimiter.
func NewNopRateLimiter() RateLimiter {
	return &NopRateLimiter{}
}

// Allow always returns true (request allowed).
func (r *NopRateLimiter) Allow(_ context.Context, _ string) (bool, error) {
	return true, nil
}

// Limit is a no-op and always returns nil.
func (r *NopRateLimiter) Limit(_ context.Context, _ string, _ Rate) error {
	return nil
}
