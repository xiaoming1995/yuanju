## Context

后端 `BuildPersonalDay` 把双喜用标记为 `supportive`，单边喜用、另一边中性标记为 `undetermined`，并分别给出方向。首页文案原先只将 `supportive` 映射成积极语句，所以后者看起来与完全中性相同。

## Goals / Non-Goals

**Goals:** 让有单边正向依据的日期获得较轻的积极文案，并与双喜用、纯中性及谨慎状态保持区别。

**Non-Goals:** 不修改后端状态或四柱算法，不把一项喜用说成确定的「好运」，不增加新 UI 区块。

## Decisions

只在 `undetermined` 且日干或日支主气有一项 `favorable`、另一项没有 `adverse` 时显示「今天有点顺风」。先检查任一 `adverse`，防止冲突方向落入积极分支。后端相冲与调候分歧已优先产出 `caution`／`mixed`，不会进入此分支。双喜用 `supportive` 保留「今天节奏挺顺」。

## Risks / Trade-offs

- [轻正向被误读为完整好运] → 文案限定为「有点顺风」，行动建议仍是「从容迈一步」。
- [旧接口方向缺失] → 不满足正向条件，维持原平稳文案。

## Migration Plan

仅更新小程序文案映射，无数据迁移；回退映射即可恢复旧行为。
