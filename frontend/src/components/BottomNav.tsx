import { Link, useLocation } from 'react-router-dom'
import { Compass, HeartHandshake, User } from 'lucide-react'
import { useAuth } from '../contexts/AuthContext'
import './BottomNav.css'

export default function BottomNav() {
  const { user } = useAuth()
  const location = useLocation()

  return (
    <nav className="bottom-nav">
      <Link 
        to="/" 
        className={`bottom-nav-item ${location.pathname === '/' ? 'active' : ''}`}
      >
        <Compass size={20} className="bottom-nav-icon" />
        <span>测算</span>
      </Link>

      <Link
        to="/compatibility"
        className={`bottom-nav-item ${location.pathname.startsWith('/compatibility') ? 'active' : ''}`}
      >
        <HeartHandshake size={20} className="bottom-nav-icon" />
        <span>合盘</span>
      </Link>
      
      {user ? (
        <Link 
          to="/profile" 
          className={`bottom-nav-item ${location.pathname.startsWith('/profile') ? 'active' : ''}`}
        >
          <User size={20} className="bottom-nav-icon" />
          <span>我的</span>
        </Link>
      ) : (
        <Link 
          to="/login" 
          className={`bottom-nav-item ${location.pathname === '/login' ? 'active' : ''}`}
        >
          <User size={20} className="bottom-nav-icon" />
          <span>我的</span>
        </Link>
      )}
    </nav>
  )
}
