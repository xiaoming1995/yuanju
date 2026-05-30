# 后台合盘查看 + 用户/起盘管理补齐 — 设计文档

日期：2026-05-30
范围：后台管理控制台（admin）。新增「合盘明细」只读页，补齐用户管理与起盘明细页若干缺口。

## 背景与动机

合盘（compatibility）数据已完整持久化在 4 张表
（`compatibility_readings` / `compatibility_participants` / `compatibility_evidences` / `ai_compatibility_reports`），
但后台**没有任何接口或页面**可查看——而单人八字命盘已有完整的「起盘明细」页。核心诉求即：**后台能查看合盘数据**。

同时顺手补齐相邻缺口：
- 用户列表只加载第 1 页的 Bug（`AdminUsersPage` 硬编码 `users(1, q)`，>20 用户不可达）。
- 后台无法重置用户密码、无法禁用/删除用户。
- 起盘明细页无搜索。

## 现状锚点（已核实）

- 路由注册：`backend/cmd/api/main.go` admin 路由组。
- 已有可参照实现：`AdminListCharts`（`admin_handler.go`）+ `repository.ListBaziCharts` + 前端 `AdminChartsPage.tsx`（分页 + 展开详情 + 缓存删除）。
- 合盘用户侧仓储：`backend/internal/repository/compatibility_repository.go`（在此扩展 admin 函数）。
- 前端 admin API 客户端：`frontend/src/lib/adminApi.ts`；侧边栏导航：`frontend/src/components/AdminLayout.tsx`。
- 登录校验：`backend/internal/service/auth_service.go:103`（bcrypt 比对）；按邮箱取用户：`repository.go` `GetUserByEmail`（line 30 SELECT）。
- 密码哈希：bcrypt；创建用户走 `CreateUserByAdmin`。
- 用户外键（删除语义不对称）：
  - `bazi_charts.user_id` → `ON DELETE SET NULL`（删用户后八字命盘保留，转为游客）。
  - `compatibility_readings.user_id` → `ON DELETE CASCADE`（删用户**连带删除其全部合盘记录**及级联的 participants/evidences/reports）。
- `users` 表**无**禁用/状态列。最新迁移为 `00013`，新迁移为 `00014`。

## 设计决策（已与用户确认）

1. 合盘页：**只读**（仅列表 + 详情，不做搜索/删除/清缓存）。
2. 禁用用户：新增 `disabled_at` 列 + 登录拦截（软禁用，可逆）。
3. 删除用户：**硬删除** + 强警告二次确认（明确写出"将连带删除 N 条合盘记录且不可恢复，八字命盘转为游客"）。
4. 起盘日期搜索：按**排盘时间**（`created_at`）范围。

---

## Phase 1 — 合盘明细页（只读，新建）

结构完全对标现有「起盘明细」页。

### 后端

`compatibility_repository.go` 新增两个 admin 函数：

- `AdminListCompatibilityReadings(page, pageSize int) (rows []AdminCompatListItem, total int, err error)`
  - 列表行字段：`id`、创建用户邮箱（LEFT JOIN users）、双方 participant 的 `display_name` 与生辰摘要（self/partner）、`overall_score`、`overall_level`、`relationship_stage`、`primary_question`、`analysis_version`、`created_at`。
  - 排序 `created_at DESC`，`LIMIT/OFFSET` 分页；`total` 为 `COUNT(*)`。
  - participant 摘要可在一条查询里聚合，或对每条 reading 取 self/partner 两行——优先单查询 JOIN，避免 N+1。
- `AdminGetCompatibilityReadingDetail(id string) (*AdminCompatDetail, error)`
  - 返回完整 reading（含 `dimension_scores`、`score_explanations`、`duration_assessment`、`consulting_assessment`、`summary_tags`）
    + 2 个 participant（含 `chart_snapshot` 四柱）+ evidences 列表 + 最新 `ai_compatibility_reports`（若有）。

`admin_handler.go` 新增（对标 `AdminListCharts` 的签名与错误风格）：

- `AdminListCompatReadings(c *gin.Context)` — 解析 `page`/`pageSize`（pageSize 1–100，默认 20），返回 `{data, total, page}`。
- `AdminGetCompatReadingDetail(c *gin.Context)` — 路径参数 `id`，返回 `{data}`，未找到返回 404。

`main.go` admin 路由组新增：

```
GET /api/admin/compatibility/readings        → AdminListCompatReadings
GET /api/admin/compatibility/readings/:id     → AdminGetCompatReadingDetail
```

### 前端

- 新建 `frontend/src/pages/admin/AdminCompatPage.tsx`，结构参照 `AdminChartsPage.tsx`：
  - 表头：排盘用户 · 排盘时间 · 双方命主 · 总分/等级 · 关系阶段 · 操作（查看详情）。
  - 展开详情面板：双方四柱（复用起盘页四柱卡片样式）、维度评分、期限评估（duration windows）、关系诊断 / 决策建议（consulting_assessment）、证据列表、AI 报告（结构化 content）。
  - 详情按需懒加载（展开时调 detail 接口），对标起盘页流年懒加载。
  - 分页器复用起盘页上一页/下一页控件。
- `adminApi.ts` 新增 `adminCompatAPI`：`list(page, pageSize)`、`detail(id)`。
- `AdminLayout.tsx` 在「起盘明细」后新增导航项「合盘明细」（图标 `Heart`，从 lucide-react 引入）。
- 路由注册（与 `/admin/charts` 同一处）新增 `/admin/compatibility`。

