import './ScoreOverview.css'
import type {
  CompatibilityDimensionScoresLegacy,
  CompatibilityDimensionScoresV3,
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

const dimensionHint: Record<keyof CompatibilityDimensionScoresLegacy, string> = {
  attraction: '初期靠近感与彼此牵引',
  stability: '长期承接与持续投入',
  communication: '冲突后的理解和修复',
  practicality: '现实安排、责任和节奏',
}

function clampScore(value: number) {
  return Math.max(0, Math.min(100, Math.round(value)))
}

function scoreTone(value: number) {
  if (value >= 78) return 'high'
  if (value >= 62) return 'medium'
  return 'low'
}

function getDimensionItems(scores: CompatibilityDimensionScoresLegacy) {
  return ([
    ['attraction', scores.attraction],
    ['stability', scores.stability],
    ['communication', scores.communication],
    ['practicality', scores.practicality],
  ] as Array<[keyof CompatibilityDimensionScoresLegacy, number]>).map(([key, value]) => ({
    key,
    label: dimensionText[key],
    hint: dimensionHint[key],
    value: clampScore(value),
    tone: scoreTone(value),
  }))
}

const dimensionHintV3: Record<keyof CompatibilityDimensionScoresV3, string> = {
  zodiac: '属相（年支）：六合/三合 50、五行同（双生）30、五行相生 20',
  nayin: '纳音五行：相生/相同 命中即满分 20',
  day_pillar: '日柱（亲密层）：支六合/三合 + 干合/生 10、支六合/三合 5、支五行同/相生 3',
  eight_chars: '年/月/时三柱：按日柱规则评分，最高 20',
}

const dimensionLabelV3: Record<keyof CompatibilityDimensionScoresV3, string> = {
  zodiac: '合属相',
  nayin: '合纳音',
  day_pillar: '合日柱',
  eight_chars: '合八字',
}

const dimensionMaxV3: Record<keyof CompatibilityDimensionScoresV3, number> = {
  zodiac: 50,
  nayin: 20,
  day_pillar: 10,
  eight_chars: 20,
}

export function ScoreOverviewV3({
  scores,
  overallScore,
  overallLevel,
}: {
  scores: CompatibilityDimensionScoresV3
  overallScore: number
  overallLevel: 'high' | 'medium' | 'low'
}) {
  const keys: Array<keyof CompatibilityDimensionScoresV3> = [
    'zodiac',
    'nayin',
    'day_pillar',
    'eight_chars',
  ]
  return (
    <section className="compat-score-v3">
      <header className="compat-score-v3__header">
        <span className="compat-score-v3__total">{overallScore}</span>
        <span className="compat-score-v3__unit">/100</span>
        <span className={`compat-score-v3__badge compat-score-v3__badge--${overallLevel}`}>
          {overallLevel === 'high' ? '上吉' : overallLevel === 'medium' ? '中' : '低'}
        </span>
      </header>
      <ul className="compat-score-v3__modules">
        {keys.map((key) => {
          const value = scores[key]
          const max = dimensionMaxV3[key]
          return (
            <li key={key} className="compat-score-v3__module">
              <div className="compat-score-v3__module-row">
                <span className="compat-score-v3__module-label">{dimensionLabelV3[key]}</span>
                <span className="compat-score-v3__module-value">
                  {value}<span className="compat-score-v3__module-max">/{max}</span>
                </span>
              </div>
              <div className="compat-score-v3__module-hint">{dimensionHintV3[key]}</div>
              <div className="compat-score-v3__module-bar">
                <div
                  className="compat-score-v3__module-bar-fill"
                  style={{ width: `${(value / max) * 100}%` }}
                />
              </div>
            </li>
          )
        })}
      </ul>
    </section>
  )
}

export function ScoreOverview({ scores }: { scores: CompatibilityDimensionScoresLegacy }) {
  return (
    <div className="card compatibility-quick-score">
      <div className="compatibility-section-header compatibility-section-header--stacked">
        <h2 className="serif compatibility-section-title">关系速览</h2>
        <p className="compatibility-section-desc">先看四个关键维度的强弱，再展开专业依据。</p>
      </div>
      <div className="compatibility-quick-score-list">
        {getDimensionItems(scores).map(item => (
          <div key={item.key} className={`compatibility-quick-score-row compatibility-quick-score-row--${item.tone}`}>
            <div className="compatibility-quick-score-copy">
              <div className="compatibility-quick-score-label">{item.label}</div>
              <div className="compatibility-quick-score-hint">{item.hint}</div>
            </div>
            <div className="compatibility-quick-score-meter">
              <div className="compatibility-quick-score-value serif">{item.value}</div>
              <div className="compatibility-quick-score-bar" aria-hidden="true">
                <div className="compatibility-quick-score-fill" style={{ width: `${item.value}%` }} />
              </div>
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}
