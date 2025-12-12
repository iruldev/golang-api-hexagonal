-- Drop notes table and related objects.
-- This is the rollback migration for 20251212000001_create_notes.up.sql

DROP INDEX IF EXISTS idx_notes_created_at;
DROP TABLE IF EXISTS notes;
