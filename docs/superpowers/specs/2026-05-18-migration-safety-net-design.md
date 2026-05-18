# 迁移安全网 设计文档

**日期**：2026-05-18
**作者**：刘明（brainstorming 产物）
**范围**：把 `backend/pkg/database/database.go` 886 行内联 DDL 替换为基于 `pressly/goose` 的版本化 migration 文件；单条 migration 失败不再阻塞 backend 启动

---

## 1. 背景

2026-05-18 线上 backend 启动失败：

```
2026/05/18 10:16:06 ✅ 数据库连接成功
2026/05/18 10:16:06 数据库迁移失败: pq: could not extend file ...
```

`database.go` 当前结构（节选）：

```go
func Migrate() {
    if _, err := DB.Exec(`CREATE TABLE IF NOT EXISTS ai_reports (...)`); err != nil {
        log.Fatalf("迁移失败: %v", err)
    }
    // ... 重复 50+ 次 ...
}
```

总共 **886 行**，**50+ 段 `if-err-fatal`**。任意一段失败 → 整个 backend `log.Fatalf` 退出 → docker 自动重启 → 同样的错 → 死循环。

该事故的磁盘层面问题已由「数据膨胀防御」spec 解决，但**迁移失败阻塞启动**这条故障路径仍存在。本 spec 专门解决这条。

---

## 2. 目标与非目标

**目标：**
- 把现有 886 行 DDL 提取为 `backend/pkg/database/migrations/00001_baseline.sql`（goose 格式）
- 引入 `pressly/goose` 管理后续 migration（一文件一版本）
- 启动时：0001 baseline 失败 fatal；0002+ 失败 warn-only + backend 继续启动
- 新增 `--migrate-dry-run` / `--migrate-apply` CLI flag
- 结构化日志（`evt=migrate_run`），与 `cleanup_run` 对齐
- 集成测试基于已有 testcontainers-go 基建

**非目标（明确不做）：**
- ❌ 多 DB 支持（只用 Postgres）
- ❌ ORM 化（保持 raw SQL via lib/pq）
- ❌ 自动 down / rollback CLI（人工跑 `goose down` 或部署新版本）
- ❌ `/health` 暴露 schema 版本（noise；以后可加）
- ❌ 试图回溯 git history 拆出多个 baseline 文件（投入产出比差）
- ❌ 处理 ENV-derived 种子数据（保留现有 `pkg/seed/` 不动）

---

## 3. 决策摘要

| 维度 | 决定 | 备选与放弃理由 |
|---|---|---|
| Baseline 策略 | 现有 886 行一刀切到 `00001_baseline.sql` | 按 git log 逆推太脆弱；混合 + sanity check 维护成本高 |
| 库选型 | `pressly/goose` v3 | golang-migrate 多 DB 支持是 overkill；自己撸缺锁/dirty 处理，不一定比 goose 快 |
| 失败处理 | 0001 fatal，0002+ warn-only + 仍启动 HTTP | 全 warn-only 隐藏环境问题；全 fatal 等于现状无改善 |
| CLI 暴露 | `--migrate-dry-run` + `--migrate-apply` | emergency 指令（down/force-version）现加不晚；零手动控制太黑盒 |

---

## 4. 架构

```
backend/
├── cmd/api/main.go
│   └── 加 2 个 flag: --migrate-dry-run / --migrate-apply
│
├── pkg/database/
│   ├── database.go
│   │   ├── Connect()         # 不变
│   │   └── Migrate(mode)     # 重写：从 886 行 → 调 goose 的薄封装
│   │
│   ├── migrations.go         # 新建：goose 封装（dry-run / apply / startup）
│   │
│   ├── migrations/           # 新目录，goose 默认读这里
│   │   ├── 00001_baseline.sql              # 现有 886 行 DDL
│   │   └── (空，后续 0002+ 在这里加)
│   │
│   ├── migrations_test.go    # 6 个集成测试（testcontainers）
│   │
│   └── migrations_testdata/  # 测试 fixtures
│       ├── good_v2/
│       │   ├── 00001_baseline.sql
│       │   └── 00002_valid_change.sql
│       └── bad_v2/
│           ├── 00001_baseline.sql
│           └── 00002_broken.sql
│
└── go.mod                    # + github.com/pressly/goose/v3
```

**启动时序：**

```
main.go
  ├─ flag.Parse
  ├─ if --migrate-dry-run → Connect → Migrate(ModeDryRun) → 打印 pending → exit 0
  ├─ if --migrate-apply   → Connect → Migrate(ModeApply) → 打印 RunReport →
  │                           成功 exit 0；任意失败 exit 1
  └─ default              → Connect → Migrate(ModeStartup) →
                              ├─ Phase 1: goose.UpTo(db, dir, 1)
                              │           失败 → log.Fatalf（baseline 是 schema 断言）
                              └─ Phase 2: goose.Up(db, dir)
                                          失败 → warn log + Failed 记录，继续启动 HTTP
```

