## Why

当前大运好坏评级（金不换算法）只基于**大运地支**所属的三合方位（东南西北）来判断吉凶，忽视了大运天干对命局的直接影响力。传统命理中，十年大运以干支并重，天干透干影响力往往是地支的"门面"，完全忽视天干会导致评级过于粗糙，部分命盘的大运吉凶预判出现明显偏差。

## What Changes

- 扩展 `JinBuHuanResult` 结构体，新增 `zhi_level`（地支原始评级）、`gan_modifier`（天干修正方向：加成/减损/通关/中性）、`gan_desc`（天干修正详细描述）字段
- 扩展 `JinBuHuanRule` 结构体，新增 `earth_gan_eval`（戊己土干专属评价，类型 `*JBHEval`）字段，120 条规则逐步补充土运说明
- 修改 `CalcJinBuHuanDayun` 函数，新增 `dayunGan` 入参，实现天干五行 → 喜忌集合匹配 → Level 升降的完整修正逻辑
- 土干（戊/己）命中日期：优先查 `earth_gan_eval`；若为 `nil` 则降级到解法一（通关逻辑）兜底，确保向后兼容
- 阶段一：完成框架实现，120 条 `EarthGanEval` 数据初始均为 `nil`（通关兜底）；阶段二：逐步按日干分批补全

## Capabilities

### New Capabilities
- `dayun-gan-modifier`: 大运天干修正评级能力——根据大运天干五行与命盘金不换规则中的喜忌方向匹配，对地支评级进行升降调整，并在结果中提供独立的天干修正说明

### Modified Capabilities
- 无（新增能力，不修改任何现有规格行为）

## Impact

- 后端改动：`backend/pkg/bazi/jin_bu_huan_dict.go`（结构体扩展、120条 earth_gan_eval 字段）、`backend/pkg/bazi/engine.go`（CalcJinBuHuanDayun 调用处增加 gan 参数）
- API 响应：`/api/bazi/calculate` 和 `/api/bazi/history/:id` 中 `dayun[].jin_bu_huan` 对象会新增字段，向前兼容无破坏
- 前端展示：`frontend/src/components/DayunTimeline.tsx` 可选择性展示 `gan_modifier` 天干修正徽章
