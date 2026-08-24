import { useCallback, useEffect, useMemo, useState } from 'react'
import type { ReactNode } from 'react'
import { Link } from 'react-router-dom'
import {
  Archive,
  Bot,
  CheckCircle2,
  Filter,
  Play,
  RefreshCw,
  RotateCcw,
  Search,
  Settings,
  Sparkles,
  Tags,
  Trash2,
  XCircle,
} from 'lucide-react'
import { adminArticlesAPI } from '../../lib/adminApi'
import './AdminArticlesPage.css'

type TabKey = 'articles' | 'collection' | 'content' | 'tasks'

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
}

interface QualityReason {
  type: string
  points: number
  message: string
}

interface TaxonomyRecord {
  id: string
  name: string
  slug?: string
  sort_order?: number
  active: boolean
}

interface KeywordRecord {
  id: string
  keyword: string
  active: boolean
}

interface CollectionConfigState {
  enabled: boolean
  frequency: string
  schedule_interval_minutes: number
  max_results_per_run: number
  search_page_min: number
  search_page_max: number
  auto_publish_enabled: boolean
  auto_publish_min_quality_score: number
  auto_publish_requires_body: boolean
  auto_publish_max_per_run: number
}

interface QualityConfigState {
  quality_filter_enabled: boolean
  min_quality_score: number
  allow_without_body: boolean
  bonus_keywords: string[]
  source_blacklist: string[]
  preferred_sources: string[]
  ai_quality_check_enabled: boolean
}

interface TaskRecord {
  id: string
  status: string
  keyword_count: number
  found_count?: number
  inserted_count: number
  duplicate_count: number
  failed_count: number
  error_msg?: string
  started_at?: string
  finished_at?: string
}

interface TaskItemRecord {
  id: string
  task_id: string
  article_id?: string
  original_url?: string
  keyword?: string
  status: string
  error_msg?: string
  article_title?: string
  source_name?: string
  cover_url?: string
  body_fetch_status?: string
  ai_status?: string
  quality_score?: number
  quality_reasons?: QualityReason[]
  skip_reason?: string
  auto_published?: boolean
  auto_publish_reason?: string
  created_at?: string
}

const tabs: Array<{ key: TabKey; label: string; desc: string }> = [
  { key: 'articles', label: '审核队列', desc: '处理候选、发布和下架' },
  { key: 'collection', label: '采集中心', desc: '关键词、手动采集和定时' },
  { key: 'content', label: '内容配置', desc: '分类、标签和 AI 配置' },
  { key: 'tasks', label: '任务日志', desc: '查看采集历史和失败原因' },
]

const defaultCollectionConfig: CollectionConfigState = {
  enabled: false,
  frequency: 'daily',
  schedule_interval_minutes: 1440,
  max_results_per_run: 20,
  search_page_min: 1,
  search_page_max: 5,
  auto_publish_enabled: false,
  auto_publish_min_quality_score: 90,
  auto_publish_requires_body: true,
  auto_publish_max_per_run: 3,
}

const defaultQualityConfig: QualityConfigState = {
  quality_filter_enabled: false,
  min_quality_score: 60,
  allow_without_body: true,
  bonus_keywords: [],
  source_blacklist: [],
  preferred_sources: [],
  ai_quality_check_enabled: false,
}

const collectionIntervalPresets = [
  { label: '10 分钟', value: 10 },
  { label: '30 分钟', value: 30 },
  { label: '1 小时', value: 60 },
  { label: '6 小时', value: 360 },
  { label: '每天', value: 1440 },
]

