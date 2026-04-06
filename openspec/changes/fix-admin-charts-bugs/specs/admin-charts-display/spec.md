## ADDED Requirements

### Requirement: 起盘明细列表每条命盘最多返回一行
`ListBaziCharts` 查询 SHALL 保证每个 `bazi_charts` 记录在结果中仅出现一次，即使该命盘有多条 `ai_reports`。

#### Scenario: 命盘有多条历史 AI 报告时不出现重复行
- **WHEN** 某命盘存在多条 `ai_reports` 记录（如用户多次生成报告）
- **THEN** 管理员起盘明细列表中该命盘仅显示一条，展示其中最新一条报告内容

#### Scenario: 命盘无 AI 报告时正常显示
- **WHEN** 命盘尚未生成任何 AI 报告
- **THEN** 该命盘仍出现在列表中，AI 相关字段为空

### Requirement: 起盘明细详情显示结构化 AI 报告
管理后台命盘详情面板 SHALL 优先展示 `content_structured` 结构化字段，降级时展示原始 `content` 文本摘要。

#### Scenario: 命盘有结构化报告时四章节显示
- **WHEN** 管理员展开一条命盘详情，该命盘最新 AI 报告包含 `content_structured`（含 `personality/career/romance/health` 字段）
- **THEN** 详情面板显示分章节的结构化内容，而非原始 JSON 字符串

#### Scenario: 命盘无结构化内容时降级到文本摘要
- **WHEN** 命盘最新报告 `content_structured` 为 null 但 `content`（纯文本 JSON）存在
- **THEN** 降级显示"此命盘 AI 报告为旧格式，无结构化内容"提示

#### Scenario: 命盘无 AI 报告
- **WHEN** 命盘尚未生成任何 AI 报告
- **THEN** 显示"此命盘尚未生成 AI 原局报告"占位提示

### Requirement: 流年批断记录 loading 状态按命盘隔离
流年批断记录的加载状态 SHALL 按 `chartId` 独立维护，互不干扰。

#### Scenario: 快速切换展开不同命盘时 loading 正确显示
- **WHEN** 管理员连续快速点击展开 A 命盘、B 命盘
- **THEN** A 命盘和 B 命盘各自显示独立的 loading 状态，互不覆盖
