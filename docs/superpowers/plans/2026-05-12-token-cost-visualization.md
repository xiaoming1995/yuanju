# Token 成本可视化 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在 Admin token 用量统计页按模型单价估算每次 AI 调用的 CNY 费用，并在汇总与明细视图中展示。

**Architecture:** 查询时动态计算，不改 `token_usage_logs` 表结构。定价存于 `algo_config` 表（6 个 key），启动时 seed 默认值（ON CONFLICT DO NOTHING，Admin 可在"算法配置"页修改）。后端新增 `llm_pricing.go` 服务读取定价并提供 `CalcCost` 函数，repository 查询函数接受该函数作参数，在 Go 侧逐行计算费用并聚合。

**Tech Stack:** Go 1.21, Gin, lib/pq (PostgreSQL), React 19 + TypeScript

---

## 文件改动清单

| 文件 | 类型 | 说明 |
|------|------|------|
| `backend/pkg/seed/seed.go` | 修改 | 新增 `SeedLLMPrices()` |
| `backend/cmd/api/main.go` | 修改 | 调用 `SeedLLMPrices()` |
| `backend/internal/service/llm_pricing.go` | 新建 | `matchModelTier`, `GetModelPrice`, `CalcCost` |
| `backend/internal/service/llm_pricing_test.go` | 新建 | 单元测试 `matchModelTier` |
| `backend/internal/repository/token_usage_repository.go` | 修改 | struct 加字段，两个查询函数加 costFn 参数 |
| `backend/internal/handler/token_usage_handler.go` | 修改 | 传入 `service.CalcCost` |
| `frontend/src/pages/admin/TokenUsagePage.tsx` | 修改 | 加 cost 列和 interface 字段 |

---

### Task 1: Seed LLM 默认定价

**Files:**
- Modify: `backend/pkg/seed/seed.go`
- Modify: `backend/cmd/api/main.go`

- [ ] **Step 1: 在 `seed.go` 末尾追加 `SeedLLMPrices` 函数**

在 `backend/pkg/seed/seed.go` 末尾添加（在 `SeedLLMProviders` 函数之后）：

```go
// SeedLLMPrices 将默认模型定价写入 algo_config（ON CONFLICT DO NOTHING，不覆盖 Admin 已改的值）
func SeedLLMPrices() {
	prices := []struct {
		key, value, description string
	}{
		{"llm_price_flash_input", "0.27", "deepseek-v4-flash 输入单价（CNY/百万tokens）"},
		{"llm_price_flash_output", "1.10", "deepseek-v4-flash 输出单价（CNY/百万tokens）"},
		{"llm_price_pro_input", "4.00", "deepseek-v4-pro 输入单价（CNY/百万tokens）"},
		{"llm_price_pro_output", "16.00", "deepseek-v4-pro 输出单价（CNY/百万tokens）"},
		{"llm_price_default_input", "1.00", "未知模型 fallback 输入单价（CNY/百万tokens）"},
		{"llm_price_default_output", "2.00", "未知模型 fallback 输出单价（CNY/百万tokens）"},
	}
	for _, p := range prices {
		if _, err := database.DB.Exec(
			`INSERT INTO algo_config (key, value, description)
			 VALUES ($1, $2, $3)
			 ON CONFLICT (key) DO NOTHING`,
			p.key, p.value, p.description,
		); err != nil {
			log.Printf("[seed] LLM 定价 seed 失败: key=%s err=%v", p.key, err)
		}
	}
	log.Println("✅ 种子数据：LLM 定价配置已写入 algo_config（ON CONFLICT DO NOTHING）")
}
```

- [ ] **Step 2: 在 `main.go` 中调用 `SeedLLMPrices`**

在 `backend/cmd/api/main.go` 中，`seed.SeedLLMProviders()` 调用之后，紧接着添加：

```go
seed.SeedLLMPrices()
```

结果应如下：
```go
seed.SeedLLMProviders()
seed.SeedLLMPrices()

// 加载算法配置（含调候用神 seed）
if err := service.LoadAlgoConfig(); err != nil {
```

