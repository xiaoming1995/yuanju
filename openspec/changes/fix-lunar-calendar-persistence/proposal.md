## Why

农历（含闰月）排盘后，`bazi_charts` 表没有持久化 `calendar_type` 和 `is_leap_month` 两个字段。首次排盘结果正确（引擎内部完成了 lunar→solar 转换），但查看历史记录或生成 AI 报告时，系统以 `"solar"` 强制重排，导致四柱结果与首次排盘完全不一致。这是上线农历排盘功能后的一个回归性 Bug，影响所有使用农历输入的用户。

## What Changes

- **数据库 `bazi_charts` 表**：新增 `calendar_type`（VARCHAR, 默认 `'solar'`）和 `is_leap_month`（BOOLEAN, 默认 `false`）两列
- **Model 层**：`BaziChart` 结构体新增 `CalendarType` 和 `IsLeapMonth` 字段
- **Repository 层**：`CreateChart` INSERT、`GetChartByID` SELECT、`GetChartsByUserID` SELECT 均新增这两个字段
- **Handler 层**：
  - `Calculate` handler 落库时传入用户原始的历法类型和闰月标识
  - `GetHistoryDetail` 和 `GenerateReport` 读取存储的历法参数进行重排（替代硬编码 `"solar", false`）

## Capabilities

### New Capabilities

_无新增能力。_

### Modified Capabilities

- `bazi-precision-engine`：八字排盘引擎的历法参数（`calendar_type` + `is_leap_month`）需要在数据持久化层正确保存和回放，以确保历史记录重排的一致性

## Impact

- **后端 5 个文件**：`model/model.go`、`database/database.go`（增量迁移）、`repository/repository.go`、`handler/bazi_handler.go`、`service/report_service.go`（如有重排逻辑）
- **数据库**：增量 ALTER TABLE，对已有记录自动填充默认值（`'solar'` / `false`），无破坏性
- **前端**：无需改动（前端已正确传递 `calendar_type` 和 `is_leap_month` 参数）
- **API 契约**：无变更，入参和返参格式不变
