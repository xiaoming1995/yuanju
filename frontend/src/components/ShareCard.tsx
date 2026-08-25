import { forwardRef } from 'react'
import type { StructuredReport, ExportBrand } from '../lib/api'
import type { PastEventsExportReadySegment } from '../lib/pastEventsViewModel'
import type { DayunTrendSeries } from '../lib/dayunTrend'
import { buildTrendNote } from '../lib/dayunTrend'
import { resolveFooter, showDiagonalWatermark } from '../lib/brandText'

// ── 天干地支专属 Google Fonts 子集（仅22字，< 20KB，保障截图字体渲染） ──
const PILLAR_FONT_URL =
  'https://fonts.googleapis.com/css2?family=Noto+Serif+SC:wght@700&text=%E7%94%B2%E4%B9%99%E4%B8%99%E4%B8%81%E6%88%8A%E5%B7%B1%E5%BA%9A%E8%BE%9B%E5%A3%AC%E7%99%B8%E5%AD%90%E4%B8%91%E5%AF%85%E5%8D%AF%E8%BE%B0%E5%B7%B3%E5%8D%88%E6%9C%AA%E7%94%B3%E9%85%89%E6%88%8C%E4%BA%A5&display=swap'

// 五行对色（国风明色版本）
const WX_COLOR: Record<string, string> = {
  '木': '#4a7c59',
  '火': '#c0392b',
  '土': '#a0784a',
  '金': '#7a6830',
  '水': '#2c5282',
}

function wxColor(wxStr: string) {
  for (const [k, v] of Object.entries(WX_COLOR)) {
    if (wxStr?.startsWith(k)) return v
  }
  return '#5c4a3a'
}

// 分隔线组件
function Divider() {
  return (
    <div style={{
      height: 1,
      background: 'linear-gradient(to right, transparent, #d4b896, #c9a96e, #d4b896, transparent)',
      margin: '0 20px',
    }} />
  )
}

// 章节卡片
function ChapterBlock({ icon, title, content }: { icon: string; title: string; content: string }) {
  return (
    <div style={{ padding: '16px 24px' }}>
      <div style={{
        fontSize: 13,
        fontWeight: 700,
        color: '#7a5c3a',
        marginBottom: 8,
        letterSpacing: 2,
        display: 'flex',
        alignItems: 'center',
        gap: 6,
        fontFamily: '"Noto Serif SC", serif',
      }}>
        <span>{icon}</span>
        <span>{title}</span>
      </div>
      <div style={{
        fontSize: 13,
        color: '#4a3728',
        lineHeight: 1.85,
        fontFamily: '"Noto Sans SC", "PingFang SC", "Microsoft YaHei", sans-serif',
      }}>
        {content}
      </div>
    </div>
  )
}

