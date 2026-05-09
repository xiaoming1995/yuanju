import { useEffect, useState } from 'react'
import { useNavigate, Link } from 'react-router-dom'
import { Compass } from 'lucide-react'
import { useAuth } from '../contexts/AuthContext'
import { baziAPI } from '../lib/api'
import './HistoryPage.css'

interface Chart {
  id: string
  birth_year: number; birth_month: number; birth_day: number; birth_hour: number
  gender: string
  year_gan: string; year_zhi: string; month_gan: string; month_zhi: string
  day_gan: string; day_zhi: string; hour_gan: string; hour_zhi: string
  created_at: string
}

export default function HistoryPage() {
  const { user, isLoading } = useAuth()
  const navigate = useNavigate()
  const [charts, setCharts] = useState<Chart[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    if (isLoading) return
    if (!user) { navigate('/login'); return }
    baziAPI.getHistory()
      .then(res => setCharts(res.data.charts || []))
      .catch(() => {})
      .finally(() => setLoading(false))
  }, [user, isLoading, navigate])

  if (loading || isLoading) return (
    <div className="history-page page">
      <div className="container">
        <div style={{ paddingTop: 40 }}>
          {[1,2,3].map(i => <div key={i} className="skeleton" style={{ height: 96, borderRadius: 12, marginBottom: 12 }} />)}
        </div>
      </div>
    </div>
  )

  return (
    <div className="history-page page">
      <div className="container">
        <div className="history-header animate-fade-up">
          <h1 className="history-title serif">命盘历史</h1>
          <p className="history-desc">共 {charts.length} 份记录</p>
          <div style={{ marginTop: 10 }}>
            <Link to="/compatibility/history" className="btn btn-ghost btn-sm">查看合盘历史</Link>
          </div>
        </div>

        {charts.length === 0 ? (
          <div className="history-empty card animate-fade-up">
            <div className="empty-icon"><Compass size={48} style={{ opacity: 0.5 }} /></div>
            <p className="empty-title serif">还没有命盘记录</p>
            <p className="empty-desc">起盘后登录即可自动保存</p>
            <Link to="/" className="btn btn-primary">立即起盘</Link>
          </div>
        ) : (
          <div className="history-list animate-fade-up">
            {charts.map(c => (
              <Link key={c.id} to={`/history/${c.id}`} className="history-item card">
                <div className="history-pillars serif">
                  {c.year_gan}{c.year_zhi}·{c.month_gan}{c.month_zhi}·{c.day_gan}{c.day_zhi}·{c.hour_gan}{c.hour_zhi}
                </div>
                <div className="history-meta">
                  <span>{c.birth_year}年{c.birth_month}月{c.birth_day}日 {c.birth_hour}时</span>
                  <span>{c.gender === 'male' ? '男命' : '女命'}</span>
                  <span className="history-date">{new Date(c.created_at).toLocaleDateString('zh-CN')}</span>
                </div>
              </Link>
            ))}
          </div>
        )}
      </div>
    </div>
  )
}
