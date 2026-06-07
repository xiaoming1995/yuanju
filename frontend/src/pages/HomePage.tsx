import { useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { Compass, HeartHandshake, History } from 'lucide-react'
import { useAuth } from '../contexts/AuthContext'
import { baziAPI } from '../lib/api'
import type { CalculateInput } from '../lib/api'
import { buildAuthPath } from '../lib/authRedirect'
import BirthProfileForm from '../components/BirthProfileForm'
import { initialBirthProfile, type BirthProfileFormValue, type ZiHourMode } from '../components/birthProfile'
import PillarsInputForm, { initialPillarsValue, type PillarsFormValue } from '../components/PillarsInputForm'
import type { PillarCandidate } from '../lib/api'
import './HomePage.css'

// 省份 → 中心经度映射（东8区标准为120°E）
const PROVINCE_LONGITUDE: Record<string, number> = {
  '': 0,
  '北京市': 116.4, '天津市': 117.2, '上海市': 121.5, '重庆市': 106.5,
  '河北省': 114.5, '山西省': 112.6, '内蒙古': 111.7, '辽宁省': 123.4,
  '吉林省': 125.3, '黑龙江省': 126.6, '江苏省': 119.5, '浙江省': 120.2,
  '安徽省': 117.3, '福建省': 119.3, '江西省': 116.0, '山东省': 117.0,
  '河南省': 113.7, '湖北省': 114.3, '湖南省': 112.0, '广东省': 113.3,
  '广西': 108.4, '海南省': 110.3, '四川省': 104.1, '贵州省': 106.7,
  '云南省': 102.7, '西藏': 91.1, '陕西省': 108.9, '甘肃省': 103.8,
  '青海省': 101.8, '宁夏': 106.3, '新疆': 87.6,
  '香港': 114.2, '澳门': 113.5, '台湾': 121.5,
}

export default function HomePage() {
  const navigate = useNavigate()
  const { user } = useAuth()
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  const [birthProfile, setBirthProfile] = useState<BirthProfileFormValue>(initialBirthProfile('male'))
  const [ziHourMode, setZiHourMode] = useState<ZiHourMode>('late')
  const [province, setProvince] = useState('')
  const [showAdvancedCalibration, setShowAdvancedCalibration] = useState(false)
  const [chartDisplayName, setChartDisplayName] = useState('')
  const [displayNameError, setDisplayNameError] = useState('')
  const [inputMode, setInputMode] = useState<'birth' | 'pillars'>('birth')
  const [pillars, setPillars] = useState<PillarsFormValue>(initialPillarsValue('male'))
  const [candidates, setCandidates] = useState<PillarCandidate[]>([])

  const calibrationSummary = province ? `${province}省级经度近似校准` : '按北京时间排盘'

  const trimmedDisplayName = (): string | undefined => {
    const name = chartDisplayName.trim()
    return name || undefined
  }

  // 用「公历年月日 + 小时 + 性别」直接走现有 calculate
  const castBySolar = async (
    year: number, month: number, day: number, hour: number,
    gender: 'male' | 'female', isEarlyZishi: boolean,
  ) => {
    const input: CalculateInput = {
      year, month, day, hour, gender,
      is_early_zishi: isEarlyZishi,
      longitude: PROVINCE_LONGITUDE[province] || 0,
      calendar_type: 'solar',
      is_leap_month: false,
      display_name: trimmedDisplayName(),
    }
    const res = await baziAPI.calculate(input)
    navigate('/result', { state: { result: res.data.result, chartId: res.data.chart_id, input, isGuest: !user } })
  }

  // 八字模式提交：先反查候选
  const handlePillarsSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')
    setCandidates([])
    setLoading(true)
    try {
      const res = await baziAPI.resolvePillars({
        year_pillar: pillars.yearPillar,
        month_pillar: pillars.monthPillar,
        day_pillar: pillars.dayPillar,
        hour_pillar: pillars.hourPillar,
        min_year: pillars.minYear,
        max_year: pillars.maxYear,
      })
      const list = res.data.candidates
      if (list.length === 0) {
        const isZiHour = pillars.hourPillar.endsWith('子')
        setError(isZiHour
          ? '这组八字找不到对应的真实日期。若出生于 23 点后的子时，日柱排法可能不同（早/晚子时），请核对后再试。'
          : '这组八字找不到对应的真实日期，请核对四柱')
      } else if (list.length === 1) {
        await castBySolar(list[0].year, list[0].month, list[0].day, list[0].hour, pillars.gender, false)
      } else {
        setCandidates(list)
      }
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : '反查失败，请重试')
    } finally {
      setLoading(false)
    }
  }

  // 用户从多候选里点选一个
  const handlePickCandidate = async (c: PillarCandidate) => {
    setError('')
    setLoading(true)
    try {
      await castBySolar(c.year, c.month, c.day, c.hour, pillars.gender, false)
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : '计算失败，请重试')
    } finally {
      setLoading(false)
    }
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')
    setDisplayNameError('')
    const trimmedName = chartDisplayName.trim()
    if (Array.from(trimmedName).length > 20) {
      setDisplayNameError('称呼不能超过20个字符')
      return
    }

    setLoading(true)

    try {
      // 子时现在的值是 0。
      // 若是早子时（即当天夜里 23点多），对后端而言需要传 hour=23 和 is_early_zishi=true 才能回退日柱。
      // 如果是晚子时，只需发 0。
      const rawHour = Number(birthProfile.hour)
      const isZishi = rawHour === 0
      const isEarlyZishi = ziHourMode === 'early'
      const finalHour = isZishi && isEarlyZishi ? 23 : rawHour
      const input: CalculateInput = {
        year: Number(birthProfile.year),
        month: Number(birthProfile.month),
        day: Number(birthProfile.day),
        hour: finalHour,
        gender: birthProfile.gender,
        is_early_zishi: isZishi ? isEarlyZishi : false,
        longitude: PROVINCE_LONGITUDE[province] || 0,
        calendar_type: birthProfile.calendarType,
        is_leap_month: birthProfile.isLeapMonth,
        display_name: trimmedName || undefined,
      }

      // 统一调用算法排盘（快速），AI 解读由用户在结果页按钮触发
      const res = await baziAPI.calculate(input)
      navigate('/result', { state: { result: res.data.result, chartId: res.data.chart_id, input, isGuest: !user } })
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : '计算失败，请重试')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="home-page page">
      {/* Hero 区域 */}
      <section className="hero">
        <div className="container">
          <div className="hero-content animate-fade-up">
            <div className="hero-badge serif">八字命理 · 命理解读</div>
            <h1 className="hero-title serif">
              知命理，<span className="text-gold">悟人生</span>
            </h1>
            <p className="hero-desc">
              融合传统八字命理与现代算法，为你解读命盘中的天赋与机遇
            </p>
            <div className="home-intent-grid" aria-label="常用分析入口">
              <a className="home-intent-card" href="#bazi-form">
                <Compass size={18} />
                <span>立即起盘</span>
              </a>
              <Link className="home-intent-card" to={user ? '/history' : buildAuthPath('/login', '/history')}>
                <History size={18} />
                <span>继续记录</span>
              </Link>
              <Link className="home-intent-card" to="/compatibility">
                <HeartHandshake size={18} />
                <span>合盘分析</span>
              </Link>
            </div>
          </div>
        </div>
      </section>

      {/* 表单区域 */}
      <section className="form-section">
        <div className="container">
          <div className="form-card card animate-fade-up" style={{ animationDelay: '0.1s' }}>
            <div className="form-card-header">
              <h2 className="form-card-title serif">输入生辰信息</h2>
              <p className="form-card-desc">请填写真实的出生时间，以获得准确的命理分析</p>
            </div>

            <div className="input-mode-toggle birth-profile-segmented">
              {(['birth', 'pillars'] as const).map(m => (
                <button
                  key={m}
                  type="button"
                  className={`birth-profile-option ${inputMode === m ? 'active' : ''}`}
                  onClick={() => { setInputMode(m); setError(''); setDisplayNameError(''); setCandidates([]) }}
                >
                  {m === 'birth' ? '按生辰' : '按八字'}
                </button>
              ))}
            </div>

            {inputMode === 'birth' && (
            <form onSubmit={handleSubmit} id="bazi-form">
              <BirthProfileForm
                value={birthProfile}
                onChange={setBirthProfile}
                showSummary
                summaryCalibrationText={calibrationSummary}
                ziHourMode={ziHourMode}
                onZiHourModeChange={setZiHourMode}
              />

              <div className="form-group">
                <label className="form-label" htmlFor="chart-display-name">
                  档案称呼（选填）
                  <span className="field-note">用于在历史与合盘中识别此命盘，最多 20 字</span>
                </label>
                <input
                  id="chart-display-name"
                  className="form-input"
                  type="text"
                  value={chartDisplayName}
                  onChange={(e) => {
                    setChartDisplayName(e.target.value)
                    if (displayNameError) setDisplayNameError('')
                  }}
                  maxLength={20}
                  placeholder="例如：我 / 小王"
                />
                {displayNameError && <p className="form-error">{displayNameError}</p>}
              </div>

              <div className="advanced-calibration">
                <button
                  type="button"
                  className="advanced-calibration-toggle"
                  onClick={() => setShowAdvancedCalibration(open => !open)}
                  aria-expanded={showAdvancedCalibration}
                >
                  <span>{showAdvancedCalibration ? '收起高级校准' : '高级校准'}</span>
                  <span aria-hidden="true">{showAdvancedCalibration ? '▴' : '▾'}</span>
                </button>

                {showAdvancedCalibration && (
                  <div className="advanced-calibration-panel">
                    <div className="form-group">
                      <label className="form-label" htmlFor="birth-province">
                        出生地校准
                        <span className="field-note">按省级中心经度近似修正，不选择则按北京时间</span>
                      </label>
                      <select
                        id="birth-province"
                        className="form-select"
                        value={province}
                        onChange={e => setProvince(e.target.value)}
                      >
                        <option value="">不选择，按北京时间</option>
                        {Object.keys(PROVINCE_LONGITUDE).filter(k => k !== '').map(p => (
                          <option key={p} value={p}>{p}</option>
                        ))}
                      </select>
                    </div>
                  </div>
                )}
              </div>

              {error && <p className="form-error">{error}</p>}

              <button
                type="submit"
                id="submit-bazi"
                className="btn btn-primary btn-lg submit-btn"
                disabled={loading}
              >
                {loading ? (
                  <>
                    <span className="loading-spinner" />
                    正在计算...
                  </>
                ) : (
                  <>立即起盘</>
                )}
              </button>

              {!user && (
                <p className="guest-hint">
                  <a href="/login">登录</a>后可保存记录并获得完整解读报告
                </p>
              )}
            </form>
            )}

            {inputMode === 'pillars' && (
              <form onSubmit={handlePillarsSubmit} id="pillars-form">
                <PillarsInputForm value={pillars} onChange={next => { setPillars(next); setCandidates([]) }} />

                {candidates.length > 0 && (
                  <div className="candidate-list" aria-live="polite">
                    <div className="form-label">找到多个可能的出生日期，请按年龄选择</div>
                    {candidates.map(c => (
                      <button
                        type="button"
                        key={`${c.year}-${c.month}-${c.day}-${c.hour}`}
                        className="candidate-item"
                        onClick={() => handlePickCandidate(c)}
                        disabled={loading}
                      >
                        <span>{c.year}-{String(c.month).padStart(2, '0')}-{String(c.day).padStart(2, '0')}</span>
                        <span className="candidate-lunar">{c.lunar_date}</span>
                        <span className="candidate-age">约 {c.ref_age} 岁</span>
                      </button>
                    ))}
                  </div>
                )}

                {error && <p className="form-error">{error}</p>}

                <button type="submit" id="submit-pillars" className="btn btn-primary btn-lg submit-btn" disabled={loading}>
                  {loading ? (<><span className="loading-spinner" />正在反查...</>) : (<>按八字起盘</>)}
                </button>

                {!user && (
                  <p className="guest-hint"><a href="/login">登录</a>后可保存记录并获得完整解读报告</p>
                )}
              </form>
            )}
          </div>
        </div>
      </section>

      {/* 特性展示 */}
      <section className="features-section">
        <div className="container">
          <div className="features-grid">
            {[
              { icon: '盘', title: '精准排盘', desc: '严格遵循节气历法，支持省级经度近似修正' },
              { icon: '解', title: '命理解读', desc: '结合命理知识，生成通俗易懂的个性报告' },
              { icon: '行', title: '五行分析', desc: '可视化五行分布，直观了解命局特点' },
            ].map((f, i) => (
              <div key={i} className="feature-card card">
                <div className="feature-icon">{f.icon}</div>
                <h3 className="feature-title serif">{f.title}</h3>
                <p className="feature-desc">{f.desc}</p>
              </div>
            ))}
          </div>
        </div>
      </section>
    </div>
  )
}
