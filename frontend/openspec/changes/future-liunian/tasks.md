## 1. 后端：扩展信号扫描范围

- [x] 1.1 在 `pkg/bazi/event_signals.go` 中将 `GetPastYearSignals` 重命名为 `GetAllYearSignals`，并移除 `ln.Year > currentYear` 的过滤条件（改为扫描全部流年）
- [x] 1.2 在 `internal/service/report_service.go` 中更新调用处：`GetPastYearSignals` → `GetAllYearSignals`；同时移除 DayunList 构建中的 `dy.StartYear > currentYear` break 条件，改为包含所有大运

## 2. 前端：未来年份视觉区分

- [x] 2.1 在 `PastEventsPage.tsx` 顶部计算 `currentYear = new Date().getFullYear()`，传入年份列表渲染逻辑
- [x] 2.2 渲染年份卡片时增加 `isFuture = y.year > currentYear` 判断，未来卡片使用：虚线边框（`borderStyle: 'dashed'`）、`opacity: 0.75`、右上角「未来」角标（绝对定位小标签）

## 3. 数据迁移

- [x] 3.1 清空 `ai_past_events` 表中所有旧缓存（仅含过往年份的旧格式数据）：`DELETE FROM ai_past_events;`
