import type { ShenshaAnnotation } from '../lib/api'

const WX_COLOR: Record<string, string> = {
  木: '#3d6b3a', 火: '#9b2c2c', 土: '#7a5c2e', 金: '#6b5a1e', 水: '#2c4a7a',
}
function wxColor(wxStr: string): string {
  for (const [k, v] of Object.entries(WX_COLOR)) {
    if (wxStr?.startsWith(k)) return v
  }
  return '#2a1a0a'
}

const SHA_COLOR = { ji: '#2d5a3d', xiong: '#8b1a1a', zhong: '#555' } as const
const SHA_BG   = { ji: '#f0f7f2', xiong: '#fdf0f0', zhong: '#f7f7f5' } as const

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

const gold = '#b8952a'
const darkBrown = '#2a1a0a'
const midBrown = '#5a3a1a'
const lightBg = '#fdf8f0'
const borderColor = '#e0cca0'

const sectionTitle = (text: string) => (
  <div style={{
    display: 'flex', alignItems: 'center', gap: 10,
    marginBottom: 14,
  }}>
    <div style={{ height: 1, flex: 1, background: `linear-gradient(to right, transparent, ${borderColor})` }} />
    <span style={{
      fontSize: 13, fontWeight: 700, letterSpacing: 4,
      color: midBrown, whiteSpace: 'nowrap',
    }}>
      {text}
    </span>
    <div style={{ height: 1, flex: 1, background: `linear-gradient(to left, transparent, ${borderColor})` }} />
  </div>
)

