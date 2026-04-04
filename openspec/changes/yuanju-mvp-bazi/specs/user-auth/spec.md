## ADDED Requirements

### Requirement: 用户注册
系统 SHALL 允许用户通过邮箱和密码完成注册，并对密码进行 bcrypt 加密存储。

#### Scenario: 注册成功
- **WHEN** 用户提交有效邮箱和符合规则的密码（至少8位）
- **THEN** 系统创建账号，返回 JWT token，用户自动登录

#### Scenario: 邮箱已存在
- **WHEN** 用户提交的邮箱已被注册
- **THEN** 系统返回 409 错误，提示"该邮箱已被注册"

#### Scenario: 密码不符合规则
- **WHEN** 用户提交的密码少于8位
- **THEN** 系统返回 422 错误，提示密码规则

---

### Requirement: 用户登录
系统 SHALL 支持用户通过邮箱密码登录，成功后返回 JWT access token。

#### Scenario: 登录成功
- **WHEN** 用户提交正确的邮箱和密码
- **THEN** 系统返回 JWT token（有效期7天）和用户基本信息

#### Scenario: 密码错误
- **WHEN** 用户提交的密码不匹配
- **THEN** 系统返回 401 错误，提示"邮箱或密码错误"

---

### Requirement: JWT 鉴权中间件
系统 SHALL 对需要登录的接口进行 JWT 验证，无效 token 拒绝访问。

#### Scenario: 有效 token 访问受保护接口
- **WHEN** 请求头携带有效的 Bearer token
- **THEN** 系统正常处理请求并返回数据

#### Scenario: 无效 token 访问受保护接口
- **WHEN** 请求头缺少 token 或 token 已过期
- **THEN** 系统返回 401 错误
