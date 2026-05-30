# 六十甲子 · 日柱速写 实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: 用 superpowers:executing-plans（推荐，本计划含交互式文案撰写，需逐批与用户校准）逐任务实现。步骤用 `- [ ]` 勾选跟踪。

**Goal:** 在合盘结果页 / PDF / 分享图三处，按每人日柱干支展示一段静态「一柱速写」，与现有算法完全解耦。

**Architecture:** 一张静态 `Record<干支, {tag,text}>` + 单一查表函数 `getDayPillarPortrait`，三处渲染各自只读命盘快照的 `day_gan/day_zhi` 来查表。不碰 `compatibilityPersonality.ts`、评分、后端、LLM。

**Tech Stack:** React 19 + TypeScript + Vite；测试 `node --test`（`frontend/tests/*.test.mjs`，按既有约定以源码文本断言）。

**调性（锁定）：** 双关微辣、暗示拉满、不露骨。详见设计文档 `docs/superpowers/specs/2026-05-30-sixty-jiazi-day-pillar-portrait-design.md`。

**关于文案任务的说明（重要）：** Task 4 是 60 条速写的撰写，按用户要求**分 6 批 × 10 柱、每批由用户过目校准**。因此这部分**不在计划里预先写死 50 条**（那会绕过用户的逐批审稿意图），计划只给出撰写规范 + 已校准的 3 条样例；实际文案在执行时由执行者与用户交互产出。其余任务（数据文件骨架、查表函数、组件、三处挂载、测试）均为可测代码，给出完整代码。

---

## Task 1: 数据文件骨架 + 查表函数（含 3 条已校准样例）

**Files:**
- Create: `frontend/src/lib/dayPillarPortraits.ts`

- [ ] **Step 1: 写数据文件骨架**

```ts
// frontend/src/lib/dayPillarPortraits.ts
// 六十甲子 · 日柱速写（静态附加层）。按日柱干支查一段性格/相处风格速写。
// 与 compatibilityPersonality.ts / 合盘评分 / LLM 完全解耦，仅供展示。
// 调性：双关微辣、不露骨。文案逐批撰写，见实现计划 Task 4。

export type DayPillarPortrait = {
  tag: string // 4-6 字定性钩子
  text: string // 速写正文（2-3 句）
}

const DAY_PILLAR_PORTRAITS: Record<string, DayPillarPortrait> = {
  甲子: {
    tag: '有耐心的狼',
    text:
      '白天一身正气、规矩得能领面锦旗，关了灯完全两个人。子水是沐浴之地——外头道貌岸然、里头野。慢热却记仇式专一，是只“有耐心的狼”，不动则已，一动就盯死你，耐力还出奇地好。',
  },
  乙丑: {
    tag: '冷面撩人',
    text:
      '长发一甩、腰肢一摆，天生勾人。丑是官杀库，女命坐下藏着一整支“后备队”——不是她招蜂引蝶，是异性自己排队上门；偏财一掺，浪漫起来不计成本。男命则是闷声撩人的惯犯，话不多、手不闲，冷着脸就把人带走了。',
  },
  壬申: {
    tag: '供大于求',
    text:
      '自坐长生，一眼活泉，越掏越有、越用越旺。精力、情绪、还有别的，统统源源不断，是“供大于求”型选手——鲜活、耐折腾，就是太满，得配个接得住的。',
  },
}

// 查表：命中返回速写，未知干支返回 undefined（旧数据/缺字段时由调用方跳过渲染）
export function getDayPillarPortrait(dayGan: string, dayZhi: string): DayPillarPortrait | undefined {
  return DAY_PILLAR_PORTRAITS[`${dayGan}${dayZhi}`]
}
```

- [ ] **Step 2: 类型检查通过**

Run: `cd frontend && npx tsc --noEmit`
Expected: 无新增报错（新文件自洽）。

- [ ] **Step 3: Commit**

```bash
git add frontend/src/lib/dayPillarPortraits.ts
git commit -m "feat(compat): scaffold day-pillar portrait lookup with 3 calibrated entries"
```

---

## Task 2: 查表函数 + 解耦测试（先红后绿）

**Files:**
- Create: `frontend/tests/day-pillar-portrait.test.mjs`

- [ ] **Step 1: 写测试**

