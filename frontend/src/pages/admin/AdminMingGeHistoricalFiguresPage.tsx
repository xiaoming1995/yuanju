import { useEffect, useState } from 'react'
import { Landmark, Pencil, Plus, Trash2 } from 'lucide-react'
import {
  adminMingGeHistoricalFiguresAPI,
  type AdminMingGeHistoricalFigure,
  type AdminMingGeHistoricalFigureInput,
} from '../../lib/adminApi'
import './AdminMingGeHistoricalFiguresPage.css'

const emptyForm: AdminMingGeHistoricalFigureInput = {
  ming_ge: '',
  figure_name: '',
  era: '',
  identity: '',
  historical_memory: '',
  turning_point: '',
  turning_point_year: '',
  source_title: '',
  source_url: '',
  birth_data_precision: 'unknown',
  bazi_verification_note: '',
  dayun_period: '',
  dayun_explanation: '',
  show_dayun: false,
  display_order: 0,
  review_status: 'draft',
}

function toForm(figure: AdminMingGeHistoricalFigure): AdminMingGeHistoricalFigureInput {
  const { id, ...form } = figure
  void id
  return {
    ...emptyForm,
    ...form,
    bazi_verification_note: figure.bazi_verification_note || '',
    dayun_period: figure.dayun_period || '',
    dayun_explanation: figure.dayun_explanation || '',
  }
}

