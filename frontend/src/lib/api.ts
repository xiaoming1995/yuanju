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

export interface UserProfileStats {
  chart_count: number
  ai_report_count: number
  compatibility_count: number
}

export interface UserProfileChartSummary {
  id: string
  birth_year: number
  birth_month: number
  birth_day: number
  birth_hour: number
  gender: string
  year_gan: string
  year_zhi: string
  month_gan: string
  month_zhi: string
  day_gan: string
  day_zhi: string
  hour_gan: string
  hour_zhi: string
  yongshen: string
  jishen: string
  created_at: string
}

export interface UserProfileCompatibilitySummary {
  id: string
  overall_level: string
  self_name: string
  partner_name: string
  summary_tags: string[]
  created_at: string
}

export interface UserProfileFeatureEntry {
  key: string
  title: string
  description: string
  status: 'coming_soon' | 'disabled' | 'enabled'
}

export interface UserProfileOverview {
  user: {
    id: string
    email: string
    nickname: string
    created_at: string
  }
  stats: UserProfileStats
  recent_charts: UserProfileChartSummary[]
  recent_compatibility: UserProfileCompatibilitySummary[]
  features: UserProfileFeatureEntry[]
}

export const userAPI = {
  profile: () => api.get('/api/user/profile'),
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

export interface CompatibilityProfileInput {
  year: number
  month: number
  day: number
  hour: number
  gender: 'male' | 'female'
  calendar_type?: 'solar' | 'lunar'
  is_leap_month?: boolean
}

export interface CompatibilityDimensionScores {
  attraction: number
  stability: number
  communication: number
  practicality: number
}

export interface CompatibilityDurationWindow {
  level: 'high' | 'medium' | 'low'
}

export interface CompatibilityDurationAssessment {
  overall_band: 'short_term' | 'medium_term' | 'long_term'
  summary: string
  reasons: string[]
  windows: {
    three_months: CompatibilityDurationWindow
    one_year: CompatibilityDurationWindow
    two_years_plus: CompatibilityDurationWindow
  }
}

export interface CompatibilityEvidence {
  id: string
  reading_id: string
  dimension: 'attraction' | 'stability' | 'communication' | 'practicality'
  type: string
  polarity: 'positive' | 'negative' | 'mixed' | 'neutral'
  source: string
  title: string
  detail: string
  weight: number
  created_at: string
}

export interface CompatibilityChartSnapshot {
  birth_year: number
  birth_month: number
  birth_day: number
  birth_hour: number
  gender: string
  year_gan: string
  year_zhi: string
  month_gan: string
  month_zhi: string
  day_gan: string
  day_zhi: string
  hour_gan: string
  hour_zhi: string
  wuxing?: {
    mu: number
    huo: number
    tu: number
    jin: number
    shui: number
  }
}

export interface CompatibilityParticipant {
  id: string
  reading_id: string
  role: 'self' | 'partner'
  display_name: string
  birth_profile: CompatibilityProfileInput
  chart_hash: string
  chart_snapshot?: CompatibilityChartSnapshot | null
  created_at: string
}

export interface CompatibilityReading {
  id: string
  user_id: string
  overall_level: 'high' | 'medium' | 'low'
  dimension_scores: CompatibilityDimensionScores
  duration_assessment: CompatibilityDurationAssessment
  summary_tags: string[]
  analysis_version: string
  created_at: string
  updated_at: string
}

export interface CompatibilityStructuredReport {
  summary: string
  dimensions: Array<{ key: string; title: string; content: string }>
  duration_assessment: CompatibilityDurationAssessment
  risks: string[]
  advice: string
}

export interface AICompatibilityReport {
  id: string
  reading_id: string
  content: string
  content_structured?: CompatibilityStructuredReport | null
  model: string
  created_at: string
}

export interface CompatibilityDetail {
  reading: CompatibilityReading
  participants: CompatibilityParticipant[]
  evidences: CompatibilityEvidence[]
  latest_report?: AICompatibilityReport | null
}

export interface CompatibilityHistoryItem {
  id: string
  overall_level: 'high' | 'medium' | 'low'
  dimension_scores: CompatibilityDimensionScores
  summary_tags: string[]
  self_name: string
  partner_name: string
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
  generateReportStream: async (chartId: string, onMessage: (msg: string) => void, onError: (err: string) => void, onDone: () => void, onThinking?: () => void) => {
    const token = localStorage.getItem('yj_token')
    const baseURL = import.meta.env.VITE_API_URL || ''
    let isDone = false
    const safeOnDone = () => {
      if (isDone) return
      isDone = true
      onDone()
    }
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
              safeOnDone()
              return
            } else {
              try {
                const parsed = JSON.parse(data)
                if (parsed.chunk !== undefined) {
                  onMessage(parsed.chunk)
                }
              } catch {
                // 向后容错：万一不是 JSON 则直接作为字符串
                onMessage(data)
              }
            }
          } else if (line.startsWith('event: thinking')) {
            // 推理模型进入思考阶段
            onThinking?.()
          } else if (line.startsWith('event: error')) {
            // 下一行 data: 将包含错误信息
          } else if (line.startsWith('event: done')) {
            // 由 data: [DONE] 来触发，这里无需处理
          }
        }
        // 处理 error 事件
        const errorIdx = lines.findIndex(l => l.startsWith('event: error'))
        if (errorIdx !== -1) {
          const errDataLine = lines[errorIdx + 1]
          if (errDataLine?.startsWith('data: ')) {
            onError(errDataLine.slice(6))
            return
          }
        }
      }
      safeOnDone()
    } catch (err: any) {
      onError(err.message)
    }
  },
  generateLiunianReport: (chartId: string, targetYear: number) =>
    api.post(`/api/bazi/liunian-report/${chartId}`, { target_year: targetYear }, { timeout: 300000 }),
  // 思路 E：即时返回所有年份算法叙述（无 AI）
  fetchPastEventsYears: (chartId: string) =>
    api.post(`/api/bazi/past-events/years/${chartId}`),

  // 思路 E：按大运分段流式 AI 总结
  streamDayunSummaries: async (
    chartId: string,
    onItem: (item: {
      dayun_index: number
      gan_zhi: string
      themes: string[]
      summary: string
      cached: boolean
      error?: string
    }) => void,
    onError: (err: string) => void,
    onDone: () => void,
  ) => {
    const token = localStorage.getItem('yj_token')
    const baseURL = import.meta.env.VITE_API_URL || ''
    let isDone = false
    const safeOnDone = () => { if (!isDone) { isDone = true; onDone() } }
    try {
      const response = await fetch(`${baseURL}/api/bazi/past-events/dayun-summary-stream/${chartId}`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', ...(token ? { 'Authorization': `Bearer ${token}` } : {}) },
      })
      if (!response.ok) throw new Error(`HTTP error! status: ${response.status}`)
      const reader = response.body?.getReader()
      if (!reader) throw new Error('No reader available')
      const decoder = new TextDecoder()
      let buffer = ''
      let pendingError = false
      while (true) {
        const { done, value } = await reader.read()
        if (done) break
        buffer += decoder.decode(value, { stream: true })
        const lines = buffer.split('\n')
        buffer = lines.pop() || ''
        for (const line of lines) {
          if (line.startsWith('event: error')) {
            pendingError = true
          } else if (line.startsWith('event: done')) {
            safeOnDone()
            return
          } else if (line.startsWith('data: ')) {
            const data = line.slice(6)
            if (data === '[DONE]') { safeOnDone(); return }
            if (pendingError) {
              onError(data)
              pendingError = false
              continue
            }
            try {
              const parsed = JSON.parse(data)
              onItem(parsed)
            } catch {
              // ignore unparseable
            }
          }
        }
      }
      safeOnDone()
    } catch (err: any) {
      onError(err?.message || 'unknown error')
    }
  },

  streamPastEvents: async (chartId: string, onMessage: (msg: string) => void, onError: (err: string) => void, onDone: () => void, onThinking?: () => void) => {
    const token = localStorage.getItem('yj_token')
    const baseURL = import.meta.env.VITE_API_URL || ''
    let isDone = false
    const safeOnDone = () => { if (!isDone) { isDone = true; onDone() } }
    try {
      const response = await fetch(`${baseURL}/api/bazi/past-events-stream/${chartId}`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', ...(token ? { 'Authorization': `Bearer ${token}` } : {}) }
      })
      if (!response.ok) throw new Error(`HTTP error! status: ${response.status}`)
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
            if (data === '[DONE]') { safeOnDone(); return }
            try {
              const parsed = JSON.parse(data)
              if (parsed.chunk !== undefined) onMessage(parsed.chunk)
            } catch { onMessage(data) }
          } else if (line.startsWith('event: thinking')) {
            onThinking?.()
          } else if (line.startsWith('event: error')) {
            // next line has the error
          } else if (line.startsWith('event: done')) {
            safeOnDone()
          }
        }
        const errorIdx = lines.findIndex(l => l.startsWith('event: error'))
        if (errorIdx !== -1) {
          const errLine = lines[errorIdx + 1]
          if (errLine?.startsWith('data: ')) { onError(errLine.slice(6)); return }
        }
      }
      safeOnDone()
    } catch (err: any) {
      onError(err.message)
    }
  },
  getHistory: (page = 1) => api.get(`/api/bazi/history?page=${page}`),
  getHistoryDetail: (id: string) => api.get(`/api/bazi/history/${id}`),
  fetchLiuYue: (liuNianYear: number, dayGan: string) =>
    api.post('/api/bazi/liu-yue', { liu_nian_year: liuNianYear, day_gan: dayGan }),
}

export const compatibilityAPI = {
  createReading: (data: { self: CompatibilityProfileInput; partner: CompatibilityProfileInput }) =>
    api.post('/api/compatibility/readings', data),
  getHistory: () =>
    api.get('/api/compatibility/readings'),
  getDetail: (id: string) =>
    api.get(`/api/compatibility/readings/${id}`),
  generateReport: (id: string) =>
    api.post(`/api/compatibility/readings/${id}/report`, {}, { timeout: 300000 }),
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
  category: string
  short_desc: string
  description: string
  updated_at: string
}

// 获取全部神煞注解（公开，无需鉴权）
export const fetchShenshaAnnotations = (): Promise<ShenshaAnnotation[]> =>
  api.get('/api/shensha/annotations').then(res => res.data.data ?? [])

export default api
