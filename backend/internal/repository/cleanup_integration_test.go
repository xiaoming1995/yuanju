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
	"yuanju/internal/repository"
	"yuanju/pkg/database"
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

// swapGlobalDB 把 database.DB 临时替换为容器 db，返回还原 fn。
// 跨包修改全局变量需要 database 包导出 DB（已是 var DB *sql.DB）。
func swapGlobalDB(t *testing.T, db *sql.DB) func() {
	t.Helper()
	original := database.DB
	database.DB = db
	return func() { database.DB = original }
}

// monthAgo 返回 now 当前时间的"几个月前"那一刻（保留 day/hour）。
// 用于生成跨月份的测试数据。
func monthAgo(now time.Time, n int) time.Time {
	return time.Date(now.Year(), now.Month()-time.Month(n), 15, 12, 0, 0, 0, time.UTC)
}

func TestRollupClosedMonthsAndDelete(t *testing.T) {
	db, cleanup := spinUpPG(t)
	defer cleanup()
	runMigrationsInContainer(t, db)

	// 把全局 database.DB 暂时指向容器内的 db
	// 因为 RollupClosedMonthsAndDelete 内部用的就是 database.DB
	restoreDB := swapGlobalDB(t, db)
	defer restoreDB()

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
	for i := 0; i < 5; i++ {
		if _, err := db.Exec(`
INSERT INTO token_usage_logs (user_id, call_type, model, prompt_tokens, completion_tokens, total_tokens, created_at)
VALUES ($1, 'report', 'gpt-4o', 10, 20, 30, NOW())`, userA); err != nil {
			t.Fatalf("insert current: %v", err)
		}
	}

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
	// monthly 行数：5 月 * (2 用户 * 2 模型) + 1 孤儿在 first month = 20 + 1 = 21
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
