import { useState } from 'react'
import { Link, useLocation, useNavigate } from 'react-router-dom'
import { useAuth } from '../contexts/AuthContext'
import { authAPI, baziAPI } from '../lib/api'
import { buildAuthPath, getNextTargetFromSearch, resolvePostAuthTarget } from '../lib/authRedirect'
import { clearPendingJourney, readPendingJourney } from '../lib/pendingJourney'
import './AuthPage.css'

export default function LoginPage() {
  const navigate = useNavigate()
  const location = useLocation()
  const { login } = useAuth()
  const [form, setForm] = useState({ email: '', password: '' })
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  const safeNext = getNextTargetFromSearch(location.search, '')
  const switchToRegisterPath = buildAuthPath('/register', safeNext)

  const completePostAuthNavigation = async () => {
    const pendingJourney = readPendingJourney()
    if (pendingJourney?.type === 'bazi') {
      try {
        const res = await baziAPI.calculate(pendingJourney.input)
        clearPendingJourney()
        navigate(pendingJourney.returnPath || '/result', {
          replace: true,
          state: {
            result: res.data.result,
            chartId: res.data.chart_id,
            input: pendingJourney.input,
            isGuest: false,
            pendingIntent: pendingJourney.intent,
          },
        })
        return true
      } catch (err: unknown) {
        setError(err instanceof Error ? `登录成功，但恢复刚才的命盘失败：${err.message}` : '登录成功，但恢复刚才的命盘失败')
        return false
      }
    }

    navigate(resolvePostAuthTarget({
      search: location.search,
      stateNext: (location.state as { next?: unknown } | null)?.next,
    }), { replace: true })
    return true
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')
    setLoading(true)
    try {
      const res = await authAPI.login(form)
      login(res.data.token, res.data.user)
      await completePostAuthNavigation()
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : '登录失败')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="auth-page page">
      <div className="container">
        <div className="auth-card card animate-fade-up">
          <div className="auth-header">
            <div className="auth-logo serif">☯</div>
            <h1 className="auth-title serif">欢迎回来</h1>
            <p className="auth-desc">登录缘聚命理，查看您的命盘记录</p>
          </div>

          <form onSubmit={handleSubmit} id="login-form">
            <div className="form-group">
              <label className="form-label" htmlFor="email">邮箱地址</label>
              <input
                id="email"
                type="email"
                className="form-input"
                placeholder="your@email.com"
                value={form.email}
                onChange={e => setForm(p => ({ ...p, email: e.target.value }))}
                required
              />
            </div>

            <div className="form-group">
              <label className="form-label" htmlFor="password">密码</label>
              <input
                id="password"
                type="password"
                className="form-input"
                placeholder="请输入密码"
                value={form.password}
                onChange={e => setForm(p => ({ ...p, password: e.target.value }))}
                required
              />
            </div>

            {error && <p className="form-error">⚠ {error}</p>}

            <button
              id="login-submit"
              type="submit"
              className="btn btn-primary btn-lg auth-submit"
              disabled={loading}
            >
              {loading ? <><span className="loading-spinner-dark" /> 登录中...</> : '登录'}
            </button>
          </form>

          <p className="auth-switch">
            还没有账号？<Link to={switchToRegisterPath}>立即注册</Link>
          </p>
        </div>
      </div>
    </div>
  )
}
