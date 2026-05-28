# 合盘导出（图片 + PDF）Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 给 `CompatibilityResultPage` 加上图片（分享卡）+ PDF（完整报告）双导出能力，独立新建组件，不动单盘。

**Architecture:** 两个新组件 `CompatibilityShareCard` (400px PNG 国风分享卡，含双盘四柱/综合分/四维条/3 条核心证据/一句话定调) 与 `CompatibilityPrintLayout` (A4 七章节打印版式)。结果页加两按钮：图片走 modal 预览 + `html-to-image` 三端分流；PDF 走桌面 `window.print()` / 移动 `html2canvas + jsPDF` 双轨。

**Tech Stack:** React 19 + TypeScript strict + Vite + html-to-image / html2canvas / jspdf。无前端测试框架——验证靠 `tsc -b && vite build` (`npm run build`) + `npm run lint` + 手工 smoke。

**Spec:** `docs/superpowers/specs/2026-05-28-compatibility-export-design.md`

**Branch policy:** 在 `main` 上直接迭代（项目惯例）。每个 task 一个 commit。所有 commit 必须含 `Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>` trailer。

---

## File Structure

```
新增（4 个）:
  frontend/src/components/CompatibilityShareCard.tsx     ── 分享卡（400px 长图）
  frontend/src/components/CompatibilityShareCard.css     ── 分享卡专属样式
  frontend/src/components/CompatibilityPrintLayout.tsx   ── A4 七章节打印版式
  frontend/src/components/CompatibilityPrintLayout.css   ── @media print 样式

修改（2 个）:
  frontend/src/pages/CompatibilityResultPage.tsx         ── 接入按钮 + 状态 + handlers + 组件挂载
  frontend/src/pages/CompatibilityResultPage.css         ── 按钮 / modal / @media print 样式
```

---

## Pre-flight checks (before T1)

- [ ] **A. 检查工作区状态**
  Run: `git status`
  Expected: clean working tree on `main`. 若不是，先停下让用户处理。

- [ ] **B. 检查依赖已装**
  Run: `grep -E '"html-to-image"|"html2canvas"|"jspdf"' frontend/package.json`
  Expected: 三个库都有版本号（已确认 `html-to-image: ^1.11.13`, `html2canvas: ^1.4.1`, `jspdf: ^4.2.1`）。

- [ ] **C. 读 spec 与当前合盘页关键段**
  Read: `docs/superpowers/specs/2026-05-28-compatibility-export-design.md`
  Read: `frontend/src/pages/CompatibilityResultPage.tsx:884-955`（state + 派生）
  Read: `frontend/src/components/ShareCard.tsx`（单盘分享卡参考）
  Read: `frontend/src/components/PrintLayout.tsx:1-260`（单盘打印版式参考）

- [ ] **D. 检查类型确实存在**
  Run: `grep -nE "^export interface (CompatibilityReading|CompatibilityParticipant|CompatibilityChartSnapshot|CompatibilityEvidence|CompatibilityStructuredReport|CompatibilityStageRisk|ExportBrand)\b" frontend/src/lib/api.ts`
  Expected: 7 行命中，类型俱在。

---

## Task 1: 新建 `CompatibilityShareCard` 组件 + CSS

**Files:**
- Create: `frontend/src/components/CompatibilityShareCard.tsx`
- Create: `frontend/src/components/CompatibilityShareCard.css`

**Spec 引用:** §3.3、§5（边界处理）

**身份识别**: ShareCard 内部对 `participants` 数组用 `find(p => p.role === 'self'|'partner')` 派生，不依赖位置。

- [ ] **Step 1: 创建 CompatibilityShareCard.tsx**

Write 到 `frontend/src/components/CompatibilityShareCard.tsx`：

```tsx
import { forwardRef } from 'react'
import type {
  CompatibilityEvidence,
  CompatibilityParticipant,
  CompatibilityReading,
  CompatibilityStructuredReport,
  ExportBrand,
} from '../lib/api'
import { isV3DimensionScores } from '../lib/api'
import type { DecisionDashboardData } from '../lib/compatibilityDecision'
import { resolveFooter, showDiagonalWatermark } from '../lib/brandText'
import './CompatibilityShareCard.css'

const PILLAR_FONT_URL =
  'https://fonts.googleapis.com/css2?family=Noto+Serif+SC:wght@700&text=%E7%94%B2%E4%B9%99%E4%B8%99%E4%B8%81%E6%88%8A%E5%B7%B1%E5%BA%9A%E8%BE%9B%E5%A3%AC%E7%99%B8%E5%AD%90%E4%B8%91%E5%AF%85%E5%8D%AF%E8%BE%B0%E5%B7%B3%E5%8D%88%E6%9C%AA%E7%94%B3%E9%85%89%E6%88%8C%E4%BA%A5&display=swap'

const WX_COLOR: Record<string, string> = {
  '木': '#4a7c59',
  '火': '#c0392b',
  '土': '#a0784a',
  '金': '#7a6830',
  '水': '#2c5282',
}

function wxColor(wxStr: string | undefined) {
  if (!wxStr) return '#5c4a3a'
  for (const [k, v] of Object.entries(WX_COLOR)) {
    if (wxStr.startsWith(k)) return v
  }
  return '#5c4a3a'
}

const POLARITY_COLOR: Record<string, string> = {
  positive: '#66bb6a',
  negative: '#ef5350',
  mixed: '#ffb74d',
  neutral: '#9e9e9e',
}

const DIM_LABEL_LEGACY: Record<string, string> = {
  attraction: '吸引',
  stability: '稳定',
  communication: '沟通',
  practicality: '现实',
}

const DIM_LABEL_V3: Record<string, string> = {
  zodiac: '属相',
  nayin: '纳音',
  day_pillar: '日柱',
  eight_chars: '八字',
}

const EVIDENCE_SOURCE_LABEL: Record<string, string> = {
  day_master: '日主',
  five_elements: '五行',
  spouse_palace: '夫妻宫',
  spouse_star: '配偶星',
  ganzhi: '干支',
  shensha: '神煞',
  ten_god_interaction: '十神',
  favorable_element_support: '喜忌',
  ganzhi_interaction: '干支合冲',
  relationship_pattern: '关系',
  timing_context: '阶段',
  zodiac: '属相',
  nayin: '纳音',
  day_pillar: '日柱',
  eight_chars: '八字',
}

function ChartColumn({ participant, label }: { participant?: CompatibilityParticipant; label: string }) {
  if (!participant?.chart_snapshot) {
    return (
      <div className="compat-share-col">
        <div className="compat-share-col-head">{label}</div>
        <div className="compat-share-col-empty">数据缺失</div>
      </div>
    )
  }
  const c = participant.chart_snapshot
  const genderTxt = c.gender === 'female' ? '女' : '男'
  const cells = [
    { label: '年', gan: c.year_gan, zhi: c.year_zhi },
    { label: '月', gan: c.month_gan, zhi: c.month_zhi },
    { label: '日', gan: c.day_gan, zhi: c.day_zhi },
    { label: '时', gan: c.hour_gan, zhi: c.hour_zhi },
  ]
  return (
    <div className="compat-share-col">
      <div className="compat-share-col-head">{participant.display_name || label} · {genderTxt}</div>
      <div className="compat-share-pillars">
        {cells.map(cell => (
          <div className="compat-share-pillar" key={cell.label}>
            <span className="compat-share-pillar-lbl">{cell.label}</span>
            <span className="compat-share-pillar-gan">{cell.gan}</span>
            <span className="compat-share-pillar-zhi">{cell.zhi}</span>
          </div>
        ))}
      </div>
    </div>
  )
}

function EvidenceItem({ evidence }: { evidence: CompatibilityEvidence }) {
  const color = POLARITY_COLOR[evidence.polarity] || POLARITY_COLOR.neutral
  const src = EVIDENCE_SOURCE_LABEL[evidence.source] || evidence.source
  return (
    <div className="compat-share-ev" style={{ borderLeftColor: color }}>
      <div className="compat-share-ev-head">
        <span className="compat-share-ev-src">{src}</span>
        <span className="compat-share-ev-title">{evidence.title}</span>
      </div>
      <div className="compat-share-ev-detail">{evidence.detail}</div>
    </div>
  )
}

function DiagonalWatermark({ text }: { text: string }) {
  const items = Array.from({ length: 16 }, (_, i) => i)
  return (
    <div className="compat-share-watermark" aria-hidden>
      {items.map(i => <span key={i}>{text}</span>)}
    </div>
  )
}

export interface CompatibilityShareCardProps {
  reading: CompatibilityReading
  participants: CompatibilityParticipant[]
  evidences: CompatibilityEvidence[]
  structured: CompatibilityStructuredReport | null
  decision: DecisionDashboardData
  brand?: ExportBrand | null
}

const CompatibilityShareCard = forwardRef<HTMLDivElement, CompatibilityShareCardProps>((props, ref) => {
  const { reading, participants, evidences, structured, decision, brand } = props
  const selfP = participants.find(p => p.role === 'self')
  const partnerP = participants.find(p => p.role === 'partner')

  const top3Evidences = [...evidences]
    .filter(e => Number.isFinite(e.weight))
    .sort((a, b) => b.weight - a.weight)
    .slice(0, 3)

  const isV3 = (reading.analysis_version === 'v3' || reading.analysis_version === 'v3.1')
    && isV3DimensionScores(reading.dimension_scores)
  const dimLabel = isV3 ? DIM_LABEL_V3 : DIM_LABEL_LEGACY
  const dimEntries = Object.entries(reading.dimension_scores)
    .filter(([k]) => dimLabel[k])
    .map(([k, v]) => ({ key: k, label: dimLabel[k], value: typeof v === 'number' ? v : 0 }))

  const verdict = structured?.summary?.trim() || decision.verdict
  const resolvedTitle = brand?.title || '缘 聚 合 盘'
  const resolvedFooter = resolveFooter(brand, 'yuanju.com')
  const showWatermark = showDiagonalWatermark(brand)

  return (
    <>
      <link rel="stylesheet" href={PILLAR_FONT_URL} />
      <div ref={ref} className="compat-share-card">
        {showWatermark && <DiagonalWatermark text={brand?.watermark_text || brand?.title || '缘聚命理'} />}

        <header className="compat-share-header">
          <h1>{resolvedTitle}</h1>
          <p className="compat-share-sub">YUANJU · 命理合参</p>
        </header>

        <section className="compat-share-twocol">
          <ChartColumn participant={selfP} label="我" />
          <div className="compat-share-vdivider" />
          <ChartColumn participant={partnerP} label="伴侣" />
        </section>

        <section className="compat-share-score">
          <div className="compat-share-score-label">综 合 契 合 度</div>
          <div className="compat-share-score-value" style={{ color: wxColor(selfP?.chart_snapshot?.day_gan) }}>
            {reading.overall_score}
          </div>
        </section>

        {dimEntries.length > 0 && (
          <section className="compat-share-dims">
            <h3 className="compat-share-section-h">◇ 四维</h3>
            {dimEntries.map(d => (
              <div key={d.key} className="compat-share-dim-row">
                <span className="compat-share-dim-lbl">{d.label}</span>
                <div className="compat-share-dim-bar">
                  <i style={{ width: `${Math.max(0, Math.min(100, d.value))}%` }} />
                </div>
                <span className="compat-share-dim-val">{d.value}</span>
              </div>
            ))}
          </section>
        )}

        {top3Evidences.length > 0 && (
          <section className="compat-share-evs">
            <h3 className="compat-share-section-h">◇ 核心证据</h3>
            {top3Evidences.map(e => <EvidenceItem key={e.id} evidence={e} />)}
          </section>
        )}

        <section className="compat-share-verdict">{verdict}</section>

        <footer className="compat-share-footer">{resolvedFooter}</footer>
      </div>
    </>
  )
})

CompatibilityShareCard.displayName = 'CompatibilityShareCard'
export default CompatibilityShareCard
```

