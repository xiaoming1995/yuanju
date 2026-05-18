# Migration Safety Net Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 把 `backend/pkg/database/database.go` 里 35 个 `DB.Exec` 内联 DDL 全部迁出到基于 `pressly/goose` v3 的版本化 migration 文件，0001 baseline 失败 fatal、0002+ 失败 warn-only + backend 继续启动。

**Architecture:** 新建 `pkg/database/migrations.go` 作为 goose 薄封装，提供 `Migrate(mode)` / `MigrateFromDir(mode, dir)` 双入口；migration 文件嵌入二进制（`embed.FS`）；`main.go` 新增 `--migrate-dry-run` / `--migrate-apply` flag。

**Tech Stack:** Go 1.25 / `pressly/goose/v3`（新引入）/ embed.FS / testcontainers-go（已引入）/ Postgres 16

---

## Spec Reference

**Spec:** `docs/superpowers/specs/2026-05-18-migration-safety-net-design.md` (commit `9f8f220` on `main`)

**Starting branch:** `main` HEAD=`9f8f220`. Task 0 cuts a new branch `feat/migration-safety-net`.

---

## File Structure

**Create (8 files):**
- `backend/pkg/database/migrations.go` — goose 封装 + `Migrate(mode)` 主入口
- `backend/pkg/database/migrations/00001_baseline.sql` — 现有 917 行 DDL 提取
- `backend/pkg/database/migrations_test.go` — 6 个集成测试
- `backend/pkg/database/migrations_testdata/good_v2/00001_baseline.sql` — 合成简短 baseline（测试用）
- `backend/pkg/database/migrations_testdata/good_v2/00002_valid_change.sql` — 合成 valid 新 migration
- `backend/pkg/database/migrations_testdata/bad_v2/00001_baseline.sql` — 同 good_v2 的 baseline
- `backend/pkg/database/migrations_testdata/bad_v2/00002_broken.sql` — 故意写错（测试 failure 路径）
- `backend/pkg/database/migrations_testdata/empty/` — 空目录（dry-run 测试用），含 `.gitkeep`

**Modify (4 files):**
- `backend/pkg/database/database.go` — 砍掉 `Migrate()` 里 ~820 行 DDL；旧 `Migrate()` 改为薄包装调 `Migrate(ModeStartup)`
- `backend/cmd/api/main.go` — 加 2 个 flag；按 flag 分支调 `database.Migrate(mode)`
- `backend/go.mod` — 加 `github.com/pressly/goose/v3`
- `backend/go.sum` — `go mod tidy` 自动更新

---

## Task 0: Branch + Baseline sanity

**Files:** none (git only)

- [ ] **Step 1: Create feature branch from main**

```bash
git -C /Users/liujiming/web/yuanju checkout -b feat/migration-safety-net
git -C /Users/liujiming/web/yuanju status
```
Expected: `On branch feat/migration-safety-net ... working tree clean`

- [ ] **Step 2: Baseline backend build + test**

```bash
cd /Users/liujiming/web/yuanju/backend && go build ./... && go test ./...
```
Expected: All tests PASS.

- [ ] **Step 3: No commit yet** — sanity only.

---

## Task 1: Add goose dependency

**Files:**
- Modify: `backend/go.mod`, `backend/go.sum`

- [ ] **Step 1: Add dependency**

```bash
cd /Users/liujiming/web/yuanju/backend && \
  go get github.com/pressly/goose/v3 && \
  go mod tidy
```
Expected: dep added. No build failures.

- [ ] **Step 2: Verify build**

```bash
cd /Users/liujiming/web/yuanju/backend && go build ./...
```
Expected: no errors.

- [ ] **Step 3: Commit**

```bash
git -C /Users/liujiming/web/yuanju add backend/go.mod backend/go.sum
git -C /Users/liujiming/web/yuanju commit -m "deps: add pressly/goose/v3 for migration management"
```

---

## Task 2: Create test fixtures (synthetic SQL)

**Files:**
- Create: `backend/pkg/database/migrations_testdata/good_v2/00001_baseline.sql`
- Create: `backend/pkg/database/migrations_testdata/good_v2/00002_valid_change.sql`
- Create: `backend/pkg/database/migrations_testdata/bad_v2/00001_baseline.sql`
- Create: `backend/pkg/database/migrations_testdata/bad_v2/00002_broken.sql`
- Create: `backend/pkg/database/migrations_testdata/empty/.gitkeep`

These fixtures are **synthetic** — they don't use the real production baseline. The framework's correctness is verified with simple stand-ins; the real baseline's correctness is verified separately via schema parity check (T7).

- [ ] **Step 1: Create good_v2/00001_baseline.sql**

