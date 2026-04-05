## ADDED Requirements

### Requirement: 暴露出流年交脱细节
引擎在渲染排盘结果时，必须为包含大运交接边界的“交脱年”提供精确的天文交替月份、日期标识以及上一旬大运识别码，以便调用方（前端或AI模型）能明晰该年的运势切割断层。

#### Scenario: 大运交脱年份判定
- **WHEN** 核心引擎计算并组装某一个 `DayunItem`（10年跨度）内部的 10 个流年（`LiuNianItem`）列表，且正在处理该大运下首个开启的流年时
- **THEN** 系统应在这笔首年记录上将 `IsTransition` 标记为 `true`，并注入继承自全局 `StartSolar` 的确切 `TransMonth` 与 `TransDay`。同时须赋值 `PrevDayun` 代表上一大运干支（若首次起运可置空或做默认标识）。对于另外 9 个常规流年，系统应将其 `IsTransition` 置为 `false`。
