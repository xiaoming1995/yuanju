# Data Bloat Defense Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 让 backend 自带数据保留策略：6 张 AI 缓存表 + 1 张请求日志表按 90 天滑动删除；`token_usage_logs` 按月汇总到 `token_usage_logs_monthly` 后删源行；调度由 backend goroutine 每天 3 AM 跑一次，阈值走 `algo_config`。

**Architecture:** 新增 `cleanup_service.go` 编排 8 个 repository 方法，由 `main.go` 启动 in-process scheduler。Clock 接口注入便于单测固定时间；`testcontainers-go` 跑 token rollup 集成测试。Admin UI 通过现有 `AlgoConfigPage` 暴露 3 个新 key。

**Tech Stack:** Go 1.25 / Gin / lib/pq / testcontainers-go(新增) / React 19 + TypeScript / Node 23 `--experimental-strip-types`

---

## Spec Reference

**Spec:** `docs/superpowers/specs/2026-05-18-data-bloat-defense-design.md` (commit `5fd397d` on `main`)

**Starting branch:** `main` HEAD=`5fd397d`. Task 0 cuts a new branch `feat/data-bloat-defense`.

---

## File Structure

**Create (4 files):**
- `backend/internal/service/cleanup_service.go` — 编排
- `backend/internal/service/cleanup_service_test.go` — 单测（mock 8 个 cleaner）
- `backend/internal/repository/cleanup_integration_test.go` — 集成测试（testcontainers + rollup SQL）
- `frontend/tests/cleanup-config.test.mjs` — grep 单测

**Modify (11 files):**
- `backend/cmd/api/main.go` — scheduler 启动 + `--cleanup-once` flag
- `backend/pkg/database/database.go` — `token_usage_logs_monthly` DDL + algo_config 默认值 seed
- `backend/internal/repository/repository.go` — `DeleteAIReportsOlderThan`
- `backend/internal/repository/polished_report_repo.go` — `DeletePolishedReportsOlderThan`
- `backend/internal/repository/liunian_repository.go` — `DeleteLiunianReportsOlderThan`
- `backend/internal/repository/past_events_repository.go` — `DeletePastEventsOlderThan`
- `backend/internal/repository/dayun_summary_repository.go` — `DeleteDayunSummariesOlderThan`
- `backend/internal/repository/compatibility_repository.go` — `DeleteAICompatibilityReportsOlderThan`
- `backend/internal/repository/admin_repository.go` — `DeleteRequestLogsOlderThan`
- `backend/internal/repository/token_usage_repository.go` — `RollupClosedMonthsAndDelete`
- `backend/internal/service/algo_config_service.go` — `GetCleanupConfig()` 新函数 + 3 个常量
- `frontend/src/pages/admin/AlgoConfigPage.tsx` — 3 个 PARAM_LABELS 条目

**go.mod additions:**
- `github.com/testcontainers/testcontainers-go`
- `github.com/testcontainers/testcontainers-go/modules/postgres`

---

## Task 0: Branch + Baseline

**Files:** none (git only)

- [ ] **Step 1: Create feature branch from main**

```bash
git -C /Users/liujiming/web/yuanju checkout -b feat/data-bloat-defense
git -C /Users/liujiming/web/yuanju status
```
Expected: `On branch feat/data-bloat-defense ... nothing to commit, working tree clean`

- [ ] **Step 2: Baseline build + test**

```bash
cd /Users/liujiming/web/yuanju/backend && go build ./... && go test ./...
```
Expected: PASS for everything (this is sanity — if existing tests are broken, stop and report).

```bash
cd /Users/liujiming/web/yuanju/frontend && npm run build && node --test --experimental-strip-types tests/*.test.mjs
```
Expected: PASS.

- [ ] **Step 3: No commit yet** — baseline is sanity only.

---

## Task 1: Add testcontainers-go dependency

**Files:**
- Modify: `backend/go.mod`, `backend/go.sum`
- Test: `backend/internal/repository/cleanup_integration_test.go` (just the smoke test for now)

- [ ] **Step 1: Add deps**

```bash
cd /Users/liujiming/web/yuanju/backend && \
  go get github.com/testcontainers/testcontainers-go && \
  go get github.com/testcontainers/testcontainers-go/modules/postgres && \
  go mod tidy
```
Expected: deps added to go.mod / go.sum. No build failures.

- [ ] **Step 2: Write smoke test that spins up a Postgres container**

Create `backend/internal/repository/cleanup_integration_test.go`:

```go
package repository_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/testcontainers/testcontainers-go"
	tcpg "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// spinUpPG 启一个临时 Postgres 容器，返回 *sql.DB 和清理 fn。
// 跨 integration test 复用。
func spinUpPG(t *testing.T) (*sql.DB, func()) {
	t.Helper()
	ctx := context.Background()

	pg, err := tcpg.Run(ctx,
		"postgres:16-alpine",
		tcpg.WithDatabase("yuanju_test"),
		tcpg.WithUsername("test"),
		tcpg.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second),
		),
	)
	if err != nil {
		t.Fatalf("start postgres container: %v", err)
	}

	dsn, err := pg.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		_ = pg.Terminate(ctx)
		t.Fatalf("conn string: %v", err)
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		_ = pg.Terminate(ctx)
		t.Fatalf("open db: %v", err)
	}
	if err := db.Ping(); err != nil {
		_ = pg.Terminate(ctx)
		t.Fatalf("ping db: %v", err)
	}

	cleanup := func() {
		_ = db.Close()
		_ = pg.Terminate(ctx)
	}
	return db, cleanup
}

func TestPostgresContainerSmoke(t *testing.T) {
	db, cleanup := spinUpPG(t)
	defer cleanup()

	var got int
	if err := db.QueryRow(`SELECT 1`).Scan(&got); err != nil {
		t.Fatalf("query: %v", err)
	}
	if got != 1 {
		t.Fatalf("want 1, got %d", got)
	}
}
```

- [ ] **Step 3: Run smoke test**

```bash
cd /Users/liujiming/web/yuanju/backend && go test ./internal/repository/ -run TestPostgresContainerSmoke -v
```
Expected: PASS (容器启动 + `SELECT 1` 返回 1)。第一次会拉镜像，耗时 30-60s。

如果 Docker daemon 未启动 → t.Skip 不在场，这一步会 fail；工程师需要先 `docker info` 确认 daemon 在跑。

- [ ] **Step 4: Commit**

```bash
git -C /Users/liujiming/web/yuanju add backend/go.mod backend/go.sum backend/internal/repository/cleanup_integration_test.go
git -C /Users/liujiming/web/yuanju commit -m "deps: add testcontainers-go for cleanup integration tests"
```

