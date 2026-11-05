import { useCallback, useEffect, useRef, useState } from 'react'
import { FileText, X } from 'lucide-react'
import { adminActionErrorMessage, adminTodayQuestionsAPI } from '../../lib/adminApi'
import './AdminTodayQuestionsPage.css'

interface QuestionRow {
  id: string
  user_id: string
  user_email: string
  user_nickname: string
  chart_id: string
  question_date: string
  category: string
  question: string
  status: 'processing' | 'success' | 'insufficient' | 'error' | 'legacy'
  provider_name: string
  model: string
  duration_ms: number
  prompt_tokens: number
  completion_tokens: number
  total_tokens: number
  reasoning_tokens: number
  cache_hit_tokens: number
  cache_miss_tokens: number
  estimated_cost_cny: number
  error_msg: string
  created_at: string
}

interface QuestionDetail extends QuestionRow {
  context_data: unknown
  answer_data: unknown | null
  ai_system_prompt: string
  ai_user_prompt: string
  ai_output: string
}

const statusLabels: Record<QuestionRow['status'], string> = {
  processing: '处理中', success: '成功', insufficient: '依据不足', error: '失败', legacy: '历史调用',
}
const categoryLabels: Record<string, string> = {
  interview: '面试', work: '工作沟通', travel: '重要出行', social: '见面社交', general: '其他问题',
}

function money(value: number) { return `¥${Number(value || 0).toFixed(6)}` }
function time(value: string) { return new Date(value).toLocaleString('zh-CN') }
function content(value: unknown, empty = '无记录') {
  if (value == null || value === '') return empty
  return typeof value === 'string' ? value : JSON.stringify(value, null, 2)
}

