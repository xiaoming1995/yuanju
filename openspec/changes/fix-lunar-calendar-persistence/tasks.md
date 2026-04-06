## 1. 数据库迁移

- [x] 1.1 在 `database.go` 的 `Migrate()` 中添加增量 ALTER TABLE：`bazi_charts` 表新增 `calendar_type VARCHAR(10) DEFAULT 'solar'` 和 `is_leap_month BOOLEAN DEFAULT false`

## 2. Model 层

- [x] 2.1 在 `model/model.go` 的 `BaziChart` 结构体中新增 `CalendarType string` 和 `IsLeapMonth bool` 字段（含 JSON tag）

## 3. Repository 层

- [x] 3.1 修改 `CreateChart`：INSERT 语句新增 `calendar_type` 和 `is_leap_month` 两列
- [x] 3.2 修改 `GetChartByID`：SELECT 语句新增读取 `calendar_type` 和 `is_leap_month`，空值兜底为 `'solar'` / `false`
- [x] 3.3 修改 `GetChartsByUserID`：SELECT 语句同步新增读取这两个字段

## 4. Handler 层修复

- [x] 4.1 修改 `Calculate` handler：`BaziChart` 落库时传入 `input.CalendarType` 和 `input.IsLeapMonth`
- [x] 4.2 修改 `GetHistoryDetail`：用 `chart.CalendarType` 和 `chart.IsLeapMonth` 替代硬编码的 `"solar", false`
- [x] 4.3 修改 `GenerateReport`：用 `chart.CalendarType` 和 `chart.IsLeapMonth` 替代硬编码的 `"solar", false`

## 5. 验证

- [x] 5.1 ~~编写单元测试~~ 现有 `TestCalculateLunarInput` 已覆盖农历闰月排盘一致性验证，无需新增
- [x] 5.2 全局搜索 `bazi.Calculate` 确认无遗漏的硬编码调用点（发现并修复了 `report_service.go` 第4处）
- [x] 5.3 运行现有测试套件（`go test ./...`），确保无回归