- [ ] **Step 2: 创建 CompatibilityShareCard.css**

Write 到 `frontend/src/components/CompatibilityShareCard.css`：

```css
.compat-share-card {
  position: relative;
  width: 400px;
  background: #fdf9f2;
  padding: 24px 20px 18px;
  font-family: "Noto Sans SC", "PingFang SC", "Microsoft YaHei", sans-serif;
  color: #4a3728;
  box-sizing: border-box;
  overflow: hidden;
  border: 1px solid #d4b896;
  border-radius: 8px;
}

.compat-share-header {
  text-align: center;
  position: relative;
  z-index: 1;
}
.compat-share-header h1 {
  margin: 0;
  font-family: "Noto Serif SC", serif;
  font-size: 20px;
  letter-spacing: 6px;
  color: #7a5c3a;
  font-weight: 700;
}
.compat-share-sub {
  margin: 4px 0 0;
  font-size: 10px;
  letter-spacing: 3px;
  color: #9b815c;
}

.compat-share-twocol {
  display: grid;
  grid-template-columns: 1fr 1px 1fr;
  gap: 10px;
  margin: 18px 0;
  position: relative;
  z-index: 1;
}
.compat-share-vdivider {
  background: linear-gradient(to bottom, transparent, #d4b896, transparent);
}
.compat-share-col-head {
  text-align: center;
  font-size: 12px;
  font-weight: 700;
  color: #5c4a3a;
  margin-bottom: 8px;
  letter-spacing: 1px;
}
.compat-share-col-empty {
  text-align: center;
  font-size: 11px;
  color: #9b815c;
}
.compat-share-pillars {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 4px;
}
.compat-share-pillar {
  background: #f5ebd6;
  border-radius: 4px;
  padding: 6px 2px;
  text-align: center;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 2px;
  font-family: "Noto Serif SC", serif;
}
.compat-share-pillar-lbl {
  font-size: 9px;
  color: #9b815c;
  letter-spacing: 1px;
}
.compat-share-pillar-gan,
.compat-share-pillar-zhi {
  font-size: 16px;
  font-weight: 700;
  color: #4a3728;
  line-height: 1.1;
}

.compat-share-score {
  text-align: center;
  margin: 8px 0 6px;
  position: relative;
  z-index: 1;
}
.compat-share-score-label {
  font-size: 11px;
  letter-spacing: 4px;
  color: #7a5c3a;
}
.compat-share-score-value {
  font-family: "Noto Serif SC", serif;
  font-size: 44px;
  font-weight: 800;
  color: #c9a96e;
  letter-spacing: 2px;
  margin-top: 4px;
  line-height: 1;
}

.compat-share-section-h {
  font-size: 12px;
  font-weight: 700;
  color: #7a5c3a;
  margin: 14px 0 6px;
  letter-spacing: 3px;
  font-family: "Noto Serif SC", serif;
}

.compat-share-dims {
  position: relative;
  z-index: 1;
}
.compat-share-dim-row {
  display: grid;
  grid-template-columns: 36px 1fr 28px;
  gap: 8px;
  align-items: center;
  margin: 4px 0;
  font-size: 11px;
}
.compat-share-dim-lbl {
  color: #5c4a3a;
}
.compat-share-dim-val {
  text-align: right;
  color: #7a5c3a;
  font-weight: 700;
}
.compat-share-dim-bar {
  background: #efe0bc;
  height: 6px;
  border-radius: 3px;
  overflow: hidden;
}
.compat-share-dim-bar > i {
  display: block;
  height: 100%;
  background: linear-gradient(90deg, #c9a96e, #9b815c);
}

.compat-share-evs {
  position: relative;
  z-index: 1;
}
.compat-share-ev {
  padding: 6px 8px;
  margin-top: 4px;
  border-left: 3px solid #c9a96e;
  background: #fbf3e0;
  border-radius: 0 4px 4px 0;
  font-size: 11px;
  line-height: 1.5;
}
.compat-share-ev-head {
  display: flex;
  gap: 6px;
  align-items: baseline;
  margin-bottom: 2px;
}
.compat-share-ev-src {
  display: inline-block;
  background: #c9a96e;
  color: #fff;
  font-size: 9px;
  padding: 1px 5px;
  border-radius: 2px;
  letter-spacing: 1px;
  flex-shrink: 0;
}
.compat-share-ev-title {
  font-weight: 700;
  color: #4a3728;
}
.compat-share-ev-detail {
  color: #5c4a3a;
}

.compat-share-verdict {
  background: #f5ebd6;
  padding: 10px 12px;
  border-radius: 6px;
  text-align: center;
  font-size: 12px;
  line-height: 1.6;
  margin-top: 14px;
  color: #4a3728;
  position: relative;
  z-index: 1;
}

.compat-share-footer {
  margin-top: 14px;
  text-align: center;
  font-size: 10px;
  color: #9b815c;
  letter-spacing: 2px;
  position: relative;
  z-index: 1;
}

.compat-share-watermark {
  position: absolute;
  inset: -30%;
  pointer-events: none;
  overflow: hidden;
  z-index: 0;
  transform: rotate(-28deg);
  display: grid;
  grid-template-columns: repeat(auto-fill, 130px);
  gap: 50px 30px;
  opacity: 0.06;
  color: #000;
  font-size: 14px;
  font-family: "Noto Sans SC", sans-serif;
  white-space: nowrap;
}
```

