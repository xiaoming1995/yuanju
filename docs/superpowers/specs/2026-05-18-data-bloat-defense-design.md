# 数据膨胀防御 设计文档

**日期**：2026-05-18
**作者**：刘明（brainstorming 产物）
**范围**：backend 数据保留 / token 月度汇总 / 自动清理调度

---

## 1. 背景

2026-05-18 线上 backend 启动失败：

```
2026/05/18 10:16:06 ✅ 数据库连接成功
2026/05/18 10:16:06 数据库迁移失败: pq: could not extend file
  "base/16384/3455": No space left on device (53100)
```

根因不是单条 DDL，是 Postgres 数据盘满 —— 暴露三个独立问题：
1. **表膨胀无管理**：多张 AI 缓存表 + `ai_requests_log` + `token_usage_logs` 只增不减
2. **迁移失败阻塞整启动**：单条 ALTER 失败 → backend 起不来
3. **运维盲区**：磁盘 / 表大小没有任何告警

本 spec 只解决 **问题 1**。其余两个问题留给后续独立 spec。

---

## 2. 目标与非目标

**目标：**
- 6 张 AI 缓存表 + 1 张 AI 请求日志表按 `created_at` 滑动窗口删除（默认 90 天，admin 可调）
- `token_usage_logs` 按月汇总到 `token_usage_logs_monthly` 后删原始行
- 调度由 backend 内嵌 goroutine 每 24h 跑一次
- 阈值 / 开关 / 运行时刻通过 `algo_config` 暴露给 admin 后台
- 提供 `--cleanup-once` CLI flag，便于 ops 手动触发与烟雾测试

**非目标（明确不做）：**
- ❌ 表大小 / 磁盘告警（留给"可观测性骨架" spec）
- ❌ 迁移版本化与回滚（留给"迁移安全网" spec）
- ❌ 用户业务表（命盘 / 合盘 readings / participants / evidences / 账号等）的清理 —— **永不删用户数据**
- ❌ `cleanup_history` 审计表
- ❌ 错过 tick 的 catch-up 补跑
- ❌ "Run now" admin API（已由 CLI flag 覆盖）
- ❌ 主动 VACUUM / VACUUM FULL（依赖 autovacuum）

---

## 3. 决策摘要

| 维度 | 决定 | 备选与放弃理由 |
|---|---|---|
| 保留策略 | 缓存 / 日志按时间删，token 按月汇总 | 一律按时间删会丢失成本回溯；按活跃度删要 join users 表，复杂度高 |
| 调度机制 | backend in-process goroutine ticker | 独立 cron 容器需新 binary，host crontab 依赖部署文档；当前 docker-compose 单实例无并发隐患 |
| 阈值存储 | `algo_config` 表，admin 可调 | 硬编码需改代码重启；env var 配置链路冗余 |
| AI 缓存默认 | 90 天 | 30 天会让用户回看老报告全部重生成 LLM；180 天磁盘节省有限 |
| token rollup 粒度 | (user_id, model, year_month) | 用户维度可做账单回溯；保留 model 维度才能算分模型成本 |
| token rollup 字段 | 只存 token 计数，**不存 cost** | 项目当前 cost 即时由 `costFn(llm_providers.input_price_cny, …)` 计算，rollup 表沿用该模式，避免双源 |

**Brainstorm 起初假设是「4 张缓存表」，自检阶段核对真实 schema 发现还有 `ai_polished_reports`、`ai_liunian_reports`（缓存类）和 `ai_requests_log`（日志类），都纳入本 spec。**

---

## 4. 架构

```
┌──────────────────────────────────────────────────────────────┐
│                       backend (cmd/api)                       │
│                                                              │
│   main.go ─► server.Run()                                    │
│         │                                                    │
│         └─► startCleanupScheduler(ctx, deps)                 │
│                │                                             │
│                └─► time.Ticker (24h, 等到 RunHour:00 再首跑)   │
│                       │                                      │
│                       ▼                                      │
│              ┌─────────────────────────┐                     │
│              │ cleanup_service.RunOnce │  ◄── algo_config    │
│              └────────┬────────────────┘   （ retention,      │
│                       │                       enabled,        │
│       ┌───────────────┼───────────────┐       run_hour ）     │
│       ▼               ▼               ▼                       │
│  Cache+log repos  token_usage_repo   logs (stdout)           │
│  .DeleteOlderThan  .RollupClosedMonths*                      │
│  (7 张表)          (1 个事务)                                 │
└──────────────────────────────────────────────────────────────┘

                        ┌──── 新表 ────┐
                        │ token_usage_   │
                        │ logs_monthly   │ 由 rollup 写入；只增不删
                        └────────────────┘
```

