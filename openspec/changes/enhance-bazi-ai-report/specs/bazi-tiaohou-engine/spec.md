## ADDED Requirements

### Requirement: Tiaohou Seasonal Adjustment Lookup
后端引擎 SHALL 实现基于《穷通宝鉴》的调候用神查表功能：给定日主天干与出生月令地支，返回调候用神文本说明。

#### Scenario: Standard Lookup Returns Result
- **WHEN** `LookupTiaohou(dayGan, monthZhi)` 被调用，且日主天干与月令地支均为合法值
- **THEN** 函数返回非空字符串，格式如：`"癸水（首选）、丙火（次选）"`

#### Scenario: Result Included in BaziResult
- **WHEN** `bazi.Calculate()` 执行完成
- **THEN** 返回的 `BaziResult.Tiaohou` 字段包含对应的调候用神说明文本

#### Scenario: Tiaohou Exposed via Calculate API
- **WHEN** 前端调用 `POST /api/bazi/calculate`
- **THEN** 响应体中 `result.tiaohou` 字段包含调候用神说明（非空字符串）
