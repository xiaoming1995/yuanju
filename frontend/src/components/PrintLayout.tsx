import type { ShenshaAnnotation } from '../lib/api'

const WX_COLOR: Record<string, string> = {
  木: '#3d7a55', 火: '#c0392b', 土: '#8b6940', 金: '#7a6830', 水: '#2c5282',
}
function wxColor(wxStr: string): string {
  for (const [k, v] of Object.entries(WX_COLOR)) {
    if (wxStr?.startsWith(k)) return v
  }
  return '#333'
}

const SHA_COLOR = { ji: '#3d6b4f', xiong: '#b52525', zhong: '#666' } as const
const SHA_BG   = { ji: '#f0f7f2', xiong: '#fff0f0', zhong: '#f5f5f5' } as const

interface Pillar {
  label: string
  gan: string; zhi: string
  ganWx: string; zhiWx: string
  ganShiShen: string
  zhiShiShen: string[]
  diShi: string
  xunKong: string
  shenSha: string[]
}

interface DayunItem {
  index: number
  gan: string; zhi: string
  start_age: number; start_year: number; end_year: number
  gan_shishen: string; zhi_shishen: string; di_shi: string
}

interface ReportChapter {
  title: string; detail: string; brief: string
}

interface PrintLayoutProps {
  birthYear: number; birthMonth: number; birthDay: number; birthHour: number; gender: string
  yongshen: string; jishen: string
  pillars: Pillar[]
  dayun: DayunItem[]
  structured: { chapters: ReportChapter[]; analysis?: { logic: string; summary: string } | null } | null
  shenshaMap: Map<string, ShenshaAnnotation>
}

const thStyle: React.CSSProperties = {
  padding: '8px 10px', textAlign: 'center',
  fontSize: 12, fontWeight: 700, color: '#5a3a1a',
  border: '1px solid #e8dcc8', background: '#faf5eb',
}
const tdStyle: React.CSSProperties = {
  padding: '8px 10px', textAlign: 'center',
  fontSize: 13, color: '#333', border: '1px solid #eee',
}
const tdLabelStyle: React.CSSProperties = {
  padding: '8px 10px', textAlign: 'center',
  fontSize: 13, border: '1px solid #eee',
  fontWeight: 600, color: '#5a3a1a',
  background: '#faf5eb', whiteSpace: 'nowrap',
}
const sectionTitleStyle: React.CSSProperties = {
  fontSize: 14, fontWeight: 700, color: '#7a5c3a',
  letterSpacing: 2, marginBottom: 12,
  borderLeft: '3px solid #c9a84c', paddingLeft: 10,
}

