# Token 成本可视化设计文档

**日期：** 2026-05-12
**范围：** 在现有 token 用量统计基础上，新增按模型单价估算 CNY 费用，在 Admin 汇总与明细视图中展示。

---

## 目标

让管理员在 Token 用量统计页直观看到每个用户、每次 AI 调用的预估费用，无需人工换算 token 数量。

---

## 架构

**核心思路：** 查询时动态计算，不改 `token_usage_logs` 表结构。

```
algo_config（单价 key-value）
       │
       ▼
GET /api/admin/token-usage/summary
GET /api/admin/token-usage/detail
       │
       ├── 读取 token 数量（现有逻辑）
       ├── 从 algo_config 读取模型单价
       └── Go 侧逐行计算 estimated_cost_cny
              │
              ▼
       前端汇总表 + 明细抽屉
       新增"预估费用 (¥)"列
```

**不改动：** `token_usage_logs` 表结构、`CreateTokenUsageLog` 函数、7 个写入点。

---

## 定价配置

向 `algo_config` 表 seed 以下 key（单位：CNY / 百万 tokens）：

| Key | 默认值 | 含义 |
|-----|--------|------|
| `llm_price_flash_input` | `0.27` | deepseek-v4-flash 输入单价 |
| `llm_price_flash_output` | `1.10` | deepseek-v4-flash 输出单价 |
| `llm_price_pro_input` | `4.00` | deepseek-v4-pro 输入单价 |
| `llm_price_pro_output` | `16.00` | deepseek-v4-pro 输出单价 |
| `llm_price_default_input` | `1.00` | 未知模型 fallback 输入单价 |
| `llm_price_default_output` | `2.00` | 未知模型 fallback 输出单价 |

**模型匹配规则（大小写不敏感）：**
- 模型名含 `v4-pro` 或 `reasoner` → pro 价格
- 模型名含 `v4-flash` 或 `chat` → flash 价格
- 其他 → default

**管理方式：** Admin 在现有"算法配置"页直接修改，无需新增 UI。key 在 `pkg/seed/seed.go` 启动时写入，若已存在则不覆盖。

---

## 费用计算逻辑

```
cost = (prompt_tokens  / 1,000,000) × input_price
     + (completion_tokens / 1,000,000) × output_price
```

- `reasoning_tokens` 已包含在 `completion_tokens` 内，不单独计费
- `cache_hit_tokens` 本期简化为与 `input_price` 相同（后续可单独配置 `llm_price_flash_cache`）
- 计算在 Go 侧进行，对每行 log 逐条计算后聚合，比 SQL 侧更灵活应对多模型混用场景

---

## 后端改动

### 1. `pkg/seed/seed.go`
新增 6 个定价 key 的 seed 逻辑（`INSERT ... ON CONFLICT DO NOTHING`）。

### 2. `internal/service/algo_config.go`（或新建 `internal/service/llm_pricing.go`）
新增辅助函数：
```go
// GetModelPrice 返回指定模型的 (inputPriceCNY, outputPriceCNY)，单位：CNY/百万tokens
func GetModelPrice(modelName string) (inputPrice, outputPrice float64)

// CalcCost 根据 token 数量和模型计算费用（CNY）
func CalcCost(modelName string, promptTokens, completionTokens int) float64
```

### 3. `internal/repository/token_usage_repository.go`

**`TokenUsageSummaryRow`** 新增字段：
```go
EstimatedCostCny float64 `json:"estimated_cost_cny"`
```

**`TokenUsageDetailRow`** 新增字段：
```go
EstimatedCostCny float64 `json:"estimated_cost_cny"`
```

**`GetTokenUsageSummary`**：SQL 需按 (user_id, model) 分组聚合 token，Go 侧对每组算费用后合并到用户总费用。具体：将现有聚合 SQL 改为同时 SELECT model，在 Go 侧 group by user_id 后算总 cost。

**`GetTokenUsageDetail`**：在 scan 完每行后，调用 `CalcCost(r.Model, r.PromptTokens, r.CompletionTokens)` 赋值 `r.EstimatedCostCny`。

### 4. `internal/handler/admin_handler.go`
handler 无需改动，repository 返回结构体已含新字段，JSON 序列化自动透传。

---

## 前端改动（`TokenUsagePage.tsx`）

### 汇总表新增列
在"总 tokens"右侧插入：
```tsx
<th style={{ textAlign: 'right' }}>预估费用</th>
// ...
<td style={{ textAlign: 'right', color: '#f59e0b', fontWeight: 700 }}>
  ¥ {row.estimated_cost_cny.toFixed(4)}
</td>
```

### 明细抽屉新增列
在"总计"左侧插入：
```tsx
<th style={{ textAlign: 'right' }}>费用</th>
// ...
<td style={{ textAlign: 'right', color: '#f59e0b', fontSize: 12 }}>
  ¥ {item.estimated_cost_cny.toFixed(4)}
</td>
```

### 说明文字
汇总表底部加：
```tsx
<div style={{ fontSize: 12, color: '#555', marginTop: 8 }}>
  * 基于当前 algo_config 单价估算，仅供参考
</div>
```

### Interface 更新
`SummaryRow` 和 `DetailRow` 各新增 `estimated_cost_cny: number`。

---

## 文件改动清单

| 文件 | 改动类型 |
|------|---------|
| `backend/pkg/seed/seed.go` | 新增 6 个定价 key seed |
| `backend/internal/service/algo_config.go` | 新增 `GetModelPrice` / `CalcCost` |
| `backend/internal/repository/token_usage_repository.go` | struct 加字段，两个查询函数加费用计算 |
| `frontend/src/pages/admin/TokenUsagePage.tsx` | 汇总表 + 明细各加一列，interface 更新 |

---

## 不在本次范围

- cache_hit_tokens 的折扣单价配置
- 费用数据导出（CSV/Excel）
- 按 Provider 分组的费用汇总
- 费用告警/限额功能