**核心选择：**
- **In-process scheduler** —— Docker Compose 单实例，零新组件
- **algo_config 驱动** —— 复用现有 admin 配置链路
- **每张表独立 DeleteOlderThan** —— 单表失败不影响其他
- **token 单事务原子化** —— INSERT...SELECT + DELETE 在同事务，幂等可重跑

---

## 5. Components & boundaries

### 5.1 范围内的表

**6 张缓存表（按 `created_at` 删除）：**

| 表名 | 备注 |
|---|---|
| `ai_reports` | 主报告缓存；FK `bazi_charts(id) ON DELETE CASCADE` |
| `ai_polished_reports` | 润色版报告；UNIQUE(chart_id)；FK ON DELETE CASCADE |
| `ai_liunian_reports` | 流年报告；UNIQUE(chart_id, target_year)；FK ON DELETE CASCADE |
| `ai_past_events` | 过往事件缓存；UNIQUE(chart_id)；FK ON DELETE CASCADE |
| `ai_dayun_summaries` | 大运总览段缓存；UNIQUE(chart_id, dayun_index)；FK ON DELETE CASCADE |
| `ai_compatibility_reports` | 合盘 AI 报告；FK `compatibility_readings(id) ON DELETE CASCADE` |

**1 张请求日志表（按 `created_at` 删除）：**

| 表名 | 备注 |
|---|---|
| `ai_requests_log` | AI 请求耗时 / status / error_msg；当前 admin_repository 写入 |

**1 张 token 用量表（按月汇总后删）：**

| 表名 | 真实列 |
|---|---|
| `token_usage_logs` | `user_id UUID`（FK users(id) ON DELETE SET NULL）、`chart_id UUID`、`call_type`、`model`、`provider_id`、`prompt_tokens`、`completion_tokens`、`total_tokens`、`reasoning_tokens`、`cache_hit_tokens`、`cache_miss_tokens`、`input_content TEXT`、`output_content TEXT`、`created_at`。**没有 cost 列** —— 成本由 `service.CalcCost` 实时计算 |

`input_content` / `output_content` TEXT 是 token_usage_logs 真正的膨胀源；删除源行就一并释放。

### 5.2 新建文件

| 文件 | 职责 |
|---|---|
| `backend/internal/service/cleanup_service.go` | 编排：读 algo_config → 调 8 个 repository 方法 → 写日志 |
| `backend/internal/service/cleanup_service_test.go` | 单测：mock repo，验证调用顺序、错误隔离、配置 clamp |
| `backend/internal/repository/cleanup_integration_test.go` | 集成测试：真 Postgres，验证 token rollup SQL |
| `frontend/tests/cleanup-config.test.mjs` | grep 单测：AlgoConfigPage 暴露 3 个新键 |

### 5.3 修改文件

