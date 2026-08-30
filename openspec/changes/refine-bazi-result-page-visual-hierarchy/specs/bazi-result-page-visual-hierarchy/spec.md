## ADDED Requirements

### Requirement: Result page presents a clear primary reading hierarchy
八字结果页 SHALL 将四柱身份、核心扶抑结论、命盘总评、命盘座驾和当前路况作为摘要阅读链路，并将趋势、命格参考、判断依据、专业命盘和 AI 解读置于后续深入阅读区。

#### Scenario: User opens a result with complete overview data
- **WHEN** 用户进入含座驾和当前大运路况的结果页
- **THEN** 页面 SHALL 在深入阅读模块之前呈现总评、座驾与当前路况摘要
- **AND** 首屏 SHALL 不在四柱标题与总评之间呈现命格古人映照

#### Scenario: Result lacks vehicle or road data
- **WHEN** 用户进入缺少座驾或大运路况数据的结果页
- **THEN** 页面 SHALL 保持相同的结论优先顺序
- **AND** 页面 SHALL 不为缺失数据保留空白摘要面板

### Requirement: Summary presentation remains concise and visually consistent
结果页 SHALL 使用统一的标题层级、区块间距、弱分隔和低对比辅助标签呈现摘要内容。总览 SHALL 不通过连续的大卡片或嵌套卡片展示同一层级的信息。

#### Scenario: User scans the summary area on desktop
- **WHEN** 用户在桌面端查看结果页摘要区
- **THEN** 命盘座驾与当前路况 SHALL 顶部对齐并维持独立内容高度
- **AND** 总评、摘要网格和深入阅读区 SHALL 通过一致留白和分隔形成清晰层级

#### Scenario: User scans the summary area on a narrow viewport
- **WHEN** 用户在窄屏设备查看结果页
- **THEN** 摘要内容 SHALL 按总评、座驾、路况的连续单列顺序呈现
- **AND** 文本、标签和操作入口 SHALL 不重叠、不溢出容器

### Requirement: Supporting modules use progressive reading order
命格古人映照 SHALL 作为命格参考模块出现在摘要和趋势之后。简单模式与专业模式 SHALL 共享页面级阅读顺序；专业模式仅增加已有的专业数据和依据入口。

#### Scenario: User reads simple mode
- **WHEN** 用户以小白模式查看结果页
- **THEN** 页面 SHALL 先显示结论和当前阶段摘要
- **AND** 用户 SHALL 能继续进入趋势与判断依据阅读

#### Scenario: User switches to professional mode
- **WHEN** 用户切换为专业模式
- **THEN** 核心结论、座驾与路况的页面顺序 SHALL 保持不变
- **AND** 专业依据入口 SHALL 仅在对应数据可用时出现
