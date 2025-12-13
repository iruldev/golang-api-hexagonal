// Package middleware provides HTTP middleware for rate limiting, authentication, and authorization.
package middleware

import (
	"context"
	"math"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/iruldev/golang-api-hexagonal/internal/interface/http/response"
	"github.com/iruldev/golang-api-hexagonal/internal/runtimeutil"
)

// ErrRateLimited error code for rate limiting responses.
const ErrRateLimited = "ERR_RATE_LIMITED"

// TokenBucket implements the token bucket algorithm for rate limiting.
// It is thread-safe and can be used concurrently.
//
// The token bucket algorithm works by:
//   - Adding tokens at a constant rate (rate tokens per second)
//   - Allowing requests only if there are available tokens
//   - Each allowed request consumes one token
//   - Tokens accumulate up to a maximum capacity (burst size)
//
// This provides smooth rate limiting with burst allowance.
type TokenBucket struct {
	rate       float64   // tokens per second
	capacity   float64   // max tokens (burst size)
	tokens     float64   // current tokens
	lastRefill time.Time // last refill timestamp
	mu         sync.Mutex
}

// NewTokenBucket creates a new token bucket with the given tokens per second
// and capacity. The bucket starts full (tokens = capacity).
func NewTokenBucket(tokensPerSecond, capacity float64) *TokenBucket {
	return &TokenBucket{
		rate:       tokensPerSecond,
		capacity:   capacity,
		tokens:     capacity, // Start full
		lastRefill: time.Now(),
	}
}

// Allow checks if a request is allowed. Returns true if allowed and consumes
// one token. Returns false if rate limited (no tokens available).
func (b *TokenBucket) Allow() bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Refill tokens based on time elapsed
	now := time.Now()
	elapsed := now.Sub(b.lastRefill).Seconds()
	b.tokens = math.Min(b.capacity, b.tokens+elapsed*b.rate)
	b.lastRefill = now

	// Consume token if available
	if b.tokens >= 1 {
		b.tokens--
		return true
	}
	return false
}

// RetryAfter returns the number of seconds until the next token is available.
// Returns 0 if a token is already available.
func (b *TokenBucket) RetryAfter() int {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Refill first to get accurate count
	now := time.Now()
	elapsed := now.Sub(b.lastRefill).Seconds()
	currentTokens := math.Min(b.capacity, b.tokens+elapsed*b.rate)

	if currentTokens >= 1 {
		return 0
	}

	// Calculate time to generate 1 token
	tokensNeeded := 1.0 - currentTokens
	seconds := tokensNeeded / b.rate
	return int(math.Ceil(seconds))
}

// bucketEntry holds a token bucket with its last access time for cleanup.
type bucketEntry struct {
	bucket     *TokenBucket
	lastAccess time.Time
	customRate *runtimeutil.Rate // optional per-key rate
	mu         sync.Mutex        // protects lastAccess
}

// InMemoryRateLimiter implements runtimeutil.RateLimiter using in-memory storage.
// It provides thread-safe rate limiting with automatic cleanup of expired buckets.
//
// Usage:
//
//	limiter := NewInMemoryRateLimiter(
//	    WithDefaultRate(runtimeutil.NewRate(100, time.Minute)),
//	    WithCleanupInterval(5 * time.Minute),
//	)
//	defer limiter.Stop()
//
//	allowed, _ := limiter.Allow(ctx, "user-123")
type InMemoryRateLimiter struct {
	buckets         sync.Map // map[string]*bucketEntry
	defaultRate     runtimeutil.Rate
	cleanupInterval time.Duration
	bucketTTL       time.Duration // how long to keep inactive buckets
	stopCleanup     chan struct{}
	cleanupOnce     sync.Once
	stopped         bool
	mu              sync.Mutex
}

// Option configures InMemoryRateLimiter.
type Option func(*InMemoryRateLimiter)

// WithDefaultRate sets the default rate limit for new keys.
func WithDefaultRate(rate runtimeutil.Rate) Option {
	return func(l *InMemoryRateLimiter) {
		l.defaultRate = rate
	}
}

// WithCleanupInterval sets how often expired buckets are cleaned up.
func WithCleanupInterval(d time.Duration) Option {
	return func(l *InMemoryRateLimiter) {
		l.cleanupInterval = d
	}
}

