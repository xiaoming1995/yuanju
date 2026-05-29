import './PersonalityComparison.css'
import type {
  CompatibilityPersonalityComparison,
  CompatibilityPersonalityPortrait,
  CompatibilityPersonalityPoint,
} from '../../../lib/api'

// 维度 label 由前端按 key 固定映射，detail 取自 LLM（防 LLM 标签漂移）
const DIMENSION_LABELS: Record<string, string> = {
  expression: '表达 / 沟通',
  decision: '决策与节奏',
  intimacy: '亲密里的核心需求',
  emotion: '情绪反应',
  pressure: '压力下的样子',
}

function PortraitCard({ portrait }: { portrait?: CompatibilityPersonalityPortrait }) {
  const dimensions = Array.isArray(portrait?.dimensions) ? portrait.dimensions : []
  return (
    <div className="compatibility-portrait-card">
      <div className="compatibility-portrait-headline">{portrait?.headline}</div>
      <dl className="compatibility-portrait-dims">
        {dimensions.map(dim => (
          <div key={dim.key} className="compatibility-portrait-dim">
            <dt>{DIMENSION_LABELS[dim.key] || dim.key}</dt>
            <dd>{dim.detail}</dd>
          </div>
        ))}
      </dl>
    </div>
  )
}

function PointList({ title, points }: { title: string; points?: CompatibilityPersonalityPoint[] }) {
  const safe = Array.isArray(points) ? points.filter(p => p && (p.title || p.detail)) : []
  if (safe.length === 0) return null
  return (
    <div className="compatibility-personality-list">
      <div className="compatibility-personality-list-title">{title}</div>
      {safe.map((point, index) => (
        <div key={`${title}-${index}-${point.title}`} className="compatibility-personality-point">
          <strong>{point.title}</strong>
          <p>{point.detail}</p>
        </div>
      ))}
    </div>
  )
}

export default function PersonalityComparison({ comparison }: { comparison?: CompatibilityPersonalityComparison | null }) {
  if (!comparison || (!comparison.self && !comparison.partner)) return null
  return (
    <div className="compatibility-report-section compatibility-personality-comparison">
      <div className="serif compatibility-report-title">双方性格画像与差异</div>
      <div className="compatibility-portrait-grid">
        <PortraitCard portrait={comparison.self} />
        <PortraitCard portrait={comparison.partner} />
      </div>
      <div className="compatibility-personality-grid">
        <PointList title="自然合的地方" points={comparison.fit_points} />
        <PointList title="容易冲突的地方" points={comparison.clash_points} />
      </div>
    </div>
  )
}
