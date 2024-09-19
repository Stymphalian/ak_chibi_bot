BEGIN;

DROP TRIGGER IF EXISTS users_update ON Users;
DROP TRIGGER IF EXISTS rooms_update ON Rooms;
DROP TRIGGER IF EXISTS chatters_update ON Chatters;
DROP FUNCTION IF EXISTS update_updated_at();

DROP INDEX IF EXISTS idx_rooms_channel_name;
DROP INDEX IF EXISTS idx_users_username;
DROP INDEX IF EXISTS idx_chatters_room_id_user_id;
DROP TABLE IF EXISTS Chatters;
DROP TABLE IF EXISTS Rooms;
DROP TABLE IF EXISTS Users;

COMMIT;