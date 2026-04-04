import axios from 'axios'

const adminApi = axios.create({
  baseURL: '',
  timeout: 15000,
})

adminApi.interceptors.request.use((config) => {
  const token = localStorage.getItem('yj_admin_token')
  if (token) config.headers.Authorization = `Bearer ${token}`
  return config
})

adminApi.interceptors.response.use(
  (res) => res,
  (error) => {
    if (error.response?.status === 401) {
      localStorage.removeItem('yj_admin_token')
      window.location.href = '/admin/login'
    }
    const message = error.response?.data?.error || '请求失败'
    return Promise.reject(new Error(message))
  }
)

export const adminAuthAPI = {
  register: (data: { email: string; password: string; name?: string }) =>
    adminApi.post('/api/admin/auth/register', data),
  login: (data: { email: string; password: string }) =>
    adminApi.post('/api/admin/auth/login', data),
}

export const adminLLMAPI = {
  list: () => adminApi.get('/api/admin/llm-providers'),
  create: (data: { name: string; type: string; base_url: string; model: string; api_key: string }) =>
    adminApi.post('/api/admin/llm-providers', data),
  update: (id: string, data: { name?: string; base_url?: string; model?: string; api_key?: string }) =>
    adminApi.put(`/api/admin/llm-providers/${id}`, data),
  activate: (id: string) =>
    adminApi.put(`/api/admin/llm-providers/${id}/activate`, {}),
  delete: (id: string) =>
    adminApi.delete(`/api/admin/llm-providers/${id}`),
}

export const adminStatsAPI = {
  overview: () => adminApi.get('/api/admin/stats'),
  ai: () => adminApi.get('/api/admin/stats/ai'),
  users: (page = 1, q = '') =>
    adminApi.get(`/api/admin/users?page=${page}&q=${encodeURIComponent(q)}`),
  getAILogs: (page: number = 1, pageSize: number = 20) =>
    adminApi.get(`/api/admin/ai-logs?page=${page}&pageSize=${pageSize}`),
  clearAllCache: () => adminApi.delete('/api/admin/reports/cache'),
  clearChartCache: (chartId: string) => adminApi.delete(`/api/admin/reports/cache/${chartId}`),
}

export const adminChartsAPI = {
  list: (page: number = 1, pageSize: number = 20) =>
    adminApi.get(`/api/admin/charts?page=${page}&pageSize=${pageSize}`),
}

export const adminReportAPI = {
  clearAll: () => adminApi.delete('/api/admin/reports/cache'),
  clearByChart: (chartId: string) => adminApi.delete(`/api/admin/reports/cache/${chartId}`),
}

export const adminAILogsAPI = {
  list: (page = 1, status = '') =>
    adminApi.get(`/api/admin/ai-logs?page=${page}${status ? `&status=${status}` : ''}`),
  summary: () => adminApi.get('/api/admin/ai-logs/summary'),
}

export const adminCelebritiesAPI = {
  list: () => adminApi.get('/api/admin/celebrities'),
  create: (data: { name: string; gender?: string; traits?: string; career?: string; active: boolean }) =>
    adminApi.post('/api/admin/celebrities', data),
  update: (id: string, data: { name: string; gender?: string; traits?: string; career?: string; active: boolean }) =>
    adminApi.put(`/api/admin/celebrities/${id}`, data),
  delete: (id: string) =>
    adminApi.delete(`/api/admin/celebrities/${id}`),
  generateAI: (data: { topic: string; count: number }) =>
    adminApi.post('/api/admin/celebrities/ai-generate', data, { timeout: 120000 }), // 覆盖默认 15s 超时
}

export default adminApi