export default function AdminArticlesPage() {
  const [tab, setTab] = useState<TabKey>('articles')
  const [articles, setArticles] = useState<AdminArticleRecord[]>([])
  const [categories, setCategories] = useState<TaxonomyRecord[]>([])
  const [tags, setTags] = useState<TaxonomyRecord[]>([])
  const [keywords, setKeywords] = useState<KeywordRecord[]>([])
  const [tasks, setTasks] = useState<TaskRecord[]>([])
  const [expandedTaskID, setExpandedTaskID] = useState('')
  const [taskItems, setTaskItems] = useState<Record<string, TaskItemRecord[]>>({})
  const [taskItemsLoading, setTaskItemsLoading] = useState<Record<string, boolean>>({})
  const [status, setStatus] = useState('candidate')
  const [articleQuery, setArticleQuery] = useState('')
  const [articleCategory, setArticleCategory] = useState('')
  const [articleTag, setArticleTag] = useState('')
  const [articleSort, setArticleSort] = useState<'latest' | 'hot' | 'quality'>('latest')
  const [minQualityScore, setMinQualityScore] = useState(0)
  const [selected, setSelected] = useState<string[]>([])
  const [articleTotal, setArticleTotal] = useState(0)
  const [articlePage, setArticlePage] = useState(1)
  const [articleLimit, setArticleLimit] = useState(20)
  const [message, setMessage] = useState('')
  const [collecting, setCollecting] = useState(false)
  const [retryingTaskID, setRetryingTaskID] = useState('')
  const [name, setName] = useState('')
  const [tagName, setTagName] = useState('')
  const [keyword, setKeyword] = useState('')
  const [config, setConfig] = useState<CollectionConfigState>(defaultCollectionConfig)
  const [savedConfig, setSavedConfig] = useState<CollectionConfigState>(defaultCollectionConfig)
  const [collectionConfigSaving, setCollectionConfigSaving] = useState(false)
  const [collectionConfigFeedback, setCollectionConfigFeedback] = useState<{ type: 'success' | 'error'; title: string; detail: string } | null>(null)
  const [moduleEnabled, setModuleEnabled] = useState(true)
  const [savedModuleEnabled, setSavedModuleEnabled] = useState(true)
  const [moduleSettingsSaving, setModuleSettingsSaving] = useState(false)
  const [moduleSettingsFeedback, setModuleSettingsFeedback] = useState('')
  const [qualityConfig, setQualityConfig] = useState<QualityConfigState>(defaultQualityConfig)
  const [savedQualityConfig, setSavedQualityConfig] = useState<QualityConfigState>(defaultQualityConfig)
  const [qualityConfigSaving, setQualityConfigSaving] = useState(false)
  const [qualityConfigFeedback, setQualityConfigFeedback] = useState<{ type: 'success' | 'error'; title: string; detail: string } | null>(null)
  const [aiConfig, setAIConfig] = useState({ prompt_content: '', provider_name: '', provider_type: 'openai', base_url: '', model: '', api_key: '' })

  const loadArticles = useCallback(() => {
    adminArticlesAPI.list({ status, q: articleQuery, category: articleCategory, tag: articleTag, sort: articleSort, min_quality_score: minQualityScore || undefined, page: articlePage, limit: articleLimit }).then(res => {
      setArticles(res.data.articles || [])
      setArticleTotal(res.data.total || 0)
    })
  }, [articleCategory, articleLimit, articlePage, articleQuery, articleSort, articleTag, minQualityScore, status])
  const loadTaxonomy = useCallback(() => {
    adminArticlesAPI.categories().then(res => setCategories(res.data.categories || []))
    adminArticlesAPI.tags().then(res => setTags(res.data.tags || []))
  }, [])
  const loadKeywords = useCallback(() => adminArticlesAPI.keywords().then(res => setKeywords(res.data.keywords || [])), [])
  const loadTasks = useCallback(() => adminArticlesAPI.tasks().then(res => setTasks(res.data.tasks || [])), [])

  useEffect(() => {
    loadArticles()
  }, [loadArticles])

  useEffect(() => {
    loadTaxonomy()
    loadKeywords()
    loadTasks()
    adminArticlesAPI.getModuleSettings().then(res => {
      const enabled = Boolean(res.data.module_enabled)
      setModuleEnabled(enabled)
      setSavedModuleEnabled(enabled)
    }).catch(err => {
      const error = err as { response?: { data?: { error?: string } }; message?: string }
      setModuleSettingsFeedback(error.response?.data?.error || error.message || '资讯模块设置读取失败')
    })
    adminArticlesAPI.getCollectionConfig().then(res => {
      if (res.data.config) {
        const nextConfig = normalizeCollectionConfig({ ...defaultCollectionConfig, ...res.data.config })
        setConfig(nextConfig)
        setSavedConfig(nextConfig)
      }
    })
    adminArticlesAPI.getQualityConfig().then(res => {
      if (res.data.config) {
        const nextConfig = normalizeQualityConfig({ ...defaultQualityConfig, ...res.data.config })
        setQualityConfig(nextConfig)
        setSavedQualityConfig(nextConfig)
      }
    })
    adminArticlesAPI.getAIConfig().then(res => {
      if (res.data.prompt?.content) setAIConfig(v => ({ ...v, prompt_content: res.data.prompt.content }))
    })
  }, [loadKeywords, loadTasks, loadTaxonomy])

  const currentPageIDs = articles.map(item => item.id)
  const allCurrentPageSelected = currentPageIDs.length > 0 && currentPageIDs.every(id => selected.includes(id))
  const totalPages = Math.max(1, Math.ceil(articleTotal / articleLimit))
  const visibleFrom = articleTotal === 0 ? 0 : (articlePage - 1) * articleLimit + 1
  const visibleTo = Math.min(articleTotal, articlePage * articleLimit)
  const latestTask = tasks[0]
  const moduleSettingsDirty = moduleEnabled !== savedModuleEnabled
  const collectionConfigDirty = !sameCollectionConfig(config, savedConfig)
  const collectionConfigError = getCollectionConfigError(config)
  const currentCollectionSummary = collectionConfigSummary(savedConfig)
  const qualityConfigDirty = !sameQualityConfig(qualityConfig, savedQualityConfig)
  const qualityConfigError = getQualityConfigError(qualityConfig)
  const currentQualitySummary = qualityConfigSummary(savedQualityConfig)

  const pageStats = useMemo(() => {
    const aiReady = articles.filter(item => Boolean(item.ai_analysis)).length
    const bodyReady = articles.filter(item => item.full_text_authorized && item.body_content?.trim()).length
    const bodyFailed = articles.filter(item => {
      const value = item.body_fetch_status || 'pending'
      return !['pending', 'succeeded', 'manual'].includes(value)
    }).length
    return { aiReady, bodyReady, bodyFailed }
  }, [articles])

  const toggleSelected = (id: string) => {
    setSelected(prev => prev.includes(id) ? prev.filter(item => item !== id) : [...prev, id])
  }

  const toggleCurrentPageSelected = () => {
    setSelected(prev => {
      if (allCurrentPageSelected) {
        return prev.filter(id => !currentPageIDs.includes(id))
      }
      return Array.from(new Set([...prev, ...currentPageIDs]))
    })
  }

  const resetArticlePage = (fn: () => void) => {
    setArticlePage(1)
    setSelected([])
    fn()
  }

  const goArticlePage = (nextPage: number) => {
    setArticlePage(Math.min(totalPages, Math.max(1, nextPage)))
    setSelected([])
  }

  const runBatch = async (action: string) => {
    if (selected.length === 0) return
    if (action === 'delete' && !window.confirm(`确认删除选中的 ${selected.length} 篇文章？删除后会进入“已删除”状态。`)) {
      return
    }
    const hasMissingAI = action === 'publish' && articles.some(item => selected.includes(item.id) && !item.ai_analysis)
    const allow = hasMissingAI ? window.confirm('选中的文章包含未生成 AI 分析的内容，确认发布？') : true
    if (!allow) return
    const res = await adminArticlesAPI.batchAction({ ids: selected, action, allow_publish_without_ai: hasMissingAI })
    setMessage(`已更新 ${res.data.result.updated} 条`)
    setSelected([])
    loadArticles()
  }

  const runBatchAI = async () => {
    if (selected.length === 0) return
    setMessage('批量 AI 分析中...')
    const res = await adminArticlesAPI.batchGenerateAI(selected)
    setMessage(`AI 分析完成 ${res.data.result.succeeded} 条，失败 ${res.data.result.failed.length} 条`)
    loadArticles()
  }

  const createCategory = async () => {
    if (!name.trim()) return
    await adminArticlesAPI.createCategory({ name })
    setName('')
    setMessage('分类已创建')
    loadTaxonomy()
  }

  const createTag = async () => {
    if (!tagName.trim()) return
    await adminArticlesAPI.createTag({ name: tagName })
    setTagName('')
    setMessage('标签已创建')
    loadTaxonomy()
  }

  const createKeyword = async () => {
    if (!keyword.trim()) return
    await adminArticlesAPI.createKeyword({ keyword })
    setKeyword('')
    setMessage('关键词已创建')
    loadKeywords()
  }

  const editCategory = async (item: TaxonomyRecord) => {
    const nextName = window.prompt('分类名称', item.name)
    if (!nextName) return
    await adminArticlesAPI.updateCategory(item.id, { name: nextName, slug: item.slug, sort_order: item.sort_order || 0, active: item.active })
    setMessage('分类已更新')
    loadTaxonomy()
  }

  const toggleCategory = async (item: TaxonomyRecord) => {
    await adminArticlesAPI.updateCategory(item.id, { name: item.name, slug: item.slug, sort_order: item.sort_order || 0, active: !item.active })
    setMessage(item.active ? '分类已停用' : '分类已启用')
    loadTaxonomy()
  }

  const editTag = async (item: TaxonomyRecord) => {
    const nextName = window.prompt('标签名称', item.name)
    if (!nextName) return
    await adminArticlesAPI.updateTag(item.id, { name: nextName, slug: item.slug, active: item.active })
    setMessage('标签已更新')
    loadTaxonomy()
  }

  const toggleTag = async (item: TaxonomyRecord) => {
    await adminArticlesAPI.updateTag(item.id, { name: item.name, slug: item.slug, active: !item.active })
    setMessage(item.active ? '标签已停用' : '标签已启用')
    loadTaxonomy()
  }

  const editKeyword = async (item: KeywordRecord) => {
    const nextKeyword = window.prompt('采集关键词', item.keyword)
    if (!nextKeyword) return
    await adminArticlesAPI.updateKeyword(item.id, { keyword: nextKeyword, active: item.active })
    setMessage('关键词已更新')
    loadKeywords()
  }

  const toggleKeyword = async (item: KeywordRecord) => {
    await adminArticlesAPI.updateKeyword(item.id, { keyword: item.keyword, active: !item.active })
    setMessage(item.active ? '关键词已停用' : '关键词已启用')
    loadKeywords()
  }

  const triggerCollect = async () => {
    if (collecting) return
    if (!savedModuleEnabled) {
      setMessage('资讯模块已关闭，请先打开模块总开关。')
      return
    }
    setCollecting(true)
    setMessage('采集中，请稍候...')
    try {
      const res = await adminArticlesAPI.collect()
      const task = res.data.task
      setMessage(`采集完成：发现 ${task?.found_count || 0}，新增 ${task?.inserted_count || 0}，重复 ${task?.duplicate_count || 0}，失败 ${task?.failed_count || 0}`)
      loadTasks()
      loadArticles()
    } catch (err) {
      const error = err as { response?: { data?: { error?: string } }; message?: string }
      setMessage(`采集失败：${error.response?.data?.error || error.message || '未知错误'}`)
    } finally {
      setCollecting(false)
    }
  }

  const retryTask = async (id: string) => {
    if (retryingTaskID) return
    if (!savedModuleEnabled) {
      setMessage('资讯模块已关闭，请先打开模块总开关。')
      return
    }
    setRetryingTaskID(id)
    setMessage('任务重试中...')
    try {
      const res = await adminArticlesAPI.retryTask(id)
      setMessage(`重试完成：新增 ${res.data.task?.inserted_count || 0}，失败 ${res.data.task?.failed_count || 0}`)
      loadTasks()
      loadArticles()
    } catch (err) {
      const error = err as { response?: { data?: { error?: string } }; message?: string }
      setMessage(`重试失败：${error.response?.data?.error || error.message || '未知错误'}`)
    } finally {
      setRetryingTaskID('')
    }
  }

  const toggleTaskItems = async (taskID: string) => {
    if (expandedTaskID === taskID) {
      setExpandedTaskID('')
      return
    }
    setExpandedTaskID(taskID)
    if (taskItems[taskID]) return
    setTaskItemsLoading(prev => ({ ...prev, [taskID]: true }))
    try {
      const res = await adminArticlesAPI.taskItems(taskID)
      setTaskItems(prev => ({ ...prev, [taskID]: res.data.items || [] }))
    } catch (err) {
      const error = err as { response?: { data?: { error?: string } }; message?: string }
      setMessage(`采集明细加载失败：${error.response?.data?.error || error.message || '未知错误'}`)
    } finally {
      setTaskItemsLoading(prev => ({ ...prev, [taskID]: false }))
    }
  }

  const saveCollectionConfig = async () => {
    if (collectionConfigSaving || collectionConfigError) return
    setCollectionConfigSaving(true)
    setCollectionConfigFeedback(null)
    const payload = {
      ...config,
      frequency: config.schedule_interval_minutes >= 10080 ? 'weekly' : 'daily',
    }
    try {
      const res = await adminArticlesAPI.updateCollectionConfig(payload)
      const savedConfig = normalizeCollectionConfig(res.data.config ? { ...config, ...res.data.config } : config)
      if (res.data.config) {
        setConfig(savedConfig)
      }
      setSavedConfig(savedConfig)
      const detail = collectionConfigSummary(savedConfig)
      const savedAt = new Date().toLocaleString('zh-CN', { hour12: false })
      setCollectionConfigFeedback({ type: 'success', title: `已保存 ${savedAt}`, detail })
      setMessage(`采集配置已保存：${detail}`)
    } catch (err) {
      const error = err as { response?: { data?: { error?: string } }; message?: string }
      const reason = error.response?.data?.error || error.message || '未知错误'
      setCollectionConfigFeedback({ type: 'error', title: '保存失败', detail: reason })
      setMessage(`采集配置保存失败：${reason}`)
    } finally {
      setCollectionConfigSaving(false)
    }
  }

  const resetCollectionConfig = () => {
    setConfig(savedConfig)
    setCollectionConfigFeedback(null)
  }

  const saveModuleSettings = async () => {
    if (moduleSettingsSaving || !moduleSettingsDirty) return
    setModuleSettingsSaving(true)
    setModuleSettingsFeedback('')
    try {
      const res = await adminArticlesAPI.updateModuleSettings({ module_enabled: moduleEnabled })
      const enabled = Boolean(res.data.module_enabled)
      setModuleEnabled(enabled)
      setSavedModuleEnabled(enabled)
      const text = enabled ? '资讯模块已启动' : '资讯模块已停止'
      setModuleSettingsFeedback(text)
      setMessage(text)
    } catch (err) {
      const error = err as { response?: { data?: { error?: string } }; message?: string }
      const reason = error.response?.data?.error || error.message || '未知错误'
      setModuleSettingsFeedback(`保存失败：${reason}`)
      setMessage(`资讯模块设置保存失败：${reason}`)
    } finally {
      setModuleSettingsSaving(false)
    }
  }

  const resetModuleSettings = () => {
    setModuleEnabled(savedModuleEnabled)
    setModuleSettingsFeedback('')
  }

  const saveQualityConfig = async () => {
    if (qualityConfigSaving || qualityConfigError) return
    setQualityConfigSaving(true)
    setQualityConfigFeedback(null)
    try {
      const res = await adminArticlesAPI.updateQualityConfig(normalizeQualityConfig(qualityConfig))
      const saved = normalizeQualityConfig(res.data.config ? { ...qualityConfig, ...res.data.config } : qualityConfig)
      setQualityConfig(saved)
      setSavedQualityConfig(saved)
      setQualityConfigFeedback({ type: 'success', title: '质量规则已保存', detail: qualityConfigSummary(saved) })
      setMessage(`质量筛选已保存：${qualityConfigSummary(saved)}`)
      loadArticles()
    } catch (err) {
      const error = err as { response?: { data?: { error?: string } }; message?: string }
      const reason = error.response?.data?.error || error.message || '未知错误'
      setQualityConfigFeedback({ type: 'error', title: '保存失败', detail: reason })
      setMessage(`质量筛选保存失败：${reason}`)
    } finally {
      setQualityConfigSaving(false)
    }
  }

  const resetQualityConfig = () => {
    setQualityConfig(savedQualityConfig)
    setQualityConfigFeedback(null)
  }

  const saveAIConfig = async () => {
    await adminArticlesAPI.updateAIConfig(aiConfig)
    setMessage('AI 配置已保存')
  }

  return (
    <div className="admin-articles-page">
      <header className="admin-page-header">
        <div>
          <span className="admin-page-kicker">内容运营</span>
          <h1>资讯管理</h1>
          <p>采集、审核、分类并发布八字相关文章参考。</p>
        </div>
        <div className="admin-article-summary">
          <SummaryStat label="当前列表" value={articleTotal} />
          <SummaryStat label="已选择" value={selected.length} />
          <SummaryStat label="启用关键词" value={keywords.filter(item => item.active).length} />
          {message && <div className="admin-article-summary-message">{message}</div>}
        </div>
      </header>

      <section className="admin-article-module-card">
        <div>
          <span className={savedModuleEnabled ? 'admin-article-module-state on' : 'admin-article-module-state off'}>
            {savedModuleEnabled ? '模块运行中' : '模块已停止'}
          </span>
          <h2>资讯模块总开关</h2>
          <p>关闭后普通用户入口、资讯接口、手动采集、重试采集和定时采集都会停止；后台查看、审核和配置仍可使用。</p>
          {moduleSettingsFeedback && <small>{moduleSettingsFeedback}</small>}
        </div>
        <div className="admin-article-module-actions">
          <label className="admin-article-switch-line">
            <input type="checkbox" checked={moduleEnabled} onChange={e => setModuleEnabled(e.target.checked)} />
            <span aria-hidden="true" />
            <strong>{moduleEnabled ? '启动整个模块' : '停止整个模块'}</strong>
          </label>
          <button className="admin-article-btn" onClick={resetModuleSettings} disabled={!moduleSettingsDirty || moduleSettingsSaving}>
            还原
          </button>
          <button className="admin-article-btn primary" onClick={saveModuleSettings} disabled={!moduleSettingsDirty || moduleSettingsSaving}>
            {moduleSettingsSaving ? <RefreshCw size={16} className="admin-article-spin" /> : <CheckCircle2 size={16} />}
            {moduleSettingsSaving ? '保存中' : moduleSettingsDirty ? '保存开关' : '已生效'}
          </button>
        </div>
      </section>

      <div className="admin-article-tabs">
        {tabs.map(item => (
          <button key={item.key} className={tab === item.key ? 'active' : ''} onClick={() => setTab(item.key)}>
            <strong>{item.label}</strong>
            <span>{item.desc}</span>
          </button>
        ))}
      </div>

      {tab === 'articles' && (
        <section className="admin-article-panel">
          <SectionHeader
            icon={<CheckCircle2 size={18} />}
            title="审核队列"
            desc={`当前筛选显示 ${visibleFrom}-${visibleTo}，共 ${articleTotal} 条`}
            actions={<button className="admin-article-btn" onClick={loadArticles}><RefreshCw size={16} />刷新</button>}
          />
          <div className="admin-article-queue-stats">
            <SummaryStat label="本页正文已获取" value={pageStats.bodyReady} />
            <SummaryStat label="本页正文失败" value={pageStats.bodyFailed} />
            <SummaryStat label="本页 AI 已完成" value={pageStats.aiReady} />
          </div>
          <div className="admin-article-filterbar">
            <label>
              <span>状态</span>
              <select value={status} onChange={e => resetArticlePage(() => setStatus(e.target.value))}>
                <option value="candidate">候选</option>
                <option value="published">已发布</option>
                <option value="rejected">已拒绝</option>
                <option value="taken_down">已下架</option>
                <option value="deleted">已删除</option>
              </select>
            </label>
            <label className="admin-article-search">
              <span>搜索</span>
              <div>
                <Search size={16} />
                <input placeholder="标题、来源或摘要" value={articleQuery} onChange={e => resetArticlePage(() => setArticleQuery(e.target.value))} />
              </div>
            </label>
            <label>
              <span>分类</span>
              <select value={articleCategory} onChange={e => resetArticlePage(() => setArticleCategory(e.target.value))}>
                <option value="">全部分类</option>
                {categories.map(item => <option key={item.id} value={item.id}>{item.name}</option>)}
              </select>
            </label>
            <label>
              <span>标签</span>
              <select value={articleTag} onChange={e => resetArticlePage(() => setArticleTag(e.target.value))}>
                <option value="">全部标签</option>
                {tags.map(item => <option key={item.id} value={item.id}>{item.name}</option>)}
              </select>
            </label>
            <label>
              <span>排序</span>
              <select value={articleSort} onChange={e => resetArticlePage(() => setArticleSort(e.target.value as 'latest' | 'hot' | 'quality'))}>
                <option value="latest">最新入库</option>
                <option value="quality">质量优先</option>
                <option value="hot">浏览优先</option>
              </select>
            </label>
            <label>
              <span>最低质量分</span>
              <input type="number" min={0} max={100} value={minQualityScore} onChange={e => resetArticlePage(() => setMinQualityScore(Number(e.target.value) || 0))} />
            </label>
            <button className="admin-article-btn primary" onClick={loadArticles}>
              <Filter size={16} />
              筛选
            </button>
          </div>

          <div className="admin-article-bulkbar">
            <label className="admin-article-select-page">
              <input type="checkbox" checked={allCurrentPageSelected} disabled={articles.length === 0} onChange={toggleCurrentPageSelected} />
              <span>{selected.length > 0 ? `已选择 ${selected.length} 条` : '选择本页后执行批量操作'}</span>
            </label>
            <div>
              <button className="admin-article-btn success" disabled={selected.length === 0} onClick={() => runBatch('publish')}>
                <CheckCircle2 size={16} />
                发布
              </button>
              <button className="admin-article-btn" disabled={selected.length === 0} onClick={runBatchAI}>
                <Sparkles size={16} />
                AI 分析
              </button>
              <button className="admin-article-btn" disabled={selected.length === 0} onClick={() => runBatch('reject')}>
                <XCircle size={16} />
                拒绝
              </button>
              <button className="admin-article-btn" disabled={selected.length === 0} onClick={() => runBatch('take_down')}>
                <Archive size={16} />
                下架
              </button>
              <button className="admin-article-btn danger" disabled={selected.length === 0} onClick={() => runBatch('delete')}>
                <Trash2 size={16} />
                删除
              </button>
            </div>
          </div>

          <ArticleTable
            articles={articles}
            selected={selected}
            allCurrentPageSelected={allCurrentPageSelected}
            toggleSelected={toggleSelected}
            toggleCurrentPageSelected={toggleCurrentPageSelected}
            loadArticles={loadArticles}
          />
          <Pagination
            total={articleTotal}
            visibleFrom={visibleFrom}
            visibleTo={visibleTo}
            page={articlePage}
            totalPages={totalPages}
            limit={articleLimit}
            setLimit={value => resetArticlePage(() => setArticleLimit(value))}
            goPage={goArticlePage}
          />
        </section>
      )}

      {tab === 'collection' && (
        <section className="admin-article-panel padded">
          <SectionHeader
            icon={<Play size={18} />}
            title="采集中心"
            desc="维护关键词，触发采集，并查看最近一次采集结果。"
            actions={
              <button className="admin-article-btn primary" onClick={triggerCollect} disabled={!savedModuleEnabled || collecting || keywords.filter(item => item.active).length === 0}>
                {collecting ? <RefreshCw size={16} className="admin-article-spin" /> : <Play size={16} />}
                {!savedModuleEnabled ? '模块已停止' : collecting ? '采集中' : '手动采集'}
              </button>
            }
          />
          <div className="admin-article-workbench-grid">
            <div className="admin-article-card">
              <h2>最近采集</h2>
              {latestTask ? (
                <div className="admin-article-task-summary">
                  <span className={`admin-article-task-status ${latestTask.status}`}>{taskStatusLabel(latestTask.status)}</span>
                  <strong>{formatDateTime(latestTask.started_at)}</strong>
                  <div>
                    <SummaryStat label="找到" value={latestTask.found_count || 0} />
                    <SummaryStat label="新增" value={latestTask.inserted_count} />
                    <SummaryStat label="重复" value={latestTask.duplicate_count} />
                    <SummaryStat label="失败" value={latestTask.failed_count} />
                  </div>
                  {latestTask.error_msg && <p>{latestTask.error_msg}</p>}
                </div>
              ) : (
                <div className="admin-article-empty compact">暂无采集任务。</div>
              )}
            </div>
            <div className="admin-article-card">
              <div className="admin-article-config-head">
                <div>
                  <h2>定时配置</h2>
                  <p>{currentCollectionSummary}</p>
                </div>
                <span className={`admin-article-config-state ${collectionConfigDirty ? 'dirty' : 'saved'}`}>
                  {collectionConfigDirty ? '未保存' : '已生效'}
                </span>
              </div>
              <div className="admin-article-config-presets">
                {collectionIntervalPresets.map(item => (
                  <button
                    type="button"
                    key={item.value}
                    className={config.schedule_interval_minutes === item.value ? 'active' : ''}
                    onClick={() => setConfig({ ...config, schedule_interval_minutes: item.value })}
                  >
                    {item.label}
                  </button>
                ))}
              </div>
              <div className="admin-article-settings-grid">
                <label>
                  <span>采集间隔（分钟）</span>
                  <input type="number" min={1} max={10080} value={config.schedule_interval_minutes} onChange={e => setConfig({ ...config, schedule_interval_minutes: Number(e.target.value) })} />
                </label>
                <label>
                  <span>单次任务最大采集数</span>
                  <input type="number" min={1} max={100} value={config.max_results_per_run} onChange={e => setConfig({ ...config, max_results_per_run: Number(e.target.value) })} />
                </label>
                <label>
                  <span>起始页码</span>
                  <input type="number" min={1} max={20} value={config.search_page_min} onChange={e => setConfig({ ...config, search_page_min: Number(e.target.value) })} />
                </label>
                <label>
                  <span>结束页码</span>
                  <input type="number" min={1} max={20} value={config.search_page_max} onChange={e => setConfig({ ...config, search_page_max: Number(e.target.value) })} />
                </label>
                <label className="admin-article-switch-line">
                  <input type="checkbox" checked={config.enabled} onChange={e => setConfig({ ...config, enabled: e.target.checked })} />
                  <span aria-hidden="true" />
                  <strong>{config.enabled ? '定时采集已启用' : '定时采集已停用'}</strong>
                </label>
              </div>
              <div className="admin-article-auto-publish">
                <div className="admin-article-subhead">
                  <h3>自动发布</h3>
                  <p>{autoPublishSummary(savedConfig)}</p>
                </div>
                <div className="admin-article-settings-grid">
                  <label>
                    <span>自动发布最低质量分</span>
                    <input type="number" min={0} max={100} value={config.auto_publish_min_quality_score} onChange={e => setConfig({ ...config, auto_publish_min_quality_score: Number(e.target.value) })} />
                  </label>
                  <label>
                    <span>单次最多自动发布</span>
                    <input type="number" min={0} max={20} value={config.auto_publish_max_per_run} onChange={e => setConfig({ ...config, auto_publish_max_per_run: Number(e.target.value) })} />
                  </label>
                  <label className="admin-article-switch-line">
                    <input type="checkbox" checked={config.auto_publish_enabled} onChange={e => setConfig({ ...config, auto_publish_enabled: e.target.checked })} />
                    <span aria-hidden="true" />
                    <strong>{config.auto_publish_enabled ? '质量达标自动发布' : '自动发布已关闭'}</strong>
                  </label>
                  <label className="admin-article-switch-line">
                    <input type="checkbox" checked={config.auto_publish_requires_body} onChange={e => setConfig({ ...config, auto_publish_requires_body: e.target.checked })} />
                    <span aria-hidden="true" />
                    <strong>{config.auto_publish_requires_body ? '必须有正文才自动发布' : '允许无正文自动发布'}</strong>
                  </label>
                </div>
              </div>
              {collectionConfigError && (
                <div className="admin-article-config-feedback error">
                  <strong>配置需要调整</strong>
                  <span>{collectionConfigError}</span>
                </div>
              )}
              {collectionConfigFeedback && (
                <div className={`admin-article-config-feedback ${collectionConfigFeedback.type}`}>
                  <strong>{collectionConfigFeedback.title}</strong>
                  <span>{collectionConfigFeedback.detail}</span>
                </div>
              )}
              <div className="admin-article-config-actions">
                <button className="admin-article-btn" onClick={resetCollectionConfig} disabled={!collectionConfigDirty || collectionConfigSaving}>
                  <RotateCcw size={16} />
                  还原
                </button>
                <button className="admin-article-btn primary" onClick={saveCollectionConfig} disabled={collectionConfigSaving || !collectionConfigDirty || Boolean(collectionConfigError)}>
                  {collectionConfigSaving ? <RefreshCw size={16} className="admin-article-spin" /> : <CheckCircle2 size={16} />}
                  {collectionConfigSaving ? '保存中' : collectionConfigDirty ? '保存并生效' : '已是最新'}
                </button>
              </div>
            </div>
          </div>

          <div className="admin-article-card">
            <div className="admin-article-config-head">
              <div>
                <h2>质量筛选</h2>
                <p>{currentQualitySummary}</p>
              </div>
              <span className={`admin-article-config-state ${qualityConfigDirty ? 'dirty' : 'saved'}`}>
                {qualityConfigDirty ? '未保存' : '已生效'}
              </span>
            </div>
            <div className="admin-article-settings-grid">
              <label>
                <span>最低质量分</span>
                <input type="number" min={0} max={100} value={qualityConfig.min_quality_score} onChange={e => setQualityConfig({ ...qualityConfig, min_quality_score: Number(e.target.value) })} />
              </label>
              <label>
                <span>加分关键词</span>
                <input value={qualityConfig.bonus_keywords.join('，')} onChange={e => setQualityConfig({ ...qualityConfig, bonus_keywords: parseQualityList(e.target.value) })} placeholder="财运，桃花，格局" />
              </label>
              <label>
                <span>优先来源</span>
                <input value={qualityConfig.preferred_sources.join('，')} onChange={e => setQualityConfig({ ...qualityConfig, preferred_sources: parseQualityList(e.target.value) })} placeholder="常用公众号名称" />
              </label>
              <label>
                <span>黑名单来源</span>
                <input value={qualityConfig.source_blacklist.join('，')} onChange={e => setQualityConfig({ ...qualityConfig, source_blacklist: parseQualityList(e.target.value) })} placeholder="不希望入库的来源" />
              </label>
              <label className="admin-article-switch-line">
                <input type="checkbox" checked={qualityConfig.quality_filter_enabled} onChange={e => setQualityConfig({ ...qualityConfig, quality_filter_enabled: e.target.checked })} />
                <span aria-hidden="true" />
                <strong>{qualityConfig.quality_filter_enabled ? '采集时跳过低分内容' : '仅评分不跳过'}</strong>
              </label>
              <label className="admin-article-switch-line">
                <input type="checkbox" checked={qualityConfig.allow_without_body} onChange={e => setQualityConfig({ ...qualityConfig, allow_without_body: e.target.checked })} />
                <span aria-hidden="true" />
                <strong>{qualityConfig.allow_without_body ? '允许无正文先入库' : '必须获取正文才入库'}</strong>
              </label>
            </div>
            {qualityConfigError && (
              <div className="admin-article-config-feedback error">
                <strong>配置需要调整</strong>
                <span>{qualityConfigError}</span>
              </div>
            )}
            {qualityConfigFeedback && (
              <div className={`admin-article-config-feedback ${qualityConfigFeedback.type}`}>
                <strong>{qualityConfigFeedback.title}</strong>
                <span>{qualityConfigFeedback.detail}</span>
              </div>
            )}
            <div className="admin-article-config-actions">
              <button className="admin-article-btn" onClick={resetQualityConfig} disabled={!qualityConfigDirty || qualityConfigSaving}>
                <RotateCcw size={16} />
                还原
              </button>
              <button className="admin-article-btn primary" onClick={saveQualityConfig} disabled={qualityConfigSaving || !qualityConfigDirty || Boolean(qualityConfigError)}>
                {qualityConfigSaving ? <RefreshCw size={16} className="admin-article-spin" /> : <CheckCircle2 size={16} />}
                {qualityConfigSaving ? '保存中' : qualityConfigDirty ? '保存质量规则' : '已是最新'}
              </button>
            </div>
          </div>

          <div className="admin-article-card">
            <div className="admin-article-card-head">
              <div>
                <h2>采集关键词</h2>
                <p>当前启用 {keywords.filter(item => item.active).length} 个，停用 {keywords.filter(item => !item.active).length} 个。</p>
              </div>
              <div className="admin-article-inline-form">
                <input placeholder="新增关键词" value={keyword} onChange={e => setKeyword(e.target.value)} />
                <button className="admin-article-btn" onClick={createKeyword}>新增</button>
              </div>
            </div>
            <ItemList
              items={keywords}
              empty="暂无关键词。"
              renderPrimary={item => item.keyword}
              renderSecondary={item => item.active ? '启用' : '停用'}
              renderActions={item => (
                <>
                  <button onClick={() => editKeyword(item)}>编辑</button>
                  <button onClick={() => toggleKeyword(item)}>{item.active ? '停用' : '启用'}</button>
                </>
              )}
            />
          </div>
        </section>
      )}

      {tab === 'content' && (
        <section className="admin-article-panel padded">
          <SectionHeader icon={<Settings size={18} />} title="内容配置" desc="维护文章分类、标签和资讯专用 AI 设置。" />
          <div className="admin-article-workbench-grid">
            <TaxonomyPanel
              title="分类"
              items={categories}
              value={name}
              setValue={setName}
              onCreate={createCategory}
              onEdit={editCategory}
              onToggle={toggleCategory}
            />
            <TaxonomyPanel
              title="标签"
              items={tags}
              value={tagName}
              setValue={setTagName}
              onCreate={createTag}
              onEdit={editTag}
              onToggle={toggleTag}
            />
          </div>
          <div className="admin-article-card">
            <div className="admin-article-card-head">
              <div>
                <h2>资讯 AI 配置</h2>
                <p>用于文章摘要、拆解和仿写角度分析，不影响八字报告模型。</p>
              </div>
              <button className="admin-article-btn primary" onClick={saveAIConfig}>保存 AI 配置</button>
            </div>
            <label className="admin-article-field">
              <span>Prompt</span>
              <textarea value={aiConfig.prompt_content} onChange={e => setAIConfig({ ...aiConfig, prompt_content: e.target.value })} />
            </label>
            <div className="admin-article-form-grid">
              <input placeholder="Provider 名称" value={aiConfig.provider_name} onChange={e => setAIConfig({ ...aiConfig, provider_name: e.target.value })} />
              <input placeholder="Base URL" value={aiConfig.base_url} onChange={e => setAIConfig({ ...aiConfig, base_url: e.target.value })} />
              <input placeholder="Model" value={aiConfig.model} onChange={e => setAIConfig({ ...aiConfig, model: e.target.value })} />
              <input placeholder="API Key" type="password" value={aiConfig.api_key} onChange={e => setAIConfig({ ...aiConfig, api_key: e.target.value })} />
            </div>
          </div>
        </section>
      )}

      {tab === 'tasks' && (
        <section className="admin-article-panel padded">
          <SectionHeader
            icon={<RefreshCw size={18} />}
            title="任务日志"
            desc="采集结果、入库文章、重复数据和失败原因集中在这里。"
            actions={<button className="admin-article-btn" onClick={loadTasks}><RefreshCw size={16} />刷新</button>}
          />
          <TaskLogList
            tasks={tasks}
            expandedTaskID={expandedTaskID}
            taskItems={taskItems}
            taskItemsLoading={taskItemsLoading}
            retryingTaskID={retryingTaskID}
            moduleEnabled={savedModuleEnabled}
            onToggleItems={toggleTaskItems}
            onRetry={retryTask}
          />
        </section>
      )}
    </div>
  )
}

