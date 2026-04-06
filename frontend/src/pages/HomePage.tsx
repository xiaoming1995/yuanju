import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAuth } from '../contexts/AuthContext'
import { baziAPI } from '../lib/api'
import type { CalculateInput } from '../lib/api'
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

// 十二时辰选择器（value = 该时辰中间小时，防真太阳时偏移导致落入上一时辰）
const SHICHEN = [
  { value: 0,  label: '子时（23:00 - 00:59）' },
  { value: 2,  label: '丑时（01:00 - 02:59）' },
  { value: 4,  label: '寅时（03:00 - 04:59）' },
  { value: 6,  label: '卯时（05:00 - 06:59）' },
  { value: 8,  label: '辰时（07:00 - 08:59）' },
  { value: 10, label: '巳时（09:00 - 10:59）' },
  { value: 12, label: '午时（11:00 - 12:59）' },
  { value: 14, label: '未时（13:00 - 14:59）' },
  { value: 16, label: '申时（15:00 - 16:59）' },
  { value: 18, label: '酉时（17:00 - 18:59）' },
  { value: 20, label: '戌时（19:00 - 20:59）' },
  { value: 22, label: '亥时（21:00 - 22:59）' },
]

export default function HomePage() {
  const navigate = useNavigate()
  const { user } = useAuth()
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  const [form, setForm] = useState({
    year: new Date().getFullYear() - 25,
    month: 1,
    day: 1,
    hour: 12, // 默认午时中点
    gender: 'male' as 'male' | 'female',
    is_early_zishi: false,
    province: '',  // 出生省份（可选，用于真太阳时修正）
  })

  const handleChange = (field: string, value: string | number | boolean) => {
    setForm(prev => ({ ...prev, [field]: value }))
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')
    setLoading(true)

    try {
      // 子时现在的值是 0。
      // 若是早子时（即当天夜里 23点多），对后端而言需要传 hour=23 和 is_early_zishi=true 才能回退日柱。
      // 如果是晚子时，只需发 0。
      const rawHour = Number(form.hour)
      const isZishi = rawHour === 0
      const finalHour = isZishi && form.is_early_zishi ? 23 : rawHour
      const input: CalculateInput = {
        year: Number(form.year),
        month: Number(form.month),
        day: Number(form.day),
        hour: finalHour,
        gender: form.gender,
        is_early_zishi: isZishi ? form.is_early_zishi : false,
        longitude: PROVINCE_LONGITUDE[form.province] || 0,
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

  // 计算指定年月的实际天数（自动处理闰年）
  // new Date(year, month, 0) = month 月第 0 天 = 上月最后一天
  const getDaysInMonth = (year: number, month: number) => new Date(year, month, 0).getDate()

  const currentYear = new Date().getFullYear()
  const years = Array.from({ length: 120 }, (_, i) => currentYear - i)
  const months = Array.from({ length: 12 }, (_, i) => i + 1)
  const days = Array.from({ length: getDaysInMonth(form.year, form.month) }, (_, i) => i + 1)

  return (
    <div className="home-page page">
      {/* Hero 区域 */}
      <section className="hero">
        <div className="container">
          <div className="hero-content animate-fade-up">
            <div className="hero-badge serif">✦ 八字命理 · AI 解读 ✦</div>
            <h1 className="hero-title serif">
              知命理，<span className="text-gold">悟人生</span>
            </h1>
            <p className="hero-desc">
              融合传统八字命理与人工智能，为你解读命盘中的天赋与机遇
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
              {/* 性别 */}
              <div className="gender-selector">
                {(['male', 'female'] as const).map(g => (
                  <button
                    key={g}
                    type="button"
                    className={`gender-btn ${form.gender === g ? 'active' : ''}`}
                    onClick={() => handleChange('gender', g)}
                  >
                    {g === 'male' ? '♂ 男命' : '♀ 女命'}
                  </button>
                ))}
              </div>

              {/* 日期行 */}
              <div className="date-row">
                <div className="form-group">
                  <label className="form-label">出生年份</label>
                  <select
                    id="birth-year"
                    className="form-select"
                    value={form.year}
                  onChange={e => {
                    const newYear = Number(e.target.value)
                    const maxDay = getDaysInMonth(newYear, form.month)
                    setForm(prev => ({ ...prev, year: newYear, day: prev.day > maxDay ? 1 : prev.day }))
                  }}
                  >
                    {years.map(y => <option key={y} value={y}>{y} 年</option>)}
                  </select>
                </div>

                <div className="form-group">
                  <label className="form-label">月</label>
                  <select
                    id="birth-month"
                    className="form-select"
                    value={form.month}
                  onChange={e => {
                    const newMonth = Number(e.target.value)
                    const maxDay = getDaysInMonth(form.year, newMonth)
                    setForm(prev => ({ ...prev, month: newMonth, day: prev.day > maxDay ? 1 : prev.day }))
                  }}
                  >
                    {months.map(m => <option key={m} value={m}>{m} 月</option>)}
                  </select>
                </div>

                <div className="form-group">
                  <label className="form-label">日</label>
                  <select
                    id="birth-day"
                    className="form-select"
                    value={form.day}
                    onChange={e => handleChange('day', Number(e.target.value))}
                  >
                    {days.map(d => <option key={d} value={d}>{d} 日</option>)}
                  </select>
                </div>
              </div>

              {/* 时辰（十二时辰）*/}
              <div className="form-group">
                <label className="form-label">出生时辰</label>
                <select
                  id="birth-hour"
                  className="form-select"
                  value={form.hour}
                  onChange={e => handleChange('hour', Number(e.target.value))}
                >
                  {SHICHEN.map(h => (
                    <option key={h.value} value={h.value}>{h.label}</option>
                  ))}
                </select>
              </div>

              {/* 早子时选项（仅子时显示）*/}
              {form.hour === 0 && (
                <div className="early-zishi-hint">
                  <label className="checkbox-label">
                    <input
                      type="checkbox"
                      checked={form.is_early_zishi}
                      onChange={e => handleChange('is_early_zishi', e.target.checked)}
                    />
                    <span>早子时（23:00 前，日柱按前一天算）</span>
                  </label>
                </div>
              )}

              {/* 出生省份（可选，用于真太阳时修正）*/}
              <div className="form-group">
                <label className="form-label">
                  出生省份
                  <span style={{ fontSize: '12px', color: 'var(--text-muted)', marginLeft: '6px' }}>（可选，用于真太阳时修正）</span>
                </label>
                <select
                  id="birth-province"
                  className="form-select"
                  value={form.province}
                  onChange={e => handleChange('province', e.target.value)}
                >
                  <option value="">不选择（按北京时间）</option>
                  {Object.keys(PROVINCE_LONGITUDE).filter(k => k !== '').map(p => (
                    <option key={p} value={p}>{p}</option>
                  ))}
                </select>
              </div>

              {error && <p className="form-error">⚠ {error}</p>}

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
                  <>✦ 立即起盘</>
                )}
              </button>

              {!user && (
                <p className="guest-hint">
                  <a href="/login">登录</a>后可保存记录并获得 AI 智能解读报告
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
              { icon: '☯', title: '传统算法', desc: '基于 lunar-go 天文历法库，精确到秒级节气与真太阳时' },
              { icon: '✦', title: 'AI 智能解读', desc: '大模型结合命理知识，生成通俗易懂的个性报告' },
              { icon: '◈', title: '五行分析', desc: '可视化五行分布，直观了解命局特点' },
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
