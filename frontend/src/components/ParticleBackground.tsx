import { useEffect, useRef } from 'react'

// ============================================================
// 宇宙星辰背景 —— 三层深度 + 闪烁 + 偶发流星
// ============================================================

interface Star {
  x: number
  y: number
  vx: number
  vy: number
  size: number
  alpha: number
  alphaBase: number
  alphaPhase: number    // 闪烁相位（0~2π）
  alphaSpeed: number    // 闪烁速度
  color: string         // rgba前缀
  layer: 0 | 1 | 2     // 远/中/近三层景深
}

interface Meteor {
  x: number
  y: number
  vx: number
  vy: number
  life: number    // 0~1 剩余生命
  length: number
}

// 真实星色：白色系为主，少量蓝白、暖白、淡金
const STAR_COLORS = [
  'rgba(255,255,255,',      // 纯白      45%
  'rgba(220,235,255,',      // 蓝白      25%
  'rgba(255,248,220,',      // 暖白      18%
  'rgba(180,210,255,',      // 淡蓝      8%
  'rgba(255,220,130,',      // 淡金（命理主题点缀）4%
]
const COLOR_WEIGHTS = [0.45, 0.25, 0.18, 0.08, 0.04]

function pickColor(): string {
  const r = Math.random()
  let acc = 0
  for (let i = 0; i < COLOR_WEIGHTS.length; i++) {
    acc += COLOR_WEIGHTS[i]
    if (r < acc) return STAR_COLORS[i]
  }
  return STAR_COLORS[0]
}

function createStar(W: number, H: number): Star {
  // 三层景深：远景（0）密、小、暗、慢；近景（2）稀、大、亮、快
  const layerRand = Math.random()
  const layer: 0 | 1 | 2 = layerRand < 0.65 ? 0 : layerRand < 0.90 ? 1 : 2

  const layerConfig = [
    { sizeMin: 0.2, sizeMax: 0.8,  alphaMin: 0.10, alphaMax: 0.45, speedMul: 0.04 },
    { sizeMin: 0.6, sizeMax: 1.4,  alphaMin: 0.30, alphaMax: 0.75, speedMul: 0.10 },
    { sizeMin: 1.2, sizeMax: 2.4,  alphaMin: 0.55, alphaMax: 1.00, speedMul: 0.20 },
  ][layer]

  const size = Math.random() * (layerConfig.sizeMax - layerConfig.sizeMin) + layerConfig.sizeMin
  const alphaBase = Math.random() * (layerConfig.alphaMax - layerConfig.alphaMin) + layerConfig.alphaMin

  const angle = Math.random() * Math.PI * 2
  const speed = (Math.random() * 0.06 + 0.01) * layerConfig.speedMul

  return {
    x: Math.random() * W,
    y: Math.random() * H,
    vx: Math.cos(angle) * speed,
    vy: Math.sin(angle) * speed,
    size,
    alpha: alphaBase,
    alphaBase,
    alphaPhase: Math.random() * Math.PI * 2,
    alphaSpeed: Math.random() * 0.008 + 0.002,   // 闪烁速度
    color: pickColor(),
    layer,
  }
}

function createMeteor(W: number, H: number): Meteor {
  // 从屏幕上方随机位置向右下方飞
  return {
    x: Math.random() * W * 0.7,
    y: Math.random() * H * 0.3,
    vx: Math.random() * 6 + 4,
    vy: Math.random() * 3 + 2,
    life: 1,
    length: Math.random() * 100 + 60,
  }
}

