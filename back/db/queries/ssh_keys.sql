-- name: CreateSSHKey :one
INSERT INTO ssh_keys (name, private_key, fingerprint)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetSSHKeyByID :one
SELECT * FROM ssh_keys
WHERE id = $1;

-- name: ListSSHKeys :many
SELECT id, name, fingerprint, created_at FROM ssh_keys
ORDER BY created_at DESC;

-- name: DeleteSSHKey :exec
DELETE FROM ssh_keys
WHERE id = $1;
