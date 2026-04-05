import { useEffect, useState } from 'react'
import { Bot, CheckCircle } from 'lucide-react'
import { adminLLMAPI } from '../../lib/adminApi'

interface Provider {
  id: string; name: string; type: string; base_url: string
  model: string; api_key_masked: string; active: boolean
}

interface PresetType {
  type: string; name: string; base_url: string; model: string;
}

const initialForm = { name: '', type: 'deepseek', base_url: '', model: '', api_key: '' }

export default function AdminLLMPage() {
  const [providers, setProviders] = useState<Provider[]>([])
  const [presetTypes, setPresetTypes] = useState<PresetType[]>([])
  const [loading, setLoading] = useState(true)
  const [showModal, setShowModal] = useState(false)
  const [editing, setEditing] = useState<Provider | null>(null)
  const [form, setForm] = useState(initialForm)
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState('')

  const load = () => {
    adminLLMAPI.list().then(r => {
      setProviders(r.data.providers || [])
      setPresetTypes(r.data.predefined || [])
    }).finally(() => setLoading(false))
  }

  useEffect(() => { load() }, [])

  const openCreate = () => {
    setEditing(null)
    setForm(initialForm)
    setError('')
    setShowModal(true)
  }

  const openEdit = (p: Provider) => {
    setEditing(p)
    setForm({ name: p.name, type: p.type, base_url: p.base_url, model: p.model, api_key: '' })
    setError('')
    setShowModal(true)
  }

  const handleTypeChange = (presetName: string) => {
    const preset = presetTypes.find(t => t.name === presetName)
    if (preset) setForm(f => ({ ...f, type: preset.type, name: preset.name, base_url: preset.base_url, model: preset.model }))
  }

  const handleSave = async () => {
    if (!form.name || !form.base_url || !form.model) { setError('请填写完整信息'); return }
    if (!editing && !form.api_key) { setError('新建时 API Key 必填'); return }
    setSaving(true); setError('')
    try {
      if (editing) {
        await adminLLMAPI.update(editing.id, { name: form.name, base_url: form.base_url, model: form.model, ...(form.api_key ? { api_key: form.api_key } : {}) })
      } else {
        await adminLLMAPI.create({ name: form.name, type: form.type, base_url: form.base_url, model: form.model, api_key: form.api_key })
      }
      setShowModal(false)
      load()
    } catch (e: unknown) {
      setError(e instanceof Error ? e.message : '保存失败')
    } finally { setSaving(false) }
  }

  const handleActivate = async (id: string) => {
    // 乐观更新
    setProviders(ps => ps.map(p => ({ ...p, active: p.id === id })))
    try { await adminLLMAPI.activate(id) } catch { load() }
  }

  const handleDelete = async (p: Provider) => {
    if (!confirm(`确定删除 "${p.name}"？`)) return
    try { await adminLLMAPI.delete(p.id); load() } catch (e: unknown) {
      alert(e instanceof Error ? e.message : '删除失败')
    }
  }

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 24 }}>
        <h1 className="admin-page-title" style={{ margin: 0, display: 'flex', alignItems: 'center', gap: 8 }}>
          <Bot size={24} /> LLM 管理
        </h1>
        <button id="btn-add-provider" className="admin-btn admin-btn-primary" onClick={openCreate}>+ 添加 Provider</button>
      </div>

      {loading ? <div className="admin-loading">加载中...</div> : (
        <div className="admin-card" style={{ padding: 0, overflow: 'hidden' }}>
          <table className="admin-table">
            <thead>
              <tr>
                <th>名称</th><th>类型</th><th>模型</th><th>API Key</th><th>状态</th><th>操作</th>
              </tr>
            </thead>
            <tbody>
              {providers.length === 0 && (
                <tr><td colSpan={6} style={{ textAlign: 'center', color: '#666', padding: 40 }}>
                  暂无 Provider，点击右上角添加
                </td></tr>
              )}
              {providers.map(p => (
                <tr key={p.id}>
                  <td style={{ fontWeight: 600, color: '#e8e8e8' }}>{p.name}</td>
                  <td style={{ color: '#888', fontSize: 12 }}>{p.type}</td>
                  <td style={{ color: '#aaa', fontSize: 13 }}>{p.model}</td>
                  <td><code style={{ fontSize: 12, color: '#666' }}>{p.api_key_masked || '已加密'}</code></td>
                  <td>
                    <span className={`badge ${p.active ? 'badge-active' : 'badge-inactive'}`} style={{ display: 'inline-flex', alignItems: 'center', gap: 4 }}>
                      {p.active ? <><CheckCircle size={12} /> 激活</> : '待机'}
                    </span>
                  </td>
                  <td>
                    <div style={{ display: 'flex', gap: 8 }}>
                      {!p.active && (
                        <button className="admin-btn admin-btn-ghost" style={{ padding: '4px 10px', fontSize: 12 }}
                          onClick={() => handleActivate(p.id)}>激活</button>
                      )}
                      <button className="admin-btn admin-btn-ghost" style={{ padding: '4px 10px', fontSize: 12 }}
                        onClick={() => openEdit(p)}>编辑</button>
                      {!p.active && (
                        <button className="admin-btn admin-btn-danger" style={{ padding: '4px 10px', fontSize: 12 }}
                          onClick={() => handleDelete(p)}>删除</button>
                      )}
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {/* 新增/编辑 Modal */}
      {showModal && (
        <div className="admin-modal-overlay" onClick={e => e.target === e.currentTarget && setShowModal(false)}>
          <div className="admin-modal">
            <div className="admin-modal-title">{editing ? '编辑 Provider' : '添加 Provider'}</div>

            {error && <div className="admin-error">{error}</div>}

            <div className="admin-form-group" style={{ marginBottom: 16 }}>
              <label className="admin-form-label">Provider 类型</label>
              <select className="admin-form-select" value={form.name}
                onChange={e => handleTypeChange(e.target.value)} disabled={!!editing}>
                {presetTypes.map((t, idx) => <option key={idx} value={t.name}>{t.name}</option>)}
              </select>
            </div>

            <div className="admin-form-row">
              <div className="admin-form-group">
                <label className="admin-form-label">显示名称</label>
                <input className="admin-form-input" value={form.name}
                  onChange={e => setForm(f => ({ ...f, name: e.target.value }))} placeholder="DeepSeek" />
              </div>
              <div className="admin-form-group">
                <label className="admin-form-label">模型 ID</label>
                <input className="admin-form-input" value={form.model}
                  onChange={e => setForm(f => ({ ...f, model: e.target.value }))} placeholder="deepseek-chat" />
              </div>
            </div>

            <div className="admin-form-group" style={{ marginBottom: 16 }}>
              <label className="admin-form-label">Base URL</label>
              <input className="admin-form-input" value={form.base_url}
                onChange={e => setForm(f => ({ ...f, base_url: e.target.value }))} placeholder="https://api.deepseek.com" />
            </div>

            <div className="admin-form-group" style={{ marginBottom: 16 }}>
              <label className="admin-form-label">API Key {editing && <span style={{ color: '#666' }}>(留空则不修改)</span>}</label>
              <input className="admin-form-input" type="password" value={form.api_key}
                onChange={e => setForm(f => ({ ...f, api_key: e.target.value }))}
                placeholder={editing ? '•••••• (不修改请留空)' : 'sk-xxxxxxxxxx'} />
            </div>

            <div className="admin-modal-actions">
              <button className="admin-btn admin-btn-ghost" onClick={() => setShowModal(false)}>取消</button>
              <button className="admin-btn admin-btn-primary" onClick={handleSave} disabled={saving}>
                {saving ? '保存中...' : '保存'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
