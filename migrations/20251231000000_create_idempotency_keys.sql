-- +goose Up
-- +goose StatementBegin

-- Idempotency keys table for storing cached responses
-- Used by the idempotency middleware to prevent duplicate request processing
CREATE TABLE idempotency_keys (
    key             TEXT PRIMARY KEY,
    request_hash    TEXT NOT NULL,
    status_code     INTEGER NOT NULL,
    response_headers JSONB NOT NULL,
    response_body   BYTEA NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at      TIMESTAMPTZ NOT NULL
);

-- Index for efficient cleanup of expired keys (used by cleanup job)
CREATE INDEX idx_idempotency_keys_expires_at ON idempotency_keys(expires_at);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS idempotency_keys;

-- +goose StatementEnd
