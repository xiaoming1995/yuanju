import { useLocation, useParams, useNavigate } from 'react-router-dom'
import { useEffect, useRef, useState } from 'react'
import { Diamond, X } from 'lucide-react'
import { useAuth } from '../contexts/AuthContext'
import { baziAPI, fetchShenshaAnnotations } from '../lib/api'
import type { AIReport, ShenshaAnnotation, StructuredReport } from '../lib/api'
import { cleanReportText } from '../lib/reportText'
import WuxingRadar from '../components/WuxingRadar'
import DayunTimeline from '../components/DayunTimeline'
import YongshenBadge from '../components/YongshenBadge'
import MingpanAvatar from '../components/MingpanAvatar'
import TiaohouCard from '../components/TiaohouCard'
import ShareCard from '../components/ShareCard'
import PrintLayout from '../components/PrintLayout'
import { toPng, toBlob } from 'html-to-image'
import jsPDF from 'jspdf'
import html2canvas from 'html2canvas'
import './ResultPage.css'

// 💡 特性开关 (Feature Flags)
const ENABLE_MINGPAN_AVATAR = false // 暂时隐藏专属命理头像模块

const WUXING_MAP: Record<string, string> = {
  '木': 'mu', '火': 'huo', '土': 'tu', '金': 'jin', '水': 'shui'
}

// 神煞极性表（与后端 ShenShaPolarity 保持同步）
// ji = 吉神（金色），xiong = 凶煞（红色），zhong = 中性（灰色）
const SHENSHA_POLARITY: Record<string, string> = {
  // 吉神
  '天乙贵人': 'ji', '太极贵人': 'ji', '文昌贵人': 'ji', '禄神': 'ji',
  '天德贵人': 'ji', '月德贵人': 'ji', '天德合': 'ji', '月德合': 'ji',
  '德秀贵人': 'ji', '金舆贵人': 'ji', '天喜': 'ji', '天厨贵人': 'ji',
  '国印贵人': 'ji', '三奇贵人': 'ji', '日德': 'ji', '将星': 'ji', '福星贵人': 'ji', '天医': 'ji',
  '十灵日': 'ji', '词馆': 'ji',
  // 凶煞
  '羊刃': 'xiong', '飞刃': 'xiong', '劫煞': 'xiong', '亡神': 'xiong',
  '孤辰': 'xiong', '寡宿': 'xiong', '阴差阳错': 'xiong',
  '魁罡': 'xiong', '十恶大败': 'xiong', '天罗地网': 'xiong', '地网': 'xiong', '灾煞': 'xiong', '童子煞': 'xiong',
  '流霞': 'xiong', '吊客': 'xiong', '墓门': 'xiong',
  // 中性
  '桃花': 'zhong', '驿马': 'zhong', '华盖': 'zhong', '红艳': 'zhong',
}

const LOADING_STEPS = [
  '提取四柱大运神煞...',
  '结合真太阳时精确校准星运...',
  '依据典籍推断月令与格局...',
  '深度分析抓取全局调候用神...',
  '正在汇总专属呈现命局详析...'
]

const REPORT_TERMS = [
  { term: '用神', desc: '命局中最需要扶助或调节的关键五行。' },
  { term: '忌神', desc: '容易加重失衡，需要节制或避开的五行。' },
  { term: '格局', desc: '月令与十神形成的命局主结构。' },
  { term: '大运', desc: '每十年一段的人生阶段性趋势。' },
]

const GAN_WUXING: Record<string, string> = {
  '甲': '木', '乙': '木',
  '丙': '火', '丁': '火',
  '戊': '土', '己': '土',
  '庚': '金', '辛': '金',
  '壬': '水', '癸': '水',
}

const GAN_YINYANG: Record<string, number> = {
  '甲': 1, '乙': 0,
  '丙': 1, '丁': 0,
  '戊': 1, '己': 0,
  '庚': 1, '辛': 0,
  '壬': 1, '癸': 0,
}

const WUXING_SHENG: Record<string, string> = {
  '木': '火', '火': '土', '土': '金', '金': '水', '水': '木',
}

const WUXING_KE: Record<string, string> = {
  '木': '土', '火': '金', '土': '水', '金': '木', '水': '火',
}

const TEN_GOD_META: Record<string, { relation: string; group: string; group_label: string; summary: string }> = {
  '比肩': { relation: '同我', group: 'peer', group_label: '比劫', summary: '自我、同类、独立意识与同辈协作。' },
  '劫财': { relation: '同我', group: 'peer', group_label: '比劫', summary: '同辈竞争、资源分配、行动冲劲与合作博弈。' },
  '食神': { relation: '我生', group: 'output', group_label: '食伤', summary: '表达、才艺、稳定输出、口福与享受。' },
  '伤官': { relation: '我生', group: 'output', group_label: '食伤', summary: '创意表达、突破规则、才华外放与锋芒。' },
  '正财': { relation: '我克', group: 'wealth', group_label: '财星', summary: '稳定财富、务实经营、责任感与现实积累。' },
  '偏财': { relation: '我克', group: 'wealth', group_label: '财星', summary: '机会资源、流动财富、人情经营与商业嗅觉。' },
  '正官': { relation: '克我', group: 'official', group_label: '官杀', summary: '规则、职位、责任、名誉与秩序感。' },
  '七杀': { relation: '克我', group: 'official', group_label: '官杀', summary: '外部压力、竞争、规则挑战与行动魄力。' },
  '正印': { relation: '生我', group: 'seal', group_label: '印星', summary: '学习、贵人、保护、资质与正统资源。' },
  '偏印': { relation: '生我', group: 'seal', group_label: '印星', summary: '灵感、研究、特殊资源、独特思维与保护。' },
}

