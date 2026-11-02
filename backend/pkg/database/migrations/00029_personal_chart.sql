-- +goose Up
CREATE TABLE user_personal_charts (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    chart_id UUID NOT NULL REFERENCES bazi_charts(id) ON DELETE CASCADE,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE user_personal_charts;
