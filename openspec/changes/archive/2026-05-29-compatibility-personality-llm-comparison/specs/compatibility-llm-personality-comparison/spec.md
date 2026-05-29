## ADDED Requirements

### Requirement: AI 深度解读生成双方性格画像与差异
合盘 AI 深度解读 SHALL 基于双方完整命盘（含十神 / 命格 / 旺衰）由 LLM 生成「双方性格画像与差异」，作为结构化报告字段 `personality_comparison` 输出。该字段 SHALL 包含：每人 `headline` 与 5 维画像（表达沟通 / 决策节奏 / 亲密核心需求 / 情绪反应 / 压力下样子），以及双方 `fit_points`（自然合的地方）与 `clash_points`（容易冲突的地方）。LLM 的性格判断 MUST 基于输入命盘，MUST NOT 使用绝对断语或输出确定事件日期。

#### Scenario: 生成报告时产出性格画像字段
- **WHEN** 用户生成合盘 AI 深度解读
- **THEN** 结构化报告包含 `personality_comparison`，其中 `self` 与 `partner` 各含 headline 与 5 维画像
- **AND** 包含基于双方命盘的 `fit_points` 与 `clash_points`

#### Scenario: prompt 摘要含性格所需命盘数据
- **WHEN** 后端为某条 reading 构建合盘 prompt
- **THEN** 双方命盘摘要 SHALL 包含十神、命格与旺衰信息（不止四柱/五行/用神忌神）

### Requirement: 性格画像渲染于 AI 解读块内
「双方性格画像与差异」SHALL 仅渲染在页面底部的 AI 深度解读块内（读取 `personality_comparison`）。结果页 MUST NOT 再以独立顶级 SECTION 呈现确定性公式生成的性格画像。

#### Scenario: 生成后在 AI 解读内显示性格
- **WHEN** 已生成的报告含 `personality_comparison`
- **THEN** AI 深度解读块内渲染每人 5 维画像 + headline + 合点/冲突点

#### Scenario: 未生成报告时不显示性格
- **WHEN** 用户尚未生成 AI 深度解读
- **THEN** 页面不显示任何「双方性格画像与差异」内容
- **AND** AI 深度解读空态提示其内容将包含双方性格画像与差异

### Requirement: 移除前端确定性性格引擎
确定性性格画像/差异引擎 SHALL 从前端移除：`buildParticipantPortrait`、`buildPersonalityContrast`、`buildPersonalityFitSummary` 及其映射表与辅助函数 MUST 删除。仍被其它页面复用的函数（`getPersonalityMatchType`、`buildPersonalityConsultationPreview`、`getCompatibilityQuestionLabel`、`getCompatibilityStageLabel`）MUST 保留。删除后 MUST NOT 残留对已删函数/类型的引用或孤儿导入。

#### Scenario: 删除后无悬空引用
- **WHEN** 删除确定性引擎后运行 `tsc -b` 与 eslint
- **THEN** 无类型错误、无未使用导入、无对已删符号的引用

#### Scenario: 复用函数仍可用
- **WHEN** 合盘入口页与历史页渲染
- **THEN** 性格咨询预览与 matchType 标签仍正常工作（依赖的保留函数未被误删）

### Requirement: 旧报告安全降级
不含 `personality_comparison` 的历史报告 SHALL 安全降级：后端解析得到空值、前端不渲染性格子块，MUST NOT 报错或阻断报告其余内容渲染。

#### Scenario: 旧报告无性格字段
- **WHEN** 渲染一条在本变更前生成、不含 `personality_comparison` 的报告
- **THEN** 报告其余内容（summary / 维度 / 风险 / 建议）正常渲染
- **AND** 不显示性格子块、不抛错
