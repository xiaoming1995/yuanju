import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { Link, useNavigate, useSearchParams } from 'react-router-dom'
import { HeartHandshake } from 'lucide-react'
import { useAuth } from '../contexts/AuthContext'
import {
  baziAPI,
  compatibilityAPI,
  type BaziHistoryChart,
  type CompatibilityPrimaryQuestion,
  type CompatibilityProfileInput,
  type CompatibilityRelationshipStage,
} from '../lib/api'
import BirthProfileForm from '../components/BirthProfileForm'
import {
  chartToBirthProfile,
  initialBirthProfile,
  type BirthProfileFormValue,
  type BirthProfileImportSource,
} from '../components/birthProfile'
import { chartDisplayName, formatBirth, formatPillars, genderText } from '../lib/chartLabel'
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

function buildImportSource(chart: BaziHistoryChart): BirthProfileImportSource {
  const profile = chartToBirthProfile(chart)
  return {
    chartId: chart.id,
    displayName: chart.display_name?.trim() || '',
    profile,
  }
}

export default function CompatibilityPage() {
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()
  const { user, isLoading } = useAuth()
  const handledImportQueryRef = useRef('')
  const chartPickerRef = useRef<HTMLDivElement | null>(null)
  const pickerSearchRef = useRef<HTMLInputElement | null>(null)
  const [selfProfile, setSelfProfile] = useState<BirthProfileFormValue>(initialBirthProfile('male'))
  const [partnerProfile, setPartnerProfile] = useState<BirthProfileFormValue>(initialBirthProfile('female'))
  const [relationshipStage, setRelationshipStage] = useState<CompatibilityRelationshipStage>('ambiguous')
  const [primaryQuestion, setPrimaryQuestion] = useState<CompatibilityPrimaryQuestion>('continue_investment')
  const [activeProfile, setActiveProfile] = useState<'self' | 'partner'>('self')
  const [historyCharts, setHistoryCharts] = useState<BaziHistoryChart[]>([])
  const [historyLoaded, setHistoryLoaded] = useState(false)
  const [pickerRole, setPickerRole] = useState<'self' | 'partner' | null>(null)
  const [pickerSearch, setPickerSearch] = useState('')
  const [pickerGenderFilter, setPickerGenderFilter] = useState<'all' | 'male' | 'female'>('all')
  const [selfImportSource, setSelfImportSource] = useState<BirthProfileImportSource | null>(null)
  const [partnerImportSource, setPartnerImportSource] = useState<BirthProfileImportSource | null>(null)
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState('')
  const consultationPreview = buildPersonalityConsultationPreview(relationshipStage, primaryQuestion)
  const questionProgressLabel = primaryQuestionOptions.find(option => option.value === primaryQuestion)?.label || '性格合不合'
  const stageProgressLabel = relationshipStageOptions.find(option => option.value === relationshipStage)?.label || '综合关系判断'
  const birthProfileProgressLabel = selfProfile && partnerProfile ? '双方生辰已可起盘' : '待补双方生辰'
  const userKey = user?.id || user?.email || ''

  const loadHistoryCharts = async () => {
    if (isLoading) return []
    if (!user) {
      navigate('/login')
      return []
    }
    if (historyLoaded) return historyCharts
    const res = await baziAPI.getHistory()
    const charts = (res.data.charts || []) as BaziHistoryChart[]
    setHistoryCharts(charts)
    setHistoryLoaded(true)
    return charts
  }

  const applyImportedChart = useCallback((role: 'self' | 'partner', chart: BaziHistoryChart) => {
    const source = buildImportSource(chart)
    if (role === 'self') {
      setSelfProfile(source.profile)
      setSelfImportSource(source)
      setActiveProfile('self')
    } else {
      setPartnerProfile(source.profile)
      setPartnerImportSource(source)
      setActiveProfile('partner')
    }
  }, [])

  const importChartFromHistory = useCallback(async (role: 'self' | 'partner', chartId: string) => {
    try {
      const res = await baziAPI.getHistoryDetail(chartId)
      applyImportedChart(role, res.data.chart as BaziHistoryChart)
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : '命盘不存在或无权访问')
    }
  }, [applyImportedChart])

  const importLatestChart = async (role: 'self' | 'partner') => {
    if (isLoading) return
    const charts = await loadHistoryCharts()
    if (charts.length === 0) {
      setError('还没有命盘档案，请先新建命盘')
      return
    }
    applyImportedChart(role, charts[0])
  }

  useEffect(() => {
    const chartId = searchParams.get('importChart')
    const role = searchParams.get('role')
    if (!chartId || (role !== 'self' && role !== 'partner')) return
    if (isLoading) return
    if (!userKey) {
      navigate('/login')
      return
    }
    const importKey = `${chartId}:${role}`
    if (handledImportQueryRef.current === importKey) return
    handledImportQueryRef.current = importKey
    importChartFromHistory(role, chartId)
  }, [searchParams, isLoading, userKey, navigate, importChartFromHistory])

  useEffect(() => {
    if (!pickerRole) return
    const previous = document.activeElement instanceof HTMLElement ? document.activeElement : null
    pickerSearchRef.current?.focus()
    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key === 'Escape') setPickerRole(null)
    }
    document.addEventListener('keydown', handleKeyDown)
    return () => {
      document.removeEventListener('keydown', handleKeyDown)
      previous?.focus()
    }
  }, [pickerRole])

  const handleSelfProfileChange = (next: BirthProfileFormValue) => {
    setSelfProfile(next)
  }

  const handlePartnerProfileChange = (next: BirthProfileFormValue) => {
    setPartnerProfile(next)
  }

  const isProfileModified = (source: BirthProfileImportSource | null, current: BirthProfileFormValue) => {
    if (!source) return false
    return JSON.stringify(source.profile) !== JSON.stringify(current)
  }

  const importSourceLabel = (source: BirthProfileImportSource | null, current: BirthProfileFormValue, fallback: string) => {
    if (!source) return ''
    const name = source.displayName || fallback
    return isProfileModified(source, current) ? `已基于${name}修改` : `已导入：${name}`
  }

  const displayNameForSubmit = (source: BirthProfileImportSource | null, current: BirthProfileFormValue) => {
    return source && !isProfileModified(source, current) && source.displayName ? source.displayName : undefined
  }

  const openChartPicker = async (role: 'self' | 'partner') => {
    if (isLoading) return
    setPickerSearch('')
    if (role === 'partner') {
      setPickerGenderFilter(selfProfile.gender === 'male' ? 'female' : 'male')
    } else {
      setPickerGenderFilter(selfProfile.gender === 'male' ? 'male' : 'female')
    }
    setPickerRole(role)
    await loadHistoryCharts()
  }

  const pickerCharts = useMemo(() => {
    const sameSideId = pickerRole === 'self' ? selfImportSource?.chartId : partnerImportSource?.chartId
    const otherSideId = pickerRole === 'self' ? partnerImportSource?.chartId : selfImportSource?.chartId
    const q = pickerSearch.trim().toLowerCase()

    return historyCharts
      .filter(c => pickerGenderFilter === 'all' || c.gender === pickerGenderFilter)
      .filter(c => !q || (c.display_name || '').toLowerCase().includes(q))
      .map(c => ({
        chart: c,
        isSelectedHere: c.id === sameSideId,
        isUsedByOtherSide: c.id === otherSideId,
      }))
  }, [historyCharts, pickerGenderFilter, pickerSearch, pickerRole, selfImportSource, partnerImportSource])

  const handleSubmit = async () => {
    if (isLoading) return
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
        self_display_name: displayNameForSubmit(selfImportSource, selfProfile),
        partner_display_name: displayNameForSubmit(partnerImportSource, partnerProfile),
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
            <div className="compatibility-import-toolbar">
              <button type="button" className="btn btn-ghost" onClick={() => importLatestChart('self')}>
                导入最近命盘
              </button>
              <button type="button" className="btn btn-ghost" onClick={() => openChartPicker('self')}>
                从命盘档案选择
              </button>
            </div>
            {selfImportSource && (
              <p className="compatibility-import-source">
                {importSourceLabel(selfImportSource, selfProfile, '我的命盘')}
              </p>
            )}
            <BirthProfileForm title="我的生辰" value={selfProfile} onChange={handleSelfProfileChange} showSummary />
          </div>
          <div className={`card compatibility-profile-panel ${activeProfile === 'partner' ? 'compatibility-profile-panel--active' : ''}`}>
            <div className="compatibility-import-toolbar">
              <button type="button" className="btn btn-ghost" onClick={() => importLatestChart('partner')}>
                导入最近命盘
              </button>
              <button type="button" className="btn btn-ghost" onClick={() => openChartPicker('partner')}>
                从命盘档案选择
              </button>
            </div>
            {partnerImportSource && (
              <p className="compatibility-import-source">
                {importSourceLabel(partnerImportSource, partnerProfile, '对方命盘')}
              </p>
            )}
            <BirthProfileForm title="对方的生辰" value={partnerProfile} onChange={handlePartnerProfileChange} showSummary />
          </div>
        </div>

        {error && <p className="compatibility-error">{error}</p>}

        <div className="compatibility-actions">
          <button className="btn btn-primary" onClick={handleSubmit} disabled={submitting || isLoading}>
            {submitting ? '正在起盘合盘...' : '开始合盘'}
          </button>
          <button className="btn btn-ghost" onClick={() => navigate('/compatibility/history')}>
            查看合盘历史
          </button>
        </div>
      </div>

      {pickerRole && (
        <div className="compatibility-chart-picker" role="dialog" aria-modal="true" aria-label="选择命盘档案">
          <div className="compatibility-chart-picker-panel" ref={chartPickerRef} tabIndex={-1}>
            <div className="compatibility-chart-picker-head">
              <strong>选择命盘档案</strong>
              <button type="button" className="btn btn-ghost" onClick={() => setPickerRole(null)}>
                关闭
              </button>
            </div>
            <div className="compatibility-chart-picker-toolbar">
              <input
                ref={pickerSearchRef}
                type="search"
                className="compatibility-chart-picker-search"
                placeholder="搜索命盘称呼"
                value={pickerSearch}
                onChange={event => setPickerSearch(event.target.value)}
                aria-label="搜索命盘称呼"
              />
              <div className="compatibility-chart-picker-filter" role="tablist" aria-label="性别筛选">
                {(['all', 'male', 'female'] as const).map(g => (
                  <button
                    key={g}
                    type="button"
                    role="tab"
                    aria-selected={pickerGenderFilter === g}
                    onClick={() => setPickerGenderFilter(g)}
                  >
                    {g === 'all' ? '全部' : g === 'male' ? '男命' : '女命'}
                  </button>
                ))}
              </div>
            </div>
            {!historyLoaded ? (
              <div className="compatibility-chart-picker-empty">
                <p>加载中…</p>
              </div>
            ) : pickerCharts.length === 0 ? (
              historyCharts.length === 0 ? (
                <div className="compatibility-chart-picker-empty">
                  <p>还没有命盘档案</p>
                  <Link to="/" className="btn btn-primary compatibility-chart-picker-empty-cta">
                    立即新建命盘
                  </Link>
                </div>
              ) : (
                <div className="compatibility-chart-picker-empty">
                  <p>没有匹配的命盘</p>
                  <button
                    type="button"
                    className="btn btn-ghost compatibility-chart-picker-empty-cta"
                    onClick={() => { setPickerSearch(''); setPickerGenderFilter('all') }}
                  >
                    清除筛选
                  </button>
                </div>
              )
            ) : (
              <div className="compatibility-chart-picker-list" role="listbox">
                {pickerCharts.map(({ chart, isSelectedHere, isUsedByOtherSide }) => (
                  <button
                    key={chart.id}
                    type="button"
                    role="option"
                    aria-selected={isSelectedHere}
                    className="compatibility-chart-picker-item"
                    onClick={() => {
                      applyImportedChart(pickerRole, chart)
                      setPickerRole(null)
                    }}
                  >
                    <div className="compatibility-chart-picker-item-head">
                      <span className="compatibility-chart-picker-item-name">
                        {chartDisplayName(chart)}
                      </span>
                      <span className="compatibility-chart-picker-item-badges">
                        <span className="compatibility-chart-picker-badge compatibility-chart-picker-badge-gender">
                          {genderText(chart.gender)}
                        </span>
                        {isSelectedHere && (
                          <span className="compatibility-chart-picker-badge compatibility-chart-picker-badge-selected">
                            已选中
                          </span>
                        )}
                        {!isSelectedHere && isUsedByOtherSide && (
                          <span className="compatibility-chart-picker-badge compatibility-chart-picker-badge-conflict">
                            已作为{pickerRole === 'self' ? '对方' : '我'}
                          </span>
                        )}
                      </span>
                    </div>
                    <div className="compatibility-chart-picker-item-birth">{formatBirth(chart)}</div>
                    <div className="compatibility-chart-picker-item-pillars serif">{formatPillars(chart)}</div>
                  </button>
                ))}
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  )
}
