# Tasks: Provider 级别思考模式配置

## Task 1: DB 迁移

- [x] 1.1 在 `backend/pkg/database/database.go` 新增 ALTER TABLE：`ALTER TABLE llm_providers ADD COLUMN IF NOT EXISTS thinking_enabled BOOLEAN NOT NULL DEFAULT false`

## Task 2: Model 新增字段

- [x] 2.1 在 `backend/internal/model/admin.go` 的 `LLMProvider` struct 新增 `ThinkingEnabled bool \`db:"thinking_enabled" json:"thinking_enabled"\``

## Task 3: Repository 更新

- [x] 3.1 `ListLLMProviders`：SELECT 和 Scan 新增 `thinking_enabled`（在 `active` 前插入）
- [x] 3.2 `CreateLLMProvider`：函数签名新增 `thinkingEnabled bool`，INSERT 写入
- [x] 3.3 `UpdateLLMProvider`：函数签名新增 `thinkingEnabled bool`，UPDATE 写入

## Task 4: Handler 更新

- [x] 4.1 `AdminCreateProvider`：从请求体解析 `thinking_enabled`，传给 repository
- [x] 4.2 `AdminUpdateProvider`：从请求体解析 `thinking_enabled`，传给 repository
- [x] 4.3 `AdminListProviders`：response 中透传 `thinking_enabled`（model 已有该字段，JSON 自动序列化，确认即可）

## Task 5: AI Client 重构思考模式分派

- [x] 5.1 在 `AIRequest` struct 新增 `Thinking *ThinkingConfig` 字段（`type ThinkingConfig struct { Type string \`json:"type"\` }`），用于 DeepSeek 官方格式
- [x] 5.2 `streamAIWithSystemEx` 改为从 provider 配置读取 `thinkingEnabled bool`，不再接收外部 `*bool` 参数
- [x] 5.3 新增 `applyThinkingToRequest(modelName string, thinkingEnabled bool)` 函数：model 含 `deepseek` 时设置 `Thinking` 字段；其他设置 `EnableThinking` 字段
- [x] 5.4 新增统一入口 `StreamAI`，保留 `StreamAIWithSystem/NoThink` 为兼容别名，由 provider 配置驱动

## Task 6: 前端更新

- [x] 6.1 `AdminLLMPage.tsx`：`Provider` interface 新增 `thinking_enabled: boolean`
- [x] 6.2 `initialForm` 新增 `thinking_enabled: false`；`openEdit` 时同步赋值
- [x] 6.3 表单（Modal）加"思考模式"开关（`<input type="checkbox">`）
- [x] 6.4 Provider 列表新增"思考模式"列，显示"开启"或"—"

## Task 7: 编译验证与重建

- [x] 7.1 `go build ./...` 验证后端编译通过
- [x] 7.2 `npm run build` 验证前端编译通过
- [x] 7.3 提交代码，`docker-compose up --build -d backend frontend` 重建镜像
