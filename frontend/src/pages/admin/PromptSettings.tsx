import React, { useEffect, useState } from 'react'
import { adminPromptsAPI } from '../../lib/adminApi'

interface PromptRecord {
  id: string
  module: string
  content: string
  description: string
  updated_at: string
}

// 模块分类配置
const KB_MODULES = [
  { module: 'kb_shishen', label: '十神断事口诀', icon: '📖', hint: '十神的标准定义与断事逻辑，修改后将影响所有 AI 批断的十神解读方式。' },
  { module: 'kb_gejv',    label: '格局判断规则', icon: '🏛️', hint: '子平真诠格局定义、成格破格条件，是推断命局高下的核心逻辑层。' },
  { module: 'kb_tiaohou', label: '调候用神表',   icon: '🌡️', hint: '穷通宝鉴按月调候精华，使 AI 判断寒暖燥湿时有典籍依据。' },
  { module: 'kb_yingqi',  label: '流年应期推算', icon: '📅', hint: '冲合刑害应期推算规则，使 AI 能准确定位吉凶发生的月份。' },
  { module: 'kb_tonality', label: '语调与立场', icon: '⚖️', hint: '控制 AI 分析的语气风格——中立理性 vs 温暖积极，修改后影响所有报告的文风。' },
]

const INSTRUCTION_MODULES = [
  { module: 'liunian', label: '流年运势批断', icon: '⚡', hint: '流年精批的 User Prompt 模版，支持 Go Template 占位变量注入命盘数据。' },
  { module: 'compatibility', label: '婚恋合盘解读', icon: '💞', hint: '双人合盘分析的 User Prompt 模版，输入双方命盘摘要、四维分数与结构化证据。' },
]

