## ADDED Requirements

### Requirement: AI 调用日志明细查询
管理员 SHALL 能够通过 Admin API 分页查询 AI 调用日志的完整明细记录，包括调用时间、Provider 名称、模型、耗时、状态和错误信息。

#### Scenario: 查询全部日志
- **WHEN** 管理员请求 `GET /api/admin/ai-logs?page=1`
- **THEN** 系统返回按 `created_at DESC` 排序的日志列表（每页 20 条），包含 `total` 总数字段

#### Scenario: 按状态筛选日志
- **WHEN** 管理员请求 `GET /api/admin/ai-logs?page=1&status=error`
- **THEN** 系统仅返回 `status = 'error'` 的日志记录

#### Scenario: 日志记录包含 Provider 名称
- **WHEN** 日志记录关联了 `provider_id`
- **THEN** 返回结果中 SHALL 包含该 Provider 的 `name` 字段（通过 JOIN 查询）

### Requirement: AI 调用统计摘要
管理员 SHALL 能够查看近 7 天的 AI 调用统计摘要，包括按天分组的调用量、成功/失败数、以及平均响应耗时。

#### Scenario: 查询统计摘要
- **WHEN** 管理员请求 `GET /api/admin/ai-logs/summary`
- **THEN** 系统返回近 7 天每天的 `total`、`success_count`、`error_count` 和 `avg_duration_ms`

#### Scenario: 无数据日期补零
- **WHEN** 近 7 天中某天没有 AI 调用记录
- **THEN** 该天的统计数据 SHALL 返回 `total: 0, success_count: 0, error_count: 0, avg_duration_ms: 0`

### Requirement: Admin 前端 AI 日志页面
Admin 面板 SHALL 提供一个专用的"AI 调用日志"页面，展示日志明细和统计数据。

#### Scenario: 页面展示日志列表
- **WHEN** 管理员访问 `/admin/ai-logs` 页面
- **THEN** 页面显示日志明细表格，列包含：时间、Provider、模型、耗时(ms)、状态、操作（展开错误详情）

#### Scenario: 状态筛选交互
- **WHEN** 管理员点击"失败"筛选标签
- **THEN** 列表仅显示 status 为 error 的记录，分页重置为第 1 页

#### Scenario: 错误详情展开
- **WHEN** 管理员点击一条失败日志的"查看详情"
- **THEN** 该行展开显示完整的 `error_msg` 内容

#### Scenario: 顶部统计卡片
- **WHEN** 页面加载
- **THEN** 页面顶部显示统计卡片：总调用次数、成功率、平均耗时
