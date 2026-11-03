import { useEffect, useState } from 'react'
import { AdminRequestState } from '../../components/AdminRequestState'
import { adminActionErrorMessage, adminOfficialAccountGuideAPI, type OfficialAccountGuide } from '../../lib/adminApi'
import { missingOfficialAccountFields } from './officialAccountGuideForm'
import './OfficialAccountGuidePage.css'

const emptyGuide: OfficialAccountGuide = { enabled: false, name: '', wechat_id: '', description: '', qr_path: '' }

export default function OfficialAccountGuidePage() {
  const [saved, setSaved] = useState<OfficialAccountGuide | null>(null)
  const [draft, setDraft] = useState<OfficialAccountGuide>(emptyGuide)
  const [loading, setLoading] = useState(true)
  const [loadError, setLoadError] = useState('')
  const [reloadKey, setReloadKey] = useState(0)
  const [uploading, setUploading] = useState(false)
  const [saving, setSaving] = useState(false)
  const [feedback, setFeedback] = useState('')

  useEffect(() => {
    let cancelled = false
    adminOfficialAccountGuideAPI.get().then(({ data }) => {
      if (cancelled) return
      setSaved(data.data)
      setDraft(data.data)
      setLoadError('')
    }).catch(error => {
      if (!cancelled) setLoadError(adminActionErrorMessage(error))
    }).finally(() => { if (!cancelled) setLoading(false) })
    return () => { cancelled = true }
  }, [reloadKey])

  const update = (patch: Partial<OfficialAccountGuide>) => { setDraft(current => ({ ...current, ...patch })); setFeedback('') }

  const upload = async (file?: File) => {
    if (!file) return
    if (!['image/png', 'image/jpeg', 'image/webp'].includes(file.type) || file.size > 2 * 1024 * 1024) {
      setFeedback('请上传不超过 2MB 的 PNG、JPG 或 WebP 图片')
      return
    }
    setUploading(true)
    setFeedback('')
    try {
      const { data } = await adminOfficialAccountGuideAPI.uploadQR(file)
      update({ qr_path: data.data.qr_path })
      setFeedback('二维码已上传，请保存配置后生效')
    } catch (error) {
      setFeedback(adminActionErrorMessage(error))
    } finally {
      setUploading(false)
    }
  }

  const save = async () => {
    const missing = missingOfficialAccountFields(draft)
    if (missing.length) { setFeedback(`启用前请填写：${missing.join('、')}`); return }
    setSaving(true)
    setFeedback('')
    try {
      const { data } = await adminOfficialAccountGuideAPI.update(draft)
      setSaved(data.data)
      setDraft(data.data)
      setFeedback('已保存，小程序重新进入“我的”页后生效')
    } catch (error) {
      setFeedback(adminActionErrorMessage(error))
    } finally {
      setSaving(false)
    }
  }

  if (loading) return <div className="admin-page"><h1>公众号引导</h1><AdminRequestState state="loading" message="正在读取公众号引导配置…" /></div>
  if (loadError) return <div className="admin-page"><h1>公众号引导</h1><AdminRequestState state="error" message={loadError} onRetry={() => { setLoading(true); setLoadError(''); setReloadKey(n => n + 1) }} /></div>

  return <div className="admin-page official-account-admin">
    <h1>公众号引导</h1>
    <p className="admin-hint">配置小程序“我的”页的公众号入口。二维码为展示与预览用途；用户可复制微信号搜索公众号。</p>
    <div className="admin-card official-account-admin-card">
      <label className="official-account-switch"><input type="checkbox" checked={draft.enabled} onChange={event => update({ enabled: event.target.checked })} /><span>在小程序“我的”页显示</span></label>
      <p className="admin-form-hint">默认关闭。关闭后保留已填写的信息，方便再次启用。</p>

      <div className="admin-form-group"><label className="admin-form-label" htmlFor="official-account-name">公众号名称</label><input id="official-account-name" className="admin-form-input" maxLength={40} value={draft.name} onChange={event => update({ name: event.target.value })} placeholder="用户看到的名称" /></div>
      <div className="admin-form-group"><label className="admin-form-label" htmlFor="official-account-id">微信号</label><input id="official-account-id" className="admin-form-input" maxLength={64} value={draft.wechat_id} onChange={event => update({ wechat_id: event.target.value })} placeholder="供用户复制后在微信搜索" /></div>
      <div className="admin-form-group"><label className="admin-form-label" htmlFor="official-account-description">引导说明</label><input id="official-account-description" className="admin-form-input" maxLength={120} value={draft.description} onChange={event => update({ description: event.target.value })} placeholder="例如：获取缘聚内容更新" /></div>
      <div className="admin-form-group">
        <label className="admin-form-label" htmlFor="official-account-qr">公众号二维码</label>
        <input id="official-account-qr" type="file" accept="image/png,image/jpeg,image/webp" disabled={uploading || saving} onChange={event => { void upload(event.target.files?.[0]); event.target.value = '' }} />
        <div className="admin-form-hint">上传真实公众号二维码，支持 PNG、JPG、WebP，最大 2MB。请勿使用设计稿中的示意二维码。</div>
        {draft.qr_path && <div className="official-account-qr-preview"><img src={draft.qr_path} alt="当前选择的公众号二维码" /><span>{draft.qr_path === saved?.qr_path ? '当前保存的二维码' : '新上传的二维码，保存后生效'}</span></div>}
      </div>
      <div className="official-account-admin-actions"><button type="button" className="btn btn-primary" disabled={saving || uploading || !saved || JSON.stringify(draft) === JSON.stringify(saved)} onClick={() => void save()}>{saving ? '保存中…' : '保存配置'}</button>{uploading && <span>正在上传二维码…</span>}{feedback && <span role="status">{feedback}</span>}</div>
    </div>
    <p className="admin-hint">启用前请在 iPhone 和安卓微信中用真实二维码核对预览与长按表现。小程序始终提供复制微信号入口。</p>
  </div>
}
