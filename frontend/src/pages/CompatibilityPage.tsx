import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { HeartHandshake } from 'lucide-react'
import { useAuth } from '../contexts/AuthContext'
import { compatibilityAPI, type CompatibilityProfileInput } from '../lib/api'
import BirthProfileForm, { initialBirthProfile, type BirthProfileFormValue } from '../components/BirthProfileForm'

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
    } catch (err: any) {
      setError(err?.message || '创建合盘失败')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <div className="page">
      <div className="container" style={{ paddingBottom: 96 }}>
        <div className="card" style={{ padding: 24, marginTop: 28, marginBottom: 20 }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: 10, marginBottom: 8 }}>
            <HeartHandshake size={24} />
            <h1 className="serif" style={{ fontSize: 30, margin: 0 }}>姻缘合盘</h1>
          </div>
          <p style={{ margin: 0, color: 'var(--text-muted)', lineHeight: 1.7 }}>
            第一版采用双盘结构化分析：吸引力、稳定度、沟通协同、现实磨合四维结果，生成完整解读。
          </p>
        </div>

        <div style={{ display: 'grid', gap: 16 }}>
          <div className="card" style={{ padding: 20 }}>
            <BirthProfileForm title="我的生辰" value={selfProfile} onChange={setSelfProfile} />
          </div>
          <div className="card" style={{ padding: 20 }}>
            <BirthProfileForm title="对方的生辰" value={partnerProfile} onChange={setPartnerProfile} />
          </div>
        </div>

        {error && <p style={{ color: '#e77', marginTop: 16 }}>{error}</p>}

        <div style={{ display: 'flex', gap: 12, marginTop: 20 }}>
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
