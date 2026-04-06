import { Link, useLocation } from 'react-router-dom'
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
        <span className="bottom-nav-icon">☯</span>
        <span>测算</span>
      </Link>
      
      {user ? (
        <Link 
          to="/history" 
          className={`bottom-nav-item ${location.pathname.startsWith('/history') ? 'active' : ''}`}
        >
          <span className="bottom-nav-icon">📜</span>
          <span>历史</span>
        </Link>
      ) : (
        <Link 
          to="/login" 
          className={`bottom-nav-item ${location.pathname === '/login' ? 'active' : ''}`}
        >
          <span className="bottom-nav-icon">👤</span>
          <span>我的</span>
        </Link>
      )}
    </nav>
  )
}
