## ADDED Requirements

### Requirement: Engine-Grounded Prompt Construction
后端系统 SHALL 在构建八字 AI 报告的 Prompt 时，将引擎精算的以下数据全部注入：天干十神、地支主气十神、日主十二长生（地势）、四柱神煞、旬空（空亡）、完整大运序列（共10步，含干支十神）。

#### Scenario: Report generation with engine data
- **WHEN** `buildBaziPrompt(result)` 被调用
- **THEN** Prompt 中必须包含四柱各天干的十神标注（如：年干甲=正官）
- **THEN** Prompt 中必须包含四柱各地支的主气十神标注
- **THEN** Prompt 中必须包含日主对各柱地支的十二长生状态
- **THEN** Prompt 中必须包含四柱神煞（若为空则显示「无」）
- **THEN** Prompt 中必须包含完整10步大运（干、支、干十神、支十神、起运年龄和年份）

### Requirement: Engine Yongshen as Reference Hint
后端系统 SHALL 将引擎通过算法初步推算的用神/忌神作为参考提示注入 Prompt，明确标注其为「参考」，并允许 LLM 基于完整十神数据确认或微调后输出最终结论。

#### Scenario: Reference yongshen injection
- **WHEN** `result.Yongshen` 非空时
- **THEN** Prompt 中以「引擎初步推算：用神=[X]，忌神=[Y]（供参考，请综合十神数据确认或微调）」格式注入

#### Scenario: Empty yongshen handling
- **WHEN** `result.Yongshen` 为空字符串
- **THEN** Prompt 中省略此段，不影响 LLM 正常推断

### Requirement: Five-Chapter Report Structure
LLM 生成的解读报告 SHALL 包含五个章节：【性格特质】【感情运势】【事业财运】【健康提示】【大运走势】。

#### Scenario: Report contains five chapters
- **WHEN** LLM 成功生成报告
- **THEN** 返回的 `report` 字段文本包含五个章节标题
- **THEN** 【大运走势】章节引用起运年龄和当前大运信息，解读人生各阶段整体运势

### Requirement: Increased Max Tokens for Richer Reports
AI 调用时的 `max_tokens` 参数 SHALL 设为 3500，以容纳五章报告和更丰富的精算数据输出。

#### Scenario: Token limit applied
- **WHEN** `callOpenAICompatible` 被调用
- **THEN** 请求体中 `max_tokens` 值为 3500
