-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS system_settings (
	key TEXT PRIMARY KEY,
	value TEXT NOT NULL,
	updated_at TIMESTAMPTZ DEFAULT NOW()
);

INSERT INTO system_settings (key, value)
VALUES ('articles_module_enabled', 'true')
ON CONFLICT (key) DO NOTHING;

-- +goose StatementEnd

