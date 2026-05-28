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
