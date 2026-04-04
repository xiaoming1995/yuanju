## Why

现有 AI 命理报告仅输出五章节纯文字内容，AI 的推理逻辑被刻意隐藏（Prompt 明确要求「在心中完成，不输出此步骤」）。用户读到的只有结论，看不到「为什么」，报告深度不足且缺乏专业说服力。随着用户对命理解读的期望提升，需要一套支持「精简/专业」双模式、可见推理链条的增强报告体系。

## What Changes

- **重写 `buildBaziPrompt()`**：新 Prompt 要求 AI 输出完整结构化 JSON，包含命局推理总览 + 每章精简/详细双版本
- **重写 `GenerateAIReport()` 解析逻辑**：适配新 JSON 结构（`analysis` + `chapters[]`），向下兼容旧 `content` 纯文字字段
- **数据库 Migration**：`ai_reports` 表新增 `content_structured JSONB` 字段（方案 Y），旧 `content TEXT` 字段保留
- **更新 `repository.CreateReport()`**：同时写入 `content`（纯文字兜底）和 `content_structured`（结构化数据）
- **更新 `model.AIReport`**：新增 `ContentStructured` 字段映射
- **前端结果页**：新增精简/专业切换按钮，按字段渲染不同模式的报告内容

## Capabilities

### New Capabilities

- `structured-ai-report`: AI 一次性生成包含命局推理总览、每章精简摘要与详细分析的结构化 JSON 报告，支持前端精简/专业双模式展示
- `report-mode-switcher`: 前端报告区域新增「精简 / 专业」模式切换 UI，精简模式展示摘要，专业模式展示完整推理链条与详细解读

### Modified Capabilities

- `ai-report`：报告输出格式从纯文字扩展为结构化 JSON（含 `analysis.logic`、`analysis.summary`、`chapters[].brief`、`chapters[].detail`），同时保持旧字段兼容

## Impact

- **后端**：`backend/internal/service/report_service.go`（Prompt + 解析重写）、`backend/internal/model/model.go`（新字段）、`backend/internal/repository/repository.go`（写入逻辑）、`backend/pkg/database/database.go`（DDL Migration）
- **前端**：`frontend/src/pages/ResultPage.tsx`（报告渲染区 + 模式切换 UI）、`frontend/src/lib/api.ts`（接口返回字段适配）
- **API**：`POST /api/bazi/report` 响应体中 `report` 对象新增 `content_structured` 字段，`content` 字段仍保留（向下兼容）
- **数据库**：需执行一次 `ALTER TABLE ai_reports ADD COLUMN content_structured JSONB`
