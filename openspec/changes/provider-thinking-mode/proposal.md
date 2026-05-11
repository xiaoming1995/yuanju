# Proposal: Provider 级别思考模式配置

## 背景

当前思考模式（reasoning）的控制方式存在两个问题：

1. **硬编码在调用方**：`StreamAIWithSystem` vs `StreamAIWithSystemNoThink` 分散在各 service 中，切换模型时需改代码
2. **API 字段不标准**：使用 `enable_thinking: false`（Qwen3 专用字段），DeepSeek 官方 API 用的是 `thinking: {type: "enabled/disabled"}`；用 deepseek-v4-pro 时思考模式无法正确控制

## 目标

1. 在 `llm_providers` 表新增 `thinking_enabled BOOLEAN DEFAULT false` 列
2. Admin LLM 管理界面加"思考模式"开关（创建和编辑时可配置）
3. AI client 读取 Provider 配置，自动选择正确的 API 字段：
   - DeepSeek（model 含 `deepseek`）→ `thinking: {type: "enabled/disabled"}`
   - Qwen / 其他 → 保持 `enable_thinking: bool` + `/no_think` 兜底
4. 移除 `StreamAIWithSystemNoThink` 和 `streamAIWithSystemEx` 的 `enableThinking *bool` 参数，统一由 Provider 配置驱动

## 不做的事

- 不实现按 call_type 路由不同 Provider（YAGNI）
- 不改 `callOpenAICompatible`（非流式路径，命理功能不使用）
- 不改 `TestProviderConnection`（测试探针）

## 涉及文件

| 文件 | 变更 |
|------|------|
| `backend/pkg/database/database.go` | ALTER TABLE 新增 `thinking_enabled` 列 |
| `backend/internal/model/admin.go` | `LLMProvider` struct 新增字段 |
| `backend/internal/repository/admin_repository.go` | SELECT/INSERT/UPDATE 新增字段 |
| `backend/internal/handler/admin_handler.go` | create/update 解析 `thinking_enabled` |
| `backend/internal/service/ai_client.go` | 按 model 名称分派正确思考 API；简化函数签名 |
| `frontend/src/pages/admin/AdminLLMPage.tsx` | 表单加"思考模式"开关；列表展示 |

## 预期效果

```
Admin LLM 管理
┌──────────────────┬────────────┬───────────┬──────────────┐
│ 名称             │ 模型       │ 思考模式  │ 操作         │
├──────────────────┼────────────┼───────────┼──────────────┤
│ DeepSeek Flash   │ v4-flash   │ 关闭      │ 激活 编辑 测试│
│ DeepSeek Pro ✓  │ v4-pro     │ 开启      │ 编辑 测试    │
└──────────────────┴────────────┴───────────┴──────────────┘
```

切换 Provider 即可在"有思考"和"无思考"模式间切换，无需改任何代码。