function buildReportDigestItems(structured: StructuredReport, result: BaziResult) {
  const firstChapter = structured.chapters?.[0]
  const adviceChapter = structured.chapters?.find((c) =>
    /建议|总结|总论|方向|展望|策略|布局/.test(c.title),
  )
  const yongshen = structured.yongshen || result.yongshen || ''
  const jishen = structured.jishen || result.jishen || ''
  const fallbackAdvice = yongshen && jishen
    ? `优先围绕「${yongshen}」方向布局，对「${jishen}」相关领域更克制谨慎。`
    : '先读摘要，再展开与当前问题最相关的章节。'

  return [
    {
      label: '总体判断',
      value: cleanReportText(structured.analysis?.summary || firstChapter?.brief) || '已生成完整命理解读，可继续查看各章节。',
    },
    {
      label: '喜用重点',
      value: `${yongshen || '待判定'}：优先观察能补足命局平衡的方向。`,
    },
    {
      label: '主要风险',
      value: `${jishen || '待判定'}：相关五行过旺或失衡时，需要在选择与节奏上更谨慎。`,
    },
    {
      label: '行动建议',
      value: cleanReportText(structured.analysis?.advice) || cleanReportText(adviceChapter?.brief) || fallbackAdvice,
    },
  ]
}


interface TenGodDayMaster {
  gan: string
  wuxing: string
  label: string
}

interface TenGodRelationItem {
  pillar: string
  pillar_label: string
  gan: string
  wuxing: string
  ten_god: string
  group?: string
  group_label?: string
  relation: string
  summary: string
}

interface TenGodHiddenStemItem {
  gan: string
  wuxing: string
  ten_god: string
  group?: string
  group_label?: string
  relation: string
  summary: string
}

interface TenGodHiddenStemGroup {
  pillar: string
  pillar_label: string
  branch: string
  items: TenGodHiddenStemItem[]
}

interface TenGodRelationMatrix {
  day_master: TenGodDayMaster
  heavenly_stems: TenGodRelationItem[]
  hidden_stems: TenGodHiddenStemGroup[]
}

interface BaziResult {
  year_gan: string; year_zhi: string
  month_gan: string; month_zhi: string
  day_gan: string; day_zhi: string
  hour_gan: string; hour_zhi: string
  year_gan_wuxing: string; year_zhi_wuxing: string
  month_gan_wuxing: string; month_zhi_wuxing: string
  day_gan_wuxing: string; day_zhi_wuxing: string
  hour_gan_wuxing: string; hour_zhi_wuxing: string
  // 藏干
  year_hide_gan: string[]; month_hide_gan: string[]
  day_hide_gan: string[]; hour_hide_gan: string[]
  
  // 十神和长生
  year_gan_shishen: string; month_gan_shishen: string; day_gan_shishen: string; hour_gan_shishen: string;
  year_zhi_shishen: string[]; month_zhi_shishen: string[]; day_zhi_shishen: string[]; hour_zhi_shishen: string[];
  year_di_shi: string; month_di_shi: string; day_di_shi: string; hour_di_shi: string;
  year_xing_yun: string; month_xing_yun: string; day_xing_yun: string; hour_xing_yun: string;
  year_xun_kong: string; month_xun_kong: string; day_xun_kong: string; hour_xun_kong: string;
  year_shen_sha: string[]; month_shen_sha: string[]; day_shen_sha: string[]; hour_shen_sha: string[];
  // 纳音
  year_na_yin: string; month_na_yin: string
  day_na_yin: string; hour_na_yin: string
  // 真太阳时
  true_solar_hour: number; true_solar_minute: number
  wuxing: { mu: number; huo: number; tu: number; jin: number; shui: number }
  yongshen: string; jishen: string
  tiaohou?: {
    expected: string[]
    tou: string[]
    cang: string[]
    text: string
  }
  // 交运时间
  start_yun_solar: string;
  dayun: Array<{
    index: number;
    gan: string;
    zhi: string;
    start_age: number;
    start_year: number;
    end_year: number;
    gan_shishen: string;
    zhi_shishen: string;
    di_shi: string;
    jin_bu_huan?: { qian_level: string; qian_desc: string; hou_level: string; hou_desc: string; verse: string } | null;
    liu_nian: Array<{
      year: number;
      age: number;
      gan_zhi: string;
      gan_shishen: string;
      zhi_shishen: string;
    }>;
  }>
  birth_year: number; birth_month: number; birth_day: number; birth_hour: number; gender: string
  // 命格
  ming_ge?: string
  ming_ge_desc?: string
  ten_god_relation?: TenGodRelationMatrix
}

function getShiShen(dayGan: string, targetGan: string) {
  const dayWx = GAN_WUXING[dayGan]
  const targetWx = GAN_WUXING[targetGan]
  if (!dayWx || !targetWx) return ''

  const sameYinyang = GAN_YINYANG[dayGan] === GAN_YINYANG[targetGan]
  if (dayWx === targetWx) return sameYinyang ? '比肩' : '劫财'
  if (WUXING_SHENG[dayWx] === targetWx) return sameYinyang ? '食神' : '伤官'
  if (WUXING_KE[dayWx] === targetWx) return sameYinyang ? '偏财' : '正财'
  if (WUXING_KE[targetWx] === dayWx) return sameYinyang ? '七杀' : '正官'
  if (WUXING_SHENG[targetWx] === dayWx) return sameYinyang ? '偏印' : '正印'
  return ''
}

function relationFromTenGod(tenGod: string) {
  return TEN_GOD_META[tenGod] || { relation: '', group: '', group_label: '', summary: '' }
}

function tenGodSummary(tenGod: string) {
  return relationFromTenGod(tenGod).summary
}

function buildStemRelation(
  dayGan: string,
  pillar: string,
  pillarLabel: string,
  gan: string,
  wuxing: string,
  tenGod: string,
  isDayMaster = false,
): TenGodRelationItem {
  if (isDayMaster) {
    return {
      pillar,
      pillar_label: pillarLabel,
      gan,
      wuxing: wuxing || GAN_WUXING[gan] || '',
      ten_god: '日主 / 日元',
      relation: '命主自身',
      summary: '这是命盘的参照点，其他十神都以此天干为中心推导。',
    }
  }
  const resolvedTenGod = tenGod || getShiShen(dayGan, gan)
  const meta = relationFromTenGod(resolvedTenGod)
  return {
    pillar,
    pillar_label: pillarLabel,
    gan,
    wuxing: wuxing || GAN_WUXING[gan] || '',
    ten_god: resolvedTenGod,
    group: meta.group,
    group_label: meta.group_label,
    relation: meta.relation,
    summary: tenGodSummary(resolvedTenGod),
  }
}