// WithBucketTTL sets how long to keep inactive buckets before cleanup.
func WithBucketTTL(d time.Duration) Option {
	return func(l *InMemoryRateLimiter) {
		l.bucketTTL = d
	}
}

// NewInMemoryRateLimiter creates a new in-memory rate limiter.
// Default settings: 100 requests/minute, 5-minute cleanup interval, 10-minute bucket TTL.
func NewInMemoryRateLimiter(opts ...Option) *InMemoryRateLimiter {
	l := &InMemoryRateLimiter{
		defaultRate:     runtimeutil.NewRate(100, time.Minute), // 100 req/min default
		cleanupInterval: 5 * time.Minute,
		bucketTTL:       10 * time.Minute,
		stopCleanup:     make(chan struct{}),
	}

	for _, opt := range opts {
		opt(l)
	}

	l.startCleanup()
	return l
}

// Allow checks if a request for the given key should be allowed.
// Returns true if allowed, false if rate limited.
// Implements runtimeutil.RateLimiter.
func (l *InMemoryRateLimiter) Allow(_ context.Context, key string) (bool, error) {
	entry := l.getOrCreateEntry(key)
	entry.mu.Lock()
	entry.lastAccess = time.Now()
	entry.mu.Unlock()
	return entry.bucket.Allow(), nil
}

// Limit sets a custom rate limit for the given key.
// Implements runtimeutil.RateLimiter.
func (l *InMemoryRateLimiter) Limit(_ context.Context, key string, rate runtimeutil.Rate) error {
	// Convert rate to tokens per second
	tokensPerSecond := float64(rate.Limit) / rate.Period.Seconds()
	capacity := float64(rate.Limit) // burst = limit

	entry := &bucketEntry{
		bucket:     NewTokenBucket(tokensPerSecond, capacity),
		lastAccess: time.Now(),
		customRate: &rate,
	}

	l.buckets.Store(key, entry)
	return nil
}

// RetryAfter returns the number of seconds until the next request for key is allowed.
func (l *InMemoryRateLimiter) RetryAfter(key string) int {
	if val, ok := l.buckets.Load(key); ok {
		entry := val.(*bucketEntry)
		return entry.bucket.RetryAfter()
	}
	return 0
}

// Stop stops the cleanup goroutine. Call this when done with the limiter.
func (l *InMemoryRateLimiter) Stop() {
	l.mu.Lock()
	defer l.mu.Unlock()

	if !l.stopped {
		close(l.stopCleanup)
		l.stopped = true
	}
}

// getOrCreateEntry gets or creates a bucket entry for the key.
func (l *InMemoryRateLimiter) getOrCreateEntry(key string) *bucketEntry {
	if val, ok := l.buckets.Load(key); ok {
		return val.(*bucketEntry)
	}

	// Create new bucket with default rate
	tokensPerSecond := float64(l.defaultRate.Limit) / l.defaultRate.Period.Seconds()
	capacity := float64(l.defaultRate.Limit)

	entry := &bucketEntry{
		bucket:     NewTokenBucket(tokensPerSecond, capacity),
		lastAccess: time.Now(),
	}

	// Use LoadOrStore to handle race condition
	actual, _ := l.buckets.LoadOrStore(key, entry)
	return actual.(*bucketEntry)
}

// startCleanup starts the background cleanup goroutine.
func (l *InMemoryRateLimiter) startCleanup() {
	l.cleanupOnce.Do(func() {
		go func() {
			ticker := time.NewTicker(l.cleanupInterval)
			defer ticker.Stop()

			for {
				select {
				case <-ticker.C:
					l.cleanup()
				case <-l.stopCleanup:
					return
				}
			}
		}()
	})
}

// cleanup removes expired bucket entries.
func (l *InMemoryRateLimiter) cleanup() {
	now := time.Now()
	l.buckets.Range(func(key, value interface{}) bool {
		entry := value.(*bucketEntry)
		entry.mu.Lock()
		lastAccess := entry.lastAccess
		entry.mu.Unlock()
		if now.Sub(lastAccess) > l.bucketTTL {
			l.buckets.Delete(key)
		}
		return true
	})
}

// MiddlewareConfig configures the rate limit middleware.
type MiddlewareConfig struct {
	keyExtractor      func(*http.Request) string
	retryAfterSeconds int
	onLimitExceeded   func(http.ResponseWriter, *http.Request, int) // optional custom handler
}

// MiddlewareOption configures RateLimitMiddleware.
type MiddlewareOption func(*MiddlewareConfig)

