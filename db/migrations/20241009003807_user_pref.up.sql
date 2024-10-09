BEGIN;

CREATE TABLE IF NOT EXISTS user_preferences (
    user_preference_id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    operator_info JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_user_preferences_user_id
    ON user_preferences (user_id ASC);

CREATE TRIGGER user_preferences_update
BEFORE UPDATE ON user_preferences
FOR EACH ROW
EXECUTE PROCEDURE update_updated_at();

COMMIT;