---

## Task 2: DDL — token_usage_logs_monthly + algo_config seeds

**Files:**
- Modify: `backend/pkg/database/database.go`

- [ ] **Step 1: 找到 database.go 中"✅ 数据库迁移完成"之前的位置**

```bash
grep -n "数据库迁移完成" /Users/liujiming/web/yuanju/backend/pkg/database/database.go
```
Expected: 找到 `log.Println("✅ 数据库迁移完成")` 的行号。新迁移块插在那一行之前。

- [ ] **Step 2: 在 "数据库迁移完成" 行之前追加 token_usage_logs_monthly DDL**

把这段插入到 `log.Println("✅ 数据库迁移完成")` **前面**：

```go
	// 增量迁移 (data-bloat-defense)：token 用量月度汇总表
	monthlyMigration := `
CREATE TABLE IF NOT EXISTS token_usage_logs_monthly (
    user_id            UUID         NOT NULL,
    model              VARCHAR(100) NOT NULL,
    year_month         CHAR(7)      NOT NULL,
    call_count         BIGINT       NOT NULL DEFAULT 0,
    prompt_tokens      BIGINT       NOT NULL DEFAULT 0,
    completion_tokens  BIGINT       NOT NULL DEFAULT 0,
    reasoning_tokens   BIGINT       NOT NULL DEFAULT 0,
    total_tokens       BIGINT       NOT NULL DEFAULT 0,
    aggregated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, model, year_month)
);
CREATE INDEX IF NOT EXISTS idx_token_usage_logs_monthly_ym
  ON token_usage_logs_monthly (year_month);`
	if _, err := DB.Exec(monthlyMigration); err != nil {
		log.Fatalf("增量迁移失败 (token_usage_logs_monthly): %v", err)
	}

	// 增量迁移 (data-bloat-defense)：注入清理 cron 默认 algo_config 键
	cleanupSeed := `
INSERT INTO algo_config (key, value, description) VALUES
  ('cleanup_enabled',         'true', '是否启用自动数据清理任务'),
  ('cleanup_retention_days',  '90',   'AI 缓存表与请求日志的保留天数，超期自动删除'),
  ('cleanup_run_hour',        '3',    '每日清理任务执行时刻（24h 制小时，0-23）')
ON CONFLICT (key) DO NOTHING;`
	if _, err := DB.Exec(cleanupSeed); err != nil {
		log.Fatalf("增量迁移失败 (cleanup algo_config seed): %v", err)
	}
```

- [ ] **Step 3: 验证编译通过**

```bash
cd /Users/liujiming/web/yuanju/backend && go build ./...
```
Expected: 无 error。

- [ ] **Step 4: 启动 backend 跑一次迁移（手动验证）**

```bash
cd /Users/liujiming/web/yuanju/backend && go run ./cmd/api 2>&1 | head -20
```
启动后看到 `✅ 数据库迁移完成` 一行说明 DDL 全部 OK。看到后 Ctrl+C 退出。

或者，跑迁移单测（如果有），或在 docker-compose 环境下 `docker-compose up backend` 看日志。

> 如果工程师无法访问开发 DB，跳过这一步，等 Task 7 的集成测试验证。

- [ ] **Step 5: Commit**

```bash
git -C /Users/liujiming/web/yuanju add backend/pkg/database/database.go
git -C /Users/liujiming/web/yuanju commit -m "feat(db): add token_usage_logs_monthly + cleanup algo_config seeds"
```

---

## Task 3: 7 个 repo `Delete*OlderThan` 方法

每个都是「单条 DELETE WHERE created_at < $1」+ 返回 `RowsAffected`。集中在一个 task，分 7 个小步骤。

**Files:**
- Modify: `backend/internal/repository/repository.go`
- Modify: `backend/internal/repository/polished_report_repo.go`
- Modify: `backend/internal/repository/liunian_repository.go`
- Modify: `backend/internal/repository/past_events_repository.go`
- Modify: `backend/internal/repository/dayun_summary_repository.go`
- Modify: `backend/internal/repository/compatibility_repository.go`
- Modify: `backend/internal/repository/admin_repository.go`

- [ ] **Step 1: 在 `repository.go` 文件末尾追加 `DeleteAIReportsOlderThan`**

```go
// DeleteAIReportsOlderThan 删除 created_at 早于 cutoff 的 ai_reports 行。
// 返回删除条数。
func DeleteAIReportsOlderThan(cutoff time.Time) (int64, error) {
	res, err := database.DB.Exec(`DELETE FROM ai_reports WHERE created_at < $1`, cutoff)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}
```

确认 `repository.go` 顶部 import 包含 `"time"`，没有则加。

- [ ] **Step 2: `polished_report_repo.go` 末尾追加**

```go
// DeletePolishedReportsOlderThan 删除超期润色报告。
func DeletePolishedReportsOlderThan(cutoff time.Time) (int64, error) {
	res, err := database.DB.Exec(`DELETE FROM ai_polished_reports WHERE created_at < $1`, cutoff)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}
```

确认 `polished_report_repo.go` import `"time"`。

- [ ] **Step 3: `liunian_repository.go` 末尾追加**

```go
// DeleteLiunianReportsOlderThan 删除超期流年报告。
func DeleteLiunianReportsOlderThan(cutoff time.Time) (int64, error) {
	res, err := database.DB.Exec(`DELETE FROM ai_liunian_reports WHERE created_at < $1`, cutoff)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}
```

确认 import `"time"`。

- [ ] **Step 4: `past_events_repository.go` 末尾追加**

```go
// DeletePastEventsOlderThan 删除超期过往事件缓存。
func DeletePastEventsOlderThan(cutoff time.Time) (int64, error) {
	res, err := database.DB.Exec(`DELETE FROM ai_past_events WHERE created_at < $1`, cutoff)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}
```

确认 import `"time"`。

- [ ] **Step 5: `dayun_summary_repository.go` 末尾追加**

```go
// DeleteDayunSummariesOlderThan 删除超期大运 summary。
func DeleteDayunSummariesOlderThan(cutoff time.Time) (int64, error) {
	res, err := database.DB.Exec(`DELETE FROM ai_dayun_summaries WHERE created_at < $1`, cutoff)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}
```

确认 import `"time"`。

- [ ] **Step 6: `compatibility_repository.go` 末尾追加**

```go
// DeleteAICompatibilityReportsOlderThan 删除超期合盘 AI 报告（不动业务表）。
func DeleteAICompatibilityReportsOlderThan(cutoff time.Time) (int64, error) {
	res, err := database.DB.Exec(`DELETE FROM ai_compatibility_reports WHERE created_at < $1`, cutoff)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}
```