- [ ] **Step 3: 验证 tsc + lint 通过**

Run: `cd frontend && npm run lint && npm run build`
Expected: 全绿。两个文件都被打包，无类型错误。如果报 `Cannot find module './CompatibilityShareCard.css'`，确认 .css 文件已创建。

- [ ] **Step 4: Commit**

```bash
git add frontend/src/components/CompatibilityShareCard.tsx frontend/src/components/CompatibilityShareCard.css
git commit -m "$(cat <<'EOF'
feat(compat-export): add CompatibilityShareCard component

400px-wide national-style PNG share card matching mockup C: dual-chart
pillars side-by-side, composite score hero, 4-dimension bars (v2 legacy
or v3 zodiac/nayin/day_pillar/eight_chars auto-detected via
isV3DimensionScores), top-3 evidences sorted by weight, verdict line,
and brand-aware footer + optional diagonal watermark. Inner ChartColumn
uses participants.find(role) rather than positional indexing.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

## Task 2: 新建 `CompatibilityPrintLayout` 组件 + CSS

**Files:**
- Create: `frontend/src/components/CompatibilityPrintLayout.tsx`
- Create: `frontend/src/components/CompatibilityPrintLayout.css`

**Spec 引用:** §3.4、§3.7

- [ ] **Step 1: 创建 CompatibilityPrintLayout.tsx**

Write 到 `frontend/src/components/CompatibilityPrintLayout.tsx`：

```tsx
import type {
  CompatibilityEvidence,
  CompatibilityParticipant,
  CompatibilityReading,
  CompatibilityStageRisk,
  CompatibilityStructuredReport,
  ExportBrand,
} from '../lib/api'
import { isV3DimensionScores } from '../lib/api'
import type { DecisionDashboardData } from '../lib/compatibilityDecision'
import { cleanReportText, splitParagraphs } from '../lib/reportText'
import { resolveFooter, showDiagonalWatermark } from '../lib/brandText'
import './CompatibilityPrintLayout.css'

const DIM_LABEL_LEGACY: Record<string, string> = {
  attraction: '吸引',
  stability: '稳定',
  communication: '沟通',
  practicality: '现实',
}
const DIM_LABEL_V3: Record<string, string> = {
  zodiac: '属相',
  nayin: '纳音',
  day_pillar: '日柱',
  eight_chars: '八字',
}
const POLARITY_LABEL: Record<string, string> = {
  positive: '正向',
  negative: '风险',
  mixed: '复杂',
  neutral: '中性',
}
const STAGE_WINDOW_LABEL: Record<string, string> = {
  three_months: '3 个月内',
  one_year: '1 年内',
  two_years_plus: '2 年以上',
}
const RISK_LEVEL_LABEL: Record<string, string> = {
  high: '偏高',
  medium: '中等',
  low: '偏低',
}
const EVIDENCE_SOURCE_LABEL: Record<string, string> = {
  day_master: '日主关系',
  five_elements: '五行结构',
  spouse_palace: '夫妻宫',
  spouse_star: '配偶星',
  ganzhi: '冲克总量',
  shensha: '神煞辅助',
  ten_god_interaction: '十神互动',
  favorable_element_support: '喜忌互补',
  ganzhi_interaction: '干支合冲刑害',
  relationship_pattern: '关系模式',
  timing_context: '阶段时机',
  zodiac: '合属相',
  nayin: '合纳音',
  day_pillar: '合日柱',
  eight_chars: '合八字',
}

function PrintHeader({ title, brand }: { title: string; brand?: ExportBrand | null }) {
  const isWordmark = brand?.logo_mode === 'wordmark' && !!brand?.logo_url
  return (
    <div className="compat-print-header">
      <span className="compat-print-header-left">
        {brand?.logo_url ? (
          isWordmark ? (
            <img className="compat-print-header-wordmark" src={brand.logo_url} alt={brand.title} />
          ) : (
            <img className="compat-print-header-logo" src={brand.logo_url} alt={brand.title} />
          )
        ) : null}
        <span className="compat-print-header-brand">{title}</span>
      </span>
      <span className="compat-print-header-center">命理合参报告</span>
      <span className="compat-print-header-info">YUANJU</span>
    </div>
  )
}

function ParticipantsHero({ participants, reading }: {
  participants: CompatibilityParticipant[]
  reading: CompatibilityReading
}) {
  const selfP = participants.find(p => p.role === 'self')
  const partnerP = participants.find(p => p.role === 'partner')
  return (
    <div className="compat-print-hero">
      <div className="compat-print-hero-cols">
        <PersonBlock participant={selfP} label="我" />
        <PersonBlock participant={partnerP} label="伴侣" />
      </div>
      <div className="compat-print-hero-score">
        <span className="compat-print-hero-score-lbl">综合契合度</span>
        <span className="compat-print-hero-score-value">{reading.overall_score}</span>
        <span className="compat-print-hero-score-level">{reading.overall_level === 'high' ? '偏高' : reading.overall_level === 'low' ? '偏低' : '中等'}</span>
      </div>
    </div>
  )
}

function PersonBlock({ participant, label }: { participant?: CompatibilityParticipant; label: string }) {
  if (!participant?.chart_snapshot) return <div className="compat-print-person">数据缺失</div>
  const c = participant.chart_snapshot
  const gen = c.gender === 'female' ? '女命' : '男命'
  return (
    <div className="compat-print-person">
      <div className="compat-print-person-name">{participant.display_name || label} · {gen}</div>
      <div className="compat-print-person-birth">{c.birth_year}年{c.birth_month}月{c.birth_day}日 {c.birth_hour}时</div>
      <div className="compat-print-person-pillars">
        {`${c.year_gan}${c.year_zhi} · ${c.month_gan}${c.month_zhi} · ${c.day_gan}${c.day_zhi} · ${c.hour_gan}${c.hour_zhi}`}
      </div>
    </div>
  )
}

function DecisionBlock({ decision }: { decision: DecisionDashboardData }) {
  return (
    <div className="compat-print-decision">
      <p className="compat-print-decision-verdict">{decision.verdict}</p>
      <p className="compat-print-decision-summary">{decision.summary}</p>
      <div className="compat-print-decision-grid">
        <div><span className="lbl">关系定调</span>{decision.relationshipType}</div>
        <div><span className="lbl">推进建议</span>{decision.recommendationLabel} · 信心{decision.confidenceLabel}</div>
        <div><span className="lbl">最大风险</span>{decision.maxRisk}</div>
        <div><span className="lbl">下一步动作</span>{decision.nextAction}</div>
      </div>
      {decision.avoid.length > 0 && (
        <div className="compat-print-decision-avoid">
          <span className="lbl">避免</span>
          <ul>{decision.avoid.map((a, i) => <li key={i}>{a}</li>)}</ul>
        </div>
      )}
    </div>
  )
}

function ScorePrint({ reading }: { reading: CompatibilityReading }) {
  const isV3 = (reading.analysis_version === 'v3' || reading.analysis_version === 'v3.1')
    && isV3DimensionScores(reading.dimension_scores)
  const labels = isV3 ? DIM_LABEL_V3 : DIM_LABEL_LEGACY
  const entries = Object.entries(reading.dimension_scores)
    .filter(([k]) => labels[k])
    .map(([k, v]) => ({ key: k, label: labels[k], value: typeof v === 'number' ? v : 0 }))
  return (
    <div className="compat-print-scores">
      {entries.map(e => (
        <div key={e.key} className="compat-print-score-row">
          <span className="compat-print-score-lbl">{e.label}</span>
          <div className="compat-print-score-bar"><i style={{ width: `${Math.max(0, Math.min(100, e.value))}%` }} /></div>
          <span className="compat-print-score-val">{e.value}</span>
        </div>
      ))}
    </div>
  )
}

