BEGIN;
CREATE TABLE IF NOT EXISTS asset_files (
    file_path_hash  BIGINT    PRIMARY KEY,
    file_path       TEXT      NOT NULL,
    data            BYTEA     NOT NULL,
    created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
COMMIT;
