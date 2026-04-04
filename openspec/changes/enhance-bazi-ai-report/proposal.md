## Why

现有 AI 命理报告存在多个影响解读质量的系统性问题：System/User 角色指令矛盾导致输出风格不稳定；章节分析缺乏命盘数据锚点导致内容泛化；大运走势因缺少当前年份而无法精准定位；调候用神（《穷通宝鉴》）和月令格局（《子平真诠》）等核心命理维度完全缺失。这些问题使报告的专业性和精准度均低于行业水准，亟需系统性升级。

## What Changes

- **修复 Prompt 角色冲突**：统一 System message 与 User prompt 中的风格定义为「现代解读风格」（通俗直接，结论先行，术语作点缀）
- **注入当前年份**：在 Prompt 中注入当前公历年份，要求 AI 基于起运数据精确推算用户目前处于哪步大运，并据此重点解读
- **章节数据锚点**：为六个报告章节各添加命盘数据锚点指令，明确告知 AI 分析时应参考哪些精算字段（十神位置/神煞类型/地支星运等）
- **月令格局推断（子平真诠）**：在 Prompt CoT 步骤中要求 AI 基于月令地支藏干主气推断格局名称（正官格/七杀格/食神格等），并将格局用神融入命局分析总览
- **调候用神精算（穷通宝鉴）**：后端引擎新增调候查表（日主 × 出生月份 → 调候用神），计算结果作为独立字段注入 Prompt
- **调整生成参数**：MaxTokens 4500 → 6000；Temperature 1.0 → 0.75

## Capabilities

### New Capabilities
- `bazi-tiaohou-engine`: 基于《穷通宝鉴》实现的后端调候用神精算能力，通过日主天干 × 出生月令的查表逻辑，输出调候用神及说明，供 AI 报告参考引用

### Modified Capabilities
- `bazi-ai-reasoning`: 新增格局推断要求（月令格局识别与格局用神推演）、现代解读风格约束、章节数据锚点要求、大运走势时间定位要求；输出 token 上限和随机性参数同步调整

## Impact

- **后端**：`backend/pkg/bazi/engine.go`（新增 `Tiaohou` 字段）、`backend/pkg/bazi/tiaohou.go`（新增调候查表，约 150 行 Go 代码）、`backend/internal/service/report_service.go`（Prompt 全面重写 + 参数调整）
- **前端**：无需改动
- **数据库**：无需改动
- **API**：`BaziResult` 新增 `tiaohou` 字段（随 `/api/bazi/calculate` 响应返回，向下兼容）
