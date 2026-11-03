import { describe, expect, it } from 'vitest'
import { missingOfficialAccountFields } from './officialAccountGuideForm'

describe('official account guide form', () => {
  it('allows an empty disabled draft and retains fields for later editing', () => {
    expect(missingOfficialAccountFields({ enabled: false, name: '', wechat_id: '', description: '', qr_path: '' })).toEqual([])
  })

  it('lists every missing field before enabling', () => {
    expect(missingOfficialAccountFields({ enabled: true, name: '  ', wechat_id: '', description: '', qr_path: '' }))
      .toEqual(['公众号名称', '微信号', '引导说明', '二维码'])
  })

  it('accepts a fully configured guide', () => {
    expect(missingOfficialAccountFields({ enabled: true, name: '缘聚', wechat_id: 'yuanju', description: '内容更新', qr_path: '/static/uploads/official-account-qr/example.png' })).toEqual([])
  })
})