export default function ParticleBackground() {
  const canvasRef = useRef<HTMLCanvasElement>(null)

  useEffect(() => {
    if (window.matchMedia('(prefers-reduced-motion: reduce)').matches) return

    const canvas = canvasRef.current
    if (!canvas) return
    const ctx = canvas.getContext('2d')
    if (!ctx) return

    let W = window.innerWidth
    let H = window.innerHeight

    const starCount = () => Math.min(Math.floor((W * H) / 5000), 320)

    canvas.width = W
    canvas.height = H

    let stars: Star[] = Array.from({ length: starCount() }, () => createStar(W, H))
    let meteors: Meteor[] = []
    let frame = 0
    let raf: number

    // 每隔 8~18 秒触发一颗流星
    let nextMeteorAt = frame + Math.floor(Math.random() * 600 + 480)

    const resize = () => {
      W = window.innerWidth
      H = window.innerHeight
      canvas.width = W
      canvas.height = H
      stars = Array.from({ length: starCount() }, () => createStar(W, H))
    }

    const draw = () => {
      ctx.clearRect(0, 0, W, H)
      frame++

      // ── 星星 ──────────────────────────────────────
      for (const s of stars) {
        // 漂移
        s.x += s.vx
        s.y += s.vy

        // 环绕
        if (s.x < -4) s.x = W + 4
        if (s.x > W + 4) s.x = -4
        if (s.y < -4) s.y = H + 4
        if (s.y > H + 4) s.y = -4

        // 闪烁：正弦波振荡
        s.alphaPhase += s.alphaSpeed
        const flicker = Math.sin(s.alphaPhase) * 0.18 * s.layer + Math.sin(s.alphaPhase * 2.3) * 0.06
        s.alpha = Math.max(0.02, Math.min(1, s.alphaBase + flicker))

        // 绘制星点
        ctx.beginPath()
        ctx.arc(s.x, s.y, s.size, 0, Math.PI * 2)
        ctx.fillStyle = s.color + s.alpha.toFixed(3) + ')'
        ctx.fill()

        // 近景星：加光晕（glow）
        if (s.layer === 2 && s.size > 1.6) {
          const grd = ctx.createRadialGradient(s.x, s.y, 0, s.x, s.y, s.size * 3.5)
          grd.addColorStop(0, s.color + (s.alpha * 0.4).toFixed(3) + ')')
          grd.addColorStop(1, s.color + '0)')
          ctx.beginPath()
          ctx.arc(s.x, s.y, s.size * 3.5, 0, Math.PI * 2)
          ctx.fillStyle = grd
          ctx.fill()
        }
      }

      // ── 流星 ──────────────────────────────────────
      if (frame >= nextMeteorAt) {
        meteors.push(createMeteor(W, H))
        nextMeteorAt = frame + Math.floor(Math.random() * 720 + 480)
      }

      meteors = meteors.filter(m => m.life > 0)
      for (const m of meteors) {
        const tailX = m.x - m.vx * (m.length / 10)
        const tailY = m.y - m.vy * (m.length / 10)

        const grd = ctx.createLinearGradient(tailX, tailY, m.x, m.y)
        grd.addColorStop(0, `rgba(255,255,255,0)`)
        grd.addColorStop(0.6, `rgba(220,235,255,${(m.life * 0.5).toFixed(3)})`)
        grd.addColorStop(1, `rgba(255,255,255,${(m.life * 0.9).toFixed(3)})`)

        ctx.beginPath()
        ctx.moveTo(tailX, tailY)
        ctx.lineTo(m.x, m.y)
        ctx.strokeStyle = grd
        ctx.lineWidth = m.life * 1.5
        ctx.stroke()

        m.x += m.vx
        m.y += m.vy
        m.life -= 0.016   // ~60fps 下约 1 秒消逝
      }

      raf = requestAnimationFrame(draw)
    }

    window.addEventListener('resize', resize)
    raf = requestAnimationFrame(draw)

    return () => {
      cancelAnimationFrame(raf)
      window.removeEventListener('resize', resize)
    }
  }, [])

  return (
    <canvas
      ref={canvasRef}
      style={{
        position: 'fixed',
        inset: 0,
        width: '100%',
        height: '100%',
        pointerEvents: 'none',
        zIndex: 0,
      }}
    />
  )
}