`schema_migrations` 表由 goose 自动管理（首次启动自动 CREATE TABLE），结构：

```sql
CREATE TABLE schema_migrations (
    version_id BIGINT NOT NULL PRIMARY KEY,
    is_applied BOOLEAN NOT NULL,
    tstamp     TIMESTAMP DEFAULT NOW()
);
```

---

## 5. Components & boundaries

### 5.1 新建文件

| 文件 | 职责 |
|---|---|
| `backend/pkg/database/migrations.go` | `Migrate(mode)` 实现 + `MigrationMode` / `MigrationReport` / `FailedMigration` 类型 |
| `backend/pkg/database/migrations/00001_baseline.sql` | goose 格式：Up=现有 886 行 DDL，Down=`SELECT 1;` + 注释 |
| `backend/pkg/database/migrations_test.go` | 6 个集成测试（empty + idempotent + valid v2 + broken v2 in startup + broken v2 in apply + dry-run） |
| `backend/pkg/database/migrations_testdata/good_v2/00001_baseline.sql` | 拷贝自生产 migrations |
| `backend/pkg/database/migrations_testdata/good_v2/00002_valid_change.sql` | 测试用的 valid 新 migration |
| `backend/pkg/database/migrations_testdata/bad_v2/00001_baseline.sql` | 拷贝自生产 migrations |
| `backend/pkg/database/migrations_testdata/bad_v2/00002_broken.sql` | 故意写错的 fixture |

### 5.2 修改文件

| 文件 | 改什么 |
|---|---|
| `backend/cmd/api/main.go` | + 2 个 flag；按 flag 分支调 `Migrate(mode)`；默认 ModeStartup |
| `backend/pkg/database/database.go` | 砍掉旧 `Migrate()` 中 886 行 DDL；保留 `Connect()`；旧 `Migrate()` 改为薄包装调 `Migrate(ModeStartup)` |
| `backend/go.mod` | + `github.com/pressly/goose/v3` |
| `backend/go.sum` | 由 `go mod tidy` 自动更新 |

### 5.3 关键类型

```go
package database

type MigrationMode int

const (
    ModeStartup  MigrationMode = iota
    ModeDryRun
    ModeApply
)

type FailedMigration struct {
    Version int64
    Err     error
}

type MigrationReport struct {
    Mode     MigrationMode
    Applied  []int64           // 本次跑成功的版本号
    Skipped  []int64           // 已 applied 的，跳过
    Failed   []FailedMigration // ModeStartup 下 0002+ 失败；ModeApply 下任意失败
    Pending  []int64           // 仅 ModeDryRun 填充
    Duration time.Duration
}

// Migrate 是 backend 唯一对外的迁移入口。dir 默认 "migrations"，测试可注入。
func Migrate(mode MigrationMode) (MigrationReport, error)

// MigrateFromDir 测试专用：注入 migration 目录路径
func MigrateFromDir(mode MigrationMode, dir string) (MigrationReport, error)
```

### 5.4 goose 文件格式（00001_baseline.sql）

```sql
-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS users (...);
CREATE TABLE IF NOT EXISTS bazi_charts (...);
-- ... 现有 database.go 里的全部 50+ 段 DDL，按当前顺序粘贴 ...
INSERT INTO algo_config (key, value, description) VALUES (...) ON CONFLICT (key) DO NOTHING;

-- +goose StatementEnd

-- +goose Down
-- baseline 不支持回滚；如需重置 schema，请清空 DB 重新跑
SELECT 1;
```

**为什么把整段包在 StatementBegin / StatementEnd 里**：

baseline 是「现有 schema 的一次性快照」，语义上当作一个原子 SQL blob 最简单。
goose 默认按 `;` 切分，会把单个 baseline 拆成几十条独立语句逐条 Exec ——
功能上等价（每条都有 IF NOT EXISTS 守卫），但失去了
当前 `database.go` 用大字符串一次性 Exec 的 behavior。统一交给
PG 自己一次性处理，错误信息更接近现状，也能容忍未来 seed 数据
出现 `DO $$ ... $$;` 块。

新 migration（0002+）每个文件只有一两条语句，不再需要 StatementBegin/End。

### 5.5 后续 migration 模板（0002+）

