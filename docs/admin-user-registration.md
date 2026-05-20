# Admin 用户注册控制

## 功能概览

管理员可在后台用户列表页管理普通用户进入方式：

- 开启或关闭公开注册入口。
- 使用邮箱、昵称和初始密码创建普通用户。
- 查看用户来源：公开注册或后台创建。

公开注册关闭时，前台注册页和注册按钮会隐藏或显示不可用提示；后端 `POST /api/auth/register` 同时返回 403，避免绕过前端。

## API

- `GET /api/auth/registration-settings`：公开读取注册开关。
- `GET /api/admin/settings/registration`：Admin 读取注册开关。
- `PUT /api/admin/settings/registration`：Admin 更新注册开关，请求体为 `{"registration_enabled": true}`。
- `POST /api/admin/users`：Admin 创建普通用户，请求体为 `{"email":"user@example.com","password":"password123","nickname":"昵称"}`。

## 默认行为

迁移会创建 `system_settings` 表，并写入 `registration_enabled=true`。升级后默认仍允许公开注册，不改变现有站点行为。
