# LLM 成本软阈值预警（在 TokenUsagePage 上迭代）设计

> 日期：2026-05-19
> 关联背景：刚合入的 AI-driven year narrative 实际单命盘成本约 ¥0.36，是 spec 原估 ¥0.01 的 36 倍。需要可见性 + 预警机制防止账单意外。

## 问题

`feat/ai-driven-year-narrative` 合入后，过往事件推算每个命盘的 AI 生成成本被低估了 36 倍：

- spec 估计：¥0.01/命盘（单命盘约 3K tokens 输出）
- 真实数据：¥0.36/命盘（单命盘约 117K tokens 输出，9 段大运 × 13K tokens）

按 100 命盘/天估算，月成本约 ¥1000，与 spec 预期不同量级。需要尽快建立**成本可见性 + 软阈值预警机制**，让运营者能：
- 进入 admin 后台就能看到当前预算消耗百分比
- 超阈值时立即可见（红色 banner）
- 写结构化日志为未来对接外部告警（邮件 / 企微）留口子

## 决策

> 用户偏好：**「只看不挡」soft 模式 + 在原 TokenUsagePage 上迭代，不另起页面**

1. **完全建立在现有 token_usage_logs 之上** —— 零新表、零外部依赖。
2. **三档阈值**：单日 / 单月 / 单命盘成本上限，均以 CNY 为单位（不是 tokens）。
3. **检测双路**：admin dashboard 加载时实时聚合（用户可见 banner）+ backend 5 分钟 ticker（写结构化日志钩子）。
4. **超阈值 = 红色 banner**，不阻断 AI 调用、不降级到 template 模式。
5. **阈值编辑入口在 TokenUsagePage 标题栏的 ⚙️ 按钮**，不分散到 AlgoConfigPage。

## 架构

```
┌─────────────────────────────────────────────────────────┐
│  现有打点（不改）                                         │
│  AI 调用 → CreateTokenUsageLog → token_usage_logs        │
│  含字段：user_id, chart_id, call_type, tokens, provider │
│  + estimated_cost_cny 通过 llm_providers.price 算出      │
└─────────────────────────────────────────────────────────┘
                          ↓
                  ┌───────┴────────┐
                  ↓                ↓
        【检测路径 A：被动】     【检测路径 B：主动】
        admin dashboard 加载    backend 5-min ticker
        实时聚合 + banner       聚合 + 越界结构化日志
        （用户看到）            （grep / 外部告警钩子）
                  ↓                ↓
        ┌───────────────┐  ┌─────────────────────────┐
        │ TokenUsagePage │  │ stdout JSON 日志：       │
        │ 顶部状态条    │  │ {"evt":"cost_threshold_  │
        │ + 3 块卡片    │  │   exceeded", "scope":...} │
        │ + ⚙️ 阈值入口 │  │  写一次/告警类型/小时    │
        └───────────────┘  └─────────────────────────┘
```

## 数据流 + 接口

### 新增 1 个后端接口

```
GET /api/admin/token-usage/budget-status
```

返回结构：
```json
{
  "today": {
    "total_tokens": 506184,
    "total_cost_cny": 48.3,
    "threshold_cost_cny": 5.0,
    "exceeded": true,
    "exceeded_pct": 966
  },
  "this_month": {
    "total_tokens": 1234567,
    "total_cost_cny": 95.6,
    "threshold_cost_cny": 100.0,
    "exceeded": false,
    "exceeded_pct": 96
  },
  "top_charts": [
    {"chart_id": "9eddf08a-...", "total_cost_cny": 3.2, "calls": 9, "threshold_exceeded": true},
    {"chart_id": "b2a8045f-...", "total_cost_cny": 0.65, "calls": 9, "threshold_exceeded": false}
  ],
  "per_chart_threshold_cny": 1.0,
  "last_alerted_at": {
    "daily_total":   "2026-05-19T14:23:11Z",
    "monthly_total": null,
    "per_chart":     null
  }
}
```

### 聚合 SQL

复用 TokenUsagePage 现有的成本计算口径：

```sql
estimated_cost_cny =
  ( (prompt_tokens - cache_hit_tokens) * input_price_cny / 1000000 )
  + ( cache_hit_tokens * input_price_cny * 0.1 / 1000000 )
  + ( completion_tokens * output_price_cny / 1000000 )
```

