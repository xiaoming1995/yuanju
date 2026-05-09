import { useEffect, useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { HeartHandshake } from 'lucide-react'
import { useAuth } from '../contexts/AuthContext'
import { compatibilityAPI, type CompatibilityHistoryItem } from '../lib/api'

const levelText: Record<string, string> = {
  high: '契合度高',
  medium: '有优点也有拉扯',
  low: '磨合成本偏高',
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
    return <div className="page"><div className="container" style={{ paddingTop: 40 }}>加载中...</div></div>
  }

  return (
    <div className="page">
      <div className="container" style={{ paddingBottom: 96 }}>
        <div className="card" style={{ padding: 24, marginTop: 28, marginBottom: 20 }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: 10, marginBottom: 8 }}>
            <HeartHandshake size={24} />
            <h1 className="serif" style={{ fontSize: 30, margin: 0 }}>合盘历史</h1>
          </div>
          <p style={{ margin: 0, color: 'var(--text-muted)' }}>共 {items.length} 条合盘记录</p>
        </div>

        <div style={{ display: 'grid', gap: 14 }}>
          {items.map(item => (
            <Link key={item.id} to={`/compatibility/${item.id}`} className="card" style={{ padding: 18, textDecoration: 'none', color: 'inherit' }}>
              <div style={{ display: 'flex', justifyContent: 'space-between', gap: 12, alignItems: 'center', marginBottom: 8 }}>
                <div className="serif" style={{ fontSize: 20 }}>{item.self_name} × {item.partner_name}</div>
                <div style={{ fontSize: 13, color: 'var(--text-muted)' }}>{levelText[item.overall_level] || item.overall_level}</div>
              </div>
              <div style={{ display: 'flex', flexWrap: 'wrap', gap: 8, marginBottom: 10 }}>
                {item.summary_tags.map(tag => (
                  <span key={tag} style={{ fontSize: 12, padding: '3px 8px', border: '1px solid var(--border-subtle)', borderRadius: 999 }}>{tag}</span>
                ))}
              </div>
              <div style={{ display: 'grid', gridTemplateColumns: 'repeat(2, minmax(0, 1fr))', gap: 8, fontSize: 13, color: 'var(--text-secondary)' }}>
                <span>吸引力 {item.dimension_scores.attraction}</span>
                <span>稳定度 {item.dimension_scores.stability}</span>
                <span>沟通协同 {item.dimension_scores.communication}</span>
                <span>现实磨合 {item.dimension_scores.practicality}</span>
              </div>
            </Link>
          ))}
        </div>
      </div>
    </div>
  )
}
