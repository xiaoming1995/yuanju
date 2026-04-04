## Why

当前 AI 报告的 Prompt 仅传入原始干支和五行统计，要求 LLM 自行推算十神、判断日主强弱、确定用神——这与"算法精算 + AI 解读"的核心理念相悖。引擎中已精确算出的十神、十二长生、神煞、大运等数据被白白浪费，LLM 的自行推算还可能与算法结果不一致，造成数据矛盾。

## What Changes

- **Prompt 重构**：将所有引擎已算出的关键数据（十神、十二长生、神煞、完整10步大运、旬空）注入 Prompt，LLM 的职责从「排盘+解读」收窄为「基于精算数据解读」。
- **用神传递方式调整**：引擎的初步用神/忌神作为参考提示传入，LLM 可基于更完整的十神数据确认或微调后输出最终结论。
- **CoT 步骤重构**：原「在心中推算日主强弱」改为「整合引擎精算数据，进行专业综合判断后再写报告」。
- **报告新增章节**：在原四章基础上新增「大运走势」章节（变为五章），让 LLM 基于完整大运数据解读人生各阶段走势。
- **修订 `bazi-ai-reasoning` 规格**：原规格要求「代码不得预推用神」需修订为「引擎可推算初步用神作为 LLM 参考，最终用神由 LLM 综合推理后输出」。

## Capabilities

### New Capabilities

- `ai-engine-grounded-prompt`：新的 Prompt 构建能力——将引擎精算数据（十神/长生/神煞/大运）结构化注入 AI 解读请求，实现「算法排盘、AI 解读」双层分工。

### Modified Capabilities

- `bazi-ai-reasoning`：修订用神推断策略——引擎可提供初步用神作为参考，LLM 综合十神等数据后确认或微调，不再要求 LLM 从零推算排盘。

## Impact

- **后端**：`internal/service/report_service.go` — `buildBaziPrompt()` 函数重构，接收更丰富的 `BaziResult` 字段。
- **数据**：`BaziResult` 中已有的 `ShiShen`、`DiShi`、`ShenSha`、`XunKong`、`Dayun`、`Yongshen`/`Jishen` 字段全部启用，无需新增后端字段。
- **AI 调用 token 量**：Prompt 长度将显著增加（大运10步 + 神煞），需关注各 Provider 的 token 上限（当前 `max_tokens: 2000` 应对应调整）。
- **不涉及数据库变更、前端变更或 API 接口变更**。
