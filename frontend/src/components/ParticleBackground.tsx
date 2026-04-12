import { useEffect, useRef } from 'react'

interface Particle {
  x: number
  y: number
  vx: number
  vy: number
  size: number
  baseAlpha: number
  alpha: number
  alphaDir: number
  alphaSpeed: number
  color: string
}

// 五行配色粒子颜色（极低饱和度，以免抢戏）
const COLORS = [
  'rgba(201,168,76,',   // 金
  'rgba(91,155,213,',   // 水
  'rgba(76,175,125,',   // 木
  'rgba(224,92,75,',    // 火（极少）
  'rgba(193,127,62,',   // 土
]

const COLOR_WEIGHTS = [0.40, 0.30, 0.18, 0.05, 0.07] // 金>水>木>土>火

function weightedColor(): string {
  const r = Math.random()
  let acc = 0
  for (let i = 0; i < COLOR_WEIGHTS.length; i++) {
    acc += COLOR_WEIGHTS[i]
    if (r < acc) return COLORS[i]
  }
  return COLORS[0]
}

function createParticle(W: number, H: number): Particle {
  const baseAlpha = Math.random() * 0.35 + 0.05 // 0.05 ~ 0.40
  return {
    x: Math.random() * W,
    y: Math.random() * H,
    vx: (Math.random() - 0.5) * 0.18,
    vy: (Math.random() - 0.5) * 0.18,
    size: Math.random() * 1.6 + 0.4, // 0.4 ~ 2px
    baseAlpha,
    alpha: baseAlpha,
    alphaDir: Math.random() > 0.5 ? 1 : -1,
    alphaSpeed: Math.random() * 0.003 + 0.001,
    color: weightedColor(),
  }
}

export default function ParticleBackground() {
  const canvasRef = useRef<HTMLCanvasElement>(null)

  useEffect(() => {
    // 尊重用户「减少动画」偏好
    if (window.matchMedia('(prefers-reduced-motion: reduce)').matches) return

    const canvas = canvasRef.current
    if (!canvas) return
    const ctx = canvas.getContext('2d')
    if (!ctx) return

    let W = window.innerWidth
    let H = window.innerHeight
    const PARTICLE_COUNT = Math.min(Math.floor((W * H) / 14000), 120)

    canvas.width = W
    canvas.height = H

    let particles: Particle[] = Array.from({ length: PARTICLE_COUNT }, () =>
      createParticle(W, H)
    )

    let raf: number

    const resize = () => {
      W = window.innerWidth
      H = window.innerHeight
      canvas.width = W
      canvas.height = H
      // 重建粒子，给新分辨率
      particles = Array.from({ length: Math.min(Math.floor((W * H) / 14000), 120) }, () =>
        createParticle(W, H)
      )
    }

    const draw = () => {
      ctx.clearRect(0, 0, W, H)

      for (const p of particles) {
        // 位移
        p.x += p.vx
        p.y += p.vy

        // 边界环绕
        if (p.x < -4) p.x = W + 4
        if (p.x > W + 4) p.x = -4
        if (p.y < -4) p.y = H + 4
        if (p.y > H + 4) p.y = -4

        // 呼吸（alpha 振荡）
        p.alpha += p.alphaDir * p.alphaSpeed
        if (p.alpha >= p.baseAlpha + 0.12) p.alphaDir = -1
        if (p.alpha <= Math.max(0.02, p.baseAlpha - 0.12)) p.alphaDir = 1

        // 绘制
        ctx.beginPath()
        ctx.arc(p.x, p.y, p.size, 0, Math.PI * 2)
        ctx.fillStyle = p.color + p.alpha.toFixed(3) + ')'
        ctx.fill()
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
