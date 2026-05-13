## MODIFIED Requirements

### Requirement: BaziResult 包含完整命盘计算数据
BaziResult 结构体 SHALL 包含四柱、五行、十神、用神忌神、调候用神、大运、以及命格（ming_ge、ming_ge_desc）字段，随 `POST /api/bazi/calculate` 返回给前端。

#### Scenario: 计算八字后包含命格字段
- **WHEN** 前端调用 `POST /api/bazi/calculate`，后端完成排盘计算
- **THEN** 响应 JSON 的 result 对象中包含 `ming_ge`（命格名称字符串）和 `ming_ge_desc`（命格说明文字字符串）两个字段

#### Scenario: 旧命盘快照无命格字段时降级
- **WHEN** 前端从历史记录加载旧命盘（result_json 中无 ming_ge 字段）
- **THEN** 前端命盘顶部不显示命格 Badge，不报错，其余信息正常展示
