## Context

大运评级底层由 `CalcJinBuHuanDayun(dayGan, monthZhi, dayunZhi)` 计算，输出 `JinBuHuanResult{Level, Keyword, Text}`。当前逻辑只看大运**地支**的四方方位（东/南/西/北），与金不换规则里的 GoodDirections/BadDirections 匹配后给出吉凶。大运**天干**完全未被利用，而 `DayunItem.GanShiShen` 字段（天干十神）已经存在于计算结果中，数据基础具备。

## Goals / Non-Goals

**Goals:**
- 新增天干修正层，实现「地支定基调 + 天干做加权修正」的大运评级双层模型
- 为土干（戊/己）提供专属 `EarthGanEval` 评价字段（阶段一 nil 兜底，阶段二补数据）
- 土干兜底逻辑：当 `EarthGanEval == nil` 时，检测命局喜忌方向是否有互克对，有则判定为「通关」小吉；否则中性
- 所有改动向后兼容，现有 API 字段不变，只新增字段

**Non-Goals:**
- 不在本变更内完成全部 120 条 `EarthGanEval` 的数据填充（属阶段二迭代）
- 不改变地支方向的匹配逻辑
- 不将数据迁移到数据库（数据库迁移作为独立变更评估）

## Decisions

### 1. 结构体扩展方式

扩展 `JinBuHuanResult`（输出结构）与 `JinBuHuanRule`（规则结构）：

```go
// 输出结构：新增天干修正字段
type JinBuHuanResult struct {
    Level       string  `json:"level"`        // 综合评级（天干+地支合并后）
    Keyword     string  `json:"keyword"`
    Text        string  `json:"text"`
    ZhiLevel    string  `json:"zhi_level"`    // 地支原始评级（不受天干影响）
    GanModifier string  `json:"gan_modifier"` // "加成" / "减损" / "通关" / "中性"
    GanDesc     string  `json:"gan_desc"`     // 天干修正文字说明
}

// 规则结构：新增土干专属评价
type JinBuHuanRule struct {
    // ...现有字段...
    EarthGanEval *JBHEval `json:"earth_gan_eval"` // 戊己土干专属，nil = 用通关逻辑兜底
}
```

### 2. Level 升降矩阵

| 地支级别 | 天干=加成 | 天干=通关 | 天干=中性 | 天干=减损 |
|---------|---------|---------|---------|---------|
| 大吉 | 大吉 | 大吉 | 大吉 | 吉 |
| 吉 | 大吉 | 吉 | 吉 | 平 |
| 平 | 吉 | 小吉(吉) | 平 | 凶 |
| 凶 | 平 | 平 | 凶 | 大凶 |
| 大凶 | 凶 | 凶 | 大凶 | 大凶 |

### 3. 天干五行 → 喜忌匹配路径

方向字符串反解为五行集合（`dirToWuxing map`），再与大运天干五行比对：

```
GoodDirections → goodWx set (如 {huo, jin})
BadDirections  → badWx set
dayunGan 五行  → ganWx
if ganWx ∈ goodWx → "加成"
if ganWx ∈ badWx  → "减损"
if ganWx == "tu"  → 查 EarthGanEval，nil 则跑通关检测
else              → "中性"
```

### 4. 函数签名修改

```go
// 原签名
CalcJinBuHuanDayun(dayGan, monthZhi, dayunZhi string) *JinBuHuanResult

// 新签名（向后兼容地增加参数）
CalcJinBuHuanDayun(dayGan, monthZhi, dayunGan, dayunZhi string) *JinBuHuanResult
```

调用处 `engine.go:334` 只需额外传入 `gan` 变量。

## Risks / Trade-offs

- [土干数据缺口] 阶段一 120 条 `EarthGanEval` 均为 nil，土运全走通关兜底逻辑，精准度不如真实规则填充 → 缓解：通关逻辑本身有命理依据，优于完全中性处理；阶段二补全数据
- [升降矩阵主观性] 五行升降矩阵是根据传统命理经验设计，不同流派可能略有差异 → 缓解：升降幅度保守（最多升降一档），不会产生极端误判
- [API 字段膨胀] `jin_bu_huan` 对象增加了 3 个新字段，前端需要适配 → 缓解：字段可选展示，DayunTimeline 仅展示 `gan_modifier` 小徽章，不影响现有布局
