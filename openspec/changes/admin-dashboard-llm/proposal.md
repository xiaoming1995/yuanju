## Why

当前平台的 AI 报告功能依赖硬编码在 `.env` 中的 LLM 配置，每次切换服务商或更新 API Key 都需要修改文件并重启后端服务，运维体验极差。同时平台缺乏管理入口，无法查看用户数据和系统运行状态。

## What Changes

- 新增独立的 Admin 账号体系（`admins` 表，与普通用户完全隔离）
- 新增 `llm_providers` 表，将 LLM 配置从 `.env` 迁移至数据库，支持运行时动态切换
- 支持多个 LLM Provider（DeepSeek、OpenAI、Qwen、Claude、Gemini 等），可随时扩展
- AI 客户端改为从数据库读取激活的 Provider，修改后无需重启立即生效
- 新增 `ai_requests_log` 表，记录每次 AI 调用信息（Provider、Token 数量、响应时间）
- 新增前端 Admin Dashboard（独立路由 `/admin`），包含：LLM 管理、用户列表、数据统计
- API Key 在数据库中加密存储，前端展示时 mask 处理

## Capabilities

### New Capabilities
- `admin-auth`: Admin 独立账号注册/登录/JWT 鉴权，与普通用户系统完全隔离
- `llm-provider-management`: 在 Admin 面板中对 LLM Provider 进行 CRUD 操作，支持切换激活 Provider，API Key 加密存储
- `ai-request-logging`: 记录每次 AI 调用的 Provider、模型、Token 消耗、响应时长，供统计使用
- `admin-dashboard`: Admin 前端界面，含数据统计概览、用户列表、LLM 配置管理

### Modified Capabilities
- `ai-report-generation`: AI 报告生成从静态读 `.env` 改为动态读数据库激活 Provider，其余行为不变

## Impact

- **后端**：新增 3 张数据库表（admins、llm_providers、ai_requests_log），新增 `/api/admin/*` 路由组，修改 `ai_client.go` 动态读 Provider
- **前端**：新增 `/admin` 路由及独立 AdminLayout，新增 3 个 Admin 页面
- **数据库迁移**：需执行新的 schema 迁移，`ai_client.go` 兼容旧 `.env` 配置作为初始种子数据
- **依赖**：可能引入加密库（Go 标准库 `crypto/aes` 即可，无需第三方）