```sql
-- backend/pkg/database/migrations/00002_add_user_active_at.sql

-- +goose Up
ALTER TABLE users ADD COLUMN active_at TIMESTAMPTZ;
CREATE INDEX idx_users_active_at ON users(active_at);

-- +goose Down
DROP INDEX IF EXISTS idx_users_active_at;
ALTER TABLE users DROP COLUMN IF EXISTS active_at;
```

新文件**不再用 IF NOT EXISTS**（版本号唯一性已防重）；rollback 写真实 inverse。

### 5.6 边界规则

- `migrations.go` **不接 SQL 字符串**，只是 goose 的薄包装
- baseline 之外的 migration **每个一个独立文件**，禁止在已 applied 的版本里改 SQL（goose 不会重跑）
- `pkg/seed/` 不动（ENV-derived 数据保持现状）
- `database.DB` 仍是 `*sql.DB` 全局变量（兼容现有 repository 层）
- 失败的 0002+ migration 不写 `schema_migrations` 记录 → 下次启动会自动重试

---

## 6. Data flow

### 6.1 启动流程（ModeStartup）

```
backend startup
       │
       ▼
database.Connect()
       │
       ▼
database.Migrate(ModeStartup)
       │
       ├─ Phase 1：goose.UpTo(db, dir, 1)
       │   ├─ 成功 → 继续
       │   └─ 失败 → log.Fatalf（baseline 是不可降级的环境断言）
       │
       ├─ Phase 2：goose.Up(db, dir)
       │   ├─ 全部成功 → Applied 累加每个版本号
       │   └─ 任一失败 → 停在该 version；Failed 写入 (version, err)
       │                  不中断 caller，返回 MigrationReport
       │
       └─ 写结构化日志：evt=migrate_run
       
       ▼
（即使有 Failed） continue → HTTP server 起来 → 服务用户
```

### 6.2 dry-run 流程（ModeDryRun）

```
flag.Parse → --migrate-dry-run = true
       │
       ▼
Connect → Migrate(ModeDryRun)
       │
       ├─ goose.Status(db, dir) 列出所有 (version, applied?)
       └─ Pending = 未 applied 的版本号列表
       │
       ▼
打印（stdout）：
  Pending migrations:
    00002_add_user_active_at.sql
    00003_split_charts.sql
  Will execute 2 SQL files.
  Use --migrate-apply to apply.
       │
       ▼
os.Exit(0)
```

### 6.3 apply 流程（ModeApply）

```
flag.Parse → --migrate-apply = true
       │
       ▼
Connect → Migrate(ModeApply)
       │
       ├─ goose.Up(db, dir)  # 包含 baseline
       │   ├─ 成功 → Applied 累加
       │   └─ 失败 → 停在该 version；Failed 写入；返回 err
       │
       ▼
打印 RunReport JSON
       │
       ▼
有 err → os.Exit(1)；无 err → os.Exit(0)
```

### 6.4 结构化日志格式

```json
{
  "evt": "migrate_run",
  "mode": "startup",
  "duration_ms": 48,
  "applied": [1],
  "skipped": [],
  "failed": [
    {"version": 4, "err": "pq: column \"foo\" of relation \"users\" already exists"}
  ]
}
```

`grep evt=migrate_run` 拉历史；与 `cleanup_run` 风格对齐。

---

## 7. Error handling & 防御性设计

### 7.1 故障模式 → 应对

| 场景 | 行为 |
|---|---|
| DB 连接失败 | `database.Connect()` 已 fatal；本 spec 不动 |
| baseline (0001) 跑失败 | `log.Fatalf("baseline 失败: ...")` —— 这是 schema 断言级别错误 |
| 0002+ 某条 SQL 语法错 | warn log + Failed 记录 + 继续启动；下次启动自动重试该版本 |
| `schema_migrations` 表被人手 DROP | goose 启动时自动重建；0001 baseline 视为未 applied，会再跑；现有 `IF NOT EXISTS` 守卫确保安全 no-op |
| 同一版本号文件名冲突 | goose 会报错（不允许两个文件同 version_id）；编译期靠 review 拦 |
| `--migrate-apply` 中途失败 | 进程 exit 1，已 applied 的不回退；ops 看 RunReport 决定下一步 |
| 用户跑 `--migrate-dry-run` 但 DB 没起 | `Connect()` fatal —— 跟其他模式一致 |

### 7.2 并发 / 幂等

- 单实例 docker-compose 部署 → 单进程跑 migrate → 无并发
- 即使两个进程同时启动：`schema_migrations` 表的 PK 防止双重 applied
- baseline 全部用 `IF NOT EXISTS` / `ON CONFLICT DO NOTHING` → 重跑 0001 = no-op

### 7.3 明确不做的事（YAGNI）