- [ ] **Step 3: 验证编译**

```bash
cd /Users/liujiming/web/yuanju/backend && go build ./...
```

Expected: 无错误输出

- [ ] **Step 4: Commit**

```bash
git add backend/pkg/seed/seed.go backend/cmd/api/main.go
git commit -m "feat(pricing): seed LLM 模型默认定价到 algo_config"
```

---

### Task 2: LLM Pricing 服务（新建）

**Files:**
- Create: `backend/internal/service/llm_pricing.go`
- Create: `backend/internal/service/llm_pricing_test.go`

- [ ] **Step 1: 写测试（先写测试）**

创建 `backend/internal/service/llm_pricing_test.go`：

```go
package service

import "testing"

func Test_matchModelTier(t *testing.T) {
	cases := []struct {
		model string
		want  string
	}{
		{"deepseek-v4-pro", "pro"},
		{"deepseek-v4-pro-20250101", "pro"},
		{"deepseek-reasoner", "pro"},
		{"deepseek-v4-flash", "flash"},
		{"deepseek-chat", "flash"},
		{"DEEPSEEK-V4-FLASH", "flash"},
		{"gpt-4o", "default"},
		{"qwen3-32b", "default"},
		{"", "default"},
	}
	for _, c := range cases {
		got := matchModelTier(c.model)
		if got != c.want {
			t.Errorf("matchModelTier(%q) = %q, want %q", c.model, got, c.want)
		}
	}
}
```

- [ ] **Step 2: 运行测试，确认失败（函数未定义）**

```bash
cd /Users/liujiming/web/yuanju/backend && go test ./internal/service/ -run Test_matchModelTier -v
```

Expected: `FAIL` — `undefined: matchModelTier`

- [ ] **Step 3: 创建 `llm_pricing.go`**

创建 `backend/internal/service/llm_pricing.go`：

```go
package service

import (
	"strconv"
	"strings"
	"yuanju/internal/repository"
)

// matchModelTier 将模型名映射到定价档位：pro / flash / default
func matchModelTier(modelName string) string {
	lower := strings.ToLower(modelName)
	if strings.Contains(lower, "v4-pro") || strings.Contains(lower, "reasoner") {
		return "pro"
	}
	if strings.Contains(lower, "v4-flash") || strings.Contains(lower, "chat") {
		return "flash"
	}
	return "default"
}

// GetModelPrice 返回指定模型的 (inputPriceCNY, outputPriceCNY)，单位：CNY / 百万 tokens。
// 从 algo_config 表实时读取；读取失败时使用硬编码兜底值。
func GetModelPrice(modelName string) (inputPrice, outputPrice float64) {
	rows, err := repository.GetAllAlgoConfig()
	if err != nil {
		return 1.0, 2.0
	}
	prices := make(map[string]string, 6)
	for _, r := range rows {
		if strings.HasPrefix(r.Key, "llm_price_") {
			prices[r.Key] = r.Value
		}
	}
	tier := matchModelTier(modelName)
	inputPrice = parsePrice(prices["llm_price_"+tier+"_input"], 1.0)
	outputPrice = parsePrice(prices["llm_price_"+tier+"_output"], 2.0)
	return inputPrice, outputPrice
}

// CalcCost 根据 token 数量和模型名返回预估费用（CNY）。
func CalcCost(modelName string, promptTokens, completionTokens int) float64 {
	inputPrice, outputPrice := GetModelPrice(modelName)
	return float64(promptTokens)/1_000_000*inputPrice +
		float64(completionTokens)/1_000_000*outputPrice
}

func parsePrice(s string, fallback float64) float64 {
	if s == "" {
		return fallback
	}
	v, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	if err != nil {
		return fallback
	}
	return v
}
```

- [ ] **Step 4: 运行测试，确认通过**

```bash
cd /Users/liujiming/web/yuanju/backend && go test ./internal/service/ -run Test_matchModelTier -v
```

Expected:
```
=== RUN   Test_matchModelTier
--- PASS: Test_matchModelTier (0.00s)
PASS
```

- [ ] **Step 5: Commit**

