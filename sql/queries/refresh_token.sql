-- name: CreateRefreshToken :one
INSERT INTO
  refresh_token (
    token,
    created_at,
    updated_at,
    user_id,
    expires_at,
    revoked_at
  )
VALUES
  (
    $1,
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP,
    $2,
    CURRENT_TIMESTAMP + INTERVAL '60 days',
    NULL
  )
RETURNING
  *;

-- name: CheckRefreshToken :one
SELECT
  user_id
FROM
  refresh_token
WHERE
  token = $1
  AND expires_at > CURRENT_TIMESTAMP
  AND revoked_at IS NULL;

-- name: RevokeRefreshToken :exec
UPDATE refresh_token
SET
  revoked_at = CURRENT_TIMESTAMP
WHERE
  token = $1
  AND expires_at > CURRENT_TIMESTAMP
  AND revoked_at IS NULL;
