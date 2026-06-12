import { Link, useLocation } from 'react-router-dom'
import { Compass, HeartHandshake, History, User } from 'lucide-react'
import { useAuth } from '../contexts/AuthContext'
import { buildAuthPath } from '../lib/authRedirect'
import './BottomNav.css'

type Tab = 'cast' | 'compat' | 'records' | 'me'

// 路由 → 底栏归属的显式映射；新增页面时在这里登记
function activeTab(pathname: string): Tab | null {
  if (pathname === '/' || pathname === '/result') return 'cast'
  if (pathname.startsWith('/compatibility')) return 'compat'
  if (pathname.startsWith('/history') || pathname.startsWith('/bazi/')) return 'records'
  if (pathname.startsWith('/profile') || pathname.startsWith('/settings')) return 'me'
  if (pathname === '/login' || pathname === '/register') return 'me'
  return null
}

export default function BottomNav() {
  const { user } = useAuth()
  const location = useLocation()
  const tab = activeTab(location.pathname)
  const itemClass = (t: Tab) => `bottom-nav-item ${tab === t ? 'active' : ''}`

  return (
    <nav className="bottom-nav">
      <Link to="/" className={itemClass('cast')}>
        <Compass size={20} className="bottom-nav-icon" />
        <span>测算</span>
      </Link>

      <Link to="/compatibility" className={itemClass('compat')}>
        <HeartHandshake size={20} className="bottom-nav-icon" />
        <span>合盘</span>
      </Link>

      <Link to={user ? '/history' : buildAuthPath('/login', '/history')} className={itemClass('records')}>
        <History size={20} className="bottom-nav-icon" />
        <span>记录</span>
      </Link>

      <Link to={user ? '/profile' : '/login'} className={itemClass('me')}>
        <User size={20} className="bottom-nav-icon" />
        <span>我的</span>
      </Link>
    </nav>
  )
}
