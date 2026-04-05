## ADDED Requirements

### Requirement: 流月数据查询
系统 SHALL 提供 `POST /api/bazi/liu-yue` 接口，接受流年公历年（`liu_nian_year`，整数）和日主天干（`day_gan`，单个汉字）作为输入，返回该年份 12 个流月（寅月至丑月）的完整数据，以及当前所处流月的 index（基于接口调用时的服务器日期与节气表对照计算）。

#### Scenario: 正常请求返回 12 个流月
- **WHEN** 客户端 POST `{ "liu_nian_year": 2026, "day_gan": "甲" }` 到 `/api/bazi/liu-yue`
- **THEN** 系统返回包含 12 个流月对象的数组，每个对象含 `index`(0-11)、`month_name`（地支月名，如"寅月"）、`gan_zhi`（干支，如"丙寅"）、`gan_shishen`（天干十神）、`zhi_shishen`（地支十神）、`jie_qi_name`（节气名）、`start_date`（节气起始公历日期，格式 YYYY-MM-DD）、`end_date`（节气结束日期）

#### Scenario: 返回当前流月 index
- **WHEN** 接口被调用时服务器日期为 2026-04-05（清明节气当天）
- **THEN** 响应包含 `"current_month_index": 2`（辰月，index=2）

#### Scenario: 缺少必要参数时返回错误
- **WHEN** 请求体缺少 `day_gan` 字段
- **THEN** 系统返回 HTTP 400，body 为 `{ "error": "missing required field: day_gan" }`

#### Scenario: 年份切换查询
- **WHEN** 客户端 POST `{ "liu_nian_year": 2025, "day_gan": "甲" }` 到 `/api/bazi/liu-yue`
- **THEN** 系统返回 2025 年 12 个流月的正确干支（五虎遁基于 2025 乙年重新计算），`current_month_index` 仍基于今日节气计算

---

### Requirement: 节气起止日期精确对应
系统 SHALL 为每个流月返回准确的节气起始和结束公历日期，日期以实际天文节气时刻确定（由 `lunar-go` 的 `GetJieQiTable()` 计算），而非固定日期规则。丑月（index=11）的结束日期 SHALL 为次年立春前一天。

#### Scenario: 节气日期随年份变化
- **WHEN** 查询 2024 年与 2025 年流月
- **THEN** 两年的立春日期不同（如 2024-02-04 vs 2025-02-03），系统返回各年实际节气日期而非硬编码值

#### Scenario: 丑月跨年结束日期
- **WHEN** 查询任意年份的流月，index=11（丑月）
- **THEN** `end_date` 为次年立春日期的前一天
