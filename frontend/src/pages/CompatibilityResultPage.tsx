import { useEffect, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { HeartHandshake } from 'lucide-react'
import { useAuth } from '../contexts/AuthContext'
import {
  compatibilityAPI,
  type CompatibilityChartSnapshot,
  type CompatibilityDetail,
  type CompatibilityDurationAssessment,
  type CompatibilityEvidence,
  type CompatibilityParticipant,
} from '../lib/api'
import './CompatibilityResultPage.css'

const levelText: Record<string, string> = {
  high: '契合度高',
  medium: '可发展，但需要磨合',
  low: '吸引与稳定存在明显矛盾',
}

const levelSummaryText: Record<string, string> = {
  high: '双方结构中既有吸引也有稳定支点，关系更容易自然推进。',
  medium: '关系有明显可发展空间，但节奏、边界和现实问题需要慢慢磨合。',
  low: '吸引点和拉扯点并存，是否继续深入更依赖双方现实处理能力。',
}

const dimensionText: Record<string, string> = {
  attraction: '吸引力',
  stability: '稳定度',
  communication: '沟通协同',
  practicality: '现实磨合',
}

const polarityText: Record<string, string> = {
  positive: '正向',
  negative: '风险',
  mixed: '复杂',
  neutral: '中性',
}

const polarityColor: Record<string, string> = {
  positive: '#66bb6a',
  negative: '#ef5350',
  mixed: '#ffb74d',
  neutral: 'var(--text-muted)',
}

const durationLevelText: Record<string, string> = {
  high: '偏高',
  medium: '中等',
  low: '偏低',
}

const wuxingLabel = [
  { key: 'mu', label: '木', className: 'wuxing-mu' },
  { key: 'huo', label: '火', className: 'wuxing-huo' },
  { key: 'tu', label: '土', className: 'wuxing-tu' },
  { key: 'jin', label: '金', className: 'wuxing-jin' },
  { key: 'shui', label: '水', className: 'wuxing-shui' },
] as const

function formatBirthText(snapshot?: CompatibilityChartSnapshot | null, fallback?: CompatibilityParticipant['birth_profile']) {
  if (snapshot) {
    return `${snapshot.birth_year}年${snapshot.birth_month}月${snapshot.birth_day}日 ${snapshot.birth_hour}时`
  }
  if (fallback) {
    return `${fallback.year}年${fallback.month}月${fallback.day}日 ${fallback.hour}时`
  }
  return '出生信息缺失'
}

function genderText(snapshot?: CompatibilityChartSnapshot | null, fallback?: CompatibilityParticipant['birth_profile']) {
  const value = snapshot?.gender || fallback?.gender
  return value === 'female' ? '女命' : '男命'
}

function getPillars(snapshot?: CompatibilityChartSnapshot | null) {
  if (!snapshot) return []
  return [
    { label: '年柱', value: `${snapshot.year_gan}${snapshot.year_zhi}` },
    { label: '月柱', value: `${snapshot.month_gan}${snapshot.month_zhi}` },
    { label: '日柱', value: `${snapshot.day_gan}${snapshot.day_zhi}` },
    { label: '时柱', value: `${snapshot.hour_gan}${snapshot.hour_zhi}` },
  ]
}

function getWuxingItems(snapshot?: CompatibilityChartSnapshot | null) {
  return wuxingLabel.map(item => ({
    ...item,
    value: snapshot?.wuxing?.[item.key] ?? 0,
  }))
}

function isDurationLevel(value: unknown): value is 'high' | 'medium' | 'low' {
  return value === 'high' || value === 'medium' || value === 'low'
}

function normalizeDurationAssessment(
  primary: CompatibilityDurationAssessment | null | undefined,
  fallback: CompatibilityDurationAssessment
): CompatibilityDurationAssessment {
  const hasPrimaryValue = Boolean(primary && (
    (typeof primary.summary === 'string' && primary.summary.trim()) ||
    (Array.isArray(primary.reasons) && primary.reasons.length > 0) ||
    isDurationLevel(primary.windows?.three_months?.level) ||
    isDurationLevel(primary.windows?.one_year?.level) ||
    isDurationLevel(primary.windows?.two_years_plus?.level)
  ))

  const source = hasPrimaryValue ? primary! : fallback

  return {
    overall_band: source.overall_band || fallback.overall_band,
    summary: source.summary?.trim() || fallback.summary,
    reasons: Array.isArray(source.reasons) ? source.reasons.filter(Boolean) : fallback.reasons,
    windows: {
      three_months: {
        level: isDurationLevel(source.windows?.three_months?.level) ? source.windows.three_months.level : fallback.windows.three_months.level,
      },
      one_year: {
        level: isDurationLevel(source.windows?.one_year?.level) ? source.windows.one_year.level : fallback.windows.one_year.level,
      },
      two_years_plus: {
        level: isDurationLevel(source.windows?.two_years_plus?.level) ? source.windows.two_years_plus.level : fallback.windows.two_years_plus.level,
      },
    },
  }
}

function EvidenceCard({ evidence }: { evidence: CompatibilityEvidence }) {
  const badgeColor = polarityColor[evidence.polarity] || 'var(--text-muted)'

  return (
    <div className="card" style={{ padding: 16 }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', gap: 12, flexWrap: 'wrap', marginBottom: 8 }}>
        <div className="serif" style={{ fontSize: 18 }}>{evidence.title}</div>
        <div style={{ display: 'flex', gap: 8, flexWrap: 'wrap' }}>
          <span
            style={{
              fontSize: 12,
              padding: '4px 10px',
              borderRadius: 999,
              border: '1px solid var(--border-subtle)',
              color: 'var(--text-secondary)',
              background: 'rgba(255,255,255,0.03)',
            }}
          >
            {dimensionText[evidence.dimension] || evidence.dimension}
          </span>
          <span
            style={{
              fontSize: 12,
              padding: '4px 10px',
              borderRadius: 999,
              border: `1px solid ${badgeColor}33`,
              color: badgeColor,
              background: `${badgeColor}14`,
            }}
          >
            {polarityText[evidence.polarity] || evidence.polarity}
          </span>
        </div>
      </div>
      <div style={{ color: 'var(--text-secondary)', lineHeight: 1.7 }}>{evidence.detail}</div>
    </div>
  )
}

function ParticipantSummaryCard({
  participant,
}: {
  participant: CompatibilityParticipant
}) {
  const snapshot = participant.chart_snapshot || null
  const pillars = getPillars(snapshot)
  const wuxingItems = getWuxingItems(snapshot)

  return (
    <div className="card compatibility-person-card">
      <div className="compatibility-person-header">
        <div>
          <div className="compatibility-person-name serif">{participant.display_name}</div>
          <div className="compatibility-person-meta">
            <span>{genderText(snapshot, participant.birth_profile)}</span>
            <span>{formatBirthText(snapshot, participant.birth_profile)}</span>
          </div>
        </div>
        {snapshot?.day_gan && (
          <div className="compatibility-day-master">
            <span className="compatibility-day-master-label">日主</span>
            <span className="compatibility-day-master-value serif">{snapshot.day_gan}</span>
          </div>
        )}
      </div>

      {pillars.length > 0 && (
        <div className="compatibility-pillar-grid">
          {pillars.map(pillar => (
            <div key={pillar.label} className="compatibility-pillar-cell">
              <div className="compatibility-pillar-label">{pillar.label}</div>
              <div className="compatibility-pillar-value serif">{pillar.value}</div>
            </div>
          ))}
        </div>
      )}

      <div className="compatibility-wuxing-title">五行概览</div>
      <div className="compatibility-wuxing-grid">
        {wuxingItems.map(item => (
          <div key={item.key} className="compatibility-wuxing-item">
            <span className={`wuxing-badge ${item.className}`}>{item.label}</span>
            <span className="compatibility-wuxing-value">{item.value}</span>
          </div>
        ))}
      </div>
    </div>
  )
}

export default function CompatibilityResultPage() {
  const { id } = useParams()
  const { user, isLoading } = useAuth()
  const navigate = useNavigate()
  const [detail, setDetail] = useState<CompatibilityDetail | null>(null)
  const [loading, setLoading] = useState(true)
  const [reportLoading, setReportLoading] = useState(false)
  const [error, setError] = useState('')

  const load = async () => {
    if (!id) return
    const res = await compatibilityAPI.getDetail(id)
    setDetail(res.data.data)
  }

  useEffect(() => {
    if (isLoading) {
      return
    }
    if (!user) {
      navigate('/login')
      return
    }
    load()
      .catch((err: any) => setError(err?.message || '加载失败'))
      .finally(() => setLoading(false))
  }, [id, user, isLoading, navigate])

  const handleGenerateReport = async () => {
    if (!id) return
    setReportLoading(true)
    setError('')
    try {
      await compatibilityAPI.generateReport(id)
      await load()
    } catch (err: any) {
      setError(err?.message || '生成合盘解读失败')
    } finally {
      setReportLoading(false)
    }
  }

  if (loading || isLoading) {
    return <div className="page"><div className="container" style={{ paddingTop: 40 }}>加载中...</div></div>
  }
  if (!detail) {
    return <div className="page"><div className="container" style={{ paddingTop: 40 }}>未找到合盘记录</div></div>
  }

  const reading = detail.reading
  const selfP = detail.participants.find(p => p.role === 'self')
  const partnerP = detail.participants.find(p => p.role === 'partner')
  const structuredReport = detail.latest_report?.content_structured
  const heroSummary = structuredReport?.summary || levelSummaryText[reading.overall_level] || levelText[reading.overall_level]
  const durationAssessment = normalizeDurationAssessment(structuredReport?.duration_assessment, reading.duration_assessment)
  const summaryTags = Array.isArray(reading.summary_tags) ? reading.summary_tags : []
  const reportDimensions = Array.isArray(structuredReport?.dimensions) ? structuredReport.dimensions : []
  const reportRisks = Array.isArray(structuredReport?.risks) ? structuredReport.risks.filter(Boolean) : []

  return (
    <div className="page">
      <div className="container" style={{ paddingBottom: 96 }}>
        <div className="card compatibility-hero-card">
          <div className="compatibility-hero-header">
            <HeartHandshake size={24} />
            <h1 className="serif compatibility-hero-title">
              {selfP?.display_name || '我'} × {partnerP?.display_name || '对方'}
            </h1>
          </div>
          <div className="compatibility-hero-status-row">
            <span className="compatibility-hero-status">{levelText[reading.overall_level] || reading.overall_level}</span>
          </div>
          <p className="compatibility-hero-summary">{heroSummary}</p>
          <div className="compatibility-tag-row">
            {summaryTags.map(tag => (
              <span key={tag} className="compatibility-tag">{tag}</span>
            ))}
          </div>
        </div>

        <div className="compatibility-section">
          <div className="compatibility-section-header">
            <h2 className="serif compatibility-section-title">双方命盘摘要</h2>
            <p className="compatibility-section-desc">先确认双方四柱与命盘核心信息，再看匹配维度和证据。</p>
          </div>
          <div className="compatibility-summary-grid">
            {selfP && <ParticipantSummaryCard participant={selfP} />}
            {partnerP && <ParticipantSummaryCard participant={partnerP} />}
          </div>
        </div>

        <div className="compatibility-section">
          <div className="compatibility-section-header">
            <h2 className="serif compatibility-section-title">四维评分</h2>
            <p className="compatibility-section-desc">从吸引、稳定、沟通和现实磨合四个维度看这段关系的整体结构。</p>
          </div>
          <div className="compatibility-score-grid">
          {[
            ['吸引力', reading.dimension_scores.attraction],
            ['稳定度', reading.dimension_scores.stability],
            ['沟通协同', reading.dimension_scores.communication],
            ['现实磨合', reading.dimension_scores.practicality],
          ].map(([label, value]) => (
            <div key={label} className="card compatibility-score-card">
              <div className="compatibility-score-label">{label}</div>
              <div className="serif compatibility-score-value">{value}</div>
            </div>
          ))}
        </div>
        </div>

        <div className="compatibility-section">
          <div className="compatibility-section-header">
            <h2 className="serif compatibility-section-title">缘分时长评估</h2>
            <p className="compatibility-section-desc">不是断某一天结束，而是看这段关系在不同阶段的维持潜力。</p>
          </div>
          <div className="compatibility-duration-grid">
            {[
              ['3个月', durationAssessment.windows.three_months.level],
              ['1年', durationAssessment.windows.one_year.level],
              ['2年以上', durationAssessment.windows.two_years_plus.level],
            ].map(([label, value]) => (
              <div key={label} className="card compatibility-duration-card">
                <div className="compatibility-duration-label">{label}</div>
                <div className="serif compatibility-duration-value">{durationLevelText[value as string] || value}</div>
              </div>
            ))}
          </div>
          <p className="compatibility-duration-summary">{durationAssessment.summary}</p>
          {durationAssessment.reasons.length > 0 && (
            <div className="compatibility-duration-reasons">
              {durationAssessment.reasons.map(reason => (
                <div key={reason} className="compatibility-duration-reason">{reason}</div>
              ))}
            </div>
          )}
        </div>

        <div className="compatibility-section">
          <div className="compatibility-section-header">
            <h2 className="serif compatibility-section-title">关键依据</h2>
            <p className="compatibility-section-desc">这些结构化证据是合盘结论的主要命理依据。</p>
          </div>
          <div style={{ display: 'grid', gap: 12 }}>
          {detail.evidences.map(evidence => <EvidenceCard key={evidence.id} evidence={evidence} />)}
        </div>
        </div>

        <div className="card compatibility-ai-card">
          <div className="compatibility-ai-header">
            <h2 className="serif compatibility-section-title" style={{ margin: 0 }}>合盘解读</h2>
            {!detail.latest_report && (
              <button className="btn btn-primary" onClick={handleGenerateReport} disabled={reportLoading}>
                {reportLoading ? '生成中...' : '生成解读'}
              </button>
            )}
          </div>

          {error && <p style={{ color: '#e77' }}>{error}</p>}

          {structuredReport ? (
            <div style={{ display: 'grid', gap: 12 }}>
              <p style={{ margin: 0, lineHeight: 1.8 }}>{structuredReport.summary}</p>
              {reportDimensions.map(item => (
                <div key={item.key}>
                  <div className="serif" style={{ fontSize: 18, marginBottom: 6 }}>{item.title}</div>
                  <div style={{ color: 'var(--text-secondary)', lineHeight: 1.75 }}>{item.content}</div>
                </div>
              ))}
              {reportRisks.length > 0 && (
                <div>
                  <div className="serif" style={{ fontSize: 18, marginBottom: 6 }}>风险点</div>
                  <ul style={{ margin: 0, paddingLeft: 18, color: 'var(--text-secondary)' }}>
                    {reportRisks.map(risk => <li key={risk}>{risk}</li>)}
                  </ul>
                </div>
              )}
              <div>
                <div className="serif" style={{ fontSize: 18, marginBottom: 6 }}>建议</div>
                <div style={{ color: 'var(--text-secondary)', lineHeight: 1.75 }}>{structuredReport.advice}</div>
              </div>
            </div>
          ) : detail.latest_report ? (
            <div style={{ whiteSpace: 'pre-wrap', lineHeight: 1.8 }}>{detail.latest_report.content}</div>
          ) : (
            <p style={{ margin: 0, color: 'var(--text-muted)' }}>尚未生成合盘解读。</p>
          )}
        </div>
      </div>
    </div>
  )
}