export default function PrintLayout({
  birthYear, birthMonth, birthDay, birthHour, gender,
  yongshen, jishen, pillars, dayun, structured, shenshaMap,
}: PrintLayoutProps) {
  const chapters = structured?.chapters ?? []

  const allShensha: Array<{
    pillarLabel: string
    name: string
    annotation: ShenshaAnnotation | undefined
    polarity: 'ji' | 'xiong' | 'zhong'
  }> = []
  pillars.forEach(p => {
    p.shenSha.forEach(name => {
      const annotation = shenshaMap.get(name)
      allShensha.push({
        pillarLabel: p.label,
        name,
        annotation,
        polarity: annotation?.polarity ?? 'zhong',
      })
    })
  })

  return (
    <div
      className="print-only"
      style={{
        fontFamily: '"Noto Serif SC", "SimSun", serif',
        color: '#1a1a1a', background: '#fff',
        padding: '24px 32px', maxWidth: 800, margin: '0 auto',
      }}
    >
      {/* 品牌头部 */}
      <div style={{ textAlign: 'center', borderBottom: '2px solid #c9a84c', paddingBottom: 16, marginBottom: 24 }}>
        <div style={{ fontSize: 26, fontWeight: 700, letterSpacing: 8, color: '#3a2416', marginBottom: 6 }}>
          缘 聚 命 理
        </div>
        <div style={{ fontSize: 13, color: '#666', letterSpacing: 1 }}>
          {birthYear}年{birthMonth}月{birthDay}日 {birthHour}时
          &nbsp;·&nbsp;{gender === 'male' ? '男命' : '女命'}
        </div>
      </div>

      {/* 四柱 */}
      <div style={{ marginBottom: 28 }}>
        <div style={sectionTitleStyle}>四 柱</div>
        <table style={{ width: '100%', borderCollapse: 'collapse' }}>
          <thead>
            <tr>
              <th style={thStyle}>柱</th>
              {pillars.map(p => (
                <th key={p.label} style={{ ...thStyle, color: p.label === '日柱' ? '#c9a84c' : '#3a2416' }}>
                  {p.label}
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            <tr>
              <td style={tdLabelStyle}>主星</td>
              {pillars.map(p => <td key={p.label} style={tdStyle}>{p.ganShiShen}</td>)}
            </tr>
            <tr>
              <td style={tdLabelStyle}>天干</td>
              {pillars.map(p => (
                <td key={p.label} style={{ ...tdStyle, fontWeight: 700, fontSize: 22, color: wxColor(p.ganWx) }}>
                  {p.gan}
                  <span style={{ fontSize: 10, color: '#999', marginLeft: 2 }}>({p.ganWx})</span>
                </td>
              ))}
            </tr>
            <tr>
              <td style={tdLabelStyle}>地支</td>
              {pillars.map(p => (
                <td key={p.label} style={{ ...tdStyle, fontWeight: 700, fontSize: 22, color: wxColor(p.zhiWx) }}>
                  {p.zhi}
                  <span style={{ fontSize: 10, color: '#999', marginLeft: 2 }}>({p.zhiWx})</span>
                </td>
              ))}
            </tr>
            <tr>
              <td style={tdLabelStyle}>副星</td>
              {pillars.map(p => <td key={p.label} style={tdStyle}>{p.zhiShiShen.join(' / ')}</td>)}
            </tr>
            <tr>
              <td style={tdLabelStyle}>星运</td>
              {pillars.map(p => <td key={p.label} style={tdStyle}>{p.diShi}</td>)}
            </tr>
            <tr>
              <td style={tdLabelStyle}>空亡</td>
              {pillars.map(p => <td key={p.label} style={tdStyle}>{p.xunKong}</td>)}
            </tr>
          </tbody>
        </table>
      </div>

      {/* 喜用神 / 忌神 */}
      {(yongshen || jishen) && (
        <div style={{ marginBottom: 28, display: 'flex', gap: 20, flexWrap: 'wrap' }}>
          {yongshen && (
            <div style={{ padding: '8px 16px', border: '1px solid #b5d6c3', borderRadius: 6, background: '#f4fbf7', fontSize: 13 }}>
              <span style={{ color: '#666', marginRight: 6 }}>喜用神</span>
              <span style={{ fontWeight: 700, color: '#3d6b4f' }}>{yongshen}</span>
            </div>
          )}
          {jishen && (
            <div style={{ padding: '8px 16px', border: '1px solid #f5c6c6', borderRadius: 6, background: '#fff5f5', fontSize: 13 }}>
              <span style={{ color: '#666', marginRight: 6 }}>忌神</span>
              <span style={{ fontWeight: 700, color: '#b52525' }}>{jishen}</span>
            </div>
          )}
        </div>
      )}

      {/* 神煞（内联注解） */}
      {allShensha.length > 0 && (
        <div style={{ marginBottom: 28 }}>
          <div style={sectionTitleStyle}>神 煞</div>
          <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
            {allShensha.map((sha, i) => (
              <div
                key={i}
                style={{
                  display: 'flex', alignItems: 'flex-start', gap: 10,
                  padding: '8px 12px',
                  background: SHA_BG[sha.polarity],
                  borderRadius: 6,
                  borderLeft: `3px solid ${SHA_COLOR[sha.polarity]}`,
                  pageBreakInside: 'avoid', breakInside: 'avoid',
                }}
              >
                <span style={{ fontSize: 11, color: '#888', whiteSpace: 'nowrap', paddingTop: 2, minWidth: 28 }}>
                  {sha.pillarLabel}
                </span>
                <span style={{ fontWeight: 700, color: SHA_COLOR[sha.polarity], whiteSpace: 'nowrap', minWidth: 64 }}>
                  {sha.name}
                </span>
                {sha.annotation?.short_desc && (
                  <span style={{ fontSize: 12, color: '#444', lineHeight: 1.6 }}>
                    {sha.annotation.short_desc}
                  </span>
                )}
              </div>
            ))}
          </div>
        </div>
      )}

      {/* 命理解读 */}
      <div style={{ marginBottom: 28, pageBreakBefore: 'always', breakBefore: 'page' }}>
        <div style={sectionTitleStyle}>命 理 解 读</div>
        {chapters.length === 0 ? (
          <div style={{ fontSize: 13, color: '#999', fontStyle: 'italic', padding: '16px 0' }}>
            命理解读尚未生成
          </div>
        ) : (
          <div style={{ display: 'flex', flexDirection: 'column', gap: 24 }}>
            {chapters.map((ch, i) => (
              <div key={i} style={{ pageBreakInside: 'avoid', breakInside: 'avoid' }}>
                <div style={{
                  fontSize: 14, fontWeight: 700, color: '#3a2416',
                  marginBottom: 8, paddingBottom: 6,
                  borderBottom: '1px dashed #e0cca0',
                }}>
                  {ch.title}
                </div>
                <div style={{
                  fontSize: 13, color: '#333', lineHeight: 1.9,
                  fontFamily: '"Noto Sans SC", "PingFang SC", sans-serif',
                }}>
                  {ch.detail || ch.brief}
                </div>
              </div>
            ))}
          </div>
        )}
      </div>

      {/* 大运总览（竖向列表） */}
      <div style={{ marginBottom: 28, pageBreakBefore: 'always', breakBefore: 'page' }}>
        <div style={sectionTitleStyle}>大 运 总 览</div>
        <table style={{ width: '100%', borderCollapse: 'collapse', fontSize: 12 }}>
          <thead>
            <tr>
              <th style={thStyle}>段</th>
              <th style={thStyle}>起止岁</th>
              <th style={thStyle}>起止年</th>
              <th style={thStyle}>干支</th>
              <th style={thStyle}>天干十神</th>
              <th style={thStyle}>地支十神</th>
              <th style={thStyle}>星运</th>
            </tr>
          </thead>
          <tbody>
            {dayun.map(d => (
              <tr key={d.index} style={{ borderBottom: '1px solid #eee' }}>
                <td style={tdStyle}>第{d.index + 1}运</td>
                <td style={tdStyle}>{d.start_age}—{d.start_age + 9}岁</td>
                <td style={tdStyle}>{d.start_year}—{d.end_year}</td>
                <td style={{ ...tdStyle, fontWeight: 700, fontSize: 18 }}>{d.gan}{d.zhi}</td>
                <td style={tdStyle}>{d.gan_shishen}</td>
                <td style={tdStyle}>{d.zhi_shishen}</td>
                <td style={tdStyle}>{d.di_shi}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {/* 品牌落款 */}
      <div style={{
        borderTop: '1px solid #e0cca0', paddingTop: 12,
        display: 'flex', justifyContent: 'space-between',
        fontSize: 11, color: '#999',
      }}>
        <span>本报告内容仅供参考，不构成任何决策建议。</span>
        <span style={{ color: '#c9a84c' }}>yuanju.com</span>
      </div>
    </div>
  )
}
