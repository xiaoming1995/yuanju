import { useEffect, useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { Compass, HeartHandshake } from 'lucide-react'
import { useAuth } from '../contexts/AuthContext'
import { compatibilityAPI, type CompatibilityHistoryItem } from '../lib/api'
import { getPersonalityMatchType } from '../lib/compatibilityPersonality'
import './CompatibilityHistoryPage.css'

const levelText: Record<string, string> = {
  high: '契合度高',
  medium: '有优点也有拉扯',
  low: '磨合成本偏高',
}

const relationshipStageText: Record<string, string> = {
  ambiguous: '暧昧中',
  dating: '恋爱中',
  long_distance: '异地中',
  reconciliation: '分手/复合中',
  marriage_or_engagement: '谈婚论嫁',
  crush: '单恋/暗恋',
  general: '综合关系判断',
}

const primaryQuestionText: Record<string, string> = {
  continue_investment: '值不值得继续投入',
  marriage_suitability: '适不适合结婚',
  recurring_conflict: '为什么反复拉扯',
  reconciliation_potential: '复合有没有意义',
  long_term_stability: '长期能不能稳定',
  relationship_strategy: '怎么相处更顺',
  general: '综合关系判断',
}

function getHistoryContinuationLabel(item: CompatibilityHistoryItem) {
  if (item.primary_question === 'relationship_strategy') return '继续看性格合盘'
  return '查看合盘 · 查看性格合盘'
}

export default function CompatibilityHistoryPage() {
  const { user, isLoading } = useAuth()
  const navigate = useNavigate()
  const [items, setItems] = useState<CompatibilityHistoryItem[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    if (isLoading) {
      return
    }
    if (!user) {
      navigate('/login')
      return
    }
    compatibilityAPI.getHistory()
      .then(res => setItems(res.data.data || []))
      .finally(() => setLoading(false))
  }, [user, isLoading, navigate])

  if (loading || isLoading) {
    return (
      <div className="compatibility-history-page page">
        <div className="container">
          <div className="compatibility-history-loading">加载中...</div>
        </div>
      </div>
    )
  }

  return (
    <div className="compatibility-history-page page">
      <div className="container">
        <section className="compatibility-history-hero card">
          <div className="compatibility-history-title-row">
            <HeartHandshake size={24} />
            <div>
              <p className="compatibility-history-kicker">合盘档案</p>
              <h1 className="serif">我的合盘档案</h1>
            </div>
          </div>
          <p>共 {items.length} 条合盘记录，回看双方匹配结果、四维评分和完整解读。</p>
        </section>

        <nav className="archive-switcher" aria-label="档案类型">
          <Link to="/history" className="archive-switcher-item">
            <Compass size={17} />
            <span>命盘档案</span>
          </Link>
          <Link to="/compatibility/history" className="archive-switcher-item archive-switcher-item--active">
            <HeartHandshake size={17} />
            <span>合盘档案</span>
          </Link>
        </nav>

        {items.length === 0 ? (
          <div className="compatibility-history-empty card">
            <HeartHandshake size={46} />
            <h2 className="serif">还没有合盘记录</h2>
            <p>创建合盘后，会在这里保存双方匹配结果。</p>
            <Link to="/compatibility" className="btn btn-primary">开始合盘</Link>
          </div>
        ) : (
          <div className="compatibility-history-list">
            {items.map(item => (
              <Link key={item.id} to={`/compatibility/${item.id}`} className="compatibility-history-card card">
                <div className="compatibility-history-card-head">
                  <div>
                    <div className="serif compatibility-history-names">{item.self_name} × {item.partner_name}</div>
                    <div className="compatibility-history-level">{levelText[item.overall_level] || item.overall_level}</div>
                  </div>
                  <span className="compatibility-history-action">查看合盘</span>
                </div>
                <div className="compatibility-history-personality">
                  <span>性格匹配</span>
                  <strong>{getPersonalityMatchType(item.dimension_scores, item.primary_question, item.relationship_stage)}</strong>
                </div>
                <div className="compatibility-history-context-title">关系背景</div>
                <div className="compatibility-history-context">
                  <span>{relationshipStageText[item.relationship_stage] || relationshipStageText.general}</span>
                  <span>{primaryQuestionText[item.primary_question] || primaryQuestionText.general}</span>
                </div>
                <div className="compatibility-history-tags">
                  {item.summary_tags.length > 0 ? item.summary_tags.map(tag => (
                    <span key={tag}>{tag}</span>
                  )) : <span>{levelText[item.overall_level] || item.overall_level}</span>}
                </div>
                <div className="compatibility-history-score-summary">
                  <span>分数参考</span>
                  <strong>
                    吸引 {item.dimension_scores.attraction} · 稳定 {item.dimension_scores.stability} · 沟通 {item.dimension_scores.communication} · 现实 {item.dimension_scores.practicality}
                  </strong>
                </div>
                <div className="compatibility-history-continuation">
                  <span>{getHistoryContinuationLabel(item)}</span>
                </div>
              </Link>
            ))}
          </div>
        )}
      </div>
    </div>
  )
}
