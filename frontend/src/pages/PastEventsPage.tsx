import { useCallback, useEffect, useRef, useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { ChevronDown, ChevronLeft, Download, Loader2 } from 'lucide-react'
import { baziAPI, brandAPI } from '../lib/api'
import type { ExportBrand } from '../lib/api'
import { useAuth } from '../contexts/AuthContext'
import { Button } from '../components/ui/Button'
import { EmptyState } from '../components/ui/EmptyState'
import PastEventsPrintLayout from '../components/PastEventsPrintLayout'
import { useToast } from '../components/ui/useToast'
import {
  buildPastEventsExportSegments,
  countPastEventsProgress,
  getPastEventsCurrentPeriod,
  getPastEventsFutureToggleLabel,
  getPastEventsSegmentViewState,
  getPastEventsSignalLabel,
  getPastEventsYearNarrativeState,
  type PastEventsProgressCounts,
  type PastEventsSegmentViewState,
} from '../lib/pastEventsViewModel'
import './PastEventsPage.css'

interface YearEvent {
  year: number
  age: number
  gan_zhi: string
  dayun_gan_zhi: string
  dayun_index: number
  year_in_dayun?: number
  dayun_phase?: string
  ten_god_power?: TenGodPower
  signals: string[]
  narrative: string
  evidence_summary?: string[]
}

interface TenGodPower {
  dominant: string
  group: string
  group_label: string
  strength: string
  polarity: string
  plain_title: string
  plain_text: string
  score: number
  reason?: string
}

interface DayunMeta {
  index: number
  gan_zhi: string
  start_age: number
  end_age: number
  start_year: number
  end_year: number
  ten_god_power?: TenGodPower
}

interface YearNarrativeEntry {
  year: number
  ganzhi: string
  narrative: string
}

interface DayunSummary {
  themes: string[]
  summary: string
  years?: YearNarrativeEntry[]
  cached?: boolean
  loading?: boolean
  status?: 'loading' | 'interrupted'
  error?: string
  generation?: DayunGenerationMeta
  // 未来段折叠态：true 表示该段在页面打开时不自动生成批语
  // 用户点击 [展开 ▼] 后变 false（仍未生成批语），再点"生成本段"才触发
  folded?: boolean
}

type DayunGenerationSource = 'initial' | 'manual' | 'recovery'

interface DayunGenerationMeta {
  startedAt: number
  attempt: number
  source: DayunGenerationSource
  requestId: number
}

const DAYUN_GENERATION_STALE_MS = 10_000
const DAYUN_GENERATION_MAX_AUTO_RETRIES = 1
const DAYUN_GENERATION_INTERRUPTED_COPY = '生成中断，点击重试'
const DAYUN_GENERATION_FAILED_COPY = '生成失败，请重试'

const WUXING_GAN: Record<string, string> = {
  '甲': 'mu', '乙': 'mu', '丙': 'huo', '丁': 'huo',
  '戊': 'tu', '己': 'tu', '庚': 'jin', '辛': 'jin',
  '壬': 'shui', '癸': 'shui',
}

const currentYear = new Date().getFullYear()

function SegmentStateBadge({ state }: { state: PastEventsSegmentViewState }) {
  const stateClass = state.state === 'future_folded' || state.state === 'future_expanded_ungenerated'
    ? 'future'
    : state.state
  return (
    <span className={`past-events-state-badge past-events-state-badge--${stateClass}`}>
      {state.label}
    </span>
  )
}

function PastEventsStatusPanel({
  totalYears,
  progressCounts,
  currentGroup,
  currentYearEvent,
}: {
  totalYears: number
  progressCounts: PastEventsProgressCounts
  currentGroup?: { meta: DayunMeta }
  currentYearEvent?: YearEvent
}) {
  return (
    <div className="past-events-status-panel">
      <div className="past-events-status-panel__top">
        <div>
          <div className="past-events-status-panel__title">
            年份信号已就绪，大运批语分段生成
          </div>
          <div className="past-events-status-panel__copy">
            当前年份 {currentYear} 年
            {currentGroup ? ` · 当前大运 ${currentGroup.meta.gan_zhi}（${currentGroup.meta.start_age}-${currentGroup.meta.end_age}岁）` : ''}
            {currentYearEvent ? ` · ${currentYearEvent.age}岁` : ''}
          </div>
        </div>
        {currentGroup && (
          <a className="past-events-status-panel__jump" href="#past-events-current-dayun">
            跳到当前大运
          </a>
        )}
      </div>
      <div className="past-events-status-panel__metrics" aria-label="过往推算状态">
        <span className="past-events-status-metric">
          <span>算法年份</span>
          <strong>{totalYears}</strong>
        </span>
        <span className="past-events-status-metric">
          <span>批语完成</span>
          <strong>{progressCounts.generated}</strong>
        </span>
        <span className="past-events-status-metric">
          <span>生成中</span>
          <strong>{progressCounts.generating}</strong>
        </span>
        <span className="past-events-status-metric">
          <span>可重试</span>
          <strong>{progressCounts.interrupted}</strong>
        </span>
        <span className="past-events-status-metric">
          <span>未来待生成</span>
          <strong>{progressCounts.futurePending}</strong>
        </span>
      </div>
    </div>
  )
}

export default function PastEventsPage() {
  const { chartId } = useParams<{ chartId: string }>()
  const navigate = useNavigate()
  const { user, isLoading } = useAuth()
  const { showToast } = useToast()
  const [yearsLoaded, setYearsLoaded] = useState(false)
  const [yearsError, setYearsError] = useState('')
  const [events, setEvents] = useState<YearEvent[]>([])
  const [dayunMeta, setDayunMeta] = useState<DayunMeta[]>([])
  const [summaries, setSummaries] = useState<Record<number, DayunSummary>>({})
  const [includeEvidenceInExport, setIncludeEvidenceInExport] = useState(false)
  const [exportingPDF, setExportingPDF] = useState(false)
  const [brand, setBrand] = useState<ExportBrand | null>(null)
  const [expandedEvidence, setExpandedEvidence] = useState<Record<string, boolean>>({})
  const [streamDone, setStreamDone] = useState(false)
  const [streamError, setStreamError] = useState('')
  const inflightRef = useRef(false)
  const summariesRef = useRef<Record<number, DayunSummary>>({})
  const generationRequestSeqRef = useRef(0)

  useEffect(() => {
    summariesRef.current = summaries
  }, [summaries])

  const beginDayunGeneration = useCallback((dayunIndex: number, source: DayunGenerationSource) => {
    const existing = summariesRef.current[dayunIndex]
    const previousAttempt = existing?.generation?.attempt ?? 0
    const attempt = source === 'recovery' ? previousAttempt + 1 : 0
    const requestId = ++generationRequestSeqRef.current
    const nextSummary: DayunSummary = {
      ...existing,
      themes: existing?.themes || [],
      summary: existing?.summary || '',
      loading: true,
      status: 'loading',
      error: undefined,
      folded: false,
      generation: {
        startedAt: Date.now(),
        attempt,
        source,
        requestId,
      },
    }

    summariesRef.current = {
      ...summariesRef.current,
      [dayunIndex]: nextSummary,
    }
    setSummaries((prev) => ({
      ...prev,
      [dayunIndex]: nextSummary,
    }))
    if (source !== 'initial') {
      setStreamError('')
      setStreamDone(false)
    }
    return requestId
  }, [])

  const markDayunInterrupted = useCallback((dayunIndex: number, message = DAYUN_GENERATION_INTERRUPTED_COPY) => {
    const existing = summariesRef.current[dayunIndex]
    const nextSummary: DayunSummary = {
      ...existing,
      themes: existing?.themes || [],
      summary: existing?.summary || '',
      loading: false,
      status: 'interrupted',
      error: message,
      folded: false,
      generation: undefined,
    }
    summariesRef.current = {
      ...summariesRef.current,
      [dayunIndex]: nextSummary,
    }
    setSummaries((prev) => ({
      ...prev,
      [dayunIndex]: nextSummary,
    }))
  }, [])

  const markLoadingDayunsInterrupted = useCallback((message = DAYUN_GENERATION_INTERRUPTED_COPY, source?: DayunGenerationSource) => {
    const next = { ...summariesRef.current }
    for (const [key, summary] of Object.entries(summariesRef.current)) {
      if (summary.loading && (!source || summary.generation?.source === source)) {
        next[Number(key)] = {
          ...summary,
          loading: false,
          status: 'interrupted',
          error: message,
          folded: false,
          generation: undefined,
        }
      }
    }
    summariesRef.current = next
    setSummaries(next)
  }, [])

  const writeDayunSummary = useCallback((dayunIndex: number, nextSummary: DayunSummary) => {
    summariesRef.current = {
      ...summariesRef.current,
      [dayunIndex]: nextSummary,
    }
    setSummaries((prev) => ({
      ...prev,
      [dayunIndex]: nextSummary,
    }))
  }, [])

  const loadAll = useCallback(async () => {
    if (!chartId || inflightRef.current) return
    inflightRef.current = true
    await Promise.resolve()
    setYearsLoaded(false)
    setYearsError('')
    setStreamDone(false)
    setStreamError('')
    setExpandedEvidence({})
    const initialGenerationRequestIds: Record<number, number> = {}

    // Stage 1: 即时拿所有年份（毫秒级）
    try {
      const resp = await baziAPI.fetchPastEventsYears(chartId)
      const data = resp.data
      setEvents(data.years || [])
      setDayunMeta(data.dayun_meta || [])
      // 初始化各大运 summary 占位
      // 未来段（start_year > currentYear）默认 folded=true，不参与 stage 2 自动生成
      // 已发生 + 当前段默认 loading=true，等流式拉取（可能 cache 命中也可能新生成）
      // 后端若发现 future 段在 cache 中，会主动 emit → setSummaries 把 folded 切回 false
      const init: Record<number, DayunSummary> = {}
      for (const dm of data.dayun_meta || []) {
        const isFuture = dm.start_year > currentYear
        if (isFuture) {
          init[dm.index] = { themes: [], summary: '', folded: true }
        } else {
          const requestId = ++generationRequestSeqRef.current
          initialGenerationRequestIds[dm.index] = requestId
          init[dm.index] = {
            themes: [],
            summary: '',
            loading: true,
            status: 'loading',
            generation: { startedAt: Date.now(), attempt: 0, source: 'initial', requestId },
          }
        }
      }
      summariesRef.current = init
      setSummaries(init)
      setYearsLoaded(true)
    } catch (e: unknown) {
      setYearsError(e instanceof Error ? e.message : '年份加载失败')
      inflightRef.current = false
      return
    }

    // Stage 2: 后台流式拉大运批语
    baziAPI.streamDayunSummaries(
      chartId,
      (item) => {
        const expectedRequestId = initialGenerationRequestIds[item.dayun_index]
        const current = summariesRef.current[item.dayun_index]
        if (expectedRequestId !== undefined) {
          if (current?.generation?.requestId !== expectedRequestId) {
            return
          }
        } else if (current?.generation || current?.summary || current?.years || current?.themes.length) {
          return
        }

        if (item.error) {
          writeDayunSummary(item.dayun_index, {
            themes: [],
            summary: '',
            error: item.error,
            loading: false,
            status: 'interrupted',
            generation: undefined,
            folded: false,
          })
        } else {
          // 收到 SSE 即意味着该段已成功生成（或缓存命中）→ 一律展开
          writeDayunSummary(item.dayun_index, {
            themes: item.themes || [],
            summary: item.summary || '',
            years: item.years || undefined,
            cached: Boolean(item.cached),
            loading: false,
            status: undefined,
            generation: undefined,
            folded: false,
          })
        }
      },
      (err) => {
        const initialStreamStillCurrent = Object.entries(initialGenerationRequestIds)
          .some(([index, requestId]) => {
            const current = summariesRef.current[Number(index)]
            return current?.generation?.source === 'initial' && current.generation.requestId === requestId
          })
        if (!initialStreamStillCurrent) return
        setStreamError(err)
        inflightRef.current = false
        markLoadingDayunsInterrupted(err || DAYUN_GENERATION_INTERRUPTED_COPY, 'initial')
      },
      () => {
        setStreamDone(true)
        inflightRef.current = false
      },
    )
  }, [chartId, markLoadingDayunsInterrupted, writeDayunSummary])

  // 用户点击 [展开 ▼] —— 折叠段进入"已展开但未生成批语"状态
  // 还不调任何 API，只切换 folded 状态显示 chips
  const handleExpand = useCallback((dayunIndex: number) => {
    setSummaries((prev) => ({
      ...prev,
      [dayunIndex]: { ...prev[dayunIndex], folded: false },
    }))
  }, [])

  // 用户点击 [收起 ▲] —— 折回但保留已加载的批语内容（下次展开不重新调）
  const handleCollapse = useCallback((dayunIndex: number) => {
    setSummaries((prev) => ({
      ...prev,
      [dayunIndex]: { ...prev[dayunIndex], folded: true },
    }))
  }, [])

  // 用户点击 [生成本段批语] —— 触发单段 SSE 生成
  const handleGenerateSegment = useCallback((dayunIndex: number, source: DayunGenerationSource = 'manual') => {
    if (!chartId) return
    const requestId = beginDayunGeneration(dayunIndex, source)
    baziAPI.streamDayunSummaries(
      chartId,
      (item) => {
        if (item.dayun_index !== dayunIndex) return
        const current = summariesRef.current[dayunIndex]
        if (current?.generation?.requestId !== requestId) return
        if (item.error) {
          writeDayunSummary(item.dayun_index, {
            ...current,
            themes: current?.themes || [],
            summary: current?.summary || '',
            error: item.error,
            loading: false,
            status: 'interrupted',
            generation: undefined,
            folded: false,
          })
        } else {
          writeDayunSummary(item.dayun_index, {
            themes: item.themes || [],
            summary: item.summary || '',
            years: item.years || undefined,
            cached: Boolean(item.cached),
            loading: false,
            status: undefined,
            generation: undefined,
            folded: false,
          })
        }
      },
      (err) => {
        const current = summariesRef.current[dayunIndex]
        if (current?.generation?.requestId !== requestId) return
        markDayunInterrupted(dayunIndex, err || DAYUN_GENERATION_FAILED_COPY)
      },
      () => {},
      [dayunIndex],
    )
  }, [beginDayunGeneration, chartId, markDayunInterrupted, writeDayunSummary])

  const recoverStaleDayunSummaries = useCallback(() => {
    const now = Date.now()
    const staleIndexes = Object.entries(summariesRef.current)
      .filter(([, summary]) => {
        if (!summary.loading || !summary.generation) return false
        return now - summary.generation.startedAt >= DAYUN_GENERATION_STALE_MS
      })
      .map(([index]) => Number(index))

    if (staleIndexes.length === 0) return
    inflightRef.current = false

    for (const index of staleIndexes) {
      const summary = summariesRef.current[index]
      const attempt = summary?.generation?.attempt ?? 0
      if (attempt < DAYUN_GENERATION_MAX_AUTO_RETRIES) {
        handleGenerateSegment(index, 'recovery')
      } else {
        markDayunInterrupted(index)
      }
    }
  }, [handleGenerateSegment, markDayunInterrupted])

  useEffect(() => {
    const recoverIfVisible = () => {
      if (document.visibilityState === 'visible') {
        recoverStaleDayunSummaries()
      }
    }

    document.addEventListener('visibilitychange', recoverIfVisible)
    window.addEventListener('pageshow', recoverStaleDayunSummaries)
    window.addEventListener('focus', recoverStaleDayunSummaries)
    window.addEventListener('online', recoverStaleDayunSummaries)
    return () => {
      document.removeEventListener('visibilitychange', recoverIfVisible)
      window.removeEventListener('pageshow', recoverStaleDayunSummaries)
      window.removeEventListener('focus', recoverStaleDayunSummaries)
      window.removeEventListener('online', recoverStaleDayunSummaries)
    }
  }, [recoverStaleDayunSummaries])

  useEffect(() => {
    if (isLoading) return
    if (!user) {
      navigate('/login')
      return
    }
    if (!chartId) return
    const timer = window.setTimeout(() => {
      void loadAll()
    }, 0)
    return () => window.clearTimeout(timer)
  }, [chartId, user, isLoading, navigate, loadAll])

  useEffect(() => {
    if (!user) return
    brandAPI.get()
      .then((r) => setBrand(r.data.data))
      .catch(() => setBrand(null))
  }, [user])

  // 按 dayun_index 分组（保持原顺序）
  const grouped: Array<{ meta: DayunMeta; years: YearEvent[] }> = dayunMeta.map((dm) => ({
    meta: dm,
    years: events.filter((y) => y.dayun_index === dm.index),
  })).filter((g) => g.years.length > 0)
  const summaryList = Object.values(summaries)
  const hasLoadingSummary = summaryList.some((summary) => summary.loading)
  const hasInterruptedSummary = summaryList.some((summary) => summary.status === 'interrupted')
  const progressCounts = countPastEventsProgress(summaries)
  const currentPeriod = getPastEventsCurrentPeriod(dayunMeta, events, currentYear)
  const currentGroup = currentPeriod.dayun
    ? grouped.find(({ meta }) => meta.index === currentPeriod.dayun?.index)
    : undefined
  const currentYearEvent = currentPeriod.year
  const headerStreamStatus = !yearsLoaded ? '正在加载年份时间轴……' :
    hasLoadingSummary ? '年份已就绪 · 大运批语正在后台生成' :
    hasInterruptedSummary ? '部分大运批语生成中断，可点击重试' :
    streamDone ? '已完成，所有大运批语已生成' :
    '年份已就绪 · 大运批语正在后台生成'

  const yearNarrative = (y: YearEvent): { text: string; status: 'loading' | 'ready' | 'empty' } => {
    return getPastEventsYearNarrativeState(y, summaries[y.dayun_index])
  }

  const exportSegments = buildPastEventsExportSegments(dayunMeta, events, summaries, includeEvidenceInExport)
  const hasExportableContent = exportSegments.length > 0
  const isMobileDevice = /iPhone|iPad|iPod|Android/i.test(navigator.userAgent)

  const handleExportPDF = async () => {
    if (!hasExportableContent) {
      showToast('暂无已生成的过往批语可导出', 'info')
      return
    }
    if (!isMobileDevice) {
      window.print()
      return
    }
    const el = document.querySelector('.past-events-print-layout') as HTMLElement | null
    if (!el) return
    setExportingPDF(true)
    const prevDisplay = el.style.display
    try {
      const [{ default: html2canvas }, { default: jsPDF }] = await Promise.all([
        import('html2canvas'),
        import('jspdf'),
      ])
      await document.fonts.ready
      el.style.display = 'block'
      const canvas = await html2canvas(el, { scale: 2, useCORS: true, logging: false })
      el.style.display = prevDisplay
      const imgData = canvas.toDataURL('image/jpeg', 0.92)
      const pdf = new jsPDF({ orientation: 'portrait', unit: 'mm', format: 'a4' })
      const pageW = pdf.internal.pageSize.getWidth()
      const pageH = pdf.internal.pageSize.getHeight()
      const imgH = (canvas.height * pageW) / canvas.width
      let remaining = imgH
      let offset = 0
      pdf.addImage(imgData, 'JPEG', 0, offset, pageW, imgH)
      remaining -= pageH
      while (remaining > 0) {
        offset -= pageH
        pdf.addPage()
        pdf.addImage(imgData, 'JPEG', 0, offset, pageW, imgH)
        remaining -= pageH
      }
      const suffix = new Date().toISOString().slice(0, 10).replace(/-/g, '')
      pdf.save(`缘聚-过往年运回看-${suffix}.pdf`)
    } catch {
      showToast('生成 PDF 失败，请稍后重试', 'error')
    } finally {
      el.style.display = prevDisplay
      setExportingPDF(false)
    }
  }

  return (
    <>
    <div className="screen-only" style={{ minHeight: '100vh', background: 'var(--bg-base)', paddingBottom: 60 }}>
      {/* 顶部导航 */}
      <div style={{
        display: 'flex', alignItems: 'center', gap: 12,
        padding: '16px 20px',
        background: 'var(--bg-card)',
        borderBottom: '1px solid var(--border-subtle)',
        position: 'sticky', top: 0, zIndex: 100,
      }}>
        <button
          onClick={() => navigate(-1)}
          style={{ background: 'none', border: 'none', cursor: 'pointer', color: 'var(--text-secondary)', padding: 4 }}
        >
          <ChevronLeft size={20} />
        </button>
        <div>
          <div style={{ fontFamily: 'Noto Serif SC, serif', fontSize: '1rem', color: 'var(--text-primary)', fontWeight: 600 }}>
            过往事件推算
          </div>
          <div style={{ fontSize: '0.7rem', color: 'var(--text-muted)', marginTop: 2 }}>
            {headerStreamStatus}
          </div>
        </div>
        <div className="past-events-export-actions">
          <label className="past-events-export-option">
            <input
              type="checkbox"
              checked={includeEvidenceInExport}
              onChange={(event) => setIncludeEvidenceInExport(event.target.checked)}
            />
            <span>包含依据</span>
          </label>
          <button
            type="button"
            className="past-events-export-button"
            onClick={handleExportPDF}
            disabled={!hasExportableContent || exportingPDF}
            title={hasExportableContent ? '导出已生成的过往批语' : '暂无已生成的过往批语'}
          >
            {exportingPDF ? <Loader2 size={14} className="past-events-export-spin" /> : <Download size={14} />}
            <span>{exportingPDF ? '生成中' : '导出已生成内容'}</span>
          </button>
        </div>
      </div>

      <div style={{ maxWidth: 700, margin: '0 auto', padding: '24px 16px' }}>
        {!yearsLoaded && !yearsError && (
          <div style={{ textAlign: 'center', padding: '60px 0' }}>
            <Loader2 size={32} style={{ color: 'var(--wu-jin)', animation: 'spin 1s linear infinite', marginBottom: 16 }} />
            <div style={{ color: 'var(--text-secondary)', fontSize: '0.9rem' }}>正在加载……</div>
          </div>
        )}

        {yearsError && (
          <div style={{ textAlign: 'center', padding: '40px 0' }}>
            <div style={{ color: 'var(--status-danger)', marginBottom: 16 }}>{yearsError}</div>
            <button
              onClick={loadAll}
              style={{
                background: 'var(--wu-jin)', color: '#fffdf8', border: 'none',
                borderRadius: 8, padding: '10px 24px', cursor: 'pointer', fontWeight: 600,
              }}
            >重新加载</button>
          </div>
        )}

        {yearsLoaded && events.length === 0 && (
          <EmptyState
            title="暂无过往年份数据"
            description="这份命盘暂时没有可展示的过往事件推算结果，可以返回结果页继续查看命盘或重新起盘。"
            action={
              <div style={{ display: 'flex', gap: 10, flexWrap: 'wrap' }}>
                {chartId && <Button href={`/history/${chartId}`} variant="primary">返回结果页</Button>}
                <Button href="/history" variant="secondary">查看历史命盘</Button>
              </div>
            }
          />
        )}

        {yearsLoaded && events.length > 0 && (
          <div>
            <div style={{ color: 'var(--text-muted)', fontSize: '0.75rem', marginBottom: 20, textAlign: 'center' }}>
              共推算 {events.length} 个年份 · 算法即时生成 · 大运批语后台生成
            </div>

            <PastEventsStatusPanel
              totalYears={events.length}
              progressCounts={progressCounts}
              currentGroup={currentGroup}
              currentYearEvent={currentYearEvent}
            />

            {grouped.map(({ meta, years }) => {
              const dyGan = meta.gan_zhi[0] || ''
              const dyWx = WUXING_GAN[dyGan] || 'tu'
              const dySum = summaries[meta.index]
              const segmentState = getPastEventsSegmentViewState(dySum)
              const isCurrentDayun = currentYear >= meta.start_year && currentYear <= meta.end_year
              return (
                <div
                  key={meta.index}
                  id={isCurrentDayun ? 'past-events-current-dayun' : undefined}
                  className={`past-events-segment${isCurrentDayun ? ' past-events-segment--current' : ''}`}
                >
                  {/* 大运标题 */}
                  <div style={{
                    display: 'flex', alignItems: 'center', gap: 8,
                    marginBottom: dySum ? 10 : 12,
                    paddingBottom: dySum ? 0 : 8,
                    borderBottom: dySum ? 'none' : '1px solid var(--border-subtle)',
                  }}>
                    <div style={{
                      fontFamily: 'Noto Serif SC, serif',
                      fontSize: '1.1rem',
                      fontWeight: 700,
                      color: `var(--wu-${dyWx})`,
                    }}>{meta.gan_zhi}</div>
                    <div style={{ fontSize: '0.75rem', color: 'var(--text-muted)' }}>
                      大运 {meta.start_age}-{meta.end_age}岁
                    </div>
                    {isCurrentDayun && (
                      <span className="past-events-current-badge">当前大运</span>
                    )}
                    <SegmentStateBadge state={segmentState} />
                    <div style={{
                      display: 'inline-flex',
                      alignItems: 'center',
                      gap: 6,
                      fontSize: '0.66rem',
                      color: 'var(--text-muted)',
                      border: '1px solid var(--border-subtle)',
                      borderRadius: 4,
                      padding: '2px 6px',
                      background: 'var(--bg-elevated)',
                      whiteSpace: 'nowrap',
                    }}>
                      <span>前五年天干主事</span>
                      <span style={{ opacity: 0.55 }}>·</span>
                      <span>后五年地支主事</span>
                    </div>
                    {meta.ten_god_power?.plain_title && (
                      <div style={{
                        display: 'inline-flex',
                        alignItems: 'center',
                        gap: 4,
                        fontSize: '0.66rem',
                        color: `var(--wu-${dyWx})`,
                        border: `1px solid color-mix(in srgb, var(--wu-${dyWx}) 42%, transparent)`,
                        borderRadius: 4,
                        padding: '2px 6px',
                        background: `color-mix(in srgb, var(--wu-${dyWx}) 8%, transparent)`,
                        whiteSpace: 'nowrap',
                      }}>
                        主导：{meta.ten_god_power.plain_title}
                      </div>
                    )}
                    {/* 未来段折叠/展开切换 */}
                    {dySum?.folded !== undefined && (
                      <button
                        onClick={() => dySum?.folded ? handleExpand(meta.index) : handleCollapse(meta.index)}
                        style={{
                          marginLeft: 'auto',
                          background: 'none',
                          border: '1px solid var(--border-subtle)',
                          borderRadius: 4,
                          padding: '2px 8px',
                          color: 'var(--text-muted)',
                          cursor: 'pointer',
                          fontSize: '0.72rem',
                          whiteSpace: 'nowrap',
                        }}
                      >{getPastEventsFutureToggleLabel(Boolean(dySum?.folded))}</button>
                    )}
                  </div>

                  {dySum?.folded && (
                    <div style={{
                      fontSize: '0.72rem',
                      color: 'var(--text-muted)',
                      padding: '4px 0 12px',
                      lineHeight: 1.6,
                    }}>
                      未来大运段，点击"展开年份信号"只查看算法信号；展开后再按需生成本段批语。
                    </div>
                  )}

                  {!dySum?.folded && (<>

                  {/* 大运整体总结块 */}
                  {dySum && (
                    <div style={{
                      background: `color-mix(in srgb, var(--wu-${dyWx}) 8%, var(--bg-card))`,
                      border: `1px solid color-mix(in srgb, var(--wu-${dyWx}) 25%, transparent)`,
                      borderRadius: 10,
                      padding: '12px 14px',
                      marginBottom: 12,
                    }}>
                      {dySum.loading && (
                        <div style={{ display: 'flex', alignItems: 'center', gap: 8, color: 'var(--text-muted)', fontSize: '0.78rem' }}>
                          <Loader2 size={14} style={{ animation: 'spin 1s linear infinite' }} />
                          正在生成本段大运批语……
                        </div>
                      )}
                      {dySum.status === 'interrupted' && (
                        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', gap: 10, color: 'var(--primary)', fontSize: '0.78rem' }}>
                          <span>{dySum.error || DAYUN_GENERATION_INTERRUPTED_COPY}</span>
                          <button
                            type="button"
                            onClick={() => handleGenerateSegment(meta.index, 'manual')}
                            style={{
                              border: '1px solid color-mix(in srgb, var(--wu-jin) 45%, transparent)',
                              background: 'transparent',
                              color: 'var(--wu-jin)',
                              borderRadius: 6,
                              padding: '4px 8px',
                              cursor: 'pointer',
                              fontSize: '0.72rem',
                            }}
                          >
                            重试
                          </button>
                        </div>
                      )}
                      {dySum.error && dySum.status !== 'interrupted' && (
                        <div style={{ color: 'var(--status-danger)', fontSize: '0.78rem' }}>
                          本段总结生成失败：{dySum.error}
                        </div>
                      )}
                      {!dySum.loading && dySum.status !== 'interrupted' && !dySum.error && dySum.themes.length > 0 && (
                        <>
                          <div style={{ display: 'flex', flexWrap: 'wrap', gap: 6, marginBottom: 8 }}>
                            {dySum.themes.map((theme) => {
                              const hasUp = theme.includes('↑')
                              const hasDown = theme.includes('↓')
                              const direction = hasUp ? '↑' : hasDown ? '↓' : null
                              const text = direction ? theme.replace(direction, '').trim() : theme
                              const dirColor = hasUp ? 'var(--status-success)' : hasDown ? 'var(--status-danger)' : null
                              const borderColor = dirColor ?? `var(--wu-${dyWx})`
                              return (
                                <span
                                  key={theme}
                                  style={{
                                    display: 'inline-flex',
                                    alignItems: 'center',
                                    gap: 2,
                                    fontSize: '0.68rem',
                                    padding: '2px 7px',
                                    borderRadius: 4,
                                    border: `1px solid ${borderColor}`,
                                    color: `var(--wu-${dyWx})`,
                                    whiteSpace: 'nowrap',
                                  }}
                                >
                                  {text}
                                  {direction && (
                                    <span style={{ fontSize: '0.85rem', fontWeight: 700, color: dirColor!, lineHeight: 1 }}>{direction}</span>
                                  )}
                                </span>
                              )
                            })}
                          </div>
                          <div style={{ color: 'var(--text-secondary)', fontSize: '0.83rem', lineHeight: 1.75 }}>
                            {dySum.summary}
                          </div>
                        </>
                      )}
                    </div>
                  )}

                  {/* 年份列表 */}
                  <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
                    {years.map((y) => {
                      const gan = y.gan_zhi?.[0] || ''
                      const wx = WUXING_GAN[gan] || 'tu'
                      const hasSignals = y.signals && y.signals.length > 0
                      const isFuture = y.year > currentYear
                      const isCurrentYear = y.year === currentYear
                      const evidenceKey = `${meta.index}-${y.year}`
                      const hasEvidence = Boolean(y.evidence_summary?.length)
                      const evidenceOpen = Boolean(expandedEvidence[evidenceKey])
                      return (
                        <div
                          key={y.year}
                          className={isCurrentYear ? 'past-events-year-card--current' : undefined}
                          style={{
                            position: 'relative',
                            background: 'var(--bg-card)',
                            borderRadius: 12,
                            padding: '14px 16px',
                            opacity: isFuture ? 0.75 : 1,
                            border: hasSignals
                              ? `1px ${isFuture ? 'dashed' : 'solid'} color-mix(in srgb, var(--wu-${wx}) 40%, transparent)`
                              : `1px ${isFuture ? 'dashed' : 'solid'} var(--border-subtle)`,
                          }}
                        >
                          {isFuture && (
                            <span style={{
                              position: 'absolute',
                              top: 8,
                              right: 10,
                              fontSize: '0.6rem',
                              padding: '1px 5px',
                              borderRadius: 3,
                              background: 'color-mix(in srgb, var(--wu-shui) 15%, transparent)',
                              border: '1px dashed var(--wu-shui)',
                              color: 'var(--wu-shui)',
                              letterSpacing: '0.05em',
                            }}>未来</span>
                          )}
                          <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 8, flexWrap: 'wrap' }}>
                            <span style={{
                              fontFamily: 'Noto Serif SC, serif',
                              fontWeight: 700,
                              fontSize: '1.05rem',
                              color: `var(--wu-${wx})`,
                            }}>{y.gan_zhi}</span>
                            <span style={{ color: 'var(--text-muted)', fontSize: '0.8rem' }}>
                              {y.year}年 · {y.age}岁
                            </span>
                            {isCurrentYear && (
                              <span className="past-events-current-year-badge">当前年份</span>
                            )}
                            {y.signals?.map((sig) => {
                              const signalMeta = getPastEventsSignalLabel(sig)
                              return (
                                <span
                                  key={sig}
                                  style={{
                                    fontSize: '0.65rem',
                                    padding: '2px 6px',
                                    borderRadius: 4,
                                    border: `1px solid ${signalMeta.color}`,
                                    color: signalMeta.color,
                                    whiteSpace: 'nowrap',
                                  }}
                                >{signalMeta.label}</span>
                              )
                            })}
                          </div>
                          {(() => {
                            const n = yearNarrative(y)
                            if (n.status === 'loading') {
                              return (
                                <div style={{ color: 'var(--text-muted)', fontSize: '0.78rem', fontStyle: 'italic' }}>
                                  本段批语正在生成…
                                </div>
                              )
                            }
                            if (n.status === 'ready') {
                              return (
                                <div style={{ color: 'var(--text-secondary)', fontSize: '0.85rem', lineHeight: 1.7 }}>
                                  {n.text}
                                </div>
                              )
                            }
                            // status === 'empty'：批语主动留空或被护栏拦下。
                            // 当卡片有 chips 时，用 chips 自动拼一句兜底，避免视觉断层。
                            if (y.signals && y.signals.length > 0) {
                              const chipLabels = y.signals
                                .map((sig) => getPastEventsSignalLabel(sig).label)
                                .join('、')
                              return (
                                <div style={{ color: 'var(--text-muted)', fontSize: '0.78rem', lineHeight: 1.6 }}>
                                  本年关键信号：{chipLabels}。详见下方命理依据。
                                </div>
                              )
                            }
                            return null
                          })()}
                          {(y.ten_god_power?.plain_title || y.dayun_phase || y.year_in_dayun) && (
                            <div className="past-events-year-details">
                              {y.ten_god_power?.plain_title && <span>主导力量：{y.ten_god_power.plain_title}</span>}
                              {y.year_in_dayun && <span>大运第 {y.year_in_dayun} 年</span>}
                              {y.dayun_phase && <span>{y.dayun_phase === 'gan' ? '天干主事' : '地支主事'}</span>}
                            </div>
                          )}
                          {hasEvidence && (
                            <div style={{ marginTop: 10 }}>
                              <button
                                type="button"
                                onClick={() => setExpandedEvidence((prev) => ({
                                  ...prev,
                                  [evidenceKey]: !prev[evidenceKey],
                                }))}
                                style={{
                                  display: 'inline-flex',
                                  alignItems: 'center',
                                  gap: 4,
                                  border: 'none',
                                  background: 'transparent',
                                  color: 'var(--text-muted)',
                                  fontSize: '0.72rem',
                                  padding: '2px 0',
                                  cursor: 'pointer',
                                }}
                                aria-expanded={evidenceOpen}
                              >
                                <ChevronDown
                                  size={14}
                                  style={{
                                    transform: evidenceOpen ? 'rotate(180deg)' : 'rotate(0deg)',
                                    transition: 'transform 0.16s ease',
                                  }}
                                />
                                命理依据
                              </button>
                              {evidenceOpen && (
                                <ul style={{
                                  margin: '8px 0 0',
                                  paddingLeft: 18,
                                  color: 'var(--text-muted)',
                                  fontSize: '0.72rem',
                                  lineHeight: 1.65,
                                }}>
                                  {y.evidence_summary!.map((ev, idx) => (
                                    <li key={`${evidenceKey}-${idx}`}>{ev}</li>
                                  ))}
                                </ul>
                              )}
                            </div>
                          )}
                        </div>
                      )
                    })}
                  </div>

                  {/* 展开但未生成批语 → 生成本段按钮 */}
                  {segmentState.canGenerateAi && (
                    <div style={{ marginTop: 16, textAlign: 'center' }}>
                      <button
                        onClick={() => handleGenerateSegment(meta.index, 'manual')}
                        style={{
                          background: `color-mix(in srgb, var(--wu-${dyWx}) 16%, transparent)`,
                          border: `1px solid color-mix(in srgb, var(--wu-${dyWx}) 50%, transparent)`,
                          borderRadius: 8,
                          padding: '10px 24px',
                          color: `var(--wu-${dyWx})`,
                          cursor: 'pointer',
                          fontSize: '0.85rem',
                          fontWeight: 600,
                        }}
                      >生成本段批语</button>
                    </div>
                  )}

                  </>)}
                </div>
              )
            })}

            {streamError && (
              <div style={{
                marginTop: 16, padding: '10px 14px',
                background: 'color-mix(in srgb, var(--status-danger) 12%, transparent)',
                border: '1px solid var(--status-danger)',
                borderRadius: 8,
                fontSize: '0.78rem',
                color: 'var(--status-danger)',
              }}>
                大运批语生成中断：{streamError}
                <button
                  onClick={loadAll}
                  style={{
                    marginLeft: 12,
                    background: 'none', border: '1px solid var(--status-danger)',
                    borderRadius: 4, padding: '2px 10px',
                    color: 'var(--status-danger)', cursor: 'pointer', fontSize: '0.72rem',
                  }}
                >重试</button>
              </div>
            )}

            <div style={{
              marginTop: 24, padding: '12px 16px',
              background: 'var(--bg-elevated)',
              borderRadius: 8,
              fontSize: '0.72rem',
              color: 'var(--text-muted)',
              lineHeight: 1.6,
            }}>
              本推算内容仅供参考，不构成任何决策建议。
            </div>
          </div>
        )}
      </div>
    </div>
    <PastEventsPrintLayout
      chart={{ id: chartId }}
      segments={exportSegments}
      includeEvidence={includeEvidenceInExport}
      brand={brand}
    />
    </>
  )
}
