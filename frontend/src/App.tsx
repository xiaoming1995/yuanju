import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import { AuthProvider } from './contexts/AuthContext'
import { AdminAuthProvider, useAdminAuth } from './contexts/AdminAuthContext'
import Navbar from './components/Navbar'
import AdminLayout from './components/AdminLayout'
import HomePage from './pages/HomePage'
import LoginPage from './pages/LoginPage'
import RegisterPage from './pages/RegisterPage'
import ResultPage from './pages/ResultPage'
import HistoryPage from './pages/HistoryPage'
import AdminLoginPage from './pages/admin/AdminLoginPage'
import AdminDashboardPage from './pages/admin/AdminDashboardPage'
import AdminLLMPage from './pages/admin/AdminLLMPage'
import AdminUsersPage from './pages/admin/AdminUsersPage'
import './index.css'
import './App.css'

// Admin 路由守卫（未登录跳转 /admin/login）
function AdminGuard({ children }: { children: React.ReactNode }) {
  const { admin, isLoading } = useAdminAuth()
  if (isLoading) return null
  if (!admin) return <Navigate to="/admin/login" replace />
  return <>{children}</>
}

export default function App() {
  return (
    <BrowserRouter>
      <AdminAuthProvider>
        <AuthProvider>
          <Routes>
            {/* 普通用户路由 */}
            <Route path="/" element={<><Navbar /><HomePage /></>} />
            <Route path="/login" element={<><Navbar /><LoginPage /></>} />
            <Route path="/register" element={<><Navbar /><RegisterPage /></>} />
            <Route path="/result" element={<><Navbar /><ResultPage /></>} />
            <Route path="/history" element={<><Navbar /><HistoryPage /></>} />
            <Route path="/history/:id" element={<><Navbar /><ResultPage /></>} />

            {/* Admin 路由（独立布局，无 Navbar）*/}
            <Route path="/admin/login" element={<AdminLoginPage />} />
            <Route path="/admin" element={
              <AdminGuard><AdminLayout /></AdminGuard>
            }>
              <Route index element={<Navigate to="/admin/dashboard" replace />} />
              <Route path="dashboard" element={<AdminDashboardPage />} />
              <Route path="llm" element={<AdminLLMPage />} />
              <Route path="users" element={<AdminUsersPage />} />
            </Route>
          </Routes>
        </AuthProvider>
      </AdminAuthProvider>
    </BrowserRouter>
  )
}