### 验证

- 后台「合盘明细」可见全站合盘流水（含创建者邮箱），分页可翻；展开任意一条能看到双方四柱、评分、AI 报告。
- 游客无法触发合盘（合盘需登录），故列表均有归属用户——无需游客分支，但仍对空邮箱做兜底显示。

---

## Phase 2 — 用户管理补齐

### 2.1 用户列表分页（Bug 修复，前端为主）

- `AdminUsersPage` 增加 `page` 状态；`load(q, page)` 调 `adminStatsAPI.users(page, q)`（后端已支持 `page`）。
- 复用 `AdminChartsPage` 的上一页/下一页 + 「第 X / Y 页」控件；`totalPages = ceil(total/20)`。
- 搜索时重置 `page=1`。
- 验证：构造 >20 用户，能翻到第 2 页。

### 2.2 重置用户密码

- 后端：`POST /api/admin/users/:id/reset-password`，body `{password}`（≥8 位，否则 400）。
  - 仓储新增 `UpdateUserPassword(userID, passwordHash)`（`UPDATE users SET password_hash=$1 WHERE id=$2`）。
  - 哈希复用 bcrypt（与 `CreateUserByAdmin` 同路径）。
- 前端：用户行新增「重置密码」操作 → 弹窗输入新密码（复用创建用户弹窗样式与校验），成功后提示「请通过安全渠道告知用户」。
- 验证：重置后用旧密码登录失败、新密码登录成功。

### 2.3 禁用 / 启用用户（软禁用）

- 迁移 `00014_add_user_disabled_at.sql`：`ALTER TABLE users ADD COLUMN disabled_at TIMESTAMPTZ;`（可空，NULL = 正常）。
- 模型：`User` 增加 `DisabledAt *time.Time`；`GetUserByEmail` 的 SELECT 同步取 `disabled_at`。
- 登录拦截：`auth_service.go` 登录流程中，取到 user 后若 `DisabledAt != nil` 返回明确错误（如「该账号已被禁用」），不进入 bcrypt 成功分支。
- 后端：`PUT /api/admin/users/:id/disable` body `{disabled bool}` → 置 `disabled_at = now()` 或 `NULL`。
  - 仓储 `SetUserDisabled(userID, disabled bool)`。
- 列表：`AdminGetUsers` 查询与返回结构增加 `disabled_at`（前端据此显示状态徽标 + 切换按钮）。
- 前端：用户行显示「正常/已禁用」状态；「禁用/解禁」切换按钮。
- 验证：禁用后该用户登录被拒；解禁后可登录。

### 2.4 删除用户（硬删除 + 强警告）

- 后端：`DELETE /api/admin/users/:id`。
  - 删除前统计该用户合盘记录数（`SELECT COUNT(*) FROM compatibility_readings WHERE user_id=$1`），返回给前端用于确认文案 / 或前端在列表已知。
  - 执行 `DELETE FROM users WHERE id=$1`，依赖既有外键级联（合盘 CASCADE 删除，八字 SET NULL 保留）。
  - 仓储 `DeleteUser(userID)`。
- 前端：用户行「删除」操作 → 二次确认弹窗，明确文案：
  「确认删除用户 {email}？将**连带删除其 N 条合盘记录**（不可恢复）；其八字命盘将转为游客记录保留。」
  - 需输入确认（如再次点击「确认删除」按钮）。
- 验证：删除后用户消失；其合盘记录被清除；其八字命盘仍在起盘明细中显示为游客。

---

## Phase 3 — 起盘明细搜索

- 后端：`AdminListCharts` + `repository.ListBaziCharts` 增加可选参数：
  - `q`：按创建用户邮箱 ILIKE 过滤（JOIN users；游客记录在有 `q` 时排除）。
  - `from` / `to`：按**排盘时间** `created_at` 范围过滤（闭区间，缺省不限）。
- 前端：`AdminChartsPage` 顶部新增搜索栏（复用 `AdminUsersPage` 的 `admin-search-bar` 样式）：邮箱输入 + 起止日期；搜索时 `page=1`。
- 验证：按邮箱能筛出指定用户的起盘；按日期范围能限定排盘时间段。

---

## 实施顺序与拆分

建议按 Phase 1 → 2 → 3 依次实现，各 Phase 可独立交付与验证：

1. Phase 1（合盘只读页）：最高优先，纯新增，不改动既有行为。
2. Phase 2（用户管理）：含 1 个迁移 + 登录路径改动，需回归登录/注册。
3. Phase 3（起盘搜索）：仅扩展既有查询与页面，风险最低。

## 不在本轮范围

- 合盘记录的删除 / 清 AI 缓存（明确只读）。
- 合盘按邮箱搜索（本轮不做，留待后续）。
- 用户角色 / 权限分级。
- 起盘记录删除、起盘按出生日期搜索。

## 风险与注意

- 删除用户对合盘的级联删除是不可逆的——文案必须如实写出影响条数，二次确认。
- `disabled_at` 上线后需确认登录、注册、token 校验等所有取用户路径不被禁用状态破坏（重点回归登录）。
- 合盘详情数据结构存在版本差异（v1/v2/v3，`analysis_version`），前端详情面板需对缺失字段做兜底（参照用户侧 `CompatibilityResultPage` 的降级处理）。
