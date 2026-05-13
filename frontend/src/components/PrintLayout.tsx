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
  mingGe?: string
  mingGeDesc?: string
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
    marginBottom: 10,
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
  yongshen, jishen, mingGe, mingGeDesc, pillars, dayun, structured, shenshaMap,
}: PrintLayoutProps) {
  const chapters = structured?.chapters ?? []
  const analysis = structured?.analysis ?? null
  const localVerdict = (() => {
    const logic = analysis?.logic?.trim()
    if (!logic) return ''
    const firstSentence = logic.match(/^.*?[。！？]/)?.[0]?.trim() || logic
    if (!firstSentence) return ''
    if (firstSentence.length <= 72) return firstSentence
    return firstSentence.slice(0, 72).trim() + '...'
  })()

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
    padding: '5px 8px', textAlign: 'center',
    fontSize: 11, fontWeight: 700, color: midBrown,
    borderBottom: `1px solid ${borderColor}`,
    background: '#faf3e4',
    letterSpacing: 1,
  }
  const tdS: React.CSSProperties = {
    padding: '5px 8px', textAlign: 'center',
    fontSize: 12, color: darkBrown,
    borderBottom: `1px solid #f0e8d4`,
  }
  const tdLabelS: React.CSSProperties = {
    ...tdS, fontWeight: 700, color: midBrown,
    background: '#fdf8f0', textAlign: 'left',
    paddingLeft: 10,
  }

  return (
    <div className="print-only">
      {/*
        用 table/thead 实现每页重复页头：
        CSS 规范要求 thead 在分页打印时自动重复，这比 position:fixed 更可靠，
        且不会遮挡正文内容流。
      */}
      <table className="print-page-table">
        <thead>
          <tr>
            <td>
              <div className="print-page-header">
                <span className="print-page-header-center">命　理　命　书</span>
                <span className="print-page-header-info">
                  {birthYear}年{birthMonth}月{birthDay}日&nbsp;·&nbsp;{gender === 'male' ? '男命' : '女命'}
                </span>
              </div>
              {/* 页头与正文之间的间距 */}
              <div className="print-page-header-spacer" />
            </td>
          </tr>
        </thead>
        <tbody>
          <tr>
            <td style={{ verticalAlign: 'top', padding: 0 }}>

      <div
        style={{
          fontFamily: '"Noto Serif SC", "Source Han Serif SC", "SimSun", "STSong", serif',
          color: darkBrown,
          background: '#fff',
          padding: '0 16mm 14mm',
          maxWidth: 820,
          margin: '0 auto',
          lineHeight: 1.6,
        }}
      >
      {/* ── 封面头部 ── */}
      <div style={{
        textAlign: 'center',
        borderBottom: `2px solid ${gold}`,
        paddingBottom: 12,
        marginBottom: 16,
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
      <div style={{ marginBottom: 16, breakInside: 'avoid', pageBreakInside: 'avoid' }}>
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
        <div style={{ marginBottom: 16 }}>
          {sectionTitle('神　煞')}
          <table style={{ width: '100%', borderCollapse: 'collapse', border: `1px solid ${borderColor}`, fontSize: 12 }}>
            <thead>
              <tr>
                {['柱位', '神煞名称', '性质', '简述'].map(h => (
                  <th key={h} style={{
                    padding: '6px 10px', textAlign: 'center',
                    fontSize: 11, fontWeight: 700, color: midBrown,
                    background: '#faf3e4', borderBottom: `1px solid ${borderColor}`,
                    letterSpacing: 1, whiteSpace: 'nowrap',
                  }}>{h}</th>
                ))}
              </tr>
            </thead>
            <tbody>
              {allShensha.map((sha, i) => (
                <tr key={i} style={{
                  background: i % 2 === 0 ? '#fff' : lightBg,
                  breakInside: 'avoid', pageBreakInside: 'avoid',
                }}>
                  <td style={{ padding: '4px 8px', textAlign: 'center', color: midBrown, fontWeight: 600, whiteSpace: 'nowrap', borderBottom: `1px solid #f0e8d4` }}>
                    {sha.pillarLabel}
                  </td>
                  <td style={{ padding: '4px 10px', textAlign: 'center', fontWeight: 700, color: SHA_COLOR[sha.polarity], whiteSpace: 'nowrap', borderBottom: `1px solid #f0e8d4` }}>
                    {sha.name}
                  </td>
                  <td style={{ padding: '4px 6px', textAlign: 'center', whiteSpace: 'nowrap', borderBottom: `1px solid #f0e8d4` }}>
                    <span style={{
                      fontSize: 11, fontWeight: 700,
                      color: SHA_COLOR[sha.polarity],
                      background: SHA_BG[sha.polarity],
                      padding: '1px 6px', borderRadius: 2,
                    }}>
                      {sha.polarity === 'ji' ? '吉' : sha.polarity === 'xiong' ? '凶' : '中'}
                    </span>
                  </td>
                  <td style={{ padding: '4px 10px', color: '#444', lineHeight: 1.6, borderBottom: `1px solid #f0e8d4`, fontSize: 11 }}>
                    {sha.annotation?.short_desc || sha.annotation?.description || '—'}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {/* ── 命理解读 ── */}
      <div style={{ marginBottom: 16, pageBreakBefore: 'always', breakBefore: 'page' }}>
        {sectionTitle('命　理　解　读')}

        {mingGe && (
          <div style={{
            marginBottom: 14,
            padding: '10px 14px',
            background: '#fcf7ea',
            border: `1px solid ${borderColor}`,
            borderRadius: 3,
            breakInside: 'avoid',
            pageBreakInside: 'avoid',
          }}>
            <div style={{ fontSize: 11, color: midBrown, fontWeight: 700, marginBottom: 7, letterSpacing: 2 }}>
              ▍ 命格解读
            </div>
            <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
              <div style={{ fontSize: 12, color: darkBrown, lineHeight: 1.8 }}>
                <span style={{ color: midBrown, fontWeight: 700, marginRight: 6 }}>主格</span>
                <span style={{
                  display: 'inline-block',
                  padding: '1px 8px',
                  borderRadius: 2,
                  border: `1px solid ${gold}`,
                  color: darkBrown,
                  background: '#fffaf0',
                  fontWeight: 700,
                  letterSpacing: 1,
                }}>
                  {mingGe}
                </span>
              </div>
              {mingGeDesc && (
                <div style={{ fontSize: 12, color: darkBrown, lineHeight: 1.85 }}>
                  <span style={{ color: midBrown, fontWeight: 700, marginRight: 6 }}>格义</span>
                  <span>{mingGeDesc}</span>
                </div>
              )}
              {localVerdict && (
                <div style={{ fontSize: 12, color: darkBrown, lineHeight: 1.85 }}>
                  <span style={{ color: midBrown, fontWeight: 700, marginRight: 6 }}>本局落点</span>
                  <span>{localVerdict}</span>
                </div>
              )}
            </div>
          </div>
        )}

        {/* 命局分析总览 */}
        {analysis?.logic && (
          <div style={{
            marginBottom: 14,
            padding: '10px 14px',
            background: lightBg,
            border: `1px solid ${borderColor}`,
            borderRadius: 3,
            breakInside: 'avoid',
            pageBreakInside: 'avoid',
          }}>
            <div style={{ fontSize: 11, color: midBrown, fontWeight: 700, marginBottom: 5, letterSpacing: 2 }}>
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
          <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
            {chapters.map((ch, i) => (
              <div key={i} style={{ pageBreakInside: 'avoid', breakInside: 'avoid' }}>
                <div style={{
                  fontSize: 12, fontWeight: 700, color: midBrown,
                  marginBottom: 5,
                  paddingBottom: 4,
                  borderBottom: `1px dashed ${borderColor}`,
                  letterSpacing: 1,
                }}>
                  【{ch.title}】
                </div>
                <p style={{
                  fontSize: 12, color: darkBrown, lineHeight: 1.85,
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
      <div style={{ marginBottom: 16 }}>
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

            </td>
          </tr>
        </tbody>
      </table>
    </div>
  )
}
