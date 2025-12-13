package idempotency

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// RedisStore implements Store using Redis as the backend.
// It uses atomic SET NX EX for check-and-set operations.
type RedisStore struct {
	client   *redis.Client
	prefix   string
	failMode FailMode
	logger   *zap.Logger
}

// RedisStoreOption is a functional option for configuring RedisStore.
type RedisStoreOption func(*RedisStore)

// WithFailMode sets the fail mode for Redis errors.
func WithFailMode(mode FailMode) RedisStoreOption {
	return func(s *RedisStore) {
		s.failMode = mode
	}
}

// WithLogger sets the logger for the store.
func WithLogger(logger *zap.Logger) RedisStoreOption {
	return func(s *RedisStore) {
		s.logger = logger
	}
}

// NewRedisStore creates a new Redis-backed idempotency store.
//
// Parameters:
//   - client: Redis client (from go-redis/v9)
//   - prefix: Key prefix for namespacing (e.g., "idempotency:")
//   - opts: Functional options for configuration
//
// Example:
//
//	store := idempotency.NewRedisStore(redisClient.Client(), "idempotency:",
//	    idempotency.WithFailMode(idempotency.FailClosed),
//	    idempotency.WithLogger(logger),
//	)
func NewRedisStore(client *redis.Client, prefix string, opts ...RedisStoreOption) *RedisStore {
	if prefix == "" {
		prefix = DefaultKeyPrefix
	}

	s := &RedisStore{
		client:   client,
		prefix:   prefix,
		failMode: FailOpen, // Default to fail-open
		logger:   zap.NewNop(),
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

// Check atomically checks if a key is new and marks it as seen.
// Uses SET NX EX for atomic check-and-set with TTL.
//
// Returns:
//   - (true, nil): Key is new (first occurrence), task should be processed
//   - (false, nil): Key exists (duplicate), task should be skipped
//   - (true, nil): On Redis error with FailOpen mode (process anyway)
//   - (false, error): On Redis error with FailClosed mode
func (s *RedisStore) Check(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	if key == "" {
		return true, nil // Empty key = no idempotency, process normally
	}

	fullKey := s.prefix + key

	// SET key value NX EX ttl - atomic check-and-set
	ok, err := s.client.SetNX(ctx, fullKey, "1", ttl).Result()
	if err != nil {
		return s.handleRedisError(err, key)
	}

	if !ok {
		// Key already exists = duplicate
		s.logger.Debug("duplicate task detected",
			zap.String("idempotency_key", key),
		)
	}

	return ok, nil // true = new key, false = duplicate
}

// StoreResult stores a result for a given idempotency key.
// The result is stored in a separate key with "result:" prefix.
func (s *RedisStore) StoreResult(ctx context.Context, key string, result []byte, ttl time.Duration) error {
	if key == "" {
		return nil
	}

	fullKey := s.prefix + "result:" + key

	err := s.client.Set(ctx, fullKey, result, ttl).Err()
	if err != nil {
		s.logger.Warn("failed to store idempotency result",
			zap.String("idempotency_key", key),
			zap.Error(err),
		)
		// Don't fail the operation if result storage fails
		return nil
	}

	return nil
}

// GetResult retrieves a stored result for a given idempotency key.
func (s *RedisStore) GetResult(ctx context.Context, key string) ([]byte, bool, error) {
	if key == "" {
		return nil, false, nil
	}

	fullKey := s.prefix + "result:" + key

	result, err := s.client.Get(ctx, fullKey).Bytes()
	if err == redis.Nil {
		return nil, false, nil // Not found
	}
	if err != nil {
		s.logger.Warn("failed to get idempotency result",
			zap.String("idempotency_key", key),
			zap.Error(err),
		)
		return nil, false, nil // Fail gracefully
	}

	return result, true, nil
}

// SetFailMode updates the fail mode.
func (s *RedisStore) SetFailMode(mode FailMode) {
	s.failMode = mode
}

// handleRedisError handles Redis errors based on fail mode.
func (s *RedisStore) handleRedisError(err error, key string) (bool, error) {
	if s.failMode == FailOpen {
		s.logger.Warn("idempotency check failed, processing anyway (fail-open)",
			zap.String("idempotency_key", key),
			zap.Error(err),
		)
		return true, nil // Process the task
	}

	s.logger.Error("idempotency check failed (fail-closed)",
		zap.String("idempotency_key", key),
		zap.Error(err),
	)
	return false, fmt.Errorf("idempotency check failed: %w", err)
}

// Compile-time check that RedisStore implements Store.
var _ Store = (*RedisStore)(nil)
