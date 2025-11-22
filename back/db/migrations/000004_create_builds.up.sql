CREATE TABLE builds (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    branch_id UUID NOT NULL REFERENCES branches(id) ON DELETE CASCADE,
    commit_hash TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'queued',
    started_at TIMESTAMP NOT NULL DEFAULT NOW(),
    finished_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_builds_commit_hash ON builds(commit_hash);
CREATE INDEX idx_builds_project_id ON builds(project_id);
CREATE INDEX idx_builds_branch_id ON builds(branch_id);
CREATE INDEX idx_builds_status ON builds(status);

