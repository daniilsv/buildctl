CREATE TABLE build_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    build_id UUID NOT NULL REFERENCES builds(id) ON DELETE CASCADE,
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    branch_id UUID NOT NULL REFERENCES branches(id) ON DELETE CASCADE,
    commit_hash TEXT NOT NULL,
    status TEXT NOT NULL,
    log_message TEXT NOT NULL,
    artifact_s3_key TEXT,
    artifact_url TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_build_logs_build_id ON build_logs(build_id);
CREATE INDEX idx_build_logs_commit_hash ON build_logs(commit_hash);
CREATE INDEX idx_build_logs_project_id ON build_logs(project_id);
CREATE INDEX idx_build_logs_branch_id ON build_logs(branch_id);

