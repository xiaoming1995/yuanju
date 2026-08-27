import type { LiuYueItem } from '../lib/api'
import {
  buildLiuYueTrendPath,
  liuyueTrendX,
  liuyueTrendY,
  type LiuYueTrendLevel,
  type LiuYueTrendSeries,
} from '../lib/liuyueTrend'
import './LiunianExportLayout.css'

export interface LiunianReportContent {
  career?: string
  romance?: string
  health?: string
  advice?: string
  monthly_notes?: LiunianMonthlyNote[]
}

export interface LiunianMonthlyNote {
  index: number
  month_label?: string
  liuyue_name?: string
  gan_zhi?: string
  summary?: string
  career?: string
  romance?: string
  health?: string
}

interface LiunianExportLayoutProps {
  variant: 'image' | 'pdf'
  year: number
  liuNianGanZhi: string
  gender?: string
  dayGan: string
  dayunLabel?: string
  report: LiunianReportContent
  trendSeries: LiuYueTrendSeries[]
  months: LiuYueItem[]
  currentMonthIndex?: number
}

const TREND_LEVEL_COLORS: Record<LiuYueTrendLevel, string> = {
  3: '#73b88a',
  2: '#c1ad74',
  1: '#c87968',
}

const TREND_LINE_COLORS: Record<string, string> = {
  overall: '#c9a84c',
  career: '#a88632',
  marriage: '#c87968',
  wealth: '#73a985',
  health: '#7394c8',
}

const TREND_LABELS: Record<LiuYueTrendLevel, string> = {
  3: '顺势',
  2: '平稳',
  1: '留意',
}

function genderLabel(gender?: string) {
  if (gender === 'male') return '男命'
  if (gender === 'female') return '女命'
  return '命主'
}

function fmtDate(date: string) {
  const parts = date.split('-')
  if (parts.length !== 3) return date
  return `${parseInt(parts[1])}/${parseInt(parts[2])}`
}

function fmtMonthLabel(date: string) {
  const parts = date.split('-')
  if (parts.length !== 3) return date
  return `${parseInt(parts[1])}月`
}

function compactText(value?: string, max = 90) {
  const text = (value || '暂无内容').replace(/\s+/g, ' ').trim()
  if (text.length <= max) return text
  return `${text.slice(0, max)}...`
}

function monthlyNoteForItem(notes: LiunianMonthlyNote[] | undefined, item: LiuYueItem) {
  return notes?.find(note => note.index === item.index)
}

function MonthlyNoteCard({ item, note, currentMonthIndex }: {
  item: LiuYueItem
  note: LiunianMonthlyNote
  currentMonthIndex: number
}) {
  return (
    <div className="liunian-export-month-note">
      <div className="liunian-export-month-note-head">
        <div>
          <div className="liunian-export-month-note-title">
            {note.month_label || fmtMonthLabel(item.start_date)} · {note.liuyue_name || item.month_name}
          </div>
          <div className="liunian-export-month-note-meta">
            {note.gan_zhi || item.gan_zhi} · {fmtDate(item.start_date)} - {fmtDate(item.end_date)}
          </div>
        </div>
        {item.index === currentMonthIndex && <span className="liunian-export-month-current">当前</span>}
      </div>
      <div className="liunian-export-month-note-summary">
        {note.summary || '本月宜结合流月干支与现实节奏，稳妥安排重要事项。'}
      </div>
      <div className="liunian-export-month-note-dims">
        {note.career && <div><span>事业</span>{note.career}</div>}
        {note.romance && <div><span>感情</span>{note.romance}</div>}
        {note.health && <div><span>健康</span>{note.health}</div>}
      </div>
    </div>
  )
}

function TrendChart({ series, wide = false }: { series: LiuYueTrendSeries; wide?: boolean }) {
  return (
    <div className={`liunian-export-trend-card ${wide ? 'liunian-export-trend-card--wide' : ''}`}>
      <div className="liunian-export-trend-head">
        <div className="liunian-export-trend-title">{series.title}</div>
        <div className="liunian-export-trend-summary">{series.summary}</div>
      </div>
      <svg className="liunian-export-trend-svg" viewBox="0 0 420 148" role="img" aria-label={`${series.title}流月趋势图`}>
        {[3, 2, 1].map(level => (
          <g key={level}>
            <text x="12" y={liuyueTrendY(level as LiuYueTrendLevel) + 4} fill="#8a7e68" fontSize="11">
              {TREND_LABELS[level as LiuYueTrendLevel]}
            </text>
            <line
              x1="48"
              x2="360"
              y1={liuyueTrendY(level as LiuYueTrendLevel)}
              y2={liuyueTrendY(level as LiuYueTrendLevel)}
              stroke="rgba(70, 56, 30, 0.12)"
              strokeWidth="1"
            />
          </g>
        ))}
        <path
          d={buildLiuYueTrendPath(series.points)}
          fill="none"
          stroke={TREND_LINE_COLORS[series.key] || '#c9a84c'}
          strokeWidth="3"
          strokeLinecap="round"
          strokeLinejoin="round"
        />
        {series.points.map((point, index) => {
          const x = liuyueTrendX(index, series.points.length)
          const y = liuyueTrendY(point.level)
          return (
            <g key={point.index}>
              <circle cx={x} cy={y} r="5" fill="#fffaf0" stroke={TREND_LEVEL_COLORS[point.level]} strokeWidth="3" />
              <text x={x} y="136" textAnchor="middle" fill="#8a7e68" fontSize="10" fontWeight="700">
                {fmtMonthLabel(point.startDate)}
              </text>
            </g>
          )
        })}
      </svg>
    </div>
  )
}