function ArticleTable({ articles, selected, allCurrentPageSelected, toggleSelected, toggleCurrentPageSelected, loadArticles }: {
  articles: AdminArticleRecord[]
  selected: string[]
  allCurrentPageSelected: boolean
  toggleSelected: (id: string) => void
  toggleCurrentPageSelected: () => void
  loadArticles: () => void
}) {
  return (
    <div className="admin-article-table">
      <div className="admin-article-table-head">
        <span>
          <input type="checkbox" checked={allCurrentPageSelected} disabled={articles.length === 0} onChange={toggleCurrentPageSelected} />
        </span>
        <span>文章</span>
        <span>状态</span>
        <span>质量</span>
        <span>数据</span>
        <span>操作</span>
      </div>
      {articles.map(article => (
        <div className="admin-article-item" key={article.id}>
          <div className="admin-article-row">
            <input type="checkbox" checked={selected.includes(article.id)} onChange={() => toggleSelected(article.id)} />
            <div className="admin-article-title-cell">
              <strong>{article.title}</strong>
              <span>
                {article.source_name || '未知来源'} · {formatDateTime(article.published_at_source)}
              </span>
            </div>
            <div className="admin-article-status-cell">
              <span className={`admin-article-status ${article.status}`}>{statusLabel(article.status)}</span>
              <span className={`admin-article-ai-status ${article.ai_status || 'pending'}`}>AI {aiStatusLabel(article.ai_status)}</span>
              <span
                className={`admin-article-body-fetch-status ${article.body_fetch_status || 'pending'}`}
                title={article.body_fetch_error || undefined}
              >
                正文 {bodyFetchStatusLabel(article.body_fetch_status, article.full_text_authorized, article.body_content)}
              </span>
            </div>
            <div className="admin-article-quality-cell">
              <QualityScoreBadge score={article.quality_score || 0} />
              {article.quality_reasons?.[0] && <span>{article.quality_reasons[0].message}</span>}
            </div>
            <div className="admin-article-metrics">
              <span>浏览 {article.view_count}</span>
              <span>原文 {article.original_click_count || 0}</span>
            </div>
            <div className="admin-article-row-actions">
              <Link to={`/admin/articles/${article.id}`} title="查看采集数据">查看</Link>
              <button onClick={() => adminArticlesAPI.generateAI(article.id).then(loadArticles)}>
                <Bot size={16} />
                AI
              </button>
            </div>
          </div>
        </div>
      ))}
      {articles.length === 0 && <div className="admin-article-empty">当前筛选条件下没有文章。</div>}
    </div>
  )
}