确认 import `"time"`。

- [ ] **Step 7: `admin_repository.go` 末尾追加**

```go
// DeleteRequestLogsOlderThan 删除超期 AI 请求日志。
func DeleteRequestLogsOlderThan(cutoff time.Time) (int64, error) {
	res, err := database.DB.Exec(`DELETE FROM ai_requests_log WHERE created_at < $1`, cutoff)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}
```

确认 import `"time"`。

- [ ] **Step 8: 验证编译**

```bash
cd /Users/liujiming/web/yuanju/backend && go build ./... && go vet ./...
```
Expected: 无 error / 无 warning。

- [ ] **Step 9: Commit**

```bash
git -C /Users/liujiming/web/yuanju add backend/internal/repository/
git -C /Users/liujiming/web/yuanju commit -m "feat(repo): add Delete*OlderThan for 7 cleanup-eligible tables"
```

---

## Task 4: token rollup `RollupClosedMonthsAndDelete`

**Files:**
- Modify: `backend/internal/repository/token_usage_repository.go`

- [ ] **Step 1: 在 `token_usage_repository.go` 末尾追加 RollupReport 类型与方法**

```go
// RollupReport 是 token rollup 一次运行的结构化结果。
type RollupReport struct {
	MonthsAggregated      int
	RowsInsertedOrUpdated int64
	SourceRowsDeleted     int64
}

// OrphanUserUUID 用作 user_id IS NULL 行的 sentinel（已注销用户的合并桶）。
const OrphanUserUUID = "00000000-0000-0000-0000-000000000000"

// RollupClosedMonthsAndDelete 把 token_usage_logs 里所有已闭合月份（早于本月 1 号 00:00）
// 的行按 (user_id, model, year_month) 聚合写入 token_usage_logs_monthly，再删除源行。
// 整个流程在单事务里，幂等可重跑。
func RollupClosedMonthsAndDelete() (RollupReport, error) {
	var rep RollupReport

	tx, err := database.DB.Begin()
	if err != nil {
		return rep, err
	}
	defer func() { _ = tx.Rollback() }() // 提交后是 no-op

	const insertSQL = `
INSERT INTO token_usage_logs_monthly (
    user_id, model, year_month,
    call_count, prompt_tokens, completion_tokens,
    reasoning_tokens, total_tokens, aggregated_at
)
SELECT
    COALESCE(user_id, '` + OrphanUserUUID + `'::uuid) AS user_id,
    model,
    to_char(created_at, 'YYYY-MM')      AS year_month,
    COUNT(*)                            AS call_count,
    COALESCE(SUM(prompt_tokens), 0)     AS prompt_tokens,
    COALESCE(SUM(completion_tokens), 0) AS completion_tokens,
    COALESCE(SUM(reasoning_tokens), 0)  AS reasoning_tokens,
    COALESCE(SUM(total_tokens), 0)      AS total_tokens,
    NOW()                               AS aggregated_at
FROM token_usage_logs
WHERE created_at < date_trunc('month', NOW())
GROUP BY COALESCE(user_id, '` + OrphanUserUUID + `'::uuid), model, year_month
ON CONFLICT (user_id, model, year_month) DO UPDATE SET
    call_count        = EXCLUDED.call_count,
    prompt_tokens     = EXCLUDED.prompt_tokens,
    completion_tokens = EXCLUDED.completion_tokens,
    reasoning_tokens  = EXCLUDED.reasoning_tokens,
    total_tokens      = EXCLUDED.total_tokens,
    aggregated_at     = NOW();
`
	insertRes, err := tx.Exec(insertSQL)
	if err != nil {
		return rep, err
	}
	rep.RowsInsertedOrUpdated, _ = insertRes.RowsAffected()

	// 统计 affected month 数（聚合查 distinct，便于日志）
	if err := tx.QueryRow(`
SELECT COUNT(DISTINCT to_char(created_at, 'YYYY-MM'))
FROM token_usage_logs
WHERE created_at < date_trunc('month', NOW())
`).Scan(&rep.MonthsAggregated); err != nil {
		return rep, err
	}

	delRes, err := tx.Exec(`DELETE FROM token_usage_logs WHERE created_at < date_trunc('month', NOW())`)
	if err != nil {
		return rep, err
	}
	rep.SourceRowsDeleted, _ = delRes.RowsAffected()

	if err := tx.Commit(); err != nil {
		return rep, err
	}
	return rep, nil
}
```

- [ ] **Step 2: 验证编译**

```bash
cd /Users/liujiming/web/yuanju/backend && go build ./... && go vet ./...
```
Expected: 无 error。

- [ ] **Step 3: Commit**

```bash
git -C /Users/liujiming/web/yuanju add backend/internal/repository/token_usage_repository.go
git -C /Users/liujiming/web/yuanju commit -m "feat(repo): add RollupClosedMonthsAndDelete for token usage"
```

---

## Task 5: algo_config typed getter

**Files:**
- Modify: `backend/internal/service/algo_config_service.go`

- [ ] **Step 1: 在 `algo_config_service.go` 末尾追加新类型 + getter**

```go
// CleanupConfig 是清理任务的运行时配置（每次 tick 从 algo_config 表实时读取）。
type CleanupConfig struct {
	Enabled       bool
	RetentionDays int // 已 clamp 到 [1, 3650]
	RunHour       int // 已 clamp 到 [0, 23]
}

// GetCleanupConfig 实时读取 algo_config 中清理相关键。
// 缺失的键走默认值；异常值 clamp 到合法区间。
func GetCleanupConfig() (CleanupConfig, error) {
	cfg := CleanupConfig{Enabled: true, RetentionDays: 90, RunHour: 3}

	rows, err := repository.GetAllAlgoConfig()
	if err != nil {
		return cfg, err
	}
	for _, r := range rows {
		switch r.Key {
		case "cleanup_enabled":
			cfg.Enabled = r.Value == "true" || r.Value == "1"
		case "cleanup_retention_days":
			if v, err := strconv.Atoi(r.Value); err == nil {
				cfg.RetentionDays = clampInt(v, 1, 3650)
			}
		case "cleanup_run_hour":
			if v, err := strconv.Atoi(r.Value); err == nil {
				cfg.RunHour = clampInt(v, 0, 23)
			}
		}
	}
	return cfg, nil
}

func clampInt(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
```

确认 import 已包含 `"strconv"`（既有 `LoadAlgoConfig` 已用过）。`repository` package 也已 imported。

- [ ] **Step 2: 验证编译**

