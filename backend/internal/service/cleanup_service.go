package service

import (
	"context"
	"encoding/json"
	"log"
	"time"
)

// ---- 注入接口 ----

// TableCleaner 删除某张表中 created_at < cutoff 的行，返回删除条数。
type TableCleaner func(ctx context.Context, cutoff time.Time) (int64, error)

// TokenRollup 把 token_usage_logs 闭合月份汇总到 monthly 表后删源行。
type TokenRollupFn func(ctx context.Context) (RollupReport, error)

// Clock 让单测固定时间。
type Clock interface{ Now() time.Time }

// RealClock 生产环境用。
type RealClock struct{}

func (RealClock) Now() time.Time { return time.Now() }

// ---- 输出结构 ----

type TableResult struct {
	Name    string
	Deleted int64
	Err     error
}

type RunReport struct {
	StartedAt time.Time
	Duration  time.Duration
	Tables    []TableResult
	Rollup    RollupReport
	Errors    []string
}

// RollupReport 是 service 层 token rollup 一次运行的结构化结果。
// 与 repository.RollupReport 结构等价；main.go 适配器做字段对拷。
type RollupReport struct {
	MonthsAggregated      int
	RowsInsertedOrUpdated int64
	SourceRowsDeleted     int64
	Err                   error
}

// ---- 依赖捆绑 ----

type CleanupDeps struct {
	AIReports     TableCleaner
	Polished      TableCleaner
	Liunian       TableCleaner
	PastEvents    TableCleaner
	DayunSummary  TableCleaner
	CompatReports TableCleaner
	RequestLogs   TableCleaner
	TokenRollup   TokenRollupFn
}

// ---- Service ----

type CleanupService struct {
	deps   CleanupDeps
	config func() (CleanupConfig, error)
	clock  Clock
	logger *log.Logger
}

func NewCleanupService(deps CleanupDeps) *CleanupService {
	return &CleanupService{
		deps:   deps,
		config: GetCleanupConfig,
		clock:  RealClock{},
		logger: log.Default(),
	}
}

// NewCleanupServiceForTest 是单测专用构造，允许注入 config loader + clock。
func NewCleanupServiceForTest(
	deps CleanupDeps,
	config func() (CleanupConfig, error),
	clock Clock,
) *CleanupService {
	return &CleanupService{
		deps:   deps,
		config: config,
		clock:  clock,
		logger: log.Default(),
	}
}

// RunOnce 跑一次完整清理流程。
// - 错误隔离：单张表失败收到 TableResult.Err，不中断后续表
// - rollup 在 repository 层独立事务里跑，失败也仅记录 Rollup.Err
func (s *CleanupService) RunOnce(ctx context.Context) RunReport {
	started := s.clock.Now()
	rep := RunReport{StartedAt: started}

	cfg, err := s.config()
	if err != nil {
		rep.Errors = append(rep.Errors, "load config: "+err.Error())
		rep.Duration = s.clock.Now().Sub(started)
		s.logRun(rep)
		return rep
	}
	if !cfg.Enabled {
		s.logger.Println("[cleanup] disabled by algo_config, skipping")
		rep.Duration = s.clock.Now().Sub(started)
		return rep
	}

	cutoff := s.clock.Now().Add(-time.Duration(cfg.RetentionDays) * 24 * time.Hour)

	type tableEntry struct {
		name string
		fn   TableCleaner
	}
	tables := []tableEntry{
		{"ai_reports", s.deps.AIReports},
		{"ai_polished_reports", s.deps.Polished},
		{"ai_liunian_reports", s.deps.Liunian},
		{"ai_past_events", s.deps.PastEvents},
		{"ai_dayun_summaries", s.deps.DayunSummary},
		{"ai_compatibility_reports", s.deps.CompatReports},
		{"ai_requests_log", s.deps.RequestLogs},
	}
	for _, t := range tables {
		deleted, err := t.fn(ctx, cutoff)
		rep.Tables = append(rep.Tables, TableResult{Name: t.name, Deleted: deleted, Err: err})
	}

	if rollup, err := s.deps.TokenRollup(ctx); err != nil {
		rep.Rollup = RollupReport{Err: err}
	} else {
		rep.Rollup = rollup
	}

	rep.Duration = s.clock.Now().Sub(started)
	s.logRun(rep)
	return rep
}

