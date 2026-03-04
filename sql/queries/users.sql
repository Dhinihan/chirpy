-- name: CreateUser :one
INSERT INTO users (id, email)
VALUES ($1, $2)
RETURNING *;

-- name: ResetUsers :exec
DELETE FROM users;

-- name: GetUser :one
SELECT * FROM users WHERE id = $1;
