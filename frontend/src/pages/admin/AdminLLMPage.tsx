import { useEffect, useState } from 'react'
import { Bot, CheckCircle } from 'lucide-react'
import { adminLLMAPI } from '../../lib/adminApi'

interface Provider {
  id: string; name: string; type: string; base_url: string
  model: string; api_key_masked: string; api_key_preview: string
  thinking_enabled: boolean; active: boolean
}

interface PresetType {
  type: string; name: string; base_url: string; model: string;
}

interface TestResult {
  ok: boolean; latency_ms?: number; error?: string
}

const initialForm = { name: '', type: 'deepseek', base_url: '', model: '', api_key: '', thinking_enabled: false }

export default function AdminLLMPage() {
  const [providers, setProviders] = useState<Provider[]>([])
  const [presetTypes, setPresetTypes] = useState<PresetType[]>([])
  const [loading, setLoading] = useState(true)
  const [showModal, setShowModal] = useState(false)
  const [editing, setEditing] = useState<Provider | null>(null)
  const [form, setForm] = useState(initialForm)
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState('')
  const [testingId, setTestingId] = useState<string | null>(null)
  const [testResults, setTestResults] = useState<Record<string, TestResult>>({})

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
    setForm({ name: p.name, type: p.type, base_url: p.base_url, model: p.model, api_key: '', thinking_enabled: p.thinking_enabled })
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
        await adminLLMAPI.update(editing.id, { name: form.name, base_url: form.base_url, model: form.model, thinking_enabled: form.thinking_enabled, ...(form.api_key ? { api_key: form.api_key } : {}) })
      } else {
        await adminLLMAPI.create({ name: form.name, type: form.type, base_url: form.base_url, model: form.model, api_key: form.api_key, thinking_enabled: form.thinking_enabled })
      }
      setShowModal(false)
      load()
    } catch (e: unknown) {
      setError(e instanceof Error ? e.message : '保存失败')
    } finally { setSaving(false) }
  }

  const handleActivate = async (id: string) => {
    setProviders(ps => ps.map(p => ({ ...p, active: p.id === id })))
    try { await adminLLMAPI.activate(id) } catch { load() }
  }

  const handleDelete = async (p: Provider) => {
    if (!confirm(`确定删除 "${p.name}"？`)) return
    try { await adminLLMAPI.delete(p.id); load() } catch (e: unknown) {
      alert(e instanceof Error ? e.message : '删除失败')
    }
  }

  const handleTest = async (id: string) => {
    setTestingId(id)
    try {
      const res = await adminLLMAPI.test(id)
      setTestResults(prev => ({ ...prev, [id]: res.data }))
    } catch (e: unknown) {
      setTestResults(prev => ({ ...prev, [id]: { ok: false, error: e instanceof Error ? e.message : '测试失败' } }))
    } finally {
      setTestingId(null)
    }
  }

  const keyDisplay = (p: Provider) => p.api_key_preview || p.api_key_masked || '—'

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
                <th>名称</th><th>类型</th><th>模型</th><th>API Key</th><th>思考模式</th><th>状态</th><th>操作</th>
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
                  <td>
                    <code style={{ fontSize: 12, color: '#888' }}>{keyDisplay(p)}</code>
                  </td>
                  <td>
                    {p.thinking_enabled
                      ? <span style={{ fontSize: 12, color: '#a78bfa', fontWeight: 600 }}>开启</span>
                      : <span style={{ fontSize: 12, color: '#555' }}>—</span>}
                  </td>
                  <td>
                    <div style={{ display: 'flex', flexDirection: 'column', gap: 4 }}>
                      <span className={`badge ${p.active ? 'badge-active' : 'badge-inactive'}`} style={{ display: 'inline-flex', alignItems: 'center', gap: 4 }}>
                        {p.active ? <><CheckCircle size={12} /> 激活</> : '待机'}
                      </span>
                      {testResults[p.id] && (
                        <span style={{ fontSize: 11, color: testResults[p.id].ok ? '#4ade80' : '#f87171' }}>
                          {testResults[p.id].ok
                            ? `✓ ${testResults[p.id].latency_ms}ms`
                            : `✗ ${testResults[p.id].error}`}
                        </span>
                      )}
                    </div>
                  </td>
                  <td>
                    <div style={{ display: 'flex', gap: 8, flexWrap: 'wrap' }}>
                      {!p.active && (
                        <button className="admin-btn admin-btn-ghost" style={{ padding: '4px 10px', fontSize: 12 }}
                          onClick={() => handleActivate(p.id)}>激活</button>
                      )}
                      <button
                        className="admin-btn admin-btn-ghost"
                        style={{ padding: '4px 10px', fontSize: 12 }}
                        disabled={testingId === p.id}
                        onClick={() => handleTest(p.id)}
                      >
                        {testingId === p.id ? '测试中…' : '测试'}
                      </button>
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
              <label className="admin-form-label">思考模式</label>
              <label style={{ display: 'flex', alignItems: 'center', gap: 10, cursor: 'pointer' }}>
                <input
                  type="checkbox"
                  checked={form.thinking_enabled}
                  onChange={e => setForm(f => ({ ...f, thinking_enabled: e.target.checked }))}
                  style={{ width: 16, height: 16, accentColor: '#a78bfa' }}
                />
                <span style={{ fontSize: 13, color: form.thinking_enabled ? '#a78bfa' : '#888' }}>
                  {form.thinking_enabled ? '已开启（deepseek-v4-pro 等推理模型）' : '已关闭（deepseek-v4-flash 等标准模型）'}
                </span>
              </label>
            </div>

            <div className="admin-form-group" style={{ marginBottom: 16 }}>
              <label className="admin-form-label">
                API Key {editing && <span style={{ color: '#666' }}>(留空则不修改)</span>}
              </label>
              {editing && (editing.api_key_preview || editing.api_key_masked) && (
                <div style={{ fontSize: 12, color: '#888', marginBottom: 6 }}>
                  当前：<code style={{ color: '#a78bfa' }}>{editing.api_key_preview || editing.api_key_masked}</code>
                </div>
              )}
              <input className="admin-form-input" type="password" value={form.api_key}
                onChange={e => setForm(f => ({ ...f, api_key: e.target.value }))}
                placeholder={editing ? '输入新密钥（留空保留当前）' : 'sk-xxxxxxxxxx'} />
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