```bash
git add backend/internal/service/llm_pricing.go backend/internal/service/llm_pricing_test.go
git commit -m "feat(pricing): LLM 定价服务 GetModelPrice / CalcCost"
```

---

### Task 3: Repository — 汇总查询加费用计算

**Files:**
- Modify: `backend/internal/repository/token_usage_repository.go`

- [ ] **Step 1: `TokenUsageSummaryRow` 新增 `EstimatedCostCny` 字段**

在 `token_usage_repository.go` 的 `TokenUsageSummaryRow` struct 中新增最后一个字段：

```go
type TokenUsageSummaryRow struct {
	UserID           string  `json:"user_id"`
	Email            string  `json:"email"`
	Nickname         string  `json:"nickname"`
	RequestCount     int     `json:"request_count"`
	PromptTokens     int     `json:"prompt_tokens"`
	CompletionTokens int     `json:"completion_tokens"`
	TotalTokens      int     `json:"total_tokens"`
	EstimatedCostCny float64 `json:"estimated_cost_cny"`
}
```

- [ ] **Step 2: 新增内部辅助类型 `summaryByModel`**

在 `GetTokenUsageSummary` 函数之前插入（unexported，仅此文件内部使用）：

```go
type summaryByModel struct {
	userID, email, nickname        string
	model                          string
	requestCount                   int
	promptTokens, completionTokens, totalTokens int
}
```

- [ ] **Step 3: 更新 `GetTokenUsageSummary` 函数签名与实现**

将整个 `GetTokenUsageSummary` 函数替换为：

```go
// GetTokenUsageSummary 按用户聚合 token 消耗，from/to 均为日期（含）。
// costFn(model, promptTokens, completionTokens) 用于计算预估费用；传 nil 则费用为 0。
func GetTokenUsageSummary(from, to time.Time, costFn func(string, int, int) float64) ([]TokenUsageSummaryRow, error) {
	toExcl := to.AddDate(0, 0, 1)
	rows, err := database.DB.Query(`
		SELECT
			u.id,
			u.email,
			COALESCE(u.nickname, '') AS nickname,
			COUNT(t.id)::int                           AS request_count,
			COALESCE(t.model, '')                      AS model,
			COALESCE(SUM(t.prompt_tokens), 0)::int     AS prompt_tokens,
			COALESCE(SUM(t.completion_tokens), 0)::int AS completion_tokens,
			COALESCE(SUM(t.total_tokens), 0)::int      AS total_tokens
		FROM users u
		JOIN token_usage_logs t ON t.user_id = u.id
		WHERE t.created_at >= $1 AND t.created_at < $2
		GROUP BY u.id, u.email, u.nickname, t.model
		ORDER BY u.id`,
		from, toExcl,
	)
	if err != nil {
		return nil, fmt.Errorf("GetTokenUsageSummary: %w", err)
	}
	defer rows.Close()

	// 先按 (user, model) 扫描
	var byModel []summaryByModel
	for rows.Next() {
		var r summaryByModel
		if err := rows.Scan(&r.userID, &r.email, &r.nickname, &r.requestCount,
			&r.model, &r.promptTokens, &r.completionTokens, &r.totalTokens); err != nil {
			log.Printf("[TokenUsage] Scan 失败: %v", err)
			continue
		}
		byModel = append(byModel, r)
	}

	// 在 Go 侧按 user_id 聚合，顺便累加费用
	type entry struct {
		row  TokenUsageSummaryRow
		cost float64
	}
	userMap := make(map[string]*entry)
	var userOrder []string

	for _, r := range byModel {
		e, exists := userMap[r.userID]
		if !exists {
			e = &entry{row: TokenUsageSummaryRow{
				UserID:   r.userID,
				Email:    r.email,
				Nickname: r.nickname,
			}}
			userMap[r.userID] = e
			userOrder = append(userOrder, r.userID)
		}
		e.row.RequestCount += r.requestCount
		e.row.PromptTokens += r.promptTokens
		e.row.CompletionTokens += r.completionTokens
		e.row.TotalTokens += r.totalTokens
		if costFn != nil {
			e.cost += costFn(r.model, r.promptTokens, r.completionTokens)
		}
	}

	result := make([]TokenUsageSummaryRow, 0, len(userOrder))
	for _, uid := range userOrder {
		e := userMap[uid]
		e.row.EstimatedCostCny = e.cost
		result = append(result, e.row)
	}

	// 按 total_tokens 降序排列
	sort.Slice(result, func(i, j int) bool {
		return result[i].TotalTokens > result[j].TotalTokens
	})

	return result, nil
}
```

