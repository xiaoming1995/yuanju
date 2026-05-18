# 大运总览个人化生成 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 把 `DayunTimeline.tsx` 中的硬编码「大运总览」一句话 + 静态「趋势关键词」字典，替换为基于用户用神/忌神/身强弱/调候缺位的确定性公式生成。

**Architecture:** 纯前端改动。新增 `frontend/src/lib/dayunOverview.ts`（含 7 张查表 + 4 个判定函数 + 主入口），`DayunTimeline.tsx` 接 4 个可选 props 并调用主入口替换原逻辑，`ResultPage.tsx` 多传 4 个 prop。无后端改动。

**Tech Stack:** React 19 + TypeScript，测试用 `node:test` + `--experimental-strip-types` 直接 import `.ts`（项目 Node 23.7.0 已支持）。

**Spec:** `docs/superpowers/specs/2026-05-18-dayun-overview-personalization-design.md` (commit `48f9057`)

**Branch baseline:** `feat/export-brand-customization` HEAD `48f9057`，工作树干净。

---

## File Map

| 文件 | 操作 | 责任 |
|---|---|---|
| `frontend/src/lib/dayunOverview.ts` | 新建 | 7 张查表 + 4 helpers + `buildDayunOverview()` |
| `frontend/tests/dayun-overview.test.mjs` | 新建 | 3 静态接线 + 8 行为单测 |
| `frontend/src/components/DayunTimeline.tsx` | 改 | 删旧硬编码 + 新增 4 props + 改 summary strip 渲染 |
| `frontend/src/pages/ResultPage.tsx` | 改 | 给 `<DayunTimeline>` 多传 4 prop |

---

## Task 0: Baseline check

**Files:** none

- [ ] **Step 1: 确认工作目录**

Run: `pwd`
Expected: `/Users/liujiming/web/yuanju`

- [ ] **Step 2: 确认 git 状态干净 + 分支 + HEAD**

Run: `git -C /Users/liujiming/web/yuanju status --short && git -C /Users/liujiming/web/yuanju branch --show-current && git -C /Users/liujiming/web/yuanju log -1 --oneline`

Expected:
```
(no output from status)
feat/export-brand-customization
48f9057 docs: correct tiaohou field shape in dayun overview spec
```

- [ ] **Step 3: 确认 Node 版本支持 type-stripping**

Run: `node --version`
Expected: `v23.7.0` or higher (≥ 23.6)

- [ ] **Step 4: 跑一次 lint 当 baseline（项目已有 2 个 U+3000 警告在 PrintLayout.tsx，不要试图修，是历史债）**

Run: `cd /Users/liujiming/web/yuanju/frontend && npm run lint 2>&1 | tail -20`
Expected: 0 errors（warnings 可有 2 个 `no-irregular-whitespace` 在 `PrintLayout.tsx`，无视）

- [ ] **Step 5: 跑一次 build 当 baseline**

Run: `cd /Users/liujiming/web/yuanju/frontend && npm run build 2>&1 | tail -10`
Expected: `built in <X>s` (无 tsc 错误)

---

## Task 1: RED — write the failing tests

**Files:**
- Create: `frontend/tests/dayun-overview.test.mjs`

- [ ] **Step 1: 写完整测试文件**

Create `frontend/tests/dayun-overview.test.mjs` with this exact content:

