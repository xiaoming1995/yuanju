import { useRef } from 'react'
import { WUXING_MAP, parseWuxingList } from '../lib/wuxingColorSystem'
import type { WuxingKey } from '../lib/wuxingColorSystem'
import './MingpanAvatar.css'

interface MingpanAvatarProps {
  yongshen: string
  jishen: string
  /** 日主天干，如"甲"、"乙"等 */
  dayGan: string
}

const SIZE = 320

// ======= 五行纹样渲染函数 =======

function renderWoodPattern(color: string): React.ReactNode {
  // 竹节竖纹
  const rects = []
  for (let i = 0; i < 7; i++) {
    const x = 20 + i * 42
    rects.push(
      <g key={i}>
        <rect x={x} y={40} width={18} height={240} rx={9} fill={color} opacity={0.18 - i * 0.01} />
        {/* 竹节 */}
        {[100, 160, 220].map(y => (
          <rect key={y} x={x - 2} y={y} width={22} height={5} rx={2.5} fill={color} opacity={0.35} />
        ))}
      </g>
    )
  }
  return <>{rects}</>
}

function renderFirePattern(color: string): React.ReactNode {
  // 发散三角射线
  const lines = []
  const cx = SIZE / 2, cy = SIZE / 2
  for (let i = 0; i < 12; i++) {
    const angle = (i / 12) * Math.PI * 2 - Math.PI / 2
    const innerR = 40, outerR = 140
    const x1 = cx + Math.cos(angle) * innerR
    const y1 = cy + Math.sin(angle) * innerR
    const x2 = cx + Math.cos(angle) * outerR
    const y2 = cy + Math.sin(angle) * outerR
    lines.push(
      <line key={i} x1={x1} y1={y1} x2={x2} y2={y2}
        stroke={color} strokeWidth={i % 3 === 0 ? 2.5 : 1}
        opacity={i % 3 === 0 ? 0.45 : 0.2} />
    )
  }
  return <>{lines}</>
}

function renderEarthPattern(color: string): React.ReactNode {
  // 棋盘方格纹
  const rects = []
  const step = 40
  for (let row = 0; row < 8; row++) {
    for (let col = 0; col < 8; col++) {
      if ((row + col) % 2 === 0) {
        rects.push(
          <rect key={`${row}-${col}`}
            x={col * step} y={row * step}
            width={step} height={step}
            fill={color} opacity={0.1} />
        )
      }
    }
  }
  return <>{rects}</>
}

function renderMetalPattern(color: string): React.ReactNode {
  // 同心圆弧
  const circles = []
  const cx = SIZE / 2, cy = SIZE / 2
  for (let i = 1; i <= 6; i++) {
    circles.push(
      <circle key={i} cx={cx} cy={cy} r={i * 26}
        fill="none" stroke={color}
        strokeWidth={i % 2 === 0 ? 1.5 : 0.8}
        opacity={0.25 - i * 0.02} />
    )
  }
  return <>{circles}</>
}

function renderWaterPattern(color: string): React.ReactNode {
  // 正弦波浪线
  const waves = []
  for (let w = 0; w < 6; w++) {
    const yBase = 40 + w * 50
    const points: string[] = []
    for (let x = 0; x <= SIZE; x += 4) {
      const y = yBase + Math.sin((x / SIZE) * Math.PI * 3 + w * 0.8) * 18
      points.push(`${x},${y}`)
    }
    waves.push(
      <polyline key={w}
        points={points.join(' ')}
        fill="none"
        stroke={color}
        strokeWidth={w % 2 === 0 ? 1.5 : 0.8}
        opacity={0.22} />
    )
  }
  return <>{waves}</>
}

function renderPattern(wx: WuxingKey, color: string): React.ReactNode {
  switch (wx) {
    case '木': return renderWoodPattern(color)
    case '火': return renderFirePattern(color)
    case '土': return renderEarthPattern(color)
    case '金': return renderMetalPattern(color)
    case '水': return renderWaterPattern(color)
  }
}

// ======= 下载工具函数 =======