function ReportHeader({
  year,
  liuNianGanZhi,
  gender,
  dayGan,
  dayunLabel,
}: Pick<LiunianExportLayoutProps, 'year' | 'liuNianGanZhi' | 'gender' | 'dayGan' | 'dayunLabel'>) {
  return (
    <header className="liunian-export-header">
      <div>
        <div className="liunian-export-kicker">缘聚命理 · 流年运势报告</div>
        <div className="liunian-export-title">{year} {liuNianGanZhi}年</div>
        <div className="liunian-export-meta">
          {genderLabel(gender)} · {dayGan || '未知'}日主{dayunLabel ? ` · ${dayunLabel}` : ''}
        </div>
      </div>
      <div className="liunian-export-badge">年度精批</div>
    </header>
  )
}

function AdviceBlock({ advice }: { advice?: string }) {
  return (
    <section className="liunian-export-advice">
      <div className="liunian-export-advice-label">年度锦囊</div>
      <div className="liunian-export-advice-text">{advice || '本年宜稳住节奏，结合原局与现实情况逐步推进。'}</div>
    </section>
  )
}

function ExportFooter() {
  return (
    <>
      <div className="liunian-export-note">
        说明：流月趋势只表示当前流年内部的相对起伏，不代表不同大运或不同年份之间的绝对好坏。同样上行，弱年里可能只是压力减轻，好年里可能是机会放大。
      </div>
      <footer className="liunian-export-footer">本报告由 AI 辅助生成，仅供参考，不构成任何决策建议。</footer>
    </>
  )
}

