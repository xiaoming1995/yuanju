# LLM 调用内容日志

## 问题

`token_usage_logs` 表已记录每次 LLM 调用的 token 数量，但没有存储实际的输入内容（prompt）和输出内容（response）。发生问题时无法复现，也无法审计或调优。

## 目标

每次 LLM 调用（6 种 call_type：report / report_stream / liunian / dayun / celebrity / compatibility）均完整记录：
- `input_content`：发送给 AI 的 user prompt 原文
- `output_content`：AI 返回的原始内容

## 设计决策

| 决策 | 结论 |
|------|------|
| 存储范围 | 完整存储，不截断 |
| 内容加载 | 按需：点「查看」按钮后单独请求，不随列表返回 |
| system prompt | 不重复存储（已在 ai_prompts 表维护） |
| 新列可空 | 是，历史记录 NULL，不影响现有统计查询 |

## 改动范围

**后端：**
- `pkg/database/database.go`：增量迁移 `token_usage_logs` 加 2 列
- `internal/repository/token_usage_repository.go`：`CreateTokenUsageLog` 加参，新增 `GetTokenUsageContent`
- `internal/service/report_service.go`：4 处调用传入 prompt + rawContent
- `internal/service/celebrity_service.go`：1 处调用
- `internal/service/compatibility_service.go`：1 处调用
- `internal/handler/admin_handler.go`：新增 `GET /api/admin/token-usage/content/:id`
- `cmd/api/main.go`：注册新路由

**前端：**
- `src/pages/admin/TokenUsagePage.tsx`：明细行加「查看」按钮 + content modal
- `src/lib/adminApi.ts`：新增 `getTokenUsageContent(id)` 方法
