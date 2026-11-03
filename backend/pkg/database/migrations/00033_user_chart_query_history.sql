-- +goose Up
CREATE TABLE user_chart_query_history (
    chart_id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    queried_at TIMESTAMPTZ NOT NULL
);
CREATE INDEX idx_user_chart_query_history_user_time
    ON user_chart_query_history(user_id, queried_at DESC, chart_id DESC);

-- Existing charts have their original creation time.
INSERT INTO user_chart_query_history (chart_id, user_id, queried_at)
SELECT id, user_id, created_at FROM bazi_charts WHERE user_id IS NOT NULL;

-- Recent activity can recover deleted queries that still have an event.
INSERT INTO user_chart_query_history (chart_id, user_id, queried_at)
SELECT object_id, user_id, created_at
FROM user_activity_events
WHERE event_type = 'chart_created' AND object_id IS NOT NULL
ON CONFLICT (chart_id) DO NOTHING;

UPDATE users u SET chart_query_count = GREATEST(
    u.chart_query_count,
    (SELECT COUNT(*) FROM user_chart_query_history h WHERE h.user_id = u.id)
);

-- +goose Down
DROP TABLE IF EXISTS user_chart_query_history;
