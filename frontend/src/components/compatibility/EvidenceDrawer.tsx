import './EvidenceDrawer.css'
import type {
  CompatibilityEvidence,
  CompatibilityClaimEvidenceLink,
} from '../../lib/api'

const dimensionText: Record<string, string> = {
  attraction: '会不会互相吸引？',
  stability: '能不能长期稳定？',
  communication: '吵架后能不能修复？',
  practicality: '现实条件能不能落地？',
  zodiac: '合属相',
  nayin: '合纳音',
  day_pillar: '合日柱',
  eight_chars: '合八字',
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

const evidenceSourceText: Record<string, string> = {
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

const perspectiveText: Record<string, string> = {
  self_to_partner: '我看对方',
  partner_to_self: '对方看我',
  mutual: '双方互见',
}

function EvidenceCard({ evidence }: { evidence: CompatibilityEvidence }) {
  const badgeColor = polarityColor[evidence.polarity] || 'var(--text-muted)'

  return (
    <div className="card compatibility-evidence-card">
      <div className="compatibility-evidence-header">
        <div className="serif compatibility-evidence-title">{evidence.title}</div>
        <div className="compatibility-evidence-badges">
          <span
            className="compatibility-evidence-badge"
          >
            {dimensionText[evidence.dimension] || evidence.dimension}
          </span>
          {evidence.perspective && (
            <span className="compatibility-evidence-badge">
              {perspectiveText[evidence.perspective] || evidence.perspective}
            </span>
          )}
          <span
            className="compatibility-evidence-badge"
            style={{
              border: `1px solid ${badgeColor}33`,
              color: badgeColor,
              background: `${badgeColor}14`,
            }}
          >
            {polarityText[evidence.polarity] || evidence.polarity}
          </span>
        </div>
      </div>
      <div className="compatibility-evidence-detail">{evidence.detail}</div>
      {Array.isArray(evidence.related_sources) && evidence.related_sources.length > 0 && (
        <div className="compatibility-evidence-related">
          关联：{evidence.related_sources.map(source => evidenceSourceText[source] || source).join(' / ')}
        </div>
      )}
    </div>
  )
}

function groupEvidenceBySource(evidences: CompatibilityEvidence[]) {
  const groups = new Map<string, CompatibilityEvidence[]>()
  evidences.forEach(evidence => {
    const key = evidence.source || 'unknown'
    const items = groups.get(key) || []
    items.push(evidence)
    groups.set(key, items)
  })
  return Array.from(groups.entries())
    .filter(([, items]) => items.length > 0)
    .sort(([a], [b]) => (evidenceSourceText[a] || a).localeCompare(evidenceSourceText[b] || b, 'zh-Hans-CN'))
}

function ProfessionalEvidenceGroups({ evidences }: { evidences: CompatibilityEvidence[] }) {
  const groups = groupEvidenceBySource(evidences)
  if (groups.length === 0) {
    return <p className="compatibility-report-empty">暂无结构化依据。</p>
  }

  return (
    <div className="compatibility-evidence-groups">
      {groups.map(([source, items]) => (
        <section key={source} className="compatibility-evidence-group">
          <div className="compatibility-evidence-group-header">
            <div className="serif compatibility-evidence-group-title">{evidenceSourceText[source] || source}</div>
            <div className="compatibility-evidence-group-count">{items.length} 条</div>
          </div>
          <div className="compatibility-evidence-grid">
            {items.map(evidence => <EvidenceCard key={evidence.id || evidence.evidence_key} evidence={evidence} />)}
          </div>
        </section>
      ))}
    </div>
  )
}

function EvidenceLinkedClaims({
  links,
  evidences,
}: {
  links: CompatibilityClaimEvidenceLink[]
  evidences: CompatibilityEvidence[]
}) {
  const byKey = new Map(evidences.map(evidence => [evidence.evidence_key || evidence.id, evidence]))
  const previewText = (text: string) => text.length > 72 ? `${text.slice(0, 72)}...` : text

  return (
    <div className="compatibility-claim-list">
      {links.map((link, index) => (
        <details key={link.claim_id || link.claim} className="card compatibility-claim-card" open={index === 0}>
          <summary>
            <span className="serif">{link.claim}</span>
            <span className="compatibility-claim-toggle">
              <span className="compatibility-claim-toggle-open">收起依据</span>
              <span className="compatibility-claim-toggle-closed">查看完整依据</span>
            </span>
          </summary>
          <div className="compatibility-claim-preview">{previewText(link.reasoning)}</div>
          <p>{link.reasoning}</p>
          {link.caveat && <p className="compatibility-claim-caveat">{link.caveat}</p>}
          <div className="compatibility-claim-evidence">
            {(link.evidence_keys || []).map(key => {
              const evidence = byKey.get(key)
              return evidence ? <EvidenceCard key={key} evidence={evidence} /> : null
            })}
          </div>
        </details>
      ))}
    </div>
  )
}

type Props = {
  evidences: CompatibilityEvidence[]
  claimEvidenceLinks: CompatibilityClaimEvidenceLink[]
}

export default function EvidenceDrawer({ evidences, claimEvidenceLinks }: Props) {
  return (
    <details open className="compat-evidence-drawer">
      <summary className="compat-evidence-drawer__summary">
        <span className="compat-evidence-drawer__title serif">命理证据 / 命盘细节</span>
        <span className="compat-evidence-drawer__hint">关键判断依据 + 结构化证据组</span>
      </summary>
      <div className="compat-evidence-drawer__body">
        {claimEvidenceLinks.length > 0 && (
          <div className="compat-evidence-drawer__group">
            <h3 className="compat-evidence-drawer__group-title">关键判断依据</h3>
            <EvidenceLinkedClaims links={claimEvidenceLinks} evidences={evidences} />
          </div>
        )}
        <div className="compat-evidence-drawer__group">
          <h3 className="compat-evidence-drawer__group-title">结构化证据组</h3>
          <ProfessionalEvidenceGroups evidences={evidences} />
        </div>
      </div>
    </details>
  )
}
