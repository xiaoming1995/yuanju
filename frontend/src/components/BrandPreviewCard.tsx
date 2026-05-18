import type { ExportBrand } from '../lib/api'
import { resolveFooter, showDiagonalWatermark } from '../lib/brandText'

interface Props {
  brand: ExportBrand
}

/**
 * Lightweight preview of the export card with current brand settings.
 * Mocks four-pillar area with gray placeholder bars; only purpose is to
 * give the user a live visual of how title/logo/footer/watermark render.
 */
export default function BrandPreviewCard({ brand }: Props) {
  const title = brand.title || '缘 聚 命 理'
  const footer = resolveFooter(brand, 'yuanju.com')
  const showDiagonal = showDiagonalWatermark(brand)
  const isWordmark = brand.logo_mode === 'wordmark' && !!brand.logo_url

  return (
    <div style={{
      position: 'relative',
      width: 320,
      background: '#fdf9f2',
      fontFamily: '"Noto Serif SC", serif',
      overflow: 'hidden',
      borderRadius: 8,
      border: '1px solid #e0cca0',
    }}>
      {/* Header */}
      <div style={{
        position: 'relative',
        background: 'linear-gradient(135deg, #2d1f14 0%, #4a3020 50%, #3a2416 100%)',
        padding: '16px 20px 14px',
        textAlign: 'center',
      }}>
        {isWordmark ? (
          <img
            src={brand.logo_url}
            alt=""
            style={{
              display: 'block',
              margin: '0 auto',
              maxHeight: 32,
              maxWidth: 240,
              objectFit: 'contain',
            }}
          />
        ) : (
          <>
            {brand.logo_url && (
              <img
                src={brand.logo_url}
                alt=""
                style={{
                  position: 'absolute', left: 16, top: '50%', transform: 'translateY(-50%)',
                  width: 32, height: 32, objectFit: 'contain',
                }}
              />
            )}
            <div style={{ color: '#e8c97c', fontSize: 16, letterSpacing: 5, fontWeight: 700 }}>{title}</div>
          </>
        )}
        <div style={{ color: '#c4a06a', fontSize: 10, marginTop: 4 }}>预览（占位）</div>
      </div>
      {/* Body: gray placeholders */}
      <div style={{ padding: '18px 16px', display: 'flex', gap: 6 }}>
        {[0, 1, 2, 3].map(i => (
          <div key={i} style={{ flex: 1, height: 60, background: '#e8dcc8', borderRadius: 4 }} />
        ))}
      </div>
      {/* Footer */}
      <div style={{
        background: 'linear-gradient(135deg, #2d1f14 0%, #4a3020 100%)',
        padding: '10px 16px',
        display: 'flex',
        justifyContent: 'space-between',
        alignItems: 'center',
        fontSize: 10,
      }}>
        <span style={{ color: '#9a7a5a' }}>仅供参考，不作决策依据</span>
        <span style={{ color: '#e8c97c' }}>{footer}</span>
      </div>
      {/* Diagonal watermark overlay */}
      {showDiagonal && (
        <div style={{
          position: 'absolute', inset: 0, pointerEvents: 'none',
          overflow: 'hidden', zIndex: 1,
        }}>
          <div style={{
            position: 'absolute',
            top: '-30%', left: '-30%', right: '-30%', bottom: '-30%',
            transform: 'rotate(-30deg)',
            display: 'grid',
            gridTemplateColumns: 'repeat(auto-fill, 140px)',
            gap: '40px 30px',
            opacity: 0.08,
            color: '#000',
            fontSize: 12,
            whiteSpace: 'nowrap',
          }}>
            {Array.from({ length: 40 }).map((_, i) => (
              <span key={i}>{brand.watermark_text}</span>
            ))}
          </div>
        </div>
      )}
    </div>
  )
}

