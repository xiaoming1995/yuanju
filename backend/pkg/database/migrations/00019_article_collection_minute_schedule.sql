-- +goose Up
-- +goose StatementBegin

ALTER TABLE article_collection_config
	ADD COLUMN IF NOT EXISTS schedule_interval_minutes INTEGER NOT NULL DEFAULT 1440;

UPDATE article_collection_config
SET schedule_interval_minutes = CASE
	WHEN frequency = 'weekly' THEN 10080
	WHEN COALESCE(schedule_interval_minutes, 0) > 0 THEN LEAST(GREATEST(schedule_interval_minutes, 1), 10080)
	ELSE 1440
END;

-- +goose StatementEnd