```js
// frontend/tests/day-pillar-portrait.test.mjs
import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import test from 'node:test'

const root = resolve(dirname(fileURLToPath(import.meta.url)), '..')
const read = (p) => readFileSync(resolve(root, p), 'utf8')

// 生成标准六十甲子
const GAN = ['甲', '乙', '丙', '丁', '戊', '己', '庚', '辛', '壬', '癸']
const ZHI = ['子', '丑', '寅', '卯', '辰', '巳', '午', '未', '申', '酉', '戌', '亥']
const SIXTY = Array.from({ length: 60 }, (_, i) => GAN[i % 10] + ZHI[i % 12])

test('data file exposes the lookup type and getter', () => {
  const src = read('src/lib/dayPillarPortraits.ts')
  assert.match(src, /export type DayPillarPortrait/)
  assert.match(src, /export function getDayPillarPortrait\(dayGan: string, dayZhi: string\)/)
})

test('present entries have non-empty tag and text', () => {
  const src = read('src/lib/dayPillarPortraits.ts')
  // 不允许空串占位
  assert.doesNotMatch(src, /tag:\s*''/)
  assert.doesNotMatch(src, /text:\s*''/)
})

test('all 60 jiazi keys are present (completeness gate — green only after Task 4)', () => {
  const src = read('src/lib/dayPillarPortraits.ts')
  const missing = SIXTY.filter((gz) => !new RegExp(`(?:^|[^甲-癸])${gz}:`).test(src))
  assert.deepEqual(missing, [], `缺失干支: ${missing.join(' ')}`)
})

test('day-pillar layer stays decoupled from the personality engine', () => {
  const lib = read('src/lib/dayPillarPortraits.ts')
  assert.doesNotMatch(lib, /compatibilityPersonality/)
  const comp = read('src/components/compatibility/DayPillarPortrait.tsx')
  assert.doesNotMatch(comp, /compatibilityPersonality/)
  assert.match(comp, /getDayPillarPortrait/)
})

test('screen / PDF / share-image all render via the shared lookup', () => {
  assert.match(read('src/pages/CompatibilityResultPage.tsx'), /<DayPillarPortrait/)
  assert.match(read('src/components/CompatibilityPrintLayout.tsx'), /getDayPillarPortrait/)
  assert.match(read('src/components/CompatibilityShareCard.tsx'), /getDayPillarPortrait/)
})
```

- [ ] **Step 2: 跑测试看红**

Run: `cd frontend && node --test tests/day-pillar-portrait.test.mjs`
Expected: 多条 FAIL（DayPillarPortrait.tsx 不存在、三处挂载未加、60 柱未齐）。这是预期的红。

- [ ] **Step 3: Commit**

```bash
git add frontend/tests/day-pillar-portrait.test.mjs
git commit -m "test(compat): add day-pillar portrait completeness + decoupling tests"
```

---

## Task 3: 屏幕端组件 + 挂载

**Files:**
- Create: `frontend/src/components/compatibility/DayPillarPortrait.tsx`
- Create: `frontend/src/components/compatibility/DayPillarPortrait.css`
- Modify: `frontend/src/pages/CompatibilityResultPage.tsx`（import + 在 `<SectionBasicCharts/>` 后挂载）

- [ ] **Step 1: 写组件**

```tsx
// frontend/src/components/compatibility/DayPillarPortrait.tsx
import './DayPillarPortrait.css'
import type { CompatibilityParticipant } from '../../lib/api'
import { getDayPillarPortrait } from '../../lib/dayPillarPortraits'

function PortraitCard({ participant }: { participant?: CompatibilityParticipant | null }) {
  const snap = participant?.chart_snapshot
  const portrait = snap ? getDayPillarPortrait(snap.day_gan, snap.day_zhi) : undefined
  if (!snap || !portrait) return null
  return (
    <div className="day-pillar-card">
      <div className="day-pillar-card-head">
        <span className="day-pillar-name">{participant?.display_name || '—'}</span>
        <span className="day-pillar-ganzhi">{snap.day_gan}{snap.day_zhi}日</span>
      </div>
      <div className="day-pillar-tag">{portrait.tag}</div>
      <p className="day-pillar-text">{portrait.text}</p>
    </div>
  )
}

export default function DayPillarPortrait({
  self,
  partner,
}: {
  self?: CompatibilityParticipant | null
  partner?: CompatibilityParticipant | null
}) {
  const has = (p?: CompatibilityParticipant | null) =>
    Boolean(p?.chart_snapshot && getDayPillarPortrait(p.chart_snapshot.day_gan, p.chart_snapshot.day_zhi))
  if (!has(self) && !has(partner)) return null
  return (
    <section className="day-pillar-portrait">
      <div className="day-pillar-portrait-title serif">日柱 · 本命速写</div>
      <div className="day-pillar-grid">
        <PortraitCard participant={self} />
        <PortraitCard participant={partner} />
      </div>
    </section>
  )
}
```

