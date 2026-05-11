# Token 用量按模型拆分展示 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将 token 用量汇总表从每用户一行改为按 (用户 × 模型) 拆分多行，并在明细抽屉加模型筛选 tabs。

**Architecture:** 后端 `GetTokenUsageSummary` 改为返回平铺的 per-(user,model) 数组，Go 侧按用户总 token 排序后展开；`GetTokenUsageDetail` 加可选 model 过滤参数。前端按 user_id 分组渲染：第一行显示邮箱，后续行空白，每组末尾插前端计算的合计行。

**Tech Stack:** Go 1.21, Gin, lib/pq (PostgreSQL), React 19 + TypeScript

---

## 文件改动清单

| 文件 | 改动类型 |
|------|---------|
| `backend/internal/repository/token_usage_repository.go` | 改 struct（加 Model 字段，删 summaryByModel）、重写 GetTokenUsageSummary、改 GetTokenUsageDetail 签名 |
| `backend/internal/handler/token_usage_handler.go` | AdminGetTokenUsageDetail 读 model query param |
| `frontend/src/lib/adminApi.ts` | adminTokenUsageAPI.detail 加 model 参数 |
| `frontend/src/pages/admin/TokenUsagePage.tsx` | 全面重写汇总表渲染 + 加 drawerModel state + 明细 tabs |

---

### Task 1: Repository — GetTokenUsageSummary 返回平铺 per-(user,model) 数组

**Files:**
- Modify: `backend/internal/repository/token_usage_repository.go`

- [ ] **Step 1: 更新 `TokenUsageSummaryRow` struct，新增 `Model` 字段**

将文件中的 `TokenUsageSummaryRow` 替换为：

```go
type TokenUsageSummaryRow struct {
	UserID           string  `json:"user_id"`
	Email            string  `json:"email"`
	Nickname         string  `json:"nickname"`
	Model            string  `json:"model"`
	RequestCount     int     `json:"request_count"`
	PromptTokens     int     `json:"prompt_tokens"`
	CompletionTokens int     `json:"completion_tokens"`
	TotalTokens      int     `json:"total_tokens"`
	EstimatedCostCny float64 `json:"estimated_cost_cny"`
}
```

- [ ] **Step 2: 删除 `summaryByModel` 辅助类型**

删除以下整个类型定义（现在直接 scan 进 `TokenUsageSummaryRow`）：

```go
type summaryByModel struct {
	userID, email, nickname             string
	model                               string
	requestCount                        int
	promptTokens, completionTokens, totalTokens int
}
```

- [ ] **Step 3: 重写 `GetTokenUsageSummary` 函数**

将整个 `GetTokenUsageSummary` 函数替换为：

```go
// GetTokenUsageSummary 返回平铺的 per-(user,model) 数组，按用户总 token 降序排列用户，
// 同一用户内按模型 token 降序排列。costFn 传 nil 则费用为 0。
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

	// 直接 scan 进带 Model 字段的 row，同时计算费用
	var byModel []TokenUsageSummaryRow
	for rows.Next() {
		var r TokenUsageSummaryRow
		if err := rows.Scan(&r.UserID, &r.Email, &r.Nickname, &r.RequestCount,
			&r.Model, &r.PromptTokens, &r.CompletionTokens, &r.TotalTokens); err != nil {
			log.Printf("[TokenUsage] Scan 失败: %v", err)
			continue
		}
		if costFn != nil {
			r.EstimatedCostCny = costFn(r.Model, r.PromptTokens, r.CompletionTokens)
		}
		byModel = append(byModel, r)
	}

	// 按 user_id 分组，记录每个用户的总 token 用于排序
	type userMeta struct {
		totalTokens int
		rows        []TokenUsageSummaryRow
	}
	userMap := make(map[string]*userMeta)
	var userOrder []string
	for _, r := range byModel {
		if _, exists := userMap[r.UserID]; !exists {
			userMap[r.UserID] = &userMeta{}
			userOrder = append(userOrder, r.UserID)
		}
		userMap[r.UserID].totalTokens += r.TotalTokens
		userMap[r.UserID].rows = append(userMap[r.UserID].rows, r)
	}

	// 用户按总 token 降序排列
	sort.Slice(userOrder, func(i, j int) bool {
		return userMap[userOrder[i]].totalTokens > userMap[userOrder[j]].totalTokens
	})

	// 展开：每个用户内按模型 token 降序排列
	var result []TokenUsageSummaryRow
	for _, uid := range userOrder {
		meta := userMap[uid]
		sort.Slice(meta.rows, func(i, j int) bool {
			return meta.rows[i].TotalTokens > meta.rows[j].TotalTokens
		})
		result = append(result, meta.rows...)
	}
	return result, nil
}
```