```javascript
import { readFileSync } from 'node:fs'
import { join } from 'node:path'
import test from 'node:test'
import assert from 'node:assert/strict'

const root = new URL('..', import.meta.url).pathname
function read(path) {
  return readFileSync(join(root, path), 'utf8')
}

// ─── 静态接线测试 ───────────────────────────────────────────────────────

test('DayunTimeline imports buildDayunOverview from lib', () => {
  const src = read('src/components/DayunTimeline.tsx')
  assert.match(src, /import\s+\{\s*buildDayunOverview/)
})

test('DayunTimeline no longer hardcodes 宜先看节奏', () => {
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

// ─── 行为单测: buildDayunOverview ──────────────────────────────────────
// 通过 dynamic import 加载 .ts 源（依赖 --experimental-strip-types 启动）

const { buildDayunOverview } = await import('../src/lib/dayunOverview.ts')

test('截图原例: 壬午身弱忌杀盖头缺火', () => {
  const o = buildDayunOverview({
    dayun: { gan: '壬', zhi: '午', gan_shishen: '七杀', zhi_shishen: '劫财', di_shi: '帝旺' },
    yongshen: '丙火',
    jishen: '壬癸水',
    wuxing: { mu: 10, huo: 15, tu: 10, jin: 25, shui: 40 },
    dayGanWuxing: '火',
    tiaohou: { expected: ['丙', '丁'], tou: [], cang: [], text: '冬月需丙丁火调候' },
  })
  assert.match(o.prose, /七杀为忌/)
  assert.match(o.prose, /身弱遭杀克身/)
  assert.match(o.prose, /但被壬盖头压制/)
  assert.match(o.prose, /调候未到位/)
  assert.equal(o.ganPolarity, 'ji')
})

test('身旺喜杀: 七杀 chip 取 xi', () => {
  const o = buildDayunOverview({
    dayun: { gan: '丙', zhi: '寅', gan_shishen: '七杀', zhi_shishen: '偏印', di_shi: '长生' },
    yongshen: '丙火',
    jishen: '壬癸水',
    wuxing: { mu: 30, huo: 25, tu: 20, jin: 15, shui: 10 },
    dayGanWuxing: '金',
    tiaohou: null,
  })
  assert.match(o.prose, /立威破局/)
  assert.match(o.trendKeywords, /突破/)
  assert.equal(o.ganPolarity, 'xi')
})

test('通根: 甲寅 (七杀)', () => {
  const o = buildDayunOverview({
    dayun: { gan: '甲', zhi: '寅', gan_shishen: '七杀', zhi_shishen: '七杀', di_shi: '临官' },
    yongshen: '',
    jishen: '',
    wuxing: { mu: 20, huo: 20, tu: 20, jin: 20, shui: 20 },
    dayGanWuxing: '土',
    tiaohou: null,
  })
  assert.match(o.prose, /甲通根寅得力/)
})

test('截脚: 丙子 (正官)', () => {
  const o = buildDayunOverview({
    dayun: { gan: '丙', zhi: '子', gan_shishen: '正官', zhi_shishen: '正官', di_shi: '胎' },
    yongshen: '',
    jishen: '',
    wuxing: { mu: 20, huo: 20, tu: 20, jin: 20, shui: 20 },
    dayGanWuxing: '土',
    tiaohou: null,
  })
  assert.match(o.prose, /反被子截脚虚浮/)
})

test('调候补足: 缺火, 丁未运不被克', () => {
  const o = buildDayunOverview({
    dayun: { gan: '丁', zhi: '未', gan_shishen: '正印', zhi_shishen: '伤官', di_shi: '冠带' },
    yongshen: '丙火',
    jishen: '',
    wuxing: { mu: 30, huo: 5, tu: 30, jin: 20, shui: 15 },
    dayGanWuxing: '金',
    tiaohou: { expected: ['丙', '丁'], tou: [], cang: [], text: '' },
  })
  assert.match(o.prose, /正补足命局所缺调候/)
})

test('调候未及: 缺火, 庚申运完全无火', () => {
  const o = buildDayunOverview({
    dayun: { gan: '庚', zhi: '申', gan_shishen: '比肩', zhi_shishen: '比肩', di_shi: '临官' },
    yongshen: '',
    jishen: '',
    wuxing: { mu: 30, huo: 5, tu: 30, jin: 20, shui: 15 },
    dayGanWuxing: '金',
    tiaohou: { expected: ['丙', '丁'], tou: [], cang: [], text: '' },
  })
  assert.match(o.prose, /未在此运补足/)
})

test('yongshen/jishen 都空时标题简化, BODY1 仍出', () => {
  const o = buildDayunOverview({
    dayun: { gan: '壬', zhi: '午', gan_shishen: '七杀', zhi_shishen: '劫财', di_shi: '帝旺' },
    yongshen: '',
    jishen: '',
    wuxing: { mu: 10, huo: 15, tu: 10, jin: 25, shui: 40 },
    dayGanWuxing: '火',
    tiaohou: null,
  })
  assert.doesNotMatch(o.prose, /（七杀为/)
  assert.match(o.prose, /^壬午运：/)
  assert.match(o.prose, /身弱遭杀克身/)
})

test('字典 miss → fallback', () => {
  const o = buildDayunOverview({
    dayun: { gan: 'X', zhi: 'Y', gan_shishen: '不存在', zhi_shishen: '不存在', di_shi: 'Z' },
    yongshen: '',
    jishen: '',
    wuxing: { mu: 20, huo: 20, tu: 20, jin: 20, shui: 20 },
    dayGanWuxing: '木',
    tiaohou: null,
  })
  assert.equal(o.prose, '选择一段大运后查看该十年流年节奏。')
  assert.equal(o.trendKeywords, '节奏 · 观察 · 平衡')
})
```

- [ ] **Step 2: 跑测试，确认全部失败**

