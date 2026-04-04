import { createContext, useContext, useState } from 'react'
import type { ReactNode } from 'react'

interface AdminUser { id: string; email: string; name: string }
interface AdminAuthContextType {
  admin: AdminUser | null
  token: string | null
  login: (token: string, admin: AdminUser) => void
  logout: () => void
  isLoading: boolean
}

const AdminAuthContext = createContext<AdminAuthContextType | null>(null)

export function AdminAuthProvider({ children }: { children: ReactNode }) {
  const [admin, setAdmin] = useState<AdminUser | null>(() => {
    try {
      const a = localStorage.getItem('yj_admin_user')
      return a ? JSON.parse(a) : null
    } catch {
      return null
    }
  })
  const [token, setToken] = useState<string | null>(() => localStorage.getItem('yj_admin_token'))
  const [isLoading] = useState(false)

  const login = (t: string, a: AdminUser) => {
    localStorage.setItem('yj_admin_token', t)
    localStorage.setItem('yj_admin_user', JSON.stringify(a))
    setToken(t); setAdmin(a)
  }

  const logout = () => {
    localStorage.removeItem('yj_admin_token')
    localStorage.removeItem('yj_admin_user')
    setToken(null); setAdmin(null)
  }

  return (
    <AdminAuthContext.Provider value={{ admin, token, login, logout, isLoading }}>
      {children}
    </AdminAuthContext.Provider>
  )
}

// eslint-disable-next-line react-refresh/only-export-components
export function useAdminAuth() {
  const ctx = useContext(AdminAuthContext)
  if (!ctx) throw new Error('useAdminAuth must be used within AdminAuthProvider')
  return ctx
}
