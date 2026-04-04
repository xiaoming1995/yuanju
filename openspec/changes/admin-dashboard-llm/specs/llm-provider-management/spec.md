## ADDED Requirements

### Requirement: 列出所有 LLM Provider
系统 SHALL 提供接口（`GET /api/admin/llm-providers`）返回所有已配置的 Provider 列表，API Key 以 mask 形式返回（前8位明文 + `***`）。

#### Scenario: 获取 Provider 列表
- **WHEN** Admin 请求 Provider 列表
- **THEN** 返回所有 Provider，包含 id、name、type、base_url、model、masked_key、active 字段

### Requirement: 新增 LLM Provider
系统 SHALL 提供接口（`POST /api/admin/llm-providers`）新增 Provider，API Key 使用 AES-256-GCM 加密后存入数据库。

#### Scenario: 成功新增 Provider
- **WHEN** 提交 name、type、base_url、model、api_key
- **THEN** 返回 201，Provider 已保存，api_key 加密存储

### Requirement: 切换激活 Provider
系统 SHALL 提供接口（`PUT /api/admin/llm-providers/:id/activate`）切换当前激活的 Provider，同一时间只有一个 Provider 处于激活状态。

#### Scenario: 切换激活
- **WHEN** Admin 激活 Provider B
- **THEN** Provider B 的 active=true，其余所有 Provider active=false

#### Scenario: AI 调用使用激活 Provider
- **WHEN** 用户发起八字 AI 报告生成请求
- **THEN** 后端从数据库读取 active=true 的 Provider 发起 LLM 调用

### Requirement: 更新 LLM Provider 配置
系统 SHALL 提供接口（`PUT /api/admin/llm-providers/:id`）更新 Provider 的配置，若 api_key 字段提交则重新加密存储。

#### Scenario: 更新 API Key
- **WHEN** Admin 提交新的 api_key
- **THEN** 系统重新加密存储，下次 AI 调用使用新 Key

### Requirement: 删除 LLM Provider
系统 SHALL 提供接口（`DELETE /api/admin/llm-providers/:id`）删除非激活状态的 Provider。

#### Scenario: 删除非激活 Provider
- **WHEN** 删除 active=false 的 Provider
- **THEN** 返回 200，Provider 已从数据库删除

#### Scenario: 删除激活中的 Provider
- **WHEN** 尝试删除 active=true 的 Provider
- **THEN** 返回 400，提示"请先切换到其他 Provider 再删除"
