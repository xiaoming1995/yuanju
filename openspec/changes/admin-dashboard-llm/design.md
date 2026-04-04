## Context

缘聚命理平台当前的 AI 配置硬编码在 `.env` 文件中，`ai_client.go` 按固定优先级调用 DeepSeek 或 OpenAI。切换 LLM 需要修改文件并重启服务，无法在运行时动态调整。此外平台缺少管理界面，无法查看用户数据或系统运行状态。

本次变更引入独立的 Admin 账号体系与 Admin Dashboard，将 LLM 配置迁移至数据库，实现动态切换，同时提供用户管理和数据统计功能。

## Goals / Non-Goals

**Goals:**
- Admin 账号与普通用户完全隔离（独立表、独立 JWT）
- LLM Provider 配置存入 `llm_providers` 表，运行时动态加载，无需重启
- 支持任意 OpenAI 兼容 Provider（DeepSeek、Qwen、混元等），以及 Anthropic Claude、Google Gemini
- 记录每次 AI 调用（Provider、模型、Token 数量、响应时长）
- Admin 前端：数据概览 + 用户列表 + LLM Provider 管理

**Non-Goals:**
- 不引入 Redis 缓存 Provider 配置（每次调用直接查 DB，规模不大无需缓存）
- 不做按用户计费或限流
- 不实现 Claude 和 Gemini 的原生 SDK 调用（统一用 OpenAI 兼容格式，不兼容的暂不支持）

## Decisions

### D1：LLM Provider 统一 OpenAI 兼容接口

**决策**：所有 Provider 统一使用 OpenAI Chat Completions 格式（POST `/v1/chat/completions`），通过 `base_url` 区分不同服务商。

**理由**：DeepSeek、Qwen、混元、月之暗面等国内主流 LLM 均提供 OpenAI 兼容接口，一套代码维护成本最低。Claude 和 Gemini 的 OpenAI 兼容层也已推出（`api.anthropic.com/v1` / Vertex AI Gemini），可通过配置接入。

**替代方案**：为每个 Provider 实现独立 SDK → 维护成本高，放弃。

### D2：API Key 加密存储

**决策**：使用 AES-256-GCM 对 API Key 进行加密，加密 Key 来自 `.env` 的 `ADMIN_ENCRYPTION_KEY`（32 字节）。DB 存密文，读取时解密。

**理由**：API Key 属敏感数据，不可明文存 DB。使用 Go 标准库 `crypto/aes`，无需引入第三方依赖。

**前端展示**：API Key 仅显示前 8 位 + `***`（如 `sk-abc123***`）。

### D3：Admin 账号独立表

**决策**：新建 `admins` 表，字段与 `users` 表类似（id、email、password_hash、name），但使用独立的 JWT issuer（`yuanju-admin`）与独立的鉴权中间件。

**理由**：普通用户路由和 Admin 路由完全隔离，安全边界清晰，不会因为权限漏洞让普通用户访问 Admin 接口。

### D4：ai_requests_log 按月分区写入（简化版）

**决策**：MVP 阶段单表写入，`created_at` 加索引。后续流量大时再考虑按月分区。

### D5：前端 Admin 路由独立于普通用户

**决策**：`/admin/*` 路由使用独立的 `AdminLayout`（侧边栏导航），与普通用户的 `Navbar` 布局完全隔离。AdminLayout 不带公共 Navbar。

## Risks / Trade-offs

- **[风险] 加密 Key 丢失** → DB 中所有 API Key 无法解密。缓解：`ADMIN_ENCRYPTION_KEY` 必须备份，并在部署文档中明确说明。
- **[风险] DB 读取 Provider 增加延迟** → 每次 AI 调用多一次 DB 查询（~1ms）。在 AI 本身 ~10s 响应时间面前可忽略不计。
- **[取舍] Claude/Gemini 兼容性** → 若服务商的 OpenAI 兼容层不支持某些参数（如 `max_tokens` vs `max_completion_tokens`），需要逐 Provider 适配。MVP 阶段先按标准接口处理，问题出现时再修。

## Migration Plan

1. 执行数据库迁移：新增 `admins`、`llm_providers`、`ai_requests_log` 三张表
2. 启动时将 `.env` 中已有的 API Key 作为种子数据写入 `llm_providers`（若非空）
3. `ai_client.go` 改为从 DB 读取激活 Provider，旧的 `.env` 配置仍保留作为 fallback（过渡期）
4. Admin 前端新增 `/admin/login` 入口，首次使用通过命令行创建第一个 Admin 账号

## Open Questions

- Claude 和 Gemini 是否通过官方 OpenAI 兼容层接入，还是留 TODO 待后续实现？（建议：MVP 先留配置项，调用时若 Provider 类型为 `claude`/`gemini` 返回"暂不支持"提示）
