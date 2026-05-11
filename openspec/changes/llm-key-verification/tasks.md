## Task 1: DB 迁移 — 新增 api_key_preview 列

- [x] 1.1 在 `backend/pkg/database/database.go` 最后一个迁移块后追加 `ALTER TABLE llm_providers ADD COLUMN IF NOT EXISTS api_key_preview VARCHAR(50) DEFAULT ''`

## Task 2: 新增 MaskPlainKey 函数

- [x] 2.1 在 `backend/pkg/crypto/crypto.go` 新增 `MaskPlainKey(plaintext string) string`，格式：前6位 + `***` + 后4位

## Task 3: Model 新增字段

- [x] 3.1 在 `backend/internal/model/admin.go` 的 `LLMProvider` struct 新增 `APIKeyPreview string` 字段

## Task 4: Repository 更新

- [x] 4.1 `ListLLMProviders` 的 SELECT 和 Scan 新增 `api_key_preview` 字段
- [x] 4.2 `CreateLLMProvider` 增加 `preview string` 参数，写入 DB
- [x] 4.3 `UpdateLLMProvider` 增加 `preview string` 参数，有值时更新 DB（空值则保留原值）

## Task 5: Handler 更新

- [x] 5.1 `AdminListProviders`：用 `p.APIKeyPreview` 替代 `crypto.MaskKey(p.APIKeyEncrypted)`
- [x] 5.2 `AdminCreateProvider`：加密前计算 `preview := crypto.MaskPlainKey(req.APIKey)`，传给 repository
- [x] 5.3 `AdminUpdateProvider`：有新 key 时计算 preview，传给 repository
- [x] 5.4 新增 `AdminTestProvider` handler（`POST /api/admin/llm-providers/:id/test`）

## Task 6: Service 新增 TestProviderConnection

- [x] 6.1 在 `backend/internal/service/ai_client.go` 新增 `TestProviderConnection(url, apiKey, modelName string) error`，发送 1-token 探测请求

## Task 7: 注册新路由

- [x] 7.1 在 `backend/cmd/api/main.go` 的 adminAuth 组注册 `POST /llm-providers/:id/test`

## Task 8: 前端 API 客户端

- [x] 8.1 在 `frontend/src/lib/adminApi.ts` 的 `adminLLMAPI` 新增 `test(id)` 方法

## Task 9: 前端页面更新

- [x] 9.1 `AdminLLMPage.tsx` 新增 `testingId` 和 `testResults` state
- [x] 9.2 新增 `handleTest` 函数
- [x] 9.3 列表行加"测试"按钮与结果展示
- [x] 9.4 编辑弹窗显示当前 `api_key_masked`（用 `<code>` 标签展示脱敏 key）

## Task 10: 编译验证与重建镜像

- [x] 10.1 `go build ./...` 验证后端编译通过
- [x] 10.2 `npm run build` 验证前端编译通过
- [x] 10.3 提交代码，`docker-compose up --build -d backend frontend` 重建镜像
