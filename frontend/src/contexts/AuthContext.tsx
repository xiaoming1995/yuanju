import { createContext, useContext, useState, useEffect } from 'react'
import type { ReactNode } from 'react'
import { useNavigate } from 'react-router-dom'
import { authAPI } from '../lib/api'

interface User {
  id: string
  email: string
  nickname: string
}

interface AuthContextType {
  user: User | null
  isLoading: boolean
  login: (token: string, user: User) => void
  logout: () => void
}

const AuthContext = createContext<AuthContextType | null>(null)

function getStoredUser(): User | null {
  const raw = localStorage.getItem('yj_user')
  if (!raw) return null
  try {
    return JSON.parse(raw) as User
  } catch {
    localStorage.removeItem('yj_user')
    return null
  }
}

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(() => getStoredUser())
  const [isLoading, setIsLoading] = useState(() => !!localStorage.getItem('yj_token'))
  const navigate = useNavigate()

  // api 层 401 时派发该事件（storage 已清），这里清状态并走路由跳转
  useEffect(() => {
    const onUnauthorized = () => {
      setUser(null)
      navigate('/login')
    }
    window.addEventListener('yj:unauthorized', onUnauthorized)
    return () => window.removeEventListener('yj:unauthorized', onUnauthorized)
  }, [navigate])

  useEffect(() => {
    const token = localStorage.getItem('yj_token')
    if (token) {
      authAPI.me()
        .then(res => setUser(res.data.user))
        .catch(() => {
          localStorage.removeItem('yj_token')
          localStorage.removeItem('yj_user')
        })
        .finally(() => setIsLoading(false))
    }
  }, [])

  const login = (token: string, userData: User) => {
    localStorage.setItem('yj_token', token)
    localStorage.setItem('yj_user', JSON.stringify(userData))
    setUser(userData)
  }

  const logout = () => {
    localStorage.removeItem('yj_token')
    localStorage.removeItem('yj_user')
    setUser(null)
  }

  return (
    <AuthContext.Provider value={{ user, isLoading, login, logout }}>
      {children}
    </AuthContext.Provider>
  )
}

// eslint-disable-next-line react-refresh/only-export-components
export const useAuth = () => {
  const ctx = useContext(AuthContext)
  if (!ctx) throw new Error('useAuth 必须在 AuthProvider 内使用')
  return ctx
}
