## Why

合盘结果页当前的模块顺序是「基础盘 → 是否合(分数) → 深度分析(性格/风险/策略/下一步/AI) → 依据」。性格画像被埋在「深度分析」容器的第一个折叠子块里，要先翻过分数才看到。我们希望结果页按**叙事/好奇心驱动**的弧线呈现——先认识两人是什么样的人，再揭晓合不合，再讲怎么相处，最后想深挖——让用户的阅读顺序更自然。

## What Changes

- 把「性格画像与差异」整块从「深度分析」中拎出，**升为独立的顶级 SECTION**，位置排在「基础盘」与「是否合」之间。
- 对结果页 SECTION 重新编号：`01 双方基础盘 / 02 性格画像与差异 / 03 是否合 / 04 深度分析`。
- 「深度分析(04)」内部子块顺序调整为：**关系经营策略 → 阶段风险与时段 → 下一步/避免 → AI 深度解读**（策略提到第 1）。
- 「关键依据」抽屉保持在最后；隐藏的打印版 `CompatibilityPrintLayout` 不在本次范围。
- 若顶部存在锚点导航，同步更新其顺序/编号以匹配新结构。
- **不改动**：StickyHeader（仍吸顶显示总分+结论）、合盘评分算法、AI 报告生成、各模块内部内容与数据来源。

## Capabilities

### New Capabilities
- `compatibility-result-section-ordering`: 覆盖合盘结果页顶级 SECTION 的顺序与编号、性格画像作为独立 SECTION 的位置，以及「深度分析」内部子块的排列顺序。

### Modified Capabilities
- None.

## Impact

- 前端结果页：`frontend/src/pages/CompatibilityResultPage.tsx`（调整 SECTION 渲染顺序、性格块的渲染位置）
- 前端组件：`frontend/src/components/compatibility/SectionDeepAnalysis.tsx`（移除 PersonalityFit，调整内部子块顺序）、性格块从子块升级为 SECTION 的承载方式（新增轻量 section 包装或在页面层渲染 `PersonalityFit`）
- 各 SECTION 的 kicker 编号文案（`SECTION 01/02/03/04`）
- 可能涉及：顶部锚点导航、相关静态 UX 测试 `frontend/tests/`
- 无后端、API、数据库或评分模型改动
