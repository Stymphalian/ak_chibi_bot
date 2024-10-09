BEGIN;
DROP TRIGGER IF EXISTS user_preferences_update ON user_preferences;
DROP INDEX IF EXISTS idx_user_preferences_user_id;
DROP TABLE IF EXISTS user_preferences;
COMMIT;