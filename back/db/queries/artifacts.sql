-- name: CreateArtifact :one
INSERT INTO artifacts (
  build_id,
  log_id,
  project_id,
  branch_id,
  commit_hash,
  filename,
  s3_key,
  size_bytes,
  content_type,
  artifact_type,
  image_name,
  image_tag,
  image_digest
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
) RETURNING *;

-- name: GetArtifactsByBuildID :many
SELECT * FROM artifacts
WHERE build_id = $1 AND deleted_at IS NULL
ORDER BY created_at ASC;

-- name: GetArtifact :one
SELECT * FROM artifacts
WHERE id = $1 AND deleted_at IS NULL;

-- name: SoftDeleteArtifact :exec
UPDATE artifacts
SET deleted_at = NOW()
WHERE id = $1;

-- name: SoftDeleteArtifactsByBuildID :exec
UPDATE artifacts
SET deleted_at = NOW()
WHERE build_id = $1 AND deleted_at IS NULL;
