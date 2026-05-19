package service

import (
	"context"
	"encoding/json"
	"log"
	"strconv"
	"sync"
	"time"

	"yuanju/internal/repository"
)

// BudgetStatusScope 单一维度（今日 / 本月）的预算状态
type BudgetStatusScope struct {
	TotalTokens      int     `json:"total_tokens"`
	TotalCostCny     float64 `json:"total_cost_cny"`
	ThresholdCostCny float64 `json:"threshold_cost_cny"`
	Exceeded         bool    `json:"exceeded"`
	ExceededPct      int     `json:"exceeded_pct"`
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
func snapshotLastAlertedAt() map[string]time.Time {
	return globalAlertState.snapshot()
}

// alertState 是 ticker 的内存告警状态。
// Key 为 scope 名（"daily_total" / "monthly_total" / "per_chart"），
// value 为上次告警时间。重启清零。
type alertState struct {
	mu            sync.Mutex
	lastAlertedAt map[string]time.Time
	dedupWindow   time.Duration
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
