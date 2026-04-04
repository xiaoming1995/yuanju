## ADDED Requirements

### Requirement: 前端报告精简/专业模式切换
前端报告区域 SHALL 提供「精简」/「专业」切换按钮，读取 content_structured 不同字段渲染对应内容。

#### Scenario: 默认展示精简模式
- **WHEN** 报告页面加载且 content_structured 存在
- **THEN** 默认展示 analysis.summary + 各章 brief，切换按钮默认高亮「精简」

#### Scenario: 切换到专业模式
- **WHEN** 用户点击「专业」切换按钮
- **THEN** 展示 analysis.logic（完整推理总览）+ 各章 detail（含推理依据）

#### Scenario: 旧报告无切换按钮
- **WHEN** content_structured 为空（历史报告或解析失败）
- **THEN** 不显示切换按钮，直接展示 content 纯文字，渲染行为与现有逻辑一致

#### Scenario: 精简/专业切换状态持久
- **WHEN** 用户切换模式后滚动页面再切回
- **THEN** 模式选择 SHALL 保持在当前 Session 内有效（无需持久化到 DB）
