## Why

当前命理解读的 PDF 导出已经能展示命盘与 AI 解读，但“命格”只停留在结果页顶部 badge 和弹窗中，没有进入最终命书内容。既然系统已经能稳定产出 `ming_ge` 与 `ming_ge_desc`，就应该让 PDF 直接回答“你是什么格”以及“这个格通常代表什么”，提升命书的完整度与可读性。

## What Changes

- 在 PDF 打印布局中新增独立的“命格解读”区块，放在“命理解读”页顶部、`命局分析总览` 之前
- `PrintLayout` 接收并展示 `ming_ge` 与 `ming_ge_desc`
- 命格解读区块采用轻量三段式信息结构：主格、格义、可选的本局落点短句
- 本次不扩 AI 报告结构，不新增后端字段，不修改现有 `analysis.logic` 生成协议
- 页面导出链路继续沿用现有前端打印 / PDF 方案，仅增强打印内容编排

## Capabilities

### New Capabilities
- `bazi-report-pdf`: 规范八字结果页导出 PDF 时的命格解读内容与版面要求

### Modified Capabilities
- None

## Impact

- 前端：`frontend/src/components/PrintLayout.tsx`、`frontend/src/pages/ResultPage.tsx`，以及相关打印样式
- 数据来源：复用现有 `result.ming_ge`、`result.ming_ge_desc` 与 `report.content_structured.analysis.logic`
- 无后端 API 变更，无数据库变更，无新增依赖
