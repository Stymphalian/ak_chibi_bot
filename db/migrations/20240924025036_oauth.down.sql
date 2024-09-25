BEGIN;

DROP INDEX IF EXISTS http_sessions_expiry_idx;
DROP INDEX IF EXISTS http_sessions_key_idx;
DROP TABLE IF EXISTS http_sessions;
ALTER TABLE users DROP COLUMN IF EXISTS twitch_user_id;

COMMIT;