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
  create: (data: { name: string; type: string; base_url: string; model: string; api_key: string; thinking_enabled?: boolean; input_price_cny?: number; output_price_cny?: number }) =>
    adminApi.post('/api/admin/llm-providers', data),
  update: (id: string, data: { name?: string; base_url?: string; model?: string; api_key?: string; thinking_enabled?: boolean; input_price_cny?: number; output_price_cny?: number }) =>
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

export const adminUsersAPI = {
  create: (data: { email: string; password: string; nickname?: string }) =>
    adminApi.post('/api/admin/users', data),
  resetPassword: (id: string, password: string) =>
    adminApi.post(`/api/admin/users/${id}/reset-password`, { password }),
  setDisabled: (id: string, disabled: boolean) =>
    adminApi.put(`/api/admin/users/${id}/disable`, { disabled }),
  remove: (id: string) =>
    adminApi.delete(`/api/admin/users/${id}`),
}

export const adminRegistrationSettingsAPI = {
  get: () => adminApi.get<{ registration_enabled: boolean }>('/api/admin/settings/registration'),
  update: (data: { registration_enabled: boolean }) =>
    adminApi.put<{ registration_enabled: boolean }>('/api/admin/settings/registration', data),
}

export const adminChartsAPI = {
  list: (page: number = 1, pageSize: number = 20, q = '', from = '', to = '') =>
    adminApi.get(`/api/admin/charts?page=${page}&pageSize=${pageSize}&q=${encodeURIComponent(q)}&from=${from}&to=${to}`),
  detail: (chartId: string) =>
    adminApi.get(`/api/admin/charts/${chartId}`),
  getLiunianReports: (chartId: string) =>
    adminApi.get(`/api/admin/charts/${chartId}/liunian`),
  getPastEventsRecords: (chartId: string) =>
    adminApi.get(`/api/admin/charts/${chartId}/past-events`),
  deleteLiunianReport: (id: string) =>
    adminApi.delete(`/api/admin/liunian/${id}`),
}

export const adminCompatAPI = {
  list: (page: number = 1, pageSize: number = 20) =>
    adminApi.get(`/api/admin/compatibility/readings?page=${page}&pageSize=${pageSize}`),
  detail: (id: string) =>
    adminApi.get(`/api/admin/compatibility/readings/${id}`),
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
    adminApi.put(`/api/admin/prompts/${module}`, data),
  resetToCanonical: (module: string) =>
    adminApi.post(`/api/admin/prompts/${module}/reset`),
  getCanonical: (module: string) =>
    adminApi.get(`/api/admin/prompts/${module}/canonical`),
}

export const adminAlgoConfigAPI = {
  list: () => adminApi.get('/api/admin/algo-config'),
  update: (key: string, data: { value: string; description?: string }) =>
    adminApi.put(`/api/admin/algo-config/${key}`, data),
  reload: () => adminApi.post('/api/admin/algo-config/reload'),
}

export interface CleanupConfig {
  enabled: boolean
  retention_days: number
  run_hour: number
}

export const adminCleanupConfigAPI = {
  get: () => adminApi.get<CleanupConfig>('/api/admin/cleanup-config'),
  update: (data: CleanupConfig) => adminApi.put<CleanupConfig>('/api/admin/cleanup-config', data),
}

export const adminAlgoTiaohouAPI = {
  list: (dayGan?: string) =>
    adminApi.get('/api/admin/algo-tiaohou' + (dayGan ? `?day_gan=${dayGan}` : '')),
  update: (dayGan: string, monthZhi: string, data: { xi_elements: string; text?: string }) =>
    adminApi.put(`/api/admin/algo-tiaohou/${dayGan}/${monthZhi}`, data),
  delete: (dayGan: string, monthZhi: string) =>
    adminApi.delete(`/api/admin/algo-tiaohou/${dayGan}/${monthZhi}`),
}

export const adminShenshaAPI = {
  list: () => adminApi.get('/api/shensha/annotations'),
  update: (name: string, data: { category: string; short_desc: string; description: string }) =>
    adminApi.put(`/api/admin/shensha-annotations/${encodeURIComponent(name)}`, data),
}

export const adminTokenUsageAPI = {
  summary: (from: string, to: string) =>
    adminApi.get(`/api/admin/token-usage/summary?from=${from}&to=${to}`),
  detail: (userID: string, from: string, to: string, page: number, limit: number, model: string) =>
    adminApi.get(
      `/api/admin/token-usage/detail?user_id=${userID}&from=${from}&to=${to}&page=${page}&limit=${limit}&model=${encodeURIComponent(model)}`
    ),
  content: (id: string) =>
    adminApi.get<{ input_content: string; output_content: string }>(`/api/admin/token-usage/content/${id}`),
  budgetStatus: () =>
    adminApi.get('/api/admin/token-usage/budget-status'),
}