Run: `cd /Users/liujiming/web/yuanju/frontend && node --experimental-strip-types --test tests/dayun-overview.test.mjs 2>&1 | tail -30`

Expected: 测试加载阶段直接 fail，因为 `../src/lib/dayunOverview.ts` 不存在。报错应包含 `Cannot find module` 或 `ERR_MODULE_NOT_FOUND`。**不要把这个失败视为 bug**，这是 RED 阶段预期的。

- [ ] **Step 3: 提交 RED**

Run:
```bash
git -C /Users/liujiming/web/yuanju add frontend/tests/dayun-overview.test.mjs
git -C /Users/liujiming/web/yuanju commit -m "test(dayun-overview): RED — failing tests for buildDayunOverview

11 tests total: 3 wiring (grep) + 8 behavioral (via --experimental-strip-types).
All fail because src/lib/dayunOverview.ts does not exist yet.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>"
```

---

## Task 2: GREEN — implement dayunOverview.ts

**Files:**
- Create: `frontend/src/lib/dayunOverview.ts`

- [ ] **Step 1: 写完整实现**

Create `frontend/src/lib/dayunOverview.ts` with this exact content:

```typescript
export type Polarity = 'xi' | 'ji' | 'zhong'
type Strength = 'wang' | 'ruo'
type Relation = 'tongGen' | 'gaiTou' | 'jieJiao' | 'none'
type Fit = 'buZu' | 'weiDaoWei' | 'weiJi' | 'skip'

export interface DayunOverviewInput {
  dayun: {
    gan: string
    zhi: string
    gan_shishen: string
    zhi_shishen: string
    di_shi: string
  }
  yongshen: string
  jishen: string
  wuxing: { mu: number; huo: number; tu: number; jin: number; shui: number }
  dayGanWuxing: string
  tiaohou?: {
    expected: string[]
    tou: string[]
    cang: string[]
    text: string
  } | null
}

export interface DayunOverviewOutput {
  prose: string
  trendKeywords: string
  ganPolarity: Polarity
  zhiPolarity: Polarity
}

const FALLBACK_PROSE = '选择一段大运后查看该十年流年节奏。'
const FALLBACK_KEYWORDS = '节奏 · 观察 · 平衡'

const GAN_WUXING_CN: Record<string, string> = {
  甲: '木', 乙: '木', 丙: '火', 丁: '火', 戊: '土',
  己: '土', 庚: '金', 辛: '金', 壬: '水', 癸: '水',
}

const ZHI_MAIN_WUXING: Record<string, string> = {
  子: '水', 丑: '土', 寅: '木', 卯: '木',
  辰: '土', 巳: '火', 午: '火', 未: '土',
  申: '金', 酉: '金', 戌: '土', 亥: '水',
}

const ZHI_MAIN_GAN: Record<string, string> = {
  子: '癸', 丑: '己', 寅: '甲', 卯: '乙',
  辰: '戊', 巳: '丙', 午: '丁', 未: '己',
  申: '庚', 酉: '辛', 戌: '戊', 亥: '壬',
}

const K_GRAPH: Record<string, string> = {
  木: '土', 土: '水', 水: '火', 火: '金', 金: '木',
}

const HELP_MAP: Record<string, Array<'mu' | 'huo' | 'tu' | 'jin' | 'shui'>> = {
  木: ['mu', 'shui'],
  火: ['huo', 'mu'],
  土: ['tu', 'huo'],
  金: ['jin', 'tu'],
  水: ['shui', 'jin'],
}

const DI_SHI_BUCKET: Record<string, 'wang' | 'mid' | 'shuai'> = {
  帝旺: 'wang', 临官: 'wang', 长生: 'wang', 冠带: 'wang',
  沐浴: 'mid',  养: 'mid',     胎: 'mid',   墓: 'mid',
  衰: 'shuai',  病: 'shuai',   死: 'shuai', 绝: 'shuai',
}

const DI_SHI_LABEL: Record<'wang' | 'mid' | 'shuai', string> = {
  wang: '得位有力',
  mid: '态势中等',
  shuai: '气势减弱',
}

const BODY1: Record<string, Record<Strength, string>> = {
  比肩: { wang: '同行竞争分薄资源',     ruo: '兄弟朋友助身有力' },
  劫财: { wang: '损财争夺、合作伤利',   ruo: '同道分担、压力有人共担' },
  食神: { wang: '财源外吐、口腹之享',   ruo: '才华外泄、气力分散' },
  伤官: { wang: '才名突破、敢破规则',   ruo: '才华伤身、易招是非' },
  正财: { wang: '经营得利、稳定积累',   ruo: '财多身弱、力不从心' },
  偏财: { wang: '偏门机会、流动资金',   ruo: '财来财去、难以聚守' },
  正官: { wang: '事业晋升、责任加码',   ruo: '官杀压身、易受规则约束' },
  七杀: { wang: '立威破局、事业突破',   ruo: '身弱遭杀克身，压力与突发事件增多' },
  正印: { wang: '印重身旺反招迟滞',     ruo: '学习/贵人/资格类机会成形' },
  偏印: { wang: '转型旁门、思虑成局',   ruo: '灵感/研究/孤独感提升' },
}

const TREND: Record<string, { xi: string; ji: string }> = {
  比肩: { xi: '同道 · 自立 · 稳进', ji: '分薄 · 竞争 · 节制' },
  劫财: { xi: '合伙 · 协力 · 取舍', ji: '损财 · 争夺 · 化解' },
  食神: { xi: '表达 · 享受 · 作品', ji: '泄气 · 分心 · 节用' },
  伤官: { xi: '突破 · 才名 · 创意', ji: '是非 · 锋芒 · 收敛' },
  正财: { xi: '经营 · 责任 · 积累', ji: '负重 · 守财 · 量力' },
  偏财: { xi: '机会 · 流动 · 人脉', ji: '财去 · 投机 · 谨慎' },
  正官: { xi: '事业 · 晋升 · 成就', ji: '约束 · 规矩 · 顺应' },
  七杀: { xi: '突破 · 立威 · 决断', ji: '压力 · 守势 · 化解' },
  正印: { xi: '学习 · 贵人 · 资质', ji: '迟滞 · 内耗 · 取舍' },
  偏印: { xi: '研究 · 灵感 · 转型', ji: '孤独 · 怀疑 · 沉淀' },
}

const POL_LABEL: Record<Polarity, string> = { xi: '喜', ji: '忌', zhong: '中' }

function resolvePolarity(wuxing: string, yong: string, ji: string): Polarity {
  if (!wuxing) return 'zhong'
  if (yong && yong.includes(wuxing)) return 'xi'
  if (ji && ji.includes(wuxing)) return 'ji'
  return 'zhong'
}

function resolveDayStrength(
  wuxing: DayunOverviewInput['wuxing'] | undefined,
  dayGanWuxing: string,
): Strength {
  if (!wuxing) return 'ruo'
  const help = HELP_MAP[dayGanWuxing] ?? []
  const helpPct = help.reduce((s, k) => s + (wuxing[k] ?? 0), 0)
  return helpPct > 40 ? 'wang' : 'ruo'
}

function resolveGanZhiRelation(ganWx: string, zhiWx: string): Relation {
  if (!ganWx || !zhiWx) return 'none'
  if (ganWx === zhiWx) return 'tongGen'
  if (K_GRAPH[ganWx] === zhiWx) return 'gaiTou'
  if (K_GRAPH[zhiWx] === ganWx) return 'jieJiao'
  return 'none'
}

function resolveTiaohouFit(
  input: DayunOverviewInput,
  relation: Relation,
): { fit: Fit; missingWx?: string; matchedGan?: string; coverGan?: string } {
  const t = input.tiaohou
  if (!t || !t.expected || t.expected.length === 0) return { fit: 'skip' }

  const have = new Set([...(t.tou ?? []), ...(t.cang ?? [])])
  const missingGans = t.expected.filter(g => !have.has(g))
  if (missingGans.length === 0) return { fit: 'skip' }

  const missingWxSet = new Set(
    missingGans.map(g => GAN_WUXING_CN[g]).filter(Boolean),
  )
  if (missingWxSet.size === 0) return { fit: 'skip' }

  const ganWx = GAN_WUXING_CN[input.dayun.gan]
  const zhiMainWx = ZHI_MAIN_WUXING[input.dayun.zhi]
  const ganMatches = missingWxSet.has(ganWx)
  const zhiMatches = missingWxSet.has(zhiMainWx)

  if (!ganMatches && !zhiMatches) {
    return { fit: 'weiJi', missingWx: GAN_WUXING_CN[missingGans[0]] }
  }

  let matchedGan: string
  let matchedWx: string
  if (ganMatches) {
    matchedGan = input.dayun.gan
    matchedWx = ganWx
  } else {
    matchedGan = ZHI_MAIN_GAN[input.dayun.zhi] ?? input.dayun.zhi
    matchedWx = zhiMainWx
  }

  if (relation === 'gaiTou') {
    return { fit: 'weiDaoWei', missingWx: matchedWx, coverGan: input.dayun.gan }
  }
  return { fit: 'buZu', missingWx: matchedWx, matchedGan }
}

function body2(diShi: string, gan: string, zhi: string, relation: Relation): string {
  const bucket = DI_SHI_BUCKET[diShi]
  if (!bucket) return ''
  let base = `${zhi}${diShi}${DI_SHI_LABEL[bucket]}`
  switch (relation) {
    case 'tongGen': base += `，${gan}通根${zhi}得力`; break
    case 'gaiTou':  base += `，但被${gan}盖头压制`; break
    case 'jieJiao': base += `，反被${zhi}截脚虚浮`; break
    case 'none': break
  }
  return base
}

function body3(
  fit: Fit,
  missingWx?: string,
  matchedGan?: string,
  coverGan?: string,
): string {
  switch (fit) {
    case 'buZu':
      return `${matchedGan ?? ''}${missingWx ?? ''}透出，正补足命局所缺调候`
    case 'weiDaoWei':
      return `命局所需${missingWx ?? ''}虽现于运中，却被${coverGan ?? ''}压制，调候未到位`
    case 'weiJi':
      return `命局所缺${missingWx ?? ''}未在此运补足，需外接调候助力`
    case 'skip':
      return ''
  }
}

export function buildDayunOverview(input: DayunOverviewInput): DayunOverviewOutput {
  const { dayun } = input
  const ganWx = GAN_WUXING_CN[dayun.gan]
  const zhiWx = ZHI_MAIN_WUXING[dayun.zhi]
  const inBody1 = !!BODY1[dayun.gan_shishen]
  const inBucket = !!DI_SHI_BUCKET[dayun.di_shi]

  if (!ganWx || !zhiWx || !inBody1 || !inBucket) {
    return {
      prose: FALLBACK_PROSE,
      trendKeywords: FALLBACK_KEYWORDS,
      ganPolarity: 'zhong',
      zhiPolarity: 'zhong',
    }
  }

  const ganPolarity = resolvePolarity(ganWx, input.yongshen, input.jishen)
  const zhiPolarity = resolvePolarity(zhiWx, input.yongshen, input.jishen)
  const dayStrength = resolveDayStrength(input.wuxing, input.dayGanWuxing)
  const relation = resolveGanZhiRelation(ganWx, zhiWx)
  const { fit, missingWx, matchedGan, coverGan } = resolveTiaohouFit(input, relation)

  const hasPolarity = !!input.yongshen || !!input.jishen
  const heading = hasPolarity
    ? `${dayun.gan}${dayun.zhi}运（${dayun.gan_shishen}为${POL_LABEL[ganPolarity]}·${dayun.zhi_shishen}为${POL_LABEL[zhiPolarity]}）：`
    : `${dayun.gan}${dayun.zhi}运：`

  const body1 = BODY1[dayun.gan_shishen][dayStrength]
  const body2text = body2(dayun.di_shi, dayun.gan, dayun.zhi, relation)
  const body3text = body3(fit, missingWx, matchedGan, coverGan)

  let prose = `${heading}${body1}；${body2text}。`
  if (body3text) prose += `${body3text}。`

  const trendEntry = TREND[dayun.gan_shishen]
  const trendKeywords = trendEntry
    ? (ganPolarity === 'ji' ? trendEntry.ji : trendEntry.xi)
    : FALLBACK_KEYWORDS

  return { prose, trendKeywords, ganPolarity, zhiPolarity }
}
```

