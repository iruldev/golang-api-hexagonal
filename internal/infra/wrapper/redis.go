package wrapper

import (
	"context"
	"time"
)

// DefaultRedisTimeout is applied when context has no deadline for Redis operations.
const DefaultRedisTimeout = 30 * time.Second

// RedisFunc is a function that performs a Redis operation.
// The function should use the provided context for cancellation.
type RedisFunc func(ctx context.Context) error

// DoRedis executes a Redis operation with context check.
// If ctx has no deadline, DefaultRedisTimeout is applied.
// It returns immediately if context is already done.
// Returns context.Canceled or context.DeadlineExceeded if context is done.
//
// Usage:
//
//	err := wrapper.DoRedis(ctx, func(ctx context.Context) error {
//	    return rdb.Set(ctx, "key", "value", 0).Err()
//	})
func DoRedis(ctx context.Context, fn RedisFunc) error {
	// Return early if context is already cancelled
	if err := ctx.Err(); err != nil {
		return err
	}

	// Add timeout only if no deadline is set
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, DefaultRedisTimeout)
		defer cancel()
	}

	return fn(ctx)
}

// RedisResult is a generic function that performs a Redis operation and returns a result.
type RedisResult[T any] func(ctx context.Context) (T, error)

// DoRedisResult executes a Redis operation that returns a result.
// If ctx has no deadline, DefaultRedisTimeout is applied.
// It returns immediately if context is already done.
//
// Usage:
//
//	val, err := wrapper.DoRedisResult(ctx, func(ctx context.Context) (string, error) {
//	    return rdb.Get(ctx, "key").Result()
//	})
func DoRedisResult[T any](ctx context.Context, fn RedisResult[T]) (T, error) {
	var zero T

	// Return early if context is already cancelled
	if err := ctx.Err(); err != nil {
		return zero, err
	}

	// Add timeout only if no deadline is set
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, DefaultRedisTimeout)
		defer cancel()
	}

	return fn(ctx)
}

// RedisPinger interface for Redis clients that support Ping.
type RedisPinger interface {
	Ping(ctx context.Context) error
}

// PingRedis wraps Redis Ping with context check.
// If ctx has no deadline, DefaultRedisTimeout is applied.
// Returns immediately if context is already done.
func PingRedis(ctx context.Context, client RedisPinger) error {
	// Return early if context is already cancelled
	if err := ctx.Err(); err != nil {
		return err
	}

	// Add timeout only if no deadline is set
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, DefaultRedisTimeout)
		defer cancel()
	}

	return client.Ping(ctx)
}
