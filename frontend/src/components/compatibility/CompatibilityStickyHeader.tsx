import './CompatibilityStickyHeader.css'

type Props = {
  selfName: string
  partnerName: string
  overallScore: number
  verdict: string
}

export default function CompatibilityStickyHeader({ selfName, partnerName, overallScore, verdict }: Props) {
  return (
    <header className="compat-sticky-header" aria-label="合盘摘要">
      <div className="compat-sticky-header__left">
        <span className="compat-sticky-header__names">{selfName} × {partnerName}</span>
        <a
          href="#compat-section-verdict"
          className="compat-sticky-header__verdict"
          aria-label="跳到判断详情"
        >
          {verdict}
        </a>
      </div>
      <div className="compat-sticky-header__right">
        <span className="compat-sticky-header__score serif">{overallScore}</span>
        <span className="compat-sticky-header__score-max">/100</span>
      </div>
    </header>
  )
}
