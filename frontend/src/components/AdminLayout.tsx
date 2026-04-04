import { NavLink, useNavigate, Outlet } from 'react-router-dom'
import { useAdminAuth } from '../contexts/AdminAuthContext'
import './AdminLayout.css'

export default function AdminLayout() {
  const { admin, logout } = useAdminAuth()
  const navigate = useNavigate()

  const handleLogout = () => { logout(); navigate('/admin/login') }

  return (
    <div className="admin-layout">
      <aside className="admin-sidebar">
        <div className="admin-brand">
          <span className="admin-brand-icon">⚙</span>
          <div>
            <div className="admin-brand-title">缘聚后台</div>
            <div className="admin-brand-sub">管理控制台</div>
          </div>
        </div>
        <nav className="admin-nav">
          <NavLink to="/admin/dashboard" className={({isActive}) => isActive ? 'admin-nav-item active' : 'admin-nav-item'}>
            <span>📊</span> 数据概览
          </NavLink>
          <NavLink to="/admin/llm" className={({isActive}) => isActive ? 'admin-nav-item active' : 'admin-nav-item'}>
            <span>🤖</span> LLM 管理
          </NavLink>
          <NavLink to="/admin/users" className={({isActive}) => isActive ? 'admin-nav-item active' : 'admin-nav-item'}>
            <span>👥</span> 用户列表
          </NavLink>
        </nav>
        <div className="admin-sidebar-footer">
          <div className="admin-user-info">{admin?.email}</div>
          <button className="admin-logout-btn" onClick={handleLogout}>退出登录</button>
        </div>
      </aside>
      <main className="admin-main">
        <Outlet />
      </main>
    </div>
  )
}
