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
//
// 注意：本函数使用 goose 的包级全局 BaseFS / Dialect 状态，**非 goroutine-safe**。
// 实际只在 main() 启动路径或 CLI flag 处理时单线程调用，故不冲突；
// 若未来需要并发调用（例如 admin API 触发），需引入互斥锁或改用 goose Provider API。
func runMigrations(mode MigrationMode, baseFS fs.FS, dir string) (MigrationReport, error) {
	started := time.Now()
	rep := MigrationReport{Mode: mode}

	goose.SetBaseFS(baseFS)
	defer goose.SetBaseFS(nil)
	_ = goose.SetDialect("postgres")

	switch mode {
	case ModeDryRun:
		err := computePending(baseFS, dir, &rep)
		rep.Duration = time.Since(started)
		logRun(rep)
		return rep, err

	case ModeStartup:
		before := appliedVersions()
		// Phase 1: baseline (0001) fatal-on-error
		if err := goose.UpTo(DB, dir, 1); err != nil {
			rep.Duration = time.Since(started)
			logRun(rep)
			log.Fatalf("[migrate] baseline (0001) failed: %v", err)
		}
		// Phase 2: 其余 warn-only
		upErr := goose.Up(DB, dir)
		after := appliedVersions()

		populateAppliedAndSkipped(&rep, before, after)

		if upErr != nil {
			rep.Failed = append(rep.Failed, FailedMigration{
				Version: nextPendingAfter(baseFS, dir, after),
				Err:     upErr,
			})
			log.Printf("[migrate] non-baseline migration failed (continuing): %v", upErr)
		}
		rep.Duration = time.Since(started)
		logRun(rep)
		return rep, nil

	case ModeApply:
		before := appliedVersions()
		upErr := goose.Up(DB, dir)
		after := appliedVersions()
		populateAppliedAndSkipped(&rep, before, after)
		if upErr != nil {
			rep.Failed = append(rep.Failed, FailedMigration{
				Version: nextPendingAfter(baseFS, dir, after),
				Err:     upErr,
			})
		}
		rep.Duration = time.Since(started)
		logRun(rep)
		return rep, upErr

	default:
		return rep, fmt.Errorf("unknown migration mode: %d", mode)
	}
}

// appliedVersions 读 goose_db_version 表里 is_applied=true 的版本号列表（已排除 goose 自己的 version 0）。
// 出错或表不存在时返回空切片。
func appliedVersions() []int64 {
	if DB == nil {
		return nil
	}
	rows, err := DB.Query(`SELECT version_id FROM goose_db_version WHERE is_applied=true AND version_id > 0 ORDER BY version_id`)
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

// listVersionFiles 从 baseFS:dir 里读取所有 NNNN_xxx.sql 文件，返回升序版本号列表。
func listVersionFiles(baseFS fs.FS, dir string) ([]int64, error) {
	entries, err := fs.ReadDir(baseFS, dir)
	if err != nil {
		return nil, err
	}
	var versions []int64
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
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

// computePending 把 baseFS:dir 里所有版本中 goose_db_version 没记录的填入 rep.Pending。
func computePending(baseFS fs.FS, dir string, rep *MigrationReport) error {
	all, err := listVersionFiles(baseFS, dir)
	if err != nil {
		return err
	}
	applied := map[int64]bool{}
	for _, v := range appliedVersions() {
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
func nextPendingAfter(baseFS fs.FS, dir string, after []int64) int64 {
	all, err := listVersionFiles(baseFS, dir)
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
