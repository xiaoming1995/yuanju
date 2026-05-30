# 六十甲子 · 日柱速写（静态附加层）设计

**日期：** 2026-05-30
**状态：** 已确认，待写实现计划

## 目标

在合盘结果里，给每个人按其**日柱干支**（甲子…癸亥）额外展示一段"一柱速写"——一段有性格 + 感情/相处风格味道、双关微辣但不露骨的短文。纯静态查表，**完全不影响现有任何算法**。

## 背景与方法来源

灵感来自"情色命理之六十甲子相见欢"（豆瓣，一柱论命系列）。注意：该系列名虽叫"相见欢"，**实质是单柱论命**（对每根日柱单独取象），不是两柱配对查表。可借鉴的方法内核有两条：

1. **干为表/外，支为里/内。** 一柱 = 一个人的「外显 vs 内里」。如甲子：甲(干)=表现出来的外在姿态，子(支)=内在想法/需求。能产出"外一套、内一套"的反差。
2. **一柱论命取象（走出干支自身，源自《玉井奥诀》）。** 不止读日柱十神，把挂在这根柱上的信息全拉进来整合取最象：十神、纳音意象、长生十二宫、神煞、日支的合冲刑害。

原文用情色作记忆术，**本设计只借方法，不照搬露骨措辞**。作者自己也声明"断语十不中一"——所以这是让画像**更丰富、更有人味**的补充层，不主张准确率。

## 调性（已锁定）

**双关微辣、暗示拉满，但不露骨。** 这是天花板，60 条全部按此写，不写限制级/露骨内容。校准样例：

> **甲子**：白天一身正气、规矩得能领面锦旗，关了灯完全两个人。子水是沐浴之地——外头道貌岸然、里头野。慢热却记仇式专一，是只"有耐心的狼"，不动则已，一动就盯死你，耐力还出奇地好。

> **乙丑**：长发一甩、腰肢一摆，天生勾人。丑是官杀库，女命坐下藏着一整支"后备队"——不是她招蜂引蝶，是异性自己排队上门；偏财一掺，浪漫起来不计成本。男命则是闷声撩人的惯犯，话不多、手不闲，冷着脸就把人带走了。

> **壬申**：自坐长生，一眼活泉，越掏越有、越用越旺。精力、情绪、还有别的，统统源源不断，是"供大于求"型选手——鲜活、耐折腾，就是太满，得配个接得住的。

## 架构 / 数据流

**只读、不算、不碰任何现有逻辑：**

```
participant.chart_snapshot.day_gan + day_zhi   →  "甲子"   (已有字段，只读)
        ▼
getDayPillarPortrait("甲", "子")  →  { tag, text }       (纯静态查表)
        ▼
三处渲染：结果页 / PDF / 分享图，各自复用同一张表
```

- **不进** `compatibilityPersonality.ts`、**不进**评分、**不进** LLM `personality_comparison`。三者零耦合。
- 不依赖 AI 报告是否生成（区别于现有"双方性格画像"挂在 `DeepReportNarrative` 内、要点生成才有）。日柱是命盘快照现成字段，**进结果页即有**。
- 旧数据若 `chart_snapshot` 缺失或日柱查不到 → 该人这块不渲染（返回 null），不报错。

## 数据形状

```ts
// frontend/src/lib/dayPillarPortraits.ts
export type DayPillarPortrait = { tag: string; text: string }

// tag = 4-6 字定性钩子（呼应现有画像卡 headline 风格）
// text = 双关微辣速写（2-3 句）
const DAY_PILLAR_PORTRAITS: Record<string, DayPillarPortrait> = {
  '甲子': { tag: '有耐心的狼', text: '白天一身正气……' },
  // …共 60 条
}

export function getDayPillarPortrait(dayGan: string, dayZhi: string): DayPillarPortrait | undefined {
  return DAY_PILLAR_PORTRAITS[`${dayGan}${dayZhi}`]
}
```

## 文件结构

**新增（核心逻辑全部在此，与既有文件解耦）：**
- `frontend/src/lib/dayPillarPortraits.ts` — 60 条静态表 + `getDayPillarPortrait`。
- `frontend/src/components/compatibility/DayPillarPortrait.tsx` + `.css` — 双方两张速写卡（姓名 + 日柱干支 + tag + text）；缺数据返回 null。

**修改（仅挂载，不动既有算法）：**
- `frontend/src/pages/CompatibilityResultPage.tsx` — `<SectionBasicCharts/>` 后加一行挂载。
- `frontend/src/components/CompatibilityPrintLayout.tsx` — PDF 内加一块（读 selfP/partnerP 的 day_gan/day_zhi → 同一查表函数），完整密度。
- `frontend/src/components/CompatibilityShareCard.tsx` — 分享图加一块，速写较短，原样放；过长则裁到合适长度。

**各面密度：** 屏幕 = 完整；PDF = 完整；分享图 = 完整（必要时裁短）。

## 文案生产

- 60 条**分 6 批 × 10 柱**手写。每批写完用户过一眼校准，再写下一批，避免一次性 60 条跑偏。
- 取象依据：干表/支里 + 十神 + 长生 + 纳音 + 日支刑冲；不照搬露骨原文。

## 测试

`frontend/tests/day-pillar-portrait.test.mjs`：
1. `dayPillarPortraits.ts` 含全部 60 个干支 key、无重复、无空 tag/text。
2. `getDayPillarPortrait` 命中已知柱、未知输入返回 undefined。
3. `DayPillarPortrait.tsx` 被结果页挂载；三处渲染文件均 import `getDayPillarPortrait`；该组件**不** import `compatibilityPersonality`（守解耦）。

## 边界 / 非目标

- **不动** `compatibilityPersonality.ts`、合盘评分、后端、LLM prompt、`personality_comparison`。
- 只取**日柱**，不铺年/月/时四柱。
- 不做两柱配对断语（"相见欢"原名的字面含义），本期是单人单柱速写。
- 不写露骨/限制级内容。
