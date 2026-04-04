# 缘聚 (Yuanju) — Agent 人设

> 本文档为 AI 编程助手的上下文配置文件。每次会话开始时请先阅读本文档。

---

## 项目简介

**缘聚** 是一个中文命理分析 Web 平台，MVP 阶段聚焦**八字（四柱）功能**。  
核心理念：「算法精算 + AI 自然语言解读」双层模式，同时服务普通用户与命理专业用户。

---

## 技术栈

### 后端（`/backend`）
- **语言/框架**：Go + Gin
- **数据库**：PostgreSQL（通过 `lib/pq` 驱动直接操作，无 ORM）
- **身份认证**：JWT（`golang-jwt/jwt/v5`），用户与管理员使用独立 JWT secret
- **加密**：AES-256-GCM（`pkg/crypto/crypto.go`），用于 API Key 加密存储
- **八字算法**：`github.com/6tail/lunar-go v1.4.6`（专业级天文历法库）
- **AI 集成**：主力 DeepSeek API，备选 OpenAI，均通过数据库动态配置
- **配置**：`.env` 文件 + `configs/config.go`

### 前端（`/frontend`）
- **框架**：React 18 + Vite + TypeScript
- **样式**：纯 CSS Variables（无 UI 框架，无 TailwindCSS）
- **路由**：React Router
- **HTTP 客户端**：Axios（含拦截器）
- **字体**：Google Fonts（Noto Sans SC / Inter）

### 基础设施
- **容器化**：Docker + Docker Compose（PostgreSQL + 后端 + 前端）
- **反向代理**：Nginx（前端）

---

## 目录结构

```
yuanju/
├── backend/
│   ├── cmd/api/            # 程序入口
│   ├── configs/            # 环境变量配置加载
│   ├── internal/
│   │   ├── handler/        # HTTP 路由处理器（auth/bazi/admin）
│   │   ├── middleware/     # JWT 鉴权中间件（用户/管理员独立）
│   │   ├── model/          # 数据结构定义（model.go / admin.go）
│   │   ├── repository/     # 数据库操作层（CRUD）
│   │   └── service/        # 业务逻辑层（AI客户端/认证/报告）
│   └── pkg/
│       └── crypto/         # AES-256-GCM 加解密工具
├── frontend/
│   └── src/
│       ├── components/     # 公共 UI 组件
│       ├── contexts/       # React Context（用户Auth / 管理员Auth）
│       ├── lib/            # API 请求层（api.ts / adminApi.ts）
│       └── pages/          # 页面组件
├── openspec/               # OpenSpec 变更管理系统
│   ├── changes/            # 功能变更记录（含 proposal/design/tasks）
│   └── specs/              # 能力规格说明
└── docker-compose.yml
```

---

## 数据库表结构

```sql
-- 用户表
users (id UUID, email, password_hash, created_at)

-- 八字命盘表
bazi_charts (id UUID, user_id, birth_datetime, four_pillars JSONB, wuxing JSONB)

-- AI 解读报告表
ai_reports (id UUID, chart_id, content, model, created_at)

-- 管理员表
admins (id UUID, email, password_hash, name, created_at)

-- LLM Provider 配置表（API Key 加密存储）
llm_providers (id UUID, name, type, base_url, model, api_key_encrypted, active, created_at, updated_at)

-- AI 调用日志表
ai_requests_log (id UUID, chart_id, provider_id, model, duration_ms, status, error_msg, created_at)
```

---

## API 路由概览

```
# 用户认证
POST /api/auth/register
POST /api/auth/login
GET  /api/auth/me

# 八字核心功能
POST /api/bazi/calculate      # 计算八字（无需登录）
POST /api/bazi/report         # 生成 AI 报告（需登录）
GET  /api/bazi/history        # 历史记录列表（需登录）
GET  /api/bazi/history/:id    # 历史记录详情

# 管理后台（独立 Admin JWT）
POST /api/admin/auth/register
POST /api/admin/auth/login
GET  /api/admin/stats
GET  /api/admin/stats/ai
GET  /api/admin/users
GET  /api/admin/llm-providers
POST /api/admin/llm-providers
PUT  /api/admin/llm-providers/:id
PUT  /api/admin/llm-providers/:id/activate
DELETE /api/admin/llm-providers/:id
```

