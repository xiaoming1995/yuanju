## Why

当前定价逻辑用硬编码 keyword 匹配模型名（v4-flash/v4-pro）映射到固定三档，价格存在 `algo_config` 表里。随着接入的模型增多（千问、Kimi 等），这套机制无法灵活扩展：加新模型要改 Go 代码，价格管理入口也藏在算法配置页里，与 LLM 管理割裂。

## What Changes

- **后端**：`llm_providers` 表新增 `input_price_cny` / `output_price_cny` 两列，定价随 Provider 一起管理。`CalcCost` 改为按 `model` 字段从 `llm_providers` 表查价格，找不到时 fallback 默认值。Provider 的创建/更新 API 同步支持这两个字段。
- **前端**：`AdminLLMPage` Provider 列表新增输入价/输出价两列；添加/编辑 Provider 的 Modal 新增这两个输入字段。`AlgoConfigPage` 不再展示任何定价相关内容（已还原）。

## Capabilities

### Modified Capabilities
- `llm-management`：LLM Provider 管理，新增每个 Provider 的模型定价配置（输入价/输出价），支持随 Provider 增删改时同步维护价格。

## Impact

- `pkg/database/database.go`：DDL migration 新增两列
- `internal/model/admin.go`：Provider struct 新增两字段
- `internal/repository/admin_repository.go`：Create/Update SQL 同步两字段；新增 `GetPriceByModel` 查询
- `internal/service/llm_pricing.go`：`GetModelPrice` 改为查 DB，fallback 保留
- `internal/handler/admin_handler.go`：Provider 创建/更新接口 body 解析同步两字段
- `frontend/src/pages/admin/AdminLLMPage.tsx`：列表加两列，Modal 加两个输入框
- `frontend/src/lib/adminApi.ts`：Provider 相关类型加两字段