```sql
-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS test_foo (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS test_seed_table (
    key VARCHAR(50) PRIMARY KEY,
    value TEXT
);
INSERT INTO test_seed_table (key, value) VALUES ('foo', 'bar') ON CONFLICT (key) DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- baseline 不支持回滚
SELECT 1;
```

- [ ] **Step 2: Create good_v2/00002_valid_change.sql**

```sql
-- +goose Up
CREATE TABLE test_bar (
    id SERIAL PRIMARY KEY,
    foo_id INT REFERENCES test_foo(id)
);

-- +goose Down
DROP TABLE IF EXISTS test_bar;
```

- [ ] **Step 3: Create bad_v2/00001_baseline.sql**

Same content as good_v2/00001_baseline.sql — copy verbatim.

- [ ] **Step 4: Create bad_v2/00002_broken.sql**

```sql
-- +goose Up
-- syntax error: missing table name on purpose
SELECT * FROM ;

-- +goose Down
SELECT 1;
```

- [ ] **Step 5: Create empty/.gitkeep**

Just an empty file so git tracks the directory.

```bash
mkdir -p /Users/liujiming/web/yuanju/backend/pkg/database/migrations_testdata/empty
touch /Users/liujiming/web/yuanju/backend/pkg/database/migrations_testdata/empty/.gitkeep
```

- [ ] **Step 6: Commit**

```bash
git -C /Users/liujiming/web/yuanju add backend/pkg/database/migrations_testdata/
git -C /Users/liujiming/web/yuanju commit -m "test(migrations): synthetic SQL fixtures for goose integration tests"
```

---

## Task 3: RED — write 6 failing integration tests

**Files:**
- Create: `backend/pkg/database/migrations_test.go`

This task writes integration tests against `Migrate(mode)` / `MigrateFromDir(mode, dir)` which don't exist yet. **Expect compile failure.** Do NOT create the types/functions to fix the failure — Task 4 does that.

- [ ] **Step 1: Write the test file with EXACTLY this content**

Create `backend/pkg/database/migrations_test.go`:

```go
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
// 跟 cleanup_integration_test 同模式，但本测试在 database 包外仍能访问导出变量。
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

// 测试 2：同 DB 跑两次 ModeStartup → 第二次 Applied=[]、Skipped=[1]
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
	// 0002 也已 applied，应在 Skipped 中
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
// 关键断言：Migrate 返回，没 log.Fatalf 退出（如果 fatal，测试进程会死）
func TestStartup_BrokenV2_DoesNotBlock(t *testing.T) {
	db, cleanup := spinUpPG(t)
	defer cleanup()
	restore := swapGlobalDB(t, db)
	defer restore()

	rep, err := database.MigrateFromDir(database.ModeStartup, fixturePath("bad_v2"))
	// ModeStartup 的契约：返回 err=nil，错误信息在 rep.Failed 里
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

	// 断言 DB 没被改：test_foo 不存在
	var exists bool
	row := db.QueryRow(`SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name='test_foo')`)
	if err := row.Scan(&exists); err != nil {
		t.Fatalf("query: %v", err)
	}
	if exists {
		t.Errorf("dry-run should not have created test_foo")
	}
}
```

- [ ] **Step 2: Run tests — expect compile FAIL (RED)**

```bash
cd /Users/liujiming/web/yuanju/backend && go test ./pkg/database/ -run TestStartup -v 2>&1 | head -30
```

**Expected:** COMPILE ERROR mentioning undefined symbols:
- `database.MigrateFromDir`
- `database.ModeStartup`
- `database.ModeDryRun`
- `database.ModeApply`
- `MigrationReport` field access (`Applied`, `Skipped`, `Failed`, `Pending`, `Failed[0].Version`)

This is THE POINT of RED — Task 4 will resolve all these.

- [ ] **Step 3: Commit RED**

```bash
git -C /Users/liujiming/web/yuanju add backend/pkg/database/migrations_test.go
git -C /Users/liujiming/web/yuanju commit -m "test(migrations): RED — failing integration tests for Migrate(mode)"
```

---

## Task 4: GREEN — implement migrations.go

**Files:**
- Create: `backend/pkg/database/migrations.go`
- Create: `backend/pkg/database/migrations/` (empty directory)

This is the GREEN half. Create `migrations.go` so all 6 tests from Task 3 pass.

- [ ] **Step 1: Create empty migrations/ dir with .gitkeep**

```bash
mkdir -p /Users/liujiming/web/yuanju/backend/pkg/database/migrations
touch /Users/liujiming/web/yuanju/backend/pkg/database/migrations/.gitkeep
```

(Real baseline.sql lands here in Task 5; this step just ensures `//go:embed` works.)

