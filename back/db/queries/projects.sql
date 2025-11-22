-- name: GetProjectByName :one
SELECT * FROM projects WHERE name = $1 LIMIT 1;

-- name: GetProjectByID :one
SELECT * FROM projects WHERE id = $1 LIMIT 1;

-- name: ListProjects :many
SELECT * FROM projects ORDER BY created_at DESC;

-- name: CreateProject :one
INSERT INTO projects (name, title, repository_url, repository_type, access_token, settings)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: UpdateProject :one
UPDATE projects
SET name = COALESCE($2, name),
    title = COALESCE($3, title),
    repository_url = COALESCE($4, repository_url),
    repository_type = COALESCE($5, repository_type),
    access_token = COALESCE($6, access_token),
    settings = COALESCE($7, settings)
WHERE id = $1
RETURNING *;

-- name: DeleteProject :exec
DELETE FROM projects WHERE id = $1;