- [ ] **Step 2: 写样式**

```css
/* frontend/src/components/compatibility/DayPillarPortrait.css */
.day-pillar-portrait {
  margin: 16px 0;
}
.day-pillar-portrait-title {
  font-size: 16px;
  font-weight: 700;
  color: #5a3a1a;
  margin-bottom: 10px;
}
.day-pillar-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px;
}
.day-pillar-card {
  border: 1px solid #e0cca0;
  border-radius: 8px;
  padding: 12px 14px;
  background: #fdf9f2;
}
.day-pillar-card-head {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: 8px;
  margin-bottom: 6px;
}
.day-pillar-name {
  font-weight: 700;
  color: #4a3728;
}
.day-pillar-ganzhi {
  font-family: "Noto Serif SC", serif;
  font-size: 13px;
  color: #9b815c;
}
.day-pillar-tag {
  display: inline-block;
  font-size: 13px;
  font-weight: 700;
  color: #7a5c3a;
  background: #f5ebd6;
  border-radius: 4px;
  padding: 2px 8px;
  margin-bottom: 6px;
}
.day-pillar-text {
  margin: 0;
  font-size: 14px;
  line-height: 1.7;
  color: #4a3728;
}
@media (max-width: 560px) {
  .day-pillar-grid { grid-template-columns: 1fr; }
}
```

- [ ] **Step 3: 结果页挂载**

在 `frontend/src/pages/CompatibilityResultPage.tsx` 顶部 import 区加：
```tsx
import DayPillarPortrait from '../components/compatibility/DayPillarPortrait'
```
在 `<SectionBasicCharts self={selfP || null} partner={partnerP || null} />` 后紧接一行：
```tsx
        <DayPillarPortrait self={selfP || null} partner={partnerP || null} />
```

- [ ] **Step 4: 类型检查 + 跑解耦/挂载测试**

Run: `cd frontend && npx tsc --noEmit && node --test tests/day-pillar-portrait.test.mjs`
Expected: 解耦测试、屏幕挂载断言转绿；PDF/分享图 `getDayPillarPortrait` 断言与「60 柱齐全」仍红（后续任务）。

- [ ] **Step 5: Commit**

```bash
git add frontend/src/components/compatibility/DayPillarPortrait.tsx frontend/src/components/compatibility/DayPillarPortrait.css frontend/src/pages/CompatibilityResultPage.tsx
git commit -m "feat(compat): render day-pillar portrait on result page"
```

---

## Task 4: 撰写 60 条速写（6 批 × 10，逐批用户校准）

**Files:**
- Modify: `frontend/src/lib/dayPillarPortraits.ts`（逐批补全条目）

**撰写规范（每条都遵守）：**
- 调性：双关微辣、暗示拉满、不露骨。参照 Task 1 三条样例的尺度。
- 取象依据：干为表/外、支为里/内 + 日柱十神 + 长生十二宫 + 纳音意象 + 日支刑冲。不照搬露骨原文。
- 结构：`tag` 4-6 字定性钩子；`text` 2-3 句。
- 批次顺序（六十甲子标准序）：
  - 批1 已含 3 条（甲子/乙丑/壬申），补齐：丙寅 丁卯 戊辰 己巳 庚午 辛未 癸酉
  - 批2：甲戌 乙亥 丙子 丁丑 戊寅 己卯 庚辰 辛巳 壬午 癸未
  - 批3：甲申 乙酉 丙戌 丁亥 戊子 己丑 庚寅 辛卯 壬辰 癸巳
  - 批4：甲午 乙未 丙申 丁酉 戊戌 己亥 庚子 辛丑 壬寅 癸卯
  - 批5：甲辰 乙巳 丙午 丁未 戊申 己酉 庚戌 辛亥 壬子 癸丑
  - 批6：甲寅 乙卯 丙辰 丁巳 戊午 己未 庚申 辛酉 壬戌 癸亥

- [ ] **Step 1～6: 逐批撰写并请用户过目**

