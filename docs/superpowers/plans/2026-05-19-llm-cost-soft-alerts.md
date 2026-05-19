# LLM Cost Soft Alerts Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a passive cost-visibility layer to the existing admin TokenUsagePage — daily / monthly / per-chart cost thresholds with banner + 3 stat cards + edit modal + 30s auto-refresh, plus a 5-min backend ticker emitting structured JSON logs when thresholds cross. No new tables, no enforcement.

**Architecture:** New service `cost_alert_service.go` aggregates `token_usage_logs` × `llm_providers.price` using the existing `service.CalcCost` function. Three thresholds stored as keys in the existing `algo_config` KV table. One new admin endpoint `GET /api/admin/token-usage/budget-status`. TokenUsagePage UI is extended (banner + cards) without replacing any existing controls. Scheduler reuses the goroutine ticker pattern from `cleanup_service.StartScheduler`.

**Tech Stack:** Go 1.21+, PostgreSQL, gin handlers, React 19 + TypeScript.

**Related spec:** `docs/superpowers/specs/2026-05-19-llm-cost-soft-alerts-design.md`

---

## File Structure

**Create:**
- `backend/internal/service/cost_alert_service.go` — `BuildBudgetStatus()` + `CostAlertScheduler` + ticker
- `backend/internal/service/cost_alert_service_test.go` — unit tests for both

**Modify:**
- `backend/internal/repository/token_usage_repository.go` — add `GetTokenUsageCostByModel(from, to)` and `GetChartCostBreakdown(from, to, limit)`
- `backend/internal/handler/token_usage_handler.go` — add `AdminGetBudgetStatus` handler
- `backend/internal/handler/algo_config_handler.go` — extend `validKeys` + value validation for 3 new keys
- `backend/pkg/seed/seed.go` — add `SeedCostAlertThresholds()` function
- `backend/cmd/api/main.go` — register route, call seed, start scheduler
- `frontend/src/lib/adminApi.ts` — add `budgetStatus()` method to `adminTokenUsageAPI`
- `frontend/src/pages/admin/TokenUsagePage.tsx` — banner + 3 stat cards + ⚙️ threshold modal + 30s auto-refresh

**Not touched (out of scope per spec):**
- Existing `GetTokenUsageSummary`/`Detail`/`Content` repository functions and handlers
- Existing TokenUsagePage filter / table / drawer / content modal
- `service.CalcCost` formula or `llm_providers.price` semantics

---

## Task 1: Repository aggregation functions

**Files:**
- Modify: `backend/internal/repository/token_usage_repository.go` — append two new functions at end of file

- [ ] **Step 1: Write failing tests**

Create `backend/internal/repository/token_usage_repository_test.go` (if not exists, append if exists). Add at file end:

```go
package repository

import (
	"testing"
	"time"
)

// fakeCostFn returns 0.01 CNY per 1000 tokens (any model)
var fakeCostFn = func(_ string, p, c int) float64 {
	return float64(p+c) * 0.01 / 1000
}

func TestGetTokenUsageCostByModel_AggregatesGroupedRows(t *testing.T) {
	if testing.Short() {
		t.Skip("requires DB")
	}
	now := time.Now()
	from := now.AddDate(0, 0, -1)
	to := now

	rows, totalTokens, err := GetTokenUsageCostByModel(from, to, fakeCostFn)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	// Even if DB is empty, slice should be non-nil and total should be >=0.
	if rows == nil {
		t.Error("expected non-nil rows")
	}
	if totalTokens < 0 {
		t.Errorf("totalTokens must be non-negative, got %d", totalTokens)
	}
}

func TestGetChartCostBreakdown_RespectsLimitAndSortsByCost(t *testing.T) {
	if testing.Short() {
		t.Skip("requires DB")
	}
	now := time.Now()
	from := now.AddDate(0, 0, -7)
	to := now

	rows, err := GetChartCostBreakdown(from, to, 5, fakeCostFn)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(rows) > 5 {
		t.Errorf("len > 5: %d", len(rows))
	}
	for i := 1; i < len(rows); i++ {
		if rows[i-1].TotalCostCny < rows[i].TotalCostCny {
			t.Errorf("rows[%d].TotalCostCny=%f < rows[%d].TotalCostCny=%f (must be desc)",
				i-1, rows[i-1].TotalCostCny, i, rows[i].TotalCostCny)
		}
	}
}
```

- [ ] **Step 2: Run tests — expect compile failure**

```bash
cd /Users/liujiming/web/yuanju/backend && go test ./internal/repository/ -run TestGetTokenUsageCostByModel_AggregatesGroupedRows 2>&1 | tail -5
```

Expected: `undefined: GetTokenUsageCostByModel` etc.

- [ ] **Step 3: Add `ChartCostRow` struct + two functions**

Append to end of `backend/internal/repository/token_usage_repository.go`:

