-- +goose Up
ALTER TABLE today_question_requests
    ADD COLUMN quota_charged BOOLEAN NOT NULL DEFAULT FALSE;

CREATE INDEX idx_today_question_requests_daily_quota
    ON today_question_requests (user_id, question_date)
    WHERE quota_charged AND status <> 'error';

INSERT INTO system_settings (key, value)
VALUES ('today_question_daily_limit', '1')
ON CONFLICT (key) DO NOTHING;

-- +goose Down
DELETE FROM system_settings WHERE key = 'today_question_daily_limit';
DROP INDEX IF EXISTS idx_today_question_requests_daily_quota;
ALTER TABLE today_question_requests DROP COLUMN quota_charged;