// StartScheduler 每天在 RunHour:00 跑一次 RunOnce。ctx.Done() 退出。
func (s *CleanupService) StartScheduler(ctx context.Context) {
	for {
		cfg, err := s.config()
		if err != nil {
			s.logger.Printf("[cleanup] load config in scheduler: %v; using default RunHour=3", err)
			cfg.RunHour = 3
		}
		next := s.nextRunAt(cfg.RunHour)
		wait := next.Sub(s.clock.Now())
		s.logger.Printf("[cleanup] next run at %s (in %s)", next.Format(time.RFC3339), wait)

		select {
		case <-ctx.Done():
			s.logger.Println("[cleanup] scheduler ctx done, exiting")
			return
		case <-time.After(wait):
			s.RunOnce(ctx)
		}
	}
}

func (s *CleanupService) nextRunAt(runHour int) time.Time {
	now := s.clock.Now()
	candidate := time.Date(now.Year(), now.Month(), now.Day(), runHour, 0, 0, 0, now.Location())
	if !candidate.After(now) {
		candidate = candidate.Add(24 * time.Hour)
	}
	return candidate
}

func (s *CleanupService) logRun(rep RunReport) {
	type tableLog struct {
		Name    string `json:"name"`
		Deleted int64  `json:"deleted"`
		Err     string `json:"err,omitempty"`
	}
	type rollupLog struct {
		MonthsAggregated      int    `json:"months_aggregated"`
		RowsInsertedOrUpdated int64  `json:"rows_inserted_or_updated"`
		SourceRowsDeleted     int64  `json:"source_rows_deleted"`
		Err                   string `json:"err,omitempty"`
	}
	type runLog struct {
		Evt       string     `json:"evt"`
		StartedAt string     `json:"started_at"`
		DurMs     int64      `json:"duration_ms"`
		Tables    []tableLog `json:"tables"`
		Rollup    rollupLog  `json:"rollup"`
		Errors    []string   `json:"errors,omitempty"`
	}
	out := runLog{
		Evt:       "cleanup_run",
		StartedAt: rep.StartedAt.Format(time.RFC3339),
		DurMs:     rep.Duration.Milliseconds(),
		Errors:    rep.Errors,
		Rollup: rollupLog{
			MonthsAggregated:      rep.Rollup.MonthsAggregated,
			RowsInsertedOrUpdated: rep.Rollup.RowsInsertedOrUpdated,
			SourceRowsDeleted:     rep.Rollup.SourceRowsDeleted,
		},
	}
	if rep.Rollup.Err != nil {
		out.Rollup.Err = rep.Rollup.Err.Error()
	}
	for _, t := range rep.Tables {
		entry := tableLog{Name: t.Name, Deleted: t.Deleted}
		if t.Err != nil {
			entry.Err = t.Err.Error()
		}
		out.Tables = append(out.Tables, entry)
	}
	b, _ := json.Marshal(out)
	s.logger.Println(string(b))
}

// MarshalRunReport 把 RunReport 序列化为 JSON（CLI --cleanup-once 模式打印用）。
func MarshalRunReport(rep RunReport) []byte {
	type tableOut struct {
		Name    string `json:"name"`
		Deleted int64  `json:"deleted"`
		Err     string `json:"err,omitempty"`
	}
	type rollupOut struct {
		MonthsAggregated      int    `json:"months_aggregated"`
		RowsInsertedOrUpdated int64  `json:"rows_inserted_or_updated"`
		SourceRowsDeleted     int64  `json:"source_rows_deleted"`
		Err                   string `json:"err,omitempty"`
	}
	type out struct {
		StartedAt string     `json:"started_at"`
		DurMs     int64      `json:"duration_ms"`
		Tables    []tableOut `json:"tables"`
		Rollup    rollupOut  `json:"rollup"`
		Errors    []string   `json:"errors,omitempty"`
	}
	o := out{
		StartedAt: rep.StartedAt.Format(time.RFC3339),
		DurMs:     rep.Duration.Milliseconds(),
		Errors:    rep.Errors,
		Rollup: rollupOut{
			MonthsAggregated:      rep.Rollup.MonthsAggregated,
			RowsInsertedOrUpdated: rep.Rollup.RowsInsertedOrUpdated,
			SourceRowsDeleted:     rep.Rollup.SourceRowsDeleted,
		},
	}
	if rep.Rollup.Err != nil {
		o.Rollup.Err = rep.Rollup.Err.Error()
	}
	for _, t := range rep.Tables {
		e := tableOut{Name: t.Name, Deleted: t.Deleted}
		if t.Err != nil {
			e.Err = t.Err.Error()
		}
		o.Tables = append(o.Tables, e)
	}
	b, _ := json.MarshalIndent(o, "", "  ")
	return b
}
