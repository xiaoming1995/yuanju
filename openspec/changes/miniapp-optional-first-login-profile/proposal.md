## Why

已注册用户再次微信登录时自动展开资料设置卡片，会反复打断使用流程。头像和昵称是可选资料，回访登录应直接进入“我的”。新建账号的首次资料引导由后续变更 `miniapp-first-login-profile-autosave` 处理。

## What Changes

- 已注册账号再次登录或刷新资料时，不自动展开头像昵称编辑区。
- 未设置资料时使用现有默认头像和“微信用户”名称。
- 保留点击顶部资料区主动修改头像和昵称的入口，资料编辑可以取消。

## Capabilities

### New Capabilities

无。

### Modified Capabilities

- `dual-channel-user-auth`: 明确回访登录不会重复要求完善资料，编辑入口由用户主动打开。

## Impact

仅涉及小程序“我的”页状态、文案、相关测试及说明；不改变微信登录接口、后端资料接口或数据库。
