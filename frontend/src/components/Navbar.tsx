import { useEffect, useState } from 'react'
import { Link, NavLink, useNavigate } from 'react-router-dom'
import { Clock3, Compass, HeartHandshake, Newspaper, UserRound } from 'lucide-react'
import { useAuth } from '../contexts/AuthContext'
import { articleAPI, authAPI } from '../lib/api'
import './Navbar.css'

export default function Navbar() {
  const { user, logout } = useAuth()
  const navigate = useNavigate()
  const [registration_enabled, setRegistrationEnabled] = useState(false)
  const [articleModuleEnabled, setArticleModuleEnabled] = useState(false)

  useEffect(() => {
    if (user) return
    authAPI.registrationSettings()
      .then(r => setRegistrationEnabled(r.data.registration_enabled))
      .catch(() => setRegistrationEnabled(true))
  }, [user])

  useEffect(() => {
    let alive = true
    articleAPI.settings()
      .then(r => {
        if (alive) setArticleModuleEnabled(Boolean(r.data.module_enabled))
      })
      .catch(() => {
        if (alive) setArticleModuleEnabled(false)
      })
    return () => { alive = false }
  }, [])

  const handleLogout = () => {
    logout()
    navigate('/')
  }

  const navLinkClass = ({ isActive }: { isActive: boolean }) =>
    `navbar-link ${isActive ? 'active' : ''}`

  return (
    <nav className="navbar">
      <div className="container navbar-inner">
        <Link to="/" className="navbar-logo">
          <Compass size={24} className="navbar-logo-icon" />
          <span className="serif">缘聚命理</span>
        </Link>

        <div className="navbar-links">
          <NavLink to="/" end className={navLinkClass}>
            <Compass size={16} className="navbar-link-icon" />
            <span>测算</span>
          </NavLink>
          <NavLink to="/compatibility" className={navLinkClass}>
            <HeartHandshake size={16} className="navbar-link-icon" />
            <span>合盘</span>
          </NavLink>
          {user && articleModuleEnabled && (
            <NavLink to="/articles" className={navLinkClass}>
              <Newspaper size={16} className="navbar-link-icon" />
              <span>资讯</span>
            </NavLink>
          )}
          {user && (
            <NavLink to="/profile" className={navLinkClass}>
              <UserRound size={16} className="navbar-link-icon" />
              <span>我的</span>
            </NavLink>
          )}
          {user && (
            <NavLink to="/history" className={navLinkClass}>
              <Clock3 size={16} className="navbar-link-icon" />
              <span>历史</span>
            </NavLink>
          )}
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
              {registration_enabled && <Link to="/register" className="btn btn-primary btn-sm">注册</Link>}
            </div>
          )}
        </div>
      </div>
    </nav>
  )
}
