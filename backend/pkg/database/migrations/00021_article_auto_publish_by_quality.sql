-- +goose Up
-- +goose StatementBegin

ALTER TABLE article_collection_config
  ADD COLUMN IF NOT EXISTS auto_publish_enabled BOOLEAN NOT NULL DEFAULT false,
  ADD COLUMN IF NOT EXISTS auto_publish_min_quality_score INTEGER NOT NULL DEFAULT 90,
  ADD COLUMN IF NOT EXISTS auto_publish_requires_body BOOLEAN NOT NULL DEFAULT true,
  ADD COLUMN IF NOT EXISTS auto_publish_max_per_run INTEGER NOT NULL DEFAULT 3;

UPDATE article_collection_config
SET auto_publish_min_quality_score = LEAST(GREATEST(auto_publish_min_quality_score, 0), 100),
    auto_publish_max_per_run = LEAST(GREATEST(auto_publish_max_per_run, 0), 20);

ALTER TABLE article_collection_task_items
  ADD COLUMN IF NOT EXISTS auto_published BOOLEAN NOT NULL DEFAULT false,
  ADD COLUMN IF NOT EXISTS auto_publish_reason TEXT NOT NULL DEFAULT '';

CREATE INDEX IF NOT EXISTS idx_article_collection_task_items_auto_published
  ON article_collection_task_items(task_id, auto_published);

-- +goose StatementEnd
