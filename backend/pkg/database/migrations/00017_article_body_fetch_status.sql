-- +goose Up
-- +goose StatementBegin

ALTER TABLE articles
	ADD COLUMN IF NOT EXISTS body_fetch_status VARCHAR(40) NOT NULL DEFAULT 'pending',
	ADD COLUMN IF NOT EXISTS body_fetch_error TEXT NOT NULL DEFAULT '';

UPDATE articles
SET body_fetch_status = CASE
	WHEN full_text_authorized = true AND COALESCE(body_content, '') <> '' THEN 'succeeded'
	ELSE body_fetch_status
END,
body_fetch_error = COALESCE(body_fetch_error, '');

-- +goose StatementEnd
