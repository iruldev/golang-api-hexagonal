-- name: GetUser :one
-- GetUser retrieves a user by ID
SELECT * FROM users WHERE id = $1;

-- name: GetUserByEmail :one
-- GetUserByEmail retrieves a user by email address
SELECT * FROM users WHERE email = $1;

-- name: ListUsers :many
-- ListUsers retrieves all users ordered by creation date
SELECT * FROM users ORDER BY created_at DESC;

-- name: CreateUser :one
-- CreateUser inserts a new user and returns it
INSERT INTO users (email, name)
VALUES ($1, $2)
RETURNING *;

-- name: UpdateUser :one
-- UpdateUser updates a user's name and returns the updated user
UPDATE users
SET name = $2, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteUser :exec
-- DeleteUser removes a user by ID
DELETE FROM users WHERE id = $1;
