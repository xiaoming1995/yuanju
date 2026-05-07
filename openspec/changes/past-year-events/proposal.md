## Why

用户在查看命盘后，希望了解命理角度对以往年份的推算——某年会有婚恋动象、某年财运变动、某年健康需注意——但现有「流年精批」是单年按需生成的散文报告，无法一眼纵览全部人生过往。本功能提供基于算法信号检测 + AI 语言组织的过往年份事件推算，让用户看到完整的命理人生轨迹。

## What Changes

- **新增**：算法引擎 `pkg/bazi/event_signals.go`，对每一流年计算事件信号（婚恋、事业、财运、健康、迁变）
- **新增**：`POST /api/bazi/past-events-stream/:chart_id` SSE 流式接口，一次生成所有过往年份事件推算
- **新增**：DB 表 `ai_past_events`，缓存生成结果（命盘维度，一次生成长期复用）
- **新增**：Admin Prompt 管理支持 `past_events` 模块
- **新增**：前端 `PastEventsPage.tsx`，展示纵向时间轴，按大运分组，SSE 流式渲染

## Capabilities

### New Capabilities

- `past-events-signal-engine`：基于规则的流年事件信号检测算法，输入原局四柱+大运+流年干支，输出结构化信号列表（婚恋/事业/财运/健康/迁变及其命理证据）
- `past-events-report`：过往年份事件推算报告，接收算法信号 → AI 组织语言 → SSE 流式输出，每年 2-3 句自然语言描述，全部过往年份覆盖

### Modified Capabilities

（无现有 spec 级行为变更）

## Impact

- `pkg/bazi/event_signals.go`：新文件
- `internal/handler/bazi_handler.go`：新增 `HandlePastEventsStream` handler
- `internal/service/report_service.go`：新增 `GeneratePastEventsStream` 函数
- `internal/repository/`：新增 `past_events_repository.go`
- `pkg/database/database.go`：新增 `ai_past_events` 表 DDL
- `frontend/src/pages/PastEventsPage.tsx`：新页面
- `frontend/src/lib/api.ts`：新增 SSE 请求函数
- `frontend/src/App.tsx`：注册新路由
- 无 breaking change，现有流年精批功能不受影响
