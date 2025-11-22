CREATE TABLE commits_summary (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    start_hash TEXT NOT NULL,
    end_hash TEXT NOT NULL,
    summary_ru TEXT NOT NULL,
    commit_count INTEGER NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(project_id, start_hash, end_hash)
);

CREATE INDEX idx_commits_summary_project_id ON commits_summary(project_id);
CREATE INDEX idx_commits_summary_hashes ON commits_summary(start_hash, end_hash);

