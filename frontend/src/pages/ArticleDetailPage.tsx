import { useEffect, useState } from 'react'
import { Link, Navigate, useParams } from 'react-router-dom'
import { ExternalLink } from 'lucide-react'
import ArticleCover from '../components/ArticleCover'
import { articleAPI, isArticleModuleClosedError } from '../lib/api'
import type { ArticleItem } from '../lib/api'
import { useAuth } from '../contexts/AuthContext'
import './ArticlesPage.css'

function formatArticleDate(value?: string) {
  if (!value) return ''
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return ''
  return date.toLocaleDateString('zh-CN', { year: 'numeric', month: '2-digit', day: '2-digit' })
}

function ListSection({ title, items }: { title: string; items?: string[] }) {
  if (!items || items.length === 0) return null
  return (
    <section className="article-detail-section">
      <h2>{title}</h2>
      <ul>
        {items.map((item, index) => <li key={`${title}-${index}`}>{item}</li>)}
      </ul>
    </section>
  )
}

export default function ArticleDetailPage() {
  const { id = '' } = useParams()
  const { user, isLoading } = useAuth()
  const [article, setArticle] = useState<ArticleItem | null>(null)
  const [moduleClosed, setModuleClosed] = useState(false)

  useEffect(() => {
    if (!user || !id) return
    queueMicrotask(() => setModuleClosed(false))
    articleAPI.detail(id)
      .then(res => setArticle(res.data.article))
      .catch(err => {
        if (isArticleModuleClosedError(err)) {
          setModuleClosed(true)
        }
      })
  }, [id, user])

  if (isLoading) return null
  if (!user) return <Navigate to="/login" replace />

  if (moduleClosed) {
    return (
      <div className="page articles-page">
        <div className="container article-detail-shell">
          <Link to="/" className="article-back">返回首页</Link>
          <div className="articles-empty">资讯模块暂未开放</div>
        </div>
      </div>
    )
  }

  const analysis = article?.ai_analysis
  const hasAnalysis = !!analysis && Object.values(analysis).some(value => Array.isArray(value) ? value.length > 0 : !!value)
  const hasBody = Boolean(article?.full_text_authorized && article.body_content?.trim())

  const openOriginal = async () => {
    if (!article) return
    const res = await articleAPI.trackOriginalClick(article.id)
    window.open(res.data.original_url || article.original_url, '_blank', 'noopener,noreferrer')
  }

  return (
    <div className="page articles-page">
      <div className="container article-detail-shell">
        <Link to="/articles" className="article-back">返回资讯</Link>
        {!article ? (
          <div className="articles-empty">加载中...</div>
        ) : (
          <>
            <header className="article-detail-header">
              <div className="article-detail-hero-copy">
                <div className="article-row-meta">
                  <span>{article.source_name || '未知来源'}</span>
                  {formatArticleDate(article.published_at_source) && <span>{formatArticleDate(article.published_at_source)}</span>}
                  <span>浏览 {article.view_count}</span>
                </div>
                <h1>{article.title}</h1>
                <p>{article.summary || article.search_snippet || '暂无摘要'}</p>
                <div className="article-tags">
                  {article.category && <span className="article-tag-category">{article.category.name}</span>}
                  {article.tags?.map(item => <span key={item.id}>{item.name}</span>)}
                </div>
              </div>
              <div className="article-detail-hero-action">
                <ArticleCover
                  className="article-detail-cover"
                  coverUrl={article.cover_url}
                  sourceName={article.source_name}
                  loading="eager"
                />
                <button className="btn btn-primary" onClick={openOriginal}>
                  <ExternalLink size={16} />
                  阅读原文
                </button>
              </div>
            </header>

            {hasBody && (
              <section className="article-detail-section article-body-section">
                <h2>正文内容</h2>
                <p>{article.body_content}</p>
              </section>
            )}

            {hasAnalysis && (
              <div className="article-ai-notice">
                AI 摘要与仿写拆解基于{hasBody ? '已采集正文和公开信息' : '公开搜索信息'}生成，仅供选题和写作参考。
              </div>
            )}

            {hasAnalysis && (
              <div className="article-detail-grid">
                <div className="article-detail-group">
                  <div className="article-detail-group-title">阅读辅助</div>
                  <div className="article-detail-group-grid">
                    {analysis?.one_sentence_summary && (
                      <section className="article-detail-section article-detail-section-wide">
                        <h2>一句话摘要</h2>
                        <p>{analysis.one_sentence_summary}</p>
                      </section>
                    )}
                    <ListSection title="阅读要点" items={analysis?.key_points} />
                    <ListSection title="适合读者" items={analysis?.target_readers} />
                    <ListSection title="相关主题" items={analysis?.related_topics} />
                  </div>
                </div>

                <div className="article-detail-group">
                  <div className="article-detail-group-title">写作拆解</div>
                  <div className="article-detail-group-grid">
                    {analysis?.title_pattern && (
                      <section className="article-detail-section">
                        <h2>标题模式</h2>
                        <p>{analysis.title_pattern}</p>
                      </section>
                    )}
                    {analysis?.opening_style && (
                      <section className="article-detail-section">
                        <h2>开头方式</h2>
                        <p>{analysis.opening_style}</p>
                      </section>
                    )}
                    <ListSection title="结构拆解" items={analysis?.structure_outline} />
                    <ListSection title="表达风格" items={analysis?.expression_style} />
                    <ListSection title="仿写角度" items={analysis?.rewrite_angles} />
                  </div>
                </div>
              </div>
            )}
          </>
        )}
      </div>
    </div>
  )
}
