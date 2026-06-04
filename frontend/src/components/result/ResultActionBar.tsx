import { Button } from '../ui/Button'
import './ResultActionBar.css'

interface ResultActionBarProps {
  hasReport: boolean
  reportLoading?: boolean
  exportingPDF?: boolean
  isGuest?: boolean
  targetId?: string
  onGenerateReport: () => void
  onExportPDF: () => void
}

export function ResultActionBar({
  hasReport,
  reportLoading = false,
  exportingPDF = false,
  isGuest = false,
  targetId,
  onGenerateReport,
  onExportPDF,
}: ResultActionBarProps) {
  const pastEventsHref = targetId ? `/bazi/${targetId}/past-events` : undefined

  return (
    <div className="result-action-bar" aria-label="结果页主要操作">
      <div className="result-action-bar__primary">
        {hasReport ? (
          <Button id="result-primary-ai-action" href="#result-section-ai" variant="primary" size="lg">
            查看 AI 解读
          </Button>
        ) : (
          <Button
            id="result-primary-ai-action"
            type="button"
            variant="primary"
            size="lg"
            loading={reportLoading}
            onClick={onGenerateReport}
            disabled={reportLoading || isGuest}
          >
            {isGuest ? '登录后生成 AI 解读' : '生成 AI 解读'}
          </Button>
        )}
      </div>
      <div className="result-action-bar__secondary">
        {pastEventsHref ? (
          <Button href={pastEventsHref} variant="secondary">查看过往事件</Button>
        ) : (
          <Button type="button" variant="secondary" disabled>查看过往事件</Button>
        )}
        <Button type="button" variant="ghost" onClick={onExportPDF} disabled={!hasReport || exportingPDF} loading={exportingPDF}>
          导出 PDF
        </Button>
      </div>
      <div className="result-mobile-primary-action">
        {hasReport ? (
          <Button href="#result-section-ai" variant="primary">查看 AI 解读</Button>
        ) : (
          <Button
            type="button"
            variant="primary"
            loading={reportLoading}
            onClick={onGenerateReport}
            disabled={reportLoading || isGuest}
          >
            {isGuest ? '登录后生成 AI 解读' : '生成 AI 解读'}
          </Button>
        )}
      </div>
    </div>
  )
}
