import { useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { useAuth } from '../contexts/AuthContext'
import { authAPI } from '../lib/api'
import './AuthPage.css'

export default function RegisterPage() {
  const navigate = useNavigate()
  const { login } = useAuth()
  const [form, setForm] = useState({ email: '', password: '', nickname: '' })
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')
    if (form.password.length < 8) {
      setError('密码至少需要8位')
      return
    }
    setLoading(true)
    try {
      const res = await authAPI.register(form)
      login(res.data.token, res.data.user)
      navigate('/')
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : '注册失败')
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
            <h1 className="auth-title serif">加入缘聚</h1>
            <p className="auth-desc">注册账号，开启你的命理之旅</p>
          </div>

          <form onSubmit={handleSubmit} id="register-form">
            <div className="form-group">
              <label className="form-label" htmlFor="reg-email">邮箱地址</label>
              <input
                id="reg-email"
                type="email"
                className="form-input"
                placeholder="your@email.com"
                value={form.email}
                onChange={e => setForm(p => ({ ...p, email: e.target.value }))}
                required
              />
            </div>

            <div className="form-group">
              <label className="form-label" htmlFor="reg-nickname">昵称（选填）</label>
              <input
                id="reg-nickname"
                type="text"
                className="form-input"
                placeholder="你的名字或昵称"
                value={form.nickname}
                onChange={e => setForm(p => ({ ...p, nickname: e.target.value }))}
              />
            </div>

            <div className="form-group">
              <label className="form-label" htmlFor="reg-password">密码</label>
              <input
                id="reg-password"
                type="password"
                className="form-input"
                placeholder="至少8位密码"
                value={form.password}
                onChange={e => setForm(p => ({ ...p, password: e.target.value }))}
                required
                minLength={8}
              />
            </div>

            {error && <p className="form-error">⚠ {error}</p>}

            <button
              id="register-submit"
              type="submit"
              className="btn btn-primary btn-lg auth-submit"
              disabled={loading}
            >
              {loading ? <><span className="loading-spinner-dark" /> 注册中...</> : '注册'}
            </button>
          </form>

          <p className="auth-switch">
            已有账号？<Link to="/login">立即登录</Link>
          </p>
        </div>
      </div>
    </div>
  )
}
