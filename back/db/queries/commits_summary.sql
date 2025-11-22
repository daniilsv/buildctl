-- name: CreateCommitsSummary :one
INSERT INTO commits_summary (project_id, start_hash, end_hash, summary_ru, commit_count)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetCommitsSummary :one
SELECT * FROM commits_summary
WHERE project_id = $1 AND start_hash = $2 AND end_hash = $3;