function EvidenceTable({ evidences }: { evidences: CompatibilityEvidence[] }) {
  if (!evidences.length) return <p className="compat-print-empty">暂无命理证据</p>
  return (
    <table className="compat-print-ev-table">
      <thead>
        <tr><th>来源</th><th>极性</th><th>标题</th><th>说明</th></tr>
      </thead>
      <tbody>
        {evidences.map(e => (
          <tr key={e.id} className={`compat-print-ev-row compat-print-ev-${e.polarity}`}>
            <td>{EVIDENCE_SOURCE_LABEL[e.source] || e.source}</td>
            <td>{POLARITY_LABEL[e.polarity] || e.polarity}</td>
            <td>{e.title}</td>
            <td>{e.detail}</td>
          </tr>
        ))}
      </tbody>
    </table>
  )
}

function StageRisksBlock({ risks }: { risks: CompatibilityStageRisk[] }) {
  if (!risks.length) return <p className="compat-print-empty">暂无阶段风险数据</p>
  return (
    <div className="compat-print-stages">
      {risks.map((r, i) => (
        <div key={i} className="compat-print-stage">
          <div className="compat-print-stage-head">
            <span className="compat-print-stage-window">{STAGE_WINDOW_LABEL[r.window] || r.window}</span>
            <span className={`compat-print-stage-level lvl-${r.risk_level}`}>{RISK_LEVEL_LABEL[r.risk_level] || r.risk_level}</span>
          </div>
          <div><span className="lbl">主要风险</span>{r.main_risk}</div>
          <div><span className="lbl">触发条件</span>{r.trigger}</div>
          <div><span className="lbl">应对建议</span>{r.advice}</div>
        </div>
      ))}
    </div>
  )
}

function ChapterBlock({ title, content }: { title: string; content: string }) {
  const paragraphs = splitParagraphs(content)
  return (
    <div className="compat-print-chapter">
      <h4 className="compat-print-chapter-title">{title}</h4>
      {paragraphs.length > 0
        ? paragraphs.map((p, i) => <p key={i}>{p}</p>)
        : <p>{cleanReportText(content)}</p>}
    </div>
  )
}

function ChartFull({ participant, label }: { participant?: CompatibilityParticipant; label: string }) {
  if (!participant?.chart_snapshot) return <div className="compat-print-chartfull">数据缺失</div>
  const c = participant.chart_snapshot
  const gen = c.gender === 'female' ? '女命' : '男命'
  const cells = [
    { lbl: '年柱', gan: c.year_gan, zhi: c.year_zhi },
    { lbl: '月柱', gan: c.month_gan, zhi: c.month_zhi },
    { lbl: '日柱', gan: c.day_gan, zhi: c.day_zhi },
    { lbl: '时柱', gan: c.hour_gan, zhi: c.hour_zhi },
  ]
  const wuxing = c.wuxing
  return (
    <div className="compat-print-chartfull">
      <div className="compat-print-chartfull-head">{participant.display_name || label} · {gen}</div>
      <div className="compat-print-chartfull-birth">{c.birth_year}年{c.birth_month}月{c.birth_day}日 {c.birth_hour}时</div>
      <table className="compat-print-chartfull-table">
        <tbody>
          <tr>
            {cells.map(p => <td key={p.lbl}><div className="lbl">{p.lbl}</div><div className="gz">{p.gan}<br />{p.zhi}</div></td>)}
          </tr>
        </tbody>
      </table>
      {wuxing && (
        <div className="compat-print-chartfull-wuxing">
          木 {wuxing.mu} · 火 {wuxing.huo} · 土 {wuxing.tu} · 金 {wuxing.jin} · 水 {wuxing.shui}
        </div>
      )}
    </div>
  )
}

function PrintFooter({ text }: { text: string }) {
  return <div className="compat-print-footer">{text}</div>
}

function DiagonalWatermark({ text }: { text: string }) {
  const items = Array.from({ length: 40 }, (_, i) => i)
  return (
    <div className="compat-print-watermark" aria-hidden>
      {items.map(i => <span key={i}>{text}</span>)}
    </div>
  )
}

export interface CompatibilityPrintLayoutProps {
  reading: CompatibilityReading
  participants: CompatibilityParticipant[]
  evidences: CompatibilityEvidence[]
  decision: DecisionDashboardData
  stageRisks: CompatibilityStageRisk[]
  structured: CompatibilityStructuredReport | null
  brand?: ExportBrand | null
}

export default function CompatibilityPrintLayout(props: CompatibilityPrintLayoutProps) {
  const { reading, participants, evidences, decision, stageRisks, structured, brand } = props
  const selfP = participants.find(p => p.role === 'self')
  const partnerP = participants.find(p => p.role === 'partner')
  const title = brand?.title || '缘 聚 合 盘'
  const footerText = resolveFooter(brand, 'yuanju.com')
  const showWatermark = showDiagonalWatermark(brand)

  return (
    <div className="print-only compat-print-layout">
      {showWatermark && <DiagonalWatermark text={brand?.watermark_text || brand?.title || '缘聚命理'} />}

      <table className="compat-print-table">
        <thead><tr><td><PrintHeader title={title} brand={brand} /></td></tr></thead>
        <tbody><tr><td className="compat-print-body">

          <section className="compat-print-section">
            <h2 className="compat-print-section-title">一、合参概要</h2>
            <ParticipantsHero participants={participants} reading={reading} />
          </section>
          <div className="compat-print-page-break" />

          <section className="compat-print-section">
            <h2 className="compat-print-section-title">二、决策仪表盘</h2>
            <DecisionBlock decision={decision} />
          </section>
          <div className="compat-print-page-break" />

          <section className="compat-print-section">
            <h2 className="compat-print-section-title">三、评分明细</h2>
            <ScorePrint reading={reading} />
          </section>
          <div className="compat-print-page-break" />

          <section className="compat-print-section">
            <h2 className="compat-print-section-title">四、命理证据</h2>
            <EvidenceTable evidences={evidences} />
          </section>
          <div className="compat-print-page-break" />

          <section className="compat-print-section">
            <h2 className="compat-print-section-title">五、阶段风险与验证</h2>
            <StageRisksBlock risks={stageRisks} />
          </section>
          <div className="compat-print-page-break" />

          {structured && (
            <>
              <section className="compat-print-section">
                <h2 className="compat-print-section-title">六、命理解读</h2>
                {structured.summary && <p className="compat-print-summary">{structured.summary}</p>}
                {structured.dimensions.map(chap => (
                  <ChapterBlock key={chap.key} title={chap.title} content={chap.content} />
                ))}
                {structured.advice && (
                  <div className="compat-print-chapter">
                    <h4 className="compat-print-chapter-title">综合建议</h4>
                    {splitParagraphs(structured.advice).map((p, i) => <p key={i}>{p}</p>)}
                  </div>
                )}
              </section>
              <div className="compat-print-page-break" />
            </>
          )}

          <section className="compat-print-section">
            <h2 className="compat-print-section-title">七、双盘原图</h2>
            <div className="compat-print-charts-grid">
              <ChartFull participant={selfP} label="我" />
              <ChartFull participant={partnerP} label="伴侣" />
            </div>
          </section>

        </td></tr></tbody>
        <tfoot><tr><td><PrintFooter text={footerText} /></td></tr></tfoot>
      </table>
    </div>
  )
}
```

- [ ] **Step 2: 创建 CompatibilityPrintLayout.css**

Write 到 `frontend/src/components/CompatibilityPrintLayout.css`：

```css
.compat-print-layout {
  display: none;
}

