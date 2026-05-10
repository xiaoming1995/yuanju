## ADDED Requirements

### Requirement: 神煞重煞/轻煞分级
系统 SHALL 将神煞区分为重煞与轻煞两个级别，重煞列表为：羊刃、白虎、岁破、丧门、吊客、灾煞、劫煞、亡神。不在该列表中的神煞均视为轻煞。

#### Scenario: 重煞标记存在于白名单结构
- **WHEN** `shenshaWhitelist` 中定义神煞条目
- **THEN** 每条条目包含 `IsHeavy bool` 字段，重煞列表中的神煞 `IsHeavy=true`，其余为 `false`

### Requirement: 重煞凶信号压制轻煞吉信号
当同一流年存在重煞凶信号（Polarity=凶）时，系统 SHALL 在同年轻煞吉信号的 Evidence 末尾追加「（本年有重煞，此信号仅作参考）」，轻煞信号的 Polarity 保持不变。

#### Scenario: 重煞凶信号降级同年轻煞吉信号
- **WHEN** 同一流年中存在至少一条重煞凶信号，且该年还存在 `IsHeavy=false` 的神煞吉信号
- **THEN** 该轻煞吉信号的 Evidence 末尾追加「（本年有重煞，此信号仅作参考）」，Polarity 保持吉不变

#### Scenario: 无重煞凶信号时轻煞吉信号正常输出
- **WHEN** 同一流年中不存在重煞凶信号
- **THEN** 轻煞吉信号不受降级处理，正常输出

### Requirement: 重煞不影响 Layer 4 十神信号
神煞重煞的出现 SHALL NOT 影响 Layer 4（十神宏观，Source=柱位互动）的信号，Layer 4 信号按正常逻辑输出。

#### Scenario: 重煞仅压制神煞轻煞，不影响十神
- **WHEN** 同一流年同时存在重煞凶信号和 Layer 4 吉信号
- **THEN** Layer 4 吉信号不受重煞压制，按 Layer 0 vs Layer 4 交互规则（layer0-layer4-interaction）处理

### Requirement: 旬空对所有信号一律减半
旬空（空亡）对所有类型的信号（含重煞、轻煞、Layer 0、Layer 4）均适用减半标注，不因信号类型而豁免。

#### Scenario: 重煞信号遇旬空减半
- **WHEN** 某流年神煞（含重煞）所在宫位落入旬空
- **THEN** 该神煞信号的 Evidence 标注「（旬空，力度减半）」，不因重煞级别而跳过

#### Scenario: Layer 0 信号遇旬空减半
- **WHEN** 某流年 Layer 0 应期位置信号涉及旬空宫位
- **THEN** 该信号的 Evidence 标注「（旬空，力度减半）」
