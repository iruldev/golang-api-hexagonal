-- +goose Up
-- +goose StatementBegin

-- Initial schema setup migration
-- This migration creates a project-level schema_info table for application versioning.
--
-- NOTE: This is SEPARATE from goose's internal `goose_db_version` table:
--   - goose_db_version: Tracks which migrations have been applied (managed by goose)
--   - schema_info: Tracks application/project version metadata (managed by us)
--
-- Purpose: Allows application code to query current schema version without
-- depending on goose internals, and provides a simple verification that
-- the migration system is working correctly.

CREATE TABLE IF NOT EXISTS schema_info (
    id SERIAL PRIMARY KEY,
    version VARCHAR(50) NOT NULL DEFAULT '0.0.1',
    description TEXT,
    initialized_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Insert initial record to verify migration ran (idempotent)
INSERT INTO schema_info (version, description)
VALUES ('0.0.1', 'Initial schema setup - golang-api-hexagonal')
ON CONFLICT (id) DO NOTHING;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS schema_info;

-- +goose StatementEnd
