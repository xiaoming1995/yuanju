## 1. 数据库迁移

- [x] 1.1 新增 `admins` 表（id UUID、email、password_hash、name、created_at）
- [x] 1.2 新增 `llm_providers` 表（id UUID、name、type、base_url、model、api_key_encrypted、active、created_at、updated_at）
- [x] 1.3 新增 `ai_requests_log` 表（id UUID、chart_id、provider_id、model、duration_ms、status、error_msg、created_at）
- [x] 1.4 在 `database.go` 中添加上述三张表的迁移 SQL
- [x] 1.5 启动时将 `.env` 中已有的 DEEPSEEK_API_KEY / OPENAI_API_KEY 作为种子数据写入 `llm_providers`（若非空且表为空）

## 2. Admin 认证后端

- [x] 2.1 在 `configs/config.go` 中新增 `ADMIN_ENCRYPTION_KEY`（32字节，用于 AES-256-GCM）和 `ADMIN_JWT_SECRET` 配置项
- [x] 2.2 在 `backend/.env` 中增加 `ADMIN_ENCRYPTION_KEY` 和 `ADMIN_JWT_SECRET` 默认值
- [x] 2.3 实现 AES-256-GCM 加解密工具函数（`pkg/crypto/crypto.go`）
- [x] 2.4 新增 `internal/model/admin.go`（Admin 和 LLMProvider 和 AIRequestLog 结构体）
- [x] 2.5 新增 `internal/repository/admin_repository.go`（Admin CRUD 操作）
- [x] 2.6 实现 Admin JWT 颁发（issuer: `yuanju-admin`，有效期24小时）
- [x] 2.7 实现 Admin 注册接口 `POST /api/admin/auth/register`
- [x] 2.8 实现 Admin 登录接口 `POST /api/admin/auth/login`
- [x] 2.9 实现 Admin JWT 鉴权中间件（`internal/middleware/admin_auth.go`）

## 3. LLM Provider 管理后端

- [x] 3.1 新增 `internal/repository/llm_repository.go`（Provider CRUD、激活切换、读取激活 Provider）
- [x] 3.2 实现 `GET /api/admin/llm-providers` — 列出所有 Provider（API Key mask 展示）
- [x] 3.3 实现 `POST /api/admin/llm-providers` — 新增 Provider（API Key 加密存储）
- [x] 3.4 实现 `PUT /api/admin/llm-providers/:id` — 更新 Provider 配置
- [x] 3.5 实现 `PUT /api/admin/llm-providers/:id/activate` — 切换激活 Provider（原子操作：先全部置 false，再置目标为 true）
- [x] 3.6 实现 `DELETE /api/admin/llm-providers/:id` — 删除非激活 Provider

## 4. AI 客户端重构

- [x] 4.1 修改 `ai_client.go`：从 DB 读取 active=true 的 Provider，替代硬编码逻辑
- [x] 4.2 实现 AI 调用日志写入（成功/失败均记录到 `ai_requests_log`）
- [x] 4.3 兼容 fallback：若 DB 中无激活 Provider，降级读取 `.env` 中的旧配置（过渡期）

## 5. Admin 数据统计后端

- [x] 5.1 实现 `GET /api/admin/stats` — 返回 6 项概览指标（总用户/今日新增/总命盘/今日命盘/AI调用总数/今日AI调用）
- [x] 5.2 实现 `GET /api/admin/stats/ai` — 返回 AI 调用详细统计（按 Provider 分类、成功率）
- [x] 5.3 实现 `GET /api/admin/users` — 用户分页列表（支持 `?q=` 邮箱搜索，每条附带命盘数量）

## 6. 前端 Admin 路由与布局

- [x] 6.1 新增 `AdminAuthContext`（独立于普通用户 AuthContext）
- [x] 6.2 新增 `AdminLayout` 组件（深色侧边栏导航：Dashboard / LLM管理 / 用户列表）
- [x] 6.3 新增 `/admin/*` 路由组，未登录自动跳转 `/admin/login`
- [x] 6.4 封装 Admin API 请求层（`src/lib/adminApi.ts`，token 独立存储 `yj_admin_token`）

## 7. Admin 前端页面

- [x] 7.1 开发 `AdminLoginPage`（邮箱+密码表单，登录后跳转 Dashboard）
- [x] 7.2 开发 `AdminDashboardPage`（6张统计卡片：用户数/命盘数/AI调用等）
- [x] 7.3 开发 `AdminLLMPage` — Provider 列表表格（name/type/model/status/激活状态）
- [x] 7.4 开发 LLM Provider 新增/编辑表单（含 Provider 类型下拉、API Key 输入）
- [x] 7.5 开发激活 Provider 切换功能（切换时高亮更新，乐观 UI）
- [x] 7.6 开发 `AdminUsersPage`（分页用户列表，顶部搜索框，显示每用户命盘数量）

## 8. 收尾与验证

- [x] 8.1 在 `backend/.env` 添加 `ADMIN_ENCRYPTION_KEY` 生成说明（`openssl rand -hex 16`）
- [x] 8.2 编写快速创建首个 Admin 账号的 curl 命令文档
- [x] 8.3 端到端测试：注册 Admin → 配置 Provider → 普通用户起盘 → 验证 AI 调用日志
