import type { ExportBrand } from './api'

/**
 * Resolve the bottom-right footer text for an export card.
 *
 * Precedence:
 *   1. If watermark_mode='bottom' and BOTH watermark_text and footer_text are set
 *      → join them with " · "
 *   2. If watermark_mode='bottom' and only watermark_text is set → watermark_text
 *   3. Otherwise → footer_text if non-empty, else fallback
 *
 * The fallback differs by component:
 *   - ShareCard / BrandPreviewCard: "yuanju.com" (PNG export default)
 *   - PrintLayout: "缘 聚 命 理" (preserves the PDF's original footer mark)
 *
 * Each component passes its own fallback to preserve pre-existing
 * default behavior for users who haven't customized.
 */
export function resolveFooter(brand: ExportBrand | null | undefined, fallback: string): string {
  if (!brand) return fallback
  if (brand.watermark_mode === 'bottom' && brand.watermark_text && brand.footer_text) {
    return `${brand.footer_text} · ${brand.watermark_text}`
  }
  if (brand.watermark_mode === 'bottom' && brand.watermark_text) {
    return brand.watermark_text
  }
  return brand.footer_text || fallback
}

/** Whether the diagonal watermark overlay should render. */
export function showDiagonalWatermark(brand: ExportBrand | null | undefined): boolean {
  return brand?.watermark_mode === 'diagonal' && (brand?.watermark_text?.length ?? 0) > 0
}