// WithKeyExtractor sets a custom key extractor function.
func WithKeyExtractor(fn func(*http.Request) string) MiddlewareOption {
	return func(c *MiddlewareConfig) {
		c.keyExtractor = fn
	}
}

// WithRetryAfterSeconds sets the fixed Retry-After value in seconds.
// If set to 0, the middleware will calculate the actual retry time.
func WithRetryAfterSeconds(seconds int) MiddlewareOption {
	return func(c *MiddlewareConfig) {
		c.retryAfterSeconds = seconds
	}
}

// WithLimitExceededHandler sets a custom handler for rate limit exceeded.
func WithLimitExceededHandler(fn func(http.ResponseWriter, *http.Request, int)) MiddlewareOption {
	return func(c *MiddlewareConfig) {
		c.onLimitExceeded = fn
	}
}

// IPKeyExtractor extracts the client IP address from the request.
// It checks X-Forwarded-For header first (for requests behind proxy),
// then falls back to the RemoteAddr.
//
// Security note: X-Forwarded-For can be spoofed. In high-security scenarios,
// consider using the direct IP or a trusted proxy header.
func IPKeyExtractor(r *http.Request) string {
	// Check X-Forwarded-For first (behind proxy)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// First IP in the list is the original client
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}

	// Check X-Real-IP
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}

	// Fall back to RemoteAddr
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr // Return as-is if parsing fails
	}
	return host
}

// UserIDKeyExtractor extracts the user ID from authenticated claims.
// Falls back to IP address if no claims are present.
func UserIDKeyExtractor(r *http.Request) string {
	claims, err := FromContext(r.Context())
	if err != nil {
		return IPKeyExtractor(r) // Fallback to IP
	}
	if claims.UserID == "" {
		return IPKeyExtractor(r) // Fallback if UserID is empty
	}
	return "user:" + claims.UserID
}

// defaultMiddlewareConfig returns the default middleware configuration.
func defaultMiddlewareConfig() MiddlewareConfig {
	return MiddlewareConfig{
		keyExtractor:      IPKeyExtractor,
		retryAfterSeconds: 0, // 0 means calculate dynamically
	}
}

// RateLimiterWithRetryAfter extends RateLimiter with RetryAfter support.
type RateLimiterWithRetryAfter interface {
	runtimeutil.RateLimiter
	RetryAfter(key string) int
}

// RateLimitMiddleware creates rate limiting middleware.
// It uses the provided RateLimiter to check if requests should be allowed.
//
// Features:
//   - Extracts key from request (IP, UserID, or custom)
//   - Returns 429 Too Many Requests when limit exceeded
//   - Sets Retry-After header with seconds until limit reset
//   - Fail-open: errors from limiter allow request through
//
// Usage:
//
//	limiter := NewInMemoryRateLimiter()
//	r.Use(RateLimitMiddleware(limiter))
//
//	// With custom key extractor
//	r.Use(RateLimitMiddleware(limiter, WithKeyExtractor(UserIDKeyExtractor)))
func RateLimitMiddleware(limiter runtimeutil.RateLimiter, opts ...MiddlewareOption) func(http.Handler) http.Handler {
	cfg := defaultMiddlewareConfig()
	for _, opt := range opts {
		opt(&cfg)
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := cfg.keyExtractor(r)

			allowed, err := limiter.Allow(r.Context(), key)
			if err != nil {
				// Fail-open: if rate limiter errors, allow request
				// Log error in production (using structured logger if available)
				next.ServeHTTP(w, r)
				return
			}

			if !allowed {
				// Calculate retry-after value
				retryAfter := cfg.retryAfterSeconds
				if retryAfter == 0 {
					// Try to get dynamic retry-after from limiter
					if retryLimiter, ok := limiter.(RateLimiterWithRetryAfter); ok {
						retryAfter = retryLimiter.RetryAfter(key)
					}
					if retryAfter == 0 {
						retryAfter = 60 // Default fallback
					}
				}

				// Set Retry-After header
				w.Header().Set("Retry-After", strconv.Itoa(retryAfter))

				// Custom handler if provided
				if cfg.onLimitExceeded != nil {
					cfg.onLimitExceeded(w, r, retryAfter)
					return
				}

				// Default response using project's response package
				response.Error(w, http.StatusTooManyRequests, ErrRateLimited, "Rate limit exceeded")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
