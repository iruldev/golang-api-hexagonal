-- +goose Up
-- +goose StatementBegin
-- Story 5.2: Case-Insensitive Email with CITEXT
-- Enable citext extension for case-insensitive text type
CREATE EXTENSION IF NOT EXISTS citext;

-- Change email column from VARCHAR to CITEXT
-- The existing unique index will automatically work case-insensitively
ALTER TABLE users ALTER COLUMN email TYPE CITEXT USING email::citext;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Revert email column back to VARCHAR
ALTER TABLE users ALTER COLUMN email TYPE VARCHAR(255);

-- Remove citext extension (only if no other columns use it)
DROP EXTENSION IF EXISTS citext;
-- +goose StatementEnd
