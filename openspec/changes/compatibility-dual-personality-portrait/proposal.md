## Why

合盘结果页的「性格相处画像」目前没有真正的双方性格内容：`compatibilityPersonality.ts` 的 `participantPattern()` 给「我方」和「对方」喂的是同一份合盘四维分数，只换了个日主标签，所以两边描述同源、本质上不存在「各自性格」，更谈不上对比。用户因此感到「输出内容缺少了双方性格对比」。而每个参与者的完整命盘（十神、日主、五行旺衰、用神/忌神）其实已存在 `chart_snapshot` 中并下发到前端，只是未被用于刻画个人性格。

## What Changes

- 在前端基于每个人**各自的命盘结构**（日主五行 + 主导十神 + 旺衰）推导出**可解释的个人性格画像**，替换现有两边同源的假 `participantPattern` 实现。
- 每个参与者输出 **5 条维度画像**：表达/沟通方式、决策与节奏、亲密里的核心需求、情绪反应、压力下的样子；并附一句「日主 + 主导十神」定性。
- 基于两人命盘结构推导**差异对照**：「自然合的地方」与「容易冲突的地方」。
- 在 `PersonalityFit.tsx` 渲染 **A 画像 / B 画像 / 差异对照** 三块结构。
- 补全前端 `CompatibilityChartSnapshot` 类型，使其能读取每人完整 `BaziResult`（十神、用神/忌神等字段）。
- **范围边界**：一期只做前端确定性逻辑，不改后端、零数据库迁移。LLM 增强（后端 prompt + `personality_comparison` 结构化字段）列为二期，不在本次范围。

## Capabilities

### New Capabilities
- `compatibility-personality-portrait`: 覆盖合盘结果中基于双方各自命盘的确定性性格画像、5 维度刻画、日主+主导十神定性，以及两人「合/冲」差异对照的推导与展示规则。

### Modified Capabilities
- None.

## Impact

- 前端类型：`frontend/src/lib/api.ts`（补全 `CompatibilityChartSnapshot` 字段）
- 前端逻辑：`frontend/src/lib/compatibilityPersonality.ts`（新增经典映射引擎，重写 self/partner 画像与差异推导）
- 前端 UI：`frontend/src/components/compatibility/deep-analysis/PersonalityFit.tsx` 及其 CSS（渲染双画像 + 差异对照）
- 数据来源：`chart_snapshot` 中已有完整 `BaziResult`，无需后端、API、数据库或合盘评分改动。
- 测试：`frontend/tests/` 下补充映射引擎与画像输出的静态测试。
