-- +goose Up

ALTER TABLE ai_prompts
  ADD COLUMN IF NOT EXISTS version VARCHAR(64) NOT NULL DEFAULT 'unversioned',
  ADD COLUMN IF NOT EXISTS is_customized BOOLEAN NOT NULL DEFAULT FALSE,
  ADD COLUMN IF NOT EXISTS canonical_hash CHAR(64) NOT NULL DEFAULT '';

-- 保守策略：所有现存行视为 admin 自定义，避免首次启动 SyncCanonical 覆盖。
-- 新模块或 admin 主动重置后才会变为 is_customized=false。
UPDATE ai_prompts SET is_customized = TRUE;
