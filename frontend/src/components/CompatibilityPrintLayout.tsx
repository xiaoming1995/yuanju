import type {
  CompatibilityEvidence,
  CompatibilityParticipant,
  CompatibilityPersonalityComparison,
  CompatibilityPersonalityPortrait,
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
const PERSONALITY_DIM_LABEL: Record<string, string> = {
  expression: '表达 / 沟通',
  decision: '决策与节奏',
  intimacy: '亲密里的核心需求',
  emotion: '情绪反应',
  pressure: '压力下的样子',
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

function PersonalityPrint({ comparison, selfName, partnerName }: {
  comparison: CompatibilityPersonalityComparison
  selfName: string
  partnerName: string
}) {
  const cols: Array<{ name: string; portrait?: CompatibilityPersonalityPortrait }> = [
    { name: selfName, portrait: comparison.self },
    { name: partnerName, portrait: comparison.partner },
  ]
  const fit = (comparison.fit_points || []).filter(p => p && (p.title || p.detail))
  const clash = (comparison.clash_points || []).filter(p => p && (p.title || p.detail))
  return (
    <div className="compat-print-personality">
      <div className="compat-print-personality-grid">
        {cols.map((col, i) => (
          <div key={i} className="compat-print-portrait">
            <div className="compat-print-portrait-name">{col.name}</div>
            {col.portrait?.headline && (
              <div className="compat-print-portrait-headline">{col.portrait.headline}</div>
            )}
            <dl className="compat-print-portrait-dims">
              {(col.portrait?.dimensions || []).map(d => (
                <div key={d.key} className="compat-print-portrait-dim">
                  <dt>{PERSONALITY_DIM_LABEL[d.key] || d.key}</dt>
                  <dd>{d.detail}</dd>
                </div>
              ))}
            </dl>
          </div>
        ))}
      </div>
      {(fit.length > 0 || clash.length > 0) && (
        <div className="compat-print-personality-points">
          {fit.length > 0 && (
            <div className="compat-print-personality-pcol">
              <div className="compat-print-personality-ptitle">自然合的地方</div>
              {fit.map((p, i) => (
                <div key={i} className="compat-print-personality-point">
                  <strong>{p.title}</strong>
                  <span>{p.detail}</span>
                </div>
              ))}
            </div>
          )}
          {clash.length > 0 && (
            <div className="compat-print-personality-pcol">
              <div className="compat-print-personality-ptitle">容易冲突的地方</div>
              {clash.map((p, i) => (
                <div key={i} className="compat-print-personality-point">
                  <strong>{p.title}</strong>
                  <span>{p.detail}</span>
                </div>
              ))}
            </div>
          )}
        </div>
      )}
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
    <div className="compat-print-watermark" aria-hidden="true">
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

          <section className="compat-print-section">
            <h2 className="compat-print-section-title">二、决策仪表盘</h2>
            <DecisionBlock decision={decision} />
          </section>

          <section className="compat-print-section">
            <h2 className="compat-print-section-title">三、评分明细</h2>
            <ScorePrint reading={reading} />
          </section>

          <section className="compat-print-section">
            <h2 className="compat-print-section-title">四、命理证据</h2>
            <EvidenceTable evidences={evidences} />
          </section>

          <section className="compat-print-section">
            <h2 className="compat-print-section-title">五、阶段风险与验证</h2>
            <StageRisksBlock risks={stageRisks} />
          </section>

          {structured && (
            <>
              <section className="compat-print-section">
                <h2 className="compat-print-section-title">六、命理解读</h2>
                {structured.summary && <p className="compat-print-summary">{structured.summary}</p>}
                {structured.personality_comparison &&
                  (structured.personality_comparison.self || structured.personality_comparison.partner) && (
                  <div className="compat-print-chapter">
                    <h4 className="compat-print-chapter-title">双方性格画像与差异</h4>
                    <PersonalityPrint
                      comparison={structured.personality_comparison}
                      selfName={selfP?.display_name || '我'}
                      partnerName={partnerP?.display_name || '伴侣'}
                    />
                  </div>
                )}
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