function buildHiddenStemGroup(
  dayGan: string,
  pillar: string,
  pillarLabel: string,
  branch: string,
  hiddenGans: string[],
): TenGodHiddenStemGroup {
  return {
    pillar,
    pillar_label: pillarLabel,
    branch,
    items: hiddenGans
      .map((gan) => {
        const tenGod = getShiShen(dayGan, gan)
        const meta = relationFromTenGod(tenGod)
        return {
          gan,
          wuxing: GAN_WUXING[gan] || '',
          ten_god: tenGod,
          group: meta.group,
          group_label: meta.group_label,
          relation: meta.relation,
          summary: tenGodSummary(tenGod),
        }
      })
      .filter(item => item.ten_god),
  }
}

function buildTenGodRelationMatrix(result: BaziResult): TenGodRelationMatrix {
  return {
    day_master: {
      gan: result.day_gan,
      wuxing: result.day_gan_wuxing || GAN_WUXING[result.day_gan] || '',
      label: `${result.day_gan}${result.day_gan_wuxing || GAN_WUXING[result.day_gan] || ''}`,
    },
    heavenly_stems: [
      buildStemRelation(result.day_gan, 'year', '年干', result.year_gan, result.year_gan_wuxing, result.year_gan_shishen),
      buildStemRelation(result.day_gan, 'month', '月干', result.month_gan, result.month_gan_wuxing, result.month_gan_shishen),
      buildStemRelation(result.day_gan, 'day', '日干', result.day_gan, result.day_gan_wuxing, result.day_gan_shishen, true),
      buildStemRelation(result.day_gan, 'hour', '时干', result.hour_gan, result.hour_gan_wuxing, result.hour_gan_shishen),
    ],
    hidden_stems: [
      buildHiddenStemGroup(result.day_gan, 'year', '年支', result.year_zhi, result.year_hide_gan || []),
      buildHiddenStemGroup(result.day_gan, 'month', '月支', result.month_zhi, result.month_hide_gan || []),
      buildHiddenStemGroup(result.day_gan, 'day', '日支', result.day_zhi, result.day_hide_gan || []),
      buildHiddenStemGroup(result.day_gan, 'hour', '时支', result.hour_zhi, result.hour_hide_gan || []),
    ],
  }
}


