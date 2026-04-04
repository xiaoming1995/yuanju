## ADDED Requirements

### Requirement: 数据概览统计接口
系统 SHALL 提供接口（`GET /api/admin/stats`）返回平台概览数据：总用户数、今日新增用户、总命盘数、今日新增命盘、AI 调用总次数、今日 AI 调用次数。

#### Scenario: 获取概览数据
- **WHEN** Admin 访问 Dashboard 首页
- **THEN** 返回包含上述 6 个指标的 JSON 对象

### Requirement: 用户列表接口
系统 SHALL 提供接口（`GET /api/admin/users`）返回分页用户列表（默认20条/页），支持按邮箱模糊搜索，每条记录包含：id、email、nickname、created_at、命盘数量。

#### Scenario: 获取用户列表
- **WHEN** Admin 请求第1页用户列表
- **THEN** 返回最多20条用户记录和总数

#### Scenario: 搜索用户
- **WHEN** 传入关键字 `?q=example`
- **THEN** 返回 email 包含 `example` 的用户列表

### Requirement: Admin Dashboard 前端界面
系统 SHALL 提供 `/admin` 路由组（独立 AdminLayout），包含以下页面：
- `/admin/login`：Admin 登录页
- `/admin/dashboard`：数据概览卡片
- `/admin/llm`：LLM Provider 列表与管理表单
- `/admin/users`：用户列表（分页、搜索）

#### Scenario: 未登录访问 Admin 页面
- **WHEN** 未登录用户访问 `/admin/dashboard`
- **THEN** 自动重定向到 `/admin/login`

#### Scenario: Admin 切换激活 Provider
- **WHEN** Admin 在 LLM 页面点击"激活"按钮
- **THEN** 界面立即更新激活状态，后端从下一次 AI 调用起使用新 Provider