export interface AdminArticleFilter {
  page?: number
  limit?: number
  status?: string
  q?: string
  category?: string
  tag?: string
  sort?: 'latest' | 'hot' | 'quality'
  min_quality_score?: number
}

export const adminArticlesAPI = {
  getModuleSettings: () =>
    adminApi.get<{ module_enabled: boolean }>('/api/admin/articles/module-settings'),
  updateModuleSettings: (data: { module_enabled: boolean }) =>
    adminApi.put<{ module_enabled: boolean }>('/api/admin/articles/module-settings', data),
  list: (params: AdminArticleFilter = {}) =>
    adminApi.get('/api/admin/articles', { params }),
  detail: (id: string) =>
    adminApi.get(`/api/admin/articles/${id}`),
  updateBody: (id: string, bodyContent: string) =>
    adminApi.put(`/api/admin/articles/${id}/body`, { body_content: bodyContent }),
  fetchBody: (id: string, url: string) =>
    adminApi.post(`/api/admin/articles/${id}/fetch-body`, { url }, { timeout: 120000 }),
  batchAction: (data: { ids: string[]; action: string; note?: string; allow_publish_without_ai?: boolean }) =>
    adminApi.post('/api/admin/articles/batch-action', data),
  generateAI: (id: string) =>
    adminApi.post(`/api/admin/articles/${id}/ai-analysis`, {}, { timeout: 120000 }),
  batchGenerateAI: (ids: string[]) =>
    adminApi.post('/api/admin/articles/ai-analysis/batch', { ids }, { timeout: 300000 }),
  categories: () =>
    adminApi.get('/api/admin/articles/categories'),
  createCategory: (data: { name: string; slug?: string; sort_order?: number; active?: boolean }) =>
    adminApi.post('/api/admin/articles/categories', data),
  updateCategory: (id: string, data: { name: string; slug?: string; sort_order?: number; active: boolean }) =>
    adminApi.put(`/api/admin/articles/categories/${id}`, data),
  tags: () =>
    adminApi.get('/api/admin/articles/tags'),
  createTag: (data: { name: string; slug?: string; active?: boolean }) =>
    adminApi.post('/api/admin/articles/tags', data),
  updateTag: (id: string, data: { name: string; slug?: string; active: boolean }) =>
    adminApi.put(`/api/admin/articles/tags/${id}`, data),
  keywords: () =>
    adminApi.get('/api/admin/articles/keywords'),
  createKeyword: (data: { keyword: string; active?: boolean }) =>
    adminApi.post('/api/admin/articles/keywords', data),
  updateKeyword: (id: string, data: { keyword: string; active: boolean }) =>
    adminApi.put(`/api/admin/articles/keywords/${id}`, data),
  collect: () =>
    adminApi.post('/api/admin/articles/collect', {}, { timeout: 120000 }),
  tasks: (page = 1, limit = 20) =>
    adminApi.get(`/api/admin/articles/collection-tasks?page=${page}&limit=${limit}`),
  taskItems: (id: string, page = 1, limit = 100) =>
    adminApi.get(`/api/admin/articles/collection-tasks/${id}/items?page=${page}&limit=${limit}`),
  retryTask: (id: string) =>
    adminApi.post(`/api/admin/articles/collection-tasks/${id}/retry`, {}, { timeout: 120000 }),
  getCollectionConfig: () =>
    adminApi.get('/api/admin/articles/collection-config'),
  updateCollectionConfig: (data: { enabled: boolean; frequency: string; schedule_interval_minutes?: number; max_results_per_run: number; search_page_min?: number; search_page_max?: number }) =>
    adminApi.put('/api/admin/articles/collection-config', data),
  getQualityConfig: () =>
    adminApi.get('/api/admin/articles/quality-config'),
  updateQualityConfig: (data: {
    quality_filter_enabled: boolean
    min_quality_score: number
    allow_without_body: boolean
    bonus_keywords: string[]
    source_blacklist: string[]
    preferred_sources: string[]
    ai_quality_check_enabled: boolean
  }) =>
    adminApi.put('/api/admin/articles/quality-config', data),
  getAIConfig: () =>
    adminApi.get('/api/admin/articles/ai-config'),
  updateAIConfig: (data: {
    prompt_content?: string
    prompt_description?: string
    provider_name?: string
    provider_type?: string
    base_url?: string
    model?: string
    api_key?: string
  }) =>
    adminApi.put('/api/admin/articles/ai-config', data),
}

export default adminApi