- [ ] **Step 4: 在文件顶部 import 中新增 `sort`**

`token_usage_repository.go` 的 import 块改为：

```go
import (
	"fmt"
	"log"
	"sort"
	"time"
	"yuanju/pkg/database"
)
```

- [ ] **Step 5: 验证编译（会报 handler 错误，属预期）**

```bash
cd /Users/liujiming/web/yuanju/backend && go build ./internal/repository/... 2>&1
```

Expected: 无错误（repository 包本身编译通过；handler 尚未更新会在下一步修复）

- [ ] **Step 6: Commit**

```bash
git add backend/internal/repository/token_usage_repository.go
git commit -m "feat(pricing): GetTokenUsageSummary 按模型聚合并计算费用"
```

---

### Task 4: Repository — 明细查询加费用计算

**Files:**
- Modify: `backend/internal/repository/token_usage_repository.go`

- [ ] **Step 1: `TokenUsageDetailRow` 新增 `EstimatedCostCny` 字段**

```go
type TokenUsageDetailRow struct {
	ID               string    `json:"id"`
	CallType         string    `json:"call_type"`
	Model            string    `json:"model"`
	PromptTokens     int       `json:"prompt_tokens"`
	CompletionTokens int       `json:"completion_tokens"`
	TotalTokens      int       `json:"total_tokens"`
	ReasoningTokens  int       `json:"reasoning_tokens"`
	CacheHitTokens   int       `json:"cache_hit_tokens"`
	CacheMissTokens  int       `json:"cache_miss_tokens"`
	EstimatedCostCny float64   `json:"estimated_cost_cny"`
	CreatedAt        time.Time `json:"created_at"`
}
```

- [ ] **Step 2: 更新 `GetTokenUsageDetail` 函数签名与实现**

将整个 `GetTokenUsageDetail` 函数替换为（新增 `costFn` 参数，在 scan 后逐行计算费用）：

```go
// GetTokenUsageDetail 查询单用户分页明细。
// costFn(model, promptTokens, completionTokens) 用于计算预估费用；传 nil 则费用为 0。
func GetTokenUsageDetail(userID string, from, to time.Time, page, limit int, costFn func(string, int, int) float64) (total int, items []TokenUsageDetailRow, err error) {
	toExcl := to.AddDate(0, 0, 1)
	offset := (page - 1) * limit

	if err = database.DB.QueryRow(`
		SELECT COUNT(*) FROM token_usage_logs
		WHERE user_id = $1 AND created_at >= $2 AND created_at < $3`,
		userID, from, toExcl,
	).Scan(&total); err != nil {
		return 0, nil, fmt.Errorf("GetTokenUsageDetail count: %w", err)
	}

	rows, err := database.DB.Query(`
		SELECT id, call_type, COALESCE(model, ''), prompt_tokens, completion_tokens, total_tokens,
		       reasoning_tokens, cache_hit_tokens, cache_miss_tokens, created_at
		FROM token_usage_logs
		WHERE user_id = $1 AND created_at >= $2 AND created_at < $3
		ORDER BY created_at DESC
		LIMIT $4 OFFSET $5`,
		userID, from, toExcl, limit, offset,
	)
	if err != nil {
		return 0, nil, fmt.Errorf("GetTokenUsageDetail query: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var r TokenUsageDetailRow
		if err := rows.Scan(&r.ID, &r.CallType, &r.Model,
			&r.PromptTokens, &r.CompletionTokens, &r.TotalTokens,
			&r.ReasoningTokens, &r.CacheHitTokens, &r.CacheMissTokens, &r.CreatedAt); err != nil {
			log.Printf("[TokenUsage] Scan detail 失败: %v", err)
			continue
		}
		if costFn != nil {
			r.EstimatedCostCny = costFn(r.Model, r.PromptTokens, r.CompletionTokens)
		}
		items = append(items, r)
	}
	return total, items, nil
}
```

