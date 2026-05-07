## 1. 数据库与种子数据

- [x] 1.1 在 `pkg/database/database.go` 新增 `ai_past_events` 表 DDL（字段：id, chart_id UNIQUE, content_structured JSONB, model, created_at）
- [x] 1.2 在 `pkg/seed/seed.go` 写入默认 `past_events` Prompt 模板（含 {{.Gender}} {{.DayGan}} {{.NatalSummary}} {{.YearsData}} 变量）

## 2. 算法引擎：流年事件信号检测

- [x] 2.1 新建 `pkg/bazi/event_signals.go`，定义 `EventSignal`（Type, Evidence）和 `YearSignals`（Year, Age, GanZhi, DayunGanZhi, Signals）结构体
- [x] 2.2 实现婚恋信号检测：男命财星透干、女命官星透干、日支六合/六冲引动夫妻宫、桃花临命
- [x] 2.3 实现事业信号检测：官星/印星透干且为用神、驿马动
- [x] 2.4 实现财运信号检测：财星透干得令（财运_得）、财星受克/劫财夺财（财运_损）
- [x] 2.5 实现健康信号检测：流年天干克日干、地支冲日支
- [x] 2.6 实现迁变信号检测：驿马临命、流年与年柱相冲
- [x] 2.7 实现 `GetPastYearSignals`：过滤 minAge、过滤未来年份、按年份升序返回

## 3. 后端：Repository 与 Service

- [x] 3.1 新建 `internal/repository/past_events_repository.go`，实现 `GetPastEvents(chartID)` 和 `CreatePastEvents(chartID, contentStructured, model)` 函数
- [x] 3.2 在 `internal/model/model.go` 新增 `AIPastEvents` 结构体和 `PastEventsTemplateData` 模板数据结构
- [x] 3.3 在 `internal/service/report_service.go` 新增 `GeneratePastEventsStream(chartID string, w http.ResponseWriter)` 函数，实现：检查缓存 → 算法扫描信号 → 构建 Prompt → SSE 流式 AI 调用 → 解析 JSON → 存库

## 4. 后端：Handler 与路由

- [x] 4.1 在 `internal/handler/bazi_handler.go` 新增 `HandlePastEventsStream` handler（验证 chart 归属、调用 service、SSE headers）
- [x] 4.2 在 `cmd/api/main.go` 注册路由：`bazi.POST("/past-events-stream/:chart_id", middleware.Auth(), handler.HandlePastEventsStream)`

## 5. 前端：API 层

- [x] 5.1 在 `src/lib/api.ts` 新增 `streamPastEvents(chartId: string, onChunk, onDone, onError)` SSE 请求函数（与现有 streamReport 模式一致）

## 6. 前端：页面与组件

- [x] 6.1 新建 `src/pages/PastEventsPage.tsx`，实现纵向时间轴展示（按大运分组，每年展示干支+事件类型标签+叙述文字）
- [x] 6.2 实现 SSE 流式渲染：边接收边追加年份卡片，完成后显示"生成完毕"状态
- [x] 6.3 在 `src/App.tsx` 注册路由 `/bazi/:chartId/past-events`
- [x] 6.4 在命盘详情页（`HistoryDetailPage` 或 `BaziResultPage`）添加"过往事件推算"入口按钮，跳转到新页面
