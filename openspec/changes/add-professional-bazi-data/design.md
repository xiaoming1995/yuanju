## Context

随着平台从玩具级排盘向专业级迈进。必须在“四柱牌堆”里展示“十神”（Shi Shen）。命理中十神是日主对其他七个字的生克关系的总称，是最关键的推理符号。此外，长生（Chang Sheng）代表每个天干或地支的状态生命周期。

## Goals / Non-Goals

**Goals:**
- 提供天干的十神（如正财、偏印）。
- 提供地支的主气十神，甚至包含藏干各自的十神。
- 提供日主相对于各柱地支的十二长生状态（长生、沐浴等）。
- 将这些数据呈现在 UI 内且不破坏现有界面的美学和谐。

**Non-Goals:**
- 不涉及大运或流年粒度的神煞计算（防止前端数据爆炸，先专心完成本命盘四柱的专业化）。

## Decisions

### 1. 数据模型下发设计
在 `BaziResult` 中增加扁平化字面量：
```go
	YearGanShiShen  string `json:"year_gan_shishen"`
	MonthGanShiShen string `json:"month_gan_shishen"`
	DayGanShiShen   string `json:"day_gan_shishen"` // 通常为 "日主" 或 "元男"
	HourGanShiShen  string `json:"hour_gan_shishen"`

    // 地支的主十神
	YearZhiShiShen  string `json:"year_zhi_shishen"`
    // ...以此类推

    // 十二长生状态
    YearDiShi       string `json:"year_di_shi"`
```
这些都可以通过调用现有的底层天文库接口顺手推演。

### 2. UI 组件升级
使用更细的字体或弱化的颜色在对应柱子的上下边缘打印出这些信息。
- 排布逻辑：
  `【天干十神】`
  `[ 天干字 ]`
  `[ 地支字 ]`
  `【地支十神】`

## Risks / Trade-offs

- **卡片纵向空间被拉长**：增加信息量一定会导致卡片在移动端的纵深变大，我们需要确保 CSS 在移动端的适配依然自然，十神信息字体必须控制在 12px~13px。