- [ ] **Step 4: 编译验证**

```bash
cd /Users/liujiming/web/yuanju/backend && go build ./internal/repository/... 2>&1
```

Expected: 无错误（handler 会报错，属预期，下一个 task 修复）

- [ ] **Step 5: Commit**

```bash
git add backend/internal/repository/token_usage_repository.go
git commit -m "feat(breakdown): GetTokenUsageSummary 返回平铺 per-(user,model) 数组"
```

---

### Task 2: Repository — GetTokenUsageDetail 加 model 过滤参数

**Files:**
- Modify: `backend/internal/repository/token_usage_repository.go`

- [ ] **Step 1: 更新 `GetTokenUsageDetail` 函数签名与 SQL**

将整个 `GetTokenUsageDetail` 函数替换为（新增 `model string` 参数，两个 SQL 均加 model 过滤条件）：

```go
// GetTokenUsageDetail 查询单用户分页明细。
// model 传空字符串则不过滤模型；costFn 传 nil 则费用为 0。
func GetTokenUsageDetail(userID string, from, to time.Time, page, limit int, model string, costFn func(string, int, int) float64) (total int, items []TokenUsageDetailRow, err error) {
	toExcl := to.AddDate(0, 0, 1)
	offset := (page - 1) * limit

	if err = database.DB.QueryRow(`
		SELECT COUNT(*) FROM token_usage_logs
		WHERE user_id = $1 AND created_at >= $2 AND created_at < $3
		AND ($4 = '' OR COALESCE(model, '') = $4)`,
		userID, from, toExcl, model,
	).Scan(&total); err != nil {
		return 0, nil, fmt.Errorf("GetTokenUsageDetail count: %w", err)
	}

	rows, err := database.DB.Query(`
		SELECT id, call_type, COALESCE(model, ''), prompt_tokens, completion_tokens, total_tokens,
		       reasoning_tokens, cache_hit_tokens, cache_miss_tokens, created_at
		FROM token_usage_logs
		WHERE user_id = $1 AND created_at >= $2 AND created_at < $3
		AND ($4 = '' OR COALESCE(model, '') = $4)
		ORDER BY created_at DESC
		LIMIT $5 OFFSET $6`,
		userID, from, toExcl, model, limit, offset,
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

- [ ] **Step 2: 编译验证（repository 层）**

```bash
cd /Users/liujiming/web/yuanju/backend && go build ./internal/repository/... 2>&1
```

Expected: 无错误

- [ ] **Step 3: Commit**

```bash
git add backend/internal/repository/token_usage_repository.go
git commit -m "feat(breakdown): GetTokenUsageDetail 加 model 过滤参数"
```

---

### Task 3: Handler — 读取 model query param

**Files:**
- Modify: `backend/internal/handler/token_usage_handler.go`

- [ ] **Step 1: 更新 `AdminGetTokenUsageDetail`，读取 model 参数并传入 repository**

将 `AdminGetTokenUsageDetail` 函数替换为：

```go
// AdminGetTokenUsageDetail GET /api/admin/token-usage/detail?user_id=xxx&from=&to=&page=1&limit=20&model=
func AdminGetTokenUsageDetail(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id 必填"})
		return
	}
	from, to := parseDateRange(c)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	model := c.DefaultQuery("model", "")

	total, items, err := repository.GetTokenUsageDetail(userID, from, to, page, limit, model, service.CalcCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if items == nil {
		items = []repository.TokenUsageDetailRow{}
	}
	c.JSON(http.StatusOK, gin.H{
		"total": total,
		"items": items,
	})
}
```

- [ ] **Step 2: 全量编译验证**

```bash
cd /Users/liujiming/web/yuanju/backend && go build ./... 2>&1
```

Expected: 无任何错误输出

- [ ] **Step 3: 运行测试**

```bash
cd /Users/liujiming/web/yuanju/backend && go test ./... 2>&1
```

Expected: 所有测试 PASS

- [ ] **Step 4: Commit**

```bash
git add backend/internal/handler/token_usage_handler.go
git commit -m "feat(breakdown): AdminGetTokenUsageDetail 支持 model 过滤参数"
```

---

### Task 4: API Client — detail 加 model 参数

**Files:**
- Modify: `frontend/src/lib/adminApi.ts`

- [ ] **Step 1: 更新 `adminTokenUsageAPI.detail` 函数签名**

找到 `adminTokenUsageAPI` 中的 `detail` 行：

```ts
  detail: (userId: string, from: string, to: string, page: number, limit = 20) =>
    adminApi.get(`/api/admin/token-usage/detail?user_id=${userId}&from=${from}&to=${to}&page=${page}&limit=${limit}`),