```go
// ChartCostRow 单命盘成本聚合行
type ChartCostRow struct {
	ChartID      string  `json:"chart_id"`
	TotalCostCny float64 `json:"total_cost_cny"`
	TotalTokens  int     `json:"total_tokens"`
	Calls        int     `json:"calls"`
}

// modelCostRow 内部：按模型聚合的中间结构
type modelCostRow struct {
	model            string
	promptTokens     int
	completionTokens int
	totalTokens      int
	calls            int
}

// GetTokenUsageCostByModel 在 [from, to] 区间内按模型聚合 token 用量，调用 costFn 折算成本。
// 返回各模型成本明细 + 区间总 token 数。
// to 参数同 GetTokenUsageSummary 语义：表示最后一天，内部 +1 天取右开区间。
func GetTokenUsageCostByModel(from, to time.Time, costFn func(string, int, int) float64) ([]TokenUsageSummaryRow, int, error) {
	toExcl := to.AddDate(0, 0, 1)
	rows, err := database.DB.Query(`
		SELECT
			COALESCE(model, '')                        AS model,
			COUNT(*)::int                              AS calls,
			COALESCE(SUM(prompt_tokens), 0)::int       AS prompt_tokens,
			COALESCE(SUM(completion_tokens), 0)::int   AS completion_tokens,
			COALESCE(SUM(total_tokens), 0)::int        AS total_tokens
		FROM token_usage_logs
		WHERE created_at >= $1 AND created_at < $2
		GROUP BY model`,
		from, toExcl,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("GetTokenUsageCostByModel: %w", err)
	}
	defer rows.Close()

	out := []TokenUsageSummaryRow{}
	totalTokens := 0
	for rows.Next() {
		var r TokenUsageSummaryRow
		if err := rows.Scan(&r.Model, &r.RequestCount, &r.PromptTokens, &r.CompletionTokens, &r.TotalTokens); err != nil {
			log.Printf("[TokenUsage] GetTokenUsageCostByModel Scan 失败: %v", err)
			continue
		}
		if costFn != nil {
			r.EstimatedCostCny = costFn(r.Model, r.PromptTokens, r.CompletionTokens)
		}
		out = append(out, r)
		totalTokens += r.TotalTokens
	}
	return out, totalTokens, rows.Err()
}

// GetChartCostBreakdown 返回 [from, to] 区间内成本最高的 N 个命盘。
// 按 chart_id × model 聚合后，在 Go 层按 chart_id 求总成本并排序、取 top N。
func GetChartCostBreakdown(from, to time.Time, limit int, costFn func(string, int, int) float64) ([]ChartCostRow, error) {
	if limit <= 0 {
		limit = 5
	}
	toExcl := to.AddDate(0, 0, 1)
	rows, err := database.DB.Query(`
		SELECT
			chart_id::text                             AS chart_id,
			COALESCE(model, '')                        AS model,
			COUNT(*)::int                              AS calls,
			COALESCE(SUM(prompt_tokens), 0)::int       AS prompt_tokens,
			COALESCE(SUM(completion_tokens), 0)::int   AS completion_tokens,
			COALESCE(SUM(total_tokens), 0)::int        AS total_tokens
		FROM token_usage_logs
		WHERE chart_id IS NOT NULL AND created_at >= $1 AND created_at < $2
		GROUP BY chart_id, model`,
		from, toExcl,
	)
	if err != nil {
		return nil, fmt.Errorf("GetChartCostBreakdown: %w", err)
	}
	defer rows.Close()

	type aggCell struct {
		cost   float64
		tokens int
		calls  int
	}
	byChart := map[string]*aggCell{}
	for rows.Next() {
		var chartID, model string
		var calls, prompt, completion, total int
		if err := rows.Scan(&chartID, &model, &calls, &prompt, &completion, &total); err != nil {
			log.Printf("[TokenUsage] GetChartCostBreakdown Scan 失败: %v", err)
			continue
		}
		if byChart[chartID] == nil {
			byChart[chartID] = &aggCell{}
		}
		c := byChart[chartID]
		if costFn != nil {
			c.cost += costFn(model, prompt, completion)
		}
		c.tokens += total
		c.calls += calls
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	out := make([]ChartCostRow, 0, len(byChart))
	for chartID, c := range byChart {
		out = append(out, ChartCostRow{
			ChartID:      chartID,
			TotalCostCny: c.cost,
			TotalTokens:  c.tokens,
			Calls:        c.calls,
		})
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].TotalCostCny > out[j].TotalCostCny
	})
	if len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}
```

(`sort` is already imported in that file. Verify.)

- [ ] **Step 4: Run tests — both pass**

```bash
cd /Users/liujiming/web/yuanju/backend && go test ./internal/repository/ -run 'TestGetTokenUsageCostByModel|TestGetChartCostBreakdown' -v 2>&1 | tail -10
```

Expected: both PASS.

- [ ] **Step 5: Commit**

```bash
cd /Users/liujiming/web/yuanju && git add backend/internal/repository/token_usage_repository.go backend/internal/repository/token_usage_repository_test.go
git -c commit.gpgsign=false commit -m "feat(token-usage): add cost aggregation by model and per-chart

GetTokenUsageCostByModel: aggregates token_usage_logs in [from,to] by
model, calls costFn for each group. Returns per-model rows + total
token count.

GetChartCostBreakdown: per-chart cost aggregation with top-N limit,
sorted by total cost desc. Used by BuildBudgetStatus for the
single-chart threshold check."
```

---

## Task 2: `BuildBudgetStatus` service function + tests

**Files:**
- Create: `backend/internal/service/cost_alert_service.go`
- Create: `backend/internal/service/cost_alert_service_test.go`

- [ ] **Step 1: Write failing tests**

Create `backend/internal/service/cost_alert_service_test.go`:

```go
package service

import (
	"testing"
)

func TestBuildBudgetStatus_SmokeCheck(t *testing.T) {
	if testing.Short() {
		t.Skip("requires DB")
	}
	status, err := BuildBudgetStatus()
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if status == nil {
		t.Fatal("expected non-nil status")
	}
	if status.Today.ThresholdCostCny <= 0 {
		t.Errorf("today threshold should be > 0, got %v", status.Today.ThresholdCostCny)
	}
	if status.ThisMonth.ThresholdCostCny <= 0 {
		t.Errorf("month threshold should be > 0, got %v", status.ThisMonth.ThresholdCostCny)
	}
	if status.PerChartThresholdCny <= 0 {
		t.Errorf("per-chart threshold should be > 0, got %v", status.PerChartThresholdCny)
	}
	// Exceeded flag is computed correctly
	if status.Today.TotalCostCny > status.Today.ThresholdCostCny && !status.Today.Exceeded {
		t.Error("Today.Exceeded should be true when totalCost > threshold")
	}
}

func TestExceededPct_HandlesZeroThreshold(t *testing.T) {
	// 内部辅助函数，阈值为 0 时 pct 应返回 0 不崩
	pct := exceededPct(50, 0)
	if pct != 0 {
		t.Errorf("expected 0 for zero threshold, got %d", pct)
	}
	pct = exceededPct(50, 100)
	if pct != 50 {
		t.Errorf("expected 50, got %d", pct)
	}
	pct = exceededPct(150, 100)
	if pct != 150 {
		t.Errorf("expected 150, got %d", pct)
	}
}
```

- [ ] **Step 2: Run tests — expect compile failure**

```bash
cd /Users/liujiming/web/yuanju/backend && go test ./internal/service/ -run TestBuildBudgetStatus_SmokeCheck 2>&1 | tail -5
```

Expected: undefined types and functions.

- [ ] **Step 3: Implement service file**

Create `backend/internal/service/cost_alert_service.go`:

