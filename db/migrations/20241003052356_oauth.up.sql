BEGIN;
ALTER TABLE users ADD COLUMN IF NOT EXISTS user_role VARCHAR(64) NULL;
COMMIT;