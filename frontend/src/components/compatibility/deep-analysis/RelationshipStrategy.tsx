import './RelationshipStrategy.css'
import type { CompatibilityRelationshipStrategy } from '../../../lib/api'

function AdviceList({ title, items }: { title: string; items: string[] }) {
  const safeItems = Array.isArray(items) ? items.filter(Boolean) : []

  return (
    <div className="compatibility-advice-list">
      <div className="compatibility-advice-title">{title}</div>
      {safeItems.length > 0 ? (
        safeItems.map(item => <div key={item}>{item}</div>)
      ) : (
        <div>暂无明确建议</div>
      )}
    </div>
  )
}

export default function RelationshipStrategy({ strategy }: { strategy: CompatibilityRelationshipStrategy }) {
  return (
    <details open className="compat-da-strategy">
      <summary className="compat-da-subsection-summary">关系经营策略</summary>
      <div className="compat-da-strategy__body">
        <AdviceList title="沟通" items={[strategy.communication].filter(Boolean)} />
        <AdviceList title="冲突" items={[strategy.conflict].filter(Boolean)} />
        <AdviceList title="现实" items={[strategy.reality].filter(Boolean)} />
        <AdviceList title="边界" items={[strategy.boundary].filter(Boolean)} />
      </div>
    </details>
  )
}
