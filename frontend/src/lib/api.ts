import axios from 'axios'

const api = axios.create({
  baseURL: import.meta.env.VITE_API_URL || '',  // 开发环境走 Vite proxy → 9002
  timeout: 300000, // 300s，适配最新深度推理大模型长等待
})

// 请求拦截器：自动注入 JWT
api.interceptors.request.use((config) => {
  const token = localStorage.getItem('yj_token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

// 响应拦截器：统一错误处理
api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      localStorage.removeItem('yj_token')
      localStorage.removeItem('yj_user')
      window.location.href = '/login'
    }
    const isTimeout = error.code === 'ECONNABORTED' || error.message?.includes('timeout')
    const message = isTimeout
      ? 'AI 生成超时，Kimi K2.5 是推理模型，通常需要 30~60 秒，请再试一次'
      : (error.response?.data?.error || '请求失败，请稍后重试')
    return Promise.reject(new Error(message))
  }
)

// ======= Auth API =======
export const authAPI = {
  register: (data: { email: string; password: string; nickname?: string }) =>
    api.post('/api/auth/register', data),
  login: (data: { email: string; password: string }) =>
    api.post('/api/auth/login', data),
  me: () => api.get('/api/auth/me'),
}

// ======= 结构化报告类型 =======
export interface ReportChapter {
  title: string
  brief: string
  detail: string
}

export interface ReportAnalysis {
  logic: string
  summary: string
}

export interface StructuredReport {
  yongshen: string
  jishen: string
  analysis: ReportAnalysis
  chapters: ReportChapter[]
}

export interface AIReport {
  id: string
  chart_id: string
  content: string
  content_structured?: StructuredReport | null
  model: string
  created_at: string
}

// ======= Bazi API =======
export interface CalculateInput {
  year: number
  month: number
  day: number
  hour: number
  gender: 'male' | 'female'
  is_early_zishi?: boolean
  longitude?: number  // 出生地经度，用于真太阳时修正
  calendar_type?: 'solar' | 'lunar'
  is_leap_month?: boolean
}

export const baziAPI = {
  calculate: (data: CalculateInput) => api.post('/api/bazi/calculate', data),
  generateReport: (chartId: string) =>
    api.post(`/api/bazi/report/${chartId}`, {}, { timeout: 300000 }), // 推理模型最长 300s
  generateReportStream: async (chartId: string, onMessage: (msg: string) => void, onError: (err: string) => void, onDone: () => void) => {
    const token = localStorage.getItem('yj_token')
    const baseURL = import.meta.env.VITE_API_URL || ''
    try {
      const response = await fetch(`${baseURL}/api/bazi/report-stream/${chartId}`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          ...(token ? { 'Authorization': `Bearer ${token}` } : {})
        }
      })
      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`)
      }
      const reader = response.body?.getReader()
      if (!reader) throw new Error('No reader available')

      const decoder = new TextDecoder()
      let buffer = ''
      while (true) {
        const { done, value } = await reader.read()
        if (done) break
        buffer += decoder.decode(value, { stream: true })
        const lines = buffer.split('\n')
        buffer = lines.pop() || ''
        for (const line of lines) {
          if (line.startsWith('data: ')) {
            const data = line.slice(6)
            if (data === '[DONE]') {
              onDone()
              return
            } else {
              onMessage(data)
            }
          } else if (line.startsWith('event: error')) {
            // Usually SSE format is event: error\ndata: MSG, but we send event manually as SSEvent("error", msg)
            // Gin c.SSEvent("error", "message") -> event: error\ndata: message\n\n
          }
        }
        // Handle gin SSE errors
        if (lines.some(l => l.startsWith('event: error'))) {
          const errLine = lines.find(l => l.startsWith('data: ') && lines.indexOf(l) > lines.findIndex(e => e.startsWith('event: error')))
          if (errLine) onError(errLine.slice(6))
        }
        if (lines.some(l => l.startsWith('event: done'))) {
           onDone()
           return
        }
      }
      onDone()
    } catch (err: any) {
      onError(err.message)
    }
  },
  generateLiunianReport: (chartId: string, targetYear: number) =>
    api.post(`/api/bazi/liunian-report/${chartId}`, { target_year: targetYear }, { timeout: 300000 }),
  getHistory: (page = 1) => api.get(`/api/bazi/history?page=${page}`),
  getHistoryDetail: (id: string) => api.get(`/api/bazi/history/${id}`),
  fetchLiuYue: (liuNianYear: number, dayGan: string) =>
    api.post('/api/bazi/liu-yue', { liu_nian_year: liuNianYear, day_gan: dayGan }),
}

// ======= 流月类型 =======
export interface LiuYueItem {
  index: number
  month_name: string
  gan_zhi: string
  gan_shishen: string
  zhi_shishen: string
  jie_qi_name: string
  start_date: string // YYYY-MM-DD
  end_date: string   // YYYY-MM-DD
}

export interface LiuYueResponse {
  liu_yue: LiuYueItem[]
  current_month_index: number // -1 表示不在该命理年内（无高亮）
}

// ======= 神煞注解类型 =======
export interface ShenshaAnnotation {
  id: string
  name: string
  polarity: 'ji' | 'xiong' | 'zhong'
  description: string
  updated_at: string
}

// 获取全部神煞注解（公开，无需鉴权）
export const fetchShenshaAnnotations = (): Promise<ShenshaAnnotation[]> =>
  api.get('/api/shensha/annotations').then(res => res.data.data ?? [])

export default api
