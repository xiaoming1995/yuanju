import { forwardRef } from 'react'
import type { StructuredReport } from '../lib/api'

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
  yongshen: string
  jishen: string
  structured: StructuredReport | null
}

const ShareCard = forwardRef<HTMLDivElement, ShareCardProps>((props, ref) => {
  const {
    birthYear, birthMonth, birthDay, birthHour, gender,
    yearGan, yearZhi, monthGan, monthZhi, dayGan, dayZhi, hourGan, hourZhi,
    yearGanWx, yearZhiWx, monthGanWx, monthZhiWx,
    dayGanWx, dayZhiWx, hourGanWx, hourZhiWx,
    yongshen, jishen, structured,
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
    { icon: '◈', key: 'personality' },
    { icon: '❤', key: 'romance' },
    { icon: '✦', key: 'career' },
    { icon: '☽', key: 'health' },
  ]

  return (
    <div ref={ref} style={{
      width: 400,
      background: '#fdf9f2',
      fontFamily: '"Noto Serif SC", serif',
      overflow: 'hidden',
      boxSizing: 'border-box',
    }}>
      {/* 天干地支专属字体 */}
      <style>{`@import url('${PILLAR_FONT_URL}');`}</style>

      {/* ┌ 顶部品牌栏 ── */}
      <div style={{
        background: 'linear-gradient(135deg, #2d1f14 0%, #4a3020 50%, #3a2416 100%)',
        padding: '20px 24px 18px',
      }}>
        <div style={{
          color: '#e8c97c',
          fontSize: 20,
          letterSpacing: 6,
          fontWeight: 700,
          textAlign: 'center',
          fontFamily: '"Noto Serif SC", serif',
          marginBottom: 6,
        }}>
          缘 聚 命 理
        </div>
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

      {/* ┌ 喜用神 / 忌神 ── */}
      {(yongshen || jishen) && (
        <>
          <div style={{
            padding: '14px 24px',
            display: 'flex',
            gap: 12,
            justifyContent: 'center',
            flexWrap: 'wrap',
            background: '#fdf8f0',
            borderBottom: '1px solid #e8dcc8',
          }}>
            {yongshen && (
              <span style={{
                fontSize: 13, padding: '5px 16px', borderRadius: 20,
                background: 'rgba(74,124,89,0.1)',
                color: '#3d6b4f', border: '1px solid rgba(74,124,89,0.35)',
                fontFamily: '"Noto Sans SC", sans-serif',
              }}>
                喜用神：{yongshen}
              </span>
            )}
            {jishen && (
              <span style={{
                fontSize: 13, padding: '5px 16px', borderRadius: 20,
                background: 'rgba(192,57,43,0.07)',
                color: '#8b2c1e', border: '1px solid rgba(192,57,43,0.25)',
                fontFamily: '"Noto Sans SC", sans-serif',
              }}>
                忌神：{jishen}
              </span>
            )}
          </div>
        </>
      )}

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

      {/* ┌ AI 解读章节（完整 detail 版本）── */}
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
              ── AI 命 理 解 读 ──
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
          命盘尚未生成 AI 解读，请先生成报告后再保存图片
        </div>
      )}

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
          yuanju.com
        </span>
      </div>
    </div>
  )
})

ShareCard.displayName = 'ShareCard'
export default ShareCard