- [ ] **Step 2: 跑测试，确认 8 个行为单测全 PASS**

Run: `cd /Users/liujiming/web/yuanju/frontend && node --experimental-strip-types --test tests/dayun-overview.test.mjs 2>&1 | tail -20`

Expected:
- 3 个静态接线测试仍 FAIL（DayunTimeline / ResultPage 还没改）
- 8 个行为单测全 PASS
- 类似输出：`tests 11`, `pass 8`, `fail 3`

如果任何**行为**单测 fail，**不要去改测试断言** — 检查 dayunOverview.ts 中对应字典的中文字面与测试断言是否一致。Spec 是源真实，按 spec 调整 lib，而非 tweak 测试。

- [ ] **Step 3: tsc 类型检查**

Run: `cd /Users/liujiming/web/yuanju/frontend && npx tsc --noEmit 2>&1 | tail -10`

Expected: 0 errors

- [ ] **Step 4: 提交 GREEN-1**

Run:
```bash
git -C /Users/liujiming/web/yuanju add frontend/src/lib/dayunOverview.ts
git -C /Users/liujiming/web/yuanju commit -m "feat(dayun-overview): add deterministic personalization library

7 lookup tables + 4 helpers + buildDayunOverview() main entry.
Drives the per-dayun overview prose and trend chip from
yongshen/jishen/wuxing/tiaohou.

Spec: docs/superpowers/specs/2026-05-18-dayun-overview-personalization-design.md

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>"
```