function PastEventsShareSection({ segments }: { segments: PastEventsExportReadySegment[] }) {
  if (!segments.length) return null

  return (
    <>
      <Divider />
      <div style={{
        padding: '16px 24px 12px',
        background: '#fdf8f0',
      }}>
        <div style={{
          fontSize: 11,
          color: '#a08060',
          letterSpacing: 4,
          textAlign: 'center',
          fontFamily: '"Noto Serif SC", serif',
          marginBottom: 12,
        }}>
          ── 过 往 年 运 回 看 ──
        </div>
        {segments.map((segment) => (
          <div key={segment.dayun_index} style={{
            border: '1px solid #ead8b8',
            background: '#fffaf2',
            marginBottom: 10,
          }}>
            <div style={{
              display: 'flex',
              justifyContent: 'space-between',
              gap: 8,
              padding: '7px 9px',
              borderBottom: '1px solid #ead8b8',
              background: '#faf0df',
              alignItems: 'baseline',
            }}>
              <strong style={{
                color: '#7a5c3a',
                fontSize: 14,
                letterSpacing: 2,
                fontFamily: '"Noto Serif SC", serif',
              }}>
                {segment.gan_zhi}
              </strong>
              <span style={{
                color: '#9b815c',
                fontSize: 10,
                fontFamily: '"Noto Sans SC", sans-serif',
                whiteSpace: 'nowrap',
              }}>
                {segment.start_age !== undefined && segment.end_age !== undefined
                  ? `${segment.start_age}-${segment.end_age}岁`
                  : ''}
                {segment.start_year !== undefined && segment.end_year !== undefined
                  ? `（${segment.start_year}-${segment.end_year}年）`
                  : ''}
              </span>
            </div>
            {segment.themes.length > 0 && (
              <div style={{
                display: 'flex',
                flexWrap: 'wrap',
                gap: 4,
                padding: '7px 9px 0',
              }}>
                {segment.themes.map((theme) => (
                  <span key={theme} style={{
                    border: '1px solid #e0cca0',
                    background: '#fdf8f0',
                    color: '#7a6830',
                    borderRadius: 2,
                    padding: '1px 5px',
                    fontSize: 9,
                    lineHeight: 1.5,
                    fontFamily: '"Noto Sans SC", sans-serif',
                  }}>
                    {theme}
                  </span>
                ))}
              </div>
            )}
            <p style={{
              margin: '7px 9px 8px',
              color: '#4a3728',
              fontSize: 11,
              lineHeight: 1.65,
              fontFamily: '"Noto Sans SC", "PingFang SC", "Microsoft YaHei", sans-serif',
            }}>
              {segment.summary}
            </p>
            {segment.years.map((year) => (
              <div key={`${segment.dayun_index}-${year.year}`} style={{
                padding: '7px 9px',
                borderTop: '1px solid #f0e8d4',
                background: year.year % 2 === 0 ? '#fff' : '#fdf8f0',
              }}>
                <div style={{
                  display: 'flex',
                  gap: 6,
                  alignItems: 'baseline',
                  marginBottom: 3,
                }}>
                  <strong style={{
                    color: '#7a5c3a',
                    fontSize: 12,
                    fontFamily: '"Noto Serif SC", serif',
                  }}>
                    {year.gan_zhi}
                  </strong>
                  <span style={{
                    color: '#9b815c',
                    fontSize: 9,
                    fontFamily: '"Noto Sans SC", sans-serif',
                  }}>
                    {year.year}年{year.age ? ` · ${year.age}岁` : ''}
                  </span>
                </div>
                <p style={{
                  margin: 0,
                  color: '#4a3728',
                  fontSize: 10,
                  lineHeight: 1.6,
                  fontFamily: '"Noto Sans SC", "PingFang SC", "Microsoft YaHei", sans-serif',
                }}>
                  {year.narrative}
                </p>
              </div>
            ))}
          </div>
        ))}
      </div>
    </>
  )
}