export default function ResultPage() {
  const location = useLocation()
  const { id } = useParams()
  const navigate = useNavigate()
  const { user } = useAuth()

  const [result, setResult] = useState<BaziResult | null>(location.state?.result || null)
  const [report, setReport] = useState<AIReport | null>(location.state?.report || null)
  const [isGuest] = useState(location.state?.isGuest ?? !user)
  const [loading, setLoading] = useState(!result && !!id)
  const [reportMode, setReportMode] = useState<'brief' | 'detail'>('detail')
  const [savingImage, setSavingImage] = useState(false)
  const [exportingPDF, setExportingPDF] = useState(false)
  const shareCardRef = useRef<HTMLDivElement>(null)

  // 神煞注解状态
  const [shenshaMap, setShenshaMap] = useState<Map<string, ShenshaAnnotation>>(new Map())
  const [activeAnnotation, setActiveAnnotation] = useState<ShenshaAnnotation | null>(null)
  const [activeMingGe, setActiveMingGe] = useState<{ name: string; desc: string } | null>(null)
  const hoverTimer = useRef<ReturnType<typeof setTimeout> | null>(null)

  // 预加载神煞注解
  useEffect(() => {
    fetchShenshaAnnotations()
      .then(list => {
        const map = new Map<string, ShenshaAnnotation>()
        list.forEach(a => map.set(a.name, a))
        setShenshaMap(map)
      })
      .catch(() => { /* 注解加载失败不影响主功能 */ })
  }, [])

  const handleSaveImage = async () => {
    if (!shareCardRef.current) return
    setSavingImage(true)

    const isMobile = /iPhone|iPad|iPod|Android/i.test(navigator.userAgent)
    const isIOS = /iPhone|iPad|iPod/i.test(navigator.userAgent)

    try {
      await document.fonts.ready

      if (isIOS) {
        // ✅ iOS 最佳方案：Web Share API + File Blob
        // 调起系统原生分享面板，用户可直接选“存储图像”保存到相册
        const blob = await toBlob(shareCardRef.current, {
          quality: 0.98,
          pixelRatio: 3,
          cacheBust: true,
        })
        if (!blob) throw new Error('生成图片失败')

        const fileName = `缘聚命理-${result?.year_gan ?? ''}年${result?.month_gan ?? ''}月.png`
        const file = new File([blob], fileName, { type: 'image/png' })

        if (navigator.canShare && navigator.canShare({ files: [file] })) {
          // 支持 Web Share API（iOS 15+ Safari 全款支持）
          await navigator.share({
            files: [file],
            title: `缘聚命理 · 八字命理报告`,
            text: `我的八字命理：${result?.year_gan ?? ''}${result?.year_zhi ?? ''}年`,
          })
        } else {
          // 退化到 Blob Object URL——比 base64 更靠谱
          const objectUrl = URL.createObjectURL(blob)
          Object.assign(document.createElement('a'), {
            href: objectUrl, download: fileName,
          }).click()
          setTimeout(() => URL.revokeObjectURL(objectUrl), 5000)
        }
      } else if (isMobile) {
        // Android：直接下载
        const blob = await toBlob(shareCardRef.current, {
          quality: 0.98, pixelRatio: 3, cacheBust: true,
        })
        if (!blob) throw new Error('生成图片失败')
        const objectUrl = URL.createObjectURL(blob)
        const fileName = `缘聚命理-${result?.year_gan ?? ''}年${result?.month_gan ?? ''}月.png`
        Object.assign(document.createElement('a'), {
          href: objectUrl, download: fileName,
        }).click()
        setTimeout(() => URL.revokeObjectURL(objectUrl), 5000)
      } else {
        // 桌面端：toPng + 下载
        const dataUrl = await toPng(shareCardRef.current, {
          quality: 0.98, pixelRatio: 2, cacheBust: true,
        })
        const link = document.createElement('a')
        link.download = `缘聚命理-${result?.year_gan ?? ''}年${result?.month_gan ?? ''}月.png`
        link.href = dataUrl
        link.click()
      }
    } catch (err: unknown) {
      // 用户主动取消分享不算错误
      const msg = err instanceof Error ? err.message : ''
      if (!msg.includes('AbortError') && !msg.includes('cancel')) {
        alert('生成图片失败，请稍后重试')
      }
    } finally {
      setSavingImage(false)
    }
  }

  const isMobileDevice = /iPhone|iPad|iPod|Android/i.test(navigator.userAgent)

  const handleExportPDF = async () => {
    if (!report) {
      setReportError('请先生成命理解读，再导出 PDF 报告')
      return
    }
    if (!isMobileDevice) {
      window.print()
      return
    }
    // 移动端：用 html2canvas + jsPDF 生成 PDF 文件下载
    const el = document.querySelector('.print-only') as HTMLElement | null
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
      const fileName = `缘聚命理-命书.pdf`
      pdf.save(fileName)
    } catch {
      alert('生成 PDF 失败，请稍后重试')
    } finally {
      el.style.display = prevDisplay
      setExportingPDF(false)
    }
  }

  // AI 解读状态
  const [reportLoading, setReportLoading] = useState(false)
  const [isStreaming, setIsStreaming] = useState(false)
  const [isThinking, setIsThinking] = useState(false)
  const [thinkingSeconds, setThinkingSeconds] = useState(0)
  const [streamingText, setStreamingText] = useState('')
  const [reportError, setReportError] = useState('')
  const [loadingStepIndex, setLoadingStepIndex] = useState(0)

  useEffect(() => {
    let timer: number
    if (reportLoading) {
      setLoadingStepIndex(0)
      timer = window.setInterval(() => {
        setLoadingStepIndex(prev => {
          if (prev < LOADING_STEPS.length - 1) return prev + 1
          return prev
        })
      }, 4000)
    } else {
      setLoadingStepIndex(0)
    }
    return () => window.clearInterval(timer)
  }, [reportLoading])

  // 推理计时器
  useEffect(() => {
    let timer: number
    if (isThinking) {
      setThinkingSeconds(0)
      timer = window.setInterval(() => {
        setThinkingSeconds(prev => prev + 1)
      }, 1000)
    }
    return () => window.clearInterval(timer)
  }, [isThinking])

  // 确立目前针对的独立资源 id
  // 有三个来源可能带有 id：1. 历史页面跳入 URL 的 id；2. HomePage 计算后传入的 state.chartId; 3. 新建后 result 从后端捞出的 chart.id(这里为了简化统一用 route 的方式)
  // 此页面核心判定是 targetId
  const targetId = id || location.state?.chartId

  // 从历史记录加载
  useEffect(() => {
    if (id && !result) {
      baziAPI.getHistoryDetail(id)
        .then(res => {
          setResult(res.data.result || res.data.chart || null)
          setReport(res.data.report || null)
        })
        .catch(() => navigate('/history'))
        .finally(() => setLoading(false))
    }
  }, [id]) // eslint-disable-line react-hooks/exhaustive-deps

  // 点击"生成 AI 解读"按钮
  const handleGenerateReport = async () => {
    if (!targetId) {
      setReportError('并未侦测到有效的命盘快照身份码，无法生成记录。');
      return;
    }
    setReportLoading(true)
    setIsStreaming(false)
    setIsThinking(false)
    setStreamingText('')
    setReportError('')

    let currentText = ''
    let isFirstByte = true
    await baziAPI.generateReportStream(
      targetId,
      (text) => {
        if (isFirstByte) {
          setReportLoading(false)
          setIsThinking(false)
          setIsStreaming(true)
          isFirstByte = false
        }
        currentText += text
        setStreamingText(currentText)
      },
      (err) => {
        setReportError(err)
        setIsStreaming(false)
        setIsThinking(false)
        setReportLoading(false)
      },
      () => {
        // 流结束：先保持 isStreaming=true 避免闪烁，等拉取完结构化数据后再统一切换
        baziAPI.getHistoryDetail(targetId).then(res => {
          setResult(res.data.result || res.data.chart || null)
          setReport(res.data.report || null)
        }).catch(err => {
          console.error('Failed to fetch finished report', err)
        }).finally(() => {
          setIsStreaming(false)
          setIsThinking(false)
          setReportLoading(false)
        })
      },
      () => {
        // 推理模型进入思考阶段
        setIsThinking(true)
        setReportLoading(false) // 关闭普通 loading，显示 thinking UI
      }
    )
  }

  if (loading) return <LoadingSkeleton />
  if (!result) return <div className="page container" style={{ paddingTop: 120 }}>数据加载失败</div>

  const pillars = [
    { label: '年柱', gan: result.year_gan, zhi: result.year_zhi, ganWx: result.year_gan_wuxing, zhiWx: result.year_zhi_wuxing, hideGan: result.year_hide_gan || [], naYin: result.year_na_yin || '', ganShiShen: result.year_gan_shishen, zhiShiShen: result.year_zhi_shishen || [], diShi: result.year_di_shi, xingYun: result.year_xing_yun, xunKong: result.year_xun_kong, shenSha: result.year_shen_sha || [] },
    { label: '月柱', gan: result.month_gan, zhi: result.month_zhi, ganWx: result.month_gan_wuxing, zhiWx: result.month_zhi_wuxing, hideGan: result.month_hide_gan || [], naYin: result.month_na_yin || '', ganShiShen: result.month_gan_shishen, zhiShiShen: result.month_zhi_shishen || [], diShi: result.month_di_shi, xingYun: result.month_xing_yun, xunKong: result.month_xun_kong, shenSha: result.month_shen_sha || [] },
    { label: '日柱', gan: result.day_gan, zhi: result.day_zhi, ganWx: result.day_gan_wuxing, zhiWx: result.day_zhi_wuxing, hideGan: result.day_hide_gan || [], naYin: result.day_na_yin || '', ganShiShen: result.day_gan_shishen, zhiShiShen: result.day_zhi_shishen || [], diShi: result.day_di_shi, xingYun: result.day_xing_yun, xunKong: result.day_xun_kong, shenSha: result.day_shen_sha || [] },
    { label: '时柱', gan: result.hour_gan, zhi: result.hour_zhi, ganWx: result.hour_gan_wuxing, zhiWx: result.hour_zhi_wuxing, hideGan: result.hour_hide_gan || [], naYin: result.hour_na_yin || '', ganShiShen: result.hour_gan_shishen, zhiShiShen: result.hour_zhi_shishen || [], diShi: result.hour_di_shi, xingYun: result.hour_xing_yun, xunKong: result.hour_xun_kong, shenSha: result.hour_shen_sha || [] },
  ]
  const relation = result.ten_god_relation ?? buildTenGodRelationMatrix(result)
  const hiddenStemGroups = relation.hidden_stems.filter(group => group.items.length > 0)
  const dayPillarCellClass = (index: number) => index === 2 ? ' is-day-pillar-cell' : ''
  const dayunPillarsLabel = `${result.year_gan}${result.year_zhi} ${result.month_gan}${result.month_zhi} ${result.day_gan}${result.day_zhi} ${result.hour_gan}${result.hour_zhi}`

  const structured = report?.content_structured ?? null
  const reportDigestItems = structured ? buildReportDigestItems(structured, result) : []
  // 旧报告降级：解析纯文字 content
  const reportSections = structured ? [] : parseReport(report?.content || '')

  return (
    <>
      <div className="result-page page screen-only">
        <div className="container">

        {/* 生辰标题 */}
        <div className="result-header animate-fade-up">
          <div className="result-birth-info">
            {result.birth_year}年{result.birth_month}月{result.birth_day}日 {result.birth_hour}时
            &nbsp;·&nbsp;{result.gender === 'male' ? '男命' : '女命'}
          </div>
          <h1 className="result-pillars serif">
            {pillars.map(p => `${p.gan}${p.zhi}`).join('·')}
          </h1>
          <div className="result-tags">
            <span className={`wuxing-badge ${result.yongshen ? 'wuxing-' + (WUXING_MAP[result.yongshen?.charAt(0)] || 'jin') : 'wuxing-unknown'}`}>
              喜用：{result.yongshen || (reportLoading ? '测算中...' : '待生成')}
            </span>
            <span className={`wuxing-badge ${result.jishen ? 'wuxing-' + (WUXING_MAP[result.jishen?.charAt(0)] || 'huo') : 'wuxing-unknown'}`}>
              忌：{result.jishen || (reportLoading ? '测算中...' : '待生成')}
            </span>
            {result.ming_ge && (
              <span
                className="mingge-badge"
                onClick={() => setActiveMingGe({ name: result.ming_ge!, desc: result.ming_ge_desc || '' })}
                title="点击查看格局说明"
              >
                {result.ming_ge}
              </span>
            )}
          </div>
        </div>

        {/* 命盘详情 */}
        <div className="professional-view animate-fade-up">

            {/* 四柱数据网格 (Professional Data Grid) */}
            <div className="pillars-section card bazi-primary-panel">
              <div className="result-panel-heading">
                <div>
                  <span className="result-panel-kicker">排盘核心</span>
                  <h2 className="section-title serif">基本排盘</h2>
                </div>
                <p>先确认四柱主盘，再向下查看十神、五行与调候结构。</p>
              </div>
              <div className="bazi-data-grid bazi-data-grid--primary">
                
                {/* 标尺列1：列头 */}
                <div className="grid-cell row-label">日期</div>
                {pillars.map((p, i) => <div key={i} className={`grid-cell col-header${i === 2 ? ' is-day-pillar' : ''}${dayPillarCellClass(i)}`}>{p.label}</div>)}

                {/* 主星行 */}
                <div className="grid-cell row-label">主星</div>
                {pillars.map((p, i) => <div key={i} className={`grid-cell top-shishen text-muted${dayPillarCellClass(i)}`}>{p.ganShiShen}</div>)}

                {/* 天干行 */}
                <div className="grid-cell row-label">天干</div>
                {pillars.map((p, i) => (
                  <div key={i} className={`grid-cell main-char gan wuxing-text-${WUXING_MAP[p.ganWx] || 'jin'}${dayPillarCellClass(i)}`}>
                    <span>{p.gan}</span><span className="wx-tag">{p.ganWx}</span>
                  </div>
                ))}

                {/* 地支行 */}
                <div className="grid-cell row-label">地支</div>
                {pillars.map((p, i) => (
                  <div key={i} className={`grid-cell main-char zhi wuxing-text-${WUXING_MAP[p.zhiWx] || 'jin'}${dayPillarCellClass(i)}`}>
                    <span>{p.zhi}</span><span className="wx-tag">{p.zhiWx}</span>
                  </div>
                ))}

                {/* 藏干行 */}
                <div className="grid-cell row-label">藏干</div>
                {pillars.map((p, i) => (
                  <div key={i} className={`grid-cell hide-gan-cell${dayPillarCellClass(i)}`}>
                    {p.hideGan.map((g, idx) => (
                       <div key={idx} className={`hg-row wuxing-text-${WUXING_MAP[GAN_WUXING[g]] || 'shui'}`} style={{ color: 'var(--text-color)' }}>{g}</div>
                    ))}
                  </div>
                ))}

                {/* 副星行 */}
                <div className="grid-cell row-label">副星</div>
                {pillars.map((p, i) => (
                  <div key={i} className={`grid-cell hide-gan-cell text-muted${dayPillarCellClass(i)}`}>
                    {p.zhiShiShen.map((ss, idx) => <div key={idx} className="hg-row">{ss}</div>)}
                  </div>
                ))}

                {/* 星运行 */}
                <div className="grid-cell row-label">星运</div>
                {pillars.map((p, i) => <div key={i} className={`grid-cell text-muted${dayPillarCellClass(i)}`}>{p.xingYun || p.diShi}</div>)}

                {/* 自坐行 */}
                <div className="grid-cell row-label">自坐</div>
                {pillars.map((p, i) => <div key={i} className={`grid-cell text-muted${dayPillarCellClass(i)}`}>{p.diShi}</div>)}

                {/* 空亡行 */}
                <div className="grid-cell row-label">空亡</div>
                {pillars.map((p, i) => <div key={i} className={`grid-cell text-muted${dayPillarCellClass(i)}`}>{p.xunKong}</div>)}

                {/* 纳音行 */}
                <div className="grid-cell row-label">纳音</div>
                {pillars.map((p, i) => <div key={i} className={`grid-cell text-muted nayin${dayPillarCellClass(i)}`}>{p.naYin}</div>)}

                {/* 神煞行 */}
                <div className="grid-cell row-label shensha-label">神煞</div>
                {pillars.map((p, i) => (
                  <div key={i} className={`grid-cell shensha-cell${dayPillarCellClass(i)}`}>
                    {p.shenSha.map((sh, idx) => {
                      const polarity = SHENSHA_POLARITY[sh] || 'zhong'
                      const hasAnnotation = shenshaMap.has(sh)
                      return (
                        <span
                          key={idx}
                          className={`shensha-tag shensha-tag--${polarity}${hasAnnotation ? ' shensha-tag--clickable' : ''}`}
                          onClick={() => {
                            const ann = shenshaMap.get(sh)
                            if (ann) setActiveAnnotation(ann)
                          }}
                          onMouseEnter={() => {
                            if (!hasAnnotation) return
                            hoverTimer.current = setTimeout(() => {
                              const ann = shenshaMap.get(sh)
                              if (ann) setActiveAnnotation(ann)
                            }, 300)
                          }}
                          onMouseLeave={() => {
                            if (hoverTimer.current) clearTimeout(hoverTimer.current)
                          }}
                        >{sh}</span>
                      )
                    })}
                  </div>
                ))}

              </div>
            </div>

            <section className="ten-god-relation-section" aria-labelledby="ten-god-relation-title">
              <div className="ten-god-relation-header">
                <div>
                  <h2 id="ten-god-relation-title" className="section-title serif">命主十神关系</h2>
                  <p>命主日元：<strong>{relation.day_master.label}</strong></p>
                </div>
                <span>以日干为参照点，查看其他天干与藏干对应的十神。</span>
              </div>

              <div className="ten-god-stem-grid">
                {relation.heavenly_stems.map((item) => (
                  <article key={item.pillar} className={`ten-god-relation-card ${item.pillar === 'day' ? 'is-day-master' : ''}`}>
                    <div className="ten-god-card-topline">
                      <span>{item.pillar_label}</span>
                      {item.group_label && <em>{item.group_label}</em>}
                    </div>
                    <div className="ten-god-card-main">
                      <strong className={`wuxing-text-${WUXING_MAP[item.wuxing] || 'jin'}`}>{item.gan}</strong>
                      <span>{item.ten_god}</span>
                    </div>
                    <p>{item.relation} · {item.summary}</p>
                  </article>
                ))}
              </div>

              {hiddenStemGroups.length > 0 && (
                <div className="ten-god-hidden-block">
                  <h3>地支藏干关系</h3>
                  <div className="ten-god-hidden-grid">
                    {hiddenStemGroups.map((group) => (
                      <article key={group.pillar} className="ten-god-hidden-card">
                        <div className="ten-god-hidden-title">
                          <span>{group.pillar_label}</span>
                          <strong>{group.branch}</strong>
                        </div>
                        <div className="ten-god-hidden-list">
                          {group.items.map((item) => (
                            <div key={`${group.pillar}-${item.gan}`} className="ten-god-hidden-item">
                              <span className={`wuxing-text-${WUXING_MAP[item.wuxing] || 'jin'}`}>{item.gan}</span>
                              <strong>{item.ten_god}</strong>
                              <em>{item.relation}</em>
                              <small>{item.summary}</small>
                            </div>
                          ))}
                        </div>
                      </article>
                    ))}
                  </div>
                </div>
              )}
            </section>

            <section className="result-structure-section" aria-labelledby="structure-title">
              <div className="result-section-heading">
                <span className="result-section-kicker">结构判断</span>
                <h2 id="structure-title" className="section-title serif">命局结构</h2>
                <p>把五行强弱、喜忌倾向与调候线索放在同一组里看，信息更集中。</p>
              </div>

              <div className="result-structure-grid">
                <div className="structure-card structure-card--wuxing card">
                  <h3 className="structure-card-title serif">五行分布</h3>
                  <WuxingRadar wuxing={result.wuxing} />
                </div>

                <div className="structure-card structure-card--yongshen">
                  <YongshenBadge yongshen={result.yongshen || ''} jishen={result.jishen || ''} />
                </div>

                {result.tiaohou && (
                  <div className="structure-card structure-card--tiaohou card">
                    <TiaohouCard
                      dayGan={result.day_gan}
                      monthZhi={result.month_zhi}
                      tiaohou={result.tiaohou}
                    />
                  </div>
                )}
              </div>
            </section>


            {/* 命理专属头像 (Feature Flag 控制) */}
            {ENABLE_MINGPAN_AVATAR && (
              <div className="mingpan-avatar-section card">
                <h2 className="section-title serif">专属命理头像</h2>
                <p className="section-desc">根据你的喜用神五行，程序化生成专属命元图腾</p>
                <MingpanAvatar
                  yongshen={result.yongshen || ''}
                  jishen={result.jishen || ''}
                  dayGan={result.day_gan || ''}
                />
              </div>
            )}

            {/* 大运时间轴 */}
            <section className="dayun-section">
              <DayunTimeline
                dayun={result.dayun}
                birthYear={result.birth_year}
                startYunSolar={result.start_yun_solar}
                dayGan={result.day_gan || ''}
                gender={result.gender}
                pillarsLabel={dayunPillarsLabel}
                chartId={targetId}
              />
            </section>
          </div>

        {/* AI 解读区域 */}
        <div className="report-section card animate-fade-up">
          <div className="report-section-header">
            <h2 className="section-title serif">命理解读</h2>
            <div className="report-header-actions">
              {report && (
                <>
                  <button
                    id="save-card-btn"
                    className="btn btn-ghost btn-sm"
                    onClick={handleSaveImage}
                    disabled={savingImage}
                  >
                    {savingImage ? '生成中...' : '保存分享图'}
                  </button>
                  <button
                    id="export-report-btn"
                    className="btn btn-ghost btn-sm"
                    onClick={handleExportPDF}
                    disabled={exportingPDF}
                  >
                    {exportingPDF ? '生成中...' : '导出 PDF'}
                  </button>
                </>
              )}
            </div>
          </div>

          {/* 已有报告 */}
          {report && (
            <div className="report-sections">
              {/* 精简/专业切换按钮（仅新格式报告显示） */}
              {structured && (
                <>
                  <div className="report-digest-card">
                    <div className="report-digest-heading">
                      <span>阅读摘要</span>
                      <strong className="serif">{structured.yongshen || result.yongshen || '命局'}为线索</strong>
                    </div>
                    <div className="report-digest-grid">
                      {reportDigestItems.map(item => (
                        <div key={item.label} className="report-digest-item">
                          <span>{item.label}</span>
                          <p>{item.value}</p>
                        </div>
                      ))}
                    </div>
                  </div>

                  <div className="report-mode-switcher">
                    <button
                      className={`mode-btn${reportMode === 'brief' ? ' active' : ''}`}
                      onClick={() => setReportMode('brief')}
                    >精简版</button>
                    <button
                      className={`mode-btn${reportMode === 'detail' ? ' active' : ''}`}
                      onClick={() => setReportMode('detail')}
                    >完整解读</button>
                  </div>

                  <div className="report-term-glossary" aria-label="命理术语解释">
                    {REPORT_TERMS.map(item => (
                      <div key={item.term} className="report-term-item">
                        <strong>{item.term}</strong>
                        <span>{item.desc}</span>
                      </div>
                    ))}
                  </div>
                </>
              )}

              {/* 新格式：结构化渲染 */}
              {structured ? (
                <>
                  {reportMode === 'detail' && structured.analysis?.logic && (
                    <div className="report-block report-analysis">
                      <h3 className="report-block-title serif"><Diamond size={16} className="title-diamond-icon" /> 命局分析总览</h3>
                      <div className="report-block-content">
                        {cleanReportText(structured.analysis.logic)
                          .split(/\n{2,}/)
                          .filter(Boolean)
                          .map((para, idx) => <p key={idx}>{para}</p>)}
                      </div>
                    </div>
                  )}
                  {reportMode === 'brief' && structured.analysis?.summary && (
                    <div className="report-summary">
                      <span>{cleanReportText(structured.analysis.summary)}</span>
                    </div>
                  )}
                  <div className="report-chapter-list">
                    {(structured.chapters || []).map((ch, i) => {
                      const raw = reportMode === 'brief' ? ch.brief : ch.detail
                      const paragraphs = cleanReportText(raw)
                        .split(/\n{2,}/)
                        .map(p => p.trim())
                        .filter(Boolean)
                      return (
                        <details key={i} className="report-chapter-detail" open>
                          <summary>
                            <span className="serif">【{ch.title}】</span>
                            <em>{cleanReportText(ch.brief)}</em>
                          </summary>
                          <div className="report-block-content">
                            {paragraphs.length > 0
                              ? paragraphs.map((para, idx) => <p key={idx}>{para}</p>)
                              : <p>{cleanReportText(raw)}</p>}
                          </div>
                        </details>
                      )
                    })}
                  </div>
                </>
              ) : (
                /* 旧格式：降级渲染 */
                reportSections.length > 0 ? reportSections.map((sec, i) => (
                  <div key={i} className="report-block">
                    <h3 className="report-block-title serif">{sec.title}</h3>
                    <p className="report-block-content">{cleanReportText(sec.content)}</p>
                  </div>
                )) : (
                  <div className="report-content">{cleanReportText(report.content)}</div>
                )
              )}
              <p className="report-disclaimer">
                本报告内容仅供参考，不构成任何决策建议。
              </p>
            </div>
          )}

          {/* 流式生成中 */}
          {isStreaming && (
            <div className="report-sections animate-fade-in">
              <div className="report-content" style={{ whiteSpace: 'pre-wrap', fontFamily: 'monospace', lineHeight: 1.8 }}>
                {streamingText}
                <span className="cursor-blink">|</span>
              </div>
            </div>
          )}

          {/* 推理模型正在思考 */}
          {isThinking && !isStreaming && (
            <div className="ai-loading-container animate-fade-in">
              <div className="ai-loading-icon">
                <div className="spinner"></div>
              </div>
              <div className="ai-loading-step">
                <div className="ai-loading-text">
                  正在深度推演中...  已思考 {thinkingSeconds} 秒
                </div>
              </div>
            </div>
          )}

          {/* 初始加载等待动画（SSE连接建立前） */}
          {reportLoading && !isStreaming && !isThinking && (
            <div className="ai-loading-container animate-fade-in">
              <div className="ai-loading-icon">
                <div className="spinner"></div>
              </div>
              <div className="ai-loading-step">
                <div key={loadingStepIndex} className="ai-loading-text">
                  {LOADING_STEPS[loadingStepIndex]}
                </div>
              </div>
            </div>
          )}

          {/* 报错 */}
          {reportError && !reportLoading && !isStreaming && (
            <p className="form-error" style={{ margin: '12px 0' }}>{reportError}</p>
          )}

          {/* 未生成：显示按钮或引导 */}
          {!report && !reportLoading && !isStreaming && !isThinking && (
            <>
              {!isGuest ? (
                <div className="report-cta">
                  <p className="report-cta-desc">
                    点击下方按钮，生成性格、感情、事业、健康四维解读
                  </p>
                  <button
                    id="generate-ai-report"
                    className="btn btn-primary"
                    onClick={handleGenerateReport}
                  >
                    生成命理解读
                  </button>
                </div>
              ) : (
                <div className="guest-banner">
                  <span>登录后可获得完整解读报告，并保存命盘记录</span>
                  <a href="/register" className="btn btn-primary btn-sm">立即注册</a>
                </div>
              )}
            </>
          )}

          {report && (
            <div className="report-action-bar">
              <button className="btn btn-ghost" onClick={() => navigate('/')}>重新起盘</button>
              {user && <a href="/history" className="btn btn-ghost">查看历史</a>}
              {user && targetId && (
                <button
                  className="btn btn-ghost report-action-highlight"
                  onClick={() => navigate(`/bazi/${targetId}/past-events`)}
                >过往事件</button>
              )}
              <button
                className="btn btn-ghost"
                onClick={handleExportPDF}
                disabled={exportingPDF}
              >
                {exportingPDF ? '生成中...' : '导出 PDF'}
              </button>
            </div>
          )}
        </div>
      </div>

      {/* 隐藏的分享卡片（用于生成图片，不可见） */}
      <div style={{ position: 'fixed', top: -9999, left: -9999, zIndex: -1, pointerEvents: 'none' }}>
        <ShareCard
          ref={shareCardRef}
          birthYear={result.birth_year}
          birthMonth={result.birth_month}
          birthDay={result.birth_day}
          birthHour={result.birth_hour}
          gender={result.gender}
          yearGan={result.year_gan} yearZhi={result.year_zhi}
          monthGan={result.month_gan} monthZhi={result.month_zhi}
          dayGan={result.day_gan} dayZhi={result.day_zhi}
          hourGan={result.hour_gan} hourZhi={result.hour_zhi}
          yearGanWx={result.year_gan_wuxing} yearZhiWx={result.year_zhi_wuxing}
          monthGanWx={result.month_gan_wuxing} monthZhiWx={result.month_zhi_wuxing}
          dayGanWx={result.day_gan_wuxing} dayZhiWx={result.day_zhi_wuxing}
          hourGanWx={result.hour_gan_wuxing} hourZhiWx={result.hour_zhi_wuxing}
          yongshen={result.yongshen || ''}
          jishen={result.jishen || ''}
          structured={report?.content_structured ?? null}
        />
      </div>

      {/* 神煞注解浮层卡片 */}
      {activeAnnotation && (
        <div
          className="shensha-modal-overlay"
          onClick={() => setActiveAnnotation(null)}
        >
          <div
            className="shensha-modal-card"
            onClick={e => e.stopPropagation()}
          >
            <div className="shensha-modal-header">
              <div className="shensha-modal-title">
                <span className={`shensha-modal-dot shensha-modal-dot--${activeAnnotation.polarity}`} />
                <span className="shensha-modal-name">{activeAnnotation.name}</span>
                <span className={`shensha-modal-badge shensha-modal-badge--${activeAnnotation.polarity}`}>
                  {activeAnnotation.polarity === 'ji' ? '吉神' : activeAnnotation.polarity === 'xiong' ? '凶煞' : '中性'}
                </span>
              </div>
              <button
                className="shensha-modal-close"
                onClick={() => setActiveAnnotation(null)}
                aria-label="关闭"
              >
                <X size={18} />
              </button>
            </div>
            <div className="shensha-modal-divider" />
            <div className="shensha-modal-body">
              {activeAnnotation.category && (
                <span style={{ fontSize: 11, color: '#a78bfa', background: '#2a1a4e', borderRadius: 4, padding: '2px 8px', marginBottom: 8, display: 'inline-block' }}>
                  {activeAnnotation.category}
                </span>
              )}
              {activeAnnotation.short_desc && (
                <p style={{ color: '#c0b0ff', fontSize: 13, margin: '6px 0 10px', fontStyle: 'italic' }}>{activeAnnotation.short_desc}</p>
              )}
              <p className="shensha-modal-description">{activeAnnotation.description}</p>
            </div>
          </div>
        </div>
      )}

      {/* 命格说明 Modal */}
      {activeMingGe && (
        <div
          className="shensha-modal-overlay"
          onClick={() => setActiveMingGe(null)}
        >
          <div
            className="shensha-modal-card"
            onClick={e => e.stopPropagation()}
          >
            <div className="shensha-modal-header">
              <div className="shensha-modal-title">
                <span className="mingge-modal-dot" />
                <span className="shensha-modal-name">{activeMingGe.name}</span>
                <span className="mingge-modal-badge">格局</span>
              </div>
              <button
                className="shensha-modal-close"
                onClick={() => setActiveMingGe(null)}
                aria-label="关闭"
              >
                <X size={18} />
              </button>
            </div>
            <div className="shensha-modal-divider" />
            <div className="shensha-modal-body">
              <p className="shensha-modal-description">{activeMingGe.desc}</p>
            </div>
          </div>
        </div>
      )}
      </div>
      {report && (
        <PrintLayout
          birthYear={result.birth_year}
          birthMonth={result.birth_month}
          birthDay={result.birth_day}
          birthHour={result.birth_hour}
          gender={result.gender}
          yongshen={result.yongshen || ''}
          jishen={result.jishen || ''}
          mingGe={result.ming_ge || ''}
          mingGeDesc={result.ming_ge_desc || ''}
          pillars={pillars}
          dayun={result.dayun}
          structured={structured}
          shenshaMap={shenshaMap}
          tenGodRelation={relation}
        />
      )}
    </>
  )
}

function parseReport(content: string) {
  const sections: { title: string; content: string }[] = []
  const matches = content.matchAll(/【(.+?)】\n?([\s\S]*?)(?=【|$)/g)
  for (const m of matches) {
    sections.push({ title: `【${m[1]}】`, content: m[2].trim() })
  }
  return sections
}

function LoadingSkeleton() {
  return (
    <div className="result-page page">
      <div className="container">
        <div style={{ paddingTop: 40 }}>
          <div className="skeleton" style={{ height: 32, width: 300, marginBottom: 16 }} />
          <div className="skeleton" style={{ height: 48, width: 400, marginBottom: 32 }} />
          <div className="skeleton" style={{ height: 300, borderRadius: 20 }} />
        </div>
      </div>
    </div>
  )
}
