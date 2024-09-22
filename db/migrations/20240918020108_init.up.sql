BEGIN;

CREATE TABLE IF NOT EXISTS users (
	user_id SERIAL PRIMARY KEY,
	username VARCHAR(128) NOT NULL UNIQUE,
	user_display_name VARCHAR(128) NOT NULL,
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	deleted_at TIMESTAMP NULL DEFAULT NULL
);
CREATE INDEX IF NOT EXISTS idx_users_username
    ON users (username ASC);

CREATE TABLE IF NOT EXISTS rooms (
	room_id SERIAL PRIMARY KEY,
	channel_name VARCHAR(128) NOT NULL UNIQUE,
	is_active BOOLEAN NOT NULL DEFAULT FALSE,
	default_operator_name VARCHAR(128),
	default_operator_config JSON DEFAULT '{}',
	spine_runtime_config JSON DEFAULT '{}',
	garbage_collection_period_mins INT,
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	deleted_at TIMESTAMP NULL DEFAULT NULL
);
CREATE INDEX IF NOT EXISTS idx_rooms_channel_name
    ON rooms (channel_name ASC);


CREATE TABLE IF NOT EXISTS chatters (
	chatter_id SERIAL PRIMARY KEY,
	room_id INT NOT NULL,
	user_id INT NOT NULL,
	is_active boolean NOT NULL DEFAULT TRUE,
	operator_info JSON DEFAULT '{}',
	last_chat_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	deleted_at TIMESTAMP NULL DEFAULT NULL
);
CREATE INDEX IF NOT EXISTS idx_chatters_room_id_user_id
    ON chatters (room_id ASC, user_id ASC);
	
ALTER SEQUENCE users_user_id_seq RESTART WITH 100000;
ALTER SEQUENCE rooms_room_id_seq RESTART WITH 100000;
ALTER SEQUENCE chatters_chatter_id_seq RESTART WITH 100000;	

CREATE OR REPLACE FUNCTION update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER users_update
BEFORE UPDATE ON users
FOR EACH ROW
EXECUTE PROCEDURE update_updated_at();

CREATE TRIGGER rooms_update
BEFORE UPDATE ON rooms
FOR EACH ROW
EXECUTE PROCEDURE update_updated_at();

CREATE TRIGGER chatters_update
BEFORE UPDATE ON chatters
FOR EACH ROW
EXECUTE PROCEDURE update_updated_at();

COMMIT;


