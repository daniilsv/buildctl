CREATE TABLE ssh_keys (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name TEXT NOT NULL,
    private_key TEXT NOT NULL,
    fingerprint TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);
