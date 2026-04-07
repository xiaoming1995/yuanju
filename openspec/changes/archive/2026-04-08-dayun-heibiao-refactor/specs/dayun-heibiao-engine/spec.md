## ADDED Requirements

### Requirement: DayunRule 数据结构
系统 SHALL 使用 `DayunRule` 结构体存储每个日干月令组合的大运评级规则，包含以下字段：
- `GanXi []string`：调候喜天干列表
- `GanJi []string`：调候忌天干列表
- `ZhiXi []string`：金不换喜地支列表
- `ZhiJi []string`：金不换忌地支列表
- `Note string`：特注说明

#### Scenario: 查表返回完整规则
- **WHEN** 以日干"丙"、月支"戌"查询 DayunRule
- **THEN** 返回非空规则，GanXi 包含调候喜天干，ZhiXi 包含金不换喜地支

#### Scenario: 不存在的组合返回 nil
- **WHEN** 以非法日干"X"查询 DayunRule
- **THEN** 返回 nil

---

### Requirement: 合表数据完整覆盖
系统 SHALL 包含 120 条 DayunRule 数据（10个天干 × 12个月支），数据来源为梁湘润《子平教材讲义》金不换大运——调候用神合表。

#### Scenario: 全量数据覆盖验证
- **WHEN** 遍历十天干（甲乙丙丁戊己庚辛壬癸）× 十二月支（子丑寅卯辰巳午未申酉戌亥）
- **THEN** 每个组合均能查到非空的 DayunRule

---

### Requirement: 前5年天干评级
系统 SHALL 以大运天干匹配 DayunRule.GanXi/GanJi 来判定前5年评级：
- 大运天干 ∈ GanXi → 吉
- 大运天干 ∈ GanJi → 凶
- 两者均不匹配 → 平

#### Scenario: 天干命中喜用
- **WHEN** 日干丙、月支戌、大运天干为"甲"（甲 ∈ GanXi）
- **THEN** 前5年评级为"吉"

#### Scenario: 天干命中忌神
- **WHEN** 日干丙、月支戌、大运天干为忌神天干
- **THEN** 前5年评级为"凶"

#### Scenario: 天干中性
- **WHEN** 日干丙、月支戌、大运天干不在 GanXi 也不在 GanJi 中
- **THEN** 前5年评级为"平"

---

### Requirement: 后5年地支评级
系统 SHALL 以大运地支匹配 DayunRule.ZhiXi/ZhiJi 来判定后5年评级：
- 大运地支 ∈ ZhiXi → 吉
- 大运地支 ∈ ZhiJi → 凶
- 两者均不匹配 → 平

#### Scenario: 地支命中喜用
- **WHEN** 日干丙、月支戌、大运地支为喜用地支（如"寅"）
- **THEN** 后5年评级为"吉"

#### Scenario: 地支命中忌神
- **WHEN** 日干丙、月支戌、大运地支为忌神地支（如"午"）
- **THEN** 后5年评级为"凶"

---

### Requirement: 评级描述文本生成
系统 SHALL 生成评级描述文本，格式为：
- 前5年："{大运天干}为{喜/忌/中性}天干，前5年{评级结果}。{特注说明}"
- 后5年："{大运地支}为{喜/忌/中性}地支，后5年{评级结果}。"

#### Scenario: 描述文本包含关键信息
- **WHEN** 计算完成大运评级
- **THEN** qian_desc 和 hou_desc 字段均非空，包含天干/地支名称和评级原因

---

### Requirement: API 响应格式兼容
系统 SHALL 保持 `jin_bu_huan` 字段的 JSON 格式不变：
```json
{
  "qian_level": "吉|凶|平",
  "qian_desc": "...",
  "hou_level": "吉|凶|平",
  "hou_desc": "...",
  "verse": "..."
}
```

#### Scenario: API 响应格式一致性
- **WHEN** 调用 /api/bazi/calculate 接口
- **THEN** 大运数组中每个条目的 jin_bu_huan 字段包含 qian_level、qian_desc、hou_level、hou_desc 字段
