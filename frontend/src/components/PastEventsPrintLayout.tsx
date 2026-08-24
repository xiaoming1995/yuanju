import type { ExportBrand } from '../lib/api'
import type { PastEventsExportReadySegment } from '../lib/pastEventsViewModel'
import { getPastEventsSignalLabel } from '../lib/pastEventsViewModel'
import { resolveFooter, showDiagonalWatermark } from '../lib/brandText'
import './PastEventsPrintLayout.css'

export interface PastEventsPrintChartContext {
  id?: string
  display_name?: string
  birth_year?: number
  birth_month?: number
  birth_day?: number
  birth_hour?: number
  gender?: string
}

interface PastEventsPrintSectionProps {
  segments: PastEventsExportReadySegment[]
  includeEvidence?: boolean
  compact?: boolean
}

interface PastEventsPrintLayoutProps extends PastEventsPrintSectionProps {
  chart?: PastEventsPrintChartContext | null
  brand?: ExportBrand | null
}

function sectionTitle(text: string) {
  return (
    <div className="past-events-print-section-title">
      <div />
      <span>{text}</span>
      <div />
    </div>
  )
}

function chartLabel(chart?: PastEventsPrintChartContext | null) {
  if (!chart) return ''
  const name = chart.display_name?.trim()
  const birth = chart.birth_year
    ? `${chart.birth_year}年${chart.birth_month || ''}月${chart.birth_day || ''}日${chart.birth_hour ?? ''}时`
    : ''
  const gender = chart.gender === 'male' ? '男命' : chart.gender === 'female' ? '女命' : ''
  return [name, birth, gender].filter(Boolean).join(' · ')
}

export function PastEventsPrintSection({ segments, includeEvidence = false, compact = false }: PastEventsPrintSectionProps) {
  if (!segments.length) return null

  return (
    <section className={`past-events-print-section${compact ? ' past-events-print-section--compact' : ''}`}>
      {sectionTitle('过　往　年　运　回　看')}
      {segments.map((segment) => (
        <div className="past-events-print-segment" key={segment.dayun_index}>
          <div className="past-events-print-segment-head">
            <strong>{segment.gan_zhi}</strong>
            <span>
              {segment.start_age !== undefined && segment.end_age !== undefined
                ? `${segment.start_age}-${segment.end_age}岁`
                : ''}
              {segment.start_year !== undefined && segment.end_year !== undefined
                ? `（${segment.start_year}-${segment.end_year}年）`
                : ''}
            </span>
          </div>
          {segment.themes.length > 0 && (
            <div className="past-events-print-themes">
              {segment.themes.map((theme) => <span key={theme}>{theme}</span>)}
            </div>
          )}
          <p className="past-events-print-summary">{segment.summary}</p>
          <div className="past-events-print-years">
            {segment.years.map((year) => (
              <div className="past-events-print-year" key={`${segment.dayun_index}-${year.year}`}>
                <div className="past-events-print-year-head">
                  <strong>{year.gan_zhi}</strong>
                  <span>{year.year}年{year.age ? ` · ${year.age}岁` : ''}</span>
                </div>
                <p>{year.narrative}</p>
                {includeEvidence && (year.signals?.length || year.evidence_summary?.length) && (
                  <div className="past-events-print-evidence">
                    {year.signals?.length ? (
                      <div className="past-events-print-signal-row">
                        {year.signals.map((signal) => (
                          <span key={signal}>{getPastEventsSignalLabel(signal).label}</span>
                        ))}
                      </div>
                    ) : null}
                    {year.evidence_summary?.length ? (
                      <ul>
                        {year.evidence_summary.map((evidence, idx) => (
                          <li key={`${year.year}-evidence-${idx}`}>{evidence}</li>
                        ))}
                      </ul>
                    ) : null}
                  </div>
                )}
              </div>
            ))}
          </div>
        </div>
      ))}
    </section>
  )
}

export default function PastEventsPrintLayout({
  chart,
  segments,
  includeEvidence = false,
  brand,
}: PastEventsPrintLayoutProps) {
  const title = brand?.title?.trim() || '缘 聚 命 理'
  const footer = resolveFooter(brand, '缘 聚 命 理')
  const showDiagonalMark = showDiagonalWatermark(brand)
  const context = chartLabel(chart)

  return (
    <div className="print-only past-events-print-layout">
      {showDiagonalMark && brand && (
        <div className="print-diagonal-watermark" aria-hidden="true">
          {Array.from({ length: 120 }).map((_, i) => (
            <span key={i}>{brand.watermark_text}</span>
          ))}
        </div>
      )}
      <table className="print-page-table">
        <thead>
          <tr>
            <td>
              <div className="past-events-print-page-header">
                <span>{title}</span>
                {context && <span>{context}</span>}
              </div>
              <div className="print-page-header-spacer" />
            </td>
          </tr>
        </thead>
        <tbody>
          <tr>
            <td>
              <main className="past-events-print-body">
                <header className="past-events-print-cover">
                  <div>YUAN JU MING LI</div>
                  <h1>过往年运回看</h1>
                  {context && <p>{context}</p>}
                </header>
                <PastEventsPrintSection segments={segments} includeEvidence={includeEvidence} />
                <footer className="past-events-print-footer">
                  <span>本推算内容仅供参考，不构成任何决策建议。</span>
                  <span>{footer}</span>
                </footer>
              </main>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  )
}
