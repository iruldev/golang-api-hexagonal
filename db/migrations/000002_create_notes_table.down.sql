-- 000002_create_notes_table.down.sql
-- Drops the notes table

DROP INDEX IF EXISTS idx_notes_user_id;
DROP TABLE IF EXISTS notes;