function downloadAvatarAsSvg(svgEl: SVGSVGElement, filename: string) {
  const serializer = new XMLSerializer()
  const svgStr = serializer.serializeToString(svgEl)
  const blob = new Blob([svgStr], { type: 'image/svg+xml' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = filename
  a.click()
  URL.revokeObjectURL(url)
}

function downloadAvatarAsPng(svgEl: SVGSVGElement, filename: string) {
  const canvas = document.createElement('canvas')
  canvas.width = SIZE
  canvas.height = SIZE
  const ctx = canvas.getContext('2d')
  if (!ctx) {
    downloadAvatarAsSvg(svgEl, filename.replace('.png', '.svg'))
    return
  }
  const serializer = new XMLSerializer()
  const svgStr = serializer.serializeToString(svgEl)
  const img = new Image()
  const blob = new Blob([svgStr], { type: 'image/svg+xml;charset=utf-8' })
  const url = URL.createObjectURL(blob)
  img.onload = () => {
    ctx.drawImage(img, 0, 0)
    URL.revokeObjectURL(url)
    if (canvas.toBlob) {
      canvas.toBlob(blob2 => {
        if (!blob2) return
        const url2 = URL.createObjectURL(blob2)
        const a = document.createElement('a')
        a.href = url2
        a.download = filename
        a.click()
        URL.revokeObjectURL(url2)
      }, 'image/png')
    } else {
      // Safari fallback
      const dataUrl = canvas.toDataURL('image/png')
      const a = document.createElement('a')
      a.href = dataUrl
      a.download = filename
      a.click()
    }
  }
  img.onerror = () => {
    URL.revokeObjectURL(url)
    downloadAvatarAsSvg(svgEl, filename.replace('.png', '.svg'))
  }
  img.src = url
}

// ======= 主组件 =======

export default function MingpanAvatar({ yongshen, jishen, dayGan }: MingpanAvatarProps) {
  const svgRef = useRef<SVGSVGElement>(null)
  const yongList = parseWuxingList(yongshen)
  const jiList = parseWuxingList(jishen)

  // 锁定状态
  if (yongList.length === 0) {
    return (
      <div className="mingpan-avatar-locked card">
        <div className="avatar-lock-icon">🔒</div>
        <p className="avatar-lock-text">生成 AI 报告后解锁命理头像</p>
      </div>
    )
  }

  const primary = WUXING_MAP[yongList[0]]
  const secondary = yongList[1] ? WUXING_MAP[yongList[1]] : null
  const border = jiList[0] ? WUXING_MAP[jiList[0]] : null
  const mainPattern = yongList[0]

  // 渐变 ID（防止多实例冲突）
  const gradId = `avatar-grad-${yongList[0]}-${yongList[1] || 'x'}`
  const borderGradId = `avatar-border-${jiList[0] || 'x'}`

  const handleDownload = () => {
    if (!svgRef.current) return
    downloadAvatarAsPng(svgRef.current, `mingpan-avatar-${dayGan || 'yuan'}.png`)
  }

  return (
    <div className="mingpan-avatar-wrap">
      <div className="mingpan-avatar-preview">
        <svg
          ref={svgRef}
          width={SIZE}
          height={SIZE}
          viewBox={`0 0 ${SIZE} ${SIZE}`}
          xmlns="http://www.w3.org/2000/svg"
          className="mingpan-svg"
        >
          <defs>
            {/* 背景渐变 */}
            <linearGradient id={gradId} x1="0%" y1="0%" x2="100%" y2="100%">
              <stop offset="0%" stopColor={primary.darkColor} />
              <stop offset="100%" stopColor={secondary ? secondary.darkColor : primary.color} />
            </linearGradient>
            {/* 边框渐变 */}
            {border && (
              <linearGradient id={borderGradId} x1="0%" y1="0%" x2="100%" y2="100%">
                <stop offset="0%" stopColor={border.color} />
                <stop offset="100%" stopColor={border.darkColor} />
              </linearGradient>
            )}
          </defs>

          {/* Layer 1: 背景 */}
          <rect width={SIZE} height={SIZE} fill={`url(#${gradId})`} />

          {/* Layer 2: 纹样 */}
          <g style={{ mixBlendMode: 'overlay' }}>
            {renderPattern(mainPattern, primary.color)}
          </g>

          {/* 副纹样（如有第二五行） */}
          {secondary && (
            <g style={{ mixBlendMode: 'overlay', opacity: 0.5 }}>
              {renderPattern(yongList[1], secondary.color)}
            </g>
          )}

          {/* Layer 3: 圆形遮罩（聚焦中心） */}
          <radialGradient id={`vignette-${yongList[0]}`} cx="50%" cy="50%" r="50%">
            <stop offset="40%" stopColor="transparent" />
            <stop offset="100%" stopColor="rgba(0,0,0,0.45)" />
          </radialGradient>
          <rect width={SIZE} height={SIZE} fill={`url(#vignette-${yongList[0]})`} />

          {/* Layer 4: 边框 */}
          {border ? (
            <rect x={6} y={6} width={SIZE - 12} height={SIZE - 12}
              rx={16} fill="none"
              stroke={`url(#${borderGradId})`} strokeWidth={3}
              opacity={0.7} />
          ) : (
            <rect x={6} y={6} width={SIZE - 12} height={SIZE - 12}
              rx={16} fill="none"
              stroke="rgba(255,255,255,0.2)" strokeWidth={2} />
          )}

          {/* 装饰性小圆角内框 */}
          <rect x={14} y={14} width={SIZE - 28} height={SIZE - 28}
            rx={10} fill="none"
            stroke="rgba(255,255,255,0.08)" strokeWidth={1} />

          {/* Layer 5: 中央日主天干大字 */}
          {dayGan && (
            <text
              x={SIZE / 2}
              y={SIZE / 2 + 28}
              textAnchor="middle"
              fontSize={100}
              fontWeight="700"
              fontFamily="'Noto Serif SC', serif"
              fill="rgba(255,255,255,0.92)"
              style={{ filter: 'drop-shadow(0 4px 16px rgba(0,0,0,0.6))' }}
            >
              {dayGan}
            </text>
          )}

          {/* 底部五行标注 */}
          <text
            x={SIZE / 2}
            y={SIZE - 22}
            textAnchor="middle"
            fontSize={13}
            fontFamily="'Noto Serif SC', serif"
            fill="rgba(255,255,255,0.55)"
            letterSpacing="4"
          >
            {yongList.join(' ')}
          </text>
        </svg>
      </div>

      <div className="mingpan-avatar-actions">
        <button
          className="btn btn-secondary btn-sm"
          id="download-avatar-btn"
          onClick={handleDownload}
        >
          ⬇ 下载头像
        </button>
        <div className="mingpan-avatar-hint">
          基于喜用神 <strong style={{ color: primary.color }}>{yongList.join('·')}</strong> 生成
        </div>
      </div>
    </div>
  )
}