export default function AdminTodayQuestionsPage() {
  const [filters, setFilters] = useState({ from: '', to: '', status: '', user: '' })
  const [applied, setApplied] = useState(filters)
  const [page, setPage] = useState(1)
  const [total, setTotal] = useState(0)
  const [items, setItems] = useState<QuestionRow[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [revision, setRevision] = useState(0)
  const [detail, setDetail] = useState<QuestionDetail | null>(null)
  const [detailLoading, setDetailLoading] = useState(false)
  const [detailError, setDetailError] = useState('')
  const [dailyLimit, setDailyLimit] = useState('1')
  const [savedDailyLimit, setSavedDailyLimit] = useState(1)
  const [settingsLoading, setSettingsLoading] = useState(true)
  const [settingsSaving, setSettingsSaving] = useState(false)
  const [settingsMessage, setSettingsMessage] = useState('')
  const [settingsRevision, setSettingsRevision] = useState(0)
  const listRequest = useRef(0)
  const detailRequest = useRef(0)

  const load = useCallback(async () => {
    const id = ++listRequest.current
    setLoading(true)
    setError('')
    try {
      const response = await adminTodayQuestionsAPI.list({ page, limit: 20, ...applied })
      if (id !== listRequest.current) return
      setItems(response.data.data.items || [])
      setTotal(response.data.data.total || 0)
    } catch (reason) {
      if (id === listRequest.current) setError(reason instanceof Error ? reason.message : '读取失败')
    } finally {
      if (id === listRequest.current) setLoading(false)
    }
  }, [applied, page])

  useEffect(() => {
    const requestRef = listRequest
    void load()
    return () => { requestRef.current++ }
  }, [load, revision])

  useEffect(() => {
    let active = true
    void adminTodayQuestionsAPI.settings().then(response => {
      if (!active) return
      const limit = response.data.data.daily_limit
      setDailyLimit(String(limit))
      setSavedDailyLimit(limit)
      setSettingsMessage('')
    }).catch(reason => {
      if (active) setSettingsMessage(`读取设置失败：${adminActionErrorMessage(reason)}`)
    }).finally(() => { if (active) setSettingsLoading(false) })
    return () => { active = false }
  }, [settingsRevision])

  async function saveDailyLimit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault()
    const limit = Number(dailyLimit)
    if (!Number.isInteger(limit) || limit < 1 || limit > 100) {
      setSettingsMessage('请输入 1—100 的整数')
      return
    }
    setSettingsSaving(true)
    setSettingsMessage('')
    try {
      const response = await adminTodayQuestionsAPI.updateSettings(limit)
      setSavedDailyLimit(response.data.data.daily_limit)
      setDailyLimit(String(response.data.data.daily_limit))
      setSettingsMessage('已保存，后续请求立即按新次数限制')
    } catch (reason) {
      setSettingsMessage(`保存失败：${adminActionErrorMessage(reason)}`)
    } finally {
      setSettingsSaving(false)
    }
  }

  function applyFilters(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault()
    if (filters.from && filters.to && filters.from > filters.to) {
      setError('起始日期不能晚于结束日期')
      return
    }
    setPage(1)
    setApplied({ ...filters })
  }

  async function openDetail(id: string) {
    const request = ++detailRequest.current
    setDetail(null)
    setDetailLoading(true)
    setDetailError('')
    try {
      const response = await adminTodayQuestionsAPI.detail(id)
      if (request === detailRequest.current) setDetail(response.data.data)
    } catch (reason) {
      if (request === detailRequest.current) setDetailError(reason instanceof Error ? reason.message : '读取详情失败')
    } finally {
      if (request === detailRequest.current) setDetailLoading(false)
    }
  }

  function closeDetail() {
    detailRequest.current++
    setDetail(null)
    setDetailLoading(false)
    setDetailError('')
  }

  const detailOpen = detail !== null || detailLoading || detailError !== ''
  return (
    <div className="today-admin-page">
      <div className="today-admin-heading">
        <div><h1><FileText size={23} /> 今日问事</h1><p>查看用户请求、AI 输入输出与逐次 Token 用量。完整内容仅在打开详情时读取。</p></div>
        <button type="button" className="admin-btn admin-btn-ghost" onClick={() => setRevision(value => value + 1)}>刷新</button>
      </div>

      <form className="admin-card today-admin-settings" onSubmit={event => void saveDailyLimit(event)}>
        <div><strong>每日问事次数</strong><p>每位用户每天最多问 {savedDailyLimit} 次，北京时间次日 0 点重置。AI 服务失败退回次数；「依据不足」计次。</p></div>
        <label>每人每天<input type="number" min="1" max="100" step="1" value={dailyLimit} onChange={event => setDailyLimit(event.target.value)} disabled={settingsLoading || settingsSaving} aria-label="每人每天问事次数" /></label>
        <button type="submit" className="admin-btn admin-btn-primary" disabled={settingsLoading || settingsSaving || settingsMessage.startsWith('读取设置失败') || dailyLimit === String(savedDailyLimit)}>{settingsSaving ? '保存中…' : '保存设置'}</button>
        {settingsMessage && <span className={settingsMessage.startsWith('已保存') ? 'today-admin-setting-success' : 'today-admin-error'} role="status">{settingsMessage}</span>}
        {settingsMessage.startsWith('读取设置失败') && <button type="button" className="admin-btn admin-btn-ghost" onClick={() => { setSettingsLoading(true); setSettingsRevision(value => value + 1) }}>重试读取</button>}
      </form>

      <form className="today-admin-filters admin-card" onSubmit={applyFilters}>
        <label>起始日期<input type="date" value={filters.from} onChange={event => setFilters({ ...filters, from: event.target.value })} /></label>
        <label>结束日期<input type="date" value={filters.to} onChange={event => setFilters({ ...filters, to: event.target.value })} /></label>
        <label>状态<select value={filters.status} onChange={event => setFilters({ ...filters, status: event.target.value })}><option value="">全部</option><option value="success">成功</option><option value="insufficient">依据不足</option><option value="error">失败</option><option value="processing">处理中</option><option value="legacy">历史调用</option></select></label>
        <label>用户<input value={filters.user} maxLength={100} placeholder="邮箱、昵称或用户 ID" onChange={event => setFilters({ ...filters, user: event.target.value })} /></label>
        <button type="submit" className="admin-btn admin-btn-primary">查询</button>
      </form>

      {error && <div className="today-admin-error" role="alert">{error} <button type="button" onClick={() => setRevision(value => value + 1)}>重试</button></div>}
      <div className="admin-card today-admin-results">
        <div className="today-admin-count">共 {total} 条请求{loading ? ' · 正在读取…' : ''}</div>
        {!loading && items.length === 0 && <div className="today-admin-empty">当前条件下没有问事记录</div>}
        {items.length > 0 && <div className="today-admin-table-wrap"><table className="admin-table today-admin-table"><thead><tr><th>时间 / 日期</th><th>用户</th><th>问题</th><th>状态</th><th>模型</th><th>Token</th><th>预估费用</th><th></th></tr></thead><tbody>
          {items.map(item => <tr key={item.id}>
            <td><strong>{time(item.created_at)}</strong><small>所问日 {item.question_date}</small></td>
            <td><strong>{item.user_nickname || item.user_email || item.user_id.slice(0, 8)}</strong><small>{item.user_email || item.user_id}</small></td>
            <td className="today-admin-question" title={item.question}>{item.question}</td>
            <td><span className={`today-admin-status is-${item.status}`}>{statusLabels[item.status] || item.status}</span></td>
            <td>{item.model || '未调用 AI'}</td>
            <td>{item.total_tokens ? item.total_tokens.toLocaleString('zh-CN') : '—'}</td>
            <td>{item.total_tokens ? money(item.estimated_cost_cny) : '—'}</td>
            <td><button type="button" className="admin-btn admin-btn-ghost" onClick={() => void openDetail(item.id)}>查看详情</button></td>
          </tr>)}
        </tbody></table></div>}
        <div className="today-admin-pager"><button type="button" disabled={page <= 1 || loading} onClick={() => setPage(value => value - 1)}>上一页</button><span>第 {page} / {Math.max(1, Math.ceil(total / 20))} 页</span><button type="button" disabled={page * 20 >= total || loading} onClick={() => setPage(value => value + 1)}>下一页</button></div>
      </div>

      {detailOpen && <div className="today-admin-overlay" onClick={closeDetail}>
        <section className="today-admin-detail" role="dialog" aria-modal="true" aria-label="今日问事请求详情" onClick={event => event.stopPropagation()}>
          <header><h2>今日问事请求详情</h2><button type="button" aria-label="关闭详情" onClick={closeDetail}><X size={22} /></button></header>
          {detailLoading && <p>正在读取详情…</p>}
          {detailError && <p className="today-admin-error" role="alert">{detailError}</p>}
          {detail && <div className="today-admin-detail-body">
            <dl className="today-admin-meta"><div><dt>用户</dt><dd>{detail.user_nickname || detail.user_email || detail.user_id}<small>{detail.user_id}</small></dd></div><div><dt>所问日 / 类型</dt><dd>{detail.question_date} · {categoryLabels[detail.category] || detail.category}</dd></div><div><dt>状态 / 模型</dt><dd>{statusLabels[detail.status]} · {detail.model || '未调用 AI'}<small>{detail.provider_name || '—'}</small></dd></div><div><dt>耗时</dt><dd>{detail.duration_ms ? `${detail.duration_ms.toLocaleString('zh-CN')} ms` : '—'}</dd></div><div><dt>命盘 ID</dt><dd>{detail.chart_id}</dd></div></dl>
            <section><h3>用户问题</h3><p className="today-admin-body-text">{detail.question}</p></section>
            {detail.status === 'legacy' && <p className="today-admin-legacy-note">这条记录由原 Token 日志补入；原日志没有单独保存系统指令、最终展示回答和服务端校验状态，这些内容无法还原。</p>}
            {detail.error_msg && <section><h3>失败原因</h3><p className="today-admin-error">{detail.error_msg}</p></section>}
            <section><h3>Token 与费用</h3><div className="today-admin-token-grid"><span>输入 <strong>{detail.prompt_tokens}</strong></span><span>输出 <strong>{detail.completion_tokens}</strong></span><span>合计 <strong>{detail.total_tokens}</strong></span><span>推理 <strong>{detail.reasoning_tokens}</strong></span><span>缓存命中 <strong>{detail.cache_hit_tokens}</strong></span><span>缓存未命中 <strong>{detail.cache_miss_tokens}</strong></span><span>预估费用 <strong>{detail.total_tokens ? money(detail.estimated_cost_cny) : '未产生用量'}</strong></span></div></section>
            <section><h3>服务端命理依据</h3><pre>{content(detail.context_data)}</pre></section>
            <section><h3>AI 输入 · 系统指令</h3><pre>{content(detail.ai_system_prompt, detail.status === 'legacy' ? '历史 Token 日志未单独保存系统指令，无法还原原文。' : '未调用 AI，暂无系统指令记录。')}</pre></section>
            <section><h3>AI 输入 · 任务说明与数据</h3><pre>{content(detail.ai_user_prompt, detail.status === 'legacy' ? '历史 Token 日志未保存此项。' : '未调用 AI，暂无输入记录。')}</pre></section>
            <section><h3>AI 原始输出</h3><pre>{content(detail.ai_output, detail.status === 'legacy' ? '历史 Token 日志未保存此项。' : detail.ai_system_prompt ? 'AI 未返回内容。' : '未调用 AI，暂无输出记录。')}</pre></section>
            <section><h3>最终展示回答</h3><pre>{content(detail.answer_data, detail.status === 'legacy' ? '历史 Token 日志未保存最终展示回答，无法还原。' : '无最终展示回答。')}</pre></section>
          </div>}
        </section>
      </div>}
    </div>
  )
}