| 文件 | 改什么 |
|---|---|
| `backend/cmd/api/main.go` | server 起来后 `go cleanupSvc.StartScheduler(ctx)`；新增 `--cleanup-once` flag 分支 |
| `backend/pkg/database/database.go` | 新建 `token_usage_logs_monthly` 表 DDL；algo_config 注入 3 个新键的默认值 |
| `backend/internal/repository/repository.go`（含 ai_reports） | 加 `DeleteAIReportsOlderThan(ctx, cutoff) (int64, error)` |
| `backend/internal/repository/polished_report_repo.go` | 加 `DeleteOlderThan` |
| `backend/internal/repository/liunian_repository.go` | 加 `DeleteOlderThan` |
| `backend/internal/repository/past_events_repository.go` | 加 `DeleteOlderThan` |
| `backend/internal/repository/dayun_summary_repository.go` | 加 `DeleteOlderThan` |
| `backend/internal/repository/compatibility_repository.go` | 加 `DeleteAICompatibilityReportsOlderThan`（**只动 `ai_compatibility_reports`，不动 `compatibility_readings/participants/evidences`**） |
| `backend/internal/repository/admin_repository.go` | 加 `DeleteRequestLogsOlderThan` |
| `backend/internal/repository/token_usage_repository.go` | 加 `RollupClosedMonthsAndDelete(ctx) (RollupReport, error)` |
| `backend/internal/service/algo_config_service.go` | 加 3 个 typed getter：`CleanupEnabled()` / `CleanupRetentionDays()` / `CleanupRunHour()`，内部 clamp |
| `frontend/src/pages/admin/AlgoConfigPage.tsx` | "数据保留" 章节，3 个字段；外加只读"最近一次清理"时间戳展示 |

### 5.4 关键 interfaces

```go
// backend/internal/service/cleanup_service.go

type AIReportCleaner    interface { DeleteAIReportsOlderThan(ctx context.Context, cutoff time.Time) (int64, error) }
type PolishedCleaner    interface { DeleteOlderThan(ctx context.Context, cutoff time.Time) (int64, error) }
type LiunianCleaner     interface { DeleteOlderThan(ctx context.Context, cutoff time.Time) (int64, error) }
type PastEventsCleaner  interface { DeleteOlderThan(ctx context.Context, cutoff time.Time) (int64, error) }
type DayunSummaryCleaner interface { DeleteOlderThan(ctx context.Context, cutoff time.Time) (int64, error) }
type CompatReportCleaner interface { DeleteAICompatibilityReportsOlderThan(ctx context.Context, cutoff time.Time) (int64, error) }
type RequestLogCleaner   interface { DeleteRequestLogsOlderThan(ctx context.Context, cutoff time.Time) (int64, error) }
type TokenUsageRollup    interface { RollupClosedMonthsAndDelete(ctx context.Context) (RollupReport, error) }

type Clock interface { Now() time.Time }
type RealClock struct{}
func (RealClock) Now() time.Time { return time.Now() }

type CleanupService struct {
    aiReports     AIReportCleaner
    polished      PolishedCleaner
    liunian       LiunianCleaner
    pastEvents    PastEventsCleaner
    dayunSummary  DayunSummaryCleaner
    compatReports CompatReportCleaner
    requestLogs   RequestLogCleaner
    tokenUsage    TokenUsageRollup
    cfg           *AlgoConfigService
    clock         Clock
    logger        *log.Logger
}

func (s *CleanupService) RunOnce(ctx context.Context) RunReport
func (s *CleanupService) StartScheduler(ctx context.Context)

type RunReport struct {
    StartedAt time.Time
    Duration  time.Duration
    Tables    []TableResult   // 每张表独立结果（7 项：6 缓存 + 1 日志）
    Rollup    RollupReport    // token rollup
    Errors    []string        // 顶层错误汇总
}

type TableResult struct {
    Name    string
    Deleted int64
    Err     error
}

type RollupReport struct {
    MonthsAggregated      int
    RowsInsertedOrUpdated int64
    SourceRowsDeleted     int64
    Err                   error
}
```

### 5.5 新表 schema

```sql
CREATE TABLE IF NOT EXISTS token_usage_logs_monthly (
    user_id            UUID         NOT NULL,           -- 孤儿调用用 ORPHAN_USER_UUID 兜底
    model              VARCHAR(100) NOT NULL,
    year_month         CHAR(7)      NOT NULL,           -- '2026-04'
    call_count         BIGINT       NOT NULL DEFAULT 0,
    prompt_tokens      BIGINT       NOT NULL DEFAULT 0,
    completion_tokens  BIGINT       NOT NULL DEFAULT 0,
    reasoning_tokens   BIGINT       NOT NULL DEFAULT 0,
    total_tokens       BIGINT       NOT NULL DEFAULT 0,
    aggregated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, model, year_month)
);
CREATE INDEX IF NOT EXISTS idx_token_usage_logs_monthly_ym
  ON token_usage_logs_monthly (year_month);
```

