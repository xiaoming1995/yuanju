# Token 用量统计设计文档

**日期**：2026-05-11  
**目标**：为管理员提供每个用户的 AI token 消耗统计，支持按时间聚合查看和单次调用明细下钻。

---

## 背景

项目通过 `internal/service/ai_client.go` 调用 OpenAI 兼容接口，涉及 6 个 AI 调用点（报告、流式报告、流年、大运、名人、合盘）。当前所有调用均未捕获 API 响应中的 `usage` 字段，无法统计 token 消耗。

---

## 数据捕获层（`ai_client.go`）

### 新增结构体

```go
type TokenUsage struct {
    PromptTokens     int `json:"prompt_tokens"`
    CompletionTokens int `json:"completion_tokens"`
    TotalTokens      int `json:"total_tokens"`
}
```

将 `TokenUsage` 加入 `AIResponse`：

```go
type AIResponse struct {
    Choices []struct{ ... } `json:"choices"`
    Usage   TokenUsage      `json:"usage"`
}
```

### 非流式调用

`callOpenAICompatible` 解析响应体时已包含 `usage`，直接从 `aiResp.Usage` 返回即可。

### 流式调用

请求体新增 `stream_options` 字段：

```go
type AIRequest struct {
    // ... 现有字段 ...
    StreamOptions *StreamOptions `json:"stream_options,omitempty"`
}

type StreamOptions struct {
    IncludeUsage bool `json:"include_usage"`
}
```

流式调用时设置 `StreamOptions: &StreamOptions{IncludeUsage: true}`。Provider 会在流结束前推送一个 `choices: []` + `usage: {...}` 的特殊 chunk，在 `streamOpenAICompatible` 中检测并捕获：

```go
// 检测 usage chunk：choices 为空且 usage 非零
if len(event.Choices) == 0 && event.Usage.TotalTokens > 0 {
    capturedUsage = event.Usage
}
```

### 函数签名变更

所有 `callAI*` / `StreamAI*` 函数返回值新增 `TokenUsage`：

```go
// 变更前
func callAIInternal(...) (content, model, providerID string, durationMs int, err error)

// 变更后
func callAIInternal(...) (content, model, providerID string, durationMs int, usage TokenUsage, err error)
```

涉及函数：`callAIInternal`、`callAI`、`callAIWithSystem`、`StreamAIWithSystem`、`StreamAIWithSystemNoThink`、`streamAIWithSystemEx`、`streamOpenAICompatible`、`callOpenAICompatibleWithLog`。

---

## 数据库层（`pkg/database/database.go`）

以增量迁移方式新增表：

```sql
CREATE TABLE IF NOT EXISTS token_usage_logs (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id           UUID REFERENCES users(id) ON DELETE SET NULL,
    chart_id          UUID REFERENCES bazi_charts(id) ON DELETE SET NULL,
    call_type         VARCHAR(50) NOT NULL,
    model             VARCHAR(100),
    provider_id       UUID REFERENCES llm_providers(id) ON DELETE SET NULL,
    prompt_tokens     INT NOT NULL DEFAULT 0,
    completion_tokens INT NOT NULL DEFAULT 0,
    total_tokens      INT NOT NULL DEFAULT 0,
    created_at        TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_token_usage_user_id    ON token_usage_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_token_usage_created_at ON token_usage_logs(created_at);
```

`call_type` 枚举值：`report` | `report_stream` | `liunian` | `dayun` | `celebrity` | `compatibility`

`user_id` 允许 NULL（名人生成、合盘等无登录用户场景）。

---

## Repository 层（`internal/repository/token_usage_repository.go`）

新文件，包含三个函数：

```go
// 写入一条调用记录
func CreateTokenUsageLog(userID *string, chartID *string, callType, model, providerID string, usage TokenUsage) error

// 按用户聚合，支持时间范围过滤
func GetTokenUsageSummary(from, to time.Time) ([]TokenUsageSummaryRow, error)
// TokenUsageSummaryRow: {UserID, Email, Name, RequestCount, PromptTokens, CompletionTokens, TotalTokens}

// 单用户明细，分页
func GetTokenUsageDetail(userID string, from, to time.Time, page, limit int) (total int, items []TokenUsageDetailRow, err error)
// TokenUsageDetailRow: {ID, CallType, Model, PromptTokens, CompletionTokens, TotalTokens, CreatedAt}
```

---

## 服务层写入点

AI 调用完成后，service 层负责异步写入 `token_usage_logs`（`go repository.CreateTokenUsageLog(...)`，失败仅打日志，不影响主流程）。

| 调用函数 | 所在 service 文件 | call_type | user_id 来源 |
|---|---|---|---|
| `callAIWithSystem` | `report_service.go` (原局报告) | `report` | `chart.UserID` |
| `StreamAIWithSystem` | `report_service.go` (流式报告) | `report_stream` | `chart.UserID` |
| `callAIWithSystem` | `report_service.go` (流年) | `liunian` | `chart.UserID` |
| `StreamAIWithSystemNoThink` | `report_service.go` (大运) | `dayun` | `chart.UserID` |
| `callAI` | `celebrity_service.go` | `celebrity` | `nil`（Admin 操作，无普通用户上下文）|
| `callAIWithSystem` | `compatibility_service.go` | `compatibility` | 服务层显式传入的 `userID` |

---

## Admin API（`internal/handler/token_usage_handler.go`）

注册到 `/api/admin/` 下，受 admin JWT 中间件保护。

### 汇总接口

```
GET /api/admin/token-usage/summary?from=2026-01-01&to=2026-05-31
```

`from` / `to` 默认当月第一天至今日，格式 `YYYY-MM-DD`。

响应：
```json
[
  {
    "user_id": "uuid",
    "email": "user@example.com",
    "name": "张三",
    "request_count": 12,
    "prompt_tokens": 45000,
    "completion_tokens": 18000,
    "total_tokens": 63000
  }
]
```

按 `total_tokens DESC` 排序。

### 明细接口

```
GET /api/admin/token-usage/detail?user_id=xxx&from=2026-01-01&to=2026-05-31&page=1&limit=20
```

响应：
```json
{
  "total": 48,
  "items": [
    {
      "id": "uuid",
      "call_type": "report_stream",
      "model": "deepseek-chat",
      "prompt_tokens": 3200,
      "completion_tokens": 1800,
      "total_tokens": 5000,
      "created_at": "2026-05-10T14:23:00Z"
    }
  ]
}
```

---

## Admin 前端（`frontend/src/pages/admin/TokenUsagePage.tsx`）

路由：`/admin/token-usage`，在 Admin 侧边栏加入入口。

### 布局

1. **筛选栏**：开始日期 + 结束日期输入框，默认当月，"查询"按钮
2. **汇总表格**：邮箱 | 用户名 | 请求次数 | 输入 tokens | 输出 tokens | 总 tokens | 操作（"明细"按钮）
3. **明细抽屉**：点击"明细"后展开，显示该用户分页调用记录（时间 | 类型 | 模型 | 输入 | 输出 | 总计），每页 20 条

样式复用项目现有 CSS 变量，不引入新 UI 框架。

---

## 关键约束

- token 写入失败不影响 AI 调用主流程（异步写入 + 仅打日志）
- `user_id` 为 nullable，未登录/无 chart 场景允许为 NULL
- 不做实时推送，Admin 页面按需查询
- 不实现用量限额，纯统计展示
