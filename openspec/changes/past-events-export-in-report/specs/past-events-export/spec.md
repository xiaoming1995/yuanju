## ADDED Requirements

### Requirement: 过往事件页可导出已生成内容
过往事件推算页 SHALL provide a PDF export action that exports only the current chart's generated past-event content available on the page.

#### Scenario: 有已生成大运批语时导出
- **WHEN** the user opens a past-events page with one or more generated dayun summaries and clicks the export action
- **THEN** the system produces a PDF containing those generated dayun summaries and their generated yearly narratives

#### Scenario: 存在未生成或生成中内容
- **WHEN** the user exports from the past-events page while some dayun segments are not generated, loading, interrupted, or folded pending generation
- **THEN** the exported PDF MUST include generated segments only and MUST NOT include loading placeholders as narrative正文

#### Scenario: 用户选择包含命理依据
- **WHEN** the user exports from the past-events page with an include-evidence option enabled
- **THEN** the exported PDF SHALL include available yearly signals and evidence summaries for exported years

#### Scenario: 用户不选择包含命理依据
- **WHEN** the user exports from the past-events page without the include-evidence option
- **THEN** the exported PDF SHALL omit detailed evidence summaries and keep only generated summaries, yearly narratives, and essential labels

### Requirement: 基本盘报告导出可拼接已生成过往内容
The main bazi report export SHALL append a past-events section only when the chart already has generated past-event dayun summary cache.

#### Scenario: 命盘已有过往事件缓存
- **WHEN** the user exports the main bazi report for a chart that has one or more cached past-event dayun summaries with yearly narratives
- **THEN** the PDF SHALL append a section titled "过往年运回看" after the main report content

#### Scenario: 命盘没有过往事件缓存
- **WHEN** the user exports the main bazi report for a chart without generated past-event dayun summary cache
- **THEN** the PDF MUST NOT show an empty past-events section or prompt the user to generate past events

#### Scenario: 只有部分大运段已生成
- **WHEN** the user exports the main bazi report for a chart where only some dayun segments have generated past-event cache
- **THEN** the appended past-events section SHALL include only the cached generated segments

#### Scenario: 主报告导出保持精简
- **WHEN** the main bazi report PDF includes the appended past-events section
- **THEN** the section SHALL include dayun themes, dayun summaries, and generated yearly narratives, and MUST omit detailed evidence summaries by default

### Requirement: 基本盘分享图可拼接已生成过往内容
The main bazi share image SHALL append concise past-events content only when the chart already has generated past-event dayun summary cache.

#### Scenario: 命盘已有过往事件缓存并保存分享图
- **WHEN** the user saves the main bazi share image for a chart that has one or more cached past-event dayun summaries with yearly narratives
- **THEN** the share image SHALL append a concise section titled "过往年运回看"

#### Scenario: 命盘没有过往事件缓存并保存分享图
- **WHEN** the user saves the main bazi share image for a chart without generated past-event dayun summary cache
- **THEN** the share image MUST NOT show an empty past-events section or prompt the user to generate past events

#### Scenario: 分享图保持精简
- **WHEN** the main bazi share image includes the appended past-events section
- **THEN** the section SHALL include dayun themes, dayun summaries, and generated yearly narratives, and MUST omit detailed evidence summaries and yearly signal lists

### Requirement: 导出动作不得触发过往事件 AI 生成
Past-events export flows SHALL be read-only with respect to AI generation and MUST NOT start, retry, or continue past-events AI generation.

#### Scenario: 主报告导出检查过往内容
- **WHEN** the user exports the main bazi report
- **THEN** the system MAY read existing past-event cache but MUST NOT call the past-events streaming generation endpoint

#### Scenario: 分享图保存检查过往内容
- **WHEN** the user saves the main bazi share image
- **THEN** the system MAY read existing past-event cache but MUST NOT call the past-events streaming generation endpoint

#### Scenario: 过往事件页导出当前页面状态
- **WHEN** the user exports from the past-events page
- **THEN** the system MUST use already available generated data or read-only cached data and MUST NOT generate missing dayun segments

#### Scenario: 只读接口被调用
- **WHEN** a frontend export flow calls a backend endpoint for past-events export data
- **THEN** the endpoint SHALL return existing cached generated segments only and MUST NOT invoke an LLM provider

### Requirement: 过往导出版式适配 PDF
The exported past-events content SHALL use a print/PDF-specific layout that is separate from the interactive dark timeline UI.

#### Scenario: 打印版隐藏交互元素
- **WHEN** the past-events content is exported to PDF
- **THEN** the PDF MUST NOT include interactive controls such as retry buttons, expand/collapse buttons, navigation buttons, or loading spinners

#### Scenario: 导出版式包含必要上下文
- **WHEN** the generated past-events content appears in a PDF
- **THEN** the PDF SHALL include chart identity context, dayun age/year ranges, generated content, and the standard reference disclaimer
