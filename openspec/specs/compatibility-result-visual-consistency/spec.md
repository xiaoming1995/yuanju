# compatibility-result-visual-consistency Specification

## Purpose
TBD - created by archiving change compatibility-result-style-token-convergence. Update Purpose after archive.
## Requirements
### Requirement: 顶级模块按内容性质区分裸 SECTION 与卡片
合盘结果页的顶级模块 SHALL 按内容性质呈现：核心叙事段（双方基础盘、性格画像与差异、是否合、深度分析）MUST 呈现为裸 SECTION（头部 + 内容，无卡片边框）；附属/交互层（命理证据/命盘细节抽屉、AI 深度解读）MAY 呈现为卡片以标示其可选/二级性质。基准三段以外的核心叙事段 MUST NOT 保留卡片边框（如 `border-left`）。

#### Scenario: 性格画像与基准段视觉同级
- **WHEN** 用户浏览到「性格画像与差异」SECTION
- **THEN** 它呈现为与「双方基础盘 / 是否合 / 深度分析」一致的裸 SECTION，不带卡片边框或卡片背景
- **AND** 其头部为标准的 `compat-section-kicker`（SECTION 02）+ `compat-section-title`

#### Scenario: AI 深度解读保留卡片
- **WHEN** 用户浏览到页面末尾的「AI 深度解读」
- **THEN** 它呈现为卡片（与命理证据/命盘细节同属附属/交互层），与核心叙事裸段在视觉上区分

### Requirement: 顶级模块统一采用 section 级间距 token
每个顶级模块 SHALL 通过既有 section 级 token 确定其与相邻模块的间距与页面左右边距：模块间距用 `--section-gap-mobile/desktop`，页面左右留白用 `--section-padding-mobile/desktop`，并设置 `scroll-margin-top`（裸段用 padding、卡片用 margin 留出左右边距，与既有 `EvidenceDrawer` 一致）。顶级模块 MUST NOT 依赖内层元素的 margin 塌缩来形成模块间距。

#### Scenario: 性格 SECTION 接入间距节奏
- **WHEN** 渲染「性格画像与差异」SECTION
- **THEN** 其与上方「双方基础盘」、下方「是否合」之间的间距由 `--section-gap-*` 决定，与其它基准段一致
- **AND** 其左右边距由 `--section-padding-*` 决定

#### Scenario: AI 深度解读与命理证据之间有正常间距
- **WHEN** 渲染页面末尾的「AI 深度解读」
- **THEN** 它与上方「命理证据/命盘细节」之间存在由 `--section-gap-*` 决定的间距，不再紧贴

#### Scenario: 顶部 verdict 摘要条对齐内容列
- **WHEN** 渲染顶部 verdict 摘要条（`CompatibilityStickyHeader`）
- **THEN** 其面板（背景/底边线）左右缩进通过 `margin: 0 var(--section-padding-*)` 实现，与下方 SECTION 内容列左右边缘对齐，MUST NOT 通过横向 `padding` 形成比内容列更宽的左右探出
- **AND** 条内文字（双方姓名）的左边缘与 SECTION 标题/卡片内容的左边缘对齐

#### Scenario: 顶部导出按钮组对齐内容列
- **WHEN** 渲染顶部导出按钮组（`.compat-export-actions`：分享图片 / 导出 PDF）
- **THEN** 其横向 inset 采用 `var(--section-padding-*)`，使右对齐按钮的右边缘与下方 SECTION 内容列右边缘对齐，MUST NOT 比内容列右探出一个 `--section-padding`

### Requirement: 全页仅存在一套字号 token 系统
合盘结果页 SHALL 仅使用 `--fs-*` 字号 token 系统；MUST NOT 残留第二套硬编码字号的标题系统。处于 SECTION 内部的子标题 MUST 使用子标题级字号（`--fs-subsection-title`），MUST NOT 大于其所属 SECTION 的标题字号。

#### Scenario: 速览子标题不大于父 SECTION 标题
- **WHEN** 渲染「是否合」SECTION 内的「关系速览」子块（legacy 评分路径）
- **THEN** 「关系速览」标题字号为 `--fs-subsection-title`，不大于「是否合」的 SECTION 标题字号

#### Scenario: 旧字号类不再被引用
- **WHEN** 在 `frontend/src` 全仓检索 `.compatibility-section-title` / `.compatibility-section-header` / `.compatibility-section-desc`
- **THEN** 无任何组件引用这些旧类
- **AND** 这些孤儿类已从 `CompatibilityResultPage.css` 删除

### Requirement: 收敛不改变内容、结构与数据来源
本次视觉收敛 SHALL 仅调整 CSS 与必要的顶层 className，MUST NOT 改变任一模块的内部内容、DOM 结构、组件 props、模块顺序、评分算法或 AI 报告生成行为。

#### Scenario: 收敛前后内容一致
- **WHEN** 对同一条合盘 reading 比较收敛前后
- **THEN** 各模块展示的内容、顺序与数据来源完全一致，仅间距/边框/字号呈现不同

