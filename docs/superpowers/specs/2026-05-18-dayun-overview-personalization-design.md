# 大运总览个人化生成 设计文档

**日期**：2026-05-18
**作者**：刘明（brainstorming 产物）
**范围**：`frontend/src/components/DayunTimeline.tsx` 中「大运总览」一句话 + 「趋势关键词」chip 的内容生成逻辑

---

## 问题

`DayunTimeline.tsx:106-109` 当前的 `getDayunSummary()` 是一个固定模板：

```ts
return `${gan}${zhi}运以${gan_shishen}透干、${zhi_shishen}坐支为主，${di_shi}之势宜先看节奏，再看流年触发点。`
```

读起来像废话，原因：

1. **0 个人化** — 只用了 `gan/zhi/gan_shishen/zhi_shishen/di_shi` 五个字段，全部是上面卡片已经显示过的数据，等于把卡片字面读一遍。
2. **没碰用神/忌神** — 同一个十神「七杀」，对身弱身弱、用神忌神，意义完全相反；当前模板对此一视同仁。
3. **「宜先看节奏，再看流年触发点」是套话** — 任何大运都能套上，没有指向性。
4. **趋势关键词 chip 也是静态字典**（`DayunTimeline.tsx:63-74`）—— 「七杀 → 压力 · 行动 · 突破」无论喜忌都同样输出。

## 目标

把「大运总览」改成基于用户用神/忌神/身强弱/调候缺位的确定性公式，输出 1-3 句**有判断、不重复卡片信息**的总览，并把「趋势关键词」chip 同步个人化。

非目标：
- 不调 LLM（已用现有 `ai_dayun_summaries` 缓存的 AI 长版本另在 `LiuYueDrawer` 里）
- 不改后端 API
- 不改十神主气 / 五行主气两个 chip
- 不改卡片网格、流年面板、神煞 modal

---

## 架构

```
ResultPage.tsx
  │
  │ 新增 props: yongshen, jishen, wuxing, tiaohou
  ▼
DayunTimeline.tsx ─────────► src/lib/dayunOverview.ts (新文件)
  │ 调用 buildDayunOverview(input)
  ▼
返回 { prose, trendKeywords, ganPolarity, zhiPolarity }
  │
  ▼
<span>{prose}</span>
<span>趋势关键词：{trendKeywords}</span>
```

**文件改动**：
- 新建 `frontend/src/lib/dayunOverview.ts` — 纯函数 + 三张查表
- 改 `frontend/src/components/DayunTimeline.tsx`
  - 删 `TREND_KEYWORDS`（line 63-74）、`getTrendKeywords()`（line 101-104）、`getDayunSummary()`（line 106-109）
  - 改 `DayunTimelineProps` 加 4 个可选 props
  - 改 dayun summary strip 渲染：调用 `buildDayunOverview()`
- 改 `frontend/src/pages/ResultPage.tsx` — 给 `<DayunTimeline>` 多传 4 个 props
- 新建 `frontend/tests/dayun-overview.test.mjs`

后端 0 改动。

---

## 数据契约

```ts
// frontend/src/lib/dayunOverview.ts

export type Polarity = 'xi' | 'ji' | 'zhong'  // 喜 / 忌 / 中

export interface DayunOverviewInput {
  dayun: {
    gan: string                  // '壬'
    zhi: string                  // '午'
    gan_shishen: string          // '七杀'
    zhi_shishen: string          // '劫财'
    di_shi: string               // '帝旺'
  }
  yongshen: string               // '丙火' 或 '火' 或 '木火' — '' 表示未生成
  jishen: string                 // '壬癸水' 或 '水' — '' 表示未生成
  wuxing: {                      // BaziResult.wuxing
    mu: number; huo: number; tu: number; jin: number; shui: number
  }
  dayGanWuxing: string           // '木' / '火' / '土' / '金' / '水'
  tiaohou?: {
    expected: string[]   // 理论应有的调候用神天干集合，例 ['丙','丁']
    tou: string[]        // 已透出在四柱天干的子集
    cang: string[]       // 已藏于地支的子集
    text: string         // 穷通宝鉴释义
  } | null
}

export interface DayunOverviewOutput {
  prose: string
  trendKeywords: string
  ganPolarity: Polarity
  zhiPolarity: Polarity
}

export function buildDayunOverview(input: DayunOverviewInput): DayunOverviewOutput
```

`DayunTimeline` 新增的 props（皆可选）：

