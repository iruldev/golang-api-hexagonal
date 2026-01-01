-- Idempotency key queries for sqlc
-- Story 2.5: Idempotency Storage Implementation

-- name: GetIdempotencyKey :one
-- Retrieves an idempotency key record if it exists and hasn't expired
SELECT key, request_hash, status_code, response_headers, response_body, created_at, expires_at
FROM idempotency_keys
WHERE key = $1 AND expires_at > NOW();

-- name: CreateIdempotencyKey :exec
-- Stores a new idempotency key record
INSERT INTO idempotency_keys (key, request_hash, status_code, response_headers, response_body, created_at, expires_at)
VALUES ($1, $2, $3, $4, $5, $6, $7);

-- name: DeleteExpiredIdempotencyKeys :execrows
-- Removes all expired idempotency key records and returns the count of deleted rows
DELETE FROM idempotency_keys WHERE expires_at <= NOW();
