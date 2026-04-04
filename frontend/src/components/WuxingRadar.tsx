interface WuxingProps {
  wuxing: { mu: number; huo: number; tu: number; jin: number; shui: number }
}

const WUXING_ITEMS = [
  { key: 'mu',   label: '木', color: '#4caf7d', angle: -90 },
  { key: 'huo',  label: '火', color: '#e05c4b', angle: -18 },
  { key: 'tu',   label: '土', color: '#c17f3e', angle: 54 },
  { key: 'jin',  label: '金', color: '#c9a84c', angle: 126 },
  { key: 'shui', label: '水', color: '#5b9bd5', angle: 198 },
] as const

export default function WuxingRadar({ wuxing }: WuxingProps) {
  const size = 220
  const cx = size / 2
  const cy = size / 2
  const maxR = 80
  const total = Math.max(wuxing.mu + wuxing.huo + wuxing.tu + wuxing.jin + wuxing.shui, 1)

  const toRad = (deg: number) => (deg * Math.PI) / 180
  const getPoint = (angle: number, r: number) => ({
    x: cx + r * Math.cos(toRad(angle)),
    y: cy + r * Math.sin(toRad(angle)),
  })

  // 背景网格（5级）
  const gridLevels = [0.2, 0.4, 0.6, 0.8, 1.0]

  // 数据多边形顶点
  const dataPoints = WUXING_ITEMS.map(item => {
    const ratio = (wuxing[item.key] / total) * 1.2
    const r = Math.min(ratio, 1) * maxR
    return getPoint(item.angle, r)
  })
  const dataPath = dataPoints.map((p, i) => `${i === 0 ? 'M' : 'L'} ${p.x.toFixed(1)} ${p.y.toFixed(1)}`).join(' ') + ' Z'

  return (
    <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', gap: 20 }}>
      <svg width={size} height={size} viewBox={`0 0 ${size} ${size}`}>
        {/* 背景网格 */}
        {gridLevels.map((level, li) => {
          const pts = WUXING_ITEMS.map(item => getPoint(item.angle, maxR * level))
          const path = pts.map((p, i) => `${i === 0 ? 'M' : 'L'} ${p.x.toFixed(1)} ${p.y.toFixed(1)}`).join(' ') + ' Z'
          return <path key={li} d={path} fill="none" stroke="rgba(255,255,255,0.06)" strokeWidth="1" />
        })}

        {/* 轴线 */}
        {WUXING_ITEMS.map(item => {
          const end = getPoint(item.angle, maxR)
          return <line key={item.key} x1={cx} y1={cy} x2={end.x} y2={end.y} stroke="rgba(255,255,255,0.08)" strokeWidth="1" />
        })}

        {/* 数据面 */}
        <path d={dataPath} fill="rgba(201,168,76,0.15)" stroke="#c9a84c" strokeWidth="1.5" />

        {/* 数据点 */}
        {dataPoints.map((p, i) => (
          <circle key={i} cx={p.x} cy={p.y} r={4} fill={WUXING_ITEMS[i].color} opacity={0.9} />
        ))}

        {/* 标签 */}
        {WUXING_ITEMS.map(item => {
          const pt = getPoint(item.angle, maxR + 22)
          return (
            <text key={item.key} x={pt.x} y={pt.y}
              textAnchor="middle" dominantBaseline="middle"
              fontSize="13" fontFamily="'Noto Serif SC', serif"
              fill={item.color} fontWeight="600"
            >
              {item.label}
            </text>
          )
        })}
      </svg>

      {/* 图例 */}
      <div style={{ display: 'flex', gap: 16, flexWrap: 'wrap', justifyContent: 'center' }}>
        {WUXING_ITEMS.map(item => (
          <div key={item.key} style={{ display: 'flex', alignItems: 'center', gap: 6 }}>
            <div style={{ width: 10, height: 10, borderRadius: '50%', backgroundColor: item.color }} />
            <span style={{ fontSize: 13, color: 'var(--text-secondary)' }}>
              {item.label} × {wuxing[item.key]}
            </span>
          </div>
        ))}
      </div>
    </div>
  )
}