```ts
interface DayunTimelineProps {
  // ── 现有 ──
  dayun: DayunItem[]
  birthYear: number
  startYunSolar?: string
  dayGan: string
  gender?: string
  pillarsLabel?: string
  chartId?: string

  // ── 新增 4 个，可选 ──
  yongshen?: string
  jishen?: string
  wuxing?: { mu: number; huo: number; tu: number; jin: number; shui: number }
  tiaohou?: { expected: string[]; tou: string[]; cang: string[]; text: string } | null
}
```

`dayGanWuxing` 不需新增 prop —— 由现有 `dayGan` 经 `GAN_WUXING` 字典推导。

---

## 算法

每个大运调用一次 `buildDayunOverview`，依次跑 4 步，得到 5 个标签：

```
ganPolarity      ∈ {xi, ji, zhong}
zhiPolarity      ∈ {xi, ji, zhong}
dayStrength      ∈ {wang, ruo}
ganZhiRelation   ∈ {tongGen, gaiTou, jieJiao, none}
tiaohouFit       ∈ {buZu, weiDaoWei, weiJi, skip}
```

### 步骤 1：判断喜忌 polarity

- `dayun.gan` 五行用 `GAN_WUXING` 字典（已存在）；`dayun.zhi` 主气用新增 `ZHI_MAIN_WUXING` 字典
- 五行字段统一用 1 字（木/火/土/金/水）做 substring 匹配 `yongshen.includes(wx)` —— 与项目现有约定一致
- yongshen 含该五行 → `'xi'`；jishen 含 → `'ji'`；都不含 → `'zhong'`
- yongshen/jishen 为空字符串时，两个 polarity 都强制 `'zhong'`

### 步骤 2：判断身强弱

复用 `backend/pkg/bazi/engine.go:574` 的 40% 规则，前端复算：

```ts
function resolveDayStrength(wuxing, dayGanWuxing): 'wang' | 'ruo' {
  const helpMap = {
    木: ['mu','shui'],   // 同/印
    火: ['huo','mu'],
    土: ['tu','huo'],
    金: ['jin','tu'],
    水: ['shui','jin'],
  }
  const help = helpMap[dayGanWuxing] ?? []
  const helpPct = help.reduce((s,k) => s + (wuxing[k] ?? 0), 0)
  return helpPct > 40 ? 'wang' : 'ruo'
}
```

容错：`wuxing` 为空 → 默认 `'ruo'`。

### 步骤 3：判断干支关系

四种状态，互斥：

```
通根:   dayun.gan 五行 == dayun.zhi 主气五行    例: 甲寅(木|木)
盖头:   dayun.gan 五行 克 dayun.zhi 五行         例: 壬午(水克火)
截脚:   dayun.zhi 五行 克 dayun.gan 五行         例: 丙子(水克火，子在下克丙)
none:   其余（生、被生、同类非通根）              例: 戊午(火生土)
```

克制字典：`木克土 / 土克水 / 水克火 / 火克金 / 金克木`。

### 步骤 4：判断调候补足

```
tiaohou 为 null 或 expected 为空
  ↓
  skip（第三句省略）

tiaohou.expected 有数据
  ↓
  missingGans = expected - (tou ∪ cang)
  ↓
  missingGans 为空（命局已自带调候）
    ↓ skip

  missingGans 非空
    ↓
    missing 五行集合 M（每个 gan 经 GAN_WUXING 映射）
    ↓
    dayun.gan / dayun.zhi 含 M？
      完全不含                  → weiJi     "未及"
      含且不被克                → buZu      "补足"
      含但被另一支克（盖头）     → weiDaoWei "未到位"
```

「被克」依据步骤 3 的 `ganZhiRelation === 'gaiTou'`。

---

## 文案模板

### 总览句结构

```
{gan}{zhi}运（{gan_shishen}为{POL[ganP]}·{zhi_shishen}为{POL[zhiP]}）：
{BODY1[ganShishen][dayStrength]}；
{BODY2[diShiBucket] + relationSuffix[ganZhiRelation]}。
{BODY3[tiaohouFit]}。
```

- `POL = { xi: '喜', ji: '忌', zhong: '中' }`
- `tiaohouFit === 'skip'` 时整个第三句省略（含前面的句号）
- yongshen + jishen 都为空 → 标题简化为 `{gan}{zhi}运：`（去掉括号 polarity 段）

### BODY1 查表（10 十神 × 2 身强弱 = 20）

