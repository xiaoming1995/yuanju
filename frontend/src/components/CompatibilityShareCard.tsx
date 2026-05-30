import { forwardRef } from 'react'
import type {
  CompatibilityEvidence,
  CompatibilityParticipant,
  CompatibilityReading,
  CompatibilityStageRisk,
  CompatibilityStructuredReport,
  CompatibilityRelationshipStrategy,
  ExportBrand,
} from '../lib/api'
import { isV3DimensionScores } from '../lib/api'
import { getDayPillarPortrait } from '../lib/dayPillarPortraits'
import type { DecisionDashboardData } from '../lib/compatibilityDecision'
import { resolveFooter, showDiagonalWatermark } from '../lib/brandText'
import { cleanReportText } from '../lib/reportText'
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

const STRATEGY_LABEL: Record<string, string> = {
  communication: '沟通',
  conflict: '冲突',
  reality: '现实',
  boundary: '边界',
}

const PERSONALITY_DIM_LABEL: Record<string, string> = {
  expression: '表达 / 沟通',
  decision: '决策与节奏',
  intimacy: '亲密里的核心需求',
  emotion: '情绪反应',
  pressure: '压力下的样子',
}

function firstSentence(content: string): string {
  const cleaned = cleanReportText(content)
  const match = cleaned.match(/^[^。！？]+[。！？]?/)
  return (match?.[0] || cleaned).slice(0, 60).trim()
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
    <div className="compat-share-watermark" aria-hidden="true">
      {items.map(i => <span key={i}>{text}</span>)}
    </div>
  )
}

const STAGE_WINDOW_LABEL: Record<string, string> = {
  three_months: '3 个月',
  one_year: '1 年',
  two_years_plus: '2 年+',
}
const RISK_LEVEL_LABEL: Record<string, string> = {
  high: '偏高',
  medium: '中等',
  low: '偏低',
}

export interface CompatibilityShareCardProps {
  reading: CompatibilityReading
  participants: CompatibilityParticipant[]
  evidences: CompatibilityEvidence[]
  decision: DecisionDashboardData
  stageRisks: CompatibilityStageRisk[]
  structured: CompatibilityStructuredReport | null
  relationshipStrategy?: CompatibilityRelationshipStrategy
  brand?: ExportBrand | null
}

