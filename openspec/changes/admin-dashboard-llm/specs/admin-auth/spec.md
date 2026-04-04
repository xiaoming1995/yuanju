## ADDED Requirements

### Requirement: Admin 注册
系统 SHALL 提供 Admin 账号注册接口（`POST /api/admin/auth/register`），使用 bcrypt 加密密码，禁止重复邮箱注册。

#### Scenario: 成功注册
- **WHEN** 提交合法的邮箱和密码（≥8位）
- **THEN** 返回 201，包含 admin 信息和 JWT token

#### Scenario: 邮箱已注册
- **WHEN** 提交已注册的邮箱
- **THEN** 返回 409 Conflict

### Requirement: Admin 登录
系统 SHALL 提供 Admin 登录接口（`POST /api/admin/auth/login`），验证邮箱密码，颁发独立的 Admin JWT（issuer: `yuanju-admin`，有效期 24 小时）。

#### Scenario: 成功登录
- **WHEN** 提交正确的邮箱和密码
- **THEN** 返回 200，包含 token 和 admin 信息

#### Scenario: 密码错误
- **WHEN** 密码不匹配
- **THEN** 返回 401，错误信息不透露具体原因

### Requirement: Admin JWT 鉴权中间件
系统 SHALL 对所有 `/api/admin/*` 路由（除登录/注册外）验证 Admin JWT，普通用户 token 不得访问 Admin 路由。

#### Scenario: 普通用户 token 访问 Admin 路由
- **WHEN** 使用普通用户 JWT 请求 `/api/admin/stats`
- **THEN** 返回 401 Unauthorized
