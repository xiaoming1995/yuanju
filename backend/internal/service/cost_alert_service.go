package service

import (
	"strconv"
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
// Task 3 实现 ticker 时填充。本任务里返回 nil。
func snapshotLastAlertedAt() map[string]time.Time {
	return nil
}
