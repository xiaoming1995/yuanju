import { Link, useNavigate } from 'react-router-dom'
import { Compass, HeartHandshake } from 'lucide-react'
import { useAuth } from '../contexts/AuthContext'
import './Navbar.css'

export default function Navbar() {
  const { user, logout } = useAuth()
  const navigate = useNavigate()

  const handleLogout = () => {
    logout()
    navigate('/')
  }

  return (
    <nav className="navbar">
      <div className="container navbar-inner">
        <Link to="/" className="navbar-logo">
          <Compass size={24} className="navbar-logo-icon" />
          <span className="serif">缘聚命理</span>
        </Link>

        <div className="navbar-links">
          <Link to="/" className="navbar-link">测算</Link>
          <Link to="/compatibility" className="navbar-link">
            <HeartHandshake size={16} style={{ marginRight: 6, verticalAlign: 'text-bottom' }} />
            合盘
          </Link>
          {user && <Link to="/profile" className="navbar-link">我的</Link>}
          {user && <Link to="/history" className="navbar-link">历史</Link>}
        </div>

        <div className="navbar-auth">
          {user ? (
            <div className="navbar-user">
              <Link to="/profile" className="navbar-nickname">{user.nickname}</Link>
              <button className="btn btn-ghost btn-sm" onClick={handleLogout}>退出</button>
            </div>
          ) : (
            <div className="navbar-actions">
              <Link to="/login" className="btn btn-ghost btn-sm">登录</Link>
              <Link to="/register" className="btn btn-primary btn-sm">注册</Link>
            </div>
          )}
        </div>
      </div>
    </nav>
  )
}
