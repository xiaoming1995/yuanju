## ADDED Requirements

### Requirement: 记录 AI 调用日志
系统 SHALL 在每次 AI 报告生成成功或失败后，向 `ai_requests_log` 表写入一条记录，包含：chart_id、provider_id、model、prompt_tokens（估算）、completion_tokens（估算）、duration_ms、status（success/error）、error_msg。

#### Scenario: 成功调用记录
- **WHEN** AI 调用成功返回内容
- **THEN** 写入 status=success 的日志记录，duration_ms 为实际耗时

#### Scenario: 失败调用记录
- **WHEN** AI 调用超时或返回错误
- **THEN** 写入 status=error 的日志记录，error_msg 包含错误原因

### Requirement: Admin 查询调用统计
系统 SHALL 提供接口（`GET /api/admin/stats/ai`）返回 AI 调用汇总统计：总调用次数、成功率、各 Provider 调用次数、今日调用次数。

#### Scenario: 获取 AI 统计
- **WHEN** Admin 请求 AI 统计数据
- **THEN** 返回包含上述字段的 JSON 对象
