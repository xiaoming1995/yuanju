-- +goose Up
ALTER TABLE users ADD COLUMN avatar_path VARCHAR(200) NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE users DROP COLUMN IF EXISTS avatar_path;