```

替换为：

```ts
  detail: (userId: string, from: string, to: string, page: number, limit = 20, model = '') =>
    adminApi.get(`/api/admin/token-usage/detail?user_id=${userId}&from=${from}&to=${to}&page=${page}&limit=${limit}${model ? `&model=${encodeURIComponent(model)}` : ''}`),
```

- [ ] **Step 2: Commit**

```bash
git add frontend/src/lib/adminApi.ts
git commit -m "feat(breakdown): adminTokenUsageAPI.detail 加 model 过滤参数"
```

---

### Task 5: 前端 — 分组渲染 + 模型筛选 tabs

**Files:**
- Modify: `frontend/src/pages/admin/TokenUsagePage.tsx`

- [ ] **Step 1: 更新 import，加入 `Fragment`，并更新 `SummaryRow` interface**

将文件顶部 import 行改为：

```tsx
import { useState, Fragment } from 'react'
```

将 `SummaryRow` 替换为：

```tsx
interface SummaryRow {
  user_id: string
  email: string
  nickname: string
  model: string
  request_count: number
  prompt_tokens: number
  completion_tokens: number
  total_tokens: number
  estimated_cost_cny: number
}
```

- [ ] **Step 2: 新增 `drawerModel` state，更新 `openDetail` 和 `closeDrawer`**

在 `detailLoading` state 下方新增：

```tsx
const [drawerModel, setDrawerModel] = useState('')
```

将 `openDetail` 函数替换为：

```tsx
const openDetail = async (row: SummaryRow, page = 1, model = '') => {
  setDrawerUser(row)
  setDrawerModel(model)
  setDetailPage(page)
  setDetailLoading(true)
  try {
    const res = await adminTokenUsageAPI.detail(row.user_id, from, to, page, detailLimit, model)
    setDetail(res.data)
  } finally {
    setDetailLoading(false)
  }
}
```

将 `closeDrawer` 函数替换为：

```tsx
const closeDrawer = () => {
  setDrawerUser(null)
  setDetail(null)
  setDrawerModel('')
}
```

- [ ] **Step 3: 重写汇总表 thead（新增"模型"列）**

将 `<thead>` 内容替换为：

```tsx
<thead>
  <tr>
    <th>用户邮箱</th>
    <th>昵称</th>
    <th>模型</th>
    <th style={{ textAlign: 'right' }}>请求次数</th>
    <th style={{ textAlign: 'right' }}>输入 tokens</th>
    <th style={{ textAlign: 'right' }}>输出 tokens</th>
    <th style={{ textAlign: 'right' }}>总 tokens</th>
    <th style={{ textAlign: 'right' }}>预估费用</th>
    <th>操作</th>
  </tr>
