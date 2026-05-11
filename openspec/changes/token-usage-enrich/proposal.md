# Proposal: Token 用量明细扩展（推理 + 缓存）

## 背景

当前 `token_usage_logs` 表只记录 `prompt_tokens / completion_tokens / total_tokens` 三个字段。DeepSeek 和 OpenAI 兼容接口实际返回了更丰富的用量数据，这些信息对成本分析至关重要：

| 字段 | 含义 | 当前状态 |
|------|------|---------|
| `reasoning_tokens` | 推理模型（R1/V4-pro）思考过程消耗的 token，单价远高于普通 completion | ❌ 丢失 |
| `prompt_cache_hit_tokens` | System Prompt 命中 KV Cache 的 token 数，**免费或折扣价** | ❌ 丢失 |
| `prompt_cache_miss_tokens` | 未命中缓存的 prompt token（全价） | ❌ 丢失 |

由于 `reasoning_tokens` 是计费中最贵的部分，不记录就无法分析哪些用户/场景在消耗推理预算。

## 目标

1. 扩展 `TokenUsage` Go struct，解析 DeepSeek/OpenAI 返回的完整用量字段
2. 数据库新增三列，存储每次调用的推理和缓存明细
3. Admin 后台 Token 用量明细页展示新字段

## 不做的事

- 不改计价逻辑（本期只记录，不计算费用）
- 不改 `total_tokens` 语义（保持现有汇总逻辑不变）
- 不改调用路径或 Provider 逻辑

## 涉及文件

| 文件 | 变更类型 |
|------|---------|
| `backend/internal/service/ai_client.go` | 扩展 `TokenUsage` struct + 解析新字段 |
| `backend/pkg/database/database.go` | 新增三列 ALTER TABLE 迁移 |
| `backend/internal/repository/token_usage_repository.go` | INSERT/SELECT/Scan 新增三列 |
| `backend/internal/handler/admin_handler.go` | `CreateTokenUsageLog` 调用点透传新参数 |
| `frontend/src/pages/admin/TokenUsagePage.tsx` | 明细表新增推理/缓存列 |

## 预期效果

Admin → Token 用量 → 点开用户明细后，可以看到每条记录的：
- 推理 token 数（`reasoning_tokens`，灰色标注）
- 缓存命中数（`cache_hit`，绿色标注，代表节省）
- 缓存未命中数（`cache_miss`）
