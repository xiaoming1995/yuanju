import './SectionVerdict.css'
import type {
  CompatibilityDimensionScoresLegacy,
  CompatibilityDimensionScoresV3,
} from '../../lib/api'
import type { DecisionFinding, DecisionDashboardData } from '../../lib/compatibilityDecision'

type Props = {
  dashboard: DecisionDashboardData
  isV3: boolean
  v3Scores: CompatibilityDimensionScoresV3 | null
  legacyScores: CompatibilityDimensionScoresLegacy | null
  overallScore: number
  overallLevel: string
  findings: DecisionFinding[]
}

export default function SectionVerdict({
  dashboard, isV3, v3Scores, legacyScores, overallScore, overallLevel, findings,
}: Props) {
  return (
    <section id="compat-section-verdict" className="compat-section-verdict">
      <div className="compat-section-verdict__head">
        <div className="compat-section-kicker">SECTION 02</div>
        <h2 className="serif compat-section-title">是否合</h2>
      </div>

      <div className="compat-section-verdict__columns">
        <div className="compat-section-verdict__main">
          <div className="compat-section-verdict__verdict-line serif">{dashboard.verdict}</div>
          <div className="compat-section-verdict__type">{dashboard.relationshipType}</div>

          {isV3 && v3Scores && (
            <InlineScoreOverviewV3 scores={v3Scores} overallScore={overallScore} overallLevel={overallLevel} />
          )}
          {!isV3 && legacyScores && (
            <InlineScoreOverviewLegacy scores={legacyScores} overallScore={overallScore} />
          )}
        </div>

        <div className="compat-section-verdict__findings">
          <div className="compat-section-verdict__findings-title">为什么这么判断</div>
          {findings.length === 0 ? (
            <p className="compat-section-verdict__findings-empty">暂无结构化判断要点。</p>
          ) : (
            <ol className="compat-section-verdict__findings-list">
              {findings.slice(0, 5).map((f, i) => (
                <li key={`${f.text}-${i}`}>
                  <span className="compat-section-verdict__findings-index">{i + 1}</span>
                  <span>{f.text}</span>
                </li>
              ))}
            </ol>
          )}
        </div>
      </div>
    </section>
  )
}

// 临时占位组件：依赖 Task 12 把 ScoreOverview 抽出到独立文件。在 Task 12 完成之前，
// SectionVerdict 不会被任何调用方实际渲染（Task 13 才会接到主页面），所以下面的 throw
// 不会在生产代码路径里触发。Task 12 会用真实的 import 替换这两个占位。
function InlineScoreOverviewV3(_props: { scores: CompatibilityDimensionScoresV3; overallScore: number; overallLevel: string }): never {
  throw new Error('SectionVerdict requires Task 12 (extract ScoreOverview module) to be completed first')
}
function InlineScoreOverviewLegacy(_props: { scores: CompatibilityDimensionScoresLegacy; overallScore: number }): never {
  throw new Error('SectionVerdict requires Task 12 (extract ScoreOverview module) to be completed first')
}