@media print {
  @page {
    size: A4 portrait;
    margin: 18mm 16mm 18mm 16mm;
  }

  .compat-print-layout {
    display: block;
    font-family: "Noto Serif SC", "SimSun", serif;
    color: #2a1a0a;
    background: #fff;
  }

  .compat-print-table {
    width: 100%;
    border-collapse: collapse;
  }
  .compat-print-table thead { display: table-header-group; }
  .compat-print-table tfoot { display: table-footer-group; }
  .compat-print-table td { padding: 0; }

  .compat-print-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 4mm 0;
    border-bottom: 1px solid #e0cca0;
    font-family: "Noto Serif SC", serif;
  }
  .compat-print-header-left {
    display: inline-flex;
    align-items: center;
    gap: 6px;
  }
  .compat-print-header-logo {
    width: 18px;
    height: 18px;
    object-fit: contain;
  }
  .compat-print-header-wordmark {
    height: 6mm;
    max-width: 80mm;
    object-fit: contain;
  }
  .compat-print-header-brand {
    font-size: 11pt;
    font-weight: 700;
    letter-spacing: 3px;
    color: #b8952a;
  }
  .compat-print-header-center {
    font-size: 10pt;
    letter-spacing: 4px;
    color: #5a3a1a;
  }
  .compat-print-header-info {
    font-size: 8pt;
    letter-spacing: 1px;
    color: #999;
  }

  .compat-print-body {
    padding: 8mm 0 4mm;
    font-size: 11pt;
    line-height: 1.7;
  }

  .compat-print-section {
    break-inside: avoid;
  }
  .compat-print-section-title {
    font-family: "Noto Serif SC", serif;
    font-size: 16pt;
    color: #5a3a1a;
    border-bottom: 1px solid #c9a96e;
    padding-bottom: 2mm;
    margin: 0 0 6mm 0;
    font-weight: 700;
  }
  .compat-print-page-break {
    page-break-after: always;
    height: 0;
  }
  .compat-print-empty {
    color: #888;
    font-style: italic;
  }

  /* §1 hero */
  .compat-print-hero-cols {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 8mm;
    margin-bottom: 6mm;
  }
  .compat-print-person {
    border: 1px solid #e0cca0;
    border-radius: 4pt;
    padding: 3mm 4mm;
  }
  .compat-print-person-name {
    font-size: 13pt;
    font-weight: 700;
    color: #5a3a1a;
  }
  .compat-print-person-birth {
    font-size: 10pt;
    color: #6b5638;
    margin: 1mm 0;
  }
  .compat-print-person-pillars {
    font-family: "Noto Serif SC", serif;
    font-size: 14pt;
    font-weight: 700;
    letter-spacing: 2px;
    color: #2a1a0a;
  }
  .compat-print-hero-score {
    display: flex;
    align-items: baseline;
    justify-content: center;
    gap: 8mm;
    margin-top: 4mm;
    border-top: 1px dashed #c9a96e;
    padding-top: 4mm;
  }
  .compat-print-hero-score-lbl {
    font-size: 11pt;
    letter-spacing: 3px;
    color: #7a5c3a;
  }
  .compat-print-hero-score-value {
    font-family: "Noto Serif SC", serif;
    font-size: 32pt;
    font-weight: 800;
    color: #c9a96e;
  }
  .compat-print-hero-score-level {
    font-size: 11pt;
    color: #5a3a1a;
  }

  /* §2 decision */
  .compat-print-decision-verdict {
    font-size: 14pt;
    font-weight: 700;
    color: #5a3a1a;
    margin: 0 0 2mm 0;
  }
  .compat-print-decision-summary {
    font-size: 11pt;
    color: #4a3728;
    margin: 0 0 4mm 0;
  }
  .compat-print-decision-grid {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 3mm 6mm;
    font-size: 11pt;
  }
  .compat-print-decision-grid .lbl,
  .compat-print-decision-avoid .lbl,
  .compat-print-stage .lbl {
    display: inline-block;
    min-width: 18mm;
    color: #7a5c3a;
    font-weight: 700;
    margin-right: 2mm;
  }
  .compat-print-decision-avoid {
    margin-top: 3mm;
    font-size: 11pt;
  }
  .compat-print-decision-avoid ul {
    margin: 1mm 0 0 18mm;
    padding: 0;
  }
  .compat-print-decision-avoid li { list-style: '· '; }

  /* §3 scores */
  .compat-print-score-row {
    display: grid;
    grid-template-columns: 18mm 1fr 12mm;
    gap: 4mm;
    align-items: center;
    margin: 2mm 0;
    font-size: 11pt;
  }
  .compat-print-score-lbl { color: #5a3a1a; font-weight: 700; }
  .compat-print-score-val { text-align: right; color: #7a5c3a; font-weight: 700; }
  .compat-print-score-bar {
    background: #efe0bc;
    height: 3mm;
    border-radius: 1.5mm;
    overflow: hidden;
  }
  .compat-print-score-bar > i {
    display: block;
    height: 100%;
    background: linear-gradient(90deg, #c9a96e, #9b815c);
  }

  /* §4 evidence table */
  .compat-print-ev-table {
    width: 100%;
    border-collapse: collapse;
    font-size: 10pt;
  }
  .compat-print-ev-table thead th {
    text-align: left;
    background: #f5ebd6;
    padding: 2mm 3mm;
    color: #5a3a1a;
    font-weight: 700;
    border-bottom: 1px solid #c9a96e;
  }
  .compat-print-ev-table tbody td {
    padding: 2mm 3mm;
    border-bottom: 1px solid #efe0bc;
    vertical-align: top;
  }
  .compat-print-ev-row.compat-print-ev-positive td:nth-child(2) { color: #2d5a3d; }
  .compat-print-ev-row.compat-print-ev-negative td:nth-child(2) { color: #8b1a1a; }
  .compat-print-ev-row.compat-print-ev-mixed td:nth-child(2) { color: #b8782a; }
  .compat-print-ev-row.compat-print-ev-neutral td:nth-child(2) { color: #555; }

  /* §5 stage risks */
  .compat-print-stage {
    border: 1px solid #e0cca0;
    border-radius: 4pt;
    padding: 3mm 4mm;
    margin-bottom: 3mm;
    font-size: 11pt;
    break-inside: avoid;
  }
  .compat-print-stage-head {
    display: flex;
    justify-content: space-between;
    margin-bottom: 2mm;
  }
  .compat-print-stage-window {
    font-size: 12pt;
    font-weight: 700;
    color: #5a3a1a;
  }
  .compat-print-stage-level {
    font-size: 10pt;
    padding: 0.5mm 2mm;
    border-radius: 2pt;
  }
  .compat-print-stage-level.lvl-high { background: #fdf0f0; color: #8b1a1a; }
  .compat-print-stage-level.lvl-medium { background: #fbf3e0; color: #b8782a; }
  .compat-print-stage-level.lvl-low { background: #f0f7f2; color: #2d5a3d; }

  /* §6 chapters */
  .compat-print-summary {
    font-size: 12pt;
    font-style: italic;
    color: #4a3728;
    border-left: 3px solid #c9a96e;
    padding-left: 4mm;
    margin: 0 0 5mm 0;
  }
  .compat-print-chapter {
    margin-bottom: 5mm;
    break-inside: avoid-page;
  }
  .compat-print-chapter-title {
    font-family: "Noto Serif SC", serif;
    font-size: 13pt;
    color: #5a3a1a;
    margin: 0 0 2mm 0;
    font-weight: 700;
  }
  .compat-print-chapter p {
    margin: 0 0 2mm 0;
    text-indent: 2em;
    font-size: 11pt;
    line-height: 1.8;
  }

  /* §7 dual full charts */
  .compat-print-charts-grid {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 8mm;
  }
  .compat-print-chartfull {
    border: 1px solid #e0cca0;
    border-radius: 4pt;
    padding: 3mm;
  }
  .compat-print-chartfull-head {
    font-size: 12pt;
    font-weight: 700;
    color: #5a3a1a;
    text-align: center;
  }
  .compat-print-chartfull-birth {
    font-size: 10pt;
    color: #6b5638;
    text-align: center;
    margin: 1mm 0 2mm;
  }
  .compat-print-chartfull-table {
    width: 100%;
    border-collapse: collapse;
    margin: 2mm 0;
  }
  .compat-print-chartfull-table td {
    border: 1px solid #e0cca0;
    text-align: center;
    padding: 2mm 0;
    vertical-align: middle;
  }
  .compat-print-chartfull-table .lbl {
    font-size: 9pt;
    color: #7a5c3a;
    margin-bottom: 1mm;
  }
  .compat-print-chartfull-table .gz {
    font-family: "Noto Serif SC", serif;
    font-size: 14pt;
    font-weight: 700;
    line-height: 1.2;
  }
  .compat-print-chartfull-wuxing {
    font-size: 9pt;
    color: #6b5638;
    text-align: center;
    margin-top: 2mm;
  }

  /* footer */
  .compat-print-footer {
    text-align: center;
    font-size: 9pt;
    color: #999;
    padding: 3mm 0;
    border-top: 1px solid #e0cca0;
  }

  /* diagonal watermark */
  .compat-print-watermark {
    position: fixed;
    top: -30%;
    left: -30%;
    right: -30%;
    bottom: -30%;
    pointer-events: none;
    overflow: hidden;
    z-index: 0;
    transform: rotate(-30deg);
    display: grid;
    grid-template-columns: repeat(auto-fill, 220px);
    gap: 80px 50px;
    opacity: 0.06;
    color: #000;
    font-size: 16px;
    font-family: "Noto Sans SC", sans-serif;
    white-space: nowrap;
  }
}
```

- [ ] **Step 3: 验证 tsc + lint 通过**

Run: `cd frontend && npm run lint && npm run build`
Expected: 全绿。注意：屏幕模式下 `.compat-print-layout` `display:none` 是默认，组件挂载但不占布局。

- [ ] **Step 4: Commit**

```bash
git add frontend/src/components/CompatibilityPrintLayout.tsx frontend/src/components/CompatibilityPrintLayout.css
git commit -m "$(cat <<'EOF'
feat(compat-export): add CompatibilityPrintLayout component

A4 portrait print layout for compatibility report, 7 sections with
forced page breaks: hero (双方信息 + 综合分), decision dashboard,
score detail (v2/v3 auto-detected), evidence table, stage risks,
AI chapter renderer (structured.dimensions + advice), dual full
charts grid. Reuses brandText / reportText helpers. Display:none
by default; only visible under @media print or when explicitly
shown via inline style (mobile jsPDF path).

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

## Task 3: 合盘页接入 — imports / state / handlers / brand fetch

**Files:**
- Modify: `frontend/src/pages/CompatibilityResultPage.tsx`

**Spec 引用:** §3.5、§3.7

参考 `frontend/src/pages/ResultPage.tsx:13-18`（导入风格）、`:359-516`（state + handler 范例）。

- [ ] **Step 1: 顶部新增 imports**

文件首部 import 区追加：

```tsx
// 在 CompatibilityResultPage.tsx 现有 import 区追加（约第 1-35 行的 import 之后）
import { useRef } from 'react'
import { toBlob, toPng } from 'html-to-image'
import jsPDF from 'jspdf'
import html2canvas from 'html2canvas'
import { brandAPI, type ExportBrand } from '../lib/api'
import CompatibilityShareCard from '../components/CompatibilityShareCard'
import CompatibilityPrintLayout from '../components/CompatibilityPrintLayout'
```

注意：`useEffect` / `useState` / `useCallback` 已在现有 import 中，`useRef` 可能需要补到同一行。把第 1 行的 `import { useCallback, useEffect, useState, type ReactNode } from 'react'` 改为：

```tsx
import { useCallback, useEffect, useRef, useState, type ReactNode } from 'react'
```

`brandAPI` 在 lib/api.ts:130 已 export，`ExportBrand` 在 lib/api.ts:113 已 export，无需新建。

- [ ] **Step 2: 函数顶部新增 state + 设备判定**

在 `CompatibilityResultPage` 函数体内（约 line 884 后，紧跟现有 `useState` 声明，例如 `error` 之后）插入：

```tsx
const [brand, setBrand] = useState<ExportBrand | null>(null)
const [shareModalOpen, setShareModalOpen] = useState(false)
const [savingImage, setSavingImage] = useState(false)
const [exportingPDF, setExportingPDF] = useState(false)
const shareCardRef = useRef<HTMLDivElement>(null)
const isMobileDevice = /iPhone|iPad|iPod|Android/i.test(navigator.userAgent)
const isIOS = /iPhone|iPad|iPod/i.test(navigator.userAgent)
```

- [ ] **Step 3: 新增 brand 拉取 useEffect**

紧跟现有 `useEffect`（约 line 899-910，处理 load 的那个）之后插入：

```tsx
useEffect(() => {
  if (!user) return
  brandAPI.get()
    .then(r => setBrand(r.data.data))
    .catch(() => setBrand(null))
}, [user])
```

- [ ] **Step 4: 新增 handleSaveImage handler**

在 `handleGenerateReport` 之后插入：

```tsx
const handleSaveImage = async () => {
  if (!shareCardRef.current) return
  setSavingImage(true)
  try {
    await document.fonts.ready
    const selfName = detail?.participants.find(p => p.role === 'self')?.display_name || '我'
    const partnerName = detail?.participants.find(p => p.role === 'partner')?.display_name || '伴侣'
    const fileName = `缘聚合盘-${selfName}-${partnerName}.png`

    if (isIOS) {
      const blob = await toBlob(shareCardRef.current, { quality: 0.98, pixelRatio: 3, cacheBust: true })
      if (!blob) throw new Error('生成图片失败')
      const file = new File([blob], fileName, { type: 'image/png' })
      if (navigator.canShare && navigator.canShare({ files: [file] })) {
        await navigator.share({
          files: [file],
          title: '缘聚合盘 · 命理合参',
          text: `${selfName} × ${partnerName} 综合契合度 ${detail?.reading.overall_score ?? ''} 分`,
        })
      } else {
        const objectUrl = URL.createObjectURL(blob)
        Object.assign(document.createElement('a'), { href: objectUrl, download: fileName }).click()
        setTimeout(() => URL.revokeObjectURL(objectUrl), 5000)
      }
    } else if (isMobileDevice) {
      const blob = await toBlob(shareCardRef.current, { quality: 0.98, pixelRatio: 3, cacheBust: true })
      if (!blob) throw new Error('生成图片失败')
      const objectUrl = URL.createObjectURL(blob)
      Object.assign(document.createElement('a'), { href: objectUrl, download: fileName }).click()
      setTimeout(() => URL.revokeObjectURL(objectUrl), 5000)
    } else {
      const dataUrl = await toPng(shareCardRef.current, { quality: 0.98, pixelRatio: 2, cacheBust: true })
      Object.assign(document.createElement('a'), { href: dataUrl, download: fileName }).click()
    }
  } catch (err: unknown) {
    const msg = err instanceof Error ? err.message : ''
    if (!msg.includes('AbortError') && !msg.includes('cancel')) {
      alert('生成图片失败，请稍后重试')
    }
  } finally {
    setSavingImage(false)
  }
}
```

- [ ] **Step 5: 新增 handleExportPDF handler**

紧跟 `handleSaveImage` 之后插入：

```tsx
const handleExportPDF = async () => {
  if (!detail?.latest_report?.content_structured) return
  if (!isMobileDevice) {
    window.print()
    return
  }
  const el = document.querySelector('.compat-print-layout') as HTMLElement | null
  if (!el) return
  setExportingPDF(true)
  const prevDisplay = el.style.display
  try {
    await document.fonts.ready
    el.style.display = 'block'
    const canvas = await html2canvas(el, { scale: 2, useCORS: true, logging: false })
    el.style.display = prevDisplay
    const imgData = canvas.toDataURL('image/jpeg', 0.92)
    const pdf = new jsPDF({ orientation: 'portrait', unit: 'mm', format: 'a4' })
    const pageW = pdf.internal.pageSize.getWidth()
    const pageH = pdf.internal.pageSize.getHeight()
    const imgH = (canvas.height * pageW) / canvas.width
    let remaining = imgH
    let offset = 0
    pdf.addImage(imgData, 'JPEG', 0, offset, pageW, imgH)
    remaining -= pageH
    while (remaining > 0) {
      offset -= pageH
      pdf.addPage()
      pdf.addImage(imgData, 'JPEG', 0, offset, pageW, imgH)
      remaining -= pageH
    }
    const selfName = detail?.participants.find(p => p.role === 'self')?.display_name || '我'
    const partnerName = detail?.participants.find(p => p.role === 'partner')?.display_name || '伴侣'
    pdf.save(`缘聚合盘-${selfName}-${partnerName}.pdf`)
  } catch {
    alert('生成 PDF 失败，请稍后重试')
  } finally {
    el.style.display = prevDisplay
    setExportingPDF(false)
  }
}
```

- [ ] **Step 6: 验证 tsc + lint 通过**

Run: `cd frontend && npm run lint && npm run build`
Expected: 全绿。这一步只新增逻辑（state / handler / brand fetch），尚未在 JSX 引用 `CompatibilityShareCard` / `CompatibilityPrintLayout` —— 它们被 import 但未使用，**TS 不会报错（仅 ESLint 可能 warn unused import）**。

> 若 ESLint 报 `'CompatibilityShareCard' is defined but never used` —— 这是 expected，下个 task 就会用到。临时绕过：在 Task 3 commit 信息里加一句声明；不要 disable lint rule。

实际上 ESLint 配置一般会容忍 unused（项目里看 `no-unused-vars` 设为 'off' 或 React 项目用 TS 自带检查）。如果真的报错，把这两行 import 移到下个 task 一起做。

如果 `npm run build` 通过但 lint 报 unused → 把 import 暂时移除，Task 4 时再加回去；提交前再 lint 一次。

实际操作：**先把 `CompatibilityShareCard` / `CompatibilityPrintLayout` import 暂时不加，留到 Task 4 一起加**。本步把 Step 1 中那两行删去。

修订 Step 1: 把：

```tsx
import CompatibilityShareCard from '../components/CompatibilityShareCard'
import CompatibilityPrintLayout from '../components/CompatibilityPrintLayout'
```

**移除**，留到 Task 4 Step 1。

- [ ] **Step 7: Commit**

```bash
git add frontend/src/pages/CompatibilityResultPage.tsx
git commit -m "$(cat <<'EOF'
feat(compat-export): wire state + handlers + brand fetch

Add export-related state (shareModalOpen, savingImage, exportingPDF,
brand, shareCardRef) and three-platform handleSaveImage handler
(iOS Web Share with cancel-silent + Android download + desktop toPng)
plus handleExportPDF (desktop window.print, mobile html2canvas+jsPDF
multi-page tiling). Brand fetched via brandAPI.get() on auth ready.
JSX integration follows in the next commit.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

## Task 4: 合盘页 JSX — 按钮组 + share modal + PrintLayout 挂载

**Files:**
- Modify: `frontend/src/pages/CompatibilityResultPage.tsx`

**Spec 引用:** §3.5、§3.6

- [ ] **Step 1: 加回 import**

在 CompatibilityResultPage.tsx import 区追加（Task 3 移走的两行加回来）：

```tsx
import CompatibilityShareCard from '../components/CompatibilityShareCard'
import CompatibilityPrintLayout from '../components/CompatibilityPrintLayout'
```

- [ ] **Step 2: 添加按钮组到 hero 区域**

定位 CompatibilityResultPage JSX 中页面顶部 hero / 详情卡的区域（约 line 975-1010 一带，含 `<HeartHandshake>` icon、双方姓名、综合分的那块）。在该区域结束前、紧接综合分显示之后，插入：

```tsx
<div className="compat-export-actions">
  <button
    type="button"
    className="btn btn-secondary"
    disabled={!structuredReport}
    onClick={() => setShareModalOpen(true)}
    title={!structuredReport ? '请先生成命理解读' : ''}
  >
    分享图片
  </button>
  <button
    type="button"
    className="btn btn-primary"
    disabled={!structuredReport || exportingPDF}
    onClick={handleExportPDF}
    title={!structuredReport ? '请先生成命理解读' : ''}
  >
    {exportingPDF ? '生成中…' : '导出 PDF'}
  </button>
</div>
```

> **关键约束**：这块 JSX 必须在 `<div className="page">` 或其顶层 wrapper 之**内**，不在 PrintLayout 之内（否则打印时也会显示按钮）。

- [ ] **Step 3: 在 return 末尾追加 share modal + PrintLayout 挂载**

在 CompatibilityResultPage 的 return JSX 的**最外层 wrapper 闭合之前**（通常是最后一个 `</div>` 之前），插入：

```tsx
{shareModalOpen && structuredReport && (
  <div
    className="compat-share-modal"
    role="dialog"
    aria-modal="true"
    aria-label="分享图片预览"
    onClick={() => setShareModalOpen(false)}
    onKeyDown={(e) => { if (e.key === 'Escape') setShareModalOpen(false) }}
  >
    <div className="compat-share-modal-panel" onClick={e => e.stopPropagation()}>
      <header className="compat-share-modal-head">
        <h3>分享图片预览</h3>
        <button
          type="button"
          className="compat-share-modal-close"
          onClick={() => setShareModalOpen(false)}
          aria-label="关闭"
        >×</button>
      </header>
      <div className="compat-share-modal-preview">
        <CompatibilityShareCard
          ref={shareCardRef}
          reading={reading}
          participants={detail.participants}
          evidences={detail.evidences}
          structured={structuredReport}
          decision={decisionDashboard}
          brand={brand}
        />
      </div>
      <footer className="compat-share-modal-footer">
        <button
          type="button"
          className="btn btn-primary"
          onClick={handleSaveImage}
          disabled={savingImage}
        >
          {savingImage ? '生成中…' : isIOS ? '保存 / 分享' : '保存到本地'}
        </button>
      </footer>
    </div>
  </div>
)}

<CompatibilityPrintLayout
  reading={reading}
  participants={detail.participants}
  evidences={detail.evidences}
  decision={decisionDashboard}
  stageRisks={decisionStageRisks}
  structured={structuredReport ?? null}
  brand={brand}
/>
```

- [ ] **Step 4: 验证 ESC 键关闭可达**

在 Step 3 的 `<div className="compat-share-modal">` 上加 `onKeyDown` 已经处理了 Escape。但 div 默认没有焦点，需要确保焦点能聚焦在 modal 上。补一行 `tabIndex={-1}`：

```tsx
<div
  className="compat-share-modal"
  role="dialog"
  aria-modal="true"
  aria-label="分享图片预览"
  tabIndex={-1}
  ref={(el) => { if (el && shareModalOpen) el.focus() }}
  onClick={() => setShareModalOpen(false)}
  onKeyDown={(e) => { if (e.key === 'Escape') setShareModalOpen(false) }}
>
```

> 注：这是临时方案；如果项目里已经有更规范的 modal focus 管理 hook，使用那个；否则上面这样工作。

- [ ] **Step 5: 验证 tsc + lint + build 通过**

Run: `cd frontend && npm run lint && npm run build`
Expected: 全绿。所有新增 JSX 都引用了 Task 3 添加的 state/handler 和 Task 1/2 创建的组件。

- [ ] **Step 6: Commit**

```bash
git add frontend/src/pages/CompatibilityResultPage.tsx
git commit -m "$(cat <<'EOF'
feat(compat-export): add buttons + share modal + print layout mount

Two-button row (分享图片 + 导出 PDF) sits next to the hero block,
both disabled until structuredReport is ready. Share modal opens on
button click, hosts the actual CompatibilityShareCard ref'd for
toPng/toBlob capture, and closes on backdrop/ESC/× — preview uses
real dimensions so screenshots are reliable. CompatibilityPrintLayout
is always mounted at the page bottom (display:none by default), only
visible under @media print.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

## Task 5: 合盘页 CSS — 按钮 / modal / @media print

**Files:**
- Modify: `frontend/src/pages/CompatibilityResultPage.css`

**Spec 引用:** §3.5、§3.6、§3.7

- [ ] **Step 1: 在 CompatibilityResultPage.css 末尾追加按钮组样式**

```css
/* === 导出按钮组 === */
.compat-export-actions {
  display: flex;
  gap: 12px;
  margin: 12px 0 0;
  flex-wrap: wrap;
}
.compat-export-actions .btn {
  min-width: 100px;
}
.compat-export-actions .btn:disabled {
  opacity: 0.55;
  cursor: not-allowed;
}

@media (max-width: 640px) {
  .compat-export-actions {
    width: 100%;
  }
  .compat-export-actions .btn {
    flex: 1 1 auto;
  }
}
```

- [ ] **Step 2: 在 CompatibilityResultPage.css 末尾追加 modal 样式**

```css
/* === 分享卡 Modal === */
.compat-share-modal {
  position: fixed;
  inset: 0;
  background: rgba(20, 12, 4, 0.55);
  z-index: 1200;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 16px;
  outline: none;
}

.compat-share-modal-panel {
  background: #fff;
  border-radius: 14px;
  padding: 16px 16px 14px;
  max-width: 480px;
  width: 100%;
  max-height: 92vh;
  display: flex;
  flex-direction: column;
  box-shadow: 0 16px 48px rgba(0, 0, 0, 0.3);
}

.compat-share-modal-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
}
.compat-share-modal-head h3 {
  margin: 0;
  font-size: 16px;
  font-weight: 700;
  color: var(--text-primary, #2a1a0a);
}
.compat-share-modal-close {
  background: transparent;
  border: 0;
  font-size: 24px;
  line-height: 1;
  color: var(--text-muted, #999);
  cursor: pointer;
  padding: 0 4px;
}
.compat-share-modal-close:hover {
  color: var(--text-primary, #2a1a0a);
}

.compat-share-modal-preview {
  flex: 1 1 auto;
  overflow: auto;
  display: flex;
  justify-content: center;
  padding: 8px 4px;
  background: #efe6cf;
  border-radius: 8px;
}

.compat-share-modal-footer {
  margin-top: 12px;
  display: flex;
  justify-content: flex-end;
}
```

- [ ] **Step 3: 追加 @media print 规则隐藏屏幕内容**

```css
@media print {
  /* 打印时只显示 PrintLayout，藏掉一切屏幕内容 */
  .compatibility-result-page,
  .compat-share-modal,
  .navbar,
  .bottom-nav,
  .particle-background {
    display: none !important;
  }
  /* PrintLayout 自己的 CSS 已经处理 .compat-print-layout 的 display:block */
}
```

> **注意**：选择器 `.compatibility-result-page` 是合盘结果页的根 class。若实际页面 root 不是这个名字，看一眼 CompatibilityResultPage.tsx 顶层 div 用的什么 className 改对应选择器。如果根本没有顶层专属 className，加一个 `<div className="compatibility-result-page">`。

需要先确认页面 root className：

Run: `grep -m 3 'className=' frontend/src/pages/CompatibilityResultPage.tsx | head -5`
Expected output 应该能看到根 div 的 className。如果不是 `compatibility-result-page`，要么：
  - (推荐) 在 CompatibilityResultPage.tsx 把根 div 的 className 改为或追加 `compatibility-result-page`
  - 或在本 CSS 用实际的根 className

实施步骤：先 grep，再确认 className，再写 CSS。

- [ ] **Step 4: 验证 lint + build 通过**

Run: `cd frontend && npm run lint && npm run build`
Expected: 全绿。

- [ ] **Step 5: Commit**

```bash
git add frontend/src/pages/CompatibilityResultPage.css
git commit -m "$(cat <<'EOF'
style(compat-export): button group + share modal + print hides

Button row sits inline with hero, full-width on ≤640px. Share modal
overlays with paper-tone preview background and 92vh max-height
(scrolls when the share card overflows). @media print scopes hide
the compatibility result page, modal overlay, navbar/bottom-nav so
only CompatibilityPrintLayout reaches the printer.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

## Task 6: 手动 smoke per spec §6

**Files:** none (verification only)

**Spec 引用:** §6 全部 25 条验证标准

- [ ] **Step 1: 启动 dev server**

Run: `cd frontend && npm run dev`
Expected: Vite 起在某个端口（项目里通常是 5200），不报错。

> 如果 Vite 报端口被占用，看 `frontend/vite.config.ts` 配置或停掉旧实例。

- [ ] **Step 2: 准备测试数据**

打开浏览器到 `http://localhost:5200`：
1. 登录账号（用户应已有账号；若无，先注册）
2. 进入合盘页面，提交一对样本数据（若已有合盘历史，直接打开任意一条详情页）
3. 在详情页点「生成命理解读」生成 polished AI 报告（等待完成，可能要 30s-3min）

- [ ] **Step 3: 走 spec §6 验证 25 条**

按 spec §6 的清单逐条验证：

**功能 (10 条):**
1. 顶部能看到「分享图片」「导出 PDF」两个按钮
2. 未生成 AI 报告时两按钮 disabled + hover 提示 "请先生成命理解读"
3. 生成 polished 报告后两按钮立刻可点（无需刷新）
4. 点「分享图片」打开 modal；点遮罩 / ESC / × 都能关闭
5. Modal 内 ShareCard 显示双盘四柱、综合分、四维条、3 条核心证据、verdict、brand title/footer
6. 桌面 modal 内「保存到本地」→ 下载 `缘聚合盘-{self}-{partner}.png`
7. iOS Safari 点「保存 / 分享」→ 系统分享面板弹出
8. Android 浏览器点「保存」→ .png 下载
9. 桌面点「导出 PDF」→ 浏览器打印对话框，预览中 §1–§7 全章节，"另存为 PDF" 得到 5–8 页
10. 移动端点「导出 PDF」→ loading「生成中…」→ 几秒后下载 PDF

**视觉/版面 (5 条):**
11. 分享卡 400px 宽，国风纸色 `#fdf9f2`
12. PDF 章节强制分页（每节首页新起）
13. PDF §7 双盘原图 A4 单页内左右分列
14. PDF 每页有 brand title 页眉 + footer
15. `showDiagonalWatermark(brand)` 为真时分享卡和 PDF 都有对角水印

**错误/边界 (5 条):**
16. 未生成报告时 disabled 按钮不响应点击，console 无报错
17. iOS 用户取消系统分享面板不弹 alert
18. 人为破坏 `.compat-print-layout`（DevTools 删 DOM）后点 PDF，看到 `alert('生成 PDF 失败，请稍后重试')` + loading 复位
19. 失败后按钮恢复可点，重试能正常导出
20. Modal 打开时按 Ctrl+P，打印预览里没有 modal 遮罩污染

**性能 (3 条):**
21. 详情页 FCP / TTI 与旧详情页一致（用 Chrome Lighthouse 抽测）
22. Modal 首次打开 < 500ms
23. 移动 PDF < 8s 完成

**回归 (2 条):**
24. 单盘 `ResultPage` 的图片 / PDF 按钮、ShareCard、PrintLayout 行为不变
25. `npm run lint && npm run build` 全绿（在 Step 5 已跑过）

- [ ] **Step 4: 记录不通过项**

如果有任何一条不通过：
1. 不要急着改，先记录现象
2. 把现象贴回 controller，由 controller 决定是回到 Task N 修复还是接受为已知问题
3. 修复后再跑完整 25 条

- [ ] **Step 5: 写 verify commit**

全部 25 条通过后：

```bash
git commit --allow-empty -m "$(cat <<'EOF'
verify(compat-export): manual smoke per spec §6 passed

All 25 verification points from
docs/superpowers/specs/2026-05-28-compatibility-export-design.md §6
walked through on local dev server. Image and PDF export both work
on desktop and mobile paths, print layout pages break correctly,
disabled-state UX matches spec, no regressions on single-chart
ResultPage.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

## Final review

After Task 6 commit, dispatch a final code-reviewer agent with the diff range covering Tasks 1–6 (`HEAD~6..HEAD`) to confirm spec compliance and code quality across the whole feature. Subagent-driven-development workflow handles this automatically.

---

## Self-review pass

After writing this plan, controller should verify:

**Spec coverage check (§ → task)**
- §1 背景 → covered in plan goal
- §2 范围 (in scope) → T1+T2+T3+T4+T5 cover all 6 items; out-of-scope items absent ✓
- §3.1 总体架构 → T3+T4 ✓
- §3.2 复用基建 → all helpers imported in T1/T2/T3 ✓
- §3.3 ShareCard 组件 → T1 ✓
- §3.4 PrintLayout 组件 → T2 ✓
- §3.5 页面集成 → T3+T4 ✓
- §3.6 Modal 行为 → T4 Step 3-4 ✓
- §3.7 PDF 双轨 → T3 Step 5 (handleExportPDF) ✓
- §4 数据流 → T4 prop passing ✓
- §5 错误处理 → covered in handlers (T3 Steps 4-5) ✓
- §6 验证 → T6 ✓
- §7 决策摘要 → no implementation needed
- §8 文件清单 → matches T1-T5 ✓
- §9 不引入的依赖 → none introduced ✓

**Type consistency check**
- `CompatibilityShareCardProps` (T1) signature matches usage in T4: `reading / participants / evidences / structured / decision / brand` ✓
- `CompatibilityPrintLayoutProps` (T2) matches T4: `reading / participants / evidences / decision / stageRisks / structured / brand` ✓
- `shareCardRef: useRef<HTMLDivElement>` (T3 Step 2) matches `forwardRef<HTMLDivElement>` (T1) ✓
- `brand: ExportBrand | null` consistent across T1/T2/T3 ✓

**Placeholder scan**
- No "TBD" / "TODO" / "implement later" in the plan ✓
- All code blocks are complete (no `// ...` ellipses) ✓
- Every step has explicit commands and expected outputs ✓
