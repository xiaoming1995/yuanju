import './GongJiaPanel.css'

export interface GongJiaItem {
  source: string
  source_labels?: string[]
  same_gan: string
  source_zhis?: string[]
  virtual_zhi: string
  hide_gan?: string[]
  shishen?: string[]
  shensha?: string[]
  meaning?: string
}

interface GongJiaPanelProps {
  items?: GongJiaItem[]
  onShenshaClick?: (name: string) => void
  hasShenshaAnnotation?: (name: string) => boolean
  shenshaPolarity?: (name: string) => string
}

function joinList(items?: string[]) {
  return items && items.length > 0 ? items.join('、') : '未标注'
}

export default function GongJiaPanel({
  items,
  onShenshaClick,
  hasShenshaAnnotation,
  shenshaPolarity,
}: GongJiaPanelProps) {
  if (!items || items.length === 0) return null

  return (
    <section className="gongjia-panel card" aria-labelledby="gongjia-title">
      <div className="gongjia-panel__header">
        <div>
          <span className="gongjia-panel__kicker">暗藏信号</span>
          <h2 id="gongjia-title" className="gongjia-panel__title">原局夹拱</h2>
        </div>
        <p>夹拱只提示原局可能牵动的暗藏虚支，不改原局五行与用神，也不作为第五柱处理。</p>
      </div>

      <div className="gongjia-panel__grid">
        {items.map((item, index) => {
          const shensha = item.shensha || []
          return (
            <article className="gongjia-card" key={`${item.source}-${item.virtual_zhi}-${index}`}>
              <div className="gongjia-card__topline">
                <span>{joinList(item.source_labels)}</span>
                <em>{item.source || '夹拱'}</em>
              </div>

              <div className="gongjia-card__main">
                <span className="gongjia-card__branch">{item.virtual_zhi}</span>
                <div>
                  <strong>暗藏虚支</strong>
                  <p>{item.meaning || '原局两支之间形成夹拱，可作为辅助信号观察，不直接改写主盘。'}</p>
                </div>
              </div>

              <dl className="gongjia-card__facts">
                <div>
                  <dt>同干</dt>
                  <dd>{item.same_gan || '未标注'}</dd>
                </div>
                <div>
                  <dt>来源地支</dt>
                  <dd>{joinList(item.source_zhis)}</dd>
                </div>
                <div>
                  <dt>藏干</dt>
                  <dd>{joinList(item.hide_gan)}</dd>
                </div>
                <div>
                  <dt>十神</dt>
                  <dd>{joinList(item.shishen)}</dd>
                </div>
              </dl>

              {shensha.length > 0 && (
                <div className="gongjia-card__shensha" aria-label="拱神煞">
                  <span className="gongjia-card__label">拱神煞</span>
                  <div className="gongjia-card__tags">
                    {shensha.map((name) => {
                      const hasAnnotation = hasShenshaAnnotation?.(name) || false
                      const polarity = shenshaPolarity?.(name) || 'zhong'
                      const className = `shensha-tag shensha-tag--${polarity}${hasAnnotation ? ' shensha-tag--clickable' : ''}`
                      return hasAnnotation ? (
                        <button
                          key={name}
                          type="button"
                          className={className}
                          onClick={() => onShenshaClick?.(name)}
                        >
                          {name}
                        </button>
                      ) : (
                        <span key={name} className={className}>
                          {name}
                        </span>
                      )
                    })}
                  </div>
                </div>
              )}
            </article>
          )
        })}
      </div>
    </section>
  )
}
