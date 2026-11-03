-- +goose Up
ALTER TABLE users ADD COLUMN chart_query_count BIGINT NOT NULL DEFAULT 0;
UPDATE users u SET chart_query_count = (
    SELECT COUNT(*) FROM bazi_charts b WHERE b.user_id = u.id
);

-- +goose Down
ALTER TABLE users DROP COLUMN IF EXISTS chart_query_count;
