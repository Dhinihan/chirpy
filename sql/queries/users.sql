-- name: CreateUser :one
INSERT INTO
  users (id, email, hashed_password)
VALUES
  ($1, $2, $3)
RETURNING
  *;

-- name: ResetUsers :exec
DELETE FROM users;

-- name: GetUser :one
SELECT
  *
FROM
  users
WHERE
  id = $1;

-- name: GetUserByEmail :one
SELECT
  *
FROM
  users
WHERE
  email = $1;

-- name: UpdateUserCredentials :one
UPDATE users
SET
  email = $2,
  hashed_password = $3,
  updated_at = CURRENT_TIMESTAMP
WHERE
  id = $1
RETURNING
  *;

-- name: UpdateUserSetChirpyRed :execrows
UPDATE users
SET
  is_chirpy_red = TRUE
WHERE
  id = $1;