</thead>
```

- [ ] **Step 4: 重写汇总表 tbody（分组渲染 + 合计行）**

将 `<tbody>` 内容替换为：

```tsx
<tbody>
  {(() => {
    // 按 user_id 分组（后端保证同一用户连续）
    const groups: { userID: string; rows: SummaryRow[] }[] = []
    for (const row of summary) {
      const last = groups[groups.length - 1]
      if (last && last.userID === row.user_id) {
        last.rows.push(row)
      } else {
        groups.push({ userID: row.user_id, rows: [row] })
      }
    }

    return groups.map((group, gi) => {
      const first = group.rows[0]
      const sub = group.rows.reduce(
        (acc, r) => ({
          requestCount: acc.requestCount + r.request_count,
          promptTokens: acc.promptTokens + r.prompt_tokens,
          completionTokens: acc.completionTokens + r.completion_tokens,
          totalTokens: acc.totalTokens + r.total_tokens,
          cost: acc.cost + r.estimated_cost_cny,
        }),
        { requestCount: 0, promptTokens: 0, completionTokens: 0, totalTokens: 0, cost: 0 }
      )

      return (
        <Fragment key={group.userID}>
          {gi > 0 && (
            <tr>
              <td colSpan={9} style={{ height: 4, background: '#0d0d1a', padding: 0 }} />
            </tr>
          )}
          {group.rows.map((row, ri) => (
            <tr key={`${row.user_id}-${row.model}`}>
              <td style={{ fontWeight: 600, color: '#e8e8e8' }}>
                {ri === 0 ? row.email : ''}
              </td>
              <td style={{ color: '#aaa' }}>
                {ri === 0 ? (row.nickname || '—') : ''}
              </td>
              <td style={{ fontSize: 12, color: '#888' }}>{row.model}</td>
              <td style={{ textAlign: 'right' }}>{fmt(row.request_count)}</td>
              <td style={{ textAlign: 'right' }}>{fmt(row.prompt_tokens)}</td>
              <td style={{ textAlign: 'right' }}>{fmt(row.completion_tokens)}</td>
              <td style={{ textAlign: 'right', color: '#a78bfa' }}>{fmt(row.total_tokens)}</td>
              <td style={{ textAlign: 'right', color: '#f59e0b' }}>
                ¥ {row.estimated_cost_cny.toFixed(4)}
              </td>
              <td></td>
            </tr>
          ))}
          <tr style={{ background: '#1e1e40' }}>
            <td></td>
            <td></td>
            <td style={{ fontWeight: 700, color: '#888', fontSize: 12 }}>合计</td>
            <td style={{ textAlign: 'right', fontWeight: 700 }}>{fmt(sub.requestCount)}</td>
            <td style={{ textAlign: 'right', fontWeight: 700 }}>{fmt(sub.promptTokens)}</td>
            <td style={{ textAlign: 'right', fontWeight: 700 }}>{fmt(sub.completionTokens)}</td>
            <td style={{ textAlign: 'right', fontWeight: 700, color: '#a78bfa' }}>{fmt(sub.totalTokens)}</td>
            <td style={{ textAlign: 'right', fontWeight: 700, color: '#f59e0b' }}>
              ¥ {sub.cost.toFixed(4)}
            </td>
            <td>
              <button
                className="admin-btn"
                style={{ padding: '4px 12px', fontSize: 13 }}
                onClick={() => openDetail(first)}
              >
                明细
              </button>
            </td>
          </tr>
        </Fragment>
      )
    })
  })()}