```bash
cd /Users/liujiming/web/yuanju/backend && go build ./...
```
Expected: 无 error。

- [ ] **Step 3: Commit**

```bash
git -C /Users/liujiming/web/yuanju add backend/internal/service/algo_config_service.go
git -C /Users/liujiming/web/yuanju commit -m "feat(service): GetCleanupConfig live-reads algo_config with clamp"
```

---

## Task 6: CleanupService (核心) — RED 测试先行

**Files:**
- Create: `backend/internal/service/cleanup_service_test.go`

- [ ] **Step 1: 写失败测试（RED commit）**

Create `backend/internal/service/cleanup_service_test.go`:

```go
package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"yuanju/internal/service"
)

// fakeClock 让测试固定 Now()
type fakeClock struct{ now time.Time }

func (f fakeClock) Now() time.Time { return f.now }

// stubCleaner 记录是否被调用、用什么 cutoff，返回预设值
type stubCleaner struct {
	called     bool
	gotCutoff  time.Time
	returnRows int64
	returnErr  error
}

func (s *stubCleaner) Delete(_ context.Context, cutoff time.Time) (int64, error) {
	s.called = true
	s.gotCutoff = cutoff
	return s.returnRows, s.returnErr
}

type stubRollup struct {
	called bool
	report service.RollupReport
	err    error
}

func (s *stubRollup) Rollup(_ context.Context) (service.RollupReport, error) {
	s.called = true
	return s.report, s.err
}

func makeService(t *testing.T, cfg service.CleanupConfig) (*service.CleanupService, *stubCleaner, *stubCleaner, *stubCleaner, *stubCleaner, *stubCleaner, *stubCleaner, *stubCleaner, *stubRollup) {
	t.Helper()
	aiReports := &stubCleaner{}
	polished := &stubCleaner{}
	liunian := &stubCleaner{}
	pastEvents := &stubCleaner{}
	dayunSummary := &stubCleaner{}
	compat := &stubCleaner{}
	reqLogs := &stubCleaner{}
	rollup := &stubRollup{}

	deps := service.CleanupDeps{
		AIReports:     aiReports.Delete,
		Polished:      polished.Delete,
		Liunian:       liunian.Delete,
		PastEvents:    pastEvents.Delete,
		DayunSummary:  dayunSummary.Delete,
		CompatReports: compat.Delete,
		RequestLogs:   reqLogs.Delete,
		TokenRollup:   rollup.Rollup,
	}
	svc := service.NewCleanupServiceForTest(
		deps,
		func() (service.CleanupConfig, error) { return cfg, nil },
		fakeClock{now: time.Date(2026, 5, 18, 3, 0, 0, 0, time.UTC)},
	)
	return svc, aiReports, polished, liunian, pastEvents, dayunSummary, compat, reqLogs, rollup
}

func TestRunOnce_DisabledShortCircuits(t *testing.T) {
	svc, ar, _, _, _, _, _, _, ro := makeService(t, service.CleanupConfig{Enabled: false, RetentionDays: 90, RunHour: 3})
	rep := svc.RunOnce(context.Background())
	if ar.called {
		t.Errorf("ai_reports cleaner should NOT be called when disabled")
	}
	if ro.called {
		t.Errorf("rollup should NOT be called when disabled")
	}
	if len(rep.Tables) != 0 {
		t.Errorf("disabled run should have empty Tables, got %d", len(rep.Tables))
	}
}

func TestRunOnce_PassesCutoffFromRetentionAndClock(t *testing.T) {
	svc, ar, pol, li, pe, ds, cp, rl, _ := makeService(t, service.CleanupConfig{Enabled: true, RetentionDays: 90, RunHour: 3})
	svc.RunOnce(context.Background())

	want := time.Date(2026, 5, 18, 3, 0, 0, 0, time.UTC).Add(-90 * 24 * time.Hour)
	for _, c := range []*stubCleaner{ar, pol, li, pe, ds, cp, rl} {
		if !c.called {
			t.Errorf("cleaner not called")
		}
		if !c.gotCutoff.Equal(want) {
			t.Errorf("cutoff = %v, want %v", c.gotCutoff, want)
		}
	}
}

func TestRunOnce_ErrorIsolation(t *testing.T) {
	svc, ar, pol, _, _, _, _, _, _ := makeService(t, service.CleanupConfig{Enabled: true, RetentionDays: 90, RunHour: 3})
	ar.returnErr = errors.New("simulated DB error")

	rep := svc.RunOnce(context.Background())

	if !pol.called {
		t.Errorf("polished cleaner should still run despite ai_reports failure")
	}
	if len(rep.Tables) != 7 {
		t.Fatalf("expected 7 TableResult entries, got %d", len(rep.Tables))
	}
	if rep.Tables[0].Err == nil {
		t.Errorf("expected first table (ai_reports) to have Err set")
	}
	for i := 1; i < 7; i++ {
		if rep.Tables[i].Err != nil {
			t.Errorf("Tables[%d].Err = %v, want nil", i, rep.Tables[i].Err)
		}
	}
}

func TestRunOnce_RetentionClamp(t *testing.T) {
	// 模拟 algo_config 已经 clamp 过的极小值（这里直接传 RetentionDays=1 验证传递正确）
	svc, ar, _, _, _, _, _, _, _ := makeService(t, service.CleanupConfig{Enabled: true, RetentionDays: 1, RunHour: 3})
	svc.RunOnce(context.Background())
	want := time.Date(2026, 5, 18, 3, 0, 0, 0, time.UTC).Add(-1 * 24 * time.Hour)
	if !ar.gotCutoff.Equal(want) {
		t.Errorf("cutoff = %v, want %v", ar.gotCutoff, want)
	}
}
```

- [ ] **Step 2: 跑测试验证 RED**

```bash
cd /Users/liujiming/web/yuanju/backend && go test ./internal/service/ -run TestRunOnce -v
```
Expected: FAIL with compile errors like `undefined: service.CleanupConfig` / `undefined: service.CleanupDeps` / `undefined: service.NewCleanupServiceForTest`. 这是预期的 RED。

- [ ] **Step 3: Commit RED**

```bash
git -C /Users/liujiming/web/yuanju add backend/internal/service/cleanup_service_test.go
git -C /Users/liujiming/web/yuanju commit -m "test(cleanup): RED — failing tests for CleanupService"
```

---

## Task 7: CleanupService GREEN — 实现 cleanup_service.go

**Files:**
- Create: `backend/internal/service/cleanup_service.go`

- [ ] **Step 1: 写 cleanup_service.go**

Create `backend/internal/service/cleanup_service.go`:

