import { useState } from 'react'
import { WUXING_MAP, parseWuxingList, mergeLuckyNumbers, mergeLuckyDirections, mergeLuckySeasons } from '../lib/wuxingColorSystem'
import type { WuxingKey } from '../lib/wuxingColorSystem'
import './YongshenBadge.css'

interface YongshenBadgeProps {
  yongshen: string
  jishen: string
}

interface WuxingTagProps {
  wx: WuxingKey
}

function WuxingTag({ wx }: WuxingTagProps) {
  const [expanded, setExpanded] = useState(false)
  const profile = WUXING_MAP[wx]

  return (
    <div className="wuxing-tag-wrapper">
      <button
        className="wuxing-tag"
        style={{
          '--tag-color': profile.color,
          '--tag-light': profile.lightColor,
        } as React.CSSProperties}
        onClick={() => setExpanded(prev => !prev)}
        aria-expanded={expanded}
      >
        <span className="wuxing-tag-dot" style={{ background: profile.color }} />
        <span className="wuxing-tag-emoji">{profile.emoji}</span>
        <span className="wuxing-tag-char">{wx}</span>
        <span className="wuxing-tag-chevron">{expanded ? '▴' : '▾'}</span>
      </button>
      {expanded && (
        <div className="wuxing-tag-desc animate-fade-in">
          {profile.description}
        </div>
      )}
    </div>
  )
}

export default function YongshenBadge({ yongshen, jishen }: YongshenBadgeProps) {
  const yongList = parseWuxingList(yongshen)
  const jiList = parseWuxingList(jishen)
  const hasData = yongList.length > 0 || jiList.length > 0

  const luckyNumbers = mergeLuckyNumbers(yongList)
  const luckyDirections = mergeLuckyDirections(yongList)
  const luckySeasons = mergeLuckySeasons(yongList)

  // 空数据：骨架屏占位
  if (!hasData) {
    return (
      <div className="yongshen-badge yongshen-badge--empty card">
        <div className="yongshen-badge-lock">
          <span className="yongshen-badge-lock-icon">🔮</span>
          <p className="yongshen-badge-lock-text">生成 AI 报告后可解锁命元特质</p>
        </div>
      </div>
    )
  }

  return (
    <div className="yongshen-badge card">
      <h3 className="yongshen-badge-title serif">✦ 命元特质</h3>

      {/* 喜用神 */}
      {yongList.length > 0 && (
        <div className="yongshen-row">
          <span className="yongshen-row-label yongshen-row-label--xi">喜</span>
          <div className="yongshen-tags">
            {yongList.map(wx => <WuxingTag key={wx} wx={wx} />)}
          </div>
          <div className="yongshen-color-strip">
            {yongList.map(wx => (
              <span
                key={wx}
                className="color-strip-block"
                style={{ background: WUXING_MAP[wx].color }}
                title={wx}
              />
            ))}
          </div>
        </div>
      )}

      {/* 忌神 */}
      {jiList.length > 0 && (
        <div className="yongshen-row">
          <span className="yongshen-row-label yongshen-row-label--ji">忌</span>
          <div className="yongshen-tags">
            {jiList.map(wx => <WuxingTag key={wx} wx={wx} />)}
          </div>
          <div className="yongshen-color-strip">
            {jiList.map(wx => (
              <span
                key={wx}
                className="color-strip-block color-strip-block--ji"
                style={{ background: WUXING_MAP[wx].darkColor }}
                title={wx}
              />
            ))}
          </div>
        </div>
      )}

      {/* 幸运属性 */}
      {yongList.length > 0 && (
        <div className="yongshen-lucky">
          <div className="lucky-divider" />
          <div className="lucky-grid">
            <div className="lucky-item">
              <span className="lucky-label">幸运方位</span>
              <span className="lucky-value">{luckyDirections.join(' · ')}</span>
            </div>
            <div className="lucky-item">
              <span className="lucky-label">幸运数字</span>
              <span className="lucky-value">{luckyNumbers.join(' · ')}</span>
            </div>
            <div className="lucky-item">
              <span className="lucky-label">幸运季节</span>
              <span className="lucky-value">{luckySeasons.join(' · ')}</span>
            </div>
            <div className="lucky-item">
              <span className="lucky-label">幸运质感</span>
              <span className="lucky-value">
                {yongList.map(wx => WUXING_MAP[wx].material).join('、')}
              </span>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
