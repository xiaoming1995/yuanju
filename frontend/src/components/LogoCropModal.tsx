import { useCallback, useEffect, useState } from 'react'
import Cropper from 'react-easy-crop'
import './LogoCropModal.css'

interface Props {
  sourceDataUrl: string
  open: boolean
  onConfirm: (file: File) => void
  onCancel: () => void
}

interface PixelArea {
  x: number
  y: number
  width: number
  height: number
}

const OUTPUT_SIZE = 256
const MAX_SOURCE_LONG_AXIS = 1600

export default function LogoCropModal({ sourceDataUrl, open, onConfirm, onCancel }: Props) {
  const [imgUrl, setImgUrl] = useState('')
  const [crop, setCrop] = useState({ x: 0, y: 0 })
  const [zoom, setZoom] = useState(1)
  const [areaPx, setAreaPx] = useState<PixelArea | null>(null)
  const [processing, setProcessing] = useState(false)

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

  const onCropComplete = useCallback((_: unknown, area: PixelArea) => {
    setAreaPx(area)
  }, [])

  if (!open) return null

  async function handleConfirm() {
    if (!areaPx || !imgUrl) return
    setProcessing(true)
    try {
      const blob = await cropToBlob(imgUrl, areaPx)
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
        <h3 className="logo-crop-title">调整 logo 裁剪区域</h3>

        <div className="logo-crop-canvas-area">
          {imgUrl ? (
            <Cropper
              image={imgUrl}
              crop={crop}
              zoom={zoom}
              aspect={1}
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

        <small className="logo-crop-note">动图（GIF / 动 WebP）将仅保留第一帧。输出 256×256 PNG。</small>

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

async function cropToBlob(src: string, area: PixelArea): Promise<Blob> {
  const img = await loadImage(src)
  const canvas = document.createElement('canvas')
  canvas.width = OUTPUT_SIZE
  canvas.height = OUTPUT_SIZE
  const ctx = canvas.getContext('2d')
  if (!ctx) throw new Error('canvas 2d 不可用')
  ctx.drawImage(
    img,
    area.x, area.y, area.width, area.height,
    0, 0, OUTPUT_SIZE, OUTPUT_SIZE,
  )
  return new Promise((resolve, reject) => {
    canvas.toBlob(
      blob => blob ? resolve(blob) : reject(new Error('toBlob 失败')),
      'image/png',
    )
  })
}
