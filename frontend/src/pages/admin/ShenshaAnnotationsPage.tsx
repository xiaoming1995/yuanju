import { useState, useEffect } from 'react'
import { BookOpen } from 'lucide-react'
import { adminShenshaAPI } from '../../lib/adminApi'

interface ShenshaAnnotation {
  id: string
  name: string
  polarity: string
  category: string
  short_desc: string
  description: string
  updated_at: string
}

const polarityLabel: Record<string, string> = {
  ji: '吉',
  xiong: '凶',
  zhong: '中',
}

const polarityColor: Record<string, string> = {
  ji: '#4ade80',
  xiong: '#f87171',
  zhong: '#a78bfa',
}

export default function ShenshaAnnotationsPage() {
  const [items, setItems] = useState<ShenshaAnnotation[]>([])
  const [loading, setLoading] = useState(true)
  const [editName, setEditName] = useState<string | null>(null)
  const [editCategory, setEditCategory] = useState('')
  const [editShortDesc, setEditShortDesc] = useState('')
  const [editDescription, setEditDescription] = useState('')
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState('')

  const load = async () => {
    setLoading(true)
    try {
      const res = await adminShenshaAPI.list()
      setItems(res.data.data || [])
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { load() }, [])

  const startEdit = (item: ShenshaAnnotation) => {
    setEditName(item.name)
    setEditCategory(item.category)
    setEditShortDesc(item.short_desc)
    setEditDescription(item.description)
    setError('')
  }

  const cancelEdit = () => setEditName(null)

  const save = async () => {
    if (!editName) return
    setSaving(true)
    setError('')
    try {
      await adminShenshaAPI.update(editName, {
        category: editCategory,
        short_desc: editShortDesc,
        description: editDescription,
      })
      setItems(prev => prev.map(it =>
        it.name === editName
          ? { ...it, category: editCategory, short_desc: editShortDesc, description: editDescription }
          : it
      ))
      setEditName(null)
    } catch (e: unknown) {
      setError(e instanceof Error ? e.message : '保存失败')
    } finally {
      setSaving(false)
    }
  }

  return (
    <div>
      <h1 className="admin-page-title" style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
        <BookOpen size={24} /> 神煞注解管理
      </h1>

      {loading ? (
        <div className="admin-loading">加载中…</div>
      ) : (
        <div className="admin-card">
          <table className="admin-table" style={{ width: '100%' }}>
            <thead>
              <tr>
                <th style={{ width: 60 }}>名称</th>
                <th style={{ width: 50 }}>极性</th>
                <th style={{ width: 80 }}>分类</th>
                <th>一句话简介</th>
                <th style={{ width: 80 }}>操作</th>
              </tr>
            </thead>
            <tbody>
              {items.map(item => (
                editName === item.name ? (
                  <tr key={item.name} style={{ background: '#1a1a3e' }}>
                    <td style={{ fontWeight: 700, color: polarityColor[item.polarity] ?? '#e0e0e0' }}>{item.name}</td>
                    <td>{polarityLabel[item.polarity] ?? item.polarity}</td>
                    <td>
                      <input
                        value={editCategory}
                        onChange={e => setEditCategory(e.target.value)}
                        placeholder="贵人系"
                        style={{ width: '100%', background: '#0d0d1a', color: '#e0e0e0', border: '1px solid #444', borderRadius: 4, padding: '2px 6px', fontSize: 12 }}
                      />
                    </td>
                    <td>
                      <input
                        value={editShortDesc}
                        onChange={e => setEditShortDesc(e.target.value)}
                        placeholder="一句话简介…"
                        style={{ width: '100%', background: '#0d0d1a', color: '#e0e0e0', border: '1px solid #444', borderRadius: 4, padding: '2px 6px', fontSize: 12 }}
                      />
                    </td>
                    <td>
                      <div style={{ display: 'flex', gap: 4 }}>
                        <button className="admin-btn" style={{ padding: '2px 8px', fontSize: 12 }} onClick={save} disabled={saving}>
                          {saving ? '…' : '保存'}
                        </button>
                        <button className="admin-btn" style={{ padding: '2px 8px', fontSize: 12, background: 'transparent' }} onClick={cancelEdit}>
                          取消
                        </button>
                      </div>
                    </td>
                  </tr>
                ) : (
                  <tr key={item.name} style={{ cursor: 'pointer' }} onClick={() => startEdit(item)}>
                    <td style={{ fontWeight: 700, color: polarityColor[item.polarity] ?? '#e0e0e0' }}>{item.name}</td>
                    <td style={{ color: polarityColor[item.polarity] ?? '#888' }}>{polarityLabel[item.polarity] ?? item.polarity}</td>
                    <td style={{ color: '#aaa', fontSize: 12 }}>{item.category || '—'}</td>
                    <td style={{ color: '#aaa', fontSize: 12 }}>{item.short_desc || '—'}</td>
                    <td>
                      <button className="admin-btn" style={{ padding: '2px 8px', fontSize: 12 }} onClick={e => { e.stopPropagation(); startEdit(item) }}>
                        编辑
                      </button>
                    </td>
                  </tr>
                )
              ))}
            </tbody>
          </table>
          {editName && (
            <div style={{ marginTop: 16, padding: 16, background: '#1a1a3e', borderRadius: 8 }}>
              <label style={{ color: '#aaa', fontSize: 13, display: 'block', marginBottom: 6 }}>
                {editName} — 详细说明
              </label>
              <textarea
                value={editDescription}
                onChange={e => setEditDescription(e.target.value)}
                rows={6}
                style={{ width: '100%', background: '#0d0d1a', color: '#e0e0e0', border: '1px solid #444', borderRadius: 6, padding: 10, fontSize: 13, fontFamily: 'inherit', resize: 'vertical', boxSizing: 'border-box' }}
              />
              {error && <div style={{ color: '#f87171', fontSize: 13, marginTop: 6 }}>{error}</div>}
              <div style={{ display: 'flex', gap: 8, marginTop: 10 }}>
                <button className="admin-btn" onClick={save} disabled={saving}>{saving ? '保存中…' : '保存'}</button>
                <button className="admin-btn" style={{ background: 'transparent' }} onClick={cancelEdit}>取消</button>
              </div>
            </div>
          )}
        </div>
      )}
    </div>
  )
}