```go
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
	deps    CleanupDeps
	config  func() (CleanupConfig, error)
	clock   Clock
	logger  *log.Logger
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
		cfg, _ := s.config()
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
	type out struct {
		StartedAt string       `json:"started_at"`
		DurMs     int64        `json:"duration_ms"`
		Tables    []tableOut   `json:"tables"`
		Rollup    RollupReport `json:"rollup"`
		Errors    []string     `json:"errors,omitempty"`
	}
	o := out{
		StartedAt: rep.StartedAt.Format(time.RFC3339),
		DurMs:     rep.Duration.Milliseconds(),
		Rollup:    rep.Rollup,
		Errors:    rep.Errors,
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

// RollupReport 类型定义在 repository 包但 service 也要导出。
// 这里 alias 避免 service 外部依赖 repository 类型暴露。
type RollupReport = struct {
	MonthsAggregated      int
	RowsInsertedOrUpdated int64
	SourceRowsDeleted     int64
	Err                   error
}
```

注意：`RollupReport` 在 service 包定义为独立结构（不是 repository.RollupReport 的别名）。这样 service 层不需要在公共接口里依赖 repository 类型。Task 8 的 main.go 适配器会把 repository.RollupReport 转成 service.RollupReport。

- [ ] **Step 2: 跑测试验证 GREEN**

```bash
cd /Users/liujiming/web/yuanju/backend && go test ./internal/service/ -run TestRunOnce -v
```
Expected: 4 个测试全 PASS。

- [ ] **Step 3: 跑全量测试确保没破坏既有**

```bash
cd /Users/liujiming/web/yuanju/backend && go test ./...
```
Expected: 全 PASS。

- [ ] **Step 4: Commit GREEN**

```bash
git -C /Users/liujiming/web/yuanju add backend/internal/service/cleanup_service.go
git -C /Users/liujiming/web/yuanju commit -m "feat(cleanup): implement CleanupService (RunOnce + scheduler)"
```

---

## Task 8: 集成测试 — token rollup SQL 正确性

**Files:**
- Modify: `backend/internal/repository/cleanup_integration_test.go` (新增 TestTokenRollupRollupClosedMonthsAndDelete)

- [ ] **Step 1: 在 cleanup_integration_test.go 中追加 rollup 集成测试**

把这段追加到 `cleanup_integration_test.go` 已有 `TestPostgresContainerSmoke` 之后：

```go
// runMigrationsInContainer 把 token_usage_logs / token_usage_logs_monthly 的 DDL
// 在容器里跑一遍。简化版：只声明本测试需要的两张表，不跑整个 database.Migrate()。
func runMigrationsInContainer(t *testing.T, db *sql.DB) {
	t.Helper()
	ddl := []string{
		`CREATE EXTENSION IF NOT EXISTS "pgcrypto"`,
		`CREATE TABLE token_usage_logs (
			id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id           UUID,
			chart_id          UUID,
			call_type         VARCHAR(50) NOT NULL,
			model             VARCHAR(100),
			provider_id       UUID,
			prompt_tokens     INT NOT NULL DEFAULT 0,
			completion_tokens INT NOT NULL DEFAULT 0,
			total_tokens      INT NOT NULL DEFAULT 0,
			reasoning_tokens  INT NOT NULL DEFAULT 0,
			cache_hit_tokens  INT NOT NULL DEFAULT 0,
			cache_miss_tokens INT NOT NULL DEFAULT 0,
			input_content     TEXT,
			output_content    TEXT,
			created_at        TIMESTAMPTZ DEFAULT NOW()
		)`,
		`CREATE TABLE token_usage_logs_monthly (
			user_id            UUID         NOT NULL,
			model              VARCHAR(100) NOT NULL,
			year_month         CHAR(7)      NOT NULL,
			call_count         BIGINT       NOT NULL DEFAULT 0,
			prompt_tokens      BIGINT       NOT NULL DEFAULT 0,
			completion_tokens  BIGINT       NOT NULL DEFAULT 0,
			reasoning_tokens   BIGINT       NOT NULL DEFAULT 0,
			total_tokens       BIGINT       NOT NULL DEFAULT 0,
			aggregated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
			PRIMARY KEY (user_id, model, year_month)
		)`,
	}
	for _, s := range ddl {
		if _, err := db.Exec(s); err != nil {
			t.Fatalf("ddl %q: %v", s[:60], err)
		}
	}
}

func TestRollupClosedMonthsAndDelete(t *testing.T) {
	db, cleanup := spinUpPG(t)
	defer cleanup()
	runMigrationsInContainer(t, db)

	// 把全局 database.DB 暂时指向容器内的 db
	// 因为 RollupClosedMonthsAndDelete 内部用的就是 database.DB
	original := swapGlobalDB(t, db)
	defer original()

	// 插入 5 个闭合月份 + 当月数据
	userA := "11111111-1111-1111-1111-111111111111"
	userB := "22222222-2222-2222-2222-222222222222"
	models := []string{"gpt-4o", "deepseek-v3"}

	now := time.Now().UTC()
	closedMonths := []time.Time{
		monthAgo(now, 1), monthAgo(now, 2), monthAgo(now, 3),
		monthAgo(now, 4), monthAgo(now, 5),
	}
	for _, ts := range closedMonths {
		for _, uid := range []string{userA, userB} {
			for _, m := range models {
				for i := 0; i < 3; i++ {
					if _, err := db.Exec(`
INSERT INTO token_usage_logs (user_id, call_type, model, prompt_tokens, completion_tokens, total_tokens, created_at)
VALUES ($1, 'report', $2, 100, 200, 300, $3)`, uid, m, ts); err != nil {
						t.Fatalf("insert closed: %v", err)
					}
				}
			}
		}
	}
	// 1 行孤儿用户（user_id IS NULL）
	if _, err := db.Exec(`
INSERT INTO token_usage_logs (user_id, call_type, model, prompt_tokens, completion_tokens, total_tokens, created_at)
VALUES (NULL, 'report', 'gpt-4o', 50, 75, 125, $1)`, closedMonths[0]); err != nil {
		t.Fatalf("insert orphan: %v", err)
	}
	// 5 行当月数据（不应被汇总）
	currentMonth := time.Date(now.Year(), now.Month(), 1, 12, 0, 0, 0, time.UTC).Add(-1 * time.Hour) // 实际是上月最后一刻，避免歧义
	// 重写：直接用 now 本身（当月）
	for i := 0; i < 5; i++ {
		if _, err := db.Exec(`
