import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAdminAuth } from '../../contexts/AdminAuthContext'
import { adminAuthAPI } from '../../lib/adminApi'
import '../../components/AdminLayout.css'

export default function AdminLoginPage() {
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)
  const { login } = useAdminAuth()
  const navigate = useNavigate()

  const handleLogin = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')
    setLoading(true)
    try {
      const res = await adminAuthAPI.login({ email, password })
      login(res.data.token, res.data.admin)
      navigate('/admin/dashboard')
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : '登录失败')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="admin-login-page">
      <div className="admin-login-box">
        <div className="admin-login-title">⚙ 缘聚后台</div>
        <div className="admin-login-sub">管理员登录</div>
        {error && <div className="admin-error">{error}</div>}
        <form onSubmit={handleLogin} style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
          <div className="admin-form-group">
            <label className="admin-form-label">邮箱</label>
            <input
              id="admin-email"
              className="admin-form-input"
              type="email"
              value={email}
              onChange={e => setEmail(e.target.value)}
              placeholder="admin@example.com"
              required
            />
          </div>
          <div className="admin-form-group">
            <label className="admin-form-label">密码</label>
            <input
              id="admin-password"
              className="admin-form-input"
              type="password"
              value={password}
              onChange={e => setPassword(e.target.value)}
              placeholder="••••••••"
              required
            />
          </div>
          <button type="submit" className="admin-btn admin-btn-primary" disabled={loading}
            style={{ marginTop: 8, padding: '12px', fontSize: 15 }}>
            {loading ? '登录中...' : '登录'}
          </button>
        </form>
      </div>
    </div>
  )
}
