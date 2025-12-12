-- Create notes table for the Note sample module.
-- This migration demonstrates proper table creation patterns.

CREATE TABLE IF NOT EXISTS notes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(255) NOT NULL,
    content TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Index for efficient ordering by creation time
CREATE INDEX IF NOT EXISTS idx_notes_created_at ON notes(created_at DESC);

COMMENT ON TABLE notes IS 'Sample notes table demonstrating hexagonal architecture patterns';
COMMENT ON COLUMN notes.id IS 'Unique identifier for the note';
COMMENT ON COLUMN notes.title IS 'Note title (required, max 255 chars)';
COMMENT ON COLUMN notes.content IS 'Note body content (optional)';
COMMENT ON COLUMN notes.created_at IS 'Timestamp when note was created';
COMMENT ON COLUMN notes.updated_at IS 'Timestamp when note was last updated';
