## ADDED Requirements

### Requirement: 大运总结块展示
系统 SHALL 在"过往事件推算"页面的每个大运分组顶部展示一个总结块，包含 AI 生成的主题关键词标签和叙述段落。

#### Scenario: 正常展示大运总结
- **WHEN** 过往事件推算加载完成，且 AI 响应包含 `dayun_summaries` 字段
- **THEN** 每个大运组标题下方展示主题标签（2-4 个）和叙述段（80-120 字）

#### Scenario: 旧缓存降级展示
- **WHEN** 加载的缓存数据中不含 `dayun_summaries` 字段（旧格式）
- **THEN** 大运组只显示干支标题，不渲染总结块，不报错

### Requirement: 大运总结 AI 生成
AI SHALL 在同一次调用中为每个过往大运生成总结，输出格式为 `dayun_summaries` 数组。

#### Scenario: AI 返回正确结构
- **WHEN** AI 完成推算
- **THEN** JSON 响应包含 `dayun_summaries: [{gan_zhi, themes, summary}]`，每个元素对应一段大运，`themes` 为 2-4 个短词数组，`summary` 不超过 120 字

#### Scenario: 大运数量与年份分组一致
- **WHEN** `dayun_summaries` 渲染
- **THEN** 每个 `gan_zhi` 值能在 `years[]` 的 `dayun_gan_zhi` 分组中找到对应条目