key 用 `gan_shishen`（不用 `zhi_shishen`，避免句子分裂）：

| 十神 | 身旺（wang） | 身弱（ruo） |
|---|---|---|
| 比肩 | 同行竞争分薄资源 | 兄弟朋友助身有力 |
| 劫财 | 损财争夺、合作伤利 | 同道分担、压力有人共担 |
| 食神 | 财源外吐、口腹之享 | 才华外泄、气力分散 |
| 伤官 | 才名突破、敢破规则 | 才华伤身、易招是非 |
| 正财 | 经营得利、稳定积累 | 财多身弱、力不从心 |
| 偏财 | 偏门机会、流动资金 | 财来财去、难以聚守 |
| 正官 | 事业晋升、责任加码 | 官杀压身、易受规则约束 |
| 七杀 | 立威破局、事业突破 | 身弱遭杀克身，压力与突发事件增多 |
| 正印 | 印重身旺反招迟滞 | 学习/贵人/资格类机会成形 |
| 偏印 | 转型旁门、思虑成局 | 灵感/研究/孤独感提升 |

### BODY2 — 地势 × 干支关系

12 长生归并 3 档：

```
旺档（帝旺·临官·长生·冠带）→ "{zhi}{diShi}得位有力"
中档（沐浴·养·胎·墓）        → "{zhi}{diShi}态势中等"
衰档（衰·病·死·绝）          → "{zhi}{diShi}气势减弱"
```

干支关系后缀（追加在档语之后）：

```
通根:   "，{gan}通根{zhi}得力"
盖头:   "，但被{gan}盖头压制"
截脚:   "，反被{zhi}截脚虚浮"
none:   ""
```

例：
- 壬午（帝旺+盖头） → `"午帝旺得位有力，但被壬盖头压制"`
- 甲寅（临官+通根） → `"寅临官得位有力，甲通根寅得力"`
- 丙子（胎+截脚） → `"子胎态势中等，反被子截脚虚浮"`

### BODY3 — 调候补足（3 条 + skip）

```
buZu:        "{matchedGan}{matchedWuxing}透出，正补足命局所缺调候"
weiDaoWei:   "命局所需{missingWuxing}虽现于运中，却被{coverGan}压制，调候未到位"
weiJi:       "命局所缺{missingWuxing}未在此运补足，需外接调候助力"
skip:        ""
```

- `missingGans = tiaohou.expected - (tiaohou.tou ∪ tiaohou.cang)`；`missingWuxing` 取 `missingGans[0]` 经 `GAN_WUXING` 映射（同一句话只展示一个主导五行，避免句子膨胀）
- `matchedGan` 选取规则：先看 `dayun.gan`，若其五行 ∈ missing 集合 → 用 `dayun.gan`；否则看 `dayun.zhi` 主气，若主气五行 ∈ missing 集合 → 用主气对应天干（如午→丁）；都不命中时 `tiaohouFit` 不应为 `buZu`
- `coverGan` = `dayun.gan`（盖头定义中克者必在天干）

### 趋势关键词 chip（10 × 2）

```ts
const TREND: Record<string, { xi: string; ji: string }> = {
  比肩: { xi: '同道 · 自立 · 稳进',   ji: '分薄 · 竞争 · 节制' },
  劫财: { xi: '合伙 · 协力 · 取舍',   ji: '损财 · 争夺 · 化解' },
  食神: { xi: '表达 · 享受 · 作品',   ji: '泄气 · 分心 · 节用' },
  伤官: { xi: '突破 · 才名 · 创意',   ji: '是非 · 锋芒 · 收敛' },
  正财: { xi: '经营 · 责任 · 积累',   ji: '负重 · 守财 · 量力' },
  偏财: { xi: '机会 · 流动 · 人脉',   ji: '财去 · 投机 · 谨慎' },
  正官: { xi: '事业 · 晋升 · 成就',   ji: '约束 · 规矩 · 顺应' },
  七杀: { xi: '突破 · 立威 · 决断',   ji: '压力 · 守势 · 化解' },
  正印: { xi: '学习 · 贵人 · 资质',   ji: '迟滞 · 内耗 · 取舍' },
  偏印: { xi: '研究 · 灵感 · 转型',   ji: '孤独 · 怀疑 · 沉淀' },
}
```

key 取 `dayun.gan_shishen`，按 `ganPolarity` 选 `xi` / `ji`；`ganPolarity === 'zhong'` 时取 `xi` 兜底（避免空白）。

