import type { CompatibilityParticipant, CompatibilityChartSnapshot } from '../../lib/api'

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

export default function ParticipantSummaryCard({
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
