BEGIN;

ALTER TABLE users
    ADD COLUMN twitch_user_id VARCHAR(64) NULL;


-- This is the table schema directly copied from the pgstore go library.
-- Ideally we just have the library create the table for us, but I don't want
-- to grant CREATE table access to the role in which the web app is running 
-- under so instead we create here via migration.
CREATE TABLE IF NOT EXISTS http_sessions (
    id BIGSERIAL PRIMARY KEY,
    key BYTEA,
    data BYTEA,
    created_on TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    modified_on TIMESTAMPTZ,
    expires_on TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS http_sessions_expiry_idx ON http_sessions (expires_on);
CREATE INDEX IF NOT EXISTS http_sessions_key_idx ON http_sessions (key);

COMMIT;