-- Users queries for sqlc
-- Story 5.3: Type-safe SQL queries

-- name: CreateUser :exec
INSERT INTO users (id, email, first_name, last_name, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6);

-- name: GetUserByID :one
SELECT id, email, first_name, last_name, created_at, updated_at
FROM users WHERE id = $1;

-- name: ListUsers :many
SELECT id, email, first_name, last_name, created_at, updated_at
FROM users
ORDER BY created_at DESC, id DESC
LIMIT $1 OFFSET $2;

-- name: CountUsers :one
SELECT COUNT(*) FROM users;

-- name: GetUserByEmail :one
SELECT id, email, first_name, last_name, created_at, updated_at
FROM users WHERE email = $1;