- [ ] **Step 3: 编译验证**

```bash
cd /Users/liujiming/web/yuanju/backend && go build ./internal/repository/... 2>&1
```

Expected: 无错误

- [ ] **Step 4: Commit**

```bash
git add backend/internal/repository/token_usage_repository.go
git commit -m "feat(pricing): GetTokenUsageDetail 每行计算预估费用"
```

---

### Task 5: Handler 传入 CalcCost

**Files:**
- Modify: `backend/internal/handler/token_usage_handler.go`

- [ ] **Step 1: 更新 import，加入 service**

将 `token_usage_handler.go` 的 import 块改为：

```go
import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"yuanju/internal/repository"
	"yuanju/internal/service"
)
```

- [ ] **Step 2: 更新 `AdminGetTokenUsageSummary`**

将函数替换为：

```go
// AdminGetTokenUsageSummary GET /api/admin/token-usage/summary?from=YYYY-MM-DD&to=YYYY-MM-DD
func AdminGetTokenUsageSummary(c *gin.Context) {
	from, to := parseDateRange(c)
	rows, err := repository.GetTokenUsageSummary(from, to, service.CalcCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if rows == nil {
		rows = []repository.TokenUsageSummaryRow{}
	}
	c.JSON(http.StatusOK, rows)
}
```

- [ ] **Step 3: 更新 `AdminGetTokenUsageDetail`**

将函数中的 `repository.GetTokenUsageDetail` 调用改为：

```go
total, items, err := repository.GetTokenUsageDetail(userID, from, to, page, limit, service.CalcCost)
```

- [ ] **Step 4: 全量编译验证**

```bash
cd /Users/liujiming/web/yuanju/backend && go build ./... 2>&1
```

Expected: 无错误输出

- [ ] **Step 5: 运行全部测试**

```bash
cd /Users/liujiming/web/yuanju/backend && go test ./... 2>&1
```

Expected: 所有测试 PASS（包含 `Test_matchModelTier`）

- [ ] **Step 6: Commit**

```bash
git add backend/internal/handler/token_usage_handler.go
git commit -m "feat(pricing): handler 传入 CalcCost，接口返回预估费用"
```

---

### Task 6: 前端 — 加费用列

**Files:**
- Modify: `frontend/src/pages/admin/TokenUsagePage.tsx`

- [ ] **Step 1: 更新 `SummaryRow` interface**

将 `SummaryRow` 改为：

```tsx
interface SummaryRow {
  user_id: string
  email: string
  nickname: string
  request_count: number
  prompt_tokens: number
  completion_tokens: number
  total_tokens: number
  estimated_cost_cny: number
}
```

- [ ] **Step 2: 更新 `DetailRow` interface**

将 `DetailRow` 改为：

```tsx
interface DetailRow {
  id: string
  call_type: string
  model: string
  prompt_tokens: number
  completion_tokens: number
  total_tokens: number
  reasoning_tokens: number
  cache_hit_tokens: number
  cache_miss_tokens: number
  estimated_cost_cny: number
  created_at: string
}
```

- [ ] **Step 3: 汇总表 — thead 加"预估费用"列**

找到：
```tsx
<th style={{ textAlign: 'right' }}>总 tokens</th>
<th>操作</th>
```

替换为：
```tsx
<th style={{ textAlign: 'right' }}>总 tokens</th>
<th style={{ textAlign: 'right' }}>预估费用</th>
<th>操作</th>
```

- [ ] **Step 4: 汇总表 — tbody 加费用单元格**

找到：
```tsx
<td style={{ textAlign: 'right', fontWeight: 700, color: '#a78bfa' }}>{fmt(row.total_tokens)}</td>
<td>
  <button className="admin-btn" style={{ padding: '4px 12px', fontSize: 13 }} onClick={() => openDetail(row)}>
```

