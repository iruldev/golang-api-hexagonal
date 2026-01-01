// Package postgres provides PostgreSQL database connectivity and repositories.
// This file implements the idempotency storage for safe POST request retries.
package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/iruldev/golang-api-hexagonal/internal/infra/postgres/sqlcgen"
	"github.com/iruldev/golang-api-hexagonal/internal/transport/http/middleware"
)

// pgUniqueViolationCode is the PostgreSQL error code for unique constraint violations.
const pgUniqueViolationCode = "23505"

// ErrKeyAlreadyExists is returned when trying to store a key that already exists.
var ErrKeyAlreadyExists = errors.New("idempotency key already exists")

// IdempotencyRepo implements middleware.IdempotencyStore for PostgreSQL.
// It stores idempotency records for replay of duplicate POST requests.
type IdempotencyRepo struct {
	pool Pooler
}

// NewIdempotencyRepo creates a new IdempotencyRepo instance.
func NewIdempotencyRepo(pool Pooler) *IdempotencyRepo {
	return &IdempotencyRepo{pool: pool}
}

// Get retrieves an existing record by key.
// Returns nil, nil if the key doesn't exist or is expired.
func (r *IdempotencyRepo) Get(ctx context.Context, key string) (*middleware.IdempotencyRecord, error) {
	const op = "idempotencyRepo.Get"

	pool := r.pool.Pool()
	if pool == nil {
		return nil, fmt.Errorf("%s: database not connected", op)
	}
	queries := sqlcgen.New(pool)

	row, err := queries.GetIdempotencyKey(ctx, key)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil // Key not found or expired
	}
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	// Deserialize headers from JSONB
	var headers http.Header
	if err := json.Unmarshal(row.ResponseHeaders, &headers); err != nil {
		return nil, fmt.Errorf("%s: unmarshal headers: %w", op, err)
	}

	return &middleware.IdempotencyRecord{
		Key:             row.Key,
		RequestHash:     row.RequestHash,
		StatusCode:      int(row.StatusCode),
		ResponseHeaders: headers,
		ResponseBody:    row.ResponseBody,
		CreatedAt:       row.CreatedAt.Time,
		ExpiresAt:       row.ExpiresAt.Time,
	}, nil
}

// Store saves a new idempotency record.
// Returns ErrKeyAlreadyExists if the key already exists (race condition handling).
func (r *IdempotencyRepo) Store(ctx context.Context, record *middleware.IdempotencyRecord) error {
	const op = "idempotencyRepo.Store"

	pool := r.pool.Pool()
	if pool == nil {
		return fmt.Errorf("%s: database not connected", op)
	}
	queries := sqlcgen.New(pool)

	// Serialize headers to JSON
	headersJSON, err := json.Marshal(record.ResponseHeaders)
	if err != nil {
		return fmt.Errorf("%s: marshal headers: %w", op, err)
	}

	params := sqlcgen.CreateIdempotencyKeyParams{
		Key:             record.Key,
		RequestHash:     record.RequestHash,
		StatusCode:      int32(record.StatusCode),
		ResponseHeaders: headersJSON,
		ResponseBody:    record.ResponseBody,
		CreatedAt:       pgtype.Timestamptz{Time: record.CreatedAt, Valid: true},
		ExpiresAt:       pgtype.Timestamptz{Time: record.ExpiresAt, Valid: true},
	}

	if err := queries.CreateIdempotencyKey(ctx, params); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolationCode {
			return ErrKeyAlreadyExists
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

// DeleteExpired removes all expired idempotency records.
// Returns the number of deleted records.
func (r *IdempotencyRepo) DeleteExpired(ctx context.Context) (int64, error) {
	const op = "idempotencyRepo.DeleteExpired"

	pool := r.pool.Pool()
	if pool == nil {
		return 0, fmt.Errorf("%s: database not connected", op)
	}
	queries := sqlcgen.New(pool)

	deleted, err := queries.DeleteExpiredIdempotencyKeys(ctx)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return deleted, nil
}

// Ensure IdempotencyRepo implements middleware.IdempotencyStore at compile time.
var _ middleware.IdempotencyStore = (*IdempotencyRepo)(nil)
