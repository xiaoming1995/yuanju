## Why

今日问事目前可以被同一账号在同一天反复提交，既不符合「每天一次」的使用规则，也无法在以后调整免费次数。需要在服务端原子地限制次数，并给管理员一个可调整的入口。

## What Changes

- 每位用户按北京时间自然日获得默认 1 次今日问事机会；并发提交也不得超额。
- 有效请求及「依据不足」结果计次；AI 服务或输出校验失败退回该次机会。上线前的历史请求不追扣新额度。
- 今日问事后台提供每日次数设置，管理员可调整为 1—100 次；充值与付费次数留待独立额度账本实现。
- 超额时小程序明确说明当日次数已用完及北京时间重置时间。

## Capabilities

### New Capabilities
- `today-question-daily-quota`: 每日额度的计数、并发控制、失败退回和后台配置。

### Modified Capabilities
- `miniapp-today-question`: 超额时给用户明确提示并保留问题供次日或调整额度后重试。
- `admin-today-question-requests`: 在专属管理页调整每日次数。

## Impact

涉及 `today_question_requests`、`system_settings` 数据库表，今日问事用户和管理员 API，以及小程序与管理后台页面。无新依赖；不涉及支付和充值。
