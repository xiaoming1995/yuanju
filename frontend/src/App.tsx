import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import { AuthProvider } from './contexts/AuthContext'
import { AdminAuthProvider, useAdminAuth } from './contexts/AdminAuthContext'
import Navbar from './components/Navbar'
import BottomNav from './components/BottomNav'
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
import AdminAILogsPage from './pages/admin/AdminAILogsPage'
import AdminChartsPage from './pages/admin/AdminChartsPage'
import AdminCelebritiesPage from './pages/admin/AdminCelebritiesPage'
import PromptSettings from './pages/admin/PromptSettings'
import AlgoConfigPage from './pages/admin/AlgoConfigPage'
import TokenUsagePage from './pages/admin/TokenUsagePage'
import ParticleBackground from './components/ParticleBackground'
import PastEventsPage from './pages/PastEventsPage'
import CompatibilityPage from './pages/CompatibilityPage'
import CompatibilityHistoryPage from './pages/CompatibilityHistoryPage'
import CompatibilityResultPage from './pages/CompatibilityResultPage'
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
      <ParticleBackground />
      <AdminAuthProvider>
        <AuthProvider>
          <Routes>
            {/* 普通用户路由 */}
            <Route path="/" element={<><Navbar /><BottomNav /><HomePage /></>} />
            <Route path="/login" element={<><Navbar /><BottomNav /><LoginPage /></>} />
            <Route path="/register" element={<><Navbar /><BottomNav /><RegisterPage /></>} />
            <Route path="/result" element={<><Navbar /><BottomNav /><ResultPage /></>} />
            <Route path="/history" element={<><Navbar /><BottomNav /><HistoryPage /></>} />
            <Route path="/history/:id" element={<><Navbar /><BottomNav /><ResultPage /></>} />
            <Route path="/bazi/:chartId/past-events" element={<><Navbar /><BottomNav /><PastEventsPage /></>} />
            <Route path="/compatibility" element={<><Navbar /><BottomNav /><CompatibilityPage /></>} />
            <Route path="/compatibility/history" element={<><Navbar /><BottomNav /><CompatibilityHistoryPage /></>} />
            <Route path="/compatibility/:id" element={<><Navbar /><BottomNav /><CompatibilityResultPage /></>} />

            {/* Admin 路由（独立布局，无 Navbar）*/}
            <Route path="/admin/login" element={<AdminLoginPage />} />
            <Route path="/admin" element={
              <AdminGuard><AdminLayout /></AdminGuard>
            }>
              <Route index element={<Navigate to="/admin/dashboard" replace />} />
              <Route path="dashboard" element={<AdminDashboardPage />} />
              <Route path="llm" element={<AdminLLMPage />} />
              <Route path="users" element={<AdminUsersPage />} />
              <Route path="celebrities" element={<AdminCelebritiesPage />} />
              <Route path="charts" element={<AdminChartsPage />} />
              <Route path="ai-logs" element={<AdminAILogsPage />} />
              <Route path="prompts" element={<PromptSettings />} />
              <Route path="algo-config" element={<AlgoConfigPage />} />
              <Route path="token-usage" element={<TokenUsagePage />} />
            </Route>
          </Routes>
        </AuthProvider>
      </AdminAuthProvider>
    </BrowserRouter>
  )
}