---

## Task 3: GREEN — wire DayunTimeline.tsx

**Files:**
- Modify: `frontend/src/components/DayunTimeline.tsx`

- [ ] **Step 1: 删除旧硬编码常量**

在 `DayunTimeline.tsx` 中删除以下三段（行号以当前 HEAD 为准，按字面字串匹配）：

删除 `TREND_KEYWORDS` 字典（line 63-74，10 条十神条目）：
```typescript
const TREND_KEYWORDS: Record<string, string> = {
  比肩: '自我 · 同行 · 稳定',
  劫财: '竞争 · 合伙 · 取舍',
  食神: '表达 · 作品 · 享受',
  伤官: '突破 · 创意 · 表达',
  正财: '经营 · 积累 · 责任',
  偏财: '机会 · 流动 · 人脉',
  正官: '事业 · 责任 · 成就',
  七杀: '压力 · 行动 · 突破',
  正印: '学习 · 贵人 · 资质',
  偏印: '研究 · 灵感 · 转型',
}
```

删除 `getTrendKeywords` 函数（line 101-104）：
```typescript
function getTrendKeywords(dayun?: DayunItem) {
  if (!dayun) return '节奏 · 观察 · 平衡'
  return TREND_KEYWORDS[dayun.gan_shishen] || TREND_KEYWORDS[dayun.zhi_shishen] || '节奏 · 观察 · 平衡'
}
```