function DayunTrendShareSection({
  series,
  label,
  period,
}: {
  series: DayunTrendSeries[]
  label?: string
  period?: string
}) {
  if (!series.length) return null

  return (
    <>
      <Divider />
      <div style={{
        padding: '16px 24px 14px',
        background: '#fffaf2',
      }}>
        <div style={{
          fontSize: 11,
          color: '#a08060',
          letterSpacing: 4,
          textAlign: 'center',
          fontFamily: '"Noto Serif SC", serif',
          marginBottom: 12,
        }}>
          ── 大 运 趋 势 ──
        </div>
        {(label || period) && (
          <div style={{
            marginBottom: 10,
            textAlign: 'center',
            color: '#7a5c3a',
            fontFamily: '"Noto Sans SC", sans-serif',
          }}>
            {label && <strong style={{ display: 'block', fontSize: 13 }}>{label}</strong>}
            {period && <span style={{ display: 'block', marginTop: 2, fontSize: 10, color: '#a08060' }}>{period}</span>}
          </div>
        )}
        <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
          {series.map(item => (
            <div key={item.key} style={{
              display: 'grid',
              gridTemplateColumns: '48px 1fr',
              columnGap: 10,
              alignItems: 'start',
              padding: '8px 10px',
              background: '#fdf8f0',
              borderRadius: 6,
            }}>
              <div style={{
                color: '#7a5c3a',
                fontSize: 13,
                fontWeight: 700,
                fontFamily: '"Noto Serif SC", serif',
                letterSpacing: 1,
              }}>
                {item.title}
              </div>
              <div style={{
                color: '#4a3728',
                fontSize: 10,
                lineHeight: 1.65,
                fontFamily: '"Noto Sans SC", "PingFang SC", sans-serif',
              }}>
                <span style={{ color: '#9b6b1f', fontWeight: 700 }}>{item.summary}</span>
                <span> · {buildTrendNote(item)}</span>
              </div>
            </div>
          ))}
        </div>
        <p style={{
          margin: '10px 0 0',
          color: '#a08060',
          fontSize: 9,
          lineHeight: 1.6,
          textAlign: 'center',
          fontFamily: '"Noto Sans SC", sans-serif',
        }}>
          本图为段内相对趋势：同样上行，坏运可能只是小幅改善，好运可能是机会放大，需结合大运整体判断。
        </p>
      </div>
    </>
  )
}

export interface ShareCardProps {
  birthYear: number
  birthMonth: number
  birthDay: number
  birthHour: number
  gender: string
  yearGan: string; yearZhi: string
  monthGan: string; monthZhi: string
  dayGan: string; dayZhi: string
  hourGan: string; hourZhi: string
  yearGanWx: string; yearZhiWx: string
  monthGanWx: string; monthZhiWx: string
  dayGanWx: string; dayZhiWx: string
  hourGanWx: string; hourZhiWx: string
  structured: StructuredReport | null
  brand?: ExportBrand | null
  pastEventsExportSegments?: PastEventsExportReadySegment[]
  dayunTrendSeries?: DayunTrendSeries[]
  dayunTrendLabel?: string
  dayunTrendPeriod?: string
}