const PromptSettings: React.FC = () => {
  const [prompts, setPrompts] = useState<PromptRecord[]>([])
  const [loading, setLoading] = useState(false)
  const [activeTab, setActiveTab] = useState<'kb' | 'instruction'>('kb')
  const [editingModule, setEditingModule] = useState<string | null>(null)
  const [editContent, setEditContent] = useState('')
  const [saving, setSaving] = useState(false)

  const fetchPrompts = async () => {
    setLoading(true)
    try {
      const { data } = await adminPromptsAPI.list()
      setPrompts(data || [])
    } catch (e: any) {
      alert(e.message || '获取配置失败')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { fetchPrompts() }, [])

  const handleEdit = (p: PromptRecord) => {
    setEditingModule(p.module)
    setEditContent(p.content)
  }

  const handleSave = async () => {
    if (!editingModule) return
    setSaving(true)
    try {
      await adminPromptsAPI.update(editingModule, { content: editContent })
      alert('保存成功')
      setEditingModule(null)
      fetchPrompts()
    } catch (e: any) {
      alert(e.message || '保存失败')
    } finally {
      setSaving(false)
    }
  }

  const getPrompt = (module: string) => prompts.find(p => p.module === module)

  const renderModuleCard = (
    def: { module: string; label: string; icon: string; hint: string },
    isKb: boolean
  ) => {
    const p = getPrompt(def.module)
    const isEditing = editingModule === def.module

    return (
      <div key={def.module} style={{
        background: 'var(--color-bg-secondary)',
        borderRadius: 12,
        border: isEditing
          ? '1px solid var(--color-primary)'
          : '1px solid var(--color-border)',
        overflow: 'hidden',
        marginBottom: 20,
        transition: 'border-color 0.2s',
      }}>
        {/* Header */}
        <div style={{
          padding: '16px 20px',
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'flex-start',
          borderBottom: '1px solid var(--color-border)',
          background: isKb
            ? 'rgba(167, 139, 250, 0.04)'
            : 'rgba(251, 191, 36, 0.04)',
        }}>
          <div>
            <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 4 }}>
              <span style={{ fontSize: 18 }}>{def.icon}</span>
              <h3 style={{ margin: 0, color: 'var(--color-text-primary)', fontSize: 15 }}>
                {def.label}
              </h3>
              <span style={{
                fontSize: 11, color: '#888',
                background: '#333', padding: '1px 6px', borderRadius: 4,
                fontFamily: 'monospace',
              }}>
                {def.module}
              </span>
            </div>
            <p style={{ margin: 0, color: 'var(--color-text-secondary)', fontSize: 13 }}>
              {def.hint}
            </p>
          </div>
          {p && !isEditing && (
            <button
              onClick={() => handleEdit(p)}
              style={{
                marginLeft: 16, flexShrink: 0,
                padding: '6px 14px', borderRadius: 6, fontSize: 13,
                background: 'transparent',
                border: '1px solid var(--color-border)',
                color: 'var(--color-text-secondary)',
                cursor: 'pointer',
              }}
            >
              编辑
            </button>
          )}
          {!p && !loading && (
            <span style={{ color: '#ff6b6b', fontSize: 12, flexShrink: 0 }}>⚠ 未初始化（请重启后端）</span>
          )}
        </div>

        {/* Body */}
        <div style={{ padding: '16px 20px' }}>
          {!p ? (
            <div style={{ color: '#666', fontSize: 13 }}>暂无数据</div>
          ) : isEditing ? (
            <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
              {/* Variables hint for liunian */}
              {def.module === 'liunian' && (
                <div style={{
                  fontSize: 12, color: 'var(--color-text-secondary)',
                  background: 'rgba(255,255,255,0.03)',
                  padding: '10px 14px', borderRadius: 8,
                  border: '1px solid var(--color-border)',
                }}>
                  <strong>💡 可用 Go Template 占位变量：</strong>
                  <code style={{ display: 'block', marginTop: 6, lineHeight: 1.8 }}>
                    {'{{.NatalAnalysisLogic}}'} / {'{{.CurrentDayunGanZhi}}'} / {'{{.CurrentDayunGanShiShen}}'} / {'{{.CurrentDayunZhiShiShen}}'}<br />
                    {'{{.TargetYear}}'} / {'{{.TargetYearGanZhi}}'} / {'{{.TargetYearGanShiShen}}'} / {'{{.TargetYearZhiShiShen}}'}
                  </code>
                </div>
              )}
              {isKb && (
                <div style={{
                  fontSize: 12, color: '#a78bfa',
                  background: 'rgba(167,139,250,0.06)',
                  padding: '10px 14px', borderRadius: 8,
                  border: '1px solid rgba(167,139,250,0.2)',
                }}>
                  📚 这是命理知识库模块，修改后将作为 <strong>System Prompt</strong> 注入到所有流年精批请求中。
                </div>
              )}
              <textarea
                value={editContent}
                onChange={e => setEditContent(e.target.value)}
                rows={18}
                style={{
                  width: '100%', fontFamily: 'monospace', fontSize: 13,
                  padding: '12px', background: '#1a1a2e',
                  color: '#d4d4d4', border: '1px solid #333',
                  borderRadius: 8, resize: 'vertical', boxSizing: 'border-box',
                  lineHeight: 1.6,
                }}
              />
              <div style={{ display: 'flex', gap: 10, justifyContent: 'flex-end' }}>
                <button
                  className="secondary"
                  onClick={() => setEditingModule(null)}
                  style={{ padding: '8px 16px', borderRadius: 6, cursor: 'pointer' }}
                >
                  取消
                </button>
                <button
                  onClick={handleSave}
                  disabled={saving}
                  style={{ padding: '8px 20px', borderRadius: 6, cursor: 'pointer' }}
                >
                  {saving ? '保存中...' : '保存更改'}
                </button>
              </div>
            </div>
          ) : (
            <pre style={{
              background: '#111', color: '#aaa',
              padding: '14px 16px', borderRadius: 8,
              fontSize: 12, whiteSpace: 'pre-wrap',
              overflowX: 'auto', maxHeight: 240,
              lineHeight: 1.7, margin: 0,
            }}>
              {p.content}
            </pre>
          )}
        </div>
      </div>
    )
  }

  return (
    <div>
      <div style={{ marginBottom: 28 }}>
        <h2 style={{ color: 'var(--color-primary)', margin: '0 0 6px 0' }}>
          🤖 AI 指令设定 (Prompts)
        </h2>
        <p style={{ color: 'var(--color-text-secondary)', margin: 0, fontSize: 14 }}>
          动态配置 AI 命理批断所使用的典籍知识库（System Prompt）与批断指令模版（User Prompt）。
        </p>
      </div>

      {/* Tab 导航 */}
      <div style={{ display: 'flex', gap: 4, marginBottom: 24, borderBottom: '1px solid var(--color-border)' }}>
        {[
          { key: 'kb' as const, label: '📚 命理知识库', desc: '5 个典籍模块' },
          { key: 'instruction' as const, label: '⚡ 批断指令', desc: '流年等模版' },
        ].map(tab => (
          <button
            key={tab.key}
            onClick={() => { setActiveTab(tab.key); setEditingModule(null) }}
            style={{
              padding: '10px 20px',
              borderRadius: '8px 8px 0 0',
              border: 'none',
              borderBottom: activeTab === tab.key
                ? '2px solid var(--color-primary)'
                : '2px solid transparent',
              background: activeTab === tab.key
                ? 'rgba(167,139,250,0.08)'
                : 'transparent',
              color: activeTab === tab.key
                ? 'var(--color-primary)'
                : 'var(--color-text-secondary)',
              cursor: 'pointer',
              fontSize: 14,
              fontWeight: activeTab === tab.key ? 600 : 400,
              transition: 'all 0.2s',
            }}
          >
            {tab.label}
            <span style={{ fontSize: 11, marginLeft: 6, opacity: 0.7 }}>({tab.desc})</span>
          </button>
        ))}
      </div>

      {loading ? (
        <div style={{ textAlign: 'center', padding: 40, color: 'var(--color-text-secondary)' }}>
          加载中...
        </div>
      ) : (
        <div>
          {activeTab === 'kb' && (
            <div>
              <div style={{
                background: 'rgba(167,139,250,0.06)', border: '1px solid rgba(167,139,250,0.2)',
                borderRadius: 10, padding: '12px 16px', marginBottom: 24,
                fontSize: 13, color: '#a78bfa', lineHeight: 1.6,
              }}>
                💡 <strong>知识库工作原理：</strong>
                以下 5 个模块会在每次 AI 流年精批时自动被拼入 System Prompt，让 AI 先"精通典籍"再对命盘进行批断。
                你可以随时修改并保存，清除旧流年缓存后立即生效。
              </div>
              {KB_MODULES.map(def => renderModuleCard(def, true))}
            </div>
          )}

          {activeTab === 'instruction' && (
            <div>
              <div style={{
                background: 'rgba(251,191,36,0.05)', border: '1px solid rgba(251,191,36,0.2)',
                borderRadius: 10, padding: '12px 16px', marginBottom: 24,
                fontSize: 13, color: '#fbbf24', lineHeight: 1.6,
              }}>
                💡 <strong>批断指令说明：</strong>
                此处为 User Prompt 模版，支持 Go Template 语法注入命盘结构化数据。
                AI 会在"已掌握知识库"的基础上，按此模版的要求执行具体批断并输出 JSON。
              </div>
              {INSTRUCTION_MODULES.map(def => renderModuleCard(def, false))}
            </div>
          )}
        </div>
      )}
    </div>
  )
}

export default PromptSettings
