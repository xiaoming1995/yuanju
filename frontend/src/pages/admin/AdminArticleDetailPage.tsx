import { useEffect, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { Archive, ArrowLeft, Bot, CheckCircle2, ExternalLink, LinkIcon, RefreshCw, Save, Trash2, XCircle } from 'lucide-react'
import { adminArticlesAPI } from '../../lib/adminApi'
import './AdminArticleDetailPage.css'

interface AdminArticleRecord {
  id: string
  title: string
  source_name?: string
  original_url?: string
  cover_url?: string
  published_at_source?: string
  search_snippet?: string
  summary?: string
  full_text_authorized?: boolean
  body_content?: string
  body_fetch_status?: string
  body_fetch_error?: string
  status: string
  view_count: number
  original_click_count?: number
  quality_score?: number
  quality_reasons?: QualityReason[]
  ai_analysis?: unknown
  ai_status?: string
  ai_error_msg?: string
  created_at?: string
  updated_at?: string
}

interface QualityReason {
  type: string
  points: number
  message: string
}

export default function AdminArticleDetailPage() {
  const { id } = useParams()
  const [article, setArticle] = useState<AdminArticleRecord | null>(null)
  const [loading, setLoading] = useState(true)
  const [message, setMessage] = useState('')
  const [bodyDraft, setBodyDraft] = useState('')
  const [fetchUrl, setFetchUrl] = useState('')
  const [bodySaving, setBodySaving] = useState(false)
  const [actionSaving, setActionSaving] = useState(false)

  const applyArticle = (next: AdminArticleRecord) => {
    setArticle(next)
    setBodyDraft(next.body_content || '')
    if (!next.original_url?.includes('weixin.sogou.com/link')) {
      setFetchUrl(next.original_url || '')
    }
  }

  const loadArticle = (showLoading = true) => {
    if (!id) return
    if (showLoading) setLoading(true)
    adminArticlesAPI.detail(id)
      .then(res => applyArticle(res.data.article))
      .finally(() => setLoading(false))
  }

  useEffect(() => {
    if (!id) return
    adminArticlesAPI.detail(id)
      .then(res => applyArticle(res.data.article))
      .finally(() => setLoading(false))
  }, [id])

  const generateAI = async () => {
    if (!id) return
    setActionSaving(true)
    setMessage('AI 分析生成中...')
    try {
      await adminArticlesAPI.generateAI(id)
      setMessage('AI 分析已更新')
      loadArticle(false)
    } catch (err) {
      setMessage(err instanceof Error ? err.message : 'AI 分析失败')
    } finally {
      setActionSaving(false)
    }
  }

  const runAction = async (action: string) => {
    if (!id || !article) return
    if (action === 'delete' && !window.confirm('确认删除这篇文章？删除后会进入“已删除”状态。')) return
    const allowPublishWithoutAI = action === 'publish' && !article.ai_analysis
    if (allowPublishWithoutAI && !window.confirm('这篇文章尚未生成 AI 分析，确认发布？')) return
    setActionSaving(true)
    setMessage('状态更新中...')
    try {
      const res = await adminArticlesAPI.batchAction({ ids: [id], action, allow_publish_without_ai: allowPublishWithoutAI })
      if (res.data.result.updated > 0) {
        setMessage('状态已更新')
        loadArticle(false)
      } else {
        setMessage('当前状态不支持该操作')
      }
    } catch (err) {
      setMessage(err instanceof Error ? err.message : '状态更新失败')
    } finally {
      setActionSaving(false)
    }
  }

  const saveBody = async () => {
    if (!id) return
    if (!bodyDraft.trim()) {
      setMessage('请先粘贴正文内容')
      return
    }
    setBodySaving(true)
    setMessage('正文保存中...')
    try {
      const res = await adminArticlesAPI.updateBody(id, bodyDraft)
      applyArticle(res.data.article)
      setMessage('正文已保存')
    } catch (err) {
      setMessage(err instanceof Error ? err.message : '正文保存失败')
    } finally {
      setBodySaving(false)
    }
  }

  const fetchBody = async () => {
    if (!id) return
    if (!fetchUrl.trim()) {
      setMessage('请先填写真实原文链接')
      return
    }
    setBodySaving(true)
    setMessage('正文补抓中...')
    try {
      const res = await adminArticlesAPI.fetchBody(id, fetchUrl)
      applyArticle(res.data.article)
      setMessage('正文已补抓并保存')
    } catch (err) {
      setMessage(err instanceof Error ? err.message : '正文补抓失败')
    } finally {
      setBodySaving(false)
    }
  }

  if (loading) {
    return <div className="admin-article-detail-page"><div className="admin-article-detail-loading">加载中...</div></div>
  }

  if (!article) {
    return (
      <div className="admin-article-detail-page">
        <Link className="admin-article-back" to="/admin/articles"><ArrowLeft size={16} /> 返回资讯列表</Link>
        <div className="admin-article-detail-empty">资讯不存在或已删除。</div>
      </div>
    )
  }

  const hasAuthorizedBody = Boolean(article.full_text_authorized && article.body_content?.trim())
  const isSogouRedirect = article.original_url?.includes('weixin.sogou.com/link')
  const bodyStatus = hasAuthorizedBody
    ? `已获取正文 · ${article.body_content?.length || 0} 字`
    : `未获取正文 · ${bodyFetchStatusLabel(article.body_fetch_status)}`

  return (
    <div className="admin-article-detail-page">
      <div className="admin-article-detail-topbar">
        <Link className="admin-article-back" to="/admin/articles">
          <ArrowLeft size={16} />
          返回审核队列
        </Link>
        <button className="admin-article-detail-ghost" onClick={() => loadArticle()}>
          <RefreshCw size={16} />
          刷新
        </button>
      </div>

      <header className="admin-article-detail-header">
        <div>
          <div className="admin-article-detail-kicker">
            {article.source_name || '未知来源'} · {formatDateTime(article.published_at_source)}
          </div>
          <h1>{article.title}</h1>
          <div className="admin-article-detail-badges">
            <span className={`admin-article-detail-badge ${article.status}`}>{statusLabel(article.status)}</span>
            <span className={`admin-article-detail-badge quality-${qualityLevel(article.quality_score || 0)}`}>质量 {article.quality_score || 0}</span>
            <span className={`admin-article-detail-badge ai-${article.ai_status || 'pending'}`}>AI {aiStatusLabel(article.ai_status)}</span>
            <span className={`admin-article-detail-badge body-${article.body_fetch_status || 'pending'}`}>{bodyStatus}</span>
          </div>
        </div>
        {message && <span className="admin-article-detail-message">{message}</span>}
      </header>

      <main className="admin-article-detail-grid">
        <div className="admin-article-detail-main">
          <section className="admin-article-detail-card">
            <div className="admin-article-detail-card-head">
              <h2>采集内容</h2>
              {article.original_url && (
                <a className="admin-article-inline-link" href={article.original_url} target="_blank" rel="noreferrer">
                  <ExternalLink size={16} />
                  原文
                </a>
              )}
            </div>
            {!hasAuthorizedBody && (
              <div className="admin-article-detail-note">
                当前仅保存公开搜索结果里的标题、来源、发布时间和摘要片段。右侧可以粘贴正文，或填入真实微信公众号原文链接补抓正文。
              </div>
            )}
            {!hasAuthorizedBody && article.body_fetch_error && (
              <div className="admin-article-detail-note error">
                {article.body_fetch_error}
              </div>
            )}
            <DetailField label="搜索摘要" value={article.search_snippet || '无'} multiline />
            <DetailField label="平台摘要" value={article.summary || '无'} multiline />
            {hasAuthorizedBody && <DetailField label="授权正文" value={article.body_content || ''} multiline large />}
            {isSogouRedirect && <div className="admin-article-detail-note subtle">该地址来自搜狗跳转，可能受搜狗反爬或链接过期影响。</div>}
          </section>

          <section className="admin-article-detail-card">
            <div className="admin-article-detail-card-head">
              <h2>AI 分析</h2>
              <button onClick={generateAI} disabled={actionSaving}>
                <Bot size={16} />
                生成 AI 分析
              </button>
            </div>
            {article.ai_error_msg && <div className="admin-article-detail-note error">{article.ai_error_msg}</div>}
            {article.ai_analysis ? (
              <pre className="admin-article-ai-json">{JSON.stringify(article.ai_analysis, null, 2)}</pre>
            ) : (
              <div className="admin-article-detail-empty compact">尚未生成 AI 分析。</div>
            )}
          </section>
        </div>

        <aside className="admin-article-detail-side">
          <section className="admin-article-detail-card side">
            <h2>审核操作</h2>
            <div className="admin-article-action-grid">
              <button className="success" onClick={() => runAction('publish')} disabled={actionSaving}>
                <CheckCircle2 size={16} />
                发布
              </button>
              <button onClick={() => runAction('reject')} disabled={actionSaving}>
                <XCircle size={16} />
                拒绝
              </button>
              <button onClick={() => runAction('take_down')} disabled={actionSaving}>
                <Archive size={16} />
                下架
              </button>
              <button className="danger" onClick={() => runAction('delete')} disabled={actionSaving}>
                <Trash2 size={16} />
                删除
              </button>
            </div>
            <button onClick={generateAI} disabled={actionSaving}>
              <Bot size={16} />
              生成 AI 分析
            </button>
          </section>

          <section className="admin-article-detail-card side">
            <h2>正文补全</h2>
            <label className="admin-article-body-field">
              <span>粘贴正文</span>
              <textarea
                value={bodyDraft}
                onChange={event => setBodyDraft(event.target.value)}
                placeholder="从原文页面复制完整正文后粘贴到这里"
                rows={7}
              />
            </label>
            <button onClick={saveBody} disabled={bodySaving}>
              <Save size={16} />
              保存正文
            </button>
            <label className="admin-article-body-field">
              <span>真实原文链接</span>
              <input
                value={fetchUrl}
                onChange={event => setFetchUrl(event.target.value)}
                placeholder="https://mp.weixin.qq.com/s?..."
              />
            </label>
            <button onClick={fetchBody} disabled={bodySaving}>
              <LinkIcon size={16} />
              从链接补抓
            </button>
          </section>

          <section className="admin-article-detail-card side">
            <h2>质量评估</h2>
            <div className={`admin-article-detail-quality-score ${qualityLevel(article.quality_score || 0)}`}>
              <strong>{article.quality_score || 0}</strong>
              <span>采集质量分</span>
            </div>
            <div className="admin-article-detail-quality-reasons">
              {(article.quality_reasons || []).length === 0 && <span>暂无评分原因。</span>}
              {(article.quality_reasons || []).map((reason, index) => (
                <span key={`${reason.type}-${index}`}>
                  {reason.message} {reason.points >= 0 ? '+' : ''}{reason.points}
                </span>
              ))}
            </div>
          </section>

          <section className="admin-article-detail-card side">
            <h2>元数据</h2>
            {article.cover_url && (
              <img
                className="admin-article-detail-cover"
                src={article.cover_url}
                alt=""
                onError={event => {
                  event.currentTarget.style.display = 'none'
                }}
              />
            )}
            <DetailField label="来源" value={article.source_name || '未知来源'} />
            <DetailField label="浏览/原文点击" value={`${article.view_count} / ${article.original_click_count || 0}`} />
            <DetailField label="入库时间" value={formatDateTime(article.created_at)} />
            <DetailField label="更新时间" value={formatDateTime(article.updated_at)} />
            <DetailField label="正文获取状态" value={bodyFetchStatusLabel(article.body_fetch_status)} />
            {article.body_fetch_error && <DetailField label="正文失败原因" value={article.body_fetch_error} multiline />}
            <DetailField label="文章 ID" value={article.id} />
          </section>
        </aside>
      </main>
    </div>
  )
}

function DetailField({ label, value, multiline = false, large = false }: { label: string; value: string; multiline?: boolean; large?: boolean }) {
  return (
    <div className={`${multiline ? 'admin-article-detail-field multiline' : 'admin-article-detail-field'}${large ? ' large' : ''}`}>
      <span>{label}</span>
      <p>{value}</p>
    </div>
  )
}

function formatDateTime(value?: string) {
  if (!value) return '未知'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  return date.toLocaleString('zh-CN', { hour12: false })
}

function statusLabel(value: string) {
  const labels: Record<string, string> = {
    candidate: '候选',
    published: '已发布',
    rejected: '已拒绝',
    taken_down: '已下架',
    deleted: '已删除',
  }
  return labels[value] || value
}

function aiStatusLabel(value?: string) {
  const labels: Record<string, string> = {
    pending: '待生成',
    succeeded: '已完成',
    failed: '失败',
  }
  return labels[value || 'pending'] || value || '待生成'
}

function bodyFetchStatusLabel(value?: string) {
  const labels: Record<string, string> = {
    pending: '待获取',
    succeeded: '已获取',
    manual: '手动保存',
    sogou_verify_required: '搜狗验证',
    sogou_redirect_missing: '搜狗跳转缺失',
    wechat_no_js_content: '微信页面无正文节点',
    wechat_video_page: '视频或富媒体内容',
    wechat_antispider: '微信反爬拦截',
    timeout: '请求超时',
    http_error: 'HTTP 状态异常',
    failed: '获取失败',
  }
  return labels[value || 'pending'] || value || '待获取'
}

function qualityLevel(score: number) {
  if (score >= 80) return 'high'
  if (score >= 60) return 'mid'
  return 'low'
}
