## ADDED Requirements

### Requirement: Pattern Identification via AI Reasoning
系统 SHALL 在 AI 报告生成的 CoT 推理步骤中要求 AI 明确推断月令格局名称，并将格局结论融入命局分析总览。

#### Scenario: Pattern Name in Analysis Logic
- **WHEN** AI 生成报告
- **THEN** `analysis.logic` 字段中 SHALL 包含月令格局名称的推断结论（如「此命为正印格」）

#### Scenario: Pattern Informs Yongshen Determination
- **WHEN** AI 推断用神/忌神
- **THEN** 用神推断 SHALL 综合引擎调候用神参考 + 月令格局用神两个维度，而非仅依赖五行统计

### Requirement: Chapter-Level Data Anchoring
系统 SHALL 在每个报告章节的生成指令中明确指定应参考的命盘数据字段，以确保章节分析有具体命盘依据。

#### Scenario: Relationship Chapter Uses Peach Blossom Stars
- **WHEN** AI 生成「感情运势」章节
- **THEN** 分析 SHALL 显式考察桃花、红鸾等感情类神煞的存在情况，以及日支星运状态

#### Scenario: Career Chapter Uses Official Stars
- **WHEN** AI 生成「事业财运」章节
- **THEN** 分析 SHALL 显式考察官杀/食伤天干透出情况，及文昌/天乙贵人神煞

### Requirement: Dayun Current Position Identification
系统 SHALL 在 Prompt 中注入当前公历年份，要求 AI 基于起运年份和大运序列推算用户当前所处大运步次，并作为重点解读对象。

#### Scenario: Current Dayun Identified in Report
- **WHEN** AI 生成「大运走势」章节
- **THEN** 报告 SHALL 明确指出用户当前（[当前年份]）处于第几步大运，并重点解读该步及下一步大运的运势影响

## MODIFIED Requirements

### Requirement: Analytical LLM Prompts
后端系统 SHALL 向 LLM 提供完整的精算命盘数据（含调候用神），并在 System message 中定义统一的现代解读风格（通俗直接、结论先行、术语作精准点缀）；CoT Prompt 新增格局推断与调候整合步骤；MaxTokens 不低于 6000，Temperature 不高于 0.8。

#### Scenario: Modern Style Applied Consistently
- **WHEN** 后端向 LLM 发送请求
- **THEN** System message 中 SHALL 明确指定「现代解读风格」，User message 中不得出现与 System 相矛盾的风格指令

#### Scenario: Pattern and Tiaohou in CoT
- **WHEN** 后端构建 Prompt
- **THEN** CoT 步骤 SHALL 包含：①月令格局推断子任务；②调候用神（来自引擎精算）与格局用神整合子任务

#### Scenario: Tiaohou Data Injected into Prompt
- **WHEN** 后端构建 Prompt 且 `BaziResult.Tiaohou` 非空
- **THEN** Prompt 中 SHALL 注入「===调候用神（穷通宝鉴）===」区块，包含调候用神文本