- [ ] **Step 2: Write migrations.go**

Create `backend/pkg/database/migrations.go` with EXACTLY this content:

```go
package database

import (
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"sort"
	"time"

	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var embeddedMigrations embed.FS

// MigrationMode 控制 Migrate 的失败语义。
type MigrationMode int

const (
	ModeStartup MigrationMode = iota // 0001 fatal、0002+ warn-only
	ModeDryRun                       // 仅列 pending，不动 DB
	ModeApply                        // 全部 fatal-on-error；用于 CLI / CI
)

// FailedMigration 记录某个 version 的迁移失败信息。
type FailedMigration struct {
	Version int64
	Err     error
}

// MigrationReport 是 Migrate 一次运行的结构化结果。
type MigrationReport struct {
	Mode     MigrationMode
	Applied  []int64
	Skipped  []int64
	Failed   []FailedMigration
	Pending  []int64
	Duration time.Duration
}

const defaultEmbeddedDir = "migrations"

// Migrate 是 backend 启动时使用的入口。从 embed.FS 读取 migrations。
func Migrate(mode MigrationMode) (MigrationReport, error) {
	return runMigrations(mode, embeddedMigrations, defaultEmbeddedDir)
}

// MigrateFromDir 测试用：从指定文件系统目录读取 migrations。
func MigrateFromDir(mode MigrationMode, dir string) (MigrationReport, error) {
	return runMigrations(mode, os.DirFS(dir), ".")
}

// runMigrations 是 Migrate / MigrateFromDir 的共同实现。
func runMigrations(mode MigrationMode, baseFS fs.FS, dir string) (MigrationReport, error) {
	started := time.Now()
	rep := MigrationReport{Mode: mode}

	goose.SetBaseFS(baseFS)
	defer goose.SetBaseFS(nil)
	_ = goose.SetDialect("postgres")

	switch mode {
	case ModeDryRun:
		err := computePending(dir, &rep)
		rep.Duration = time.Since(started)
		logRun(rep)
		return rep, err

	case ModeStartup:
		// Phase 1: baseline (0001) fatal-on-error
		before := appliedVersions(DB)
		if err := goose.UpTo(DB, dir, 1); err != nil {
			rep.Duration = time.Since(started)
			logRun(rep)
			log.Fatalf("[migrate] baseline (0001) failed: %v", err)
		}
		// Phase 2: 其余 warn-only
		err := goose.Up(DB, dir)
		after := appliedVersions(DB)

		populateAppliedAndSkipped(&rep, before, after)

		if err != nil {
			// goose stop 在失败 version；拿到 last attempted version 写入 Failed
			rep.Failed = append(rep.Failed, FailedMigration{
				Version: nextPendingAfter(dir, after),
				Err:     err,
			})
			log.Printf("[migrate] non-baseline migration failed (continuing): %v", err)
		}
		rep.Duration = time.Since(started)
		logRun(rep)
		return rep, nil

	case ModeApply:
		before := appliedVersions(DB)
		err := goose.Up(DB, dir)
		after := appliedVersions(DB)
		populateAppliedAndSkipped(&rep, before, after)
		if err != nil {
			rep.Failed = append(rep.Failed, FailedMigration{
				Version: nextPendingAfter(dir, after),
				Err:     err,
			})
		}
		rep.Duration = time.Since(started)
		logRun(rep)
		return rep, err

	default:
		return rep, fmt.Errorf("unknown migration mode: %d", mode)
	}
}

// appliedVersions 读 schema_migrations 表里 applied=true 的版本号列表。
// 出错或表不存在时返回空切片。
func appliedVersions(db interface{ /* sql.DB */ }) []int64 {
	if DB == nil {
		return nil
	}
	rows, err := DB.Query(`SELECT version_id FROM goose_db_version WHERE is_applied=true ORDER BY version_id`)
	if err != nil {
		return nil
	}
	defer rows.Close()
	var out []int64
	for rows.Next() {
		var v int64
		if err := rows.Scan(&v); err == nil {
			out = append(out, v)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

// populateAppliedAndSkipped 根据 before/after 计算 Applied 和 Skipped。
func populateAppliedAndSkipped(rep *MigrationReport, before, after []int64) {
	beforeSet := map[int64]bool{}
	for _, v := range before {
		beforeSet[v] = true
	}
	for _, v := range after {
		if beforeSet[v] {
			rep.Skipped = append(rep.Skipped, v)
		} else {
			rep.Applied = append(rep.Applied, v)
		}
	}
}

// listVersionFiles 从 dir 里读取所有 NNNN_xxx.sql 文件，返回升序版本号列表。
func listVersionFiles(dir string) ([]int64, error) {
	entries, err := fs.ReadDir(goose.BaseFS(), dir)
	if err != nil {
		return nil, err
	}
	var versions []int64
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		// 文件名 00001_xxx.sql → 解析前面那段数字
		i := 0
		for i < len(name) && name[i] >= '0' && name[i] <= '9' {
			i++
		}
		if i == 0 {
			continue
		}
		var v int64
		if _, err := fmt.Sscan(name[:i], &v); err != nil {
			continue
		}
		versions = append(versions, v)
	}
	sort.Slice(versions, func(i, j int) bool { return versions[i] < versions[j] })
	return versions, nil
}

// computePending 把 dir 里所有版本中 schema_migrations 没记录的填入 rep.Pending。
func computePending(dir string, rep *MigrationReport) error {
	all, err := listVersionFiles(dir)
	if err != nil {
		return err
	}
	applied := map[int64]bool{}
	for _, v := range appliedVersions(DB) {
		applied[v] = true
	}
	for _, v := range all {
		if !applied[v] {
			rep.Pending = append(rep.Pending, v)
		}
	}
	return nil
}

// nextPendingAfter 找 after 之后下一个 pending 版本号；失败时返回 0。
func nextPendingAfter(dir string, after []int64) int64 {
	all, err := listVersionFiles(dir)
	if err != nil {
		return 0
	}
	appliedMax := int64(0)
	for _, v := range after {
		if v > appliedMax {
			appliedMax = v
		}
	}
	for _, v := range all {
		if v > appliedMax {
			return v
		}
	}
	return 0
}

// logRun 写一行结构化 JSON 日志，与 cleanup_run 风格对齐。
func logRun(rep MigrationReport) {
	type failed struct {
		Version int64  `json:"version"`
		Err     string `json:"err"`
	}
	type out struct {
		Evt     string   `json:"evt"`
		Mode    string   `json:"mode"`
		DurMs   int64    `json:"duration_ms"`
		Applied []int64  `json:"applied,omitempty"`
		Skipped []int64  `json:"skipped,omitempty"`
		Failed  []failed `json:"failed,omitempty"`
		Pending []int64  `json:"pending,omitempty"`
	}
	o := out{
		Evt:     "migrate_run",
		Mode:    modeString(rep.Mode),
		DurMs:   rep.Duration.Milliseconds(),
		Applied: rep.Applied,
		Skipped: rep.Skipped,
		Pending: rep.Pending,
	}
	for _, f := range rep.Failed {
		errStr := ""
		if f.Err != nil {
			errStr = f.Err.Error()
		}
		o.Failed = append(o.Failed, failed{Version: f.Version, Err: errStr})
	}
	b, _ := json.Marshal(o)
	log.Println(string(b))
}

func modeString(m MigrationMode) string {
	switch m {
	case ModeStartup:
		return "startup"
	case ModeDryRun:
		return "dry-run"
	case ModeApply:
		return "apply"
	}
	return "unknown"
}

// MarshalMigrationReport 为 CLI 模式打印 JSON。
func MarshalMigrationReport(rep MigrationReport) []byte {
	type failed struct {
		Version int64  `json:"version"`
		Err     string `json:"err"`
	}
	type out struct {
		Mode    string   `json:"mode"`
		DurMs   int64    `json:"duration_ms"`
		Applied []int64  `json:"applied"`
		Skipped []int64  `json:"skipped"`
		Failed  []failed `json:"failed"`
		Pending []int64  `json:"pending"`
	}
	o := out{
		Mode:    modeString(rep.Mode),
		DurMs:   rep.Duration.Milliseconds(),
		Applied: rep.Applied,
		Skipped: rep.Skipped,
		Pending: rep.Pending,
	}
	for _, f := range rep.Failed {
		errStr := ""
		if f.Err != nil {
			errStr = f.Err.Error()
		}
		o.Failed = append(o.Failed, failed{Version: f.Version, Err: errStr})
	}
	b, _ := json.MarshalIndent(o, "", "  ")
	return b
}

// ErrNoBaseline 表示 migrations 目录里找不到 0001 baseline。
var ErrNoBaseline = errors.New("0001 baseline migration not found")
```

