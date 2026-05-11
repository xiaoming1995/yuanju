# Tasks: Token 用量明细扩展（推理 + 缓存）

## Task 1: 扩展 TokenUsage struct + 解析新字段

- [x] 1.1 在 `backend/internal/service/ai_client.go` 扩展 `TokenUsage` struct，新增 `ReasoningTokens`、`CacheHitTokens`、`CacheMissTokens` 字段
- [x] 1.2 更新流式解析中的 `event` 匿名 struct，解析 `completion_tokens_details.reasoning_tokens` 和 `prompt_cache_hit_tokens` / `prompt_cache_miss_tokens`

## Task 2: DB 迁移

- [x] 2.1 在 `backend/pkg/database/database.go` 新增 ALTER TABLE 迁移，为 `token_usage_logs` 增加 `reasoning_tokens`、`cache_hit_tokens`、`cache_miss_tokens` 三列（INT DEFAULT 0）

## Task 3: Repository 更新

- [x] 3.1 `TokenUsageDetailRow` struct 新增三个字段
- [x] 3.2 `CreateTokenUsageLog` 函数签名新增三个参数，INSERT 语句写入新列
- [x] 3.3 `GetTokenUsageDetail` 的 SELECT 和 Scan 新增三列

## Task 4: Handler 调用点更新

- [x] 4.1 搜索所有调用 `repository.CreateTokenUsageLog` 的地方，透传 `usage.ReasoningTokens`、`usage.CacheHitTokens`、`usage.CacheMissTokens`

## Task 5: 前端明细展示

- [x] 5.1 更新 `frontend/src/pages/admin/TokenUsagePage.tsx` 的 `UsageDetail` interface，新增三个字段
- [x] 5.2 明细表新增"推理"和"缓存命中"两列，推理 token > 0 时用灰色标注，cache_hit > 0 时用绿色标注

## Task 6: 编译验证与重建

- [x] 6.1 `go build ./...` 验证后端编译通过
- [x] 6.2 `npm run build` 验证前端编译通过
- [x] 6.3 提交代码并 `docker-compose up --build -d backend frontend`