---

## 容错矩阵

| 缺什么 | polarity | 身强弱 | 关系 | tiaohou 句 | 总览句样式 |
|---|---|---|---|---|---|
| 一切齐全 | 喜/忌 | 旺/弱 | 4 状态 | 3 状态 | 完整 3 句 |
| yongshen+jishen 全空 | 全 zhong | 仍计算 | 仍计算 | 仍计算 | 去掉括号 polarity 段，正文照旧 |
| wuxing 空 | 仍计算 | 默认 ruo | 仍计算 | 仍计算 | 完整 3 句（按身弱） |
| tiaohou 空 / `expected` 空 / `missingGans` 空 | 仍计算 | 仍计算 | 仍计算 | skip | 只 2 句 |
| dayun 字段缺失（极端） | — | — | — | — | 兜底原文案：`选择一段大运后查看该十年流年节奏。`；chip = `节奏 · 观察 · 平衡` |
| 十神 / di_shi 在字典里 miss | — | — | — | — | 同上兜底 |

---

## 测试

文件：`frontend/tests/dayun-overview.test.mjs`，使用 `node:test`，对齐项目现有 `brand-settings.test.mjs` / `ten-god-relation-ux.test.mjs` 风格。

### 静态接线测试

```js
test('DayunTimeline imports buildDayunOverview from lib', () => {
  const src = read('src/components/DayunTimeline.tsx')
  assert.match(src, /import\s+\{\s*buildDayunOverview/)
})

test('DayunTimeline no longer hardcodes "宜先看节奏"', () => {
  const src = read('src/components/DayunTimeline.tsx')
  assert.doesNotMatch(src, /宜先看节奏/)
})

test('ResultPage passes yongshen/jishen/wuxing/tiaohou to DayunTimeline', () => {
  const src = read('src/pages/ResultPage.tsx')
  assert.match(src, /<DayunTimeline[\s\S]+?yongshen=/)
  assert.match(src, /<DayunTimeline[\s\S]+?jishen=/)
  assert.match(src, /<DayunTimeline[\s\S]+?wuxing=/)
  assert.match(src, /<DayunTimeline[\s\S]+?tiaohou=/)
})
```

### `buildDayunOverview()` 单元测试（8 个）

| # | 场景 | 关键断言（用 `assert.match`） |
|---|---|---|
| 1 | 截图原例：身弱+忌杀+午火帝旺+盖头+缺火 | prose 含「七杀为忌」「身弱遭杀克身」「但被壬盖头压制」「调候未到位」 |
| 2 | 身旺+喜杀（用神火，大运七杀为火） | prose 含「立威破局」；chip 含「突破 · 立威」 |
| 3 | 通根例：甲寅 | prose 含「甲通根寅得力」 |
| 4 | 截脚例：丙子 | prose 含「反被子截脚虚浮」 |
| 5 | 调候补足：缺火，丁未运不被克 | prose 含「正补足命局所缺调候」 |
| 6 | 调候未及：缺火，庚申运完全无火 | prose 含「未在此运补足」 |
| 7 | yongshen/jishen 全空 | 标题不含「（七杀为」，BODY1/BODY2 仍出 |
| 8 | dayun.gan_shishen 不在字典 | 返回 fallback「节奏 · 观察 · 平衡」 |

每个 case 构造完整 `DayunOverviewInput` 字面量，断言用 `match`（不字面比对整句，允许后续微调措辞）。

### 手动验收清单

实施完成后人工过：

- [ ] 打开已生成报告的 chartId，点不同大运卡片，总览 3 句应该立刻刷新且各不相同
- [ ] 卸载报告（未生成报告的 chartId）总览还能出无 polarity 的 2 句版本
- [ ] chip「趋势关键词」在身旺/身弱不同盘下对同一个十神应展示不同短语
- [ ] PDF 打印 / PNG 分享导出时这块 UI 不应破版（CSS 不动，但要看一眼）

---

## 不做的事（YAGNI）

- 与命局四支的冲/合/刑/害判定 —— 第三句留给调候已经够；冲合关系数据量大且与「总览」语境略错位
- 关键流年标注 —— 事件信号引擎里有此能力，但属流年面板的职责，不属总览
- AI 生成版本 —— 已有 `ai_dayun_summaries` 表存的是 LiuYueDrawer 用的长版本，与本短总览不冲突，且本次仅做确定性公式版
- 十神主气 / 五行主气 chip 个人化 —— 信息重复
