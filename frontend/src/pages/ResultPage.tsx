import { useLocation, useParams, useNavigate } from 'react-router-dom'
import { useEffect, useRef, useState } from 'react'
import { useAuth } from '../contexts/AuthContext'
import { baziAPI } from '../lib/api'
import type { AIReport } from '../lib/api'
import WuxingRadar from '../components/WuxingRadar'
import DayunTimeline from '../components/DayunTimeline'
import YongshenBadge from '../components/YongshenBadge'
import MingpanAvatar from '../components/MingpanAvatar'
import ShareCard from '../components/ShareCard'
import { toPng, toBlob } from 'html-to-image'
import './ResultPage.css'

const WUXING_MAP: Record<string, string> = {
  '木': 'mu', '火': 'huo', '土': 'tu', '金': 'jin', '水': 'shui'
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
    liu_nian: Array<{
      year: number;
      age: number;
      gan_zhi: string;
      gan_shishen: string;
      zhi_shishen: string;
    }>;
  }>
  birth_year: number; birth_month: number; birth_day: number; birth_hour: number; gender: string
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
  const [reportMode, setReportMode] = useState<'brief' | 'detail'>('brief')
  const [savingImage, setSavingImage] = useState(false)
  const shareCardRef = useRef<HTMLDivElement>(null)

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

  // 检测是否移动端（用于隐藏 PDF 按鈕）
  const isMobileDevice = /iPhone|iPad|iPod|Android/i.test(navigator.userAgent)

  // AI 解读状态
  const [reportLoading, setReportLoading] = useState(false)
  const [reportError, setReportError] = useState('')
  const [loadingStepIndex, setLoadingStepIndex] = useState(0)

  const LOADING_STEPS = [
    '☯️ 飞盘排签，提取四柱大运神煞...',
    '🔍 校准星运，结合真太阳时精算...',
    '📚 翻阅《子平真诠》推断月令格局...',
    '🌙 对照《穷通宝鉴》抓取调候用神...',
    '✒️ 宗师沉思，正在精排你专属的命局详析...'
  ]

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
    setReportError('')
    try {
      const res = await baziAPI.generateReport(targetId)
      setReport(res.data.report || null)
      // 生成后后端可能已经提取了 yongshen / jishen，同步更新到界面
      if (res.data.chart) {
        setResult((prev: any) => ({ ...prev, ...res.data.chart }))
      }
    } catch (err: unknown) {
      setReportError(err instanceof Error ? err.message : 'AI 生成失败，请重试')
    } finally {
      setReportLoading(false)
    }
  }

  if (loading) return <LoadingSkeleton />
  if (!result) return <div className="page container" style={{ paddingTop: 120 }}>数据加载失败</div>

  const pillars = [
    { label: '年柱', gan: result.year_gan, zhi: result.year_zhi, ganWx: result.year_gan_wuxing, zhiWx: result.year_zhi_wuxing, hideGan: result.year_hide_gan || [], naYin: result.year_na_yin || '', ganShiShen: result.year_gan_shishen, zhiShiShen: result.year_zhi_shishen || [], diShi: result.year_di_shi, xingYun: result.year_xing_yun, xunKong: result.year_xun_kong, shenSha: result.year_shen_sha || [] },
    { label: '月柱', gan: result.month_gan, zhi: result.month_zhi, ganWx: result.month_gan_wuxing, zhiWx: result.month_zhi_wuxing, hideGan: result.month_hide_gan || [], naYin: result.month_na_yin || '', ganShiShen: result.month_gan_shishen, zhiShiShen: result.month_zhi_shishen || [], diShi: result.month_di_shi, xingYun: result.month_xing_yun, xunKong: result.month_xun_kong, shenSha: result.month_shen_sha || [] },
    { label: '日柱', gan: result.day_gan, zhi: result.day_zhi, ganWx: result.day_gan_wuxing, zhiWx: result.day_zhi_wuxing, hideGan: result.day_hide_gan || [], naYin: result.day_na_yin || '', ganShiShen: result.day_gan_shishen, zhiShiShen: result.day_zhi_shishen || [], diShi: result.day_di_shi, xingYun: result.day_xing_yun, xunKong: result.day_xun_kong, shenSha: result.day_shen_sha || [] },
    { label: '时柱', gan: result.hour_gan, zhi: result.hour_zhi, ganWx: result.hour_gan_wuxing, zhiWx: result.hour_zhi_wuxing, hideGan: result.hour_hide_gan || [], naYin: result.hour_na_yin || '', ganShiShen: result.hour_gan_shishen, zhiShiShen: result.hour_zhi_shishen || [], diShi: result.hour_di_shi, xingYun: result.hour_xing_yun, xunKong: result.hour_xun_kong, shenSha: result.hour_shen_sha || [] },
  ]

  const structured = report?.content_structured ?? null
  // 旧报告降级：解析纯文字 content
  const reportSections = structured ? [] : parseReport(report?.content || '')

  return (
    <div className="result-page page">
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
              喜用：{result.yongshen || (reportLoading ? 'AI测算中...' : '待生成')}
            </span>
            <span className={`wuxing-badge ${result.jishen ? 'wuxing-' + (WUXING_MAP[result.jishen?.charAt(0)] || 'huo') : 'wuxing-unknown'}`}>
              忌：{result.jishen || (reportLoading ? 'AI测算中...' : '待生成')}
            </span>
          </div>
        </div>

        {/* 命盘详情 */}
        <div className="professional-view animate-fade-up">

            {/* 四柱数据网格 (Professional Data Grid) */}
            <div className="pillars-section card">
              <h2 className="section-title serif">基本排盘</h2>
              <div className="bazi-data-grid">
                
                {/* 标尺列1：列头 */}
                <div className="grid-cell row-label">日期</div>
                {pillars.map((p, i) => <div key={i} className={`grid-cell col-header ${i === 2 ? 'is-day-pillar' : ''}`}>{p.label}</div>)}

                {/* 主星行 */}
                <div className="grid-cell row-label">主星</div>
                {pillars.map((p, i) => <div key={i} className="grid-cell top-shishen text-muted">{p.ganShiShen}</div>)}

                {/* 天干行 */}
                <div className="grid-cell row-label">天干</div>
                {pillars.map((p, i) => (
                  <div key={i} className={`grid-cell main-char gan wuxing-text-${WUXING_MAP[p.ganWx] || 'jin'}`}>
                    <span>{p.gan}</span><span className="wx-tag">{p.ganWx}</span>
                  </div>
                ))}

                {/* 地支行 */}
                <div className="grid-cell row-label">地支</div>
                {pillars.map((p, i) => (
                  <div key={i} className={`grid-cell main-char zhi wuxing-text-${WUXING_MAP[p.zhiWx] || 'jin'}`}>
                    <span>{p.zhi}</span><span className="wx-tag">{p.zhiWx}</span>
                  </div>
                ))}

                {/* 藏干行 */}
                <div className="grid-cell row-label">藏干</div>
                {pillars.map((p, i) => (
                  <div key={i} className="grid-cell hide-gan-cell">
                    {p.hideGan.map((g, idx) => (
                       <div key={idx} className={`hg-row wuxing-text-${WUXING_MAP['TODO'] || 'shui'}`} style={{ color: 'var(--text-color)' }}>{g}</div>
                    ))}
                  </div>
                ))}

                {/* 副星行 */}
                <div className="grid-cell row-label">副星</div>
                {pillars.map((p, i) => (
                  <div key={i} className="grid-cell hide-gan-cell text-muted">
                    {p.zhiShiShen.map((ss, idx) => <div key={idx} className="hg-row">{ss}</div>)}
                  </div>
                ))}

                {/* 星运行 */}
                <div className="grid-cell row-label">星运</div>
                {pillars.map((p, i) => <div key={i} className="grid-cell text-muted">{p.xingYun || p.diShi}</div>)}

                {/* 自坐行 */}
                <div className="grid-cell row-label">自坐</div>
                {pillars.map((p, i) => <div key={i} className="grid-cell text-muted">{p.diShi}</div>)}

                {/* 空亡行 */}
                <div className="grid-cell row-label">空亡</div>
                {pillars.map((p, i) => <div key={i} className="grid-cell text-muted">{p.xunKong}</div>)}

                {/* 纳音行 */}
                <div className="grid-cell row-label">纳音</div>
                {pillars.map((p, i) => <div key={i} className="grid-cell text-muted nayin">{p.naYin}</div>)}

                {/* 神煞行 */}
                <div className="grid-cell row-label shensha-label">神煞</div>
                {pillars.map((p, i) => (
                  <div key={i} className="grid-cell shensha-cell">
                    {p.shenSha.map((sh, idx) => (
                      <span key={idx} className="shensha-tag">{sh}</span>
                    ))}
                  </div>
                ))}

              </div>
            </div>

            {/* 五行雷达图 */}
            <div className="wuxing-section card">
              <h2 className="section-title serif">五行分布</h2>
              <WuxingRadar wuxing={result.wuxing} />
            </div>

            {/* 喜用神命元特质 */}
            <div className="yongshen-section">
              <h2 className="section-title serif">命元特质</h2>
              <YongshenBadge yongshen={result.yongshen || ''} jishen={result.jishen || ''} />
            </div>

            {/* 命理专属头像 */}
            <div className="mingpan-avatar-section card">
              <h2 className="section-title serif">✦ 专属命理头像</h2>
              <p className="section-desc">根据你的喜用神五行，程序化生成专属命元图腾</p>
              <MingpanAvatar
                yongshen={result.yongshen || ''}
                jishen={result.jishen || ''}
                dayGan={result.day_gan || ''}
              />
            </div>

            {/* 大运时间轴 */}
            <div className="dayun-section card">
              <h2 className="section-title serif">大运时间轴</h2>
              <DayunTimeline dayun={result.dayun} birthYear={result.birth_year} startYunSolar={result.start_yun_solar} dayGan={result.day_gan || ''} chartId={targetId} />
            </div>
          </div>

        {/* AI 解读区域 */}
        <div className="report-section card animate-fade-up">
          <div className="report-section-header">
            <h2 className="section-title serif">✦ AI 命理解读</h2>
            {report && (
              <div style={{ display: 'flex', gap: 8, flexWrap: 'wrap' }}>
                <button
                  id="save-card-btn"
                  className="btn btn-ghost btn-sm"
                  onClick={handleSaveImage}
                  disabled={savingImage}
                >
                  {savingImage ? '生成中...' : '🖼️ 保存分享图'}
                </button>
                {/* PDF 导出移动端不支持，仅桂面展示 */}
                {!isMobileDevice && (
                  <button
                    id="export-report-btn"
                    className="btn btn-ghost btn-sm"
                    onClick={() => window.print()}
                  >
                    📄 导出 PDF
                  </button>
                )}
              </div>
            )}
          </div>

          {/* 已有报告 */}
          {report && (
            <div className="report-sections">
              {/* 精简/专业切换按钮（仅新格式报告显示） */}
              {structured && (
                <div className="report-mode-switcher">
                  <button
                    className={`mode-btn${reportMode === 'brief' ? ' active' : ''}`}
                    onClick={() => setReportMode('brief')}
                  >精简</button>
                  <button
                    className={`mode-btn${reportMode === 'detail' ? ' active' : ''}`}
                    onClick={() => setReportMode('detail')}
                  >专业</button>
                </div>
              )}

              {/* 新格式：结构化渲染 */}
              {structured ? (
                <>
                  {reportMode === 'detail' && structured.analysis?.logic && (
                    <div className="report-block report-analysis">
                      <h3 className="report-block-title serif">🔍 命局分析总览</h3>
                      <p className="report-block-content">{structured.analysis.logic}</p>
                    </div>
                  )}
                  {reportMode === 'brief' && structured.analysis?.summary && (
                    <div className="report-summary">
                      <span className="report-summary-icon">✦</span>
                      <span>{structured.analysis.summary}</span>
                    </div>
                  )}
                  {(structured.chapters || []).map((ch, i) => (
                    <div key={i} className="report-block">
                      <h3 className="report-block-title serif">【{ch.title}】</h3>
                      <p className="report-block-content">
                        {reportMode === 'brief' ? ch.brief : ch.detail}
                      </p>
                    </div>
                  ))}
                </>
              ) : (
                /* 旧格式：降级渲染 */
                reportSections.length > 0 ? reportSections.map((sec, i) => (
                  <div key={i} className="report-block">
                    <h3 className="report-block-title serif">{sec.title}</h3>
                    <p className="report-block-content">{sec.content}</p>
                  </div>
                )) : (
                  <div className="report-content">{report.content}</div>
                )
              )}
              <p className="report-disclaimer">
                本报告由 AI 辅助生成，内容仅供参考，不构成任何决策建议。
              </p>
            </div>
          )}

          {/* 生成中 */}
          {reportLoading && (
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
          {reportError && !reportLoading && (
            <p className="form-error" style={{ margin: '12px 0' }}>⚠ {reportError}</p>
          )}

          {/* 未生成：显示按钮或引导 */}
          {!report && !reportLoading && (
            <>
              {!isGuest ? (
                <div className="report-cta">
                  <p className="report-cta-desc">
                    点击下方按钮，AI 将根据你的命盘生成性格、感情、事业、健康四维解读
                  </p>
                  <button
                    id="generate-ai-report"
                    className="btn btn-primary"
                    onClick={handleGenerateReport}
                  >
                    ✨ 生成 AI 命理解读
                  </button>
                </div>
              ) : (
                <div className="guest-banner">
                  <span>💡 登录后可获得完整 AI 解读报告，并保存命盘记录</span>
                  <a href="/register" className="btn btn-primary btn-sm">立即注册</a>
                </div>
              )}
            </>
          )}
        </div>

        <div className="result-footer">
          <button className="btn btn-ghost" onClick={() => navigate('/')}>← 重新起盘</button>
          {user && <a href="/history" className="btn btn-ghost">查看历史记录</a>}
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
    </div>
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