INSERT INTO token_usage_logs (user_id, call_type, model, prompt_tokens, completion_tokens, total_tokens, created_at)
VALUES ($1, 'report', 'gpt-4o', 10, 20, 30, NOW())`, userA); err != nil {
			t.Fatalf("insert current: %v", err)
		}
	}
	_ = currentMonth // 抑制 unused

	// 跑 rollup
	rep, err := repository.RollupClosedMonthsAndDelete()
	if err != nil {
		t.Fatalf("rollup: %v", err)
	}

	// 断言：5 个月被聚合
	if rep.MonthsAggregated != 5 {
		t.Errorf("MonthsAggregated = %d, want 5", rep.MonthsAggregated)
	}
	// 源行被删 = 5 月 * 2 用户 * 2 模型 * 3 条 + 1 孤儿 = 61
	wantDeleted := int64(5*2*2*3 + 1)
	if rep.SourceRowsDeleted != wantDeleted {
		t.Errorf("SourceRowsDeleted = %d, want %d", rep.SourceRowsDeleted, wantDeleted)
	}
	// monthly 行数：5 月 * (2 用户 * 2 模型 + 1 孤儿在 first month) = 5*4 + 1 = 21
	var monthlyCount int
	if err := db.QueryRow(`SELECT COUNT(*) FROM token_usage_logs_monthly`).Scan(&monthlyCount); err != nil {
		t.Fatalf("count monthly: %v", err)
	}
	if monthlyCount != 21 {
		t.Errorf("monthly row count = %d, want 21", monthlyCount)
	}
	// 源表只剩当月 5 行
	var srcCount int
	if err := db.QueryRow(`SELECT COUNT(*) FROM token_usage_logs`).Scan(&srcCount); err != nil {
		t.Fatalf("count src: %v", err)
	}
	if srcCount != 5 {
		t.Errorf("token_usage_logs row count = %d, want 5", srcCount)
	}

	// 断言聚合数值正确（任选 userA + gpt-4o + 1 月前那一行）
	var promptSum, completionSum, callCount int64
	ym := closedMonths[0].Format("2006-01")
	if err := db.QueryRow(`
SELECT prompt_tokens, completion_tokens, call_count
FROM token_usage_logs_monthly
WHERE user_id = $1 AND model = 'gpt-4o' AND year_month = $2`, userA, ym).Scan(&promptSum, &completionSum, &callCount); err != nil {
		t.Fatalf("read aggregate: %v", err)
	}
	if promptSum != 300 || completionSum != 600 || callCount != 3 {
		t.Errorf("aggregate userA gpt-4o %s = (prompt=%d, completion=%d, count=%d), want (300, 600, 3)",
			ym, promptSum, completionSum, callCount)
	}

	// 孤儿行落到 sentinel UUID 桶
	var orphanCount int64
	if err := db.QueryRow(`
SELECT call_count FROM token_usage_logs_monthly
WHERE user_id = $1 AND model = 'gpt-4o' AND year_month = $2`, repository.OrphanUserUUID, ym).Scan(&orphanCount); err != nil {
		t.Fatalf("read orphan bucket: %v", err)
	}
	if orphanCount != 1 {
		t.Errorf("orphan call_count = %d, want 1", orphanCount)
	}

	// 幂等：再跑一次，monthly 行数不变，源表还剩 5 行
	rep2, err := repository.RollupClosedMonthsAndDelete()
	if err != nil {
		t.Fatalf("rollup re-run: %v", err)
	}
	if rep2.SourceRowsDeleted != 0 {
		t.Errorf("re-run SourceRowsDeleted = %d, want 0", rep2.SourceRowsDeleted)
	}
	var monthlyCount2 int
	if err := db.QueryRow(`SELECT COUNT(*) FROM token_usage_logs_monthly`).Scan(&monthlyCount2); err != nil {
		t.Fatalf("count monthly re-run: %v", err)
	}
	if monthlyCount2 != 21 {
		t.Errorf("after re-run monthly count = %d, want 21 (idempotent)", monthlyCount2)
	}
}

// monthAgo 返回 now 当前时间的"几个月前"那一刻（保留 day/hour）。
// 用于生成跨月份的测试数据。
func monthAgo(now time.Time, n int) time.Time {
	return time.Date(now.Year(), now.Month()-time.Month(n), 15, 12, 0, 0, 0, time.UTC)
}

// swapGlobalDB 把 database.DB 临时替换为容器 db，返回还原 fn。
// 跨包修改全局变量需要 database 包导出 SetDBForTest（见 Task 8 Step 0）。
func swapGlobalDB(t *testing.T, db *sql.DB) func() {
	t.Helper()
	original := database.DB
	database.DB = db
	return func() { database.DB = original }
}
```

测试需要新增几个 import：

```go
import (
	"yuanju/internal/repository"
	"yuanju/pkg/database"
)
```

- [ ] **Step 2: 验证测试运行**

```bash
cd /Users/liujiming/web/yuanju/backend && go test ./internal/repository/ -run TestRollupClosedMonthsAndDelete -v
```
Expected: PASS（如果 Docker daemon 没起，会因 testcontainer 启动失败而 fail —— 需要先启 docker）。

测试逻辑要点：
- 插入 5 个闭合月 × 2 用户 × 2 模型 × 3 调用 = 60 行 + 1 孤儿 + 5 当月 = 66 行
- rollup 后期待：21 monthly 行 + 5 当月源行；删 61 行
- 再跑一次，monthly 行数不变，源表 5 行不变（幂等）

- [ ] **Step 3: Commit**

```bash
git -C /Users/liujiming/web/yuanju add backend/internal/repository/cleanup_integration_test.go
git -C /Users/liujiming/web/yuanju commit -m "test(cleanup): integration test for token rollup SQL correctness"
```

---

## Task 9: main.go 接线 — scheduler + --cleanup-once CLI flag

**Files:**
- Modify: `backend/cmd/api/main.go`

- [ ] **Step 1: 改写 main.go 顶部，引入 flag、context、cleanup 适配器**

打开 `backend/cmd/api/main.go`。在 `import (...)` 块加入：

```go
import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
	"yuanju/configs"
	"yuanju/internal/handler"
	"yuanju/internal/middleware"
	"yuanju/internal/repository"
	"yuanju/internal/service"
	"yuanju/pkg/database"
	"yuanju/pkg/seed"

	"github.com/gin-gonic/gin"
)
```

- [ ] **Step 2: 替换 main() 的开头，在加载配置/连接 DB 之前先解析 flag**

把 `func main()` 改成：

