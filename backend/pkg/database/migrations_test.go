package database_test

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/testcontainers/testcontainers-go"
	tcpg "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"yuanju/pkg/database"
)

// spinUpPG 启一个临时 Postgres 容器。与 repository 包同名 helper 重复定义
// 是有意为之 —— 跨包测试 helper 共享需要 testsupport 包，目前 YAGNI。
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

// swapGlobalDB 临时替换 database.DB，返回还原 fn。
func swapGlobalDB(t *testing.T, db *sql.DB) func() {
	t.Helper()
	orig := database.DB
	database.DB = db
	return func() { database.DB = orig }
}

func containsInt64(s []int64, v int64) bool {
	for _, x := range s {
		if x == v {
			return true
		}
	}
	return false
}

func fixturePath(rel string) string {
	wd, _ := os.Getwd() // pkg/database
	return wd + "/migrations_testdata/" + rel
}

// 测试 1：fresh DB → ModeStartup → 0001 applied，schema 出现 test_foo
func TestStartup_FreshDB_AppliesBaseline(t *testing.T) {
	db, cleanup := spinUpPG(t)
	defer cleanup()
	restore := swapGlobalDB(t, db)
	defer restore()

	rep, err := database.MigrateFromDir(database.ModeStartup, fixturePath("good_v2"))
	if err != nil {
		t.Fatalf("migrate: %v", err)
	}
	if !containsInt64(rep.Applied, 1) {
		t.Errorf("expected version 1 in Applied, got %v", rep.Applied)
	}

	var n int
	if err := db.QueryRow(`SELECT COUNT(*) FROM test_foo`).Scan(&n); err != nil {
		t.Errorf("test_foo not created: %v", err)
	}
	var seed string
	if err := db.QueryRow(`SELECT value FROM test_seed_table WHERE key='foo'`).Scan(&seed); err != nil {
		t.Errorf("test_seed_table seed missing: %v", err)
	}
	if seed != "bar" {
		t.Errorf("seed value=%q, want \"bar\"", seed)
	}
}

// 测试 2：同 DB 跑两次 ModeStartup → 第二次 Applied=[]、Skipped=[1,2]
func TestStartup_Idempotent(t *testing.T) {
	db, cleanup := spinUpPG(t)
	defer cleanup()
	restore := swapGlobalDB(t, db)
	defer restore()

	if _, err := database.MigrateFromDir(database.ModeStartup, fixturePath("good_v2")); err != nil {
		t.Fatalf("first migrate: %v", err)
	}
	rep, err := database.MigrateFromDir(database.ModeStartup, fixturePath("good_v2"))
	if err != nil {
		t.Fatalf("second migrate: %v", err)
	}
	if len(rep.Applied) != 0 {
		t.Errorf("second run Applied should be empty, got %v", rep.Applied)
	}
	if !containsInt64(rep.Skipped, 1) {
		t.Errorf("second run should skip 1, Skipped=%v", rep.Skipped)
	}
	if !containsInt64(rep.Skipped, 2) {
		t.Errorf("second run should skip 2, Skipped=%v", rep.Skipped)
	}
}

// 测试 3：fresh DB → ModeStartup with good_v2 (含 0001+0002) → 两个都被 Applied
func TestStartup_AppliesNewMigration(t *testing.T) {
	db, cleanup := spinUpPG(t)
	defer cleanup()
	restore := swapGlobalDB(t, db)
	defer restore()

	rep, err := database.MigrateFromDir(database.ModeStartup, fixturePath("good_v2"))
	if err != nil {
		t.Fatalf("migrate: %v", err)
	}
	if !containsInt64(rep.Applied, 1) || !containsInt64(rep.Applied, 2) {
		t.Errorf("expected 1 and 2 applied, got %v", rep.Applied)
	}

	var n int
	if err := db.QueryRow(`SELECT COUNT(*) FROM test_bar`).Scan(&n); err != nil {
		t.Errorf("test_bar not created: %v", err)
	}
}

// 测试 4：fresh DB → ModeStartup with bad_v2 → 0001 ok、0002 fail
// 关键断言：Migrate 返回 err=nil，错误在 rep.Failed 里；没有 log.Fatalf
func TestStartup_BrokenV2_DoesNotBlock(t *testing.T) {
	db, cleanup := spinUpPG(t)
	defer cleanup()
	restore := swapGlobalDB(t, db)
	defer restore()

	rep, err := database.MigrateFromDir(database.ModeStartup, fixturePath("bad_v2"))
	if err != nil {
		t.Fatalf("ModeStartup should swallow 0002+ errors, got err=%v", err)
	}
	if !containsInt64(rep.Applied, 1) {
		t.Errorf("expected 1 in Applied, got %v", rep.Applied)
	}
	if len(rep.Failed) != 1 || rep.Failed[0].Version != 2 {
		t.Errorf("expected Failed=[{Version:2,...}], got %v", rep.Failed)
	}
}

// 测试 5：fresh DB → ModeApply with bad_v2 → 0001 ok、0002 fail → err 非 nil
func TestApply_BrokenV2_ReturnsError(t *testing.T) {
	db, cleanup := spinUpPG(t)
	defer cleanup()
	restore := swapGlobalDB(t, db)
	defer restore()

	rep, err := database.MigrateFromDir(database.ModeApply, fixturePath("bad_v2"))
	if err == nil {
		t.Fatalf("ModeApply should return error on broken migration")
	}
	if !containsInt64(rep.Applied, 1) {
		t.Errorf("0001 should still have applied, Applied=%v", rep.Applied)
	}
	if len(rep.Failed) != 1 || rep.Failed[0].Version != 2 {
		t.Errorf("expected Failed=[{Version:2,...}], got %v", rep.Failed)
	}
}

// 测试 6：fresh DB → ModeDryRun with good_v2 → Pending=[1,2]，DB 未变
func TestDryRun_ReturnsPendingNoChanges(t *testing.T) {
	db, cleanup := spinUpPG(t)
	defer cleanup()
	restore := swapGlobalDB(t, db)
	defer restore()

	rep, err := database.MigrateFromDir(database.ModeDryRun, fixturePath("good_v2"))
	if err != nil {
		t.Fatalf("dry-run: %v", err)
	}
	if !containsInt64(rep.Pending, 1) || !containsInt64(rep.Pending, 2) {
		t.Errorf("expected Pending to contain 1 and 2, got %v", rep.Pending)
	}

	var exists bool
	row := db.QueryRow(`SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name='test_foo')`)
	if err := row.Scan(&exists); err != nil {
		t.Fatalf("query: %v", err)
	}
	if exists {
		t.Errorf("dry-run should not have created test_foo")
	}
}
