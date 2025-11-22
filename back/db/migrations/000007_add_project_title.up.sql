ALTER TABLE projects ADD COLUMN title TEXT NOT NULL DEFAULT '';

UPDATE projects SET title = name WHERE title = '';

CREATE INDEX idx_projects_title ON projects(title);

