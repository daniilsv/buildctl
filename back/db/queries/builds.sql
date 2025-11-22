-- name: GetBuildByID :one
SELECT * FROM builds WHERE id = $1 LIMIT 1;

-- name: GetBuildByCommitHash :one
SELECT * FROM builds WHERE commit_hash = $1 LIMIT 1;

-- name: GetBuildByProjectBranchCommit :one
SELECT * FROM builds WHERE project_id = $1 AND branch_id = $2 AND commit_hash = $3 LIMIT 1;

-- name: ListBuilds :many
SELECT * FROM builds
WHERE (($1::uuid IS NULL OR $1::uuid = '00000000-0000-0000-0000-000000000000'::uuid) OR project_id = $1)
  AND (($2::uuid IS NULL OR $2::uuid = '00000000-0000-0000-0000-000000000000'::uuid) OR branch_id = $2)
ORDER BY started_at DESC
LIMIT $3 OFFSET $4;

-- name: GetBuildsByBranch :many
SELECT * FROM builds WHERE branch_id = $1 ORDER BY started_at DESC;

-- name: CreateBuild :one
INSERT INTO builds (project_id, branch_id, commit_hash, commit_message, status, started_at)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: UpdateBuildStatus :one
UPDATE builds
SET status = $2,
    finished_at = $3
WHERE id = $1
RETURNING *;