`input_price_cny` 和 `output_price_cny` JOIN `llm_providers` 表。

```sql
-- 今日累计
SELECT SUM(<estimated_cost_cny>) FROM token_usage_logs ... WHERE created_at >= today_start
-- 本月累计
SELECT SUM(<estimated_cost_cny>) FROM token_usage_logs ... WHERE created_at >= month_start
-- 单命盘 TOP 5
SELECT chart_id, SUM(<estimated_cost_cny>) AS cost, COUNT(*) AS calls
FROM token_usage_logs ...
WHERE chart_id IS NOT NULL AND created_at >= NOW() - INTERVAL '7 days'
GROUP BY chart_id ORDER BY cost DESC LIMIT 5
```

### 触发时机

| 路径 | 时机 | 行为 |
|---|---|---|
| A（被动） | admin 打开 TokenUsagePage / 30s 自动刷新 | 渲染 banner + 3 stat card |
| B（主动） | backend 每 5 min ticker | 越界 + 距上次告警 ≥1 小时 → 写一行 JSON 日志 |

```go
// 内存状态
var lastAlertedAt map[string]time.Time // scope → 上次告警时间

func tick() {
    status := BuildBudgetStatus()
    for _, scope := range []string{"daily_total", "monthly_total", "per_chart"} {
        if status[scope].exceeded && now.Sub(lastAlertedAt[scope]) > time.Hour {
            log.Printf("%s", JSON{"evt":"cost_threshold_exceeded","scope":scope,...})
            lastAlertedAt[scope] = now
        }
    }
}
```

服务重启 → `lastAlertedAt` 清零 → 重启后若已超阈值立即重新打日志（这是想要的：ops 应该收到提醒）。

## 前端 UI（TokenUsagePage 叠加）

```
┌────────────────────────────────────────────────────┐
│ 📊 Token 用量统计              ⚙️ 编辑预算阈值     │ ← 标题栏（+ 阈值入口）
├────────────────────────────────────────────────────┤
│ ⚠️ 本月已用 ¥95.6（96% 月预算 ¥100）               │ ← banner（仅越界时显示）
│    今日已用 ¥48.3（966% 日预算 ¥5）                │
├────────────────────────────────────────────────────┤
│ ┌─────────────┐ ┌─────────────┐ ┌────────────────┐ │
│ │ 今日累计    │ │ 本月累计    │ │ 单命盘 TOP 5   │ │
│ │ ¥48.3       │ │ ¥95.6       │ │ 1. 9eddf08a ⚠ │ │
│ │ ↑ 966%      │ │ ↑ 96%       │ │ 2. b2a8045f    │ │
│ │ 日预算 ¥5   │ │ 月预算 ¥100 │ │ 3. ...         │ │
│ └─────────────┘ └─────────────┘ └────────────────┘ │
├────────────────────────────────────────────────────┤
│ [筛选栏：日期 + 查询] ← 现有，不动                  │
│ [汇总表格：按用户 × 模型 → tokens / 费用] ← 现有    │
│ [详情抽屉 / 内容 modal] ← 现有                      │
└────────────────────────────────────────────────────┘
```

**Banner** —— 红色背景，多条越界堆同一个 banner 内（不并列 3 个），不可关闭。仅 daily 或 monthly 超阈值时渲染；per_chart 阈值越界只在 TOP 5 卡片里以 ⚠ 标注。

**3 张 stat card**：
- 今日 / 本月：进度条样式，<80% 绿、80-100% 黄、>100% 红
- 单命盘 TOP 5：列出 chart_id 前 8 字符（hover 显示完整）+ 调用次数 + 总成本，超阈值前面 ⚠

**⚙️ 阈值 modal**：3 个 number 输入框（CNY 单位），保存调 `PUT /api/admin/algo-config/:key`。复用现有 algo_config 上传机制，需要在 `validKeys` 白名单加 3 个 cost_alert_* 键 + 数字 validation。

**自动刷新**：组件 mount 后 setInterval 30s 调一次 `budget-status`，卸载清理。**summary 表格保持手动"查询"刷**，不耦合两路。

