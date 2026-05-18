import { useCallback, useEffect, useState } from 'react'
import Cropper from 'react-easy-crop'
import './LogoCropModal.css'

interface Props {
  sourceDataUrl: string
  open: boolean
  mode: 'icon' | 'wordmark'
  onConfirm: (file: File) => void
  onCancel: () => void
}

interface PixelArea {
  x: number
  y: number
  width: number
  height: number
}

const ICON_OUTPUT_SIZE = 256
const WORDMARK_OUTPUT_HEIGHT = 128
const WORDMARK_ASPECT_PRESETS = [2, 3, 4] as const
const DEFAULT_WORDMARK_ASPECT = 3
const MAX_SOURCE_LONG_AXIS = 1600

export default function LogoCropModal({ sourceDataUrl, open, mode, onConfirm, onCancel }: Props) {
  const [imgUrl, setImgUrl] = useState('')
  const [crop, setCrop] = useState({ x: 0, y: 0 })
  const [zoom, setZoom] = useState(1)
  const [areaPx, setAreaPx] = useState<PixelArea | null>(null)
  const [processing, setProcessing] = useState(false)
  const [wordmarkAspect, setWordmarkAspect] = useState<number>(DEFAULT_WORDMARK_ASPECT)

  // icon mode uses aspect={1} (square); wordmark uses selected ratio
  const cropperAspect = mode === 'icon' ? 1 : wordmarkAspect

  useEffect(() => {
    let cancelled = false
    if (!sourceDataUrl) {
      setImgUrl('')
      return
    }
    setCrop({ x: 0, y: 0 })
    setZoom(1)
    setAreaPx(null)
    downscaleIfLarge(sourceDataUrl)
      .then(url => { if (!cancelled) setImgUrl(url) })
      .catch(() => { if (!cancelled) setImgUrl(sourceDataUrl) })
    return () => { cancelled = true }
  }, [sourceDataUrl])

  // 切换比例 chip 时重置 crop 位置，让用户立刻看到新比例下的预览
  useEffect(() => {
    setCrop({ x: 0, y: 0 })
    setZoom(1)
    setAreaPx(null)
  }, [cropperAspect])

  const onCropComplete = useCallback((_: unknown, area: PixelArea) => {
    setAreaPx(area)
  }, [])

  if (!open) return null

  async function handleConfirm() {
    if (!areaPx || !imgUrl) return
    setProcessing(true)
    try {
      let outW: number, outH: number
      if (mode === 'icon') {
        outW = ICON_OUTPUT_SIZE
        outH = ICON_OUTPUT_SIZE
      } else {
        outH = WORDMARK_OUTPUT_HEIGHT
        outW = Math.round(outH * wordmarkAspect)
      }
      const blob = await cropToBlob(imgUrl, areaPx, outW, outH)
      const file = new File([blob], 'logo.png', { type: 'image/png' })
      onConfirm(file)
    } catch (err) {
      console.error('crop failed', err)
    } finally {
      setProcessing(false)
    }
  }

  return (
    <div className="logo-crop-overlay" onClick={onCancel} role="dialog" aria-modal="true">
      <div className="logo-crop-modal" onClick={e => e.stopPropagation()}>
        <h3 className="logo-crop-title">
          {mode === 'wordmark' ? '调整商标裁剪区域' : '调整 logo 裁剪区域'}
        </h3>

        {mode === 'wordmark' && (
          <div className="logo-crop-aspect-chips" role="radiogroup" aria-label="商标比例">
            {WORDMARK_ASPECT_PRESETS.map(a => (
              <button
                key={a}
                type="button"
                role="radio"
                aria-checked={wordmarkAspect === a}
                className={`logo-crop-aspect-chip${wordmarkAspect === a ? ' is-active' : ''}`}
                onClick={() => setWordmarkAspect(a)}
              >
                {a}:1
              </button>
            ))}
          </div>
        )}

        <div className="logo-crop-canvas-area">
          {imgUrl ? (
            <Cropper
              image={imgUrl}
              crop={crop}
              zoom={zoom}
              aspect={cropperAspect}
              cropShape="rect"
              showGrid
              onCropChange={setCrop}
              onZoomChange={setZoom}
              onCropComplete={onCropComplete}
            />
          ) : (
            <div className="logo-crop-loading">加载中...</div>
          )}
        </div>

        <div className="logo-crop-zoom-row">
          <span>缩放</span>
          <input
            type="range"
            min={1}
            max={3}
            step={0.01}
            value={zoom}
            onChange={e => setZoom(Number(e.target.value))}
          />
        </div>

        <small className="logo-crop-note">
          {mode === 'icon'
            ? '动图（GIF / 动 WebP）将仅保留第一帧。输出 256×256 PNG。'
            : '动图（GIF / 动 WebP）将仅保留第一帧。输出高 128 PNG，宽随比例（最多 512）。'}
        </small>

        <div className="logo-crop-actions">
          <button type="button" className="btn btn-ghost" onClick={onCancel} disabled={processing}>
            取消
          </button>
          <button
            type="button"
            className="btn btn-primary"
            onClick={handleConfirm}
            disabled={processing || !areaPx || !imgUrl}
          >
            {processing ? '处理中...' : '确认裁剪'}
          </button>
        </div>
      </div>
    </div>
  )
}

function loadImage(src: string): Promise<HTMLImageElement> {
  return new Promise((resolve, reject) => {
    const img = new Image()
    img.onload = () => resolve(img)
    img.onerror = () => reject(new Error('图片加载失败'))
    img.src = src
  })
}

async function downscaleIfLarge(dataUrl: string): Promise<string> {
  const img = await loadImage(dataUrl)
  const longest = Math.max(img.naturalWidth, img.naturalHeight)
  if (longest <= MAX_SOURCE_LONG_AXIS) return dataUrl
  const scale = MAX_SOURCE_LONG_AXIS / longest
  const canvas = document.createElement('canvas')
  canvas.width = Math.round(img.naturalWidth * scale)
  canvas.height = Math.round(img.naturalHeight * scale)
  const ctx = canvas.getContext('2d')
  if (!ctx) return dataUrl
  ctx.drawImage(img, 0, 0, canvas.width, canvas.height)
  return canvas.toDataURL('image/png')
}

async function cropToBlob(src: string, area: PixelArea, outW: number, outH: number): Promise<Blob> {
  const img = await loadImage(src)
  const canvas = document.createElement('canvas')
  canvas.width = outW
  canvas.height = outH
  const ctx = canvas.getContext('2d')
  if (!ctx) throw new Error('canvas 2d 不可用')
  ctx.drawImage(
    img,
    area.x, area.y, area.width, area.height,
    0, 0, outW, outH,
  )
  return new Promise((resolve, reject) => {
    canvas.toBlob(
      blob => blob ? resolve(blob) : reject(new Error('toBlob 失败')),
      'image/png',
    )
  })
}