```go
func main() {
	cleanupOnce := flag.Bool("cleanup-once", false, "运行一次清理任务后退出，不启动 HTTP server")
	flag.Parse()

	// 加载配置
	configs.Load()

	// 连接数据库
	database.Connect()
	database.Migrate()

	// 确保 logo 上传目录存在
	if err := os.MkdirAll(filepath.Join(configs.AppConfig.UploadDir, "brand-logos"), 0755); err != nil {
		log.Fatalf("创建上传目录失败: %v", err)
	}

	// 种子数据
	seed.SeedLLMProviders()
	seed.SeedLLMPrices()

	// 加载算法配置
	if err := service.LoadAlgoConfig(); err != nil {
		log.Printf("算法配置加载失败（使用默认值）: %v", err)
	}

	// 构造 cleanup 服务（在判断 --cleanup-once 之前，因为两种模式都用到）
	cleanupSvc := service.NewCleanupService(makeCleanupDeps())

	if *cleanupOnce {
		rep := cleanupSvc.RunOnce(context.Background())
		fmt.Println(string(service.MarshalRunReport(rep)))
		os.Exit(0)
	}

	// 起 cleanup scheduler（后台 goroutine）
	schedCtx, cancelSched := context.WithCancel(context.Background())
	go cleanupSvc.StartScheduler(schedCtx)
	defer cancelSched()

	// SIGTERM/SIGINT 触发 cancelSched 让 scheduler 干净退出
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)
		<-sigCh
		log.Println("收到退出信号，取消 cleanup scheduler")
		cancelSched()
	}()

	// 初始化路由
	r := gin.Default()
	r.MaxMultipartMemory = 4 << 20
	r.Use(middleware.CORS())
	r.Static("/static/uploads/brand-logos", filepath.Join(configs.AppConfig.UploadDir, "brand-logos"))

	// ... 以下保留原有路由注册代码（健康检查 + api 路由组），不要改动 ...
```

**注意：** 原 main.go 中从 `r := gin.Default()` 到 `r.Run` 的所有路由注册代码**不要动**；只是把新代码插在它之前。

- [ ] **Step 3: 在 main.go 文件底部追加 makeCleanupDeps 适配器**

放在 `func main()` 之后：

```go
// makeCleanupDeps 把 repository 层的具体函数 wrap 成 service.CleanupDeps，
// 同时把 repository.RollupReport 适配成 service.RollupReport（结构等价）。
func makeCleanupDeps() service.CleanupDeps {
	wrapTime := func(f func(time.Time) (int64, error)) service.TableCleaner {
		return func(_ context.Context, cutoff time.Time) (int64, error) {
			return f(cutoff)
		}
	}
	return service.CleanupDeps{
		AIReports:     wrapTime(repository.DeleteAIReportsOlderThan),
		Polished:      wrapTime(repository.DeletePolishedReportsOlderThan),
		Liunian:       wrapTime(repository.DeleteLiunianReportsOlderThan),
		PastEvents:    wrapTime(repository.DeletePastEventsOlderThan),
		DayunSummary:  wrapTime(repository.DeleteDayunSummariesOlderThan),
		CompatReports: wrapTime(repository.DeleteAICompatibilityReportsOlderThan),
		RequestLogs:   wrapTime(repository.DeleteRequestLogsOlderThan),
		TokenRollup: func(_ context.Context) (service.RollupReport, error) {
			r, err := repository.RollupClosedMonthsAndDelete()
			return service.RollupReport{
				MonthsAggregated:      r.MonthsAggregated,
				RowsInsertedOrUpdated: r.RowsInsertedOrUpdated,
				SourceRowsDeleted:     r.SourceRowsDeleted,
			}, err
		},
	}
}
```

- [ ] **Step 4: 验证编译**

```bash
cd /Users/liujiming/web/yuanju/backend && go build ./...
```
Expected: 无 error。

- [ ] **Step 5: 验证 --cleanup-once flag 行为（不接 DB）**

```bash
cd /Users/liujiming/web/yuanju/backend && go run ./cmd/api --cleanup-once 2>&1 | head -30
```
Expected: 如果 DB 可连，打印 RunReport JSON 后退出。如果 DB 不可连，`database.Connect()` 会 fatal log；这是预期，因为这条路径就依赖 DB。

- [ ] **Step 6: Commit**

```bash
git -C /Users/liujiming/web/yuanju add backend/cmd/api/main.go
git -C /Users/liujiming/web/yuanju commit -m "feat(cleanup): wire scheduler + --cleanup-once CLI flag in main.go"
```

---

## Task 10: Frontend — AlgoConfigPage 暴露 3 个 cleanup keys

**Files:**
- Modify: `frontend/src/pages/admin/AlgoConfigPage.tsx`
- Create: `frontend/tests/cleanup-config.test.mjs`

`AlgoConfigPage.tsx` 已经把 `params` 数组里的全部 key 都渲染出来。新加的 3 个 key 会自动出现，**唯一需要做的是让它们有友好显示名**（通过现有的 `PARAM_LABELS` map）。

- [ ] **Step 1: 修改 AlgoConfigPage.tsx 中的 PARAM_LABELS**

找到 `const PARAM_LABELS: Record<string, string> = {...}` 块（约 line 24-28），改成：

```typescript
const PARAM_LABELS: Record<string, string> = {
  jixiong_jiHan_min: '极寒阈值（寒性元素最低数量）',
  jixiong_jiRe_min: '极热阈值（暖性元素最低数量）',
  jixiong_shenQiang_pct: '身强判定阈值（生助比例 %）',
  cleanup_enabled: '自动清理任务是否启用（true / false）',
  cleanup_retention_days: 'AI 缓存与请求日志保留天数（默认 90，clamp 到 [1, 3650]）',
  cleanup_run_hour: '清理任务每日执行时刻（小时，clamp 到 [0, 23]）',
}
```

- [ ] **Step 2: 写 grep 单测**

Create `frontend/tests/cleanup-config.test.mjs`:

```javascript
import { readFileSync } from 'node:fs'
import { join } from 'node:path'
import test from 'node:test'
import assert from 'node:assert/strict'

const root = new URL('..', import.meta.url).pathname

function read(path) {
  return readFileSync(join(root, path), 'utf8')
}

test('AlgoConfigPage exposes cleanup_enabled label', () => {
  const src = read('src/pages/admin/AlgoConfigPage.tsx')
  assert.match(src, /cleanup_enabled:/)
})

test('AlgoConfigPage exposes cleanup_retention_days label', () => {
  const src = read('src/pages/admin/AlgoConfigPage.tsx')
  assert.match(src, /cleanup_retention_days:/)
})

test('AlgoConfigPage exposes cleanup_run_hour label', () => {
  const src = read('src/pages/admin/AlgoConfigPage.tsx')
  assert.match(src, /cleanup_run_hour:/)
})
```

