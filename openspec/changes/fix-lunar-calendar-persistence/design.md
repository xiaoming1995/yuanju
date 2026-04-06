## Context

农历排盘功能已上线，`Calculate` 引擎支持 `calendarType="lunar"` + `isLeapMonth=true` 参数，内部先将农历转公历再排盘。但 `bazi_charts` 表从未持久化这两个入参字段，导致：

- `GetHistoryDetail` 以 `"solar", false` 硬编码重排 → 四柱错误
- `GenerateReport` 同样以 `"solar", false` 重排 → AI 报告基于错误命盘

现有存量数据全部是公历入参（农历功能刚上线），补充默认值 `'solar'` / `false` 不影响已有记录。

## Goals / Non-Goals

**Goals:**
- 在 `bazi_charts` 表持久化 `calendar_type` 和 `is_leap_month` 字段
- 所有对 `bazi.Calculate()` 的回调点（历史详情、AI 报告生成）均使用存储的历法参数
- 存量数据自动获得安全的默认值（`'solar'` / `false`）

**Non-Goals:**
- 不修改前端逻辑（已正确传参）
- 不修改引擎 `Calculate()` 函数签名
- 不处理经度 / 早子时的持久化（这些是独立问题）

## Decisions

### 1. 增量 ALTER TABLE（非重建表）

使用 `ALTER TABLE ... ADD COLUMN IF NOT EXISTS` 在 `database.go` 的 `Migrate()` 函数中追加增量迁移。

**理由**：与项目现有的迁移模式完全一致（参考 `content_structured` 和 `chart_hash` 的增量迁移），对已有数据零破坏。

**备选方案**：使用独立迁移文件 — 但项目未引入迁移管理框架，维持现有模式更一致。

### 2. 默认值选择

- `calendar_type` 默认 `'solar'`：旧数据全部为公历输入
- `is_leap_month` 默认 `false`：旧数据无闰月标记

**理由**：农历功能刚上线不久，存量数据均为公历，默认值逻辑安全。

### 3. 所有重排调用点统一修复

`bazi_handler.go` 中有 3 处对 `bazi.Calculate()` 的调用，其中 2 处（`GenerateReport` 第 107 行、`GetHistoryDetail` 第 246 行）硬编码了 `"solar", false`，需改为从 `chart` 对象读取。

## Risks / Trade-offs

- **[风险] 遗漏调用点** → 使用 `grep` 全局搜索 `bazi.Calculate` 确保所有重排点均已修复
- **[风险] 旧数据 calendar_type 为空** → 使用 `DEFAULT 'solar'` 且 Go 代码中用 `COALESCE` 或空值判断兜底
- **[权衡] 不持久化经度和早子时** → 这些是独立问题，本次不扩大范围，避免改动过大
