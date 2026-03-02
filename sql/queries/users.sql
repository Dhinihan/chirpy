-- name: CreateUser :one
INSERT INTO users (id, email)
VALUES ($1, $2)
RETURNING *;
