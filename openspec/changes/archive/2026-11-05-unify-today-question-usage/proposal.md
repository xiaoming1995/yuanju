## Why

今日问事专属后台只读取后建的请求表，导致已有的 AI/Token 日志在页面中不可见；新请求还重复保存 Token 数字。管理员需要一条连续的请求记录，同时沿用全站现有的 Token 计算与统计。

## What Changes

- 将今日问事请求记录与原有 `token_usage_logs` 建立明确的一对一关联，Token 用量以原有日志为准。
- 把能够可靠解析问题与上下文的旧今日问事 Token 日志补入专属后台，标为历史调用，不伪造最终用户回答。
- 管理页面统一展示历史与新请求；后续次数限制可按请求记录的账号和北京时间日期计数。

## Capabilities

### New Capabilities

无。

### Modified Capabilities

- `admin-today-question-requests`: 统一历史和新请求的后台视图，并复用共享 Token 用量记录。

## Impact

新增关联/回填迁移，调整今日问事落库事务和后台查询、状态展示；保留现有全站 Token 统计。按日赠送次数、后台配置 N 与充值账本属于后续能力，本次只保留稳定的账号、日期和请求 ID。