```go
package service

import (
	"strconv"
	"time"

	"yuanju/internal/repository"
)

// BudgetStatusScope 单一维度（今日 / 本月）的预算状态
type BudgetStatusScope struct {
	TotalTokens       int     `json:"total_tokens"`
	TotalCostCny      float64 `json:"total_cost_cny"`
	ThresholdCostCny  float64 `json:"threshold_cost_cny"`
	Exceeded          bool    `json:"exceeded"`
	ExceededPct       int     `json:"exceeded_pct"`
}

// TopChartItem 单命盘成本卡片项
type TopChartItem struct {
	ChartID           string  `json:"chart_id"`
	TotalCostCny      float64 `json:"total_cost_cny"`
	Calls             int     `json:"calls"`
	ThresholdExceeded bool    `json:"threshold_exceeded"`
}

// BudgetStatus 完整预算状态快照
type BudgetStatus struct {
	Today                BudgetStatusScope    `json:"today"`
	ThisMonth            BudgetStatusScope    `json:"this_month"`
	TopCharts            []TopChartItem       `json:"top_charts"`
	PerChartThresholdCny float64              `json:"per_chart_threshold_cny"`
	LastAlertedAt        map[string]time.Time `json:"last_alerted_at,omitempty"`
}

// 三个阈值键
const (
	keyDailyCostThreshold    = "cost_alert_daily_cost_cny"
	keyMonthlyCostThreshold  = "cost_alert_monthly_cost_cny"
	keyPerChartCostThreshold = "cost_alert_per_chart_cost_cny"
)

// 默认阈值（admin 未配置时）
const (
	defaultDailyCostThreshold    = 5.0
	defaultMonthlyCostThreshold  = 100.0
	defaultPerChartCostThreshold = 1.0
)

// BuildBudgetStatus 聚合今日 / 本月 / 单命盘 TOP 5 成本状态。
//
// 复用 service.CalcCost 计算单元成本，保证与 TokenUsagePage 现有显示口径一致。
// 阈值从 algo_config 表实时读取，不缓存（admin 改完即时生效）。
func BuildBudgetStatus() (*BudgetStatus, error) {
	now := time.Now()

	dayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	dayEnd := dayStart // GetTokenUsageCostByModel 内部 +1 天取右开区间

	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	monthEnd := now

	sevenDaysAgo := now.AddDate(0, 0, -7)

	dailyThreshold := readFloatConfig(keyDailyCostThreshold, defaultDailyCostThreshold)
	monthlyThreshold := readFloatConfig(keyMonthlyCostThreshold, defaultMonthlyCostThreshold)
	perChartThreshold := readFloatConfig(keyPerChartCostThreshold, defaultPerChartCostThreshold)

	// Today
	todayRows, todayTokens, err := repository.GetTokenUsageCostByModel(dayStart, dayEnd, CalcCost)
	if err != nil {
		return nil, err
	}
	todayCost := sumCost(todayRows)

	// This month
	monthRows, monthTokens, err := repository.GetTokenUsageCostByModel(monthStart, monthEnd, CalcCost)
	if err != nil {
		return nil, err
	}
	monthCost := sumCost(monthRows)

	// Top charts (7 day window)
	chartRows, err := repository.GetChartCostBreakdown(sevenDaysAgo, now, 5, CalcCost)
	if err != nil {
		return nil, err
	}
	topCharts := make([]TopChartItem, 0, len(chartRows))
	for _, r := range chartRows {
		topCharts = append(topCharts, TopChartItem{
			ChartID:           r.ChartID,
			TotalCostCny:      r.TotalCostCny,
			Calls:             r.Calls,
			ThresholdExceeded: r.TotalCostCny > perChartThreshold,
		})
	}

	return &BudgetStatus{
		Today: BudgetStatusScope{
			TotalTokens:      todayTokens,
			TotalCostCny:     todayCost,
			ThresholdCostCny: dailyThreshold,
			Exceeded:         todayCost > dailyThreshold,
			ExceededPct:      exceededPct(todayCost, dailyThreshold),
		},
		ThisMonth: BudgetStatusScope{
			TotalTokens:      monthTokens,
			TotalCostCny:     monthCost,
			ThresholdCostCny: monthlyThreshold,
			Exceeded:         monthCost > monthlyThreshold,
			ExceededPct:      exceededPct(monthCost, monthlyThreshold),
		},
		TopCharts:            topCharts,
		PerChartThresholdCny: perChartThreshold,
		LastAlertedAt:        snapshotLastAlertedAt(),
	}, nil
}

func sumCost(rows []repository.TokenUsageSummaryRow) float64 {
	var sum float64
	for _, r := range rows {
		sum += r.EstimatedCostCny
	}
	return sum
}

// exceededPct 返回 current / threshold 的百分比整数。threshold == 0 时返回 0。
func exceededPct(current, threshold float64) int {
	if threshold <= 0 {
		return 0
	}
	return int(current / threshold * 100)
}

// readFloatConfig 从 algo_config 读取一个 key，解析失败 / key 不存在则返回 fallback。
func readFloatConfig(key string, fallback float64) float64 {
	all, err := repository.GetAllAlgoConfig()
	if err != nil {
		return fallback
	}
	for _, r := range all {
		if r.Key == key {
			v, perr := strconv.ParseFloat(r.Value, 64)
			if perr == nil && v > 0 {
				return v
			}
			return fallback
		}
	}
	return fallback
}

// snapshotLastAlertedAt 返回当前 ticker 内存中的 lastAlertedAt 副本。
// Task 3 实现 ticker 时填充。本任务里返回 nil。
func snapshotLastAlertedAt() map[string]time.Time {
	return nil
}
```

- [ ] **Step 4: Run tests — pass**

```bash
cd /Users/liujiming/web/yuanju/backend && go test ./internal/service/ -run 'TestBuildBudgetStatus_SmokeCheck|TestExceededPct_HandlesZeroThreshold' -v 2>&1 | tail -10
```

Expected: both PASS.

- [ ] **Step 5: Commit**

```bash
cd /Users/liujiming/web/yuanju && git add backend/internal/service/cost_alert_service.go backend/internal/service/cost_alert_service_test.go
git -c commit.gpgsign=false commit -m "feat(cost-alert): BuildBudgetStatus aggregates today/month/top-charts

Service-layer function that aggregates token_usage_logs into a single
snapshot: daily/monthly totals with thresholds, top-5 charts by cost
in the last 7-day window. Thresholds read live from algo_config table
on every call (admin changes take effect immediately).

snapshotLastAlertedAt returns nil for now; Task 3 wires it to the
ticker's in-memory state."
```

---

## Task 3: Cost alert ticker (scheduler) + tests

**Files:**
- Modify: `backend/internal/service/cost_alert_service.go` — add `CostAlertScheduler` + ticker logic + update `snapshotLastAlertedAt`
- Modify: `backend/internal/service/cost_alert_service_test.go` — add ticker dedup test

- [ ] **Step 1: Write failing ticker test**

Append to `backend/internal/service/cost_alert_service_test.go`:

