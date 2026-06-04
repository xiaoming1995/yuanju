import { useCallback, useEffect, useRef, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { useAuth } from '../contexts/AuthContext'
import {
  compatibilityAPI,
  isV3DimensionScores,
  type CompatibilityDetail,
  type CompatibilityDimensionScoresLegacy,
  type CompatibilityDimensionScoresV3,
  type CompatibilityDurationAssessment,
  type CompatibilityRelationshipStrategy,
} from '../lib/api'
import {
  buildDecisionDashboardData,
  buildDecisionStageRisks,
} from '../lib/compatibilityDecision'
import {
  buildPersonalityValidationPlan,
  getCompatibilityQuestionLabel,
  getPersonalityMatchType,
} from '../lib/compatibilityPersonality'
import { toBlob, toPng } from 'html-to-image'
import jsPDF from 'jspdf'
import html2canvas from 'html2canvas'
import { brandAPI, type ExportBrand } from '../lib/api'
import CompatibilityShareCard from '../components/CompatibilityShareCard'
import CompatibilityPrintLayout from '../components/CompatibilityPrintLayout'
import CompatibilityStickyHeader from '../components/compatibility/CompatibilityStickyHeader'
import SectionBasicCharts from '../components/compatibility/SectionBasicCharts'
import DayPillarPortrait from '../components/compatibility/DayPillarPortrait'
import SectionVerdict from '../components/compatibility/SectionVerdict'
import SectionDeepAnalysis from '../components/compatibility/SectionDeepAnalysis'
import DeepReportNarrative from '../components/compatibility/deep-analysis/DeepReportNarrative'
import EvidenceDrawer from '../components/compatibility/EvidenceDrawer'
import { useToast } from '../components/ui/useToast'
import './CompatibilityResultPage.css'


function isDurationLevel(value: unknown): value is 'high' | 'medium' | 'low' {
  return value === 'high' || value === 'medium' || value === 'low'
}

function normalizeDurationAssessment(
  primary: CompatibilityDurationAssessment | null | undefined,
  fallback: CompatibilityDurationAssessment
): CompatibilityDurationAssessment {
  const hasPrimaryValue = Boolean(primary && (
    (typeof primary.summary === 'string' && primary.summary.trim()) ||
    (Array.isArray(primary.reasons) && primary.reasons.length > 0) ||
    isDurationLevel(primary.windows?.three_months?.level) ||
    isDurationLevel(primary.windows?.one_year?.level) ||
    isDurationLevel(primary.windows?.two_years_plus?.level)
  ))

  const source = hasPrimaryValue ? primary! : fallback

  return {
    overall_band: source.overall_band || fallback.overall_band,
    summary: source.summary?.trim() || fallback.summary,
    reasons: Array.isArray(source.reasons) ? source.reasons.filter(Boolean) : fallback.reasons,
    windows: {
      three_months: {
        level: isDurationLevel(source.windows?.three_months?.level) ? source.windows.three_months.level : fallback.windows.three_months.level,
      },
      one_year: {
        level: isDurationLevel(source.windows?.one_year?.level) ? source.windows.one_year.level : fallback.windows.one_year.level,
      },
      two_years_plus: {
        level: isDurationLevel(source.windows?.two_years_plus?.level) ? source.windows.two_years_plus.level : fallback.windows.two_years_plus.level,
      },
    },
  }
}

// LLM 报告里的 relationship_strategy 可能整体缺失或字段为空串（多见于历史 v1 报告），
// 逐字段回退到 reading.consulting_assessment 里始终满的规则基线，避免显示「暂无明确建议」。
function resolveRelationshipStrategy(
  llm: CompatibilityRelationshipStrategy | undefined,
  base: CompatibilityRelationshipStrategy | undefined,
): CompatibilityRelationshipStrategy | undefined {
  if (!llm && !base) return undefined
  const pick = (a?: string, b?: string) => (a && a.trim() ? a : (b || ''))
  return {
    communication: pick(llm?.communication, base?.communication),
    conflict: pick(llm?.conflict, base?.conflict),
    reality: pick(llm?.reality, base?.reality),
    boundary: pick(llm?.boundary, base?.boundary),
  }
}

function normalizeConsultingAssessment(detail: CompatibilityDetail) {
  const report = detail.latest_report?.content_structured
  const base = detail.reading.consulting_assessment
  return {
    relationship_diagnosis: report?.relationship_diagnosis || base?.relationship_diagnosis,
    decision_advice: report?.decision_advice || base?.decision_advice,
    stage_risks: report?.stage_risks?.length ? report.stage_risks : base?.stage_risks || [],
    relationship_strategy: resolveRelationshipStrategy(report?.relationship_strategy, base?.relationship_strategy),
    claim_evidence_links: report?.claim_evidence_links?.length ? report.claim_evidence_links : base?.claim_evidence_links || [],
  }
}

export default function CompatibilityResultPage() {
  const { id } = useParams()
  const { user, isLoading } = useAuth()
  const navigate = useNavigate()
  const { showToast } = useToast()
  const [detail, setDetail] = useState<CompatibilityDetail | null>(null)
  const [loading, setLoading] = useState(true)
  const [reportLoading, setReportLoading] = useState(false)
  const [error, setError] = useState('')
  const [brand, setBrand] = useState<ExportBrand | null>(null)
  const [shareModalOpen, setShareModalOpen] = useState(false)
  const [savingImage, setSavingImage] = useState(false)
  const [exportingPDF, setExportingPDF] = useState(false)
  const shareCardRef = useRef<HTMLDivElement>(null)
  const shareModalCloseBtnRef = useRef<HTMLButtonElement>(null)
  const shareTriggerBtnRef = useRef<HTMLButtonElement>(null)
  const prevShareModalOpenRef = useRef(false)
  const isMobileDevice = /iPhone|iPad|iPod|Android/i.test(navigator.userAgent)
  const isIOS = /iPhone|iPad|iPod/i.test(navigator.userAgent)

  const load = useCallback(async () => {
    if (!id) return
    const res = await compatibilityAPI.getDetail(id)
    setDetail(res.data.data)
  }, [id])

  useEffect(() => {
    if (isLoading) {
      return
    }
    if (!user) {
      navigate('/login')
      return
    }
    load()
      .catch((err: unknown) => setError(err instanceof Error ? err.message : '加载失败'))
      .finally(() => setLoading(false))
  }, [user, isLoading, navigate, load])

  useEffect(() => {
    if (!user) return
    brandAPI.get()
      .then(r => setBrand(r.data.data))
      .catch(() => setBrand(null))
  }, [user])

  useEffect(() => {
    if (shareModalOpen) {
      shareModalCloseBtnRef.current?.focus()
    } else if (prevShareModalOpenRef.current) {
      shareTriggerBtnRef.current?.focus()
    }
    prevShareModalOpenRef.current = shareModalOpen
  }, [shareModalOpen])

  const handleGenerateReport = async () => {
    if (reportLoading) return
    if (!id) return
    setReportLoading(true)
    setError('')
    try {
      await compatibilityAPI.generateReport(id)
      await load()
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : '生成合盘解读失败')
    } finally {
      setReportLoading(false)
    }
  }

  if (loading || isLoading) {
    return <div className="page"><div className="container" style={{ paddingTop: 40 }}>加载中...</div></div>
  }
  if (!detail) {
    return <div className="page"><div className="container" style={{ paddingTop: 40 }}>未找到合盘记录</div></div>
  }

  const reading = detail.reading
  const selfP = detail.participants.find(p => p.role === 'self')
  const partnerP = detail.participants.find(p => p.role === 'partner')

  const handleSaveImage = async () => {
    if (!shareCardRef.current) return
    setSavingImage(true)
    try {
      await document.fonts.ready
      const selfName = selfP?.display_name || '我'
      const partnerName = partnerP?.display_name || '伴侣'
      const fileName = `缘聚合盘-${selfName}-${partnerName}.png`

      if (isIOS) {
        const blob = await toBlob(shareCardRef.current, { quality: 0.98, pixelRatio: 3, cacheBust: true })
        if (!blob) throw new Error('生成图片失败')
        const file = new File([blob], fileName, { type: 'image/png' })
        if (navigator.canShare && navigator.canShare({ files: [file] })) {
          await navigator.share({
            files: [file],
            title: '缘聚合盘 · 命理合参',
            text: `${selfName} × ${partnerName} 综合契合度 ${detail?.reading.overall_score ?? ''} 分`,
          })
        } else {
          const objectUrl = URL.createObjectURL(blob)
          Object.assign(document.createElement('a'), { href: objectUrl, download: fileName }).click()
          setTimeout(() => URL.revokeObjectURL(objectUrl), 5000)
        }
      } else if (isMobileDevice) {
        const blob = await toBlob(shareCardRef.current, { quality: 0.98, pixelRatio: 3, cacheBust: true })
        if (!blob) throw new Error('生成图片失败')
        const objectUrl = URL.createObjectURL(blob)
        Object.assign(document.createElement('a'), { href: objectUrl, download: fileName }).click()
        setTimeout(() => URL.revokeObjectURL(objectUrl), 5000)
      } else {
        const dataUrl = await toPng(shareCardRef.current, { quality: 0.98, pixelRatio: 2, cacheBust: true })
        Object.assign(document.createElement('a'), { href: dataUrl, download: fileName }).click()
      }
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : ''
      if (!msg.includes('AbortError') && !msg.includes('cancel')) {
        showToast('生成图片失败，请稍后重试', 'error')
      }
    } finally {
      setSavingImage(false)
    }
  }

  const handleExportPDF = async () => {
    if (!detail?.latest_report?.content_structured) return
    if (!isMobileDevice) {
      window.print()
      return
    }
    const el = document.querySelector('.compat-print-layout') as HTMLElement | null
    if (!el) return
    setExportingPDF(true)
    const prevDisplay = el.style.display
    try {
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
      const selfName = selfP?.display_name || '我'
      const partnerName = partnerP?.display_name || '伴侣'
      pdf.save(`缘聚合盘-${selfName}-${partnerName}.pdf`)
    } catch {
      showToast('生成 PDF 失败，请稍后重试', 'error')
    } finally {
      el.style.display = prevDisplay
      setExportingPDF(false)
    }
  }

  const structuredReport = detail.latest_report?.content_structured
  const durationAssessment = normalizeDurationAssessment(structuredReport?.duration_assessment, reading.duration_assessment)
  const reportDimensions = Array.isArray(structuredReport?.dimensions) ? structuredReport.dimensions : []
  const reportRisks = Array.isArray(structuredReport?.risks) ? structuredReport.risks.filter(Boolean) : []
  const consulting = normalizeConsultingAssessment(detail)
  const decisionStageRisks = buildDecisionStageRisks(consulting.stage_risks, durationAssessment)
  const decisionDashboard = buildDecisionDashboardData({
    diagnosis: consulting.relationship_diagnosis,
    advice: consulting.decision_advice,
    stageRisks: consulting.stage_risks,
    duration: durationAssessment,
    evidences: detail.evidences,
    overallLevel: reading.overall_level,
  })
  const isV3 = (reading.analysis_version === 'v3' || reading.analysis_version === 'v3.1') && isV3DimensionScores(reading.dimension_scores)
  const legacyScores = isV3 ? null : (reading.dimension_scores as CompatibilityDimensionScoresLegacy)
  const v3Scores = isV3 ? (reading.dimension_scores as CompatibilityDimensionScoresV3) : null
  // 双方性格画像与差异已改由 AI 深度解读（LLM）生成，渲染在 DeepReportNarrative 内。
  // 行动计划/验证计划维持 legacy 门控，避免改动其在 V3 下的可见性
  const personalityValidationPlan = legacyScores
    ? buildPersonalityValidationPlan({
        questionLabel: getCompatibilityQuestionLabel(reading.primary_question),
        matchType: getPersonalityMatchType(legacyScores, reading.primary_question, reading.relationship_stage),
        advice: consulting.decision_advice,
        stageRisks: consulting.stage_risks,
        duration: durationAssessment,
        hasEvidence: detail.evidences.length > 0 || consulting.claim_evidence_links.length > 0,
      })
    : null

  return (
    <>
    <div className="page compatibility-result-page">
      <div className="container compatibility-result-container">
        <div className="compat-export-actions">
          <button
            type="button"
            ref={shareTriggerBtnRef}
            className="btn btn-secondary"
            disabled={!structuredReport}
            onClick={() => setShareModalOpen(true)}
            title={!structuredReport ? '请先生成命理解读' : ''}
          >
            分享图片
          </button>
          <button
            type="button"
            className="btn btn-primary"
            disabled={!structuredReport || exportingPDF}
            onClick={handleExportPDF}
            title={!structuredReport ? '请先生成命理解读' : ''}
          >
            {exportingPDF ? '生成中…' : '导出 PDF'}
          </button>
        </div>

        <CompatibilityStickyHeader
          selfName={selfP?.display_name || '我'}
          partnerName={partnerP?.display_name || '对方'}
          overallScore={reading.overall_score}
          verdict={decisionDashboard.verdict}
        />
        <SectionBasicCharts self={selfP || null} partner={partnerP || null} />
        <DayPillarPortrait self={selfP || null} partner={partnerP || null} />
        <SectionVerdict
          dashboard={decisionDashboard}
          isV3={isV3}
          v3Scores={v3Scores}
          legacyScores={legacyScores}
          overallScore={reading.overall_score}
          overallLevel={reading.overall_level}
          findings={decisionDashboard.findings}
        />
        <SectionDeepAnalysis
          personalityValidationPlan={personalityValidationPlan}
          decisionStageRisks={decisionStageRisks}
          durationAssessment={durationAssessment}
          relationshipStrategy={consulting.relationship_strategy}
          dashboard={decisionDashboard}
        />
        <EvidenceDrawer
          evidences={detail.evidences}
          claimEvidenceLinks={consulting.claim_evidence_links}
        />
        <DeepReportNarrative
          hasReport={Boolean(detail.latest_report)}
          structuredReport={structuredReport}
          reportDimensions={reportDimensions}
          reportRisks={reportRisks}
          rawContent={detail.latest_report?.content}
          error={error}
          reportLoading={reportLoading}
          onGenerateReport={handleGenerateReport}
        />
      </div>

      {shareModalOpen && structuredReport && (
        <div
          className="compat-share-modal"
          role="dialog"
          aria-modal="true"
          aria-label="分享图片预览"
          tabIndex={-1}
          onClick={() => setShareModalOpen(false)}
          onKeyDown={(e) => { if (e.key === 'Escape') setShareModalOpen(false) }}
        >
          <div className="compat-share-modal-panel" onClick={e => e.stopPropagation()}>
            <header className="compat-share-modal-head">
              <h3>分享图片预览</h3>
              <button
                type="button"
                className="compat-share-modal-close"
                ref={shareModalCloseBtnRef}
                onClick={() => setShareModalOpen(false)}
                aria-label="关闭"
              >×</button>
            </header>
            <div className="compat-share-modal-preview">
              <CompatibilityShareCard
                ref={shareCardRef}
                reading={reading}
                participants={detail.participants}
                evidences={detail.evidences}
                decision={decisionDashboard}
                stageRisks={decisionStageRisks}
                structured={structuredReport ?? null}
                relationshipStrategy={consulting.relationship_strategy}
                brand={brand}
              />
            </div>
            <p className="compat-share-modal-hint">导出的图片为完整版 · 预览可上下滚动</p>
            <footer className="compat-share-modal-footer">
              <button
                type="button"
                className="btn btn-primary"
                onClick={handleSaveImage}
                disabled={savingImage}
              >
                {savingImage ? '生成中…' : isIOS ? '保存 / 分享' : '保存到本地'}
              </button>
            </footer>
          </div>
        </div>
      )}

    </div>
    <CompatibilityPrintLayout
      reading={reading}
      participants={detail.participants}
      evidences={detail.evidences}
      decision={decisionDashboard}
      stageRisks={decisionStageRisks}
      structured={structuredReport ?? null}
      relationshipStrategy={consulting.relationship_strategy}
      brand={brand}
    />
    </>
  )
}