每批：把 10 条加进 `DAY_PILLAR_PORTRAITS`，贴给用户校准；用户认可后进入下一批。

- [ ] **Step 7: 全 60 条齐全后跑完整测试**

Run: `cd frontend && node --test tests/day-pillar-portrait.test.mjs`
Expected: 全绿（含「60 柱齐全」完整性门）。

- [ ] **Step 8: Commit**

```bash
git add frontend/src/lib/dayPillarPortraits.ts
git commit -m "feat(compat): author all 60 day-pillar portraits"
```

---

## Task 5: PDF（CompatibilityPrintLayout）内渲染

**Files:**
- Modify: `frontend/src/components/CompatibilityPrintLayout.tsx`
- Modify: `frontend/src/components/CompatibilityPrintLayout.css`

放在「§一、合参概要」末尾（`ParticipantsHero` 之后），不打乱后续编号。

- [ ] **Step 1: 加 import**

```tsx
import { getDayPillarPortrait } from '../lib/dayPillarPortraits'
```

- [ ] **Step 2: 加一个打印用小组件**（放在 `PersonalityPrint` 同区域附近）

```tsx
function DayPillarPrint({ selfP, partnerP }: {
  selfP?: CompatibilityParticipant
  partnerP?: CompatibilityParticipant
}) {
  const cards = [
    { label: '我', p: selfP },
    { label: '伴侣', p: partnerP },
  ]
    .map(({ label, p }) => {
      const snap = p?.chart_snapshot
      const portrait = snap ? getDayPillarPortrait(snap.day_gan, snap.day_zhi) : undefined
      return snap && portrait ? { name: p?.display_name || label, snap, portrait } : null
    })
    .filter((x): x is NonNullable<typeof x> => x !== null)
  if (cards.length === 0) return null
  return (
    <div className="compat-print-chapter">
      <h4 className="compat-print-chapter-title">本命日柱速写</h4>
      <div className="compat-print-daypillar-grid">
        {cards.map((c) => (
          <div key={c.name} className="compat-print-daypillar">
            <div className="compat-print-daypillar-head">
              <span className="compat-print-daypillar-name">{c.name}</span>
              <span className="compat-print-daypillar-gz">{c.snap.day_gan}{c.snap.day_zhi}日</span>
            </div>
            <div className="compat-print-daypillar-tag">{c.portrait.tag}</div>
            <p className="compat-print-daypillar-text">{c.portrait.text}</p>
          </div>
        ))}
      </div>
    </div>
  )
}
```
> 注：`CompatibilityParticipant` 已在该文件类型 import 中（line 1-10 区块）。若未含，补进 import。

- [ ] **Step 3: 在 §一 内挂载**

在 `<ParticipantsHero participants={participants} reading={reading} />` 之后、`</section>` 之前加：
```tsx
            <DayPillarPrint selfP={selfP} partnerP={partnerP} />
```

- [ ] **Step 4: 加打印样式**（`@media print` 块内）

```css
  /* §一 本命日柱速写 */
  .compat-print-daypillar-grid {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 6mm;
    margin-top: 3mm;
  }
  .compat-print-daypillar {
    border: 1px solid #e0cca0;
    border-radius: 4pt;
    padding: 3mm 4mm;
  }
  .compat-print-daypillar-head {
    display: flex;
    justify-content: space-between;
    align-items: baseline;
    margin-bottom: 1.5mm;
  }
  .compat-print-daypillar-name { font-size: 12pt; font-weight: 700; color: #5a3a1a; }
  .compat-print-daypillar-gz { font-size: 10pt; color: #9b815c; }
  .compat-print-daypillar-tag {
    display: inline-block;
    font-size: 10pt;
    font-weight: 700;
    color: #7a5c3a;
    background: #f5ebd6;
    border-radius: 2pt;
    padding: 0.5mm 2mm;
    margin-bottom: 1.5mm;
  }
  .compat-print-daypillar-text { margin: 0; font-size: 10.5pt; line-height: 1.7; color: #3a2a1a; }
```

- [ ] **Step 5: 类型检查 + 测试**

Run: `cd frontend && npx tsc --noEmit && node --test tests/day-pillar-portrait.test.mjs`
Expected: PDF 的 `getDayPillarPortrait` 断言转绿。

- [ ] **Step 6: Commit**

