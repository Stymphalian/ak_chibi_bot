
DROP TRIGGER IF EXISTS users_update ON Users;
DROP TRIGGER IF EXISTS rooms_update ON Rooms;
DROP TRIGGER IF EXISTS chatters_update ON Chatters;
DROP FUNCTION IF EXISTS update_updated_at();
DROP TABLE IF EXISTS Chatters;
DROP TABLE IF EXISTS Rooms;
DROP TABLE IF EXISTS Users;