**注意点：**
- `appliedVersions` 用 `DB` 全局变量（与 cleanup_service 模式一致）；测试通过 `swapGlobalDB` 注入
- `goose.BaseFS()` 返回当前设置的 FS；如果未设置 SetBaseFS，则返回 nil
- goose 用的表名是 `goose_db_version`（注意：不是 spec §4 写的 `schema_migrations`；spec 用的是通用术语，goose 的实际表名以代码为准）

- [ ] **Step 3: Run T3 tests — expect PASS (GREEN)**

```bash
cd /Users/liujiming/web/yuanju/backend && go test ./pkg/database/ -run TestStartup -v -count=1
cd /Users/liujiming/web/yuanju/backend && go test ./pkg/database/ -run TestApply -v -count=1
cd /Users/liujiming/web/yuanju/backend && go test ./pkg/database/ -run TestDryRun -v -count=1
```

**Expected:** 6 tests PASS.

- [ ] **Step 4: Run full backend test suite — no regressions**

```bash
cd /Users/liujiming/web/yuanju/backend && go test ./...
```
Expected: all PASS.

- [ ] **Step 5: Commit GREEN**

```bash
git -C /Users/liujiming/web/yuanju add backend/pkg/database/migrations.go backend/pkg/database/migrations/.gitkeep
git -C /Users/liujiming/web/yuanju commit -m "feat(migrations): implement Migrate(mode) + MigrateFromDir with goose"
```

