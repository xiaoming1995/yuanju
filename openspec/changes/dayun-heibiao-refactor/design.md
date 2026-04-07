## Context

当前大运评级基于两套独立字典：
- `jinBuHuanDict`：以 `GoodDirections/BadDirections`（方向字符串）为判定依据
- `tiaohouDict`：以 `Yongshen`（调候用神天干）为数据，但未参与大运评级

两套字典存在矛盾（如丙_戌：调候喜壬水 vs 金不换忌北方水），导致评级结果不可靠。

梁湘润《子平教材讲义》提供了权威「合表」，直接给出每个日干月令下的：
1. **调候用神喜忌天干**（前5年判定依据）
2. **金不换喜忌地支**（后5年判定依据）
3. **特注**（附加条件说明）

## Goals / Non-Goals

**Goals:**
- 用合表数据完全替换现有的方向映射逻辑
- 实现精确到天干/地支的大运评级（而非方向级别）
- 保持 API 响应格式 `qian_level/qian_desc/hou_level/hou_desc` 不变
- 前端零改动

**Non-Goals:**
- 不做格局用神判定（后续阶段）
- 不修改调候用神展示模块（`tiaohou.go` 保持原样）
- 不处理流年评级

## Decisions

### 决策1：数据结构采用双层喜忌模型

**选择**：新建 `DayunRule` 结构体，包含 `GanXi/GanJi`（天干喜忌）和 `ZhiXi/ZhiJi`（地支喜忌）两层。

**替代方案A**：从 `tiaohouDict.Yongshen` 动态推导忌神（反克关系）
- 否决原因：合表已给出精确忌神天干，推导引入误差

**替代方案B**：在 `TiaohouRule` 中增加忌神字段
- 否决原因：调候字典和金不换字典职责不同，应保持独立

```go
type DayunRule struct {
    GanXi  []string `json:"gan_xi"`   // 调候喜天干（前5年依据）
    GanJi  []string `json:"gan_ji"`   // 调候忌天干（前5年依据）
    ZhiXi  []string `json:"zhi_xi"`   // 金不换喜地支（后5年依据）
    ZhiJi  []string `json:"zhi_ji"`   // 金不换忌地支（后5年依据）
    Note   string   `json:"note"`     // 特注（如"身强"、"天折"）
}
```

### 决策2：评级算法为直接查表匹配

```
前5年评级：大运天干 ∈ GanXi → 吉 | ∈ GanJi → 凶 | 否则 → 平
后5年评级：大运地支 ∈ ZhiXi → 吉 | ∈ ZhiJi → 凶 | 否则 → 平
```

不再做五行转换或方向映射，完全是字符串精确匹配。

### 决策3：特注字段纳入描述但不影响评级

合表中的「特注」（如"身强"、"有水"、"天折"等）作为文本描述展示，当前阶段不作为评级修正条件。

### 决策4：保留 `Verse` 诗句和旧 `tiaohouDict`

- `Verse` 诗句从 `JinBuHuanEntry` 迁移到 `DayunRule` 中（作为展示文本）
- `tiaohouDict` 保持原样，继续用于 `BaziResult.Tiaohou` 字段展示

### 决策5：合表中的「天」标注处理

合表中某些地支后标「(天)」表示有天折之险。在 `ZhiJi` 中录入该地支，同时在 `Note` 中标注"天折"。

## Risks / Trade-offs

- **数据录入量大**：120条 × 多字段，录入可能出错 → 通过单元测试验证关键命例
- **合表截图可能有遗漏**：浏览器截图未覆盖全部10个天干 → 需要再次截取壬、癸等天干的表格
- **比劫天干统一归"平"**：合表中有些月令对比劫也有明确喜忌标注 → 以合表数据为准录入，不做特殊处理
