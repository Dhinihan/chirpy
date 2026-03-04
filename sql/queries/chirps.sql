-- name: CreateChirp :one
INSERT INTO chirps (id, body, user_id)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetAllChirps :many
SELECT * FROM chirps ORDER BY created_at ASC LIMIT 100 OFFSET 0; 