function Pagination({ total, visibleFrom, visibleTo, page, totalPages, limit, setLimit, goPage }: {
  total: number
  visibleFrom: number
  visibleTo: number
  page: number
  totalPages: number
  limit: number
  setLimit: (value: number) => void
  goPage: (page: number) => void
}) {
  return (
    <div className="admin-article-pagination">
      <div>
        <span>共 {total} 条</span>
        <span>显示 {visibleFrom}-{visibleTo}</span>
      </div>
      <label>
        每页
        <select value={limit} onChange={e => setLimit(Number(e.target.value))}>
          <option value={10}>10</option>
          <option value={20}>20</option>
          <option value={50}>50</option>
          <option value={100}>100</option>
        </select>
      </label>
      <div className="admin-article-page-actions">
        <button disabled={page <= 1} onClick={() => goPage(page - 1)}>上一页</button>
        <span>第 {page} / {totalPages} 页</span>
        <button disabled={page >= totalPages} onClick={() => goPage(page + 1)}>下一页</button>
      </div>
    </div>
  )
}

function SectionHeader({ icon, title, desc, actions }: { icon: ReactNode; title: string; desc: string; actions?: ReactNode }) {
  return (
    <div className="admin-article-section-head">
      <div>
        <span>{icon}</span>
        <div>
          <h2>{title}</h2>
          <p>{desc}</p>
        </div>
      </div>
      {actions && <div className="admin-article-section-actions">{actions}</div>}
    </div>
  )
}

