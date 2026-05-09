## ADDED Requirements

### Requirement: yongshen 输出契约扩展

`Calculate()` 排盘入口的返回结构 `BaziResult` SHALL 在不破坏现有 `Yongshen` / `Jishen` 字符串字段的前提下，新增 `YongshenStatus`、`YongshenGans`、`JishenGans`、`YongshenMissing` 四个字段，用于反映 yongshen 推算来源（调候命中、调候缺位 fallback 至扶抑、字典缺失）与具体调候用神天干。

#### Scenario: 排盘返回 yongshen 状态字段

- **WHEN** 调用 `Calculate(birthYear, birthMonth, birthDay, birthHour, gender, ...)` 完成排盘
- **THEN** 返回的 `BaziResult` 同时包含旧字段 `Yongshen`、`Jishen` 与新字段 `YongshenStatus`、`YongshenGans`、`JishenGans`、`YongshenMissing`
- **AND** `YongshenStatus` 取值为 `"tiaohou_hit"` / `"tiaohou_miss_fallback_fuyi"` / `"tiaohou_dict_missing"` 之一
- **AND** 旧的下游消费者（如 `getYongshenBaseline`、`caiIsJi`）继续以 `strings.Contains(natal.Yongshen, ...)` 形式访问 `Yongshen` 字段且行为不变

#### Scenario: 历史命盘字段缺失向后兼容

- **WHEN** 后端读取一个 yongshen_status 字段缺失的历史 `bazi_charts` 记录
- **THEN** 视该字段为零值（空字符串），不影响其他字段渲染与 JSON 序列化
- **AND** 不强制对历史数据回填新字段

### Requirement: yongshen 主流程不再依赖月令权重短路

`Calculate()` SHALL 调用统一的 yongshen 推算入口（调候优先、扶抑 fallback），不再在月令为"三冬"且火 = 0 或月令为"三夏"且水 = 0 的极端场景下使用硬编码"火木 / 水金"短路。

#### Scenario: 三冬无火命盘走调候字典

- **WHEN** 排盘日干为"甲"、月支为"子"、原局五行火 = 0、藏干中无丙也无庚
- **THEN** yongshen 推算路径首先查询调候字典 `LookupTiaohouYongshen("甲","子")`，得到字典定义的调候天干列表
- **AND** 因调候用神天干在原局缺位，`YongshenStatus` 为 `"tiaohou_miss_fallback_fuyi"`
- **AND** 最终 yongshen/jishen 由扶抑层 `calcWeightedYongshen` 提供，与原"三冬急用 → 火木"硬编码结果可能不同

#### Scenario: 三夏无水但藏干含癸

- **WHEN** 排盘日干为"甲"、月支为"午"、原局五行水 = 0，但日支或时支地支藏干含"癸"
- **THEN** `YongshenStatus` 为 `"tiaohou_hit"`
- **AND** `YongshenGans` 包含"癸"
- **AND** `Yongshen` 字符串包含"水"
- **AND** 不再走"水金 / 火木土"的硬编码短路
