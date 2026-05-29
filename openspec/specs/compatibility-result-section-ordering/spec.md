# compatibility-result-section-ordering Specification

## Purpose
TBD - created by archiving change compatibility-result-narrative-reorder. Update Purpose after archive.
## Requirements
### Requirement: 结果页顶级 SECTION 顺序与编号
合盘结果页 SHALL 按以下顺序渲染顶级 SECTION，并以连续编号标注：`01 双方基础盘`、`02 性格画像与差异`、`03 是否合`、`04 深度分析`；「关键依据」抽屉 SHALL 仍位于最后。

#### Scenario: 主体模块按叙事顺序渲染
- **WHEN** 用户打开合盘结果页
- **THEN** 主体内容依次为：双方基础盘 → 性格画像与差异 → 是否合（分数）→ 深度分析 → 关键依据
- **AND** 各顶级 SECTION 的编号为连续的 01、02、03、04，无重复、无跳号

### Requirement: 性格画像作为独立 SECTION 排在分数之前
「性格画像与差异」SHALL 作为独立顶级 SECTION 呈现，位置 MUST 在「双方基础盘」之后、「是否合（分数）」之前，且 MUST NOT 再嵌套在「深度分析」容器内。

#### Scenario: 性格画像先于分数出现
- **WHEN** 用户从上往下浏览结果页
- **THEN** 「性格画像与差异」整块在「是否合（分数）」之前出现
- **AND** 该 SECTION 带有与其它 SECTION 一致的头部（编号 + 标题），不出现标题重复

#### Scenario: 深度分析不再包含性格画像
- **WHEN** 渲染「深度分析」SECTION
- **THEN** 其内部 MUST NOT 包含「性格画像与差异」子块

### Requirement: 深度分析内部子块顺序
「深度分析」SECTION 内部子块 SHALL 按以下顺序渲染：关系经营策略 → 阶段风险与时段 → 下一步/避免。关系经营策略为条件渲染，缺失时其余子块 SHALL 顺延，不留空位。AI 深度解读 MUST NOT 再包含在「深度分析」内（见下条）。

#### Scenario: 深度分析子块按新顺序
- **WHEN** 渲染「深度分析」SECTION 且关系经营策略存在
- **THEN** 子块顺序为：关系经营策略、阶段风险与时段、下一步/避免

#### Scenario: 关系经营策略缺失时顺延
- **WHEN** 关系经营策略数据不存在
- **THEN** 「深度分析」从「阶段风险与时段」开始渲染，不出现空白占位

### Requirement: AI 深度解读作为页面最后一环
AI 深度解读 SHALL 从「深度分析」中拎出，作为独立模块渲染在页面**最后**；命理证据/命盘细节（EvidenceDrawer）SHALL 排在其之前。

#### Scenario: AI 深度解读在命理证据之后、页面末尾
- **WHEN** 用户浏览到结果页底部
- **THEN** 顺序为：深度分析 → 命理证据/命盘细节 → AI 深度解读
- **AND** AI 深度解读是页面正文的最后一个模块

### Requirement: 重排不改变模块内容与数据来源
本次重排 SHALL 仅调整模块顺序与编号，MUST NOT 改变任一模块的内部内容、数据来源、评分算法或 AI 报告生成行为。

#### Scenario: 模块内容保持不变
- **WHEN** 重排前后对比同一条合盘 reading
- **THEN** 各模块展示的内容与数据来源一致，仅出现位置/编号不同

