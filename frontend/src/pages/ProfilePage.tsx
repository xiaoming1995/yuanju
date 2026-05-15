import { useEffect, useState } from 'react'
import { Link, Navigate } from 'react-router-dom'
import { ArrowRight, CalendarDays, CreditCard, FileText, HeartHandshake, History, Sparkles, UserRound } from 'lucide-react'
import { useAuth } from '../contexts/AuthContext'
import { userAPI } from '../lib/api'
import type { UserProfileOverview } from '../lib/api'
import './ProfilePage.css'

function formatDate(value: string) {
  if (!value) return '-'
  return new Date(value).toLocaleDateString('zh-CN')
}

function genderText(gender: string) {
  return gender === 'female' ? '女命' : '男命'
}

export default function ProfilePage() {
  const { user, isLoading } = useAuth()
  const [profile, setProfile] = useState<UserProfileOverview | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  useEffect(() => {
    if (!user) {
      return
    }
    userAPI.profile()
      .then(res => setProfile(res.data.data))
      .catch(err => setError(err.message || '个人中心加载失败'))
      .finally(() => setLoading(false))
  }, [user])

  if (!isLoading && !user) {
    return <Navigate to="/login" replace />
  }

  if (isLoading || loading) {
    return (
      <main className="profile-page container page">
        <div className="profile-loading">正在进入个人中心...</div>
      </main>
    )
  }

  if (error) {
    return (
      <main className="profile-page container page">
        <section className="profile-panel profile-error">
          <h1>个人中心</h1>
          <p>{error}</p>
          <button className="btn btn-primary" onClick={() => window.location.reload()}>重新加载</button>
        </section>
      </main>
    )
  }

  if (!profile) {
    return (
      <main className="profile-page container page">
        <section className="profile-panel profile-empty">
          <h1>个人中心</h1>
          <p>暂时没有账户数据。</p>
        </section>
      </main>
    )
  }

  const stats = [
    { label: '命盘记录', value: profile.stats.chart_count, icon: CalendarDays, to: '/history' },
    { label: 'AI 报告', value: profile.stats.ai_report_count, icon: FileText },
    { label: '合盘记录', value: profile.stats.compatibility_count, icon: HeartHandshake, to: '/compatibility/history' },
  ]
  const latestChart = profile.recent_charts[0]
  const latestCompatibility = profile.recent_compatibility[0]
  const continueTarget = latestChart
    ? {
        label: '最近命盘',
        title: `${latestChart.year_gan}${latestChart.year_zhi} · ${latestChart.month_gan}${latestChart.month_zhi} · ${latestChart.day_gan}${latestChart.day_zhi} · ${latestChart.hour_gan}${latestChart.hour_zhi}`,
        desc: `${latestChart.birth_year}年${latestChart.birth_month}月${latestChart.birth_day}日 ${latestChart.birth_hour}时 · ${genderText(latestChart.gender)}`,
        href: `/history/${latestChart.id}`,
      }
    : latestCompatibility
      ? {
          label: '最近合盘',
          title: `${latestCompatibility.self_name || '我'} 与 ${latestCompatibility.partner_name || '对方'}`,
          desc: latestCompatibility.summary_tags.length > 0 ? latestCompatibility.summary_tags.join(' · ') : latestCompatibility.overall_level,
          href: `/compatibility/${latestCompatibility.id}`,
        }
      : {
          label: '新的分析',
          title: '创建新的八字命盘',
          desc: '填写生辰信息，开始一次新的命理分析',
          href: '/',
        }

  return (
    <main className="profile-page container page">
      <section className="profile-header profile-panel">
        <div className="profile-avatar"><UserRound size={28} /></div>
        <div>
          <h1>{profile.user.nickname || '缘聚用户'}</h1>
          <p>{profile.user.email}</p>
          <span>加入时间：{formatDate(profile.user.created_at)}</span>
        </div>
      </section>

      <section className="profile-workbench profile-panel">
        <div>
          <span className="profile-workbench-kicker">{continueTarget.label}</span>
          <h2>继续上次分析</h2>
          <p>{continueTarget.title}</p>
          <small>{continueTarget.desc}</small>
        </div>
        <Link className="btn btn-primary profile-workbench-action" to={continueTarget.href}>
          继续查看 <ArrowRight size={16} />
        </Link>
      </section>

      <section className="profile-stats">
        {stats.map(item => {
          const Icon = item.icon
          const content = (
            <>
              <Icon size={18} />
              <strong>{item.value}</strong>
              <span>{item.label}</span>
            </>
          )
          return item.to ? (
            <Link className="profile-stat-card profile-stat-card--link" key={item.label} to={item.to}>
              {content}
            </Link>
          ) : (
            <div className="profile-stat-card profile-stat-card--static" key={item.label}>
              {content}
            </div>
          )
        })}
      </section>

      <section className="profile-actions">
        <Link className="profile-action-card" to="/history">
          <History size={20} />
          <div>
            <strong>历史记录</strong>
            <span>查看已保存的命盘与报告</span>
          </div>
        </Link>
        <Link className="profile-action-card" to="/compatibility/history">
          <HeartHandshake size={20} />
          <div>
            <strong>合盘记录</strong>
            <span>回看已生成的合盘分析</span>
          </div>
        </Link>
        <Link className="profile-action-card" to="/">
          <Sparkles size={20} />
          <div>
            <strong>继续测算</strong>
            <span>创建新的八字命盘</span>
          </div>
        </Link>
      </section>

      <section className="profile-grid">
        <div className="profile-panel">
          <div className="profile-section-title">
            <h2>最近命盘</h2>
            <Link to="/history">全部</Link>
          </div>
          {profile.recent_charts.length > 0 ? (
            <div className="profile-list">
              {profile.recent_charts.map(chart => (
                <Link className="profile-list-item" key={chart.id} to={`/history/${chart.id}`}>
                  <strong>{chart.year_gan}{chart.year_zhi} · {chart.month_gan}{chart.month_zhi} · {chart.day_gan}{chart.day_zhi} · {chart.hour_gan}{chart.hour_zhi}</strong>
                  <span>{chart.birth_year}年{chart.birth_month}月{chart.birth_day}日 {chart.birth_hour}时 · {genderText(chart.gender)}</span>
                </Link>
              ))}
            </div>
          ) : (
            <p className="profile-muted">还没有保存命盘。</p>
          )}
        </div>

        <div className="profile-panel">
          <div className="profile-section-title">
            <h2>最近合盘</h2>
            <Link to="/compatibility/history">全部</Link>
          </div>
          {profile.recent_compatibility.length > 0 ? (
            <div className="profile-list">
              {profile.recent_compatibility.map(reading => (
                <Link className="profile-list-item" key={reading.id} to={`/compatibility/${reading.id}`}>
                  <strong>{reading.self_name || '我'} 与 {reading.partner_name || '对方'}</strong>
                  <span>{reading.summary_tags.length > 0 ? reading.summary_tags.join(' · ') : reading.overall_level}</span>
                </Link>
              ))}
            </div>
          ) : (
            <p className="profile-muted">还没有合盘记录。</p>
          )}
        </div>
      </section>

      <section className="profile-panel">
        <div className="profile-section-title">
          <h2>充值点数与 PDF 模板</h2>
        </div>
        <div className="profile-feature-row">
          {profile.features.map(feature => (
            <div className="profile-feature-card" key={feature.key}>
              {feature.key === 'wallet' ? <CreditCard size={22} /> : <FileText size={22} />}
              <div>
                <strong>{feature.title}</strong>
                <span>{feature.description}</span>
              </div>
              <em className="profile-feature-status">即将开放</em>
            </div>
          ))}
        </div>
      </section>
    </main>
  )
}
