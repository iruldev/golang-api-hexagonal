// Package redis provides Redis-based infrastructure implementations.
package redis

import (
	"context"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/iruldev/golang-api-hexagonal/internal/runtimeutil"
)

// luaRateLimitScript is a Lua script for atomic sliding window rate limiting.
// KEYS[1] = rate limit key
// ARGV[1] = limit (max requests)
// ARGV[2] = window (seconds)
// Returns: 1 if allowed, 0 if rate limited
const luaRateLimitScript = `
local current = redis.call('INCR', KEYS[1])
if current == 1 then
    redis.call('EXPIRE', KEYS[1], ARGV[2])
end
if current > tonumber(ARGV[1]) then
    return 0
end
return 1
`

// circuitBreaker implements a simple circuit breaker pattern for Redis failures.
type circuitBreaker struct {
	failures     int
	threshold    int
	lastFailure  time.Time
	recoveryTime time.Duration
	mu           sync.Mutex
}

// newCircuitBreaker creates a new circuit breaker with the given threshold and recovery time.
func newCircuitBreaker(threshold int, recoveryTime time.Duration) *circuitBreaker {
	return &circuitBreaker{
		threshold:    threshold,
		recoveryTime: recoveryTime,
	}
}

// isOpen returns true if the circuit breaker is open (failures exceeded threshold).
// Returns false if recovery time has passed, allowing a retry.
func (cb *circuitBreaker) isOpen() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if cb.failures >= cb.threshold {
		// Check if recovery time has passed
		if time.Since(cb.lastFailure) > cb.recoveryTime {
			cb.failures = 0 // Reset, attempt recovery
			return false
		}
		return true // Still open
	}
	return false // Closed, use Redis
}

// recordFailure records a failure and updates the last failure time.
func (cb *circuitBreaker) recordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.failures++
	cb.lastFailure = time.Now()
}

// recordSuccess resets the failure count on successful operations.
func (cb *circuitBreaker) recordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.failures = 0
}

// RedisRateLimiter implements runtimeutil.RateLimiter using Redis for distributed rate limiting.
// It provides atomic rate limiting across multiple application instances using Lua scripts.
//
// Features:
//   - Sliding window counter algorithm using Redis INCR + EXPIRE
//   - Atomic operations via Lua script (no race conditions)
//   - Fallback to in-memory rate limiter on Redis failure
//   - Circuit breaker to prevent repeated Redis failures
//   - Configurable timeout for Redis operations
//
// Usage:
//
//	limiter := NewRedisRateLimiter(
//	    redisClient.Client(),
//	    WithRedisDefaultRate(runtimeutil.NewRate(100, time.Minute)),
//	    WithKeyPrefix("api:"),
//	)
//
//	allowed, _ := limiter.Allow(ctx, "user-123")
type RedisRateLimiter struct {
	client      *redis.Client
	defaultRate runtimeutil.Rate
	keyPrefix   string
	timeout     time.Duration

	// Fallback limiter for when Redis is unavailable
	fallback runtimeutil.RateLimiter
	circuit  *circuitBreaker

	// Script hash caching
	scriptSHA string
	scriptMu  sync.Mutex

	// Per-key rate configuration
	keyRates sync.Map // map[string]runtimeutil.Rate
}

// RedisOption configures the RedisRateLimiter.
type RedisOption func(*RedisRateLimiter)

// WithRedisDefaultRate sets the default rate limit for new keys.
func WithRedisDefaultRate(rate runtimeutil.Rate) RedisOption {
	return func(r *RedisRateLimiter) {
		r.defaultRate = rate
	}
}

// WithKeyPrefix sets the Redis key prefix for rate limit keys.
func WithKeyPrefix(prefix string) RedisOption {
	return func(r *RedisRateLimiter) {
		r.keyPrefix = prefix
	}
}

// WithRedisTimeout sets the timeout for Redis operations.
func WithRedisTimeout(timeout time.Duration) RedisOption {
	return func(r *RedisRateLimiter) {
		r.timeout = timeout
	}
}

// WithFallbackLimiter sets the fallback rate limiter for when Redis is unavailable.
func WithFallbackLimiter(fallback runtimeutil.RateLimiter) RedisOption {
	return func(r *RedisRateLimiter) {
		r.fallback = fallback
	}
}

// WithCircuitBreakerConfig configures the circuit breaker threshold and recovery time.
func WithCircuitBreakerConfig(threshold int, recoveryTime time.Duration) RedisOption {
	return func(r *RedisRateLimiter) {
		r.circuit = newCircuitBreaker(threshold, recoveryTime)
	}
}

