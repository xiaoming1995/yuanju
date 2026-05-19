-- +goose Up
ALTER TABLE ai_dayun_summaries ADD COLUMN IF NOT EXISTS algorithm_version VARCHAR(32);
ALTER TABLE ai_reports ADD COLUMN IF NOT EXISTS algorithm_version VARCHAR(32);

COMMENT ON COLUMN ai_dayun_summaries.algorithm_version IS 'Algorithm version under which the row was generated. NULL = v1 (pre-yongshen-priority-realignment baseline).';
COMMENT ON COLUMN ai_reports.algorithm_version IS 'Algorithm version under which the row was generated. NULL = v1 (pre-yongshen-priority-realignment baseline).';

-- +goose Down
ALTER TABLE ai_dayun_summaries DROP COLUMN IF EXISTS algorithm_version;
ALTER TABLE ai_reports DROP COLUMN IF EXISTS algorithm_version;