字段单位与 `token_usage_logs` 列对齐 —— 全部 token 计数；**不存 cost**。Admin 后台显示成本时用现有 `service.CalcCost(model, prompt_tokens, completion_tokens)` 对汇总值即时计算（线性，结果与单行求和等价）。

**孤儿处理：** 源表 `user_id` 可空（FK `ON DELETE SET NULL`）；Postgres PK 不能含 NULL。在 rollup SELECT 里用 `COALESCE(user_id, '00000000-0000-0000-0000-000000000000'::uuid)` 把孤儿统一归到 sentinel UUID 桶里，业务侧解读 = "已注销用户的合并统计"。Admin UI 显示时识别该 UUID 做特殊标签。

### 5.6 新增 algo_config 键

| key | type | default | clamp | 含义 |
|---|---|---|---|---|
| `cleanup_enabled` | bool | `true` | — | 关掉则 scheduler 仍 tick 但早返回 |
| `cleanup_retention_days` | int | `90` | `[1, 3650]` | 7 张表的统一超期阈值 |
| `cleanup_run_hour` | int | `3` | `[0, 23]` | 每天运行时刻（24h 制） |

读取入口走 `algo_config_service.go` typed getter，clamp 在 getter 里做（防 admin 错填）。

### 5.7 边界规则

- `cleanup_service` **不写 SQL**，全部走 repository
- 每张表的清理**彼此独立** —— 单表失败收集到 `Tables[i].Err`，不中断其他表
- token rollup **在单事务里** —— INSERT...SELECT + DELETE 同事务
- `algo_config` 是阈值**唯一来源** —— service 每次 tick 重读（admin 改后下一轮生效）
- service **不知道 bazi 算法的任何东西** —— 它只是个保留 / 汇总编排器
- `Clock` 接口注入 —— 单测能"伪造 2026-04-30 23:00"，绕开真实时钟

---

## 6. Data flow

### 6.1 Scheduler tick

```
backend startup
       │
       ▼
go StartScheduler(ctx)
       │
       ▼
loop:
  next := today at RunHour:00, or tomorrow if past
  sleep until next  ◄─ ctx.Done() 中断退出
       │
       ▼
  RunOnce(ctx)  ─► 见 6.2
       │
       ▼
  log structured report
       │
       └──► loop again
```

### 6.2 RunOnce 步骤

| 步骤 | 操作 |
|---|---|
| ① 读 algo_config | `enabled, retention_days, run_hour := cfg.LoadCleanup()`；enabled=false → 写 `[cleanup] disabled` 日志、立即返回 |
| ② 计算 cutoff | `cutoff := clock.Now().Add(-retention_days * 24h)` |
| ③ 表清理（顺序、独立） | 对 7 张表各执行 `DELETE FROM <t> WHERE created_at < $cutoff;`；`RowsAffected` 写 `TableResult.Deleted`；错误进 `TableResult.Err`，**不中断后续表** |
| ④ token rollup（单事务） | 见 6.3 |
| ⑤ 写一条结构化日志 | 见 6.4 |

### 6.3 token rollup SQL

```sql
BEGIN;

INSERT INTO token_usage_logs_monthly (
    user_id, model, year_month,
    call_count, prompt_tokens, completion_tokens,
    reasoning_tokens, total_tokens,
    aggregated_at
)
SELECT
    COALESCE(user_id, '00000000-0000-0000-0000-000000000000'::uuid) AS user_id,
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
GROUP BY COALESCE(user_id, '00000000-0000-0000-0000-000000000000'::uuid), model, year_month
ON CONFLICT (user_id, model, year_month) DO UPDATE SET
    call_count        = EXCLUDED.call_count,
    prompt_tokens     = EXCLUDED.prompt_tokens,
    completion_tokens = EXCLUDED.completion_tokens,
    reasoning_tokens  = EXCLUDED.reasoning_tokens,
    total_tokens      = EXCLUDED.total_tokens,
    aggregated_at     = NOW();

DELETE FROM token_usage_logs
WHERE created_at < date_trunc('month', NOW());

COMMIT;
```