- ❌ goose 的 advisory lock（无并发场景）
- ❌ schema 版本暴露 `/health`
- ❌ 自动 down / 部分 rollback
- ❌ migration 进度告警（grep 日志够用，可观测性 spec 再加）

---

## 8. Testing strategy

### 8.1 集成测试（`migrations_test.go`）

全部用 testcontainers-go（已在 `feat/data-bloat-defense` 引入）。

| # | 名称 | 验证 |
|---|---|---|
| 1 | `TestStartup_FreshDB_AppliesBaseline` | empty DB → `Migrate(ModeStartup)` → Applied=[1]；断言 schema 含 `users` / `bazi_charts` / `algo_config` 行 |
| 2 | `TestStartup_Idempotent` | 同一 DB 跑两次 ModeStartup → 第二次 Applied=[]、Skipped=[1] |
| 3 | `TestStartup_AppliesNewMigration` | 跑 0001 后注入 `good_v2/00002_valid_change.sql` → 再跑 ModeStartup → Applied=[2] |
| 4 | `TestStartup_BrokenV2_DoesNotBlock` | 0001 done，加 `bad_v2/00002_broken.sql` → ModeStartup 返回 Report.Failed=[{Version:2, Err:...}] 且 **无 panic/fatal**；HTTP 应能起来（实际测试只验证 Migrate 返回值，HTTP 单测独立） |
| 5 | `TestApply_BrokenV2_ReturnsError` | 同 fixture，但用 ModeApply → 返回 error；MigrationReport.Failed 仍准确 |
| 6 | `TestDryRun_ReturnsPendingNoChanges` | 0001 done，0002 pending → ModeDryRun → Pending=[2]；DB 没有任何写入（断言 `schema_migrations` 还是只有 1 条） |

### 8.2 手动验证（合并前必跑）

**Schema parity check**（关键，证明 baseline 等价于现状）：

```bash
# 1. 起容器 A，跑 main 分支当前的 database.Migrate()
docker run -d --name pg_a -e POSTGRES_PASSWORD=test postgres:16-alpine
sleep 5
# （从 main 分支编译 backend 跑一次）
docker exec pg_a pg_dump --schema-only -U postgres test_db > /tmp/before.sql

# 2. 起容器 B，跑 feat 分支的 goose Migrate(ModeStartup)
docker run -d --name pg_b -e POSTGRES_PASSWORD=test postgres:16-alpine
# （从 feat 分支编译 backend 跑一次）
docker exec pg_b pg_dump --schema-only -U postgres test_db > /tmp/after.sql

# 3. diff
diff /tmp/before.sql /tmp/after.sql
```

**预期差异：**
- `schema_migrations` 表（仅在 after.sql 出现）
- `tstamp` 默认值的格式差异（可忽略）
- 其他都一致

**Production smoke**（上生产前在 staging 跑）：

```bash
# 拿生产 DB 的 schema dump
pg_dump --schema-only -U yuanju -h prod-db yuanju > /tmp/prod_schema.sql

# 起空容器导入
docker run -d --name pg_smoke -e POSTGRES_PASSWORD=test postgres:16-alpine
docker exec -i pg_smoke psql -U postgres -d test_db < /tmp/prod_schema.sql

# 跑新分支的 backend 启动
DATABASE_URL=... ./backend  # 应当观察到 Migrate(ModeStartup) Skipped=[1]
```

### 8.3 测试禁忌

- ❌ 不允许测试连开发者本机 dev DB（用 testcontainers）
- ❌ 不允许测试共享容器（每个测试独立 spinUpPG）
- ❌ 不允许测试 fixture 跨目录共享（good_v2 / bad_v2 互不污染）

### 8.4 TDD 节奏

- RED：先写 6 个集成测试（编译失败因为 `Migrate(mode)` 还没实现）
- GREEN-1：写 `migrations.go` 让前 3 个测试过（基础流程）
- GREEN-2：补足错误隔离 / dry-run 让剩下 3 个测试过
- 然后才提取 baseline.sql（顺序很重要：先有运行框架再喂内容）
- 最后改 `database.go` + `main.go` 接线

---

## 9. 范围估计

- baseline 提取：1-2 小时手工粘 + 1 次 schema parity check
- `migrations.go` + 6 集成测试：~250 行 Go
- `main.go` + `database.go` 改造：~50 行 Go
- `00001_baseline.sql`：~900 行 SQL（提取自 database.go）
- 文档：本 spec + plan + 实现 commit messages

**预估时长：1.5 个工作日**

---

## 10. 后续 spec

| 主题 | 触发条件 |
|---|---|
| 可观测性骨架 | 本 spec 落地后；用 `evt=migrate_run` 日志作为信号源之一 |
| 大文件拆分 | 与 ops 工作解耦，可任何时候做 |