---

## Task 5: Extract 00001_baseline.sql from database.go

**Files:**
- Create: `backend/pkg/database/migrations/00001_baseline.sql`

This is the **largest task in the plan**. ~900 lines of SQL extracted from `database.go::Migrate()`. **Do NOT reorder statements** — preserve source order exactly.

### Step 1: Read the existing `Migrate()` body

```bash
grep -n "^func Migrate" /Users/liujiming/web/yuanju/backend/pkg/database/database.go
# Expected line 95
```

The function body runs from line ~95 to ~916. Every `DB.Exec(...)` call is a SQL statement to extract.

### Step 2: Extraction methodology

Read `database.go` lines 95-916 sequentially. For each construct:

| Construct | What to extract |
|---|---|
| `DB.Exec(\`...inline SQL...\`)` | The backtick string content, verbatim |
| `varName := \`...sql...\`; DB.Exec(varName)` | The string assigned to varName |
| `for _, sql := range []string{a, b, c} { DB.Exec(sql) }` | Each of a, b, c as a separate statement, in slice order |
| Concatenation: `DB.Exec(\`ALTER TABLE ... ADD COLUMN \` + col.name + \` \` + col.def)` | Expand the loop's values inline. E.g., for `col := range []struct{name,def string}{{"category", "VARCHAR(20)..."}, {"short_desc", "VARCHAR(200)..."}}`, emit two literal ALTER TABLE statements with category and short_desc filled in. |
| `log.Println(...)` / `log.Printf(...)` | **Skip** — not SQL |
| `if err != nil { log.Fatalf(...) }` | **Skip** — Go wrapping |
| Plain Go variables that prepare SQL strings (e.g., `schema := \`...\``) | Yes, extract their content where used |

### Step 3: Specific known constructs (use as a reference checklist)

Concrete locations / patterns to address — verify nothing is missed:

| Line | Pattern | Notes |
|---|---|---|
| 242 | `DB.Exec(schema)` | Variable `schema` is a multi-line string. Extract its full content. |
| 265 | `DB.Exec(insertPromptSQL, defaultLiunianPrompt)` | **Parameterized query.** The `$1` placeholder must be literally replaced with the value of `defaultLiunianPrompt` quoted as a string literal. Look up the Go const value. |
| 319, 327, 377, 518, 564, 617, 630, 709, 726, 742, 778, 791, 824, 838, 873, 881, 901, 912 | `DB.Exec(varName)` | Each is a single SQL block. Extract verbatim. |
| 509 | `for _, sql := range resourceUnbindMigrations { DB.Exec(sql) }` | Read the slice literal (lines ~585-597); emit each entry as separate statement |
| 576 | `DB.Exec(alterSQL)` | Inline single ALTER |
| 586 | `for _, migSQL := range calendarMigrations { DB.Exec(migSQL) }` | Same expand pattern |
| 599 | `for _, sql := range resourceUnbindMigrations { DB.Exec(sql) }` | Same |
| 610 | `for _, migSQL := range ... { DB.Exec(migSQL) }` | Same |
| 683 | `DB.Exec(...)` inside a `for shenshaList := range ... { ... }` loop with parameterized INSERT | Expand to a series of `INSERT ... ON CONFLICT DO NOTHING` statements with the literal Go-side `shenshaList` values. See lines 660-680 for the list. |
| 697 | `DB.Exec(\`ALTER TABLE shensha_annotations ADD COLUMN IF NOT EXISTS \` + col.name + \` \` + col.def)` | Loop body emits `category VARCHAR(20)...` and `short_desc VARCHAR(200)...` — 2 ALTER statements |
| 745 | Inline ALTER | Verbatim |
| 803 | `for _, stmt := range compatibilityIndexes { DB.Exec(stmt) }` | Read slice at lines ~792-800 |
| 843, 848, 851, 856 | Inline ALTERs | Verbatim |