// NewRedisRateLimiter creates a new Redis-backed rate limiter.
// The client parameter is the underlying Redis client from go-redis.
func NewRedisRateLimiter(client *redis.Client, opts ...RedisOption) *RedisRateLimiter {
	r := &RedisRateLimiter{
		client:      client,
		defaultRate: runtimeutil.NewRate(100, time.Minute), // 100 req/min default
		keyPrefix:   "rl:",
		timeout:     100 * time.Millisecond,
		circuit:     newCircuitBreaker(5, 30*time.Second),
	}

	for _, opt := range opts {
		opt(r)
	}

	return r
}

// Allow checks if a request for the given key should be allowed.
// Returns true if allowed, false if rate limited.
// Falls back to in-memory limiter on Redis failure.
// Implements runtimeutil.RateLimiter.
func (r *RedisRateLimiter) Allow(ctx context.Context, key string) (bool, error) {
	// Check circuit breaker state
	if r.circuit.isOpen() {
		if r.fallback != nil {
			return r.fallback.Allow(ctx, key)
		}
		// Fail-open if no fallback configured
		return true, nil
	}

	// Set timeout
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	// Get rate for this key (default if not set)
	rate := r.getRateForKey(key)

	// Ensure script is loaded
	sha, err := r.ensureScript(ctx)
	if err != nil {
		r.circuit.recordFailure()
		if r.fallback != nil {
			return r.fallback.Allow(ctx, key)
		}
		// Fail-open if no fallback configured
		return true, nil
	}

	// Execute script
	result, err := r.client.EvalSha(ctx, sha,
		[]string{r.keyPrefix + key},
		rate.Limit,
		int(rate.Period.Seconds()),
	).Int()

	if err != nil {
		// Check if script was flushed (NOSCRIPT error)
		if isNoScriptError(err) {
			// Reset script hash and retry with EVAL
			r.scriptMu.Lock()
			r.scriptSHA = ""
			r.scriptMu.Unlock()

			result, err = r.client.Eval(ctx, luaRateLimitScript,
				[]string{r.keyPrefix + key},
				rate.Limit,
				int(rate.Period.Seconds()),
			).Int()
		}

		if err != nil {
			r.circuit.recordFailure()
			if r.fallback != nil {
				return r.fallback.Allow(ctx, key)
			}
			// Fail-open if no fallback configured
			return true, nil
		}
	}

	r.circuit.recordSuccess()
	return result == 1, nil
}

// Limit sets a custom rate limit for the given key.
// Implements runtimeutil.RateLimiter.
func (r *RedisRateLimiter) Limit(_ context.Context, key string, rate runtimeutil.Rate) error {
	r.keyRates.Store(key, rate)
	return nil
}

// RetryAfter returns the number of seconds until the next request for key is allowed.
// Checks the TTL of the rate limit key in Redis.
func (r *RedisRateLimiter) RetryAfter(key string) int {
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()

	// Check circuit breaker - if open, delegate to fallback
	if r.circuit.isOpen() {
		if retryLimiter, ok := r.fallback.(interface{ RetryAfter(string) int }); ok {
			return retryLimiter.RetryAfter(key)
		}
		return 60 // Default fallback
	}

	ttl, err := r.client.TTL(ctx, r.keyPrefix+key).Result()
	if err != nil || ttl <= 0 {
		return 0 // Key doesn't exist or no TTL
	}

	return int(ttl.Seconds())
}

// getRateForKey returns the rate limit configuration for the given key.
// Returns the default rate if no custom rate is set.
func (r *RedisRateLimiter) getRateForKey(key string) runtimeutil.Rate {
	if rate, ok := r.keyRates.Load(key); ok {
		return rate.(runtimeutil.Rate)
	}
	return r.defaultRate
}

// ensureScript loads the Lua script into Redis and returns the SHA1 hash.
// Uses SCRIPT LOAD to cache the script for EVALSHA efficiency.
func (r *RedisRateLimiter) ensureScript(ctx context.Context) (string, error) {
	r.scriptMu.Lock()
	defer r.scriptMu.Unlock()

	if r.scriptSHA != "" {
		return r.scriptSHA, nil
	}

	sha, err := r.client.ScriptLoad(ctx, luaRateLimitScript).Result()
	if err != nil {
		return "", err
	}

	r.scriptSHA = sha
	return sha, nil
}

// isNoScriptError checks if the error is a NOSCRIPT error (script not in cache).
func isNoScriptError(err error) bool {
	if err == nil {
		return false
	}
	return err.Error() == "NOSCRIPT No matching script. Please use EVAL." ||
		(len(err.Error()) >= 8 && err.Error()[:8] == "NOSCRIPT")
}