```bash
git add frontend/src/components/CompatibilityPrintLayout.tsx frontend/src/components/CompatibilityPrintLayout.css
git commit -m "feat(compat): add day-pillar portrait to PDF print layout"
```

---

## Task 6: 分享图（CompatibilityShareCard）内渲染

**Files:**
- Modify: `frontend/src/components/CompatibilityShareCard.tsx`
- Modify: `frontend/src/components/CompatibilityShareCard.css`

加一节 `◇ 日柱速写`，放在 `◇ 双方性格` 节之后。

- [ ] **Step 1: 加 import**

```tsx
import { getDayPillarPortrait } from '../lib/dayPillarPortraits'
```

- [ ] **Step 2: 计算两人速写**（在组件内 `personalityCols` 计算附近）

```tsx
  const dayPillarCols = [selfP, partnerP]
    .map((p) => {
      const snap = p?.chart_snapshot
      const portrait = snap ? getDayPillarPortrait(snap.day_gan, snap.day_zhi) : undefined
      return snap && portrait ? { name: p?.display_name || '—', gz: `${snap.day_gan}${snap.day_zhi}`, portrait } : null
    })
    .filter((x): x is NonNullable<typeof x> => x !== null)
```

- [ ] **Step 3: 渲染该节**（紧接 `◇ 双方性格` 那段 JSX 之后）

```tsx
      {dayPillarCols.length > 0 && (
        <section className="compat-share-daypillar">
          <h3 className="compat-share-section-h">◇ 日柱速写</h3>
          <div className="compat-share-daypillar-grid">
            {dayPillarCols.map((c) => (
              <div key={c.name} className="compat-share-daypillar-card">
                <div className="compat-share-daypillar-head">
                  <span className="compat-share-daypillar-name">{c.name}</span>
                  <span className="compat-share-daypillar-gz">{c.gz}日</span>
                </div>
                <div className="compat-share-daypillar-tag">{c.portrait.tag}</div>
                <p className="compat-share-daypillar-text">{c.portrait.text}</p>
              </div>
            ))}
          </div>
        </section>
      )}
```

- [ ] **Step 4: 加样式**

```css
/* ◇ 日柱速写 */
.compat-share-daypillar { position: relative; z-index: 1; }
.compat-share-daypillar-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 8px;
}
.compat-share-daypillar-card {
  background: #fbf3e0;
  border-radius: 6px;
  padding: 8px 10px;
}
.compat-share-daypillar-head {
  display: flex;
  justify-content: space-between;
  align-items: baseline;
  margin-bottom: 3px;
}
.compat-share-daypillar-name { font-size: 12px; font-weight: 700; color: #5c4a3a; }
.compat-share-daypillar-gz { font-size: 10px; color: #9b815c; font-family: "Noto Serif SC", serif; }
.compat-share-daypillar-tag {
  display: inline-block;
  font-size: 10px;
  font-weight: 700;
  color: #7a5c3a;
  background: #efe0bc;
  border-radius: 3px;
  padding: 1px 5px;
  margin-bottom: 3px;
}
.compat-share-daypillar-text { margin: 0; font-size: 10px; line-height: 1.5; color: #4a3728; }
```

- [ ] **Step 5: 类型检查 + 全测试**

Run: `cd frontend && npx tsc --noEmit && node --test tests/day-pillar-portrait.test.mjs`
Expected: 全绿（分享图 `getDayPillarPortrait` 断言转绿）。

- [ ] **Step 6: Commit**

```bash
git add frontend/src/components/CompatibilityShareCard.tsx frontend/src/components/CompatibilityShareCard.css
git commit -m "feat(compat): add day-pillar portrait to share card"
```

---

## Task 7: 全量回归 + 收尾

- [ ] **Step 1: 跑前端全测试**

Run: `cd frontend && node --test tests/`
Expected: `day-pillar-portrait.test.mjs` 全绿；其余测试不因本次改动新增失败（既有的 chart-archive / import-idempotency 等历史失败与本改动无关，不在范围内）。

- [ ] **Step 2: 类型检查**

Run: `cd frontend && npx tsc --noEmit`
Expected: 无新增报错。

- [ ] **Step 3: 视觉自查（:5200）**

打开一个合盘结果页，确认：结果页「日柱·本命速写」双卡渲染；导出 PDF 在概要页有「本命日柱速写」；分享图有「◇ 日柱速写」。旧数据（无 chart_snapshot）该块整体不渲染、不报错。