function SummaryStat({ label, value }: { label: string; value: ReactNode }) {
  return (
    <div className="admin-article-summary-stat">
      <span>{label}</span>
      <strong>{value}</strong>
    </div>
  )
}

function QualityScoreBadge({ score }: { score: number }) {
  const level = score >= 80 ? 'high' : score >= 60 ? 'mid' : 'low'
  return <span className={`admin-article-quality-badge ${level}`}>质量 {score}</span>
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

function bodyFetchStatusLabel(value?: string, authorized?: boolean, body?: string) {
  if (authorized && body?.trim()) return '已获取'
  const labels: Record<string, string> = {
    pending: '待获取',
    succeeded: '已获取',
    manual: '手动保存',
    sogou_verify_required: '搜狗验证',
    sogou_redirect_missing: '跳转缺失',
    wechat_no_js_content: '无正文节点',
    wechat_video_page: '视频/富媒体',
    wechat_antispider: '微信拦截',
    timeout: '超时',
    http_error: 'HTTP异常',
    failed: '失败',
  }
  return labels[value || 'pending'] || value || '待获取'
}

function taskStatusLabel(value: string) {
  const labels: Record<string, string> = {
    pending: '待执行',
    running: '执行中',
    succeeded: '成功',
    failed: '失败',
  }
  return labels[value] || value
}

function collectionConfigSummary(config: CollectionConfigState) {
  const enabled = config.enabled ? '已启用' : '已停用'
  return `${enabled}，每 ${config.schedule_interval_minutes} 分钟采集一次，单次最多 ${config.max_results_per_run} 条，随机搜索第 ${config.search_page_min}-${config.search_page_max} 页，${autoPublishSummary(config)}`
}

function autoPublishSummary(config: CollectionConfigState) {
  if (!config.auto_publish_enabled) return '自动发布关闭'
  const body = config.auto_publish_requires_body ? '且有正文' : '不要求正文'
  return `质量分 >= ${config.auto_publish_min_quality_score} ${body}，每次最多 ${config.auto_publish_max_per_run} 篇`
}

function qualityConfigSummary(config: QualityConfigState) {
  const mode = config.quality_filter_enabled ? `低于 ${config.min_quality_score} 分跳过` : '只打分不跳过'
  const body = config.allow_without_body ? '允许无正文' : '必须有正文'
  return `${mode}，${body}`
}

function normalizeCollectionConfig(config: CollectionConfigState): CollectionConfigState {
  return {
    enabled: Boolean(config.enabled),
    frequency: config.frequency || 'daily',
    schedule_interval_minutes: Number(config.schedule_interval_minutes) || defaultCollectionConfig.schedule_interval_minutes,
    max_results_per_run: Number(config.max_results_per_run) || defaultCollectionConfig.max_results_per_run,
    search_page_min: Number(config.search_page_min) || defaultCollectionConfig.search_page_min,
    search_page_max: Number(config.search_page_max) || defaultCollectionConfig.search_page_max,
    auto_publish_enabled: Boolean(config.auto_publish_enabled),
    auto_publish_min_quality_score: Math.min(100, Math.max(0, Number(config.auto_publish_min_quality_score) || 0)),
    auto_publish_requires_body: config.auto_publish_requires_body !== false,
    auto_publish_max_per_run: Math.min(20, Math.max(0, Number(config.auto_publish_max_per_run) || 0)),
  }
}

function normalizeQualityConfig(config: QualityConfigState): QualityConfigState {
  return {
    quality_filter_enabled: Boolean(config.quality_filter_enabled),
    min_quality_score: Math.min(100, Math.max(0, Number(config.min_quality_score) || 0)),
    allow_without_body: Boolean(config.allow_without_body),
    bonus_keywords: normalizeQualityList(config.bonus_keywords),
    source_blacklist: normalizeQualityList(config.source_blacklist),
    preferred_sources: normalizeQualityList(config.preferred_sources),
    ai_quality_check_enabled: Boolean(config.ai_quality_check_enabled),
  }
}

function normalizeQualityList(values: string[]) {
  return Array.from(new Set(values.map(item => item.trim()).filter(Boolean)))
}

function parseQualityList(value: string) {
  return normalizeQualityList(value.split(/[,，\n]/))
}

function sameCollectionConfig(a: CollectionConfigState, b: CollectionConfigState) {
  const left = normalizeCollectionConfig(a)
  const right = normalizeCollectionConfig(b)
  return left.enabled === right.enabled &&
    left.schedule_interval_minutes === right.schedule_interval_minutes &&
    left.max_results_per_run === right.max_results_per_run &&
    left.search_page_min === right.search_page_min &&
    left.search_page_max === right.search_page_max &&
    left.auto_publish_enabled === right.auto_publish_enabled &&
    left.auto_publish_min_quality_score === right.auto_publish_min_quality_score &&
    left.auto_publish_requires_body === right.auto_publish_requires_body &&
    left.auto_publish_max_per_run === right.auto_publish_max_per_run
}

function sameQualityConfig(a: QualityConfigState, b: QualityConfigState) {
  return JSON.stringify(normalizeQualityConfig(a)) === JSON.stringify(normalizeQualityConfig(b))
}

function getCollectionConfigError(config: CollectionConfigState) {
  if (config.schedule_interval_minutes < 1 || config.schedule_interval_minutes > 10080) {
    return '采集间隔需要在 1 到 10080 分钟之间。'
  }
  if (config.max_results_per_run < 1 || config.max_results_per_run > 100) {
    return '单次任务最大采集数需要在 1 到 100 条之间。'
  }
  if (config.search_page_min < 1 || config.search_page_max < 1 || config.search_page_min > 20 || config.search_page_max > 20) {
    return '搜索页码需要在 1 到 20 页之间。'
  }
  if (config.search_page_min > config.search_page_max) {
    return '起始页码不能大于结束页码。'
  }
  if (config.auto_publish_min_quality_score < 0 || config.auto_publish_min_quality_score > 100) {
    return '自动发布最低质量分需要在 0 到 100 之间。'
  }
  if (config.auto_publish_max_per_run < 0 || config.auto_publish_max_per_run > 20) {
    return '单次最多自动发布需要在 0 到 20 篇之间。'
  }
  return ''
}

function getQualityConfigError(config: QualityConfigState) {
  if (config.min_quality_score < 0 || config.min_quality_score > 100) {
    return '最低质量分需要在 0 到 100 之间。'
  }
  return ''
}

function collectionItemStatusLabel(value: string) {
  const labels: Record<string, string> = {
    inserted: '新增',
    duplicate: '重复',
    failed: '失败',
    skipped: '跳过',
  }
  return labels[value] || value
}

function formatDateTime(value?: string) {
  if (!value) return '未知'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  return date.toLocaleString('zh-CN', { hour12: false })
}

function TaskLogList({ tasks, expandedTaskID, taskItems, taskItemsLoading, retryingTaskID, moduleEnabled, onToggleItems, onRetry }: {
  tasks: TaskRecord[]
  expandedTaskID: string
  taskItems: Record<string, TaskItemRecord[]>
  taskItemsLoading: Record<string, boolean>
  retryingTaskID: string
  moduleEnabled: boolean
  onToggleItems: (taskID: string) => void
  onRetry: (taskID: string) => void
}) {
  if (tasks.length === 0) {
    return <div className="admin-article-empty compact">暂无采集任务。</div>
  }
  return (
    <div className="admin-article-task-log">
      {tasks.map(task => {
        const expanded = expandedTaskID === task.id
        const items = taskItems[task.id] || []
        const retrying = retryingTaskID === task.id
        return (
          <div className="admin-article-task-row" key={task.id}>
            <div className="admin-article-task-row-main">
              <div>
                <span className={`admin-article-task-status ${task.status}`}>{taskStatusLabel(task.status)}</span>
                <strong>{formatDateTime(task.started_at)}</strong>
                <span>
                  关键词 {task.keyword_count} · 找到 {task.found_count || 0} · 新增 {task.inserted_count} · 重复 {task.duplicate_count} · 失败 {task.failed_count}
                </span>
                {task.error_msg && <small>{task.error_msg}</small>}
              </div>
              <div className="admin-article-row-actions">
                <button onClick={() => onToggleItems(task.id)}>{expanded ? '收起明细' : '查看明细'}</button>
                {task.failed_count > 0 && <button onClick={() => onRetry(task.id)} disabled={!moduleEnabled || Boolean(retryingTaskID)}>{!moduleEnabled ? '模块已停止' : retrying ? '重试中' : '重试'}</button>}
              </div>
            </div>
            {expanded && (
              <div className="admin-article-task-items">
                {taskItemsLoading[task.id] && <div className="admin-article-empty compact">明细加载中...</div>}
                {!taskItemsLoading[task.id] && items.length === 0 && <div className="admin-article-empty compact">暂无采集明细。</div>}
                {!taskItemsLoading[task.id] && items.map(item => (
                  <div className="admin-article-task-item" key={item.id}>
                    <div className="admin-article-task-item-head">
                      <span className={`admin-article-task-item-status ${item.status}`}>{collectionItemStatusLabel(item.status)}</span>
                      {item.auto_published && <span className="admin-article-auto-publish-badge">自动发布</span>}
                      <strong>{item.article_title || item.original_url || '未获取标题'}</strong>
                      <QualityScoreBadge score={item.quality_score || 0} />
                    </div>
                    <div className="admin-article-task-item-meta">
                      <span>关键词 {item.keyword || '-'}</span>
                      <span>{item.source_name || '未知来源'}</span>
                      <span>正文 {bodyFetchStatusLabel(item.body_fetch_status)}</span>
                      <span>AI {aiStatusLabel(item.ai_status)}</span>
                    </div>
                    {item.skip_reason && <p>跳过原因：{item.skip_reason}</p>}
                    {item.auto_publish_reason && <p>自动发布：{item.auto_publish_reason}</p>}
                    {item.quality_reasons && item.quality_reasons.length > 0 && (
                      <div className="admin-article-quality-reasons">
                        {item.quality_reasons.slice(0, 4).map((reason, index) => (
                          <span key={`${reason.type}-${index}`}>{reason.message} {reason.points >= 0 ? '+' : ''}{reason.points}</span>
                        ))}
                      </div>
                    )}
                    {item.error_msg && <p>{item.error_msg}</p>}
                    <div className="admin-article-task-item-actions">
                      {item.article_id && <Link to={`/admin/articles/${item.article_id}`}>查看文章</Link>}
                      {item.original_url && <a href={item.original_url} target="_blank" rel="noreferrer">打开原文</a>}
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>
        )
      })}
    </div>
  )
}

function TaxonomyPanel({ title, items, value, setValue, onCreate, onEdit, onToggle }: {
  title: string
  items: TaxonomyRecord[]
  value: string
  setValue: (v: string) => void
  onCreate: () => void
  onEdit: (item: TaxonomyRecord) => void
  onToggle: (item: TaxonomyRecord) => void
}) {
  return (
    <div className="admin-article-card">
      <div className="admin-article-card-head">
        <div>
          <h2><Tags size={16} /> {title}</h2>
          <p>启用 {items.filter(item => item.active).length} 个，停用 {items.filter(item => !item.active).length} 个。</p>
        </div>
        <div className="admin-article-inline-form">
          <input placeholder={`新增${title}`} value={value} onChange={e => setValue(e.target.value)} />
          <button className="admin-article-btn" onClick={onCreate}>新增</button>
        </div>
      </div>
      <ItemList
        items={items}
        empty={`暂无${title}。`}
        renderPrimary={item => item.name}
        renderSecondary={item => item.active ? '启用' : '停用'}
        renderActions={item => (
          <>
            <button onClick={() => onEdit(item)}>编辑</button>
            <button onClick={() => onToggle(item)}>{item.active ? '停用' : '启用'}</button>
          </>
        )}
      />
    </div>
  )
}

function ItemList<T extends { id: string }>({ items, empty, renderPrimary, renderSecondary, renderActions }: {
  items: T[]
  empty?: string
  renderPrimary: (item: T) => ReactNode
  renderSecondary?: (item: T) => ReactNode
  renderActions?: (item: T) => ReactNode
}) {
  if (items.length === 0) {
    return <div className="admin-article-empty compact">{empty || '暂无数据。'}</div>
  }
  return (
    <div className="admin-article-list">
      {items.map(item => (
        <div className="admin-article-list-row" key={item.id}>
          <div>
            <strong>{renderPrimary(item)}</strong>
            {renderSecondary && <span>{renderSecondary(item)}</span>}
          </div>
          {renderActions && <div className="admin-article-row-actions">{renderActions(item)}</div>}
        </div>
      ))}
    </div>
  )
}
