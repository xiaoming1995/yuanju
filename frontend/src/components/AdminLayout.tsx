import { NavLink, useNavigate, Outlet } from 'react-router-dom'
import { Hexagon, LayoutDashboard, Bot, Users, FileText, BookOpen, SlidersHorizontal, BarChart2 } from 'lucide-react'
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
          <span className="admin-brand-icon" style={{ display: 'flex', alignItems: 'center' }}>
            <Hexagon size={24} />
          </span>
          <div>
            <div className="admin-brand-title">缘聚后台</div>
            <div className="admin-brand-sub">管理控制台</div>
          </div>
        </div>
        <nav className="admin-nav">
          <NavLink to="/admin/dashboard" className={({isActive}) => isActive ? 'admin-nav-item active' : 'admin-nav-item'}>
            <span style={{ display: 'flex', alignItems: 'center', marginRight: 10 }}><LayoutDashboard size={18} /></span> 数据概览
          </NavLink>
          <NavLink to="/admin/llm" className={({isActive}) => isActive ? 'admin-nav-item active' : 'admin-nav-item'}>
            <span style={{ display: 'flex', alignItems: 'center', marginRight: 10 }}><Bot size={18} /></span> LLM 管理
          </NavLink>
          <NavLink to="/admin/users" className={({isActive}) => isActive ? 'admin-nav-item active' : 'admin-nav-item'}>
            <span style={{ display: 'flex', alignItems: 'center', marginRight: 10 }}><Users size={18} /></span> 用户管理
          </NavLink>
          <NavLink to="/admin/celebrities" className={({isActive}) => isActive ? 'admin-nav-item active' : 'admin-nav-item'}>
            <span style={{ display: 'flex', alignItems: 'center', marginRight: 10 }}><Users size={18} /></span> 名人库管理
          </NavLink>
          <NavLink to="/admin/charts" className={({isActive}) => isActive ? 'admin-nav-item active' : 'admin-nav-item'}>
            <span style={{ display: 'flex', alignItems: 'center', marginRight: 10 }}><BookOpen size={18} /></span> 起盘明细
          </NavLink>
          <NavLink to="/admin/ai-logs" className={({isActive}) => isActive ? 'admin-nav-item active' : 'admin-nav-item'}>
            <span style={{ display: 'flex', alignItems: 'center', marginRight: 10 }}><FileText size={18} /></span> AI 调用日志
          </NavLink>
          <NavLink to="/admin/prompts" className={({isActive}) => isActive ? 'admin-nav-item active' : 'admin-nav-item'}>
            <span style={{ display: 'flex', alignItems: 'center', marginRight: 10 }}><Bot size={18} /></span> AI 指令设定
          </NavLink>
          <NavLink to="/admin/algo-config" className={({isActive}) => isActive ? 'admin-nav-item active' : 'admin-nav-item'}>
            <span style={{ display: 'flex', alignItems: 'center', marginRight: 10 }}><SlidersHorizontal size={18} /></span> 算法参数配置
          </NavLink>
          <NavLink to="/admin/token-usage" className={({isActive}) => isActive ? 'admin-nav-item active' : 'admin-nav-item'}>
            <span style={{ display: 'flex', alignItems: 'center', marginRight: 10 }}><BarChart2 size={18} /></span> Token 用量
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