If anything else exists between lines 95 and 916 that calls `DB.Exec`, also extract.

### Step 4: Wrap the extracted bundle in goose markers

Create `backend/pkg/database/migrations/00001_baseline.sql`:

```sql
-- +goose Up
-- +goose StatementBegin

-- 从 backend/pkg/database/database.go::Migrate() 提取（源代码顺序保留）
-- 命名约定：每一段保留原 database.go 的中文注释以便对照

<all extracted SQL statements in source order>

-- +goose StatementEnd

-- +goose Down
-- baseline 不支持回滚 —— 如需重置 schema，请清空 DB 重新跑
SELECT 1;
```

### Step 5: Replace the existing migrations/.gitkeep

```bash
rm /Users/liujiming/web/yuanju/backend/pkg/database/migrations/.gitkeep
```

(The baseline file now lives in this directory; .gitkeep no longer needed.)

### Step 6: Verify embed picks up the new file

```bash
cd /Users/liujiming/web/yuanju/backend && go build ./pkg/database/
```
Expected: clean build. `//go:embed` directive will fail at compile time if the pattern matches no files.

### Step 7: Verify production tests still pass

```bash
cd /Users/liujiming/web/yuanju/backend && go test ./pkg/database/ -v -count=1
```

**Expected:** 6 integration tests still PASS (they use synthetic fixtures, unaffected). NO new test for baseline content — verified separately in T7.

### Step 8: Commit

```bash
git -C /Users/liujiming/web/yuanju add backend/pkg/database/migrations/
git -C /Users/liujiming/web/yuanju commit -m "feat(migrations): extract baseline SQL from database.go to 00001_baseline.sql"
```

---

## Task 6: Wire main.go + database.go shim

**Files:**
- Modify: `backend/pkg/database/database.go`
- Modify: `backend/cmd/api/main.go`

- [ ] **Step 1: Delete the entire old Migrate() function in database.go**

Open `backend/pkg/database/database.go`. The function `Migrate()` (no-arg) currently spans line ~95 to ~916 with all the inline DDL. **Delete the entire function body and signature.**

The new `Migrate(mode MigrationMode)` defined in `migrations.go` (T4) takes its place. They are in the same package, so removing the old signature avoids a name collision.

**Delete:**
- The whole `func Migrate() { ... }` block (~822 lines)
- Any imports that become unused after the deletion (likely `strings`)

