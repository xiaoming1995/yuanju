## 1. 数据库与模型

- [x] 1.1 新增迁移文件，创建 `system_settings` 表并 seed `registration_enabled=true`
- [x] 1.2 为 `users` 表新增 `source` 字段，默认 `self_registered`，Admin 创建用户写入 `admin_created`
- [x] 1.3 更新普通用户模型与相关查询扫描逻辑，兼容返回用户来源字段

## 2. 后端设置与注册控制

- [x] 2.1 新增系统设置 repository 方法：读取布尔设置、写入布尔设置、缺省值 fallback
- [x] 2.2 新增公开只读接口 `GET /api/auth/registration-settings`
- [x] 2.3 新增 Admin 设置接口 `GET /api/admin/settings/registration` 和 `PUT /api/admin/settings/registration`
- [x] 2.4 修改普通注册服务，在 `registration_enabled=false` 时拒绝 `POST /api/auth/register` 并返回 403
- [x] 2.5 增加后端测试覆盖注册开启、注册关闭、设置更新无需重启生效

## 3. Admin 创建普通用户后端

- [x] 3.1 抽出共享普通用户创建函数，统一处理邮箱唯一校验、bcrypt 加密、昵称默认值和 source 写入
- [x] 3.2 新增 Admin 接口 `POST /api/admin/users`，接收邮箱、昵称和初始密码
- [x] 3.3 确保 Admin 创建用户不受公开注册开关影响
- [x] 3.4 增加后端测试覆盖 Admin 创建成功、重复邮箱冲突、密码长度校验、非 Admin 请求拒绝

## 4. Admin 前端

- [x] 4.1 扩展 `frontend/src/lib/adminApi.ts`，增加创建用户和注册设置读写 API
- [x] 4.2 改造 `AdminUsersPage`，增加公开注册开关并展示保存状态
- [x] 4.3 在 `AdminUsersPage` 增加创建用户弹窗或表单，字段包括邮箱、昵称、初始密码
- [x] 4.4 创建用户成功后刷新或局部更新用户列表，并显示明确成功/失败反馈
- [x] 4.5 可选展示用户 `source` 字段，用于区分公开注册与后台创建用户

## 5. 普通前端注册入口

- [x] 5.1 扩展普通 API 层，增加公开注册设置读取方法
- [x] 5.2 修改 `Navbar` 未登录状态，根据注册开关展示或隐藏注册按钮
- [x] 5.3 修改 `RegisterPage`，注册关闭时展示不可用提示并禁用提交表单
- [x] 5.4 修改结果页等公开注册 CTA，注册关闭时避免引导用户注册
- [x] 5.5 增加前端测试覆盖注册开关开启/关闭下的关键展示状态

## 6. 验证与收尾

- [x] 6.1 运行后端相关 Go 测试
- [x] 6.2 运行前端测试或构建
- [x] 6.3 手动验证：关闭公开注册后普通注册 API 返回 403，Admin 创建用户仍可登录
- [x] 6.4 手动验证：重新开启公开注册后普通注册页和注册接口恢复可用
- [x] 6.5 更新必要的本地开发说明或 Admin 使用说明