```go
func TestAlertDedup_FirstCrossLogsThenSilences(t *testing.T) {
	state := newAlertState()

	// First exceeded → should log
	if !state.shouldAlert("daily_total", time.Now()) {
		t.Error("first exceeded crossing should log")
	}
	state.markAlerted("daily_total", time.Now())

	// Immediately again → should NOT log (within 1h window)
	if state.shouldAlert("daily_total", time.Now().Add(30*time.Minute)) {
		t.Error("within 1h dedup window should not log")
	}

	// >1h later → should log again
	if !state.shouldAlert("daily_total", time.Now().Add(2*time.Hour)) {
		t.Error("after 1h should log again")
	}
}

func TestAlertState_DifferentScopesIndependent(t *testing.T) {
	state := newAlertState()
	now := time.Now()
	state.markAlerted("daily_total", now)

	// 不同 scope 的告警互不影响
	if !state.shouldAlert("monthly_total", now) {
		t.Error("monthly_total should be allowed even when daily_total just alerted")
	}
}
```

- [ ] **Step 2: Run tests — expect compile failure**

```bash
cd /Users/liujiming/web/yuanju/backend && go test ./internal/service/ -run TestAlertDedup 2>&1 | tail -5
```

Expected: `undefined: newAlertState`.

- [ ] **Step 3: Add scheduler + alert state to `cost_alert_service.go`**

Replace the `snapshotLastAlertedAt` stub from Task 2 (at the bottom of the file) with this full ticker implementation:

```go
// alertState 是 ticker 的内存告警状态。
// Key 为 scope 名（"daily_total" / "monthly_total" / "per_chart"），
// value 为上次告警时间。重启清零。
type alertState struct {
	mu              sync.Mutex
	lastAlertedAt   map[string]time.Time
	dedupWindow     time.Duration
}

func newAlertState() *alertState {
	return &alertState{
		lastAlertedAt: map[string]time.Time{},
		dedupWindow:   time.Hour,
	}
}

func (s *alertState) shouldAlert(scope string, now time.Time) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	last, ok := s.lastAlertedAt[scope]
	if !ok {
		return true
	}
	return now.Sub(last) >= s.dedupWindow
}

func (s *alertState) markAlerted(scope string, now time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.lastAlertedAt[scope] = now
}

func (s *alertState) snapshot() map[string]time.Time {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make(map[string]time.Time, len(s.lastAlertedAt))
	for k, v := range s.lastAlertedAt {
		out[k] = v
	}
	return out
}

// 包级全局 alertState（被 ticker 和 BuildBudgetStatus 共享）
var globalAlertState = newAlertState()

// CostAlertScheduler ticker 调度器
type CostAlertScheduler struct {
	interval time.Duration
	logger   *log.Logger
}

// NewCostAlertScheduler 构造默认 5 分钟 tick 间隔的 scheduler
func NewCostAlertScheduler() *CostAlertScheduler {
	return &CostAlertScheduler{
		interval: 5 * time.Minute,
		logger:   log.Default(),
	}
}

// StartScheduler 启动后台 ticker。调用方应在 main 里 go scheduler.StartScheduler(ctx)。
// ctx 取消时退出。
func (s *CostAlertScheduler) StartScheduler(ctx context.Context) {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()
	s.logger.Printf("[cost_alert] scheduler started (interval=%s)", s.interval)
	for {
		select {
		case <-ctx.Done():
			s.logger.Printf("[cost_alert] scheduler stopped")
			return
		case <-ticker.C:
			s.runOnce()
		}
	}
}

// runOnce 单次检测：聚合预算状态，必要时打日志
func (s *CostAlertScheduler) runOnce() {
	status, err := BuildBudgetStatus()
	if err != nil {
		s.logger.Printf("[cost_alert] BuildBudgetStatus 失败: %v", err)
		return
	}
	now := time.Now()
	if status.Today.Exceeded && globalAlertState.shouldAlert("daily_total", now) {
		s.emitAlert("daily_total", status.Today.TotalCostCny, status.Today.ThresholdCostCny)
		globalAlertState.markAlerted("daily_total", now)
	}
	if status.ThisMonth.Exceeded && globalAlertState.shouldAlert("monthly_total", now) {
		s.emitAlert("monthly_total", status.ThisMonth.TotalCostCny, status.ThisMonth.ThresholdCostCny)
		globalAlertState.markAlerted("monthly_total", now)
	}
	// per_chart：只要 TOP 列表里有任何越界命盘就 alert（一次性告知，不区分到具体命盘）
	for _, c := range status.TopCharts {
		if c.ThresholdExceeded {
			if globalAlertState.shouldAlert("per_chart", now) {
				s.emitAlert("per_chart", c.TotalCostCny, status.PerChartThresholdCny)
				globalAlertState.markAlerted("per_chart", now)
			}
			break
		}
	}
}

// emitAlert 写一行结构化 JSON 日志
func (s *CostAlertScheduler) emitAlert(scope string, currentCny, thresholdCny float64) {
	type out struct {
		Evt          string  `json:"evt"`
		Scope        string  `json:"scope"`
		CurrentCny   float64 `json:"current_cny"`
		ThresholdCny float64 `json:"threshold_cny"`
		ExceededPct  int     `json:"exceeded_pct"`
		TimestampISO string  `json:"timestamp"`
	}
	o := out{
		Evt:          "cost_threshold_exceeded",
		Scope:        scope,
		CurrentCny:   currentCny,
		ThresholdCny: thresholdCny,
		ExceededPct:  exceededPct(currentCny, thresholdCny),
		TimestampISO: time.Now().UTC().Format(time.RFC3339),
	}
	b, _ := json.Marshal(o)
	s.logger.Println(string(b))
}
```

Update the existing `snapshotLastAlertedAt()` stub to delegate to the global state:

```go
// snapshotLastAlertedAt 返回当前 ticker 内存中的 lastAlertedAt 副本。
func snapshotLastAlertedAt() map[string]time.Time {
	return globalAlertState.snapshot()
}
```

Add to imports at top of file:

```go
import (
	"context"
	"encoding/json"
	"log"
	"strconv"
	"sync"
	"time"

	"yuanju/internal/repository"
)
```

- [ ] **Step 4: Run all cost_alert tests**

```bash
cd /Users/liujiming/web/yuanju/backend && go test ./internal/service/ -run 'TestBuildBudgetStatus|TestExceededPct|TestAlertDedup|TestAlertState' -v 2>&1 | tail -15
```

Expected: all 4 PASS.

- [ ] **Step 5: Commit**

```bash
cd /Users/liujiming/web/yuanju && git add backend/internal/service/cost_alert_service.go backend/internal/service/cost_alert_service_test.go
git -c commit.gpgsign=false commit -m "feat(cost-alert): scheduler ticker + in-memory dedup state

CostAlertScheduler ticks every 5 minutes (configurable), calls
BuildBudgetStatus, emits a structured JSON log line per scope when
threshold first crosses + dedups subsequent crossings within 1 hour.

Reset on process restart — that's intentional: ops gets immediate
re-notification of any active breach after a deploy."
```

---

## Task 4: Admin handler endpoint

