-- name: CreateSession :one
INSERT INTO sessions (
    id,
    user_id,
    username,
    refresh_token,
    user_agent,
    client_ip,
    is_blocked,
    expires_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
) RETURNING *;

-- name: UpdateSession :one
UPDATE sessions 
SET 
    username = COALESCE($2, username)
WHERE id = $1
RETURNING *;

-- name: GetSession :one
SELECT * FROM sessions
WHERE id = $1 LIMIT 1;

-- name: ListSessionsByUsername :many
SELECT * FROM sessions
WHERE username = $1;

-- name: ListSessionsByUser :many
SELECT * FROM sessions
WHERE user_id = $1;

-- name: UpdateSessionsUsername :many
UPDATE sessions
SET username = $2
WHERE username = $1
RETURNING *;

-- name: BlockSession :exec
UPDATE sessions
SET is_blocked = true
WHERE id = $1;