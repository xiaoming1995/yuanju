import { useEffect, useMemo, useState } from 'react'
import { Link, Navigate } from 'react-router-dom'
import { Search, ExternalLink } from 'lucide-react'
import ArticleCover from '../components/ArticleCover'
import { articleAPI, isArticleModuleClosedError } from '../lib/api'
import type { ArticleCategory, ArticleItem, ArticleTag } from '../lib/api'
import { useAuth } from '../contexts/AuthContext'
import './ArticlesPage.css'

function formatArticleDate(value?: string) {
  if (!value) return ''
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return ''
  return date.toLocaleDateString('zh-CN', { year: 'numeric', month: '2-digit', day: '2-digit' })
}

export default function ArticlesPage() {
  const { user, isLoading } = useAuth()
  const [articles, setArticles] = useState<ArticleItem[]>([])
  const [categories, setCategories] = useState<ArticleCategory[]>([])
  const [tags, setTags] = useState<ArticleTag[]>([])
  const [category, setCategory] = useState('')
  const [tag, setTag] = useState('')
  const [query, setQuery] = useState('')
  const [sort, setSort] = useState<'latest' | 'hot'>('latest')
  const [page, setPage] = useState(1)
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(false)
  const [moduleClosed, setModuleClosed] = useState(false)

  const params = useMemo(() => ({ category, tag, q: query, sort, page, limit: 20 }), [category, tag, page, query, sort])

  useEffect(() => {
    if (!user) return
    queueMicrotask(() => setModuleClosed(false))
    queueMicrotask(() => setLoading(true))
    articleAPI.list(params)
      .then(res => {
        setArticles(res.data.articles || [])
        setCategories(res.data.categories || [])
        setTags(res.data.tags || [])
        setTotal(res.data.total || 0)
      })
      .catch(err => {
        if (isArticleModuleClosedError(err)) {
          setModuleClosed(true)
          setArticles([])
          setTotal(0)
        }
      })
      .finally(() => setLoading(false))
  }, [params, user])

  const resetPage = (fn: () => void) => {
    setPage(1)
    fn()
  }

  if (isLoading) return null
  if (!user) return <Navigate to="/login" replace />

  const totalPages = Math.max(1, Math.ceil(total / 20))

  if (moduleClosed) {
    return (
      <div className="page articles-page">
        <div className="container articles-shell">
          <div className="articles-empty">
            资讯模块暂未开放
          </div>
        </div>
      </div>
    )
  }

  return (
    <div className="page articles-page">
      <div className="container articles-shell">
        <header className="articles-header">
          <div>
            <span className="articles-eyebrow">内容灵感库</span>
            <h1>资讯</h1>
            <p>参考公开搜索信息与 AI 拆解，收集八字文章选题和仿写灵感。</p>
          </div>
          <div className="articles-header-aside">
            <span>{total} 篇</span>
            <div className="article-sort" aria-label="资讯排序">
              <button className={sort === 'latest' ? 'active' : ''} onClick={() => resetPage(() => setSort('latest'))}>最新</button>
              <button className={sort === 'hot' ? 'active' : ''} onClick={() => resetPage(() => setSort('hot'))}>热门</button>
            </div>
          </div>
        </header>

        <div className="articles-toolbar">
          <div className="articles-toolbar-title">
            <span>筛选</span>
          </div>
          <div className="articles-toolbar-controls">
            <label className="article-search">
              <Search size={16} />
              <input value={query} onChange={e => resetPage(() => setQuery(e.target.value))} placeholder="搜索标题、来源或摘要" />
            </label>
            <select value={category} onChange={e => resetPage(() => setCategory(e.target.value))}>
              <option value="">全部分类</option>
              {categories.map(item => <option key={item.id} value={item.id}>{item.name}</option>)}
            </select>
            <select value={tag} onChange={e => resetPage(() => setTag(e.target.value))}>
              <option value="">全部标签</option>
              {tags.map(item => <option key={item.id} value={item.id}>{item.name}</option>)}
            </select>
          </div>
        </div>

        {loading ? (
          <div className="articles-empty">加载中...</div>
        ) : articles.length === 0 ? (
          <div className="articles-empty">暂无可读资讯</div>
        ) : (
          <>
            <div className="articles-list">
              {articles.map(article => (
                <Link to={`/articles/${article.id}`} className="article-row" key={article.id}>
                  <ArticleCover
                    className="article-row-cover"
                    coverUrl={article.cover_url}
                    sourceName={article.source_name}
                  />
                  <div className="article-row-body">
                    <div className="article-row-meta">
                      <span>{article.source_name || '未知来源'}</span>
                      {formatArticleDate(article.published_at_source) && <span>{formatArticleDate(article.published_at_source)}</span>}
                      <span>浏览 {article.view_count}</span>
                    </div>
                    <h2>{article.title}</h2>
                    <p>{article.summary || article.search_snippet || '暂无摘要'}</p>
                    <div className="article-row-footer">
                      <div className="article-tags">
                        {article.category && <span className="article-tag-category">{article.category.name}</span>}
                        {article.tags?.map(item => <span key={item.id}>{item.name}</span>)}
                      </div>
                      <span className="article-row-action">
                        <ExternalLink size={16} />
                      </span>
                    </div>
                  </div>
                </Link>
              ))}
            </div>
            <div className="article-pagination">
              <button disabled={page <= 1} onClick={() => setPage(v => Math.max(1, v - 1))}>上一页</button>
              <span>第 {page} 页 / 共 {totalPages} 页</span>
              <button disabled={page >= totalPages} onClick={() => setPage(v => v + 1)}>下一页</button>
            </div>
          </>
        )}
      </div>
    </div>
  )
}