**幂等性证明：**
- `ON CONFLICT DO UPDATE` —— 同月再聚合一次结果一致
- 范围条件 `< date_trunc('month', NOW())` —— 永不碰当月数据
- 单事务 —— INSERT 成功 / DELETE 失败时 ROLLBACK，无半状态
- 重跑安全 —— 第二次 INSERT 对同样 (user, model, ym) 用 UPDATE 覆盖，第二次 DELETE 因第一次已删，0 行

### 6.4 Logging（一行结构化）

```
{
  "evt": "cleanup_run",
  "started_at": "2026-05-19T03:00:00Z",
  "duration_ms": 8243,
  "tables": [
    {"name": "ai_reports", "deleted": 1234},
    {"name": "ai_polished_reports", "deleted": 612},
    {"name": "ai_liunian_reports", "deleted": 805},
    {"name": "ai_past_events", "deleted": 890},
    {"name": "ai_dayun_summaries", "deleted": 412},
    {"name": "ai_compatibility_reports", "deleted": 73},
    {"name": "ai_requests_log", "deleted": 2456}
  ],
  "rollup": {
    "months_aggregated": 4,
    "rows_inserted_or_updated": 312,
    "source_rows_deleted": 9180
  },
  "errors": []
}
```

`grep evt=cleanup_run` 即可从 docker logs / journalctl 拉历史。

### 6.5 关键决策

| 选择 | 为什么 |
|---|---|
| 表删除顺序、不在事务里 | 单表失败不污染其他；DELETE 锁很小；3 AM 流量低 |
| token rollup 在单事务里 | 防止"汇总写入成功但删除失败 → 下次再汇总 → 双倍计数" |
| `ON CONFLICT DO UPDATE` | 让 rollup 幂等；同月二次运行结果一致 |
| 跳过当月（`< date_trunc('month', NOW())`） | 当月还在累计，提前汇总不准 |
| 错过 3 AM 不补跑 | 90 天保留期下，少跑一天无差别 |

---

## 7. Error handling & 防御性设计

### 7.1 故障模式 → 应对

| 场景 | 行为 |
|---|---|
| DB 连接挂 / 读 algo_config 失败 | log error, return；下次 tick 重试；scheduler goroutine 不退出 |
| 某张表 DELETE 报错 | 收到 `TableResult.Err`，**继续清下一张表**；其他表不受影响 |
| token rollup 事务里 INSERT 失败 | 整个事务 ROLLBACK，日志记错误，24h 后重试（幂等保证安全） |
| token rollup INSERT 失败 = 磁盘还满 | 上一步表 DELETE 已 commit；autovacuum 终会回收页，下次 INSERT 成功 |
| Backend crash mid-cleanup | 各表 DELETE 是 per-statement auto-commit，已删的不丢；token rollup 在事务内，无半状态 |
| algo_config 值异常（retention_days=-1） | typed getter 内部 clamp 到 `[1, 3650]`；run_hour clamp 到 `[0,23]`；enabled 默认 true |
| 用户首次开启时磁盘已经爆了 | 这次 cleanup 自身可能也跑不动（DELETE 也写 WAL）。首启文档建议先手动 `VACUUM FULL` 一次释放空间，之后稳态靠 autovacuum |

### 7.2 并发 / 幂等保证

- 单实例 → 单 scheduler goroutine → 无双跑
- `ON CONFLICT DO UPDATE` → rollup 重跑结果一致
- token 表 DELETE 范围 = `created_at < date_trunc('month', NOW())` → 永不碰当月数据，重跑安全

### 7.3 手动触发 / 烟雾测试

新增 CLI 分支：

```bash
go run ./cmd/api --cleanup-once
```

行为：跑完 RunOnce 后打印 RunReport JSON 到 stdout，**不启动 HTTP server**，进程退出码 0（即使个别表失败 —— 因为 RunReport 已经把错误暴露）。

---

## 8. Testing strategy

### 8.1 单测（`cleanup_service_test.go`）