**Files:**
- Modify: `backend/internal/handler/token_usage_handler.go` — add `AdminGetBudgetStatus` handler

- [ ] **Step 1: Add the handler**

Append to `backend/internal/handler/token_usage_handler.go`:

```go
// AdminGetBudgetStatus GET /api/admin/token-usage/budget-status
// 返回今日 / 本月 / 单命盘 TOP 5 成本快照 + 阈值。
func AdminGetBudgetStatus(c *gin.Context) {
	status, err := service.BuildBudgetStatus()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, status)
}
```

- [ ] **Step 2: Verify compile**

```bash
cd /Users/liujiming/web/yuanju/backend && go build ./...
```

Expected: no errors.

- [ ] **Step 3: Commit**

```bash
cd /Users/liujiming/web/yuanju && git add backend/internal/handler/token_usage_handler.go
git -c commit.gpgsign=false commit -m "feat(token-usage): AdminGetBudgetStatus endpoint

GET /api/admin/token-usage/budget-status — returns the BudgetStatus
JSON shape (today, this_month, top_charts, thresholds). Route
registered in Task 7."
```

---

## Task 5: Extend `algo_config_handler` whitelist + value validation

**Files:**
- Modify: `backend/internal/handler/algo_config_handler.go`

- [ ] **Step 1: Add the 3 keys to `validKeys` + validation switch**

In `backend/internal/handler/algo_config_handler.go`, find the `validKeys` map (around line 58) and the `switch key { ... }` validation block (around line 79).

Inside `validKeys` add three entries:

```go
	validKeys := map[string]bool{
		"jixiong_jiHan_min":    true,
		"jixiong_jiRe_min":     true,
		"jixiong_shenQiang_pct": true,
		"year_narrative_mode":  true,
		"cost_alert_daily_cost_cny":     true,
		"cost_alert_monthly_cost_cny":   true,
		"cost_alert_per_chart_cost_cny": true,
	}
```

(Preserve the entries already there; only ADD the three cost_alert_* lines.)

Inside the value-validation `switch key { ... }` block, add a new case BEFORE the existing `case "year_narrative_mode":`:

```go
	case "cost_alert_daily_cost_cny", "cost_alert_monthly_cost_cny", "cost_alert_per_chart_cost_cny":
		v, perr := strconv.ParseFloat(value, 64)
		if perr != nil || v <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": key + " 必须是大于 0 的数字"})
			return
		}
```

If `strconv` isn't yet imported at the top of the file, add it.

- [ ] **Step 2: Verify compile**

```bash
cd /Users/liujiming/web/yuanju/backend && go build ./...
```

Expected: no errors.

- [ ] **Step 3: Verify all tests still pass**

```bash
cd /Users/liujiming/web/yuanju/backend && go test ./...
```

Expected: all green.

- [ ] **Step 4: Commit**

```bash
cd /Users/liujiming/web/yuanju && git add backend/internal/handler/algo_config_handler.go
git -c commit.gpgsign=false commit -m "feat(algo_config): whitelist cost_alert_* threshold keys

Allow admin to write three cost threshold keys via PUT /api/admin/
algo-config/:key. Value must parse as float > 0 (CNY)."
```

---

## Task 6: Seed default thresholds

**Files:**
- Modify: `backend/pkg/seed/seed.go` — add `SeedCostAlertThresholds()`

- [ ] **Step 1: Add the seed function**

Append to `backend/pkg/seed/seed.go`:

```go
// SeedCostAlertThresholds 将默认成本阈值写入 algo_config（ON CONFLICT DO NOTHING）。
// 启动时调用一次，admin 已改的值不会被覆盖。
func SeedCostAlertThresholds() {
	defaults := []struct {
		key   string
		value string
		desc  string
	}{
		{"cost_alert_daily_cost_cny", "5", "单日 AI 总成本告警阈值（CNY）"},
		{"cost_alert_monthly_cost_cny", "100", "单月 AI 总成本告警阈值（CNY）"},
		{"cost_alert_per_chart_cost_cny", "1", "单命盘 AI 总成本告警阈值（CNY）"},
	}
	for _, d := range defaults {
		if _, err := database.DB.Exec(
			`INSERT INTO algo_config (key, value, description)
			 VALUES ($1, $2, $3)
			 ON CONFLICT (key) DO NOTHING`,
			d.key, d.value, d.desc,
		); err != nil {
			log.Printf("[seed] SeedCostAlertThresholds %s 失败: %v", d.key, err)
			return
		}
	}
	log.Println("✅ 种子数据：成本告警阈值已写入 algo_config（ON CONFLICT DO NOTHING）")
}
```

(Verify `database` package and `log` are imported at file top — they should be from the existing seed functions.)

- [ ] **Step 2: Verify compile**

```bash
cd /Users/liujiming/web/yuanju/backend && go build ./...
```

Expected: no errors.

- [ ] **Step 3: Commit**

```bash
cd /Users/liujiming/web/yuanju && git add backend/pkg/seed/seed.go
git -c commit.gpgsign=false commit -m "feat(seed): default cost alert thresholds

Inserts cost_alert_daily / monthly / per_chart rows into algo_config
on startup with ON CONFLICT DO NOTHING. Defaults: ¥5/day, ¥100/month,
¥1/chart — chosen based on observed ~¥0.36/chart cost from the recent
dayun AI rollout."
```

---

## Task 7: Wire route + seed + scheduler in `main.go`

**Files:**
- Modify: `backend/cmd/api/main.go`

- [ ] **Step 1: Add seed call after existing seeds**

In `backend/cmd/api/main.go`, find the existing seed block (around line 65):

```go
	seed.SeedLLMProviders()
	seed.SeedLLMPrices()
```

Add a third line:

```go
	seed.SeedLLMProviders()
	seed.SeedLLMPrices()
	seed.SeedCostAlertThresholds()
```

- [ ] **Step 2: Start the cost alert scheduler**

In `backend/cmd/api/main.go`, find the existing `cleanupSvc.StartScheduler(schedCtx)` line (around line 84). Right after it, add:

```go
	go cleanupSvc.StartScheduler(schedCtx)

	// Cost alert scheduler — 5 min ticker, emits structured JSON log on threshold breach
	costAlertScheduler := service.NewCostAlertScheduler()
	go costAlertScheduler.StartScheduler(schedCtx)
```

Verify `"yuanju/internal/service"` is imported at the top (it should be — `cleanupSvc` already uses it).

- [ ] **Step 3: Register the budget-status route**

Find the existing token-usage routes (around line 198):

