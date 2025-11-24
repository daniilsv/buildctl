CREATE TABLE artifacts (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  build_id UUID NOT NULL REFERENCES builds(id) ON DELETE CASCADE,
  log_id UUID REFERENCES build_logs(id) ON DELETE SET NULL,
  project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  branch_id UUID NOT NULL REFERENCES branches(id) ON DELETE CASCADE,
  commit_hash TEXT NOT NULL,

  -- Базовая информация о файле
  filename TEXT NOT NULL,
  s3_key TEXT NOT NULL,
  size_bytes BIGINT NOT NULL,
  content_type TEXT,

  -- Тип артефакта
  artifact_type TEXT NOT NULL DEFAULT 'file', -- 'file' или 'container_image'

  -- Для контейнерных образов (NULL для обычных файлов)
  image_name TEXT,
  image_tag TEXT,
  image_digest TEXT,

  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  deleted_at TIMESTAMP
);

CREATE INDEX idx_artifacts_build_id ON artifacts(build_id);
CREATE INDEX idx_artifacts_commit_hash ON artifacts(commit_hash);
CREATE INDEX idx_artifacts_deleted_at ON artifacts(deleted_at);
