import { useState } from 'react'
import { NavLink, useNavigate, Outlet } from 'react-router-dom'
import { Hexagon, LayoutDashboard, Bot, Users, FileText, BookOpen, SlidersHorizontal, BarChart2, Star, Trash2, Heart, Newspaper, Landmark, Menu, X, QrCode } from 'lucide-react'
import { useAdminAuth } from '../contexts/AdminAuthContext'
import './AdminLayout.css'

export default function AdminLayout() {
  const { admin, logout } = useAdminAuth()
  const navigate = useNavigate()
  const [menuOpen, setMenuOpen] = useState(false)

  const handleLogout = () => { logout(); navigate('/admin/login') }

  return (
    <div className="admin-layout">
      <header className="admin-mobile-header">
        <button className="admin-menu-toggle" type="button" aria-label="打开管理导航" aria-expanded={menuOpen} aria-controls="admin-sidebar" onClick={() => setMenuOpen(true)}><Menu size={20} /></button>
        缘聚后台
      </header>
      <button className={`admin-sidebar-backdrop ${menuOpen ? 'is-open' : ''}`} type="button" aria-label="关闭管理导航" onClick={() => setMenuOpen(false)} />
      <aside id="admin-sidebar" className={`admin-sidebar ${menuOpen ? 'is-open' : ''}`}>
        <div className="admin-brand">
          <span className="admin-brand-icon" style={{ display: 'flex', alignItems: 'center' }}>
            <Hexagon size={24} />
          </span>
          <div>
            <div className="admin-brand-title">缘聚后台</div>
            <div className="admin-brand-sub">管理控制台</div>
          </div>
          <button className="admin-sidebar-close" type="button" aria-label="关闭管理导航" onClick={() => setMenuOpen(false)}><X size={18} /></button>
        </div>
        <nav className="admin-nav" onClick={event => { if ((event.target as HTMLElement).closest('a')) setMenuOpen(false) }}>
          <div className="admin-nav-group">工作台</div>
          <NavLink to="/admin/dashboard" className={({isActive}) => isActive ? 'admin-nav-item active' : 'admin-nav-item'}>
            <span style={{ display: 'flex', alignItems: 'center', marginRight: 10 }}><LayoutDashboard size={18} /></span> 数据概览
          </NavLink>
          <NavLink to="/admin/users" className={({isActive}) => isActive ? 'admin-nav-item active' : 'admin-nav-item'}>
            <span style={{ display: 'flex', alignItems: 'center', marginRight: 10 }}><Users size={18} /></span> 用户管理
          </NavLink>
          <NavLink to="/admin/celebrities" className={({isActive}) => isActive ? 'admin-nav-item active' : 'admin-nav-item'}>
            <span style={{ display: 'flex', alignItems: 'center', marginRight: 10 }}><Users size={18} /></span> 名人库管理
          </NavLink>
          <NavLink to="/admin/mingge-historical-figures" className={({isActive}) => isActive ? 'admin-nav-item active' : 'admin-nav-item'}>
            <span style={{ display: 'flex', alignItems: 'center', marginRight: 10 }}><Landmark size={18} /></span> 命格古人映照
          </NavLink>
          <NavLink to="/admin/charts" className={({isActive}) => isActive ? 'admin-nav-item active' : 'admin-nav-item'}>
            <span style={{ display: 'flex', alignItems: 'center', marginRight: 10 }}><BookOpen size={18} /></span> 起盘明细
          </NavLink>
          <NavLink to="/admin/compatibility" className={({isActive}) => isActive ? 'admin-nav-item active' : 'admin-nav-item'}>
            <span style={{ display: 'flex', alignItems: 'center', marginRight: 10 }}><Heart size={18} /></span> 合盘明细
          </NavLink>
          <div className="admin-nav-group">内容与 AI</div>
          <NavLink to="/admin/articles" className={({isActive}) => isActive ? 'admin-nav-item active' : 'admin-nav-item'}>
            <span style={{ display: 'flex', alignItems: 'center', marginRight: 10 }}><Newspaper size={18} /></span> 资讯管理
          </NavLink>
          <NavLink to="/admin/llm" className={({isActive}) => isActive ? 'admin-nav-item active' : 'admin-nav-item'}>
            <span style={{ display: 'flex', alignItems: 'center', marginRight: 10 }}><Bot size={18} /></span> LLM 管理
          </NavLink>
          <NavLink to="/admin/ai-logs" className={({isActive}) => isActive ? 'admin-nav-item active' : 'admin-nav-item'}>
            <span style={{ display: 'flex', alignItems: 'center', marginRight: 10 }}><FileText size={18} /></span> AI 调用日志
          </NavLink>
          <NavLink to="/admin/prompts" className={({isActive}) => isActive ? 'admin-nav-item active' : 'admin-nav-item'}>
            <span style={{ display: 'flex', alignItems: 'center', marginRight: 10 }}><Bot size={18} /></span> AI 指令设定
          </NavLink>
          <NavLink to="/admin/token-usage" className={({isActive}) => isActive ? 'admin-nav-item active' : 'admin-nav-item'}>
            <span style={{ display: 'flex', alignItems: 'center', marginRight: 10 }}><BarChart2 size={18} /></span> Token 用量
          </NavLink>
          <div className="admin-nav-group">系统配置</div>
          <NavLink to="/admin/official-account-guide" className={({isActive}) => isActive ? 'admin-nav-item active' : 'admin-nav-item'}>
            <span style={{ display: 'flex', alignItems: 'center', marginRight: 10 }}><QrCode size={18} /></span> 公众号引导
          </NavLink>
          <NavLink to="/admin/algo-config" className={({isActive}) => isActive ? 'admin-nav-item active' : 'admin-nav-item'}>
            <span style={{ display: 'flex', alignItems: 'center', marginRight: 10 }}><SlidersHorizontal size={18} /></span> 算法参数配置
          </NavLink>
          <NavLink to="/admin/cleanup-config" className={({isActive}) => isActive ? 'admin-nav-item active' : 'admin-nav-item'}>
            <span style={{ display: 'flex', alignItems: 'center', marginRight: 10 }}><Trash2 size={18} /></span> 数据清理配置
          </NavLink>
          <NavLink to="/admin/shensha-annotations" className={({isActive}) => isActive ? 'admin-nav-item active' : 'admin-nav-item'}>
            <span style={{ display: 'flex', alignItems: 'center', marginRight: 10 }}><Star size={18} /></span> 神煞注解
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
