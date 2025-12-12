-- Note CRUD queries for sqlc.
-- This file demonstrates type-safe query patterns using sqlc.

-- name: CreateNote :one
-- CreateNote inserts a new note and returns the created record.
INSERT INTO notes (id, title, content, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetNote :one
-- GetNote retrieves a single note by ID.
SELECT * FROM notes WHERE id = $1;

-- name: ListNotes :many
-- ListNotes retrieves notes with pagination, ordered by creation time descending.
SELECT * FROM notes ORDER BY created_at DESC LIMIT $1 OFFSET $2;

-- name: CountNotes :one
-- CountNotes returns the total number of notes (for pagination).
SELECT COUNT(*) FROM notes;

-- name: UpdateNote :one
-- UpdateNote updates a note's title and content, returns the updated record.
UPDATE notes 
SET title = $2, content = $3, updated_at = $4 
WHERE id = $1
RETURNING *;

-- name: DeleteNote :exec
-- DeleteNote removes a note by ID.
DELETE FROM notes WHERE id = $1;
