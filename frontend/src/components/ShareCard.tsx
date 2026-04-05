import { forwardRef } from 'react'
import type { StructuredReport } from '../lib/api'

// 五行对应色彩（明色国风版本）
const WX_COLOR: Record<string, string> = {
  '木': '#4a7c59', '火': '#c0392b', '土': '#a0784a',
  '金': '#8b7536', '水': '#2c5282',
}

function wxColor(char: string) {
  for (const [k, v] of Object.entries(WX_COLOR)) {
    if (char?.startsWith(k)) return v
  }
  return '#5c4a3a'
}

interface ShareCardProps {
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
    yearGanWx, yearZhiWx, monthGanWx, monthZhiWx, dayGanWx, dayZhiWx, hourGanWx, hourZhiWx,
    yongshen, jishen, structured,
  } = props

  const pillars = [
    { label: '年', gan: yearGan, zhi: yearZhi, ganWx: yearGanWx, zhiWx: yearZhiWx },
    { label: '月', gan: monthGan, zhi: monthZhi, ganWx: monthGanWx, zhiWx: monthZhiWx },
    { label: '日', gan: dayGan, zhi: dayZhi, ganWx: dayGanWx, zhiWx: dayZhiWx },
    { label: '时', gan: hourGan, zhi: hourZhi, ganWx: hourGanWx, zhiWx: hourZhiWx },
  ]

  // AI 章节摘要（取前三章 brief）
  const chapters = structured?.chapters?.slice(0, 3) ?? []
  const chapterIcons = ['✦ 事业', '✦ 感情', '✦ 健康']

  return (
    <div ref={ref} style={{
      width: 360,
      background: '#fdf8f0',
      border: '1.5px solid #c9a96e',
      borderRadius: 16,
      fontFamily: '"Noto Serif SC", "Source Han Serif CN", serif',
      overflow: 'hidden',
      boxShadow: '0 4px 24px rgba(139,101,53,0.15)',
    }}>

      {/* 顶部品牌栏 */}
      <div style={{
        background: 'linear-gradient(135deg, #3d2b1f 0%, #5c3d2e 100%)',
        padding: '16px 20px',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'space-between',
      }}>
        <div style={{ color: '#e8c97c', fontSize: 18, letterSpacing: 4, fontWeight: 700 }}>
          ✦ 缘 聚 命 理
        </div>
        <div style={{ color: '#b89a6a', fontSize: 11, letterSpacing: 1 }}>
          {birthYear}年{birthMonth}月{birthDay}日{birthHour}时 · {gender === 'male' ? '男' : '女'}
        </div>
      </div>

      {/* 四柱区 */}
      <div style={{
        padding: '20px 20px 16px',
        borderBottom: '1px solid #e8dcc8',
        background: '#fdfaf4',
      }}>
        <div style={{
          display: 'grid',
          gridTemplateColumns: 'repeat(4, 1fr)',
          gap: 8,
        }}>
          {pillars.map((p, i) => (
            <div key={i} style={{
              display: 'flex',
              flexDirection: 'column',
              alignItems: 'center',
              gap: 4,
              padding: '12px 4px',
              background: i === 2 ? 'rgba(201,168,76,0.08)' : 'transparent',
              borderRadius: 8,
              border: i === 2 ? '1px solid rgba(201,168,76,0.3)' : '1px solid transparent',
            }}>
              <span style={{ fontSize: 9, color: '#a08060', letterSpacing: 2 }}>{p.label}柱</span>
              <span style={{
                fontSize: 32, fontWeight: 700, lineHeight: 1.1,
                color: wxColor(p.ganWx), letterSpacing: 0,
              }}>{p.gan}</span>
              <span style={{
                fontSize: 28, fontWeight: 600, lineHeight: 1.1,
                color: wxColor(p.zhiWx),
              }}>{p.zhi}</span>
              <span style={{ fontSize: 9, color: '#c9a96e', marginTop: 2 }}>
                {i === 2 ? '(日元)' : ''}
              </span>
            </div>
          ))}
        </div>
      </div>

      {/* 用神忌神 */}
      <div style={{
        padding: '12px 20px',
        display: 'flex',
        gap: 12,
        borderBottom: '1px solid #e8dcc8',
        background: '#faf5eb',
      }}>
        {yongshen && (
          <span style={{
            fontSize: 12, padding: '4px 12px', borderRadius: 20,
            background: 'rgba(74,124,89,0.12)',
            color: '#3d6b4f', border: '1px solid rgba(74,124,89,0.3)',
          }}>
            喜用神：{yongshen}
          </span>
        )}
        {jishen && (
          <span style={{
            fontSize: 12, padding: '4px 12px', borderRadius: 20,
            background: 'rgba(192,57,43,0.08)',
            color: '#9b2c1e', border: '1px solid rgba(192,57,43,0.25)',
          }}>
            忌神：{jishen}
          </span>
        )}
      </div>

      {/* AI 分析摘要 */}
      {chapters.length > 0 ? (
        <div style={{ padding: '16px 20px', background: '#fdfbf6' }}>
          <div style={{
            fontSize: 10, color: '#a08060', letterSpacing: 3,
            marginBottom: 12, textAlign: 'center',
          }}>
            ── AI 命 理 解 读 ──
          </div>
          {chapters.map((ch, i) => (
            <div key={i} style={{
              marginBottom: i < chapters.length - 1 ? 12 : 0,
              paddingBottom: i < chapters.length - 1 ? 12 : 0,
              borderBottom: i < chapters.length - 1 ? '1px dashed #e8dcc8' : 'none',
            }}>
              <div style={{
                fontSize: 11, color: '#8b6e4e', fontWeight: 700,
                marginBottom: 4, letterSpacing: 1,
              }}>
                {chapterIcons[i] || `✦ ${ch.title}`}
              </div>
              <div style={{
                fontSize: 12, color: '#5c4a3a', lineHeight: 1.7,
                fontFamily: '"Noto Sans SC", sans-serif',
              }}>
                {ch.brief?.length > 60 ? ch.brief.slice(0, 60) + '……' : ch.brief}
              </div>
            </div>
          ))}
        </div>
      ) : (
        <div style={{
          padding: '20px', textAlign: 'center',
          color: '#a08060', fontSize: 12,
        }}>
          命盘尚未生成 AI 解读
        </div>
      )}

      {/* 品牌落款 */}
      <div style={{
        padding: '10px 20px',
        background: 'linear-gradient(135deg, #3d2b1f 0%, #5c3d2e 100%)',
        display: 'flex',
        justifyContent: 'space-between',
        alignItems: 'center',
      }}>
        <span style={{ fontSize: 10, color: '#b89a6a', letterSpacing: 1 }}>
          仅供参考，不作决策依据
        </span>
        <span style={{ fontSize: 11, color: '#e8c97c', letterSpacing: 1 }}>
          yuanju.com
        </span>
      </div>
    </div>
  )
})

ShareCard.displayName = 'ShareCard'
export default ShareCard
