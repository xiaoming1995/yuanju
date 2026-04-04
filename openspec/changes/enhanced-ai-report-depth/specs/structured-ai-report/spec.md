## ADDED Requirements

### Requirement: AI 生成结构化命理报告
系统 SHALL 调用 LLM 一次性生成包含命局推理总览和五章节精简/详细双版本的结构化 JSON 报告。

#### Scenario: 成功生成结构化报告
- **WHEN** 用户触发生成 AI 报告
- **THEN** 系统在 ai_reports 表写入 content_structured JSONB，包含 analysis.logic、analysis.summary 和 chapters 数组（每章含 title/brief/detail）

#### Scenario: 推理总览包含双层语言
- **WHEN** AI 报告的 analysis.logic 字段被渲染
- **THEN** 内容中每个推理点 SHALL 先写专业术语版，再紧跟通俗白话解释

#### Scenario: 每章节具有双版本内容
- **WHEN** AI 生成五章节内容
- **THEN** 每章 brief 字段约 100 字精简摘要，detail 字段约 350 字含推理依据的详细版本

#### Scenario: JSON 解析失败时降级
- **WHEN** LLM 返回内容无法解析为有效结构化 JSON
- **THEN** 系统 SHALL 将原始文本存入 content 字段，content_structured 置 NULL，报告仍可正常展示

### Requirement: 数据库存储结构化报告
`ai_reports` 表 SHALL 新增 `content_structured JSONB` 可空字段，与旧 `content TEXT` 字段共存。

#### Scenario: 新报告双字段写入
- **WHEN** 成功解析结构化 JSON 报告
- **THEN** content_structured 写入完整 JSON，content 写入各章 brief 拼接的纯文字（兜底用）

#### Scenario: 旧报告字段兼容
- **WHEN** 读取 content_structured 为 NULL 的历史报告
- **THEN** 系统 SHALL 使用 content 字段的纯文字内容正常渲染，不报错