```go
				adminAuth.GET("/token-usage/summary", handler.AdminGetTokenUsageSummary)
				adminAuth.GET("/token-usage/detail", handler.AdminGetTokenUsageDetail)
				adminAuth.GET("/token-usage/content/:id", handler.AdminGetTokenUsageContent)
```

Add a fourth line:

```go
				adminAuth.GET("/token-usage/summary", handler.AdminGetTokenUsageSummary)
				adminAuth.GET("/token-usage/detail", handler.AdminGetTokenUsageDetail)
				adminAuth.GET("/token-usage/content/:id", handler.AdminGetTokenUsageContent)
				adminAuth.GET("/token-usage/budget-status", handler.AdminGetBudgetStatus)
```

- [ ] **Step 4: Verify compile**

```bash
cd /Users/liujiming/web/yuanju/backend && go build ./...
```

Expected: no errors.

- [ ] **Step 5: Run full test suite**

```bash
cd /Users/liujiming/web/yuanju/backend && go test ./...
```

Expected: all green.

- [ ] **Step 6: Commit**

```bash
cd /Users/liujiming/web/yuanju && git add backend/cmd/api/main.go
git -c commit.gpgsign=false commit -m "feat(main): wire seed + scheduler + route for cost soft alerts

- Call seed.SeedCostAlertThresholds() at boot (idempotent)
- Start CostAlertScheduler as goroutine alongside cleanup scheduler
- Register GET /api/admin/token-usage/budget-status (admin auth)"
```

---

## Task 8: Frontend — adminApi + TokenUsagePage UI

**Files:**
- Modify: `frontend/src/lib/adminApi.ts` — add `budgetStatus()` method
- Modify: `frontend/src/pages/admin/TokenUsagePage.tsx` — banner + cards + modal + auto-refresh

- [ ] **Step 1: Add `budgetStatus()` to adminApi**

In `frontend/src/lib/adminApi.ts`, find `adminTokenUsageAPI` (around line 100 likely). Add a new method:

```ts
export const adminTokenUsageAPI = {
  summary: (from: string, to: string) =>
    adminApi.get(`/api/admin/token-usage/summary?from=${from}&to=${to}`),
  detail: (userID: string, from: string, to: string, page: number, limit: number, model: string) =>
    adminApi.get(
      `/api/admin/token-usage/detail?user_id=${userID}&from=${from}&to=${to}&page=${page}&limit=${limit}&model=${encodeURIComponent(model)}`
    ),
  content: (id: string) =>
    adminApi.get(`/api/admin/token-usage/content/${id}`),
  budgetStatus: () =>
    adminApi.get('/api/admin/token-usage/budget-status'),
}
```

(Preserve existing methods; only ADD `budgetStatus`.)

- [ ] **Step 2: Define TS interfaces in TokenUsagePage.tsx**

In `frontend/src/pages/admin/TokenUsagePage.tsx`, after the existing `interface ContentModal` declaration (around line 41), add:

```tsx
interface BudgetStatusScope {
  total_tokens: number
  total_cost_cny: number
  threshold_cost_cny: number
  exceeded: boolean
  exceeded_pct: number
}

interface TopChartItem {
  chart_id: string
  total_cost_cny: number
  calls: number
  threshold_exceeded: boolean
}

interface BudgetStatus {
  today: BudgetStatusScope
  this_month: BudgetStatusScope
  top_charts: TopChartItem[]
  per_chart_threshold_cny: number
  last_alerted_at?: Record<string, string>
}

interface ThresholdEdit {
  daily: string
  monthly: string
  per_chart: string
}
```

- [ ] **Step 3: Add state hooks to the component**

In the `TokenUsagePage` component body (right after the existing `useState` hooks like `const [from, setFrom] = useState(...)`), add:

```tsx
  const [budget, setBudget] = useState<BudgetStatus | null>(null)
  const [thresholdEditOpen, setThresholdEditOpen] = useState(false)
  const [thresholdDraft, setThresholdDraft] = useState<ThresholdEdit>({ daily: '', monthly: '', per_chart: '' })
  const [thresholdSaving, setThresholdSaving] = useState(false)
  const [thresholdError, setThresholdError] = useState('')
```

- [ ] **Step 4: Add the fetch + auto-refresh effect**

In the same component body, after the existing effects, add a new effect that loads budget on mount + refreshes every 30s:

```tsx
  useEffect(() => {
    let cancelled = false

    const load = async () => {
      try {
        const resp = await adminTokenUsageAPI.budgetStatus()
        if (!cancelled) setBudget(resp.data)
      } catch (e) {
        // Silent — banner just won't render. Don't disturb the summary table.
        console.warn('[budget-status] fetch failed', e)
      }
    }

    void load()
    const t = window.setInterval(() => void load(), 30000)
    return () => {
      cancelled = true
      window.clearInterval(t)
    }
  }, [])
```

- [ ] **Step 5: Add banner + 3 stat card render**

In the `TokenUsagePage` return JSX, find the existing `<h1>` title block (around line 169). Replace the entire `<h1>` line and what follows up to (but not including) `{/* 筛选栏 */}` with:

