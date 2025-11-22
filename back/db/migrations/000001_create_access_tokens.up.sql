CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE access_tokens (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name TEXT NOT NULL,
    token_hash TEXT NOT NULL UNIQUE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    last_used_at TIMESTAMP,
    expires_at TIMESTAMP
);

CREATE INDEX idx_access_tokens_token_hash ON access_tokens(token_hash);

