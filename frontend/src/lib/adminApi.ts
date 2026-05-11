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
  test: (id: string) =>
    adminApi.post(`/api/admin/llm-providers/${id}/test`, {}),
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
  getLiunianReports: (chartId: string) =>
    adminApi.get(`/api/admin/charts/${chartId}/liunian`),
  deleteLiunianReport: (id: string) =>
    adminApi.delete(`/api/admin/liunian/${id}`),
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

export const adminPromptsAPI = {
  list: () => adminApi.get('/api/admin/prompts'),
  update: (module: string, data: { content: string }) =>
    adminApi.put(`/api/admin/prompts/${module}`, data)
}

export const adminAlgoConfigAPI = {
  list: () => adminApi.get('/api/admin/algo-config'),
  update: (key: string, data: { value: string; description?: string }) =>
    adminApi.put(`/api/admin/algo-config/${key}`, data),
  reload: () => adminApi.post('/api/admin/algo-config/reload'),
}

export const adminAlgoTiaohouAPI = {
  list: (dayGan?: string) =>
    adminApi.get('/api/admin/algo-tiaohou' + (dayGan ? `?day_gan=${dayGan}` : '')),
  update: (dayGan: string, monthZhi: string, data: { xi_elements: string; text?: string }) =>
    adminApi.put(`/api/admin/algo-tiaohou/${dayGan}/${monthZhi}`, data),
  delete: (dayGan: string, monthZhi: string) =>
    adminApi.delete(`/api/admin/algo-tiaohou/${dayGan}/${monthZhi}`),
}

export const adminTokenUsageAPI = {
  summary: (from: string, to: string) =>
    adminApi.get(`/api/admin/token-usage/summary?from=${from}&to=${to}`),
  detail: (userId: string, from: string, to: string, page: number, limit = 20) =>
    adminApi.get(
      `/api/admin/token-usage/detail?user_id=${userId}&from=${from}&to=${to}&page=${page}&limit=${limit}`
    ),
}

export default adminApi