export default function AdminMingGeHistoricalFiguresPage() {
  const [figures, setFigures] = useState<AdminMingGeHistoricalFigure[]>([])
  const [loading, setLoading] = useState(true)
  const [showEditor, setShowEditor] = useState(false)
  const [editing, setEditing] = useState<AdminMingGeHistoricalFigure | null>(null)
  const [form, setForm] = useState<AdminMingGeHistoricalFigureInput>(emptyForm)
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState('')

  const load = async () => {
    try {
      const response = await adminMingGeHistoricalFiguresAPI.list()
      setFigures(response.data.data || [])
    } catch (requestError) {
      setError(requestError instanceof Error ? requestError.message : '加载失败')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { void load() }, [])

  const openCreate = () => {
    setEditing(null)
    setForm(emptyForm)
    setError('')
    setShowEditor(true)
  }

  const openEdit = (figure: AdminMingGeHistoricalFigure) => {
    setEditing(figure)
    setForm(toForm(figure))
    setError('')
    setShowEditor(true)
  }

  const canShowDayun = form.birth_data_precision === 'exact_hour'
    && Boolean(form.bazi_verification_note?.trim())
    && Boolean(form.turning_point.trim())
    && Boolean(form.turning_point_year.trim())
    && Boolean(form.dayun_period?.trim())
    && Boolean(form.dayun_explanation?.trim())

  const handleSave = async () => {
    if (form.show_dayun && !canShowDayun) {
      setError('展示大运前，请补齐精确时辰、命盘核验、转折信息与大运呼应说明。')
      return
    }
    setSaving(true)
    setError('')
    try {
      if (editing) {
        await adminMingGeHistoricalFiguresAPI.update(editing.id, form)
      } else {
        await adminMingGeHistoricalFiguresAPI.create(form)
      }
      setShowEditor(false)
      setLoading(true)
      await load()
    } catch (requestError) {
      setError(requestError instanceof Error ? requestError.message : '保存失败')
    } finally {
      setSaving(false)
    }
  }

  const handleArchive = async (figure: AdminMingGeHistoricalFigure) => {
    if (!window.confirm(`归档「${figure.figure_name}」？归档后将不再在排盘页展示。`)) return
    try {
      await adminMingGeHistoricalFiguresAPI.archive(figure.id)
      setLoading(true)
      await load()
    } catch (requestError) {
      setError(requestError instanceof Error ? requestError.message : '归档失败')
    }
  }

  return (
    <div className="admin-page admin-historical-figures-page">
      <div className="admin-historical-figures-header">
        <div>
          <h1 className="admin-page-title"><Landmark size={24} /> 命格古人映照</h1>
          <p>仅维护人工审核的历史资料；不使用名人库 AI 自动生成内容。</p>
        </div>
        <button className="admin-btn admin-btn-primary" onClick={openCreate}><Plus size={16} /> 添加映照</button>
      </div>

      {error && !showEditor && <div className="admin-error">{error}</div>}
      {loading ? <div className="admin-loading">加载中...</div> : (
        <div className="admin-card admin-historical-figures-table-wrap">
          <table className="admin-table">
            <thead>
              <tr><th>命格</th><th>人物</th><th>身份 / 时代</th><th>资料与核验</th><th>状态</th><th>操作</th></tr>
            </thead>
            <tbody>
              {figures.length === 0 && <tr><td colSpan={6} className="admin-historical-empty">暂无映照资料。</td></tr>}
              {figures.map((figure) => (
                <tr key={figure.id}>
                  <td>{figure.ming_ge}</td>
                  <td><strong>{figure.figure_name}</strong></td>
                  <td>{figure.era}<br /><span>{figure.identity}</span></td>
                  <td>
                    <a href={figure.source_url} target="_blank" rel="noreferrer">{figure.source_title}</a>
                    <small>{figure.birth_data_precision === 'exact_hour' ? '时辰已核验' : '不展示大运'}</small>
                  </td>
                  <td><span className={`badge admin-historical-status status-${figure.review_status}`}>{figure.review_status}</span></td>
                  <td className="admin-historical-actions">
                    <button className="admin-icon-button" title="编辑" onClick={() => openEdit(figure)}><Pencil size={15} /></button>
                    {figure.review_status !== 'archived' && <button className="admin-icon-button is-danger" title="归档" onClick={() => void handleArchive(figure)}><Trash2 size={15} /></button>}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {showEditor && (
        <div className="admin-modal-overlay" onClick={(event) => event.target === event.currentTarget && setShowEditor(false)}>
          <div className="admin-modal admin-historical-editor">
            <div className="admin-modal-title">{editing ? '编辑古人映照' : '添加古人映照'}</div>
            {error && <div className="admin-error">{error}</div>}
            <div className="admin-form-row">
              <Field label="命格"><input className="admin-form-input" value={form.ming_ge} onChange={(event) => setForm((value) => ({ ...value, ming_ge: event.target.value }))} placeholder="例如：伤官格" /></Field>
              <Field label="人物"><input className="admin-form-input" value={form.figure_name} onChange={(event) => setForm((value) => ({ ...value, figure_name: event.target.value }))} placeholder="例如：李白" /></Field>
            </div>
            <div className="admin-form-row">
              <Field label="时代"><input className="admin-form-input" value={form.era} onChange={(event) => setForm((value) => ({ ...value, era: event.target.value }))} placeholder="例如：唐代" /></Field>
              <Field label="身份"><input className="admin-form-input" value={form.identity} onChange={(event) => setForm((value) => ({ ...value, identity: event.target.value }))} placeholder="例如：诗人" /></Field>
            </div>
            <Field label="后世如何记住他 / 她"><textarea className="admin-form-input" rows={3} value={form.historical_memory} onChange={(event) => setForm((value) => ({ ...value, historical_memory: event.target.value }))} /></Field>
            <div className="admin-form-row">
              <Field label="资料标题"><input className="admin-form-input" value={form.source_title} onChange={(event) => setForm((value) => ({ ...value, source_title: event.target.value }))} /></Field>
              <Field label="资料链接"><input className="admin-form-input" type="url" value={form.source_url} onChange={(event) => setForm((value) => ({ ...value, source_url: event.target.value }))} placeholder="https://..." /></Field>
            </div>
            <div className="admin-historical-divider">仅时辰已核验的资料可展示大运</div>
            <div className="admin-form-row">
              <Field label="出生资料精度">
                <select className="admin-form-select" value={form.birth_data_precision} onChange={(event) => setForm((value) => ({ ...value, birth_data_precision: event.target.value as AdminMingGeHistoricalFigureInput['birth_data_precision'], show_dayun: event.target.value === 'exact_hour' ? value.show_dayun : false }))}>
                  <option value="unknown">未知</option><option value="date_only">仅日期</option><option value="exact_hour">精确时辰</option>
                </select>
              </Field>
              <Field label="审核状态">
                <select className="admin-form-select" value={form.review_status} onChange={(event) => setForm((value) => ({ ...value, review_status: event.target.value as AdminMingGeHistoricalFigureInput['review_status'] }))}>
                  <option value="draft">草稿</option><option value="reviewed">已审核</option><option value="published">已发布</option><option value="archived">已归档</option>
                </select>
              </Field>
            </div>
            <Field label="命盘核验说明"><textarea className="admin-form-input" rows={2} value={form.bazi_verification_note} onChange={(event) => setForm((value) => ({ ...value, bazi_verification_note: event.target.value }))} placeholder="未展示大运时可留空。" /></Field>
            <div className="admin-form-row">
              <Field label="人生转折"><input className="admin-form-input" value={form.turning_point} onChange={(event) => setForm((value) => ({ ...value, turning_point: event.target.value }))} /></Field>
              <Field label="转折年份 / 时期"><input className="admin-form-input" value={form.turning_point_year} onChange={(event) => setForm((value) => ({ ...value, turning_point_year: event.target.value }))} /></Field>
            </div>
            <div className="admin-form-row">
              <Field label="大运阶段"><input className="admin-form-input" value={form.dayun_period} onChange={(event) => setForm((value) => ({ ...value, dayun_period: event.target.value }))} /></Field>
              <Field label="大运呼应说明"><input className="admin-form-input" value={form.dayun_explanation} onChange={(event) => setForm((value) => ({ ...value, dayun_explanation: event.target.value }))} /></Field>
            </div>
            <div className="admin-historical-editor-footer">
              <label><input type="checkbox" checked={form.show_dayun} disabled={!canShowDayun} onChange={(event) => setForm((value) => ({ ...value, show_dayun: event.target.checked }))} /> 在专业模式展示大运</label>
              {!canShowDayun && <small>补齐精确时辰、核验、转折和大运信息后可开启。</small>}
              <Field label="排序"><input className="admin-form-input" type="number" value={form.display_order} onChange={(event) => setForm((value) => ({ ...value, display_order: Number(event.target.value) || 0 }))} /></Field>
            </div>
            <div className="admin-modal-actions">
              <button className="admin-btn admin-btn-ghost" onClick={() => setShowEditor(false)} disabled={saving}>取消</button>
              <button className="admin-btn admin-btn-primary" onClick={() => void handleSave()} disabled={saving}>{saving ? '保存中...' : '保存'}</button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

function Field({ label, children }: { label: string; children: React.ReactNode }) {
  return <div className="admin-form-group"><label className="admin-form-label">{label}</label>{children}</div>
}
