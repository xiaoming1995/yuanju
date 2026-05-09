## Why

"过往事件推算"页面目前只覆盖过往年份，用户无法在同一视图中看到未来流年的命理走势，需要跳转其他页面才能了解未来运势，体验割裂。将未来年份纳入同一时间轴，能让用户在一个页面内完整感知人生各阶段的命理走势。

## What Changes

- 后端算法函数 `GetPastYearSignals` 改为 `GetAllYearSignals`，同时扫描过往和未来所有流年信号
- AI Prompt 模板扩展，传入全部流年数据（含未来），生成统一格式的叙述内容
- 前端 `PastEventsPage` 时间轴向后延伸，展示全部剩余大运的年份卡片
- 未来年份卡片视觉区分：虚线边框 + 略透明 +「未来」角标
- 旧缓存（仅含过往年份）自动失效，重新生成时包含全部年份

## Capabilities

### New Capabilities

- `future-liunian-display`: 未来流年展示 — 在现有"过往事件推算"页面时间轴底部，连续展示全部剩余大运的年份卡片，视觉上与过往区分

### Modified Capabilities

- `past-year-events`: 算法函数重命名（GetPastYearSignals → GetAllYearSignals），扫描范围从"过往年份"扩展至"全部年份"；AI Prompt 不变，但传入数据量增加

## Impact

- `backend/pkg/bazi/event_signals.go`：函数重命名，移除 `ln.Year > currentYear` 过滤条件
- `backend/internal/service/report_service.go`：调用处更新函数名，DayunList 构建去掉 StartYear 截止判断
- `frontend/src/pages/PastEventsPage.tsx`：年份卡片增加 `isFuture` 判断，未来卡片用虚线边框+透明度+角标区分
- 不涉及新 API 端点或数据库 schema 变更
- 旧 `ai_past_events` 缓存需清除（不含未来年份数据）
