import { useState } from 'react'
import type { PolishedReport, StructuredReport } from '../lib/api'
import { cleanReportText, splitParagraphs } from '../lib/reportText'
import './PolishedPanel.css'

interface Props {
  polishedReport: PolishedReport | null
  hasOriginalReport: boolean
  loading: boolean
  errorMsg: string | null
  onSubmit: (userSituation: string) => Promise<void>
}

const MIN_LEN = 20
const MAX_LEN = 300

function paragraphsOf(text: string | undefined | null): string[] {
  return splitParagraphs(text)
}

function chaptersFrom(structured: StructuredReport | null | undefined): { title: string; detail: string }[] {
  return structured?.chapters?.map(c => ({ title: c.title, detail: c.detail || c.brief })) ?? []
}

export default function PolishedPanel({ polishedReport, hasOriginalReport, loading, errorMsg, onSubmit }: Props) {
  const [editing, setEditing] = useState<boolean>(!polishedReport)
  const [input, setInput] = useState<string>(polishedReport?.user_situation ?? '')

  if (!hasOriginalReport) {
    return (
      <div className="polished-panel">
        <div className="polished-empty-state">
          <p className="polished-empty-tip">请先生成「原版」命理解读，再尝试润色版。</p>
        </div>
      </div>
    )
  }

  const inputLen = input.trim().length
  const canSubmit = inputLen >= MIN_LEN && inputLen <= MAX_LEN && !loading

  const submit = async () => {
    if (!canSubmit) return
    await onSubmit(input.trim())
    setEditing(false)
  }

  // 已有润色版且不在编辑：展示态
  if (polishedReport && !editing) {
    const chapters = chaptersFrom(polishedReport.content_structured)
    return (
      <div className="polished-panel">
        <div className="polished-content-area">
          <div className="polished-context-bar">
            <span className="polished-context-label">你的情况描述：</span>
            <span className="polished-context-text">"{polishedReport.user_situation}"</span>
            <button className="polished-context-edit" onClick={() => setEditing(true)}>
              修改 / 重新润色
            </button>
          </div>
          {chapters.length === 0 ? (
            <p className="polished-empty-tip">润色版解析为空，请重新润色。</p>
          ) : (
            <div className="polished-chapter-list">
              {chapters.map((ch, i) => {
                const paras = paragraphsOf(ch.detail)
                return (
                  <div key={i} className="polished-chapter">
                    <h3 className="polished-chapter-title serif">【{cleanReportText(ch.title)}】</h3>
                    {paras.length > 0
                      ? paras.map((p, j) => <p key={j} className="polished-chapter-body">{p}</p>)
                      : <p className="polished-chapter-body">{cleanReportText(ch.detail)}</p>}
                  </div>
                )
              })}
            </div>
          )}
        </div>
      </div>
    )
  }

  // 空态 / 编辑态：输入区
  return (
    <div className="polished-panel">
      <div className="polished-input-area">
        <p className="polished-input-hint">
          简单说说你最近最关心的情况，AI 师傅会基于这个写一份贴你处境的润色版。
        </p>
        <textarea
          className="polished-input"
          value={input}
          onChange={(e) => setInput(e.target.value)}
          maxLength={300}
          placeholder="例：今年在考虑跳槽，想做创意类工作；跟对象因发展节奏不一有点摩擦..."
          rows={5}
        />
        <div className="polished-input-meta">
          <span className={inputLen < MIN_LEN ? 'polished-len-warn' : 'polished-len-ok'}>
            {inputLen} / {MAX_LEN} 字（至少 {MIN_LEN} 字）
          </span>
          <div className="polished-input-actions">
            {polishedReport && (
              <button className="polished-btn-ghost" onClick={() => { setEditing(false); setInput(polishedReport.user_situation) }}>
                取消
              </button>
            )}
            <button className="polished-btn-primary" disabled={!canSubmit} onClick={submit}>
              {loading ? '润色中…' : (polishedReport ? '重新润色' : '生成润色版')}
            </button>
          </div>
        </div>
        {errorMsg && <p className="polished-error">{errorMsg}</p>}
      </div>
    </div>
  )
}