mock 8 个 repository，验证：
- `disabled=true` → 直接返回，不调任何 repo
- `retention_days=90` → 调每个 cleaner 时 cutoff 参数 = clock.Now() - 90d
- 让 `ai_reports.DeleteAIReportsOlderThan` 返回 error → 断言其他 6 张表的 cleaner 仍被调用、`TableResult[0].Err != nil`、其他 `TableResult.Err == nil`
- token rollup 调一次、`Rollup.Err == nil`
- algo_config 异常值 `retention_days=-1` → clamp 到 1，传给 cleaner 的 cutoff = now() - 1d
- Clock 接口注入 → 单测固定 "2026-04-30 23:00"，验证 cutoff 算对

### 8.2 集成测试（`cleanup_integration_test.go`）

只测 1 件最难单测的事：token rollup SQL 的正确性。

用 `testcontainers-go` 起临时 PG 容器（项目当前没有集成测试基建，本次引入）。

测试步骤：
1. 起 PG 容器，跑 DDL 迁移
2. 插入 6 个月 token_usage_logs 数据（含当月）—— 多用户、多模型、含 1 个 `user_id IS NULL` 的孤儿行
3. 跑 `RollupClosedMonthsAndDelete`
4. 断言：
   - `token_usage_logs_monthly` 5 个月（不含当月）有正确聚合行
   - `token_usage_logs` 只剩当月行
   - 聚合行的 SUM 与插入数据吻合
   - 孤儿行（user_id NULL）被聚合到 sentinel UUID `00000000-0000-0000-0000-000000000000` 桶下
5. 再跑一次（幂等）→ monthly 行数不变、source 仍只剩当月行

### 8.3 前端 grep 单测（`cleanup-config.test.mjs`）

断言 `AlgoConfigPage.tsx` 出现：
- `/cleanup_enabled/`
- `/cleanup_retention_days/`
- `/cleanup_run_hour/`

不做更深的 React 渲染测试 —— 与项目现有 grep 风格保持一致。

### 8.4 手动验收

合并前在本地或预发跑一遍：

| 测试项 | 预期 |
|---|---|
| 种子脚本：插 7 个月数据，`--cleanup-once` | report 打印 7 张表 deleted > 0，rollup months_aggregated=6 |
| `retention_days=0` | 7 张表全删除（不影响 token_usage_logs_monthly） |
| `retention_days=99999` | 7 张表 deleted=0；rollup 正常 |
| `cleanup_enabled=false` | RunReport 直接返回，所有 cleaner 未调 |
| admin 后台改 `retention_days` 为 30 | 下一次 tick 立刻按 30 天裁切 |
| `--cleanup-once` 跑两次 | 第二次表 deleted=0；rollup source_rows_deleted=0 |

### 8.5 测试禁忌

- ❌ 不允许写"调真 LLM"的测试
- ❌ 不允许测试里硬编码当前日期 —— 必须用注入的 `Clock`
- ❌ 不允许集成测试连开发者本地 dev DB（要么 testcontainers，要么跳过）

### 8.6 TDD 节奏

- 每个 RED commit 单独成 commit
- GREEN 拆成 ≤200 行/commit，便于回滚
- 单测先于实现；集成测试在 service + repository 写完后

---

## 9. 范围估计

- 后端：cleanup_service + 8 个 repo 扩展 + DDL + main.go 接线 ≈ 900-1100 行（含测试）
- 前端：AlgoConfigPage 增 3 字段 + 只读时间戳 ≈ 60 行 + 12 行测试
- 集成测试基础：首次引入 testcontainers-go ≈ 100 行 setup
- 文档：本 spec + plan + 实现 commit messages

**预估时长：2 个工作日**（比 brainstorm 时估的 1.5 天略长，因为表数比预想多 3 张）

---

## 10. 后续 spec（明确分支）

| 主题 | 触发条件 |
|---|---|
| 迁移版本化 + 回滚 | 本 spec 落地后，下一次部署前 |
| 可观测性骨架（磁盘 / 表大小 / SSE / 5xx 告警） | 本 spec 落地后，等线上跑稳一周 |
| 大文件拆分（`event_signals.go` 1998 行 / `ResultPage.tsx` 1322 行等） | 与 ops 工作解耦，可并行 |
