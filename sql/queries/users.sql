-- name: CreateUser :one
INSERT INTO users (id, email, hashed_password)
VALUES ($1, $2, $3)
RETURNING *;

-- name: ResetUsers :exec
DELETE FROM users;

-- name: GetUser :one
SELECT * FROM users WHERE id = $1;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;
