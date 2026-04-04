import { useEffect, useState } from 'react'
import { Users, CheckCircle, XCircle } from 'lucide-react'
import { adminCelebritiesAPI } from '../../lib/adminApi'

interface Celebrity {
  id: string
  name: string
  gender: string
  traits: string
  career: string
  active: boolean
  created_at: string
}

const initialForm = { name: '', gender: '男', traits: '', career: '', active: true }

export default function AdminCelebritiesPage() {
  const [celebs, setCelebs] = useState<Celebrity[]>([])
  const [loading, setLoading] = useState(true)
  const [showModal, setShowModal] = useState(false)
  const [editing, setEditing] = useState<Celebrity | null>(null)
  const [form, setForm] = useState(initialForm)
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState('')
  const [showAIModal, setShowAIModal] = useState(false)
  const [aiGenerating, setAIGenerating] = useState(false)
  const [aiTopic, setAiTopic] = useState('')
  const [aiCount, setAiCount] = useState(10)
  const [aiError, setAiError] = useState('')

  const load = () => {
    adminCelebritiesAPI.list().then(r => {
      setCelebs(r.data.data || [])
    }).finally(() => setLoading(false))
  }

  useEffect(() => { load() }, [])

  const openCreate = () => {
    setEditing(null)
    setForm(initialForm)
    setError('')
    setShowModal(true)
  }

  const openEdit = (c: Celebrity) => {
    setEditing(c)
    setForm({
      name: c.name,
      gender: c.gender || '男',
      traits: c.traits || '',
      career: c.career || '',
      active: c.active
    })
    setError('')
    setShowModal(true)
  }

  const handleSave = async () => {
    if (!form.name) { setError('请填写姓名'); return }
    setSaving(true); setError('')
    try {
      if (editing) {
        await adminCelebritiesAPI.update(editing.id, form)
      } else {
        await adminCelebritiesAPI.create(form)
      }
      setShowModal(false)
      load()
    } catch (e: unknown) {
      setError(e instanceof Error ? e.message : '保存失败')
    } finally { setSaving(false) }
  }

  const toggleActive = async (c: Celebrity) => {
    try {
      await adminCelebritiesAPI.update(c.id, {
        name: c.name,
        gender: c.gender,
        traits: c.traits,
        career: c.career,
        active: !c.active
      })
      load()
    } catch (e: unknown) {
      alert(e instanceof Error ? e.message : '切换失败')
    }
  }

  const handleDelete = async (c: Celebrity) => {
    if (!confirm(`确定删除名人 "${c.name}"？`)) return
    try { await adminCelebritiesAPI.delete(c.id); load() } catch (e: unknown) {
      alert(e instanceof Error ? e.message : '删除失败')
    }
  }

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 24 }}>
        <h1 className="admin-page-title" style={{ margin: 0, display: 'flex', alignItems: 'center', gap: 8 }}>
          <Users size={24} /> 名人库管理
        </h1>
        <div style={{ display: 'flex', gap: 12 }}>
          <button className="admin-btn admin-btn-ghost" onClick={() => { setAiTopic(''); setAiCount(10); setAiError(''); setShowAIModal(true); }}>✨ AI 自动收集</button>
          <button className="admin-btn admin-btn-primary" onClick={openCreate}>+ 添加名人</button>
        </div>
      </div>

      {loading ? <div className="admin-loading">加载中...</div> : (
        <div className="admin-card" style={{ padding: 0, overflow: 'hidden' }}>
          <table className="admin-table">
            <thead>
              <tr>
                <th style={{ width: '120px' }}>姓名</th>
                <th style={{ width: '60px' }}>性别</th>
                <th style={{ width: '150px' }}>职业/头衔</th>
                <th>八字特征 / 命理标签</th>
                <th style={{ width: '80px' }}>状态</th>
                <th style={{ width: '150px' }}>操作</th>
              </tr>
            </thead>
            <tbody>
              {celebs.length === 0 && (
                <tr><td colSpan={6} style={{ textAlign: 'center', color: '#666', padding: 40 }}>
                  暂无名人数据，点击右上角添加
                </td></tr>
              )}
              {celebs.map(c => (
                <tr key={c.id}>
                  <td style={{ fontWeight: 600, color: '#e8e8e8' }}>{c.name}</td>
                  <td style={{ color: '#aaa' }}>{c.gender}</td>
                  <td style={{ color: '#aaa' }}>{c.career}</td>
                  <td style={{ color: '#888', fontSize: 13, maxWidth: 300, whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis' }} title={c.traits}>
                    {c.traits || '-'}
                  </td>
                  <td>
                    <span className={`badge ${c.active ? 'badge-active' : 'badge-inactive'}`} style={{ display: 'inline-flex', alignItems: 'center', gap: 4 }}>
                      {c.active ? <><CheckCircle size={12} /> 启用</> : <><XCircle size={12} /> 停用</>}
                    </span>
                  </td>
                  <td>
                    <div style={{ display: 'flex', gap: 8 }}>
                      <button className="admin-btn admin-btn-ghost" style={{ padding: '4px 10px', fontSize: 12 }}
                        onClick={() => toggleActive(c)}>{c.active ? '下线' : '上线'}</button>
                      <button className="admin-btn admin-btn-ghost" style={{ padding: '4px 10px', fontSize: 12 }}
                        onClick={() => openEdit(c)}>编辑</button>
                      <button className="admin-btn admin-btn-danger" style={{ padding: '4px 10px', fontSize: 12 }}
                        onClick={() => handleDelete(c)}>删除</button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {showModal && (
        <div className="admin-modal-overlay" onClick={e => e.target === e.currentTarget && setShowModal(false)}>
          <div className="admin-modal">
            <div className="admin-modal-title">{editing ? '编辑名人' : '添加名人'}</div>

            {error && <div className="admin-error">{error}</div>}

            <div className="admin-form-row">
              <div className="admin-form-group">
                <label className="admin-form-label">姓名</label>
                <input className="admin-form-input" value={form.name}
                  onChange={e => setForm(f => ({ ...f, name: e.target.value }))} placeholder="例如：乔布斯" />
              </div>
              <div className="admin-form-group">
                <label className="admin-form-label">性别</label>
                <select className="admin-form-select" value={form.gender}
                  onChange={e => setForm(f => ({ ...f, gender: e.target.value }))}>
                  <option value="男">男</option>
                  <option value="女">女</option>
                </select>
              </div>
            </div>

            <div className="admin-form-group" style={{ marginBottom: 16 }}>
              <label className="admin-form-label">职业 / 身份</label>
              <input className="admin-form-input" value={form.career}
                onChange={e => setForm(f => ({ ...f, career: e.target.value }))} placeholder="例如：苹果公司联合创始人" />
            </div>

            <div className="admin-form-group" style={{ marginBottom: 16 }}>
              <label className="admin-form-label">八字关键特征（最重要，供给 AI 作为参考）</label>
              <textarea className="admin-form-input" rows={6} value={form.traits}
                onChange={e => setForm(f => ({ ...f, traits: e.target.value }))}
                placeholder="例如：日主丙火生于子月，水火既济，伤官生财，性格追求极致且极具创新。" />
            </div>

            <div className="admin-form-group" style={{ marginBottom: 16, display: 'flex', alignItems: 'center', gap: 8 }}>
              <input type="checkbox" id="active-checkbox" checked={form.active}
                onChange={e => setForm(f => ({ ...f, active: e.target.checked }))} />
              <label htmlFor="active-checkbox" style={{ cursor: 'pointer', color: '#ccc' }}>
                立即作为 AI 推荐候选（启用）
              </label>
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

      {showAIModal && (
        <div className="admin-modal-overlay" onClick={e => e.target === e.currentTarget && !aiGenerating && setShowAIModal(false)}>
          <div className="admin-modal">
            <div className="admin-modal-title">✨ AI 自动收集名人</div>
            <p style={{ color: '#888', fontSize: 13, marginBottom: 16 }}>
              设定主题和数量，系统将调用大模型自动生成相关的名人记录并填充至数据库。
            </p>

            {aiError && <div className="admin-error">{aiError}</div>}

            <div className="admin-form-group" style={{ marginBottom: 16 }}>
              <label className="admin-form-label">收集主题范围</label>
              <input className="admin-form-input" value={aiTopic}
                onChange={e => setAiTopic(e.target.value)} placeholder="例如：盛唐诗人、诺贝尔物理学家、当代科技大佬" disabled={aiGenerating} />
            </div>

            <div className="admin-form-group" style={{ marginBottom: 24 }}>
              <label className="admin-form-label">生成数量 ({aiCount}个)</label>
              <input type="range" min={1} max={20} value={aiCount}
                onChange={e => setAiCount(Number(e.target.value))} style={{ width: '100%', cursor: aiGenerating ? 'not-allowed' : 'pointer' }} disabled={aiGenerating} />
            </div>

            <div className="admin-modal-actions">
              <button className="admin-btn admin-btn-ghost" onClick={() => setShowAIModal(false)} disabled={aiGenerating}>取消</button>
              <button className="admin-btn admin-btn-primary" 
                onClick={async () => {
                  if (!aiTopic.trim()) { setAiError('请输入生成主题范围'); return }
                  setAIGenerating(true); setAiError('');
                  try {
                    const res = await adminCelebritiesAPI.generateAI({ topic: aiTopic, count: aiCount });
                    alert(`成功生成并入库 ${res.data.data.inserted_count} 条名人记录！`);
                    setShowAIModal(false);
                    load();
                  } catch (e: any) {
                    setAiError(e.message || '生成失败');
                  } finally {
                    setAIGenerating(false);
                  }
                }} disabled={aiGenerating} style={{ width: 140 }}>
                {aiGenerating ? 'AI 沉思中...' : '开始生成'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