const CompatibilityShareCard = forwardRef<HTMLDivElement, CompatibilityShareCardProps>((props, ref) => {
  const { reading, participants, evidences, decision, stageRisks, structured, relationshipStrategy, brand } = props
  const selfP = participants.find(p => p.role === 'self')
  const partnerP = participants.find(p => p.role === 'partner')

  if (!selfP || !partnerP) {
    return (
      <div ref={ref} className="compat-share-card">
        <header className="compat-share-header">
          <h1>{brand?.title || '缘 聚 合 盘'}</h1>
        </header>
        <p className="compat-share-col-empty" style={{ marginTop: 24, fontSize: 13 }}>双盘数据缺失</p>
      </div>
    )
  }

  const topEvidences = [...evidences]
    .filter(e => Number.isFinite(e.weight))
    .sort((a, b) => b.weight - a.weight)
    .slice(0, 6)

  const isV3 = (reading.analysis_version === 'v3' || reading.analysis_version === 'v3.1')
    && isV3DimensionScores(reading.dimension_scores)
  const dimLabel = isV3 ? DIM_LABEL_V3 : DIM_LABEL_LEGACY
  const dimEntries = Object.entries(reading.dimension_scores)
    .filter(([k]) => dimLabel[k])
    .map(([k, v]) => ({ key: k, label: dimLabel[k], value: typeof v === 'number' ? v : 0 }))

  const explByDim = Object.fromEntries(
    (reading.score_explanations || []).map(e => [e.dimension, e])
  )
  const topFindings = (structured?.relationship_diagnosis?.top_findings || []).slice(0, 3)
  const strategy = relationshipStrategy
  const strategyEntries = strategy ? [
    { key: 'communication', label: STRATEGY_LABEL.communication, value: strategy.communication },
    { key: 'conflict',      label: STRATEGY_LABEL.conflict,      value: strategy.conflict },
    { key: 'reality',       label: STRATEGY_LABEL.reality,       value: strategy.reality },
    { key: 'boundary',      label: STRATEGY_LABEL.boundary,      value: strategy.boundary },
  ].filter(e => e.value?.trim()) : []
  const dimDigests = (structured?.dimensions || []).map(d => ({
    key: d.key,
    title: d.title,
    digest: firstSentence(d.content),
  }))
  const personality = structured?.personality_comparison
  const personalityCols = personality && (personality.self || personality.partner)
    ? [
        { name: selfP.display_name || '我', portrait: personality.self },
        { name: partnerP.display_name || '伴侣', portrait: personality.partner },
      ]
    : []
  const dayPillarCols = [
    { label: '我', p: selfP },
    { label: '伴侣', p: partnerP },
  ]
    .map(({ label, p }) => {
      const snap = p?.chart_snapshot
      const portrait = snap ? getDayPillarPortrait(snap.day_gan, snap.day_zhi) : undefined
      return snap && portrait ? { name: p.display_name || label, gz: `${snap.day_gan}${snap.day_zhi}`, portrait } : null
    })
    .filter((x): x is NonNullable<typeof x> => x !== null)

  const resolvedTitle = brand?.title || '缘 聚 合 盘'
  const resolvedFooter = resolveFooter(brand, 'yuanju.com')
  const showWatermark = showDiagonalWatermark(brand)

  return (
    <div ref={ref} className="compat-share-card">
      <style>{`@import url('${PILLAR_FONT_URL}');`}</style>
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
          {dimEntries.map(d => {
            const expl = explByDim[d.key]
            const pos = expl?.positive_factor?.trim()
            const neg = expl?.negative_factor?.trim()
            const showWhy = Boolean(pos || neg)
            return (
              <div key={d.key} className="compat-share-dim-block">
                <div className="compat-share-dim-row">
                  <span className="compat-share-dim-lbl">{d.label}</span>
                  <div className="compat-share-dim-bar">
                    <i style={{ width: `${Math.max(0, Math.min(100, d.value))}%` }} />
                  </div>
                  <span className="compat-share-dim-val">{d.value}</span>
                </div>
                {showWhy && (
                  <div className="compat-share-dim-why">
                    {pos && <span className="compat-share-dim-why-pos">正：{pos}</span>}
                    {pos && neg && <span className="compat-share-dim-why-sep"> · </span>}
                    {neg && <span className="compat-share-dim-why-neg">负：{neg}</span>}
                  </div>
                )}
              </div>
            )
          })}
        </section>
      )}

      {/* §5 关系决策 */}
      <section className="compat-share-decision">
        <h3 className="compat-share-section-h">◇ 关系决策</h3>
        <div className="compat-share-decision-verdict">{decision.verdict}</div>
        <div className="compat-share-decision-row"><span className="lbl">关系定调</span>{decision.relationshipType}</div>
        <div className="compat-share-decision-row"><span className="lbl">推进建议</span>{decision.recommendationLabel} · 信心{decision.confidenceLabel}</div>
        <div className="compat-share-decision-row"><span className="lbl">最大风险</span>{decision.maxRisk}</div>
        <div className="compat-share-decision-row"><span className="lbl">下一步</span>{decision.nextAction}</div>
      </section>

      {topFindings.length > 0 && (
        <section className="compat-share-findings">
          <h3 className="compat-share-section-h">◇ 关系诊断 · 核心发现</h3>
          {topFindings.map((f, i) => (
            <div key={i} className="compat-share-finding-item">· {f.text}</div>
          ))}
        </section>
      )}

      {personalityCols.length > 0 && (
        <section className="compat-share-personality">
          <h3 className="compat-share-section-h">◇ 双方性格</h3>
          <div className="compat-share-personality-grid">
            {personalityCols.map((col, i) => (
              <div key={i} className="compat-share-personality-col">
                <div className="compat-share-personality-name">{col.name}</div>
                {col.portrait?.headline && (
                  <div className="compat-share-personality-headline">{col.portrait.headline}</div>
                )}
                {(col.portrait?.dimensions || []).map(d => (
                  <div key={d.key} className="compat-share-personality-dim">
                    <span className="compat-share-personality-dim-lbl">{PERSONALITY_DIM_LABEL[d.key] || d.key}</span>
                    <span className="compat-share-personality-dim-val">{d.detail}</span>
                  </div>
                ))}
              </div>
            ))}
          </div>
        </section>
      )}

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

      {topEvidences.length > 0 && (
        <section className="compat-share-evs">
          <h3 className="compat-share-section-h">◇ 核心证据</h3>
          {topEvidences.map(e => <EvidenceItem key={e.id} evidence={e} />)}
        </section>
      )}

      {/* §7 阶段风险 */}
      {stageRisks.length > 0 && (
        <section className="compat-share-stages">
          <h3 className="compat-share-section-h">◇ 阶段风险</h3>
          {stageRisks.slice(0, 3).map((r, i) => (
            <div key={i} className="compat-share-stage">
              <span className="compat-share-stage-win">{STAGE_WINDOW_LABEL[r.window] || r.window}</span>
              <span className={`compat-share-stage-lvl lvl-${r.risk_level}`}>{RISK_LEVEL_LABEL[r.risk_level] || r.risk_level}</span>
              <span className="compat-share-stage-risk">{r.main_risk}</span>
            </div>
          ))}
        </section>
      )}

      {/* §8 避免 */}
      {decision.avoid.length > 0 && (
        <section className="compat-share-avoid">
          <h3 className="compat-share-section-h">⚠ 避免</h3>
          {decision.avoid.slice(0, 2).map((a, i) => (
            <div key={i} className="compat-share-avoid-item">· {a}</div>
          ))}
        </section>
      )}

      {strategyEntries.length > 0 && (
        <section className="compat-share-strategy">
          <h3 className="compat-share-section-h">◇ 关系策略</h3>
          {strategyEntries.map(e => (
            <div key={e.key} className="compat-share-strategy-row">
              <span className="compat-share-strategy-lbl">{e.label}</span>
              <span className="compat-share-strategy-val">{e.value}</span>
            </div>
          ))}
        </section>
      )}

      {dimDigests.length > 0 && (
        <section className="compat-share-digests">
          <h3 className="compat-share-section-h">◇ 命理解读摘要</h3>
          {dimDigests.map(d => (
            <div key={d.key} className="compat-share-digest-item">
              <div className="compat-share-digest-title">• {d.title}</div>
              <div className="compat-share-digest-line">{d.digest}</div>
            </div>
          ))}
        </section>
      )}

      <footer className="compat-share-footer">{resolvedFooter}</footer>
    </div>
  )
})

CompatibilityShareCard.displayName = 'CompatibilityShareCard'
export default CompatibilityShareCard
