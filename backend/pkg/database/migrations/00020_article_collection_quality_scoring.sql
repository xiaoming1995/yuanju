-- +goose Up
-- +goose StatementBegin

ALTER TABLE articles
  ADD COLUMN IF NOT EXISTS quality_score INTEGER NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS quality_reasons JSONB NOT NULL DEFAULT '[]'::jsonb;

ALTER TABLE article_collection_task_items
  ADD COLUMN IF NOT EXISTS quality_score INTEGER NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS quality_reasons JSONB NOT NULL DEFAULT '[]'::jsonb,
  ADD COLUMN IF NOT EXISTS skip_reason TEXT NOT NULL DEFAULT '';

CREATE TABLE IF NOT EXISTS article_quality_config (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  quality_filter_enabled BOOLEAN NOT NULL DEFAULT false,
  min_quality_score INTEGER NOT NULL DEFAULT 60,
  allow_without_body BOOLEAN NOT NULL DEFAULT true,
  bonus_keywords TEXT[] NOT NULL DEFAULT ARRAY[]::TEXT[],
  source_blacklist TEXT[] NOT NULL DEFAULT ARRAY[]::TEXT[],
  preferred_sources TEXT[] NOT NULL DEFAULT ARRAY[]::TEXT[],
  ai_quality_check_enabled BOOLEAN NOT NULL DEFAULT false,
  updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

INSERT INTO article_quality_config (
  quality_filter_enabled,
  min_quality_score,
  allow_without_body,
  bonus_keywords,
  source_blacklist,
  preferred_sources,
  ai_quality_check_enabled
)
SELECT false, 60, true, ARRAY[]::TEXT[], ARRAY[]::TEXT[], ARRAY[]::TEXT[], false
WHERE NOT EXISTS (SELECT 1 FROM article_quality_config);

UPDATE article_quality_config
SET min_quality_score = LEAST(GREATEST(min_quality_score, 0), 100),
    bonus_keywords = COALESCE(bonus_keywords, ARRAY[]::TEXT[]),
    source_blacklist = COALESCE(source_blacklist, ARRAY[]::TEXT[]),
    preferred_sources = COALESCE(preferred_sources, ARRAY[]::TEXT[]);

CREATE INDEX IF NOT EXISTS idx_articles_quality_score ON articles(quality_score DESC);
CREATE INDEX IF NOT EXISTS idx_articles_status_quality ON articles(status, quality_score DESC);
CREATE INDEX IF NOT EXISTS idx_article_collection_task_items_quality ON article_collection_task_items(task_id, quality_score DESC);

-- +goose StatementEnd
