import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { HeartHandshake } from 'lucide-react'
import { useAuth } from '../contexts/AuthContext'
import { compatibilityAPI, type CompatibilityProfileInput } from '../lib/api'
import BirthProfileForm from '../components/BirthProfileForm'
import { initialBirthProfile, type BirthProfileFormValue } from '../components/birthProfile'
import './CompatibilityPage.css'

function toCompatibilityProfileInput(value: BirthProfileFormValue): CompatibilityProfileInput {
  return {
    year: value.year,
    month: value.month,
    day: value.day,
    hour: value.hour,
    gender: value.gender,
    calendar_type: value.calendarType,
    is_leap_month: value.isLeapMonth,
  }
}

export default function CompatibilityPage() {
  const navigate = useNavigate()
  const { user } = useAuth()
  const [selfProfile, setSelfProfile] = useState<BirthProfileFormValue>(initialBirthProfile('male'))
  const [partnerProfile, setPartnerProfile] = useState<BirthProfileFormValue>(initialBirthProfile('female'))
  const [activeProfile, setActiveProfile] = useState<'self' | 'partner'>('self')
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState('')

  const handleSubmit = async () => {
    if (!user) {
      navigate('/login')
      return
    }
    setSubmitting(true)
    setError('')
    try {
      const { data } = await compatibilityAPI.createReading({
        self: toCompatibilityProfileInput(selfProfile),
        partner: toCompatibilityProfileInput(partnerProfile),
      })
      navigate(`/compatibility/${data.data.reading.id}`)
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : '创建合盘失败')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <div className="page compatibility-page">
      <div className="container compatibility-container">
        <div className="card compatibility-intro">
          <div className="compatibility-intro-title">
            <HeartHandshake size={24} />
            <h1 className="serif">姻缘合盘</h1>
          </div>
          <p>
            第一版采用双盘结构化分析：吸引力、稳定度、沟通协同、现实磨合四维结果，生成完整解读。
          </p>
        </div>

        <div className="compatibility-mobile-tabs" role="tablist" aria-label="合盘资料">
          <button
            type="button"
            className={`compatibility-mobile-tab ${activeProfile === 'self' ? 'active' : ''}`}
            onClick={() => setActiveProfile('self')}
          >
            我的生辰
          </button>
          <button
            type="button"
            className={`compatibility-mobile-tab ${activeProfile === 'partner' ? 'active' : ''}`}
            onClick={() => setActiveProfile('partner')}
          >
            对方生辰
          </button>
        </div>

        <div className="compatibility-forms">
          <div className={`card compatibility-profile-panel ${activeProfile === 'self' ? 'compatibility-profile-panel--active' : ''}`}>
            <BirthProfileForm title="我的生辰" value={selfProfile} onChange={setSelfProfile} showSummary />
          </div>
          <div className={`card compatibility-profile-panel ${activeProfile === 'partner' ? 'compatibility-profile-panel--active' : ''}`}>
            <BirthProfileForm title="对方的生辰" value={partnerProfile} onChange={setPartnerProfile} showSummary />
          </div>
        </div>

        {error && <p className="compatibility-error">{error}</p>}

        <div className="compatibility-actions">
          <button className="btn btn-primary" onClick={handleSubmit} disabled={submitting}>
            {submitting ? '正在起盘合盘...' : '开始合盘'}
          </button>
          <button className="btn btn-ghost" onClick={() => navigate('/compatibility/history')}>
            查看合盘历史
          </button>
        </div>
      </div>
    </div>
  )
}