删除 `getDayunSummary` 函数（line 106-109）：
```typescript
function getDayunSummary(dayun?: DayunItem) {
  if (!dayun) return '选择一段大运后查看该十年流年节奏。'
  return `${dayun.gan}${dayun.zhi}运以${dayun.gan_shishen}透干、${dayun.zhi_shishen}坐支为主，${dayun.di_shi}之势宜先看节奏，再看流年触发点。`
}
```

- [ ] **Step 2: 加 import**

在文件顶部 import 块新增一行（紧跟现有 `import { fetchShenshaAnnotations, ...` 之后）：

```typescript
import { buildDayunOverview } from '../lib/dayunOverview'
```

- [ ] **Step 3: 扩展 `DayunTimelineProps`**

将 line 40-48 的 `DayunTimelineProps` 接口替换为：

```typescript
interface DayunTimelineProps {
  dayun: DayunItem[]
  birthYear: number
  startYunSolar: string
  dayGan: string
  gender?: string
  pillarsLabel?: string
  chartId?: string
  yongshen?: string
  jishen?: string
  wuxing?: { mu: number; huo: number; tu: number; jin: number; shui: number }
  tiaohou?: { expected: string[]; tou: string[]; cang: string[]; text: string } | null
}
```

- [ ] **Step 4: 扩展函数签名 destructuring**

将 line 111（`export default function DayunTimeline({ ... }: DayunTimelineProps)`）替换为：

```typescript
export default function DayunTimeline({
  dayun, birthYear, startYunSolar, dayGan, gender, pillarsLabel, chartId,
  yongshen, jishen, wuxing, tiaohou,
}: DayunTimelineProps) {
```

- [ ] **Step 5: 改 summary strip 的渲染调用**

定位 line 294-306 的 `{activeDayun && (...)}` 块（关键字 "大运总览"）。把整块替换为：

```tsx
{activeDayun && (() => {
  const dayGanWx = GAN_WUXING[dayGan] ?? ''
  const dayGanWxCn = WUXING_LABEL[dayGanWx] ?? ''
  const overview = buildDayunOverview({
    dayun: {
      gan: activeDayun.gan,
      zhi: activeDayun.zhi,
      gan_shishen: activeDayun.gan_shishen,
      zhi_shishen: activeDayun.zhi_shishen,
      di_shi: activeDayun.di_shi,
    },
    yongshen: yongshen ?? '',
    jishen: jishen ?? '',
    wuxing: wuxing ?? { mu: 20, huo: 20, tu: 20, jin: 20, shui: 20 },
    dayGanWuxing: dayGanWxCn,
    tiaohou: tiaohou ?? null,
  })
  return (
    <div className="dayun-summary-strip">
      <div className="dayun-summary-copy">
        <strong>大运总览</strong>
        <span>{overview.prose}</span>
      </div>
      <div className="dayun-summary-tags">
        <span>十神主气：{activeDayun.gan_shishen}</span>
        <span>五行主气：{WUXING_LABEL[GAN_WUXING[activeDayun.gan] || 'jin']}</span>
        <span>趋势关键词：{overview.trendKeywords}</span>
      </div>
    </div>
  )
})()}
```

