-- +goose Up
ALTER TABLE ai_dayun_summaries ADD COLUMN years JSONB;
COMMENT ON COLUMN ai_dayun_summaries.years IS '10 个年份卡片 [{year,ganzhi,narrative}, ...]，AI dayun 调用同时产出。NULL=旧缓存需重生';

-- +goose Down
ALTER TABLE ai_dayun_summaries DROP COLUMN years;
