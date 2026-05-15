import { useEffect, useState } from 'react'
import { useNavigate, Link } from 'react-router-dom'
import { CalendarDays, Compass, HeartHandshake, Sparkles } from 'lucide-react'
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

function formatDate(value: string) {
  if (!value) return '-'
  return new Date(value).toLocaleDateString('zh-CN')
}

function genderText(gender: string) {
  return gender === 'female' ? '女命' : '男命'
}

function formatPillars(chart: Chart) {
  return `${chart.year_gan}${chart.year_zhi} · ${chart.month_gan}${chart.month_zhi} · ${chart.day_gan}${chart.day_zhi} · ${chart.hour_gan}${chart.hour_zhi}`
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

  const latestChart = charts[0]
  const maleCount = charts.filter(chart => chart.gender === 'male').length
  const femaleCount = charts.filter(chart => chart.gender === 'female').length
  const stats = [
    { label: '命盘记录', value: charts.length ? `${charts.length}` : '0', icon: CalendarDays },
    { label: '最近起盘', value: latestChart ? formatDate(latestChart.created_at) : '-', icon: Sparkles },
    { label: '男女命', value: `${maleCount}/${femaleCount}`, icon: Compass },
  ]

  if (loading || isLoading) return (
    <div className="history-page page">
      <div className="container">
        <div className="history-loading">
          {[1,2,3].map(i => <div key={i} className="skeleton history-skeleton" />)}
        </div>
      </div>
    </div>
  )

  return (
    <div className="history-page page">
      <div className="container">
        <section className="history-archive-hero card animate-fade-up">
          <div>
            <p className="history-kicker">命盘档案</p>
            <h1 className="history-title serif">我的命盘档案</h1>
            <p className="history-desc">已保存 {charts.length} 份命盘，回看四柱、出生信息和后续分析入口。</p>
          </div>
          <Link to="/" className="btn btn-primary">新建命盘</Link>
        </section>

        <nav className="archive-switcher animate-fade-up" aria-label="档案类型">
          <Link to="/history" className="archive-switcher-item archive-switcher-item--active">
            <Compass size={17} />
            <span>命盘档案</span>
          </Link>
          <Link to="/compatibility/history" className="archive-switcher-item">
            <HeartHandshake size={17} />
            <span>合盘档案</span>
          </Link>
        </nav>

        <section className="history-stat-grid animate-fade-up">
          {stats.map(item => {
            const Icon = item.icon
            return (
              <div className="history-stat-card" key={item.label}>
                <Icon size={18} />
                <span>{item.label}</span>
                <strong>{item.value}</strong>
              </div>
            )
          })}
        </section>

        <div className="history-section-title animate-fade-up">
          <div>
            <h2 className="serif">最近命盘</h2>
            <p>按保存时间排列，点击可查看完整命盘。</p>
          </div>
        </div>

        {charts.length === 0 ? (
          <div className="history-empty card animate-fade-up">
            <div className="empty-icon"><Compass size={48} /></div>
            <p className="empty-title serif">还没有命盘记录</p>
            <p className="empty-desc">起盘后登录即可自动保存</p>
            <Link to="/" className="btn btn-primary">立即起盘</Link>
          </div>
        ) : (
          <div className="history-list animate-fade-up">
            {charts.map(c => (
              <Link key={c.id} to={`/history/${c.id}`} className="history-record-card card">
                <div className="history-record-main">
                  <div>
                    <div className="history-pillars serif">{formatPillars(c)}</div>
                    <div className="history-meta">
                      <span>{c.birth_year}年{c.birth_month}月{c.birth_day}日 {c.birth_hour}时</span>
                      <span>{genderText(c.gender)}</span>
                    </div>
                  </div>
                  <span className="history-record-action">查看命盘</span>
                </div>
                <div className="history-record-footer">
                  <span>保存于 {formatDate(c.created_at)}</span>
                  <span>四柱档案</span>
                </div>
              </Link>
            ))}
          </div>
        )}
      </div>
    </div>
  )
}
