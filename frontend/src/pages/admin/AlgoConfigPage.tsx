import React, { useEffect, useState } from 'react'
import { adminAlgoConfigAPI, adminAlgoTiaohouAPI } from '../../lib/adminApi'

interface AlgoConfigItem {
  key: string
  value: string
  description: string
  source: 'db' | 'default'
}

interface TiaohouRow {
  DayGan: string
  MonthZhi: string
  XiElements: string
  Text: string
}

const PARAM_LABELS: Record<string, string> = {
  jixiong_jiHan_min: '极寒阈值（寒性元素最低数量）',
  jixiong_jiRe_min: '极热阈值（暖性元素最低数量）',
  jixiong_shenQiang_pct: '身强判定阈值（生助比例 %）',
}

const TIAN_GAN = ['甲', '乙', '丙', '丁', '戊', '己', '庚', '辛', '壬', '癸']

const AlgoConfigPage: React.FC = () => {
  const [params, setParams] = useState<AlgoConfigItem[]>([])
  const [editingKey, setEditingKey] = useState<string | null>(null)
  const [editValue, setEditValue] = useState('')
  const [savingKey, setSavingKey] = useState<string | null>(null)
  const [reloading, setReloading] = useState(false)
  const [loadingParams, setLoadingParams] = useState(false)

  // 调候规则
  const [activeGan, setActiveGan] = useState<string>('甲')
  const [tiaohouRows, setTiaohouRows] = useState<TiaohouRow[]>([])
  const [loadingTiaohou, setLoadingTiaohou] = useState(false)
  const [editingTiaohou, setEditingTiaohou] = useState<string | null>(null) // key = "甲_子"
  const [editXi, setEditXi] = useState('')
  const [editText, setEditText] = useState('')
  const [savingTiaohou, setSavingTiaohou] = useState(false)

  const fetchParams = async () => {
    setLoadingParams(true)
    try {
      const { data } = await adminAlgoConfigAPI.list()
      setParams(data || [])
    } catch (e: any) {
      alert(e.message || '获取参数失败')
    } finally {
      setLoadingParams(false)
    }
  }

  const fetchTiaohou = async (dayGan: string) => {
    setLoadingTiaohou(true)
    try {
      const { data } = await adminAlgoTiaohouAPI.list(dayGan)
      setTiaohouRows(data || [])
    } catch (e: any) {
      alert(e.message || '获取调候规则失败')
    } finally {
      setLoadingTiaohou(false)
    }
  }

  useEffect(() => { fetchParams() }, [])
  useEffect(() => { fetchTiaohou(activeGan) }, [activeGan])

  const handleSaveParam = async (key: string) => {
    setSavingKey(key)
    try {
      await adminAlgoConfigAPI.update(key, { value: editValue })
      setEditingKey(null)
      fetchParams()
    } catch (e: any) {
      alert(e.message || '保存失败')
    } finally {
      setSavingKey(null)
    }
  }

  const handleReload = async () => {
    setReloading(true)
    try {
      await adminAlgoConfigAPI.reload()
      alert('算法配置已重载')
    } catch (e: any) {
      alert(e.message || '重载失败')
    } finally {
      setReloading(false)
    }
  }

  const handleSaveTiaohou = async (row: TiaohouRow) => {
    setSavingTiaohou(true)
    try {
      await adminAlgoTiaohouAPI.update(row.DayGan, row.MonthZhi, {
        xi_elements: editXi,
        text: editText,
      })
      setEditingTiaohou(null)
      fetchTiaohou(activeGan)
    } catch (e: any) {
      alert(e.message || '保存失败')
    } finally {
      setSavingTiaohou(false)
    }
  }

  return (
    <div style={{ padding: '24px', maxWidth: '900px' }}>
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: '24px' }}>
        <h2 style={{ margin: 0, fontSize: '20px', fontWeight: 600 }}>算法参数配置</h2>
        <button
          onClick={handleReload}
          disabled={reloading}
          style={{
            padding: '8px 16px',
            background: 'var(--color-primary, #7c6f64)',
            color: '#fff',
            border: 'none',
            borderRadius: '6px',
            cursor: reloading ? 'not-allowed' : 'pointer',
            opacity: reloading ? 0.6 : 1,
          }}
        >
          {reloading ? '重载中...' : '重载缓存'}
        </button>
      </div>

      {/* 算法参数列表 */}
      <section style={{ marginBottom: '40px' }}>
        <h3 style={{ fontSize: '16px', fontWeight: 600, marginBottom: '12px' }}>吉凶判定参数</h3>
        {loadingParams ? (
          <p style={{ color: 'var(--color-muted, #888)' }}>加载中...</p>
        ) : (
          <table style={{ width: '100%', borderCollapse: 'collapse', fontSize: '14px' }}>
            <thead>
              <tr style={{ background: 'var(--color-surface, #f5f0eb)', textAlign: 'left' }}>
                <th style={{ padding: '10px 12px', fontWeight: 600 }}>参数</th>
                <th style={{ padding: '10px 12px', fontWeight: 600 }}>当前值</th>
                <th style={{ padding: '10px 12px', fontWeight: 600 }}>来源</th>
                <th style={{ padding: '10px 12px', fontWeight: 600 }}>操作</th>
              </tr>
            </thead>
            <tbody>
              {params.map((p) => (
                <tr key={p.key} style={{ borderBottom: '1px solid var(--color-border, #e8e0d8)' }}>
                  <td style={{ padding: '10px 12px' }}>
                    <div style={{ fontWeight: 500 }}>{PARAM_LABELS[p.key] || p.key}</div>
                    <div style={{ fontSize: '12px', color: 'var(--color-muted, #888)', marginTop: '2px' }}>{p.key}</div>
                  </td>
                  <td style={{ padding: '10px 12px' }}>
                    {editingKey === p.key ? (
                      <input
                        value={editValue}
                        onChange={(e) => setEditValue(e.target.value)}
                        style={{ padding: '4px 8px', border: '1px solid var(--color-border, #ccc)', borderRadius: '4px', width: '80px' }}
                        autoFocus
                      />
                    ) : (
                      <span>{p.value}</span>
                    )}
                  </td>
                  <td style={{ padding: '10px 12px' }}>
                    <span style={{
                      padding: '2px 8px',
                      borderRadius: '12px',
                      fontSize: '12px',
                      background: p.source === 'db' ? '#e8f5e9' : '#f5f5f5',
                      color: p.source === 'db' ? '#2e7d32' : '#666',
                    }}>
                      {p.source === 'db' ? '数据库' : '默认值'}
                    </span>
                  </td>
                  <td style={{ padding: '10px 12px' }}>
                    {editingKey === p.key ? (
                      <span style={{ display: 'flex', gap: '8px' }}>
                        <button
                          onClick={() => handleSaveParam(p.key)}
                          disabled={savingKey === p.key}
                          style={{ padding: '4px 10px', background: '#4caf50', color: '#fff', border: 'none', borderRadius: '4px', cursor: 'pointer' }}
                        >
                          {savingKey === p.key ? '...' : '保存'}
                        </button>
                        <button
                          onClick={() => setEditingKey(null)}
                          style={{ padding: '4px 10px', background: '#e0e0e0', border: 'none', borderRadius: '4px', cursor: 'pointer' }}
                        >
                          取消
                        </button>
                      </span>
                    ) : (
                      <button
                        onClick={() => { setEditingKey(p.key); setEditValue(p.value) }}
                        style={{ padding: '4px 10px', background: 'transparent', border: '1px solid var(--color-border, #ccc)', borderRadius: '4px', cursor: 'pointer' }}
                      >
                        编辑
                      </button>
                    )}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </section>

      {/* 调候用神规则 */}
      <section>
        <h3 style={{ fontSize: '16px', fontWeight: 600, marginBottom: '12px' }}>调候用神规则</h3>
        <div style={{ display: 'flex', gap: '8px', marginBottom: '16px', flexWrap: 'wrap' }}>
          {TIAN_GAN.map((gan) => (
            <button
              key={gan}
              onClick={() => { setActiveGan(gan); setEditingTiaohou(null) }}
              style={{
                padding: '6px 14px',
                border: '1px solid var(--color-border, #ccc)',
                borderRadius: '6px',
                cursor: 'pointer',
                background: activeGan === gan ? 'var(--color-primary, #7c6f64)' : 'transparent',
                color: activeGan === gan ? '#fff' : 'inherit',
                fontWeight: activeGan === gan ? 600 : 400,
              }}
            >
              {gan}
            </button>
          ))}
        </div>

        {loadingTiaohou ? (
          <p style={{ color: 'var(--color-muted, #888)' }}>加载中...</p>
        ) : (
          <table style={{ width: '100%', borderCollapse: 'collapse', fontSize: '14px' }}>
            <thead>
              <tr style={{ background: 'var(--color-surface, #f5f0eb)', textAlign: 'left' }}>
                <th style={{ padding: '10px 12px', fontWeight: 600 }}>月支</th>
                <th style={{ padding: '10px 12px', fontWeight: 600 }}>喜用天干</th>
                <th style={{ padding: '10px 12px', fontWeight: 600 }}>原文释义</th>
                <th style={{ padding: '10px 12px', fontWeight: 600 }}>操作</th>
              </tr>
            </thead>
            <tbody>
              {tiaohouRows.map((row) => {
                const rowKey = `${row.DayGan}_${row.MonthZhi}`
                const isEditing = editingTiaohou === rowKey
                return (
                  <tr key={rowKey} style={{ borderBottom: '1px solid var(--color-border, #e8e0d8)', verticalAlign: 'top' }}>
                    <td style={{ padding: '10px 12px', fontWeight: 500 }}>{row.MonthZhi}</td>
                    <td style={{ padding: '10px 12px' }}>
                      {isEditing ? (
                        <input
                          value={editXi}
                          onChange={(e) => setEditXi(e.target.value)}
                          placeholder="如: 丙,癸"
                          style={{ padding: '4px 8px', border: '1px solid var(--color-border, #ccc)', borderRadius: '4px', width: '100px' }}
                        />
                      ) : (
                        <span>{row.XiElements}</span>
                      )}
                    </td>
                    <td style={{ padding: '10px 12px', maxWidth: '300px' }}>
                      {isEditing ? (
                        <textarea
                          value={editText}
                          onChange={(e) => setEditText(e.target.value)}
                          rows={3}
                          style={{ padding: '4px 8px', border: '1px solid var(--color-border, #ccc)', borderRadius: '4px', width: '100%', resize: 'vertical' }}
                        />
                      ) : (
                        <span style={{ whiteSpace: 'pre-wrap', fontSize: '13px', color: 'var(--color-muted, #666)' }}>{row.Text || '-'}</span>
                      )}
                    </td>
                    <td style={{ padding: '10px 12px' }}>
                      {isEditing ? (
                        <span style={{ display: 'flex', flexDirection: 'column', gap: '6px' }}>
                          <button
                            onClick={() => handleSaveTiaohou(row)}
                            disabled={savingTiaohou}
                            style={{ padding: '4px 10px', background: '#4caf50', color: '#fff', border: 'none', borderRadius: '4px', cursor: 'pointer' }}
                          >
                            {savingTiaohou ? '...' : '保存'}
                          </button>
                          <button
                            onClick={() => setEditingTiaohou(null)}
                            style={{ padding: '4px 10px', background: '#e0e0e0', border: 'none', borderRadius: '4px', cursor: 'pointer' }}
                          >
                            取消
                          </button>
                        </span>
                      ) : (
                        <button
                          onClick={() => { setEditingTiaohou(rowKey); setEditXi(row.XiElements); setEditText(row.Text) }}
                          style={{ padding: '4px 10px', background: 'transparent', border: '1px solid var(--color-border, #ccc)', borderRadius: '4px', cursor: 'pointer' }}
                        >
                          编辑
                        </button>
                      )}
                    </td>
                  </tr>
                )
              })}
            </tbody>
          </table>
        )}
      </section>
    </div>
  )
}

export default AlgoConfigPage
