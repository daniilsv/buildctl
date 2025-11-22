-- name: GetBranchByID :one
SELECT * FROM branches WHERE id = $1 LIMIT 1;

-- name: GetBranchByProjectAndName :one
SELECT * FROM branches WHERE project_id = $1 AND name = $2 LIMIT 1;

-- name: ListBranches :many
SELECT * FROM branches WHERE project_id = $1 ORDER BY created_at DESC;

-- name: CreateBranch :one
INSERT INTO branches (project_id, name, settings)
VALUES ($1, $2, $3)
RETURNING *;

-- name: UpdateBranch :one
UPDATE branches
SET settings = COALESCE($2, settings),
    last_successful_commit = COALESCE($3, last_successful_commit),
    last_successful_at = COALESCE($4, last_successful_at)
WHERE id = $1
RETURNING *;

-- name: UpdateBranchLastSuccessful :one
UPDATE branches
SET last_successful_commit = $2,
    last_successful_at = $3
WHERE id = $1
RETURNING *;