```tsx
      <h1 className="admin-page-title" style={{ display: 'flex', alignItems: 'center', gap: 8, justifyContent: 'space-between' }}>
        <span style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
          <BarChart2 size={24} /> Token 用量统计
        </span>
        <button
          className="admin-btn"
          onClick={() => {
            if (budget) {
              setThresholdDraft({
                daily: String(budget.today.threshold_cost_cny),
                monthly: String(budget.this_month.threshold_cost_cny),
                per_chart: String(budget.per_chart_threshold_cny),
              })
            }
            setThresholdError('')
            setThresholdEditOpen(true)
          }}
          style={{ fontSize: 13 }}
        >
          ⚙️ 编辑预算阈值
        </button>
      </h1>

      {/* 越界 banner */}
      {budget && (budget.today.exceeded || budget.this_month.exceeded) && (
        <div
          style={{
            background: '#7f1d1d',
            color: '#fee2e2',
            padding: '12px 16px',
            borderRadius: 8,
            marginBottom: 20,
            border: '1px solid #b91c1c',
            fontSize: 14,
            lineHeight: 1.7,
          }}
        >
          ⚠️ 预算告警
          {budget.today.exceeded && (
            <div>
              今日已用 <strong>¥{budget.today.total_cost_cny.toFixed(2)}</strong>
              （{budget.today.exceeded_pct}% 日预算 ¥{budget.today.threshold_cost_cny.toFixed(2)}）
            </div>
          )}
          {budget.this_month.exceeded && (
            <div>
              本月已用 <strong>¥{budget.this_month.total_cost_cny.toFixed(2)}</strong>
              （{budget.this_month.exceeded_pct}% 月预算 ¥{budget.this_month.threshold_cost_cny.toFixed(2)}）
            </div>
          )}
        </div>
      )}

      {/* 预算状态卡片 */}
      {budget && (
        <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr 1fr', gap: 16, marginBottom: 24 }}>
          {/* Today card */}
          <div className="admin-card" style={{ padding: 16 }}>
            <div style={{ color: '#888', fontSize: 12, marginBottom: 8 }}>今日累计</div>
            <div style={{ color: budget.today.exceeded ? '#fca5a5' : budget.today.exceeded_pct >= 80 ? '#fbbf24' : '#86efac', fontSize: 26, fontWeight: 700 }}>
              ¥{budget.today.total_cost_cny.toFixed(2)}
            </div>
            <div style={{ fontSize: 12, color: '#aaa', marginTop: 6 }}>
              {budget.today.exceeded_pct}% · 阈值 ¥{budget.today.threshold_cost_cny.toFixed(2)}
            </div>
          </div>

          {/* This month card */}
          <div className="admin-card" style={{ padding: 16 }}>
            <div style={{ color: '#888', fontSize: 12, marginBottom: 8 }}>本月累计</div>
            <div style={{ color: budget.this_month.exceeded ? '#fca5a5' : budget.this_month.exceeded_pct >= 80 ? '#fbbf24' : '#86efac', fontSize: 26, fontWeight: 700 }}>
              ¥{budget.this_month.total_cost_cny.toFixed(2)}
            </div>
            <div style={{ fontSize: 12, color: '#aaa', marginTop: 6 }}>
              {budget.this_month.exceeded_pct}% · 阈值 ¥{budget.this_month.threshold_cost_cny.toFixed(2)}
            </div>
          </div>

          {/* Top charts card */}
          <div className="admin-card" style={{ padding: 16 }}>
            <div style={{ color: '#888', fontSize: 12, marginBottom: 8 }}>
              单命盘 TOP 5（近 7 天，阈值 ¥{budget.per_chart_threshold_cny.toFixed(2)}）
            </div>
            {budget.top_charts.length === 0 ? (
              <div style={{ color: '#666', fontSize: 13 }}>暂无数据</div>
            ) : (
              <div style={{ display: 'flex', flexDirection: 'column', gap: 4 }}>
                {budget.top_charts.map((c, i) => (
                  <div
                    key={c.chart_id}
                    title={c.chart_id}
                    style={{ fontSize: 12, display: 'flex', justifyContent: 'space-between', color: c.threshold_exceeded ? '#fca5a5' : '#e0e0e0' }}
                  >
                    <span>{i + 1}. {c.chart_id.slice(0, 8)}…{c.threshold_exceeded ? ' ⚠' : ''}</span>
                    <span>¥{c.total_cost_cny.toFixed(2)}（{c.calls} 次）</span>
                  </div>
                ))}
              </div>
            )}
          </div>
        </div>
      )}
```

- [ ] **Step 6: Add the threshold-edit modal**

At the end of the `return ( ... )` block (just before the closing `</div>` of the page root), add:

```tsx
      {/* 阈值编辑 modal */}
      {thresholdEditOpen && (
        <div
          onClick={() => setThresholdEditOpen(false)}
          style={{
            position: 'fixed', top: 0, left: 0, right: 0, bottom: 0,
            background: 'rgba(0,0,0,0.6)', display: 'flex', alignItems: 'center', justifyContent: 'center', zIndex: 999,
          }}
        >
          <div
            onClick={(e) => e.stopPropagation()}
            style={{ background: '#1a1a2e', padding: 24, borderRadius: 12, width: 420, color: '#e0e0e0' }}
          >
            <h3 style={{ margin: '0 0 16px', fontSize: 18 }}>编辑预算阈值（CNY）</h3>
            <label style={{ display: 'block', fontSize: 13, color: '#aaa', marginBottom: 4 }}>日累计上限</label>
            <input
              type="number"
              step="0.5"
              value={thresholdDraft.daily}
              onChange={(e) => setThresholdDraft({ ...thresholdDraft, daily: e.target.value })}
              style={{ width: '100%', background: '#0d0d1a', color: '#e0e0e0', border: '1px solid #333', borderRadius: 6, padding: 8, marginBottom: 12 }}
            />
            <label style={{ display: 'block', fontSize: 13, color: '#aaa', marginBottom: 4 }}>月累计上限</label>
            <input
              type="number"
              step="5"
              value={thresholdDraft.monthly}
              onChange={(e) => setThresholdDraft({ ...thresholdDraft, monthly: e.target.value })}
              style={{ width: '100%', background: '#0d0d1a', color: '#e0e0e0', border: '1px solid #333', borderRadius: 6, padding: 8, marginBottom: 12 }}
            />
            <label style={{ display: 'block', fontSize: 13, color: '#aaa', marginBottom: 4 }}>单命盘上限</label>
            <input
              type="number"
              step="0.1"
              value={thresholdDraft.per_chart}
              onChange={(e) => setThresholdDraft({ ...thresholdDraft, per_chart: e.target.value })}
              style={{ width: '100%', background: '#0d0d1a', color: '#e0e0e0', border: '1px solid #333', borderRadius: 6, padding: 8, marginBottom: 16 }}
            />
            {thresholdError && <div style={{ color: '#fca5a5', fontSize: 13, marginBottom: 12 }}>{thresholdError}</div>}
            <div style={{ display: 'flex', gap: 8, justifyContent: 'flex-end' }}>
              <button className="admin-btn" onClick={() => setThresholdEditOpen(false)} disabled={thresholdSaving}>取消</button>
              <button
                className="admin-btn admin-btn-primary"
                disabled={thresholdSaving}
                onClick={async () => {
                  setThresholdSaving(true)
                  setThresholdError('')
                  try {
                    await adminApi.put('/api/admin/algo-config/cost_alert_daily_cost_cny', { value: thresholdDraft.daily })
                    await adminApi.put('/api/admin/algo-config/cost_alert_monthly_cost_cny', { value: thresholdDraft.monthly })
                    await adminApi.put('/api/admin/algo-config/cost_alert_per_chart_cost_cny', { value: thresholdDraft.per_chart })
                    // Re-fetch budget
                    const resp = await adminTokenUsageAPI.budgetStatus()
                    setBudget(resp.data)
                    setThresholdEditOpen(false)
                  } catch (e: unknown) {
                    const err = e as { response?: { data?: { error?: string } }; message?: string }
                    setThresholdError(err.response?.data?.error || err.message || '保存失败')
                  } finally {
                    setThresholdSaving(false)
                  }
                }}
              >
                {thresholdSaving ? '保存中…' : '保存'}
              </button>
            </div>
          </div>
        </div>
      )}
```