注：`dayGanWxCn` 把 `GAN_WUXING` 输出的英文短码（`'mu'/'huo'/...`）通过 `WUXING_LABEL` 翻译成中文 1 字（`'木'/'火'/...`），这是 `buildDayunOverview` `dayGanWuxing` 字段要求的格式。

- [ ] **Step 6: 更新既有 `dayun-timeline-ux.test.mjs` 中受影响的断言**

`frontend/tests/dayun-timeline-ux.test.mjs:99` 当前断言：
```js
assert.match(component, /getTrendKeywords/)
```
此函数本任务已删，将该行替换为：
```js
assert.match(component, /buildDayunOverview/)
```

不动其他断言（dayun-summary-strip / -copy / -tags 类名仍保留，ZHI_MAIN_GAN 仅出现在新 lib 而非 DayunTimeline）。

- [ ] **Step 7: 跑 tsc 类型检查**

Run: `cd /Users/liujiming/web/yuanju/frontend && npx tsc --noEmit 2>&1 | tail -10`

Expected: 0 errors

- [ ] **Step 8: 跑 dayun-overview 测试，确认 2 个 DayunTimeline 接线测试 PASS**

Run: `cd /Users/liujiming/web/yuanju/frontend && node --experimental-strip-types --test tests/dayun-overview.test.mjs 2>&1 | tail -10`

Expected: pass 10, fail 1（只剩 ResultPage props 测试 fail）

- [ ] **Step 9: 跑 dayun-timeline-ux 测试，确认更新后通过**

Run: `cd /Users/liujiming/web/yuanju/frontend && node --test tests/dayun-timeline-ux.test.mjs 2>&1 | tail -10`

Expected: 全 PASS

- [ ] **Step 10: 提交 GREEN-2**

Run:
```bash
git -C /Users/liujiming/web/yuanju add frontend/src/components/DayunTimeline.tsx frontend/tests/dayun-timeline-ux.test.mjs
git -C /Users/liujiming/web/yuanju commit -m "feat(dayun-timeline): replace hardcoded overview with buildDayunOverview

Drop TREND_KEYWORDS / getTrendKeywords / getDayunSummary hardcoded
helpers; route through new dayunOverview lib. Add 4 optional props
(yongshen, jishen, wuxing, tiaohou) wired to per-dayun computation.
Update dayun-timeline-ux test to grep for the new symbol.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>"
```

---

## Task 4: GREEN — wire ResultPage.tsx

**Files:**
- Modify: `frontend/src/pages/ResultPage.tsx`

- [ ] **Step 1: 给 `<DayunTimeline>` 加 4 个 prop**

定位 line 896-904 的 `<DayunTimeline ... />` 元素。把它替换为：

```tsx
<DayunTimeline
  dayun={result.dayun}
  birthYear={result.birth_year}
  startYunSolar={result.start_yun_solar}
  dayGan={result.day_gan || ''}
  gender={result.gender}
  pillarsLabel={dayunPillarsLabel}
  chartId={targetId}
  yongshen={result.yongshen || ''}
  jishen={result.jishen || ''}
  wuxing={result.wuxing}
  tiaohou={result.tiaohou ?? null}
/>
```

- [ ] **Step 2: 跑 tsc 类型检查**

Run: `cd /Users/liujiming/web/yuanju/frontend && npx tsc --noEmit 2>&1 | tail -10`

Expected: 0 errors

- [ ] **Step 3: 跑测试，所有 11 个全 PASS**

Run: `cd /Users/liujiming/web/yuanju/frontend && node --experimental-strip-types --test tests/dayun-overview.test.mjs 2>&1 | tail -15`

Expected: `tests 11`, `pass 11`, `fail 0`

- [ ] **Step 4: 提交 GREEN-3**

Run:
```bash
git -C /Users/liujiming/web/yuanju add frontend/src/pages/ResultPage.tsx
git -C /Users/liujiming/web/yuanju commit -m "feat(result-page): pass yongshen/jishen/wuxing/tiaohou to DayunTimeline

Plumb the four chart-level signals needed for per-dayun overview
personalization. result.tiaohou already has the {expected, tou, cang,
text} shape consumed by the new lib.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>"
```

---

## Task 5: Final verification

**Files:** none

- [ ] **Step 1: 全套 lint**