**Keep:**
- Package declaration `package database`
- `var DB *sql.DB`
- `legacyKBGejv` const (it's used elsewhere)
- `func Connect()` and its imports
- All other top-level declarations

After the edit, `database.go` should be ~80-100 lines (down from 917). Run `go build` after to catch unused imports.

- [ ] **Step 2: Modify main.go — add 2 flags and dispatch**

Open `backend/cmd/api/main.go`.

**Add to imports** (`flag` and `database` already imported from T9 of previous spec):

No new imports needed. `flag`, `fmt`, `os`, `database` are already there.

**Replace the existing flag parsing block** at top of `func main()`:

```go
func main() {
	cleanupOnce := flag.Bool("cleanup-once", false, "运行一次清理任务后退出，不启动 HTTP server")
	migrateDryRun := flag.Bool("migrate-dry-run", false, "打印 pending migration 后退出，不动 DB")
	migrateApply := flag.Bool("migrate-apply", false, "强制跑一次 migration 后退出，不启动 HTTP server")
	flag.Parse()

	// 加载配置
	configs.Load()

	// 连接数据库
	database.Connect()

	// CLI 分支：迁移工具命令在 Migrate 之前处理，避免重复迁移
	if *migrateDryRun {
		rep, err := database.Migrate(database.ModeDryRun)
		if err != nil {
			log.Fatalf("dry-run 失败: %v", err)
		}
		fmt.Println(string(database.MarshalMigrationReport(rep)))
		os.Exit(0)
	}
	if *migrateApply {
		rep, err := database.Migrate(database.ModeApply)
		fmt.Println(string(database.MarshalMigrationReport(rep)))
		if err != nil {
			os.Exit(1)
		}
		os.Exit(0)
	}

	// 默认启动路径：跑 ModeStartup 迁移（0001 fatal、0002+ warn-only）
	if _, err := database.Migrate(database.ModeStartup); err != nil {
		// 这条路径理论上不会触发（ModeStartup 失败已经在 Migrate 内 fatal 了），
		// 但保留一个守卫日志以防 goose 行为变化。
		log.Printf("[migrate] startup unexpected error: %v", err)
	}

	// 确保 logo 上传目录存在（保留原逻辑）
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

	// ... 以下保留原 main.go 后续代码（cleanup 服务构造 / scheduler / gin 启动等）不动 ...
```

**Critical:** delete the old `database.Migrate()` call (which used the legacy no-arg signature) wherever it currently is. The new version takes a MigrationMode parameter.

- [ ] **Step 3: Verify build + full tests**

```bash
cd /Users/liujiming/web/yuanju/backend && go build ./... && go vet ./...
cd /Users/liujiming/web/yuanju/backend && go test ./...
```
Expected: all PASS.

- [ ] **Step 4: Verify --migrate-dry-run runs against current dev DB**

```bash
docker exec yuanju_backend /root/server --migrate-dry-run 2>&1 | tail -20
```

**Expected output (dev DB already at baseline equivalent):**
```json
{
  "mode": "dry-run",
  "duration_ms": <small>,
  "applied": [],
  "skipped": [],
  "failed": [],
  "pending": [1]
}
```

Note: `pending` will be `[1]` because the dev DB doesn't yet have a `goose_db_version` table — to goose, 0001 baseline is unapplied. The actual schema is already there (via legacy Migrate), so when we run `--migrate-apply` later (T7), 0001 will run as no-op via `IF NOT EXISTS` and write into `goose_db_version`.

(If the dev container is not running the new binary yet, this step is informational — confirm visually it produces JSON output.)

- [ ] **Step 5: Commit**

```bash
git -C /Users/liujiming/web/yuanju add backend/cmd/api/main.go backend/pkg/database/database.go
git -C /Users/liujiming/web/yuanju commit -m "feat(migrations): wire main.go --migrate-dry-run/apply + delete legacy Migrate body"
```

---

## Task 7: Schema parity check (manual verification)

**Files:** none (verification only)

This proves that the new goose-driven path produces a schema equivalent to the legacy `database.go::Migrate()`.

### Step 1: Take a snapshot of the current production-equivalent schema

The running `yuanju_postgres` container has the schema produced by the legacy Migrate. Dump it:

```bash
docker exec yuanju_postgres pg_dump --schema-only --no-owner --no-privileges -U yuanju -d yuanju \
  > /tmp/before_legacy_schema.sql
wc -l /tmp/before_legacy_schema.sql
```

Expected: 1000-2000 lines of SQL.

### Step 2: Spin up a fresh container and apply the new baseline

```bash
docker run --rm -d --name pg_parity_check \
  -e POSTGRES_USER=yuanju -e POSTGRES_PASSWORD=test -e POSTGRES_DB=yuanju \
  -p 15432:5432 postgres:16-alpine

# wait for ready
until docker exec pg_parity_check pg_isready -U yuanju -d yuanju; do sleep 1; done

# build feat branch backend
cd /Users/liujiming/web/yuanju/backend && go build -o /tmp/api-feat ./cmd/api

# run apply mode against the parity-check container.
# configs/config.go reads a single DATABASE_URL env var; override .env loading
# by ensuring no .env file is in cwd or use a temp dir.
cd /tmp && DATABASE_URL="postgres://yuanju:test@localhost:15432/yuanju?sslmode=disable" \
  /tmp/api-feat --migrate-apply 2>&1 | tail -15
```

Expected JSON output: `applied: [1]`, `failed: []`. If `failed` is non-empty, the baseline has a bug — go fix `00001_baseline.sql` before continuing.

### Step 3: Dump the new schema

```bash
docker exec pg_parity_check pg_dump --schema-only --no-owner --no-privileges -U yuanju -d yuanju \
  > /tmp/after_goose_schema.sql
```

### Step 4: Diff

```bash
diff -u /tmp/before_legacy_schema.sql /tmp/after_goose_schema.sql | head -80
```

**Expected differences:**
- New table `goose_db_version` (only in /tmp/after) — that's goose's internal tracking
- Possibly minor formatting / whitespace differences in pg_dump output (irrelevant)

**Unexpected differences (would be a real bug):**
- A table missing in `after.sql` → baseline extraction skipped it
- A column missing → baseline missed an ALTER
- A different column type → baseline got the type wrong

### Step 5: Cleanup

```bash
docker stop pg_parity_check
rm /tmp/api-feat /tmp/before_legacy_schema.sql /tmp/after_goose_schema.sql
```

### Step 6: Commit (only if there were fixes)

If the diff revealed missing statements, the implementer must go back to `00001_baseline.sql`, add the missing SQL, and re-run the parity check. Commit:

```bash
git -C /Users/liujiming/web/yuanju add backend/pkg/database/migrations/00001_baseline.sql
git -C /Users/liujiming/web/yuanju commit -m "fix(migrations): restore <description> missed in baseline extraction"
```

If the diff was clean, no commit needed.

---

## Task 8: Full verification + final review

**Files:** none (verification only)

- [ ] **Step 1: Run full backend test suite**

```bash
cd /Users/liujiming/web/yuanju/backend && go test ./... -count=1
```
Expected: all PASS.

- [ ] **Step 2: Run frontend tests + build (no FE changes expected)**

```bash
cd /Users/liujiming/web/yuanju/frontend && \
  node --test --experimental-strip-types tests/*.test.mjs && \
  npm run build
```
Expected: still PASS (no FE diff in this branch).

- [ ] **Step 3: Verify database.go is now small + grep for leftover DDL**

```bash
wc -l /Users/liujiming/web/yuanju/backend/pkg/database/database.go
grep -c "DB.Exec" /Users/liujiming/web/yuanju/backend/pkg/database/database.go
```
Expected: `wc -l` ≤ 300; `grep -c "DB.Exec"` = 0 (all DDL Execs gone).

- [ ] **Step 4: Inspect commits**

```bash
git -C /Users/liujiming/web/yuanju log --oneline main..HEAD
```
Expected: ~7-8 commits matching the task list.

- [ ] **Step 5: No commit** — this task is purely verification.

---

## Spec Coverage Cross-Check

| Spec section | Implementing task |
|---|---|
| §2 baseline 一刀切 | T5 (extract) |
| §2 goose 引入 | T1 (deps) + T4 (migrations.go) |
| §2 0001 fatal / 0002+ warn-only | T4 (Migrate logic) verified by T3 tests #4 #5 |
| §2 --migrate-dry-run / --migrate-apply | T6 (main.go) |
| §2 evt=migrate_run 结构化日志 | T4 logRun |
| §2 集成测试 6 项 | T3 (RED) + T4 (GREEN) |
| §5.3 类型 / 函数签名 | T4 |
| §5.4 baseline 格式 | T5 |
| §5.5 后续 migration 模板 | Documented; no code in this branch |
| §5.6 边界规则 | T4 (signatures), T6 (database.go shim) |
| §6.1 启动流程 | T4 ModeStartup + T6 main.go wiring |
| §6.2 dry-run | T4 ModeDryRun |
| §6.3 apply | T4 ModeApply + T6 main.go |
| §6.4 结构化日志 | T4 logRun |
| §7.1 故障模式 | T4 + tested by T3 #4 #5 |
| §8.1 6 集成测试 | T3 + T4 |
| §8.2 schema parity check | T7 |

**No gaps.**

---

## TDD 节奏总结

| Task | 类型 | Commits |
|---|---|---|
| 0 | branch + sanity | 0 |
| 1 | deps | 1 |
| 2 | test fixtures | 1 |
| 3 | RED tests | 1 |
| 4 | GREEN migrations.go | 1 |
| 5 | baseline extraction | 1（可能 +1 修复） |
| 6 | main.go + database.go wiring | 1 |
| 7 | schema parity check | 0（无 fix 时）或 1 |
| 8 | verification | 0 |

**总计 6-8 个 commit**，T5 是最大单 commit（~900 LOC SQL），其他都在 200 LOC 内。

---

## 已知风险与注意点

1. **baseline 提取的人工失误**：T5 是 mechanical 但易遗漏（35 个 Exec 散布在 820 行里，含动态拼接 / 循环展开）。T7 的 schema parity check 是兜底防线；强烈建议合并前必跑。
2. **embed.FS 路径**：`//go:embed migrations/*.sql` 的相对路径是 `migrations.go` 所在的 `pkg/database/` 目录下的 `migrations/`。Task 4 已经准备空目录确保 embed 不报错；Task 5 才填充内容。
3. **goose 的 `goose_db_version` 表名**：与 spec §4 写的 `schema_migrations` 名字不同（spec 用的是通用术语）。实际代码读 `goose_db_version`。Plan 已修正。
4. **Docker daemon 必须运行** —— Task 3 / 4 / 7 都需要。
5. **第一次 prod 部署的迁移行为**：dev / prod 数据库当前没有 `goose_db_version` 表。新版本 backend 启动时，goose 会把 0001 baseline 视为 pending → 跑一次，全部 IF NOT EXISTS / ON CONFLICT 守卫起效，等价 no-op，结尾在 `goose_db_version` 写一行 `(1, true, NOW())`。**该启动路径已被 T3 测试 1 + T7 parity check 双重覆盖**。
6. **`appliedVersions` 在 `goose_db_version` 不存在时返回 nil**：第一次启动时这是正常的；populateAppliedAndSkipped 正确处理空 before。
7. **测试不能 `t.Parallel()`**：和 cleanup_integration_test 一样，因为都 swap 全局 `database.DB`。
