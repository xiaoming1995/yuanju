import React, { useState } from 'react'
import { Diamond } from 'lucide-react'
import './TiaohouCard.css'

interface TiaohouCardProps {
  dayGan: string
  monthZhi: string
  tiaohou: {
    expected: string[]
    tou: string[]
    cang: string[]
    text: string
  }
}

const TiaohouCard: React.FC<TiaohouCardProps> = ({ dayGan, monthZhi, tiaohou }) => {
  const [expanded, setExpanded] = useState(true)

  if (!tiaohou || !tiaohou.expected || tiaohou.expected.length === 0) {
    return null
  }

  // 计算高亮：如果理论用神在命局中出现（透或藏）则高亮
  const isMatched = (ys: string) => tiaohou.tou.includes(ys) || tiaohou.cang.includes(ys)

  return (
    <div className="tiaohou-card-container">
      <div className="tiaohou-header-section">
        <div className="tiaohou-expected-row">
          <span className="tiaohou-label"><Diamond size={14} className="title-diamond-icon" />调候用神提示 <span className="tiaohou-question-mark">?</span></span>
          <span className="tiaohou-expected-values">
            {tiaohou.expected.map((ys, idx) => (
              <span key={`expected-${idx}`} className={isMatched(ys) ? 'matched-text' : 'unmatched-text'}>
                {ys}
              </span>
            ))}
          </span>
        </div>
        
        <div className="tiaohou-actual-row">
          <span className="tiaohou-label">本八字</span>
          <div className="tiaohou-actual-values">
            <span className="actual-type">透</span>
            {tiaohou.tou.length > 0 ? (
              <div className="actual-boxes">
                {tiaohou.tou.map((t, i) => <span key={`tou-${i}`} className="actual-box">{t}</span>)}
              </div>
            ) : <span className="actual-empty">无</span>}
            
            <span className="actual-type actual-type-cang">藏</span>
            {tiaohou.cang.length > 0 ? (
              <div className="actual-boxes">
                {tiaohou.cang.map((c, i) => <span key={`cang-${i}`} className="actual-box">{c}</span>)}
              </div>
            ) : <span className="actual-empty">无</span>}
          </div>
        </div>
      </div>

      <div className="tiaohou-detail-section">
        <div className="tiaohou-detail-title" onClick={() => setExpanded(!expanded)}>
          <span>论{dayGan}生{monthZhi}月</span>
          <span className="expand-icon">{expanded ? '▲' : '▼'}</span>
        </div>
        
        {expanded && (
          <div className="tiaohou-detail-content">
            <p className="tiaohou-classic-text">{tiaohou.text}</p>
          </div>
        )}
      </div>
    </div>
  )
}

export default TiaohouCard