## 阈值默认值

存 `algo_config` 表（KV），通过 `pkg/seed/seed.go` 启动时 `ON CONFLICT DO NOTHING` 插入：

```sql
INSERT INTO algo_config (key, value, description) VALUES
  ('cost_alert_daily_cost_cny',     '5',   '单日 AI 总成本告警阈值（CNY）'),
  ('cost_alert_monthly_cost_cny',   '100', '单月 AI 总成本告警阈值（CNY）'),
  ('cost_alert_per_chart_cost_cny', '1',   '单命盘 AI 总成本告警阈值（CNY）')
ON CONFLICT (key) DO NOTHING;
```

| 阈值 | 默认 | 推算依据 |
|---|---|---|
| 单日 | ¥5 | 单命盘 ¥0.36 × ~10 命盘/天 + 一些 report_stream ≈ ¥4。25% buffer |
| 单月 | ¥100 | 日 × 30 ≈ ¥150，月度保守取 ¥100 |
| 单命盘 | ¥1 | 健康命盘 ~¥0.36，¥1 ≈ 3 倍正常值。识别 prompt 膨胀 / 重生过多次 |

**理由：** soft 告警越早叫越好。调高比调低省事，主动权在 admin 手里。

## 范围边界

### 不做
- 强制阻断 / 限流（超阈值仍走 AI）
- Per-user 配额 / 多租户分摊
- 邮件 / 企微 / 钉钉外推
- 历史趋势图 / 月度报表导出
- 超预算自动降级 template 模式
- 多级阶梯告警（80% 黄、100% 红、150% 紫）

### 做但保持最小
- backend ticker 走现有 cleanup_service 的 cron pattern
- 内存 lastAlertedAt 重启会丢——可接受
- 单命盘 TOP 5 用 7 天滑窗，不长期累计

### 改动文件
后端（约 5 文件）：
- `internal/repository/token_usage_repository.go` —— 加 2 个聚合查询函数
- `internal/service/cost_alert_service.go`（新）—— `BuildBudgetStatus()` + ticker
- `internal/handler/token_usage_handler.go` —— 加 `AdminGetBudgetStatus`
- `internal/handler/algo_config_handler.go` —— `validKeys` 加 3 个键 + 数字校验
- `pkg/seed/seed.go` —— seed 默认阈值
- `cmd/api/main.go` —— 注册路由 + 启动 ticker

前端（1 文件）：
- `frontend/src/pages/admin/TokenUsagePage.tsx` —— banner + 3 stat card + ⚙️ modal + 自动刷新
- `frontend/src/lib/adminApi.ts` —— 加 `budgetStatus()` 方法

测试：
- `service/cost_alert_service_test.go`（新）—— `BuildBudgetStatus` + ticker 去重逻辑
- 前端不加新测试

## 风险与权衡

| 风险 | 缓解 |
|---|---|
| Banner 长期红色让人麻木 | 阈值可 admin 调；主动权在用户手里 |
| ticker 5min 频率太稀疏，业务突发被错过 | 30s 前端自刷补足；ticker 主要是为外部告警预留钩子 |
| 阈值改频繁失去意义 | 不强制；可观察后续是否需要"修改历史"审计 |
| `estimated_cost_cny` 依赖 llm_providers.price 正确 | 复用既有 TokenUsagePage 口径——根问题不在本次范围 |
| 30s 刷新轻微占 backend | SQL 聚合 < 10ms，无压力。命盘量过万时可加缓存（未来） |
| top_charts 7 天窗口可能错过更早异常 | 7 天足够覆盖最近调试期。需要更长可改可配置（不做） |
| 内存 lastAlertedAt 重启丢 | 重启后会立即重打日志 = 重新提醒 ops（想要的行为） |

## 接力

进入 `superpowers:writing-plans`，把本设计拆解为 step-by-step 实施计划，含：
- repository 聚合函数 + 单测
- cost_alert_service + ticker + 去重单测
- token_usage handler 加 budget-status endpoint
- algo_config_handler 扩 validKeys
- seed.go 插入默认阈值
- main.go 注册路由 + 启动 ticker
- TokenUsagePage UI 区块 + modal + 30s 刷新
- adminApi 新方法
