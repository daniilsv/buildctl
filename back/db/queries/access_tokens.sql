-- name: CreateAccessToken :one
INSERT INTO access_tokens (name, token_hash, expires_at)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetAccessTokenByHash :one
SELECT * FROM access_tokens
WHERE token_hash = $1;

-- name: ListAccessTokens :many
SELECT * FROM access_tokens
ORDER BY created_at DESC;

-- name: UpdateAccessTokenLastUsed :exec
UPDATE access_tokens
SET last_used_at = NOW()
WHERE id = $1;

-- name: DeleteAccessToken :exec
DELETE FROM access_tokens
WHERE id = $1;

