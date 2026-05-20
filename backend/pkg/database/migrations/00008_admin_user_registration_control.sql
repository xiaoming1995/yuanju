-- +goose Up

CREATE TABLE IF NOT EXISTS system_settings (
	key TEXT PRIMARY KEY,
	value TEXT NOT NULL,
	updated_at TIMESTAMPTZ DEFAULT NOW()
);

INSERT INTO system_settings (key, value)
VALUES ('registration_enabled', 'true')
ON CONFLICT (key) DO NOTHING;

ALTER TABLE users
	ADD COLUMN IF NOT EXISTS source VARCHAR(30) NOT NULL DEFAULT 'self_registered';