const ShareCard = forwardRef<HTMLDivElement, ShareCardProps>((props, ref) => {
  const {
    birthYear, birthMonth, birthDay, birthHour, gender,
    yearGan, yearZhi, monthGan, monthZhi, dayGan, dayZhi, hourGan, hourZhi,
    yearGanWx, yearZhiWx, monthGanWx, monthZhiWx,
    dayGanWx, dayZhiWx, hourGanWx, hourZhiWx,
    structured,
    brand,
    pastEventsExportSegments = [],
    dayunTrendSeries = [],
    dayunTrendLabel,
    dayunTrendPeriod,
  } = props

  const pillars = [
    { label: '年', gan: yearGan, zhi: yearZhi, ganWx: yearGanWx, zhiWx: yearZhiWx },
    { label: '月', gan: monthGan, zhi: monthZhi, ganWx: monthGanWx, zhiWx: monthZhiWx },
    { label: '日', gan: dayGan,   zhi: dayZhi,   ganWx: dayGanWx,   zhiWx: dayZhiWx   },
    { label: '时', gan: hourGan,  zhi: hourZhi,  ganWx: hourGanWx,  zhiWx: hourZhiWx  },
  ]

  const analysis = structured?.analysis
  const chapters = structured?.chapters ?? []

  const chapterDefs = [
    { icon: '命', key: 'personality' },
    { icon: '情', key: 'romance' },
    { icon: '业', key: 'career' },
    { icon: '身', key: 'health' },
  ]

  const resolvedTitle = brand?.title || '缘 聚 命 理'
  const resolvedFooter = resolveFooter(brand, 'yuanju.com')
  const showDiagonalMark = showDiagonalWatermark(brand)
  const isWordmark = brand?.logo_mode === 'wordmark' && !!brand?.logo_url

  return (
    <div ref={ref} style={{
      width: 400,
      background: '#fdf9f2',
      fontFamily: '"Noto Serif SC", serif',
      overflow: 'hidden',
      boxSizing: 'border-box',
      position: 'relative',
    }}>
      {/* 天干地支专属字体 */}
      <style>{`@import url('${PILLAR_FONT_URL}');`}</style>

      {/* ┌ 顶部品牌栏 ── */}
      <div style={{
        background: 'linear-gradient(135deg, #2d1f14 0%, #4a3020 50%, #3a2416 100%)',
        padding: '20px 24px 18px',
        position: 'relative',
      }}>
        {isWordmark ? (
          <img
            src={brand!.logo_url}
            alt=""
            crossOrigin="anonymous"
            style={{
              display: 'block',
              margin: '0 auto 6px',
              maxHeight: 48,
              maxWidth: 320,
              objectFit: 'contain',
            }}
          />
        ) : (
          <>
            {brand?.logo_url && (
              <img
                src={brand.logo_url}
                alt=""
                crossOrigin="anonymous"
                style={{
                  position: 'absolute',
                  left: 24,
                  top: '50%',
                  transform: 'translateY(-50%)',
                  width: 40,
                  height: 40,
                  objectFit: 'contain',
                }}
              />
            )}
            <div style={{
              color: '#e8c97c',
              fontSize: 20,
              letterSpacing: 6,
              fontWeight: 700,
              textAlign: 'center',
              fontFamily: '"Noto Serif SC", serif',
              marginBottom: 6,
            }}>
              {resolvedTitle}
            </div>
          </>
        )}
        <div style={{
          color: '#c4a06a',
          fontSize: 12,
          letterSpacing: 1,
          textAlign: 'center',
          fontFamily: '"Noto Sans SC", sans-serif',
        }}>
          {birthYear}年{birthMonth}月{birthDay}日&nbsp;{birthHour}时 · {gender === 'male' ? '男命' : '女命'}
        </div>
      </div>

      {/* ┌ 四柱大展示 ── */}
      <div style={{
        background: '#faf5eb',
        padding: '24px 16px 20px',
        borderBottom: '1px solid #e8dcc8',
      }}>
        <div style={{
          display: 'grid',
          gridTemplateColumns: 'repeat(4, 1fr)',
          gap: 6,
        }}>
          {pillars.map((p, i) => (
            <div key={i} style={{
              display: 'flex',
              flexDirection: 'column',
              alignItems: 'center',
              padding: '14px 4px 10px',
              background: i === 2 ? 'rgba(201,168,76,0.1)' : 'rgba(255,255,255,0.6)',
              borderRadius: 10,
              border: i === 2 ? '1px solid rgba(201,168,76,0.4)' : '1px solid rgba(210,190,160,0.3)',
              gap: 2,
            }}>
              <span style={{ fontSize: 10, color: '#a08060', letterSpacing: 2 }}>{p.label}柱</span>
              <span style={{
                fontSize: 38, fontWeight: 700, lineHeight: 1,
                color: wxColor(p.ganWx),
                fontFamily: '"Noto Serif SC", serif',
                marginTop: 6,
              }}>{p.gan}</span>
              <span style={{
                fontSize: 34, fontWeight: 600, lineHeight: 1,
                color: wxColor(p.zhiWx),
                fontFamily: '"Noto Serif SC", serif',
                marginTop: 4,
              }}>{p.zhi}</span>
              {i === 2 && (
                <span style={{ fontSize: 9, color: '#c9a96e', marginTop: 4, letterSpacing: 1 }}>日 元</span>
              )}
            </div>
          ))}
        </div>

        {/* 总柱标签行 */}
        <div style={{
          textAlign: 'center',
          marginTop: 12,
          fontSize: 11,
          color: '#a08060',
          letterSpacing: 3,
          fontFamily: '"Noto Sans SC", sans-serif',
        }}>
          {yearGan}{yearZhi} · {monthGan}{monthZhi} · {dayGan}{dayZhi} · {hourGan}{hourZhi}
        </div>
      </div>

      {/* ┌ 命局格局分析（专业模式）── */}
      {analysis?.logic && (
        <>
          <div style={{
            padding: '18px 24px 14px',
            background: 'rgba(201,168,76,0.04)',
          }}>
            <div style={{
              fontSize: 12, color: '#8b6e4e', fontWeight: 700,
              letterSpacing: 3, marginBottom: 10,
              fontFamily: '"Noto Serif SC", serif',
              borderBottom: '1px dashed #e0cca0',
              paddingBottom: 8,
            }}>
              ── 格 局 推 断 ──
            </div>
            <div style={{
              fontSize: 13, color: '#4a3728', lineHeight: 1.85,
              fontFamily: '"Noto Sans SC", "PingFang SC", sans-serif',
            }}>
              {analysis.logic}
            </div>
          </div>
          <Divider />
        </>
      )}

      {/* ┌ 命理解读章节（完整 detail 版本）── */}
      {chapters.length > 0 && (
        <>
          <div style={{
            padding: '14px 24px 6px',
            background: '#fdf8f0',
          }}>
            <div style={{
              fontSize: 11, color: '#a08060', letterSpacing: 4,
              textAlign: 'center',
              fontFamily: '"Noto Serif SC", serif',
            }}>
              ── 命 理 解 读 ──
            </div>
          </div>

          {chapters.map((ch, i) => (
            <div key={i}>
              <ChapterBlock
                icon={chapterDefs[i]?.icon ?? '◆'}
                title={ch.title}
                content={ch.detail || ch.brief || ''}
              />
              {i < chapters.length - 1 && <Divider />}
            </div>
          ))}
        </>
      )}

      {!structured && (
        <div style={{
          padding: '32px 24px',
          textAlign: 'center',
          color: '#a08060',
          fontSize: 13,
          fontFamily: '"Noto Sans SC", sans-serif',
        }}>
          命盘尚未生成命理解读，请先生成报告后再保存图片
        </div>
      )}

      <DayunTrendShareSection series={dayunTrendSeries} label={dayunTrendLabel} period={dayunTrendPeriod} />

      <PastEventsShareSection segments={pastEventsExportSegments} />

      {/* ┌ 品牌落款 ── */}
      <div style={{
        background: 'linear-gradient(135deg, #2d1f14 0%, #4a3020 100%)',
        padding: '14px 24px',
        display: 'flex',
        justifyContent: 'space-between',
        alignItems: 'center',
        marginTop: 4,
      }}>
        <span style={{
          fontSize: 11, color: '#9a7a5a', letterSpacing: 0.5,
          fontFamily: '"Noto Sans SC", sans-serif',
        }}>
          仅供参考，不作决策依据
        </span>
        <span style={{
          fontSize: 12, color: '#e8c97c', letterSpacing: 1,
          fontFamily: '"Noto Serif SC", serif',
        }}>
          {resolvedFooter}
        </span>
      </div>
      {showDiagonalMark && brand && (
        <div style={{
          position: 'absolute', inset: 0, pointerEvents: 'none',
          overflow: 'hidden', zIndex: 1,
        }}>
          <div style={{
            position: 'absolute',
            top: '-30%', left: '-30%', right: '-30%', bottom: '-30%',
            transform: 'rotate(-30deg)',
            display: 'grid',
            gridTemplateColumns: 'repeat(auto-fill, 180px)',
            gap: '60px 40px',
            opacity: 0.06,
            color: '#000',
            fontSize: 14,
            fontFamily: '"Noto Sans SC", sans-serif',
            whiteSpace: 'nowrap',
          }}>
            {Array.from({ length: 60 }).map((_, i) => (
              <span key={i}>{brand.watermark_text}</span>
            ))}
          </div>
        </div>
      )}
    </div>
  )
})

ShareCard.displayName = 'ShareCard'
export default ShareCard