</tbody>
```

- [ ] **Step 5: 在明细抽屉标题下方加模型筛选 tabs**

找到：

```tsx
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 20 }}>
              <h2 style={{ fontSize: 16, fontWeight: 700, color: '#e0e0e0' }}>
                {drawerUser.email} 的调用明细
              </h2>
              <button className="admin-btn" onClick={closeDrawer} style={{ padding: '4px 12px' }}>关闭</button>
            </div>
```

替换为：

```tsx
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 12 }}>
              <h2 style={{ fontSize: 16, fontWeight: 700, color: '#e0e0e0' }}>
                {drawerUser.email} 的调用明细
              </h2>
              <button className="admin-btn" onClick={closeDrawer} style={{ padding: '4px 12px' }}>关闭</button>
            </div>
            {/* 模型筛选 tabs */}
            {(() => {
              const userModels = summary
                .filter(r => r.user_id === drawerUser.user_id)
                .map(r => r.model)
              return (
                <div style={{ display: 'flex', gap: 6, flexWrap: 'wrap', marginBottom: 16 }}>
                  {['', ...userModels].map(m => (
                    <button
                      key={m}
                      onClick={() => { setDrawerModel(m); openDetail(drawerUser, 1, m) }}
                      style={{
                        background: drawerModel === m ? '#a78bfa' : 'transparent',
                        color: drawerModel === m ? '#fff' : '#888',
                        border: `1px solid ${drawerModel === m ? '#a78bfa' : '#333'}`,
                        borderRadius: 4, padding: '4px 12px', fontSize: 12, cursor: 'pointer',
                      }}
                    >
                      {m === '' ? '全部' : m}
                    </button>
                  ))}
                </div>
              )
            })()}
```

- [ ] **Step 6: 更新明细抽屉分页按钮，传入 drawerModel**

将分页区两个按钮的 `onClick` 更新：

```tsx
onClick={() => openDetail(drawerUser, detailPage - 1, drawerModel)}
// ...
onClick={() => openDetail(drawerUser, detailPage + 1, drawerModel)}
```

- [ ] **Step 7: 前端编译验证**

```bash
cd /Users/liujiming/web/yuanju/frontend && npm run build 2>&1 | tail -10
```

Expected: `✓ built in ...` 无 TypeScript 错误

- [ ] **Step 8: Commit**

```bash
git add frontend/src/pages/admin/TokenUsagePage.tsx
git commit -m "feat(breakdown): 汇总表按模型分组渲染 + 明细抽屉加模型筛选 tabs"
```

---

### Task 6: 部署验证

**Files:** 无代码改动

- [ ] **Step 1: Docker 重建**

```bash
cd /Users/liujiming/web/yuanju && docker-compose up --build -d backend frontend 2>&1 | tail -10
```

Expected: `Started` 或 `Recreated`，无 error

- [ ] **Step 2: Admin Token 用量统计页验证**

打开 `/admin` → Token 用量统计，选择时间范围后点击"查询"：

1. 汇总表每个用户显示多行（每个模型一行）
2. 每组第一行显示邮箱，后续行邮箱为空
3. 每组末尾有深色背景的"合计"行，含"明细"按钮
4. 用户组之间有细分隔线

- [ ] **Step 3: 明细抽屉验证**

点击某用户合计行的"明细"按钮：

1. 抽屉顶部出现模型筛选 tabs（[全部] [deepseek-v4-flash] [deepseek-v4-pro] 等）
2. 默认选中"全部"，显示该用户所有调用记录
3. 点击具体模型 tab，表格只显示该模型的记录，分页重置到第 1 页
4. 切换模型后分页正常工作

- [ ] **Step 4: 手动验算费用**

在明细抽屉选"全部"，取前几条记录的费用列，与 Flash/Pro 各自的定价对比是否匹配（Flash: 输入 ¥0.27/M，输出 ¥1.10/M；Pro: 输入 ¥4/M，输出 ¥16/M）