替换为：
```tsx
<td style={{ textAlign: 'right', fontWeight: 700, color: '#a78bfa' }}>{fmt(row.total_tokens)}</td>
<td style={{ textAlign: 'right', color: '#f59e0b', fontWeight: 700 }}>
  ¥ {row.estimated_cost_cny.toFixed(4)}
</td>
<td>
  <button className="admin-btn" style={{ padding: '4px 12px', fontSize: 13 }} onClick={() => openDetail(row)}>
```

- [ ] **Step 5: 汇总表底部加说明文字**

找到汇总表 `</table>` 关闭标签之后（在 `</div>` 前）插入：

```tsx
<div style={{ fontSize: 12, color: '#555', marginTop: 8, padding: '0 4px' }}>
  * 基于当前 algo_config 单价估算，仅供参考
</div>
```

具体位置：在 `summary.length === 0` 分支的表格闭合 `</table>` 后，`)}` 前。

- [ ] **Step 6: 明细抽屉 — thead 加"费用"列**

找到：
```tsx
<th style={{ textAlign: 'right' }}>缓存命中</th>
<th style={{ textAlign: 'right' }}>总计</th>
```

替换为：
```tsx
<th style={{ textAlign: 'right' }}>缓存命中</th>
<th style={{ textAlign: 'right' }}>费用</th>
<th style={{ textAlign: 'right' }}>总计</th>
```

- [ ] **Step 7: 明细抽屉 — tbody 加费用单元格**

找到：
```tsx
<td style={{ textAlign: 'right', fontWeight: 700, color: '#a78bfa' }}>{fmt(item.total_tokens)}</td>
```

在其**前面**插入：
```tsx
<td style={{ textAlign: 'right', color: '#f59e0b', fontSize: 12 }}>
  ¥ {item.estimated_cost_cny.toFixed(4)}
</td>
```

- [ ] **Step 8: 前端编译验证**

```bash
cd /Users/liujiming/web/yuanju/frontend && npm run build 2>&1 | tail -20
```

Expected: `✓ built in ...` 无 TypeScript 错误

- [ ] **Step 9: Commit**

```bash
git add frontend/src/pages/admin/TokenUsagePage.tsx
git commit -m "feat(pricing): 前端 token 用量统计加预估费用列"
```

---

### Task 7: 部署验证

**Files:** 无代码改动，只运行命令

- [ ] **Step 1: Docker 重建 backend + frontend**

```bash
cd /Users/liujiming/web/yuanju && docker-compose up --build -d backend frontend
```

Expected: 两个容器 Started/Recreated，无 error

- [ ] **Step 2: 检查 seed 日志**

```bash
docker-compose logs backend 2>&1 | grep -E "seed|LLM 定价|llm_price"
```

Expected: 包含 `✅ 种子数据：LLM 定价配置已写入 algo_config`

- [ ] **Step 3: 在 Admin"算法配置"页验证**

打开 `/admin` → 算法配置，确认出现以下 6 个 key：
- `llm_price_flash_input` = 0.27
- `llm_price_flash_output` = 1.10
- `llm_price_pro_input` = 4.00
- `llm_price_pro_output` = 16.00
- `llm_price_default_input` = 1.00
- `llm_price_default_output` = 2.00

- [ ] **Step 4: 在 Admin Token 用量统计页验证**

打开 `/admin` → Token 用量统计，选择时间范围后点击"查询"：
- 汇总表出现"预估费用"列，显示 `¥ 0.xxxx` 格式金额
- 表格底部有 `* 基于当前 algo_config 单价估算，仅供参考`
- 点击"明细"，抽屉表格出现"费用"列

- [ ] **Step 5: 手动验算一条记录**

取一条 deepseek-v4-flash 的明细记录：
- 假设 prompt_tokens=10000, completion_tokens=2000
- 预期费用 = (10000/1000000×0.27) + (2000/1000000×1.10) = 0.0027 + 0.0022 = **¥ 0.0049**
- 与页面显示的值比对，误差在 0.0001 以内则通过
