-- +goose Up
UPDATE algo_config
SET key = 'cost_alert_per_user_cost_cny',
    value = '5',
    description = '单用户 AI 总成本告警阈值（CNY，7 天滑窗）'
WHERE key = 'cost_alert_per_chart_cost_cny';

-- +goose Down
UPDATE algo_config
SET key = 'cost_alert_per_chart_cost_cny',
    value = '1',
    description = '单命盘 AI 总成本告警阈值（CNY）'
WHERE key = 'cost_alert_per_user_cost_cny';