Verify `adminApi` is imported at the top of the file. Check existing import block — `import { adminApi } from '../../lib/adminApi'` may or may not be present. If not, add it.

- [ ] **Step 7: Add `useEffect` import + verify other imports**

At the top of `frontend/src/pages/admin/TokenUsagePage.tsx`, change:

```tsx
import { useState, Fragment } from 'react'
```

to:

```tsx
import { useState, useEffect, Fragment } from 'react'
```

And ensure `adminApi` is imported alongside `adminTokenUsageAPI`:

```tsx
import { adminApi, adminTokenUsageAPI } from '../../lib/adminApi'
```

- [ ] **Step 8: Build + lint**

```bash
cd /Users/liujiming/web/yuanju/frontend && npm run build 2>&1 | tail -10
```

Expected: build succeeds, no TypeScript errors (pre-existing chunk-size warning is fine).

- [ ] **Step 9: Commit**

```bash
cd /Users/liujiming/web/yuanju && git add frontend/src/lib/adminApi.ts frontend/src/pages/admin/TokenUsagePage.tsx
git -c commit.gpgsign=false commit -m "feat(admin): cost soft alerts UI on TokenUsagePage

- New banner (red, only renders on exceedance) at top of page
- 3 stat cards: today / this_month / top_charts (near-budget yellow,
  exceeded red, healthy green)
- ⚙️ button in title opens modal to edit 3 threshold values; saves
  via existing PUT /api/admin/algo-config/:key endpoint
- 30s auto-refresh of budget status (cleared on unmount)
- adminApi gains budgetStatus() method

Existing summary table, filters, drawer, content modal untouched."
```

---

## Task 9: End-to-end verification

**Files:** none modified — verification only.

- [ ] **Step 1: Rebuild + restart both containers**

```bash
cd /Users/liujiming/web/yuanju && docker-compose up -d --build backend frontend 2>&1 | tail -10
```

Expected: both build and start.

- [ ] **Step 2: Verify seeded thresholds + scheduler started**

```bash
sleep 5
docker-compose exec postgres psql -U yuanju -d yuanju -c "SELECT key, value FROM algo_config WHERE key LIKE 'cost_alert%' ORDER BY key;"
docker-compose logs --tail=20 backend 2>&1 | grep "cost_alert"
```

Expected:
- Three rows in algo_config: `cost_alert_daily_cost_cny=5`, `cost_alert_monthly_cost_cny=100`, `cost_alert_per_chart_cost_cny=1`
- Log line: `[cost_alert] scheduler started (interval=5m0s)`

- [ ] **Step 3: Curl the budget-status endpoint**

```bash
TOKEN="<get a valid admin token from your test login>"
curl -s -H "Authorization: Bearer $TOKEN" "http://localhost:9002/api/admin/token-usage/budget-status" | python3 -m json.tool | head -30
```

Expected JSON shape:
```json
{
  "today": {"total_tokens": ..., "total_cost_cny": ..., "threshold_cost_cny": 5, "exceeded": ..., "exceeded_pct": ...},
  "this_month": {...},
  "top_charts": [...],
  "per_chart_threshold_cny": 1,
  "last_alerted_at": null
}
```

- [ ] **Step 4: Visual verification in browser**

Open `http://localhost:3000/admin/token-usage` (or wherever TokenUsagePage is routed).

Verify:
- ⚙️ "编辑预算阈值" button appears in the title bar
- 3 stat cards render (today / month / top 5)
- Red banner appears if today.exceeded or this_month.exceeded
- Clicking ⚙️ opens modal with 3 number inputs populated from current thresholds
- Saving modal triggers a refetch and banner/cards update
- Existing summary table, filters, drawer all still work as before

- [ ] **Step 5: Force a threshold breach to test banner + ticker log**

```bash
# Lower the daily threshold to ¥0.01 so any cost triggers a breach
docker-compose exec postgres psql -U yuanju -d yuanju -c \
  "UPDATE algo_config SET value = '0.01' WHERE key = 'cost_alert_daily_cost_cny';"
```

Refresh the browser (or wait 30s for auto-refresh). Expect:
- Red banner appears immediately
- "今日累计" card flips to red

Wait up to 5 minutes for the ticker to detect, then check backend logs:

```bash
docker-compose logs --tail=50 backend 2>&1 | grep "cost_threshold_exceeded"
```

Expected: a JSON log line like:
```json
{"evt":"cost_threshold_exceeded","scope":"daily_total","current_cny":...,"threshold_cny":0.01,...}
```

Restore the daily threshold:

```bash
docker-compose exec postgres psql -U yuanju -d yuanju -c \
  "UPDATE algo_config SET value = '5' WHERE key = 'cost_alert_daily_cost_cny';"
```

- [ ] **Step 6: No commit (verification only)**

Document any anomalies in the PR description if this branch is being PR'd.

---

## Self-Review

**Spec coverage:**

| Spec section | Tasks |
|---|---|
| 架构（双路检测、零新表）| 1, 2, 3, 7 |
| 数据流 + 新增接口（/budget-status）| 1, 2, 4, 7 |
| 聚合 SQL（复用 service.CalcCost）| 1, 2 |
| 触发时机（被动 + ticker 5min + 1h dedup）| 3, 8 (auto-refresh) |
| 前端 UI（banner + 3 cards + modal + 自动刷新）| 8 |
| 阈值默认值（¥5/¥100/¥1，algo_config seed）| 6 |
| 成本计算口径（复用 CalcCost）| 1, 2 |
| 范围边界 | "Not touched" section in File Structure |
| 风险表 | 不需任务（运营观察项）|

All spec items mapped to tasks.

**Placeholder scan:** No "TBD", no "fill in details", no "similar to Task N". All steps contain exact code or commands. Step 5 of Task 9 has a `<get a valid admin token...>` placeholder for the user-driven verification — that's a runtime token, not a code placeholder, so it's acceptable.

**Type consistency:**
- `BudgetStatus` / `BudgetStatusScope` / `TopChartItem` defined consistently in Task 2 (Go) and Task 8 (TS) with identical JSON tags
- `ChartCostRow` (repository) → `TopChartItem` (service) — naming differs but shape converts cleanly in BuildBudgetStatus
- `readFloatConfig`, `exceededPct`, `alertState` introduced in Task 2/3 referenced only there — no drift
- Threshold key strings used consistently: `cost_alert_daily_cost_cny`, `cost_alert_monthly_cost_cny`, `cost_alert_per_chart_cost_cny` (Tasks 2, 5, 6, 7, 8)
- 5-minute ticker interval and 1-hour dedup window are constants in `CostAlertScheduler` (Task 3) referenced nowhere else — no drift

No spec gaps. No placeholders. No type drift. Plan ready.
