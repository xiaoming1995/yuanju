import { useEffect, useRef, useState } from 'react'
import { Link, Navigate, useNavigate } from 'react-router-dom'
import { ArrowLeft } from 'lucide-react'
import { useAuth } from '../contexts/AuthContext'
import { brandAPI } from '../lib/api'
import type { ExportBrand } from '../lib/api'
import BrandPreviewCard from '../components/BrandPreviewCard'
import LogoCropModal from '../components/LogoCropModal'
import './BrandSettingsPage.css'

const DEFAULT_BRAND: ExportBrand = {
  title: '',
  footer_text: '',
  logo_url: '',
  watermark_mode: 'none',
  watermark_text: '',
}

const MAX_TITLE = 20
const MAX_FOOTER = 40
const MAX_WATERMARK = 30

export default function BrandSettingsPage() {
  const { user, isLoading: authLoading } = useAuth()
  const navigate = useNavigate()
  const [serverState, setServerState] = useState<ExportBrand>(DEFAULT_BRAND)
  const [draft, setDraft] = useState<ExportBrand>(DEFAULT_BRAND)
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [uploading, setUploading] = useState(false)
  const [error, setError] = useState('')
  const [cropSourceUrl, setCropSourceUrl] = useState<string | null>(null)
  const fileInputRef = useRef<HTMLInputElement>(null)

  useEffect(() => {
    if (!user) return
    brandAPI.get()
      .then(r => {
        setServerState(r.data.data)
        setDraft(r.data.data)
      })
      .catch(err => setError(err.message || '加载品牌设置失败'))
      .finally(() => setLoading(false))
  }, [user])

  if (!authLoading && !user) return <Navigate to="/login" replace />
  if (authLoading || loading) {
    return <main className="brand-page container page"><div className="brand-loading">加载中...</div></main>
  }

  const dirty =
    draft.title !== serverState.title ||
    draft.footer_text !== serverState.footer_text ||
    draft.watermark_mode !== serverState.watermark_mode ||
    draft.watermark_text !== serverState.watermark_text

  async function onSave() {
    setSaving(true)
    setError('')
    try {
      const r = await brandAPI.update({
        title: draft.title,
        footer_text: draft.footer_text,
        watermark_mode: draft.watermark_mode,
        watermark_text: draft.watermark_text,
      })
      setServerState(r.data.data)
      setDraft(r.data.data)
    } catch (e) {
      setError(e instanceof Error ? e.message : '保存失败')
    } finally {
      setSaving(false)
    }
  }

  function onLogoChange(e: React.ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0]
    e.target.value = ''
    if (!file) return
    if (file.size > 2 * 1024 * 1024) {
      setError('logo 文件不能超过 2MB')
      return
    }
    if (!['image/png', 'image/jpeg', 'image/webp'].includes(file.type)) {
      setError('仅支持 PNG / JPG / WebP 格式')
      return
    }
    setError('')
    const reader = new FileReader()
    reader.onload = () => {
      const result = reader.result
      if (typeof result === 'string') setCropSourceUrl(result)
    }
    reader.onerror = () => setError('读取文件失败')
    reader.readAsDataURL(file)
  }

  async function handleCropConfirm(file: File) {
    setCropSourceUrl(null)
    setUploading(true)
    setError('')
    try {
      const r = await brandAPI.uploadLogo(file)
      const next = { ...serverState, logo_url: r.data.data.logo_url }
      setServerState(next)
      setDraft(d => ({ ...d, logo_url: r.data.data.logo_url }))
    } catch (e) {
      setError(e instanceof Error ? e.message : '上传失败')
    } finally {
      setUploading(false)
    }
  }

  async function onLogoDelete() {
    if (!serverState.logo_url) return
    if (!window.confirm('确定要删除当前 logo 吗？')) return
    try {
      await brandAPI.deleteLogo()
      const next = { ...serverState, logo_url: '' }
      setServerState(next)
      setDraft(d => ({ ...d, logo_url: '' }))
    } catch (e) {
      setError(e instanceof Error ? e.message : '删除失败')
    }
  }

  async function onReset() {
    if (!window.confirm('重置为默认设置？已上传的 logo 也会被删除。')) return
    try {
      await brandAPI.update({ title: '', footer_text: '', watermark_mode: 'none', watermark_text: '' })
      if (serverState.logo_url) await brandAPI.deleteLogo()
      const fresh = await brandAPI.get()
      setServerState(fresh.data.data)
      setDraft(fresh.data.data)
    } catch (e) {
      setError(e instanceof Error ? e.message : '重置失败')
    }
  }

  return (
    <main className="brand-page container page">
      <header className="brand-page-header">
        <button className="brand-back" onClick={() => navigate(-1)}>
          <ArrowLeft size={16} /> 返回
        </button>
        <h1>导出品牌设置</h1>
      </header>

      {error && <div className="brand-error">{error}</div>}
      {dirty && <div className="brand-unsaved">有未保存的修改</div>}

      <section className="brand-section">
        <h2>顶部品牌</h2>
        <label className="brand-field">
          <span>品牌标题</span>
          <input
            type="text"
            maxLength={MAX_TITLE}
            value={draft.title}
            onChange={e => setDraft(d => ({ ...d, title: e.target.value }))}
            placeholder='留空则使用默认 "缘聚 命 理"'
          />
          <small>{draft.title.length} / {MAX_TITLE}</small>
        </label>

        <div className="brand-logo-row">
          <div className="brand-logo-preview">
            {serverState.logo_url ? (
              <img src={serverState.logo_url} alt="当前 logo" />
            ) : (
              <span className="brand-logo-empty">未上传</span>
            )}
          </div>
          <div className="brand-logo-actions">
            <button
              type="button"
              onClick={() => fileInputRef.current?.click()}
              disabled={uploading}
            >
              {uploading ? '上传中...' : (serverState.logo_url ? '更换' : '上传')}
            </button>
            {serverState.logo_url && (
              <button type="button" onClick={onLogoDelete} disabled={uploading}>删除</button>
            )}
            <small>PNG / JPG / WebP，≤ 2MB</small>
          </div>
          <input
            ref={fileInputRef}
            type="file"
            accept="image/png,image/jpeg,image/webp"
            style={{ display: 'none' }}
            onChange={onLogoChange}
          />
        </div>
      </section>

      <section className="brand-section">
        <h2>底部品牌</h2>
        <label className="brand-field">
          <span>底部文字</span>
          <input
            type="text"
            maxLength={MAX_FOOTER}
            value={draft.footer_text}
            onChange={e => setDraft(d => ({ ...d, footer_text: e.target.value }))}
            placeholder='留空则使用默认 "yuanju.com"'
          />
          <small>{draft.footer_text.length} / {MAX_FOOTER}</small>
        </label>
      </section>

      <section className="brand-section">
        <h2>水印</h2>
        <div className="brand-radio-row">
          {([
            ['none', '无水印'],
            ['bottom', '底部文字水印'],
            ['diagonal', '满页对角水印'],
          ] as const).map(([val, label]) => (
            <label key={val}>
              <input
                type="radio"
                name="watermark_mode"
                value={val}
                checked={draft.watermark_mode === val}
                onChange={() => setDraft(d => ({ ...d, watermark_mode: val }))}
              />
              {label}
            </label>
          ))}
        </div>
        <label className="brand-field">
          <span>水印文字</span>
          <input
            type="text"
            maxLength={MAX_WATERMARK}
            value={draft.watermark_text}
            onChange={e => setDraft(d => ({ ...d, watermark_text: e.target.value }))}
            disabled={draft.watermark_mode === 'none'}
            placeholder={draft.watermark_mode === 'none' ? '请先选择水印模式' : '建议填写'}
          />
          <small>{draft.watermark_text.length} / {MAX_WATERMARK}</small>
        </label>
      </section>

      <section className="brand-section brand-preview-section">
        <h2>预览</h2>
        <BrandPreviewCard brand={draft} />
      </section>

      <div className="brand-footer-actions">
        <button type="button" className="btn btn-ghost" onClick={onReset}>重置默认</button>
        <button
          type="button"
          className="btn btn-primary"
          onClick={onSave}
          disabled={!dirty || saving}
        >
          {saving ? '保存中...' : '保存'}
        </button>
      </div>

      <p className="brand-tip">
        提示：设置不影响 ResultPage 网页本身展示，仅作用于"保存图片"与"导出 PDF"的产物。
      </p>

      <Link to="/profile" className="brand-bottom-link">返回个人中心</Link>

      <LogoCropModal
        sourceDataUrl={cropSourceUrl ?? ''}
        open={!!cropSourceUrl}
        onConfirm={handleCropConfirm}
        onCancel={() => setCropSourceUrl(null)}
      />
    </main>
  )
}
