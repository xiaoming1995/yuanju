## ADDED Requirements

### Requirement: 调候优先 yongshen 推算

系统 SHALL 在排盘时按"调候用神主导，扶抑用神 fallback"的优先级推算 yongshen 与 jishen，输出五行集合字符串与具体调候用神天干列表。

#### Scenario: 调候命中（透干）

- **WHEN** 排盘日干为"甲"、月支为"寅"、原局四柱中至少一个天干为"丙"
- **THEN** `BaziResult.YongshenStatus` 等于 `"tiaohou_hit"`
- **AND** `BaziResult.YongshenGans` 包含 `"丙"`
- **AND** `BaziResult.Yongshen` 字符串包含五行 "火"

#### Scenario: 调候命中（藏干）

- **WHEN** 排盘日干为"甲"、月支为"寅"、原局四个天干都不是"丙"，但月支寅藏干含"丙"
- **THEN** `BaziResult.YongshenStatus` 等于 `"tiaohou_hit"`
- **AND** `BaziResult.YongshenGans` 包含 `"丙"`
- **AND** `BaziResult.Yongshen` 字符串包含 "火"

#### Scenario: 调候字典给多个用神，原局部分命中

- **WHEN** 调候字典给定 yongshen 天干列表为 `["丙", "癸"]`、原局含"丙"、不含"癸"（透与藏均无）
- **THEN** `BaziResult.YongshenStatus` 等于 `"tiaohou_hit"`
- **AND** `BaziResult.YongshenGans` 等于 `["丙"]`
- **AND** `BaziResult.YongshenMissing` 等于 `["癸"]`

#### Scenario: 调候用神在原局完全缺位 → 回退扶抑

- **WHEN** 调候字典给定 yongshen 天干列表为 `["丙", "癸"]`、原局透干与所有藏干中都不含"丙"也不含"癸"
- **THEN** `BaziResult.YongshenStatus` 等于 `"tiaohou_miss_fallback_fuyi"`
- **AND** `BaziResult.YongshenMissing` 等于 `["丙", "癸"]`
- **AND** `BaziResult.Yongshen` 与 `BaziResult.Jishen` 由 `calcWeightedYongshen`（扶抑）逻辑产生
- **AND** `BaziResult.YongshenGans` 为空数组

#### Scenario: 调候字典查不到该 dayGan_monthZhi 组合

- **WHEN** 排盘传入的日干_月支组合在 tiaohou_dict 中不存在（理论 120 条全覆盖，仅作为防御性分支）
- **THEN** `BaziResult.YongshenStatus` 等于 `"tiaohou_dict_missing"`
- **AND** `BaziResult.Yongshen` 与 `BaziResult.Jishen` 由扶抑逻辑产生

### Requirement: yongshen/jishen 输出格式向后兼容

系统 SHALL 保持 `BaziResult.Yongshen` 与 `BaziResult.Jishen` 仍为五行中文名集合的字符串（含义为"集合"，无分隔符），以保持下游字符串包含匹配（如 `strings.Contains(natal.Yongshen, "火")`）的兼容性。

#### Scenario: 调候命中单一五行

- **WHEN** 调候用神天干列表为 `["丁", "丙"]`，且至少一个命中（丁与丙皆为火）
- **THEN** `BaziResult.Yongshen` 等于 `"火"`（去重后单一五行）

#### Scenario: 调候命中多五行

- **WHEN** 调候用神天干列表为 `["丙", "癸"]`，丙与癸均命中（丙=火、癸=水）
- **THEN** `BaziResult.Yongshen` 字符串同时包含 "火" 与 "水"
- **AND** `BaziResult.Jishen` 字符串包含克/泄火与克/泄水的五行

#### Scenario: 下游 strings.Contains 匹配

- **WHEN** 流年信号引擎 `getYongshenBaseline` 调用 `strings.Contains(natal.Yongshen, lnWxCN)` 检查流年天干所属五行是否属用神
- **THEN** 命中条件与历史扶抑算法时代保持一致（即五行包含语义不变）

### Requirement: 调候字典封装查询函数

系统 SHALL 暴露 `LookupTiaohouYongshen(dayGan, monthZhi string) []string` 供 yongshen 主流程查询调候用神天干列表，函数返回字典中 `Yongshen` 字段的副本（防止外部修改）。

#### Scenario: 命中已知组合

- **WHEN** 以 `dayGan="甲"`、`monthZhi="寅"` 调用 `LookupTiaohouYongshen`
- **THEN** 返回切片包含 `"丙"` 与 `"癸"`（按字典定义）

#### Scenario: 未命中组合返回 nil

- **WHEN** 以非法 `dayGan="X"` 调用 `LookupTiaohouYongshen`
- **THEN** 返回 `nil` 或空切片
- **AND** 不引发 panic

### Requirement: 调候用神原局命中检测

系统 SHALL 提供调候命中检测逻辑：将排盘四柱（年/月/日/时）的天干合集，与各地支的全部藏干合集合并为"原局可见与潜在天干集合"，与调候字典返回的天干列表求交集。

#### Scenario: 透干集合构造

- **WHEN** 排盘四柱天干为 ["甲","乙","丙","丁"]
- **THEN** 原局透干集合包含 "甲"、"乙"、"丙"、"丁"

#### Scenario: 藏干集合构造

- **WHEN** 排盘四柱地支为 ["寅","卯","巳","午"]
- **THEN** 原局藏干集合包含 寅藏（甲、丙、戊）、卯藏（乙）、巳藏（丙、戊、庚）、午藏（丁、己）的并集

#### Scenario: 调候命中判定

- **WHEN** 调候字典天干列表为 `["丙", "癸"]`、原局透+藏并集为 `{"甲","乙","丙","丁","戊","庚","己"}`
- **THEN** 命中天干为 `["丙"]`、缺位天干为 `["癸"]`

### Requirement: yongshen_status 字段对外可见

系统 SHALL 在 `BaziResult` 结构体上暴露 `YongshenStatus` / `YongshenGans` / `JishenGans` / `YongshenMissing` 字段，并在排盘 JSON 响应中包含这些字段（snake_case），供前端与运营诊断面板读取。

#### Scenario: 命盘 JSON 响应包含状态

- **WHEN** 调用 `POST /api/bazi/calculate` 返回排盘结果
- **THEN** 响应 JSON 中包含 `yongshen_status` 字段、且取值为 `"tiaohou_hit"` / `"tiaohou_miss_fallback_fuyi"` / `"tiaohou_dict_missing"` 之一
- **AND** 响应 JSON 中包含 `yongshen_gans` 数组、`jishen_gans` 数组、`yongshen_missing` 数组（其中任一可能为空数组）

### Requirement: 删除现有"急用"短路

系统 SHALL 在 yongshen 计算入口中删除"三冬无火 → 火木 / 三夏无水 → 水金"的硬编码短路；该路径的命盘统一改走调候字典主流程。

#### Scenario: 子月日干甲走字典而非短路

- **WHEN** 排盘日干为"甲"、月支为"子"、原局火 = 0
- **THEN** yongshen 推算调用 `LookupTiaohouYongshen("甲","子")` 得到字典定义的天干列表（如 `["丙","庚"]`）
- **AND** 不再硬编码返回 `("火木","水金土")`
- **AND** 若原局缺丙也缺庚，则回退至扶抑（YongshenStatus = `"tiaohou_miss_fallback_fuyi"`）