Run: `cd /Users/liujiming/web/yuanju/frontend && npm run lint 2>&1 | tail -30`

Expected: 0 errors。预存的 2 个 `no-irregular-whitespace` 警告在 `PrintLayout.tsx` 来自历史 commit `29155258`，不要试图修。

- [ ] **Step 2: 全套 build**

Run: `cd /Users/liujiming/web/yuanju/frontend && npm run build 2>&1 | tail -15`

Expected: `built in <X>s`，0 tsc 错误。

- [ ] **Step 3: 全套测试**

Run: `cd /Users/liujiming/web/yuanju/frontend && node --experimental-strip-types --test tests/dayun-overview.test.mjs 2>&1 | tail -15`

Expected: `tests 11`, `pass 11`, `fail 0`

- [ ] **Step 4: 跑一遍其他既有测试，确保没有 collateral**

Run: `cd /Users/liujiming/web/yuanju/frontend && node --test tests/brand-settings.test.mjs tests/brand-settings-dark-theme.test.mjs tests/dayun-timeline-ux.test.mjs tests/ten-god-relation-ux.test.mjs 2>&1 | tail -10`

Expected: 全 PASS。`dayun-timeline-ux.test.mjs` 的 `getTrendKeywords → buildDayunOverview` 替换已在 Task 3 Step 6 完成，此处应原样通过。

如果有其他**未在 Task 3 提到过的** grep 测试因符号删除而 fail，停下来反馈给 controller，**不要静默改测试断言** —— 可能是 spec 漏覆盖。

- [ ] **Step 5: 查看 git log 确认 4 个 commit 都在**

Run: `git -C /Users/liujiming/web/yuanju log --oneline -6`

Expected 顶部:
```
<sha> feat(result-page): pass yongshen/jishen/wuxing/tiaohou to DayunTimeline
<sha> feat(dayun-timeline): replace hardcoded overview with buildDayunOverview
<sha> feat(dayun-overview): add deterministic personalization library
<sha> test(dayun-overview): RED — failing tests for buildDayunOverview
48f9057 docs: correct tiaohou field shape in dayun overview spec
0ddbcc6 docs: brainstorm spec for dayun overview personalization
```

---

## Task 6: Manual acceptance (user-run)

**Files:** none — 由用户在浏览器人工验证

- [ ] **Step 1: 重启前端 dev server**

由用户执行 `cd frontend && npm run dev`（或 `docker compose up -d --build frontend`）。

- [ ] **Step 2: 视觉验收清单**

用户按以下顺序检查：

1. **已生成报告的 chart**：
   - 打开任一已经生成过 AI 报告的 `chartId`
   - 滚到大运时间轴，点不同的大运卡片
   - 总览 3 句应**立刻刷新且各不相同** —— 同一个十神（如多个大运都是七杀）会因身强弱/喜忌/调候不同而产生不同表述
   - 旁边「趋势关键词」chip 应跟着切，喜忌不同时显示不同短语
   - 「十神主气」「五行主气」chip 保持不变（本次不动）

2. **未生成报告的 chart**：
   - 用 `isGuest = true` 或刚算完八字未点报告的 `chartId`
   - 总览**应仍能出文字**（2 句无 polarity 版本，比如「壬午运：身弱遭杀克身，压力与突发事件增多；午帝旺得位有力，但被壬盖头压制。」）
   - chip「趋势关键词」此时取 xi 兜底（无 polarity）

3. **导出场景**：
   - 点「PDF 导出」+「PNG 分享图」，看 print/share UI 不破版
   - 本次未动 CSS 与 PrintLayout，理论上零影响 —— 看一眼确认

- [ ] **Step 3: 若发现文案不对，反馈给开发**

对照 spec 文档 `docs/superpowers/specs/2026-05-18-dayun-overview-personalization-design.md` 的「BODY1 查表」「BODY2」「BODY3」三段，定位是哪一个槽位的措辞需要 tweak，单独提一笔 commit 微调。**不要因为文案不爽就推翻整个公式**。

---

## YAGNI 边界（提醒实施者）

实施过程**不做**以下任何一项（spec 已明确排除，subagent 看到就停手）：
- 与命局四支的冲合刑害判定
- 关键流年识别
- AI 长版本接入
- 十神主气 / 五行主气 chip 个人化
- 后端任何修改
- 修复 `PrintLayout.tsx` U+3000 历史警告
- 重构 `DayunTimeline.tsx` 其他部分

如果在实施中**发现 spec 与现有代码冲突无法继续**（比如 `result.wuxing` 字段实际不叫这个），停下来反馈给 controller，**不要自作主张改 spec**。
