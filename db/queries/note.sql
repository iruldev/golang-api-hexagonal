-- name: GetNote :one
-- GetNote retrieves a note by ID
SELECT * FROM notes WHERE id = $1;

-- name: ListNotesByUser :many
-- ListNotesByUser retrieves all notes for a user
SELECT * FROM notes WHERE user_id = $1 ORDER BY created_at DESC;

-- name: CreateNote :one
-- CreateNote inserts a new note and returns it
INSERT INTO notes (user_id, title, content)
VALUES ($1, $2, $3)
RETURNING *;

-- name: UpdateNote :one
-- UpdateNote updates a note's title and content
UPDATE notes
SET title = $2, content = $3, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteNote :exec
-- DeleteNote removes a note by ID
DELETE FROM notes WHERE id = $1;