---

## 已完成的功能模块

| 模块 | 状态 | OpenSpec 变更 |
|------|------|--------------|
| 项目脚手架（Go + React + Docker） | ✅ 完成 | yuanju-mvp-bazi |
| 用户认证（注册/登录/JWT） | ✅ 完成 | yuanju-mvp-bazi |
| 八字算法引擎（lunar-go） | ✅ 完成 | yuanju-mvp-bazi |
| AI 报告生成（DeepSeek/OpenAI） | ✅ 完成 | yuanju-mvp-bazi |
| 历史记录模块 | ✅ 完成 | yuanju-mvp-bazi |
| 前端设计系统 + 页面 | ✅ 完成 | yuanju-mvp-bazi |
| 管理后台（Admin 认证） | ✅ 完成 | admin-dashboard-llm |
| LLM Provider 动态管理 | ✅ 完成 | admin-dashboard-llm |
| AI 调用日志与统计 | ✅ 完成 | admin-dashboard-llm |
| 管理后台前端 UI | ✅ 完成 | admin-dashboard-llm |

---

## 开发规范

### 后端
- 所有数据库操作集中在 `internal/repository/` 层，严禁在 handler 中直接写 SQL
- API Key 必须使用 `pkg/crypto` 中的 AES-256-GCM 函数加密后再存入数据库
- 环境变量统一通过 `configs/config.go` 中的 `Config` 结构体读取
- 错误返回格式统一：`{ "error": "message" }`
- 成功返回格式统一：`{ "data": ... }` 或直接返回对象

### 前端
- **禁止**引入 UI 组件库（Ant Design / MUI 等），使用纯 CSS Variables
- 管理员 token 存储 key 为 `yj_admin_token`，普通用户 token 为 `yj_token`
- Admin 和 普通用户使用独立的 Context（`AdminAuthContext` / `AuthContext`）
- API 请求层文件：用户用 `src/lib/api.ts`，管理员用 `src/lib/adminApi.ts`

### OpenSpec 变更管理
- 新功能使用 `/opsx-propose` 工作流创建变更提案
- 实现任务使用 `/opsx-apply` 工作流执行
- 完成后使用 `/opsx-archive` 归档变更
- 变更目录：`openspec/changes/<change-name>/`

---

## 关键业务逻辑

### 八字计算流程
```
用户输入（年月日时 + 性别 + 出生地经度）
  → lunar-go 天文历法库
  → 年柱/月柱/日柱/时柱（天干地支）
  → 五行分布统计
  → 用神喜忌推算
  → 大运序列（起运年龄 + 10步大运）
```

### AI 报告生成流程
```
结构化八字数据
  → 构造 Prompt 模板（性格/感情/事业/健康四章节）
  → 从 DB 读取 active LLM Provider（降级读 .env）
  → 调用 AI API（DeepSeek 优先）
  → 写入 ai_requests_log
  → 缓存报告（相同命盘 hash 命中缓存）
```

---

## 本地开发启动

```bash
# 启动所有服务
docker-compose up -d

# 后端单独开发
cd backend && go run ./cmd/api

# 前端单独开发
cd frontend && npm run dev

# 创建首个管理员账号
curl -X POST http://localhost:8080/api/admin/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@yuanju.com","password":"your-password","name":"Admin"}'
```

---

## 当前阶段与后续规划

**当前阶段**：MVP 已完整交付，包含八字计算、AI报告、管理后台。

**下一阶段候选方向**（未启动）：
- 紫微斗数模块
- 合盘配对功能
- 用户付费订阅
- 命理师入驻平台
- 移动端 App（React Native）

---

## 工程行为准则

详见 → **[ENGINEERING.md](./ENGINEERING.md)**
