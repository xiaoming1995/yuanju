-- +goose Up
-- +goose StatementBegin

ALTER TABLE article_collection_config
	ADD COLUMN IF NOT EXISTS search_page_min INTEGER NOT NULL DEFAULT 1,
	ADD COLUMN IF NOT EXISTS search_page_max INTEGER NOT NULL DEFAULT 5;

UPDATE article_collection_config
SET search_page_min = LEAST(GREATEST(1, COALESCE(search_page_min, 1)), 20),
	search_page_max = LEAST(
		20,
		GREATEST(LEAST(GREATEST(1, COALESCE(search_page_min, 1)), 20), COALESCE(search_page_max, 5))
	);

-- +goose StatementEnd