- [ ] **Step 3: 跑测试**

```bash
cd /Users/liujiming/web/yuanju/frontend && node --test --experimental-strip-types tests/cleanup-config.test.mjs
```
Expected: 3/3 PASS。

- [ ] **Step 4: 前端 build + lint**

```bash
cd /Users/liujiming/web/yuanju/frontend && npm run build && npm run lint
```
Expected: build 成功，lint 不增加新 error。

- [ ] **Step 5: Commit**

```bash
git -C /Users/liujiming/web/yuanju add frontend/src/pages/admin/AlgoConfigPage.tsx frontend/tests/cleanup-config.test.mjs
git -C /Users/liujiming/web/yuanju commit -m "feat(admin): surface 3 cleanup config keys in AlgoConfigPage"
```

---

## Task 11: 全量验证 + 手动验收

**Files:** none (verification only)

- [ ] **Step 1: 全后端测试**

```bash
cd /Users/liujiming/web/yuanju/backend && go test ./...
```
Expected: 全 PASS（包含 cleanup_service_test 4 个 + cleanup integration 2 个）。

- [ ] **Step 2: 全前端测试 + build**

```bash
cd /Users/liujiming/web/yuanju/frontend && \
  node --test --experimental-strip-types tests/*.test.mjs && \
  npm run build && \
  npm run lint
```
Expected: 全 PASS、build 成功。

- [ ] **Step 3: 手动跑 --cleanup-once 模式**（如果工程师有可用的 dev DB）

```bash
cd /Users/liujiming/web/yuanju/backend && go run ./cmd/api --cleanup-once
```
Expected: 打印形如：

```json
{
  "started_at": "2026-05-18T...Z",
  "duration_ms": <small number>,
  "tables": [
    {"name": "ai_reports", "deleted": 0},
    {"name": "ai_polished_reports", "deleted": 0},
    ...7 tables...
  ],
  "rollup": {
    "months_aggregated": 0,
    "rows_inserted_or_updated": 0,
    "source_rows_deleted": 0
  }
}
```

第一次跑 deleted 都是 0（开发库通常没有 90 天前的数据）；如果有则会显示数字。

- [ ] **Step 4: 手动验证 admin UI**

```bash
cd /Users/liujiming/web/yuanju && docker-compose up -d --build --force-recreate frontend backend
```

浏览器打开 admin 后台 → 算法参数页面 → 应看到 3 个新键 `cleanup_enabled` / `cleanup_retention_days` / `cleanup_run_hour`，可点编辑。

改 `cleanup_retention_days` 为 30 → 保存 → 下一次 scheduler tick 会按 30 天裁切。

- [ ] **Step 5: 检查日志格式**

```bash
docker-compose logs backend | grep '"evt":"cleanup_run"' | head -5
```
Expected: 每天 3 AM 会有一行 `cleanup_run` 结构化日志（如果 backend 已经过过 3 AM）。

- [ ] **Step 6: 不需要 commit**（这一步只是验证）。

---

## Spec Coverage Cross-Check

对照 `docs/superpowers/specs/2026-05-18-data-bloat-defense-design.md`：

| Spec 节 | 实现 task |
|---|---|
| §2 目标：6+1 张表按 created_at 删 | Task 3（7 个 repo 方法）+ Task 7（service 顺序调用） |
| §2 目标：token 按月汇总 | Task 4（rollup repo）+ Task 7（service 集成）+ Task 8（集成测试） |
| §2 目标：goroutine 每 24h 跑 | Task 7 StartScheduler + Task 9 main.go 启动 |
| §2 目标：algo_config 暴露阈值 | Task 2（seed 3 keys）+ Task 5（typed getter）+ Task 10（admin UI） |
| §2 目标：--cleanup-once CLI flag | Task 9 |
| §5.4 接口（8 个 cleaner + Clock） | Task 7 类型定义 |
| §5.5 monthly schema | Task 2 DDL |
| §5.6 3 个 algo_config keys + clamp | Task 2 seed + Task 5 clamp |
| §6.3 rollup SQL with COALESCE 孤儿 | Task 4 实现 + Task 8 集成测试断言孤儿桶 |
| §6.4 结构化日志 | Task 7 logRun |
| §7.1 错误隔离 | Task 6 RED test + Task 7 实现 |
| §7.2 幂等保证 | Task 8 集成测试 re-run 断言 |
| §7.3 --cleanup-once | Task 9 |
| §8.1 单测 4 个场景 | Task 6 |
| §8.2 集成测试 | Task 8 |
| §8.3 前端 grep | Task 10 |
| §8.4 手动验收 | Task 11 |

**无遗漏。**

---

## TDD 节奏总结

| Task | 类型 | Commits |
|---|---|---|
| 0 | branch + sanity | 0 |
| 1 | deps + smoke | 1 |
| 2 | DDL | 1 |
| 3 | 7 个 repo 方法（mechanical） | 1 |
| 4 | rollup repo（mechanical） | 1 |
| 5 | typed getter（mechanical） | 1 |
| 6 | **RED** service 单测 | 1 |
| 7 | **GREEN** service 实现 | 1 |
| 8 | 集成测试 | 1 |
| 9 | main.go 接线 | 1 |
| 10 | 前端 + grep test | 1 |
| 11 | 验证 | 0 |

**总计 10 个 commit**，每个 ≤ 200 行新增（除 Task 7 cleanup_service.go 约 230 行，仍可控）。

---

## 已知风险与注意点

1. **Docker daemon 必须运行** —— 否则 Task 1/8 的 testcontainers 测试无法跑。CI 环境也要装 Docker-in-Docker。
2. **第一次拉镜像耗时 30-60 秒** —— `postgres:16-alpine` 约 90MB。本地拉一次后会缓存。
3. **`database.DB` 是全局变量** —— Task 8 的集成测试通过 swap 替换；切换不是 thread-safe，因此集成测试不能与 `go test -parallel` 并发跑。建议在 CI 单独跑或加 `-p 1`。
4. **--cleanup-once 仍连真 DB** —— 如果工程师本机没起 Postgres，这条命令会 fatal log。这是有意的：CLI 模式专为线上 / staging 用。
5. **Pre-existing 0 lint warning baseline** —— Task 10 npm run lint 不增加新 error。Vendor warnings 如已存在（PrintLayout.tsx U+3000 等）保持不动。
6. **service.RollupReport 与 repository.RollupReport 结构等价但是不同类型** —— Task 9 的 makeCleanupDeps 显式做了字段对拷。后续若 repository.RollupReport 加字段，需要同步 service.RollupReport 和 makeCleanupDeps。
