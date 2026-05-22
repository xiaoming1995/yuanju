import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { HeartHandshake } from 'lucide-react'
import { useAuth } from '../contexts/AuthContext'
import {
  compatibilityAPI,
  type CompatibilityPrimaryQuestion,
  type CompatibilityProfileInput,
  type CompatibilityRelationshipStage,
} from '../lib/api'
import BirthProfileForm from '../components/BirthProfileForm'
import { initialBirthProfile, type BirthProfileFormValue } from '../components/birthProfile'
import { buildPersonalityConsultationPreview } from '../lib/compatibilityPersonality'
import './CompatibilityPage.css'

const relationshipStageOptions: Array<{ value: CompatibilityRelationshipStage; label: string }> = [
  { value: 'ambiguous', label: '暧昧中' },
  { value: 'dating', label: '恋爱中' },
  { value: 'long_distance', label: '异地中' },
  { value: 'reconciliation', label: '分手/复合中' },
  { value: 'marriage_or_engagement', label: '谈婚论嫁' },
  { value: 'crush', label: '单恋/暗恋' },
]

const primaryQuestionOptions: Array<{ value: CompatibilityPrimaryQuestion; label: string }> = [
  { value: 'continue_investment', label: '性格合不合 / 值不值得继续' },
  { value: 'marriage_suitability', label: '适不适合结婚' },
  { value: 'recurring_conflict', label: '为什么反复拉扯' },
  { value: 'reconciliation_potential', label: '复合有没有意义' },
  { value: 'long_term_stability', label: '长期能不能稳定' },
  { value: 'relationship_strategy', label: '怎么相处更顺' },
]

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
  const [relationshipStage, setRelationshipStage] = useState<CompatibilityRelationshipStage>('ambiguous')
  const [primaryQuestion, setPrimaryQuestion] = useState<CompatibilityPrimaryQuestion>('continue_investment')
  const [activeProfile, setActiveProfile] = useState<'self' | 'partner'>('self')
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState('')
  const consultationPreview = buildPersonalityConsultationPreview(relationshipStage, primaryQuestion)
  const questionProgressLabel = primaryQuestionOptions.find(option => option.value === primaryQuestion)?.label || '性格合不合'
  const stageProgressLabel = relationshipStageOptions.find(option => option.value === relationshipStage)?.label || '综合关系判断'
  const birthProfileProgressLabel = selfProfile && partnerProfile ? '双方生辰已可起盘' : '待补双方生辰'

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
        relationship_stage: relationshipStage,
        primary_question: primaryQuestion,
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
            先批两个人性格合不合，再看吸引力、稳定度、沟通协同和现实磨合，给出可验证的相处建议。
          </p>
        </div>

        <div className="card compatibility-context-card compatibility-personality-consultation">
          <div className="compatibility-context-heading">性格合盘咨询 · 关系背景</div>
          <div className="compatibility-step-progress" aria-label="合盘流程">
            <div className="compatibility-step-item">
              <span>第 1 步</span>
              <strong>{questionProgressLabel}</strong>
            </div>
            <div className="compatibility-step-item">
              <span>第 2 步</span>
              <strong>{stageProgressLabel}</strong>
            </div>
            <div className="compatibility-step-item">
              <span>第 3 步</span>
              <strong>{birthProfileProgressLabel}</strong>
            </div>
          </div>
          <div className="compatibility-context-group">
            <div className="compatibility-context-title">你最想知道什么？</div>
            <div className="compatibility-context-options">
              {primaryQuestionOptions.map(option => (
                <button
                  key={option.value}
                  type="button"
                  className={`compatibility-context-option ${primaryQuestion === option.value ? 'active' : ''}`}
                  onClick={() => setPrimaryQuestion(option.value)}
                >
                  {option.label}
                </button>
              ))}
            </div>
          </div>
          <div className="compatibility-context-group">
            <div className="compatibility-context-title">你们目前是什么关系？</div>
            <div className="compatibility-context-options">
              {relationshipStageOptions.map(option => (
                <button
                  key={option.value}
                  type="button"
                  className={`compatibility-context-option ${relationshipStage === option.value ? 'active' : ''}`}
                  onClick={() => setRelationshipStage(option.value)}
                >
                  {option.label}
                </button>
              ))}
            </div>
          </div>
          <div className="compatibility-personality-preview">
            <div className="compatibility-personality-preview-title">{consultationPreview.title}</div>
            <p>{consultationPreview.description}</p>
            <div className="compatibility-personality-preview-list">
              {consultationPreview.bullets.map(item => <span key={item}>{item}</span>)}
            </div>
          </div>
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
