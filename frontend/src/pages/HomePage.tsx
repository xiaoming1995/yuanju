import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAuth } from '../contexts/AuthContext'
import { baziAPI } from '../lib/api'
import type { CalculateInput } from '../lib/api'
import BirthProfileForm from '../components/BirthProfileForm'
import { initialBirthProfile, type BirthProfileFormValue, type ZiHourMode } from '../components/birthProfile'
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

  const calibrationSummary = province ? `${province}真太阳时校准` : '按北京时间排盘'

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')
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

            <form onSubmit={handleSubmit} id="bazi-form">
              <BirthProfileForm
                value={birthProfile}
                onChange={setBirthProfile}
                showSummary
                summaryCalibrationText={calibrationSummary}
                ziHourMode={ziHourMode}
                onZiHourModeChange={setZiHourMode}
              />

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
                        <span className="field-note">用于真太阳时修正，不选择则按北京时间</span>
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
          </div>
        </div>
      </section>

      {/* 特性展示 */}
      <section className="features-section">
        <div className="container">
          <div className="features-grid">
            {[
              { icon: '', title: '精准排盘', desc: '严格遵循古法节气历法，支持真太阳时地区修正' },
              { icon: '', title: '命理解读', desc: '结合命理知识，生成通俗易懂的个性报告' },
              { icon: '', title: '五行分析', desc: '可视化五行分布，直观了解命局特点' },
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
