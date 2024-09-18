CREATE TABLE IF NOT EXISTS users (
	user_id SERIAL PRIMARY KEY,
	username VARCHAR(128) NOT NULL UNIQUE,
	user_display_name VARCHAR(128) NOT NULL,
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	deleted_at TIMESTAMP NULL DEFAULT NULL
);

CREATE TABLE IF NOT EXISTS rooms (
	room_id SERIAL PRIMARY KEY,
	channel_name VARCHAR(128) NOT NULL UNIQUE,
	is_active BOOLEAN NOT NULL DEFAULT FALSE,

	default_operator_name VARCHAR(128),
	default_operator_skin VARCHAR(128),
	default_operator_stance  VARCHAR(128),
	default_operator_animations TEXT[],
	default_operator_position_x DOUBLE PRECISION,

	garbage_collection_period_mins INT,

	spine_runtime_config_default_animation_speed DOUBLE PRECISION,
	spine_runtime_config_min_animation_speed DOUBLE PRECISION,
	spine_runtime_config_max_animation_speed DOUBLE PRECISION,
	spine_runtime_config_default_scale_size DOUBLE PRECISION,
	spine_runtime_config_min_scale_size DOUBLE PRECISION,
	spine_runtime_config_max_scale_size DOUBLE PRECISION,
	spine_runtime_config_max_sprite_pixel_size INT,
	spine_runtime_config_reference_movement_speed_px INT,
	spine_runtime_config_default_movement_speed DOUBLE PRECISION,
	spine_runtime_config_min_movement_speed DOUBLE PRECISION,
	spine_runtime_config_max_movement_speed DOUBLE PRECISION,

	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	deleted_at TIMESTAMP NULL DEFAULT NULL
);
CREATE INDEX IF NOT EXISTS idx_rooms_channel_name
    ON rooms (channel_name ASC);
ALTER SEQUENCE rooms_room_id_seq RESTART WITH 100000;


CREATE TABLE IF NOT EXISTS chatters (
	chatter_id SERIAL PRIMARY KEY,
	room_id INT NOT NULL REFERENCES rooms(room_id),
	user_id INT NOT NULL REFERENCES users(user_id),
	operator_id VARCHAR(128),
	last_chat_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	deleted_at TIMESTAMP NULL DEFAULT NULL
);

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



