import type { OfficialAccountGuide } from '../../lib/adminApi'

export function missingOfficialAccountFields(guide: OfficialAccountGuide): string[] {
  if (!guide.enabled) return []
  const missing: string[] = []
  if (!guide.name.trim()) missing.push('公众号名称')
  if (!guide.wechat_id.trim()) missing.push('微信号')
  if (!guide.description.trim()) missing.push('引导说明')
  if (!guide.qr_path.trim()) missing.push('二维码')
  return missing
}
