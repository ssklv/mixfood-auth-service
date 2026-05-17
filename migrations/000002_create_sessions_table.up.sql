CREATE TABLE sessions (
    user_id       BIGINT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    refresh_token TEXT NOT NULL,
    expires_at    TIMESTAMPTZ NOT NULL
);