export default function PrintLayout({
  birthYear, birthMonth, birthDay, birthHour, gender,
  yongshen, jishen, pillars, dayun, structured, shenshaMap,
}: PrintLayoutProps) {
  const chapters = structured?.chapters ?? []
  const analysis = structured?.analysis ?? null

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

  const thS: React.CSSProperties = {
    padding: '7px 10px', textAlign: 'center',
    fontSize: 11, fontWeight: 700, color: midBrown,
    borderBottom: `1px solid ${borderColor}`,
    background: '#faf3e4',
    letterSpacing: 1,
  }
  const tdS: React.CSSProperties = {
    padding: '7px 10px', textAlign: 'center',
    fontSize: 12, color: darkBrown,
    borderBottom: `1px solid #f0e8d4`,
  }
  const tdLabelS: React.CSSProperties = {
    ...tdS, fontWeight: 700, color: midBrown,
    background: '#fdf8f0', textAlign: 'left',
    paddingLeft: 12,
  }

  return (
    <div
      className="print-only"
      style={{
        fontFamily: '"Noto Serif SC", "Source Han Serif SC", "SimSun", "STSong", serif',
        color: darkBrown,
        background: '#fff',
        padding: '20mm 18mm',
        maxWidth: 820,
        margin: '0 auto',
        lineHeight: 1.7,
      }}
    >
      {/* ── 封面头部 ── */}
      <div style={{
        textAlign: 'center',
        borderBottom: `2px solid ${gold}`,
        paddingBottom: 18,
        marginBottom: 24,
      }}>
        <div style={{ fontSize: 9, letterSpacing: 6, color: '#999', marginBottom: 6 }}>
          YUAN JU MING LI
        </div>
        <div style={{ fontSize: 28, fontWeight: 900, letterSpacing: 10, color: darkBrown, marginBottom: 8 }}>
          命　理　命　书
        </div>
        <div style={{ fontSize: 13, color: midBrown, letterSpacing: 2 }}>
          {birthYear} 年 {birthMonth} 月 {birthDay} 日 {birthHour} 时
          &nbsp;·&nbsp;{gender === 'male' ? '男　命' : '女　命'}
        </div>
        {(yongshen || jishen) && (
          <div style={{ marginTop: 10, display: 'flex', justifyContent: 'center', gap: 16 }}>
            {yongshen && (
              <span style={{ fontSize: 11, color: '#2d5a3d', background: '#f0f7f2', padding: '3px 12px', borderRadius: 2, border: '1px solid #b5d6c3' }}>
                喜用神：{yongshen}
              </span>
            )}
            {jishen && (
              <span style={{ fontSize: 11, color: '#8b1a1a', background: '#fdf0f0', padding: '3px 12px', borderRadius: 2, border: '1px solid #f5c6c6' }}>
                忌　神：{jishen}
              </span>
            )}
          </div>
        )}
      </div>

      {/* ── 四柱 ── */}
      <div style={{ marginBottom: 24, breakInside: 'avoid', pageBreakInside: 'avoid' }}>
        {sectionTitle('四　柱　排　盘')}
        <table style={{ width: '100%', borderCollapse: 'collapse', border: `1px solid ${borderColor}` }}>
          <thead>
            <tr>
              <th style={{ ...thS, width: 64 }}></th>
              {pillars.map(p => (
                <th key={p.label} style={{
                  ...thS,
                  color: p.label === '日柱' ? gold : midBrown,
                  fontSize: 12,
                }}>
                  {p.label}
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            <tr>
              <td style={tdLabelS}>主星</td>
              {pillars.map(p => <td key={p.label} style={tdS}>{p.ganShiShen}</td>)}
            </tr>
            <tr>
              <td style={tdLabelS}>天干</td>
              {pillars.map(p => (
                <td key={p.label} style={{ ...tdS, fontWeight: 900, fontSize: 20, color: wxColor(p.ganWx) }}>
                  {p.gan}
                  <span style={{ fontSize: 9, color: '#aaa', marginLeft: 2 }}>({p.ganWx})</span>
                </td>
              ))}
            </tr>
            <tr>
              <td style={tdLabelS}>地支</td>
              {pillars.map(p => (
                <td key={p.label} style={{ ...tdS, fontWeight: 900, fontSize: 20, color: wxColor(p.zhiWx) }}>
                  {p.zhi}
                  <span style={{ fontSize: 9, color: '#aaa', marginLeft: 2 }}>({p.zhiWx})</span>
                </td>
              ))}
            </tr>
            <tr>
              <td style={tdLabelS}>副星</td>
              {pillars.map(p => <td key={p.label} style={{ ...tdS, fontSize: 11 }}>{p.zhiShiShen.join(' / ')}</td>)}
            </tr>
            <tr>
              <td style={tdLabelS}>星运</td>
              {pillars.map(p => <td key={p.label} style={tdS}>{p.diShi}</td>)}
            </tr>
            <tr>
              <td style={{ ...tdLabelS, borderBottom: 'none' }}>空亡</td>
              {pillars.map(p => <td key={p.label} style={{ ...tdS, borderBottom: 'none' }}>{p.xunKong}</td>)}
            </tr>
          </tbody>
        </table>
      </div>

      {/* ── 神煞 ── */}
      {allShensha.length > 0 && (
        <div style={{ marginBottom: 24, breakInside: 'avoid', pageBreakInside: 'avoid' }}>
          {sectionTitle('神　煞')}
          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 6 }}>
            {allShensha.map((sha, i) => (
              <div
                key={i}
                style={{
                  display: 'flex', alignItems: 'flex-start', gap: 8,
                  padding: '6px 10px',
                  background: SHA_BG[sha.polarity],
                  borderLeft: `3px solid ${SHA_COLOR[sha.polarity]}`,
                  borderRadius: 2,
                  pageBreakInside: 'avoid', breakInside: 'avoid',
                }}
              >
                <span style={{ fontSize: 10, color: '#999', whiteSpace: 'nowrap', paddingTop: 2, minWidth: 26 }}>
                  {sha.pillarLabel}
                </span>
                <span style={{ fontWeight: 700, color: SHA_COLOR[sha.polarity], whiteSpace: 'nowrap', minWidth: 56, fontSize: 12 }}>
                  {sha.name}
                </span>
                {sha.annotation?.short_desc && (
                  <span style={{ fontSize: 11, color: '#444', lineHeight: 1.5 }}>
                    {sha.annotation.short_desc}
                  </span>
                )}
              </div>
            ))}
          </div>
        </div>
      )}

      {/* ── 命理解读 ── */}
      <div style={{ marginBottom: 24, pageBreakBefore: 'always', breakBefore: 'page' }}>
        {sectionTitle('命　理　解　读')}

        {/* 命局分析总览 */}
        {analysis?.logic && (
          <div style={{
            marginBottom: 18,
            padding: '12px 16px',
            background: lightBg,
            border: `1px solid ${borderColor}`,
            borderRadius: 3,
            breakInside: 'avoid',
            pageBreakInside: 'avoid',
          }}>
            <div style={{ fontSize: 11, color: midBrown, fontWeight: 700, marginBottom: 6, letterSpacing: 2 }}>
              ▍ 命局分析总览
            </div>
            <p style={{ fontSize: 12, color: darkBrown, lineHeight: 1.9, margin: 0 }}>
              {analysis.logic}
            </p>
          </div>
        )}

        {chapters.length === 0 ? (
          <div style={{ fontSize: 12, color: '#aaa', fontStyle: 'italic', padding: '20px 0', textAlign: 'center' }}>
            命理解读尚未生成
          </div>
        ) : (
          <div style={{ display: 'flex', flexDirection: 'column', gap: 18 }}>
            {chapters.map((ch, i) => (
              <div key={i} style={{ pageBreakInside: 'avoid', breakInside: 'avoid' }}>
                <div style={{
                  fontSize: 13, fontWeight: 700, color: midBrown,
                  marginBottom: 7,
                  paddingBottom: 5,
                  borderBottom: `1px dashed ${borderColor}`,
                  letterSpacing: 1,
                }}>
                  【{ch.title}】
                </div>
                <p style={{
                  fontSize: 12, color: darkBrown, lineHeight: 2,
                  margin: 0,
                  fontFamily: '"Noto Sans SC", "PingFang SC", "Microsoft YaHei", sans-serif',
                }}>
                  {ch.detail || ch.brief}
                </p>
              </div>
            ))}
          </div>
        )}
      </div>

      {/* ── 大运总览 ── */}
      <div style={{ marginBottom: 24, pageBreakBefore: 'always', breakBefore: 'page' }}>
        {sectionTitle('大　运　总　览')}
        <table style={{ width: '100%', borderCollapse: 'collapse', fontSize: 11, border: `1px solid ${borderColor}` }}>
          <thead>
            <tr>
              {['段', '起止岁', '起止年', '干支', '天干十神', '地支十神', '星运'].map(h => (
                <th key={h} style={thS}>{h}</th>
              ))}
            </tr>
          </thead>
          <tbody>
            {dayun.map((d, idx) => (
              <tr key={d.index} style={{ background: idx % 2 === 0 ? '#fff' : lightBg }}>
                <td style={{ ...tdS, color: midBrown, fontWeight: 700 }}>第{d.index + 1}运</td>
                <td style={tdS}>{d.start_age}—{d.start_age + 9}岁</td>
                <td style={{ ...tdS, fontSize: 10, color: '#777' }}>{d.start_year}—{d.end_year}</td>
                <td style={{ ...tdS, fontWeight: 900, fontSize: 17, letterSpacing: 2 }}>{d.gan}{d.zhi}</td>
                <td style={tdS}>{d.gan_shishen}</td>
                <td style={tdS}>{d.zhi_shishen}</td>
                <td style={tdS}>{d.di_shi}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {/* ── 落款 ── */}
      <div style={{
        borderTop: `1px solid ${borderColor}`,
        paddingTop: 12,
        display: 'flex',
        justifyContent: 'space-between',
        alignItems: 'center',
        fontSize: 10,
        color: '#bbb',
      }}>
        <span>本报告内容仅供参考，不构成任何决策建议。</span>
        <span style={{ color: gold, letterSpacing: 2 }}>缘 聚 命 理</span>
      </div>
    </div>
  )
}
