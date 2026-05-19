-- +goose Up
UPDATE token_usage_logs
SET call_type = 'report'
WHERE call_type = 'report_stream';

-- +goose Down
-- 注意：合并后无法精确还原哪些行原本是流式，统一回退到 report_stream 仅用于回滚演练
UPDATE token_usage_logs
SET call_type = 'report_stream'
WHERE call_type = 'report';
