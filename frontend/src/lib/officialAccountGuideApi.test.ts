import { afterEach, expect, it, vi } from 'vitest'
import { AxiosError, type AxiosAdapter } from 'axios'
import adminApi, { adminOfficialAccountGuideAPI, type OfficialAccountGuide } from './adminApi'

const originalAdapter = adminApi.defaults.adapter
afterEach(() => { adminApi.defaults.adapter = originalAdapter; vi.unstubAllGlobals() })

it('uses the protected configuration and upload endpoints', async () => {
  vi.stubGlobal('localStorage', { getItem: () => 'admin-token' })
  const requests: Array<{ method?: string; url?: string; data: unknown; authorization: unknown }> = []
  adminApi.defaults.adapter = (async config => {
    requests.push({ method: config.method, url: config.url, data: config.data, authorization: config.headers?.Authorization })
    return { data: { data: { enabled: false, name: '', wechat_id: '', description: '', qr_path: '' } }, status: 200, statusText: 'OK', headers: {}, config }
  }) as AxiosAdapter
  const guide: OfficialAccountGuide = { enabled: false, name: '缘聚', wechat_id: 'yuanju', description: '内容更新', qr_path: '' }
  await adminOfficialAccountGuideAPI.get()
  await adminOfficialAccountGuideAPI.update(guide)
  await adminOfficialAccountGuideAPI.uploadQR(new File(['png'], 'qr.png', { type: 'image/png' }))
  expect(requests.map(({ method, url }) => [method, url])).toEqual([
    ['get', '/api/admin/settings/official-account-guide'],
    ['put', '/api/admin/settings/official-account-guide'],
    ['post', '/api/admin/settings/official-account-guide/qr'],
  ])
  expect(requests.every(request => request.authorization === 'Bearer admin-token')).toBe(true)
  expect(JSON.parse(requests[1].data as string)).toEqual(guide)
  expect((requests[2].data as FormData).get('file')).toBeInstanceOf(File)
})

it('preserves a save failure for the admin page to display', async () => {
  vi.stubGlobal('localStorage', { getItem: () => null })
  adminApi.defaults.adapter = (async () => { throw new AxiosError('offline', 'ERR_NETWORK') }) as AxiosAdapter
  await expect(adminOfficialAccountGuideAPI.update({ enabled: false, name: '', wechat_id: '', description: '', qr_path: '' }))
    .rejects.toMatchObject({ kind: 'network' })
})

it('preserves an upload failure for the admin page to display', async () => {
  vi.stubGlobal('localStorage', { getItem: () => null })
  adminApi.defaults.adapter = (async () => { throw new AxiosError('offline', 'ERR_NETWORK') }) as AxiosAdapter
  await expect(adminOfficialAccountGuideAPI.uploadQR(new File(['bad'], 'qr.png', { type: 'image/png' })))
    .rejects.toMatchObject({ kind: 'network' })
})
