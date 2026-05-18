import React, { useEffect, useState } from 'react'
import { adminCleanupConfigAPI, type CleanupConfig } from '../../lib/adminApi'

function errorMessage(e: unknown, fallback: string) {
  return e instanceof Error ? e.message : fallback
}

const CleanupConfigPage: React.FC = () => {
  const [cfg, setCfg] = useState<CleanupConfig | null>(null)
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [enabled, setEnabled] = useState(true)
  const [retentionDays, setRetentionDays] = useState(90)
  const [runHour, setRunHour] = useState(3)

  useEffect(() => {
    let cancelled = false
    setLoading(true)
    adminCleanupConfigAPI.get()
      .then(({ data }) => {
        if (cancelled) return
        setCfg(data)
        setEnabled(data.enabled)
        setRetentionDays(data.retention_days)
        setRunHour(data.run_hour)
      })
      .catch(e => { if (!cancelled) alert(errorMessage(e, '加载清理配置失败')) })
      .finally(() => { if (!cancelled) setLoading(false) })
    return () => { cancelled = true }
  }, [])

  const dirty = !!cfg && (
    enabled !== cfg.enabled ||
    retentionDays !== cfg.retention_days ||
    runHour !== cfg.run_hour
  )

  const handleSave = async () => {
    setSaving(true)
    try {
      const { data } = await adminCleanupConfigAPI.update({
        enabled,
        retention_days: retentionDays,
        run_hour: runHour,
      })
      setCfg(data)
      setEnabled(data.enabled)
      setRetentionDays(data.retention_days)
      setRunHour(data.run_hour)
      alert('已保存。下一次 scheduler tick 起按新配置执行。')
    } catch (e: unknown) {
      alert(errorMessage(e, '保存失败'))
    } finally {
      setSaving(false)
    }
  }

  if (loading) {
    return <div className="admin-page"><h1>数据清理配置</h1><p>加载中…</p></div>
  }

  return (
    <div className="admin-page">
      <h1>数据清理配置</h1>
      <p className="admin-hint">
        Backend 每天在「执行时刻」运行一次清理任务：
        删除 6 张 AI 缓存表 + ai_requests_log 中超过保留天数的行，
        并把 token_usage_logs 已闭合月份汇总到 token_usage_logs_monthly 后删源行。
      </p>

      <div className="admin-form-group">
        <label className="admin-form-label">
          <input
            type="checkbox"
            checked={enabled}
            onChange={e => setEnabled(e.target.checked)}
          />
          <span style={{ marginLeft: 8 }}>启用自动清理任务</span>
        </label>
        <div className="admin-form-hint">关闭后 scheduler 仍按时唤醒但直接返回，不删任何数据。</div>
      </div>

      <div className="admin-form-group">
        <label className="admin-form-label" htmlFor="retention_days">保留天数</label>
        <input
          id="retention_days"
          type="number"
          min={1}
          max={3650}
          value={retentionDays}
          onChange={e => setRetentionDays(Number(e.target.value) || 0)}
          className="admin-form-input"
          style={{ width: 120 }}
        />
        <div className="admin-form-hint">
          AI 缓存表与 ai_requests_log 超过该天数的行会被删除；token_usage_logs 不受此值影响（按月汇总）。
          合法区间 [1, 3650]，越界自动 clamp。
        </div>
      </div>

      <div className="admin-form-group">
        <label className="admin-form-label" htmlFor="run_hour">执行时刻（小时）</label>
        <select
          id="run_hour"
          value={runHour}
          onChange={e => setRunHour(Number(e.target.value))}
          className="admin-form-input"
          style={{ width: 120 }}
        >
          {Array.from({ length: 24 }, (_, h) => (
            <option key={h} value={h}>{h.toString().padStart(2, '0')}:00</option>
          ))}
        </select>
        <div className="admin-form-hint">每日固定时刻运行（24h 制）。改完保存，下一次 tick 起按新时刻调度。</div>
      </div>

      <div style={{ marginTop: 24 }}>
        <button
          className="btn btn-primary"
          disabled={!dirty || saving}
          onClick={handleSave}
        >
          {saving ? '保存中…' : '保存'}
        </button>
        {!dirty && cfg && <span className="admin-form-hint" style={{ marginLeft: 12 }}>无改动</span>}
      </div>
    </div>
  )
}

export default CleanupConfigPage
