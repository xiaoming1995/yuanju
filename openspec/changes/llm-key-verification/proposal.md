## Why

LLM Provider 管理中存在两个密钥验证盲区：

1. **掩码失效**：列表和编辑弹窗中展示的 `api_key_masked` 是加密密文的前8位（base64 字符），而非原始 key 的掩码，用户无法判断当前存储的是哪个密钥。
2. **输入无反馈**：保存新密钥后，用户无法确认输入是否正确、key 是否能调通对应的 AI 服务。

## What Changes

- 修复 API Key 掩码：创建/更新 Provider 时，额外存储原始明文的脱敏版本（前6位 + `***` + 后4位），列表展示改用该字段
- 新增"测试连接"功能：后端提供 `POST /api/admin/llm-providers/:id/test` 接口，用该 Provider 的 key 向 AI API 发送 1-token 探测请求，验证连通性；前端在列表每行显示测试按钮及结果

## Capabilities

### New Capabilities
- `llm-key-preview`：列表和编辑弹窗正确展示原始 API Key 的脱敏前缀，如 `sk-abcd***1234`
- `llm-connection-test`：管理员可一键测试任意 Provider 的密钥连通性，返回成功/失败及延迟

### Modified Capabilities
- `llm-provider-management`：列表新增"测试"按钮；编辑弹窗"当前密钥"由无意义字符变为可识别掩码

## Impact

- **数据库**：`llm_providers` 表新增 `api_key_preview VARCHAR(50)` 列
- **后端**：`crypto.go` 新增 `MaskPlainKey` 函数；`admin_repository.go` 更新建、改接口；`admin_handler.go` 新增 test handler；`main.go` 注册新路由
- **前端**：`adminApi.ts` 新增 `test(id)` 方法；`AdminLLMPage.tsx` 添加测试按钮与状态展示
- **安全**：明文 key 仅在创建/更新请求的内存中短暂存在，`api_key_preview` 只存脱敏片段，不影响加密存储策略
