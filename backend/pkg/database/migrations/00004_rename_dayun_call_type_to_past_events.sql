-- +goose Up
UPDATE token_usage_logs
SET call_type = 'past_events'
WHERE call_type = 'dayun';

-- +goose Down
UPDATE token_usage_logs
SET call_type = 'dayun'
WHERE call_type = 'past_events';
