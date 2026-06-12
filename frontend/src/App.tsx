import { lazy, Suspense } from 'react'
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import { AuthProvider } from './contexts/AuthContext'
import { AdminAuthProvider, useAdminAuth } from './contexts/AdminAuthContext'
import Navbar from './components/Navbar'
import BottomNav from './components/BottomNav'
import AdminLayout from './components/AdminLayout'
import HomePage from './pages/HomePage'
import ParticleBackground from './components/ParticleBackground'
import { ToastProvider } from './components/ui/Toast'
import './index.css'
import './App.css'

// 除首页外的页面按路由懒加载，避免首屏下载 admin/合婚/导出等全部代码
const LoginPage = lazy(() => import('./pages/LoginPage'))
const RegisterPage = lazy(() => import('./pages/RegisterPage'))
const ResultPage = lazy(() => import('./pages/ResultPage'))
const HistoryPage = lazy(() => import('./pages/HistoryPage'))
const ProfilePage = lazy(() => import('./pages/ProfilePage'))
const BrandSettingsPage = lazy(() => import('./pages/BrandSettingsPage'))
const PastEventsPage = lazy(() => import('./pages/PastEventsPage'))
const CompatibilityPage = lazy(() => import('./pages/CompatibilityPage'))
const CompatibilityHistoryPage = lazy(() => import('./pages/CompatibilityHistoryPage'))
const CompatibilityResultPage = lazy(() => import('./pages/CompatibilityResultPage'))
const AdminLoginPage = lazy(() => import('./pages/admin/AdminLoginPage'))
const AdminDashboardPage = lazy(() => import('./pages/admin/AdminDashboardPage'))
const AdminLLMPage = lazy(() => import('./pages/admin/AdminLLMPage'))
const AdminUsersPage = lazy(() => import('./pages/admin/AdminUsersPage'))
const AdminAILogsPage = lazy(() => import('./pages/admin/AdminAILogsPage'))
const AdminChartsPage = lazy(() => import('./pages/admin/AdminChartsPage'))
const AdminCompatPage = lazy(() => import('./pages/admin/AdminCompatPage'))
const AdminCelebritiesPage = lazy(() => import('./pages/admin/AdminCelebritiesPage'))
const PromptSettings = lazy(() => import('./pages/admin/PromptSettings'))
const AlgoConfigPage = lazy(() => import('./pages/admin/AlgoConfigPage'))
const CleanupConfigPage = lazy(() => import('./pages/admin/CleanupConfigPage'))
const TokenUsagePage = lazy(() => import('./pages/admin/TokenUsagePage'))
const ShenshaAnnotationsPage = lazy(() => import('./pages/admin/ShenshaAnnotationsPage'))

// 懒加载 chunk 下载期间的占位（复用全局 .skeleton 样式）
function RouteFallback() {
  return (
    <div className="page">
      <div className="container">
        <div className="skeleton route-fallback-skeleton" />
      </div>
    </div>
  )
}

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
          <ToastProvider>
            <Suspense fallback={<RouteFallback />}>
            <Routes>
              {/* 普通用户路由 */}
              <Route path="/" element={<><Navbar /><BottomNav /><HomePage /></>} />
              <Route path="/login" element={<><Navbar /><BottomNav /><LoginPage /></>} />
              <Route path="/register" element={<><Navbar /><BottomNav /><RegisterPage /></>} />
              <Route path="/result" element={<><Navbar /><BottomNav /><ResultPage /></>} />
              <Route path="/history" element={<><Navbar /><BottomNav /><HistoryPage /></>} />
              <Route path="/history/:id" element={<><Navbar /><BottomNav /><ResultPage /></>} />
              <Route path="/profile" element={<><Navbar /><BottomNav /><ProfilePage /></>} />
              <Route path="/settings/brand" element={<><Navbar /><BottomNav /><BrandSettingsPage /></>} />
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
                <Route path="compatibility" element={<AdminCompatPage />} />
                <Route path="ai-logs" element={<AdminAILogsPage />} />
                <Route path="prompts" element={<PromptSettings />} />
                <Route path="algo-config" element={<AlgoConfigPage />} />
                <Route path="cleanup-config" element={<CleanupConfigPage />} />
                <Route path="token-usage" element={<TokenUsagePage />} />
                <Route path="shensha-annotations" element={<ShenshaAnnotationsPage />} />
              </Route>
            </Routes>
            </Suspense>
          </ToastProvider>
        </AuthProvider>
      </AdminAuthProvider>
    </BrowserRouter>
  )
}
