## Why

当前系统已经在算法层产出了 `ming_ge` / `ming_ge_desc`，并在结果页顶部作为命格 Badge 展示；但原局 AI 报告的 `analysis.logic`（命局分析总览）仍通过 Prompt 要求 LLM 按 `kb_gejv` 自行再定一次格。这样会形成“双轨格局判断”：

- 顶部命格 Badge 使用算法 `MingGe`
- 命局分析总览使用 LLM 二次格局判断

两条链路一旦结论不一致，用户会看到“顶部是正官格，但总览写成偏印格/伤官格”的冲突，直接削弱平台的专业可信度。既然命盘数据中已经增加了定格结果，本轮应明确将其提升为 AI 解读中的**唯一格局真源**。

## What Changes

- 修改原局 AI 报告 Prompt，使其显式注入 `MingGe` / `MingGeDesc`
- 将格局模块从“重新定格”改为“解释系统主格是否成格、破格、有救、偏弱”
- 为 `analysis.logic` 增加主格优先的固定展开顺序：先写主格，再写格局状态，再写与调候/喜用神的关系
- 将 `kb_gejv` 从“判格规则库”收束为“解释型格局知识库”，避免与主 Prompt 形成冲突

## Capabilities

### Modified Capabilities

- `bazi-ai-reasoning`: AI 报告中的格局结论改为以系统定格结果为唯一真源，LLM 只负责解释主格的成立程度、调候关系、喜忌落点与行运表现，不再重新改判格名

## Impact

- **后端 Prompt 构造**：`backend/internal/service/report_service.go` 的 `buildBaziPrompt()` 需要注入系统定格结果，并重写格局相关指令
- **知识库模块**：`kb_gejv` 的默认内容与 Admin 配置语义需要从“重新定格”调整为“解释系统主格”
- **无 API 结构变更**：响应字段不新增不删减，仍沿用现有 `content_structured.analysis.logic`
- **无数据库 schema 变更**：`ming_ge` / `ming_ge_desc` 已存在于 `result_json` 快照中，本变更只改变其在 AI 报告中的使用方式