export default function LiunianExportLayout({
  variant,
  year,
  liuNianGanZhi,
  gender,
  dayGan,
  dayunLabel,
  report,
  trendSeries,
  months,
  currentMonthIndex = -1,
}: LiunianExportLayoutProps) {
  const isImage = variant === 'image'
  const displaySeries = isImage ? trendSeries.slice(0, 1) : trendSeries
  const chapters = [
    { title: '事业财运', text: report.career },
    { title: '感情桃花', text: report.romance },
    { title: '健康风险', text: report.health },
  ]
  const monthlyNoteItems = months
    .map(item => ({ item, note: monthlyNoteForItem(report.monthly_notes, item) }))
    .filter((entry): entry is { item: LiuYueItem; note: LiunianMonthlyNote } => Boolean(entry.note))

  if (!isImage) {
    const overallSeries = trendSeries.find(series => series.key === 'overall')
    const secondarySeries = trendSeries.filter(series => series.key !== 'overall')
    const monthlyNotePages = [
      monthlyNoteItems.slice(0, 6),
      monthlyNoteItems.slice(6, 12),
    ].filter(page => page.length > 0)

    return (
      <article className="liunian-export-layout liunian-export-layout--pdf">
        <section className="liunian-export-page liunian-export-page--overview">
          <ReportHeader
            year={year}
            liuNianGanZhi={liuNianGanZhi}
            gender={gender}
            dayGan={dayGan}
            dayunLabel={dayunLabel}
          />
          <AdviceBlock advice={report.advice} />
          <section className="liunian-export-section liunian-export-section--compact">
            <div className="liunian-export-section-kicker">四项精批</div>
            <div className="liunian-export-section-title">这一年重点看什么</div>
            <div className="liunian-export-chapters">
              {chapters.map(chapter => (
                <div key={chapter.title} className="liunian-export-chapter">
                  <div className="liunian-export-chapter-title">{chapter.title}</div>
                  <div className="liunian-export-chapter-text">{chapter.text || '暂无内容'}</div>
                </div>
              ))}
            </div>
          </section>
          <div className="liunian-export-page-footer">01 / 年度总览</div>
        </section>

        <section className="liunian-export-page liunian-export-page--trends">
          <div className="liunian-export-section-kicker">年内趋势</div>
          <div className="liunian-export-section-title">五项流月趋势</div>
          <div className="liunian-export-trend-grid">
            {overallSeries && <TrendChart series={overallSeries} wide />}
            {secondarySeries.map(series => (
              <TrendChart key={series.key} series={series} />
            ))}
          </div>
          <div className="liunian-export-page-footer">02 / 年内趋势</div>
        </section>

        {monthlyNotePages.length > 0 && monthlyNotePages.map((pageItems, pageIndex) => (
          <section key={pageIndex} className="liunian-export-page liunian-export-page--month-notes">
            <div className="liunian-export-section-kicker">月度重点</div>
            <div className="liunian-export-section-title">
              十二流月注意点{pageIndex === 0 ? '（上）' : '（下）'}
            </div>
            <div className="liunian-export-month-note-grid">
              {pageItems.map(({ item, note }) => (
                <MonthlyNoteCard
                  key={item.index}
                  item={item}
                  note={note}
                  currentMonthIndex={currentMonthIndex}
                />
              ))}
            </div>
            {pageIndex === monthlyNotePages.length - 1 && <ExportFooter />}
            <div className="liunian-export-page-footer">
              {String(3 + pageIndex).padStart(2, '0')} / 月度重点
            </div>
          </section>
        ))}

        {monthlyNotePages.length === 0 && months.length > 0 && (
          <section className="liunian-export-page liunian-export-page--months">
            <div className="liunian-export-section-kicker">流月简表</div>
            <div className="liunian-export-section-title">十二个月令节奏</div>
            <div className="liunian-export-months liunian-export-months--pdf">
              {months.map(item => (
                <div key={item.index} className="liunian-export-month">
                  <div className="liunian-export-month-top">
                    <span className="liunian-export-month-name">
                      {fmtMonthLabel(item.start_date)} · {item.month_name}
                    </span>
                    {item.index === currentMonthIndex && <span className="liunian-export-month-current">当前</span>}
                  </div>
                  <div className="liunian-export-month-main">
                    <div className="liunian-export-month-gz">{item.gan_zhi}</div>
                    <div>
                      <div className="liunian-export-month-range">{fmtDate(item.start_date)} - {fmtDate(item.end_date)}</div>
                      <div className="liunian-export-month-gods">{item.gan_shishen} · {item.zhi_shishen}</div>
                    </div>
                  </div>
                </div>
              ))}
            </div>
            <ExportFooter />
            <div className="liunian-export-page-footer">03 / 流月简表</div>
          </section>
        )}
      </article>
    )
  }

  return (
    <article className={`liunian-export-layout liunian-export-layout--${variant}`}>
      <ReportHeader
        year={year}
        liuNianGanZhi={liuNianGanZhi}
        gender={gender}
        dayGan={dayGan}
        dayunLabel={dayunLabel}
      />
      <AdviceBlock advice={report.advice} />

      <section className="liunian-export-section">
        <div className="liunian-export-section-kicker">四项精批</div>
        <div className="liunian-export-section-title">这一年重点看什么</div>
        <div className="liunian-export-chapters">
          {chapters.map(chapter => (
            <div key={chapter.title} className="liunian-export-chapter">
              <div className="liunian-export-chapter-title">{chapter.title}</div>
              <div className="liunian-export-chapter-text">
                {isImage ? compactText(chapter.text) : chapter.text || '暂无内容'}
              </div>
            </div>
          ))}
        </div>
      </section>

      {displaySeries.length > 0 && (
        <section className="liunian-export-section">
          <div className="liunian-export-section-kicker">年内趋势</div>
          <div className="liunian-export-section-title">{isImage ? '综合流月起伏' : '五项流月趋势'}</div>
          <div className="liunian-export-trend-grid">
            {displaySeries.map(series => (
              <TrendChart key={series.key} series={series} wide={isImage || series.key === 'overall'} />
            ))}
          </div>
        </section>
      )}

      {monthlyNoteItems.length > 0 && (
        <section className="liunian-export-section">
          <div className="liunian-export-section-kicker">月度重点</div>
          <div className="liunian-export-section-title">流月注意点摘录</div>
          <div className="liunian-export-month-note-grid liunian-export-month-note-grid--image">
            {monthlyNoteItems.slice(0, 3).map(({ item, note }) => (
              <MonthlyNoteCard
                key={item.index}
                item={item}
                note={{ ...note, summary: compactText(note.summary, 72) }}
                currentMonthIndex={currentMonthIndex}
              />
            ))}
          </div>
        </section>
      )}

      {!isImage && months.length > 0 && (
        <section className="liunian-export-section">
          <div className="liunian-export-section-kicker">流月简表</div>
          <div className="liunian-export-section-title">十二个月令节奏</div>
          <div className="liunian-export-months">
            {months.map(item => (
              <div key={item.index} className="liunian-export-month">
                <div className="liunian-export-month-top">
                  <span className="liunian-export-month-name">
                    {fmtMonthLabel(item.start_date)} · {item.month_name}
                  </span>
                  {item.index === currentMonthIndex && <span className="liunian-export-month-range">当前</span>}
                </div>
                <div className="liunian-export-month-gz">{item.gan_zhi}</div>
                <div className="liunian-export-month-range">{fmtDate(item.start_date)} - {fmtDate(item.end_date)}</div>
                <div className="liunian-export-month-gods">{item.gan_shishen} · {item.zhi_shishen}</div>
              </div>
            ))}
          </div>
        </section>
      )}

      <ExportFooter />
    </article>
  )
}
