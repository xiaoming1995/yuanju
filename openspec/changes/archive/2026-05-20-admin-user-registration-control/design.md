## Context

普通用户注册当前由 `POST /api/auth/register` 直接创建 `users` 记录，注册成功后立即返回普通用户 JWT。Admin 后台已有独立 JWT 鉴权、用户列表页和 `GET /api/admin/users`，但没有普通用户创建入口，也没有控制公开注册入口的配置。

本变更跨后端认证、Admin API、数据库迁移和前端普通/后台页面。公开注册控制属于安全边界，必须由后端强制执行；前端隐藏入口只是体验优化。

## Goals / Non-Goals

**Goals:**

- 支持 Admin 在后台创建普通用户，初始密码由管理员设置。
- 支持 Admin 开启或关闭公开注册入口。
- 公开注册关闭时，普通注册接口不可被绕过；Admin 创建用户不受影响。
- 默认保持公开注册开启，保证升级后现有部署行为不变。
- 前端在未登录导航、注册页和结果页注册 CTA 中体现注册开关状态。

**Non-Goals:**

- 不实现邀请码、短信/邮箱验证或一次性改密流程。
- 不实现用户禁用、删除、重置密码或角色权限体系。
- 不改变 Admin 账号体系，也不允许普通用户自行成为管理员。
- 不引入第三方 UI 框架或新的认证依赖。

## Decisions

### 1. 使用独立系统设置表保存注册开关

新增 `system_settings` 表或同等持久化机制，至少支持：

```sql
key TEXT PRIMARY KEY
value TEXT NOT NULL
updated_at TIMESTAMPTZ DEFAULT NOW()
```

初始 seed 写入 `registration_enabled = true`。

选择原因：

- 注册开关是站点级配置，不属于算法参数，不应混入 `algo_config`。
- 后续可承载站点公告、维护模式、商业化开关等通用配置。
- 文本 value 足够简单，避免本轮引入 JSON schema 或复杂类型系统。

备选方案：

- 环境变量：部署后修改需要重启，且无法由 Admin 页面管理。
- `algo_config`：已有接口可复用，但语义错误，后续维护成本高。

### 2. 普通注册与 Admin 创建用户分离入口，共享底层创建逻辑

普通注册继续使用 `POST /api/auth/register`，但在服务层创建前检查 `registration_enabled`。Admin 创建用户新增 `POST /api/admin/users`，由 Admin JWT 保护，不检查公开注册开关。

底层应抽出“创建普通用户”的共享函数，统一处理：

- 邮箱唯一校验
- bcrypt 密码加密
- 昵称默认值
- `users` 表插入

选择原因：

- 防止 Admin 创建用户时复制一套不一致的密码/邮箱逻辑。
- 注册关闭只限制公开入口，不限制运营后台。

### 3. 用户来源字段作为推荐实现

建议为 `users` 增加 `source VARCHAR(30) NOT NULL DEFAULT 'self_registered'`，Admin 创建时写入 `admin_created`。用户列表可展示来源，但不是登录鉴权条件。

选择原因：

- 便于运营区分公开注册用户与后台导入/邀约用户。
- 不改变现有认证流程，历史用户通过默认值平滑迁移。

如果实现时希望进一步缩小范围，也可以先不展示来源，但数据库字段仍建议保留。

### 4. 提供公开只读设置接口给前端

新增公开只读接口，例如 `GET /api/auth/registration-settings`，返回：

```json
{ "registration_enabled": true }
```

Admin 侧新增受保护读写接口，例如：

- `GET /api/admin/settings/registration`
- `PUT /api/admin/settings/registration`

选择原因：

- 未登录导航和注册页需要在没有 token 的情况下知道是否展示注册入口。
- Admin 写接口必须受 Admin JWT 保护。

## Risks / Trade-offs

- [Risk] 前端未加载到开关状态时短暂展示注册按钮。→ 默认可先保守展示登录，注册按钮在设置加载成功后再显示；后端仍会强制拒绝。
- [Risk] 管理员设置初始密码后通过不安全渠道告知用户。→ UI 文案提示管理员使用安全渠道发送初始密码；本轮不实现一次性改密。
- [Risk] 设置表 value 为 TEXT 需要解析布尔值。→ repository 层集中解析，只接受 `true` / `false` 写入。
- [Risk] 旧数据库没有 `source` 字段。→ 迁移使用 `DEFAULT 'self_registered' NOT NULL`，历史数据无需回填脚本。

## Migration Plan

1. 新增迁移文件，创建 `system_settings` 表并 seed `registration_enabled=true`。
2. 可选新增 `users.source` 字段，默认 `self_registered`。
3. 后端上线后默认行为仍为公开注册开启。
4. Admin 可在后台关闭公开注册；回滚时重新开启开关即可恢复公开注册。

## Open Questions

- 是否要在用户列表中展示“来源”列？推荐展示，便于验证 Admin 创建用户链路。
- 是否要在 Admin 创建成功后显示一次性“复制账号信息”提示？推荐做轻量成功提示，但不强制保存明文密码。
