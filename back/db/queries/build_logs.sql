-- name: CreateBuildLog :one
INSERT INTO build_logs (build_id, project_id, branch_id, commit_hash, status, log_message, artifact_s3_key, artifact_url)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: GetBuildLogsByBuildID :many
SELECT * FROM build_logs
WHERE build_id = $1
ORDER BY created_at ASC;

-- name: GetBuildLogsByCommitHash :many
SELECT * FROM build_logs
WHERE commit_hash = $1
ORDER BY created_at ASC;

-- name: UpdateBuildLogArtifact :exec
UPDATE build_logs
SET artifact_s3_key = $2, artifact_url = $3
WHERE id = $1;

