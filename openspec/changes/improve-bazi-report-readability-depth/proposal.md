## Why

当前命理解读报告虽然已有精简/专业双模式，但默认展示精简内容，专业版字数也缺少稳定约束，导致用户第一眼觉得解读偏短、不够深入。报告正文固定为较小字号，术语缺少持续白话转译，也会降低普通用户的阅读友好度。

## What Changes

- 将命理解读报告的默认阅读体验调整为更完整的详解视图，避免用户默认只看到短版结论。
- 优化报告正文排版，提高正文字号、行高和段落阅读舒适度。
- 强化 AI 报告 Prompt 的内容深度要求：专业章节需有稳定字数下限，并按“结论、命理依据、现实表现、建议”组织。
- 增加术语白话化约束：关键命理术语出现后必须自然解释其现实含义，避免普通用户读不懂。
- 保持现有结构化报告字段与历史报告降级渲染兼容，不新增数据库字段，不改变 API 响应结构。

## Capabilities

### New Capabilities

- `bazi-report-readability`: 命理解读报告的前端阅读体验，包括默认展示模式、正文排版、精简/专业切换文案和普通用户可读性。

### Modified Capabilities

- `bazi-ai-reasoning`: AI 命理解读生成要求增加稳定字数、章节结构和术语白话化约束。

## Impact

- **后端**：`backend/internal/service/report_service.go` 中 `buildBaziPrompt()` 的报告生成约束。
- **前端**：`frontend/src/pages/ResultPage.tsx` 的默认报告模式与切换文案，`frontend/src/pages/ResultPage.css` 的报告正文排版样式。
- **测试**：后端 Prompt 单元测试或生成内容结构测试；前端报告渲染测试或手动验证。
- **兼容性**：不改变数据库 schema；不改变 `content_structured` 结构；旧报告继续按现有降级路径展示。
