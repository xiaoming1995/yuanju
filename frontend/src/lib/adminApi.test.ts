import { afterEach, describe, expect, it, vi } from 'vitest'
import { AxiosError, type AxiosResponse } from 'axios'
import adminApi, { adminActionErrorMessage, toAdminRequestError } from './adminApi'

afterEach(() => vi.unstubAllGlobals())

describe('admin request errors', () => {
  it('distinguishes a timeout and warns that a write result is uncertain', () => {
    const error = toAdminRequestError(new AxiosError('timeout', 'ECONNABORTED'))
    expect(error.kind).toBe('timeout')
    expect(error.message).toContain('连接超时')
    expect(adminActionErrorMessage(error)).toContain('结果尚未确认')
  })

  it('keeps a server error and status for a failed request', () => {
    const response = { status: 503, data: { error: '服务暂不可用' } } as AxiosResponse<{ error: string }>
    const error = toAdminRequestError(new AxiosError('server error', 'ERR_BAD_RESPONSE', undefined, undefined, response))
    expect(error.kind).toBe('http')
    expect(error.status).toBe(503)
    expect(error.message).toBe('服务暂不可用')
  })

  it('gives a useful message for a disconnected request', () => {
    const error = toAdminRequestError(new AxiosError('Network Error', 'ERR_NETWORK'))
    expect(error.kind).toBe('network')
    expect(error.message).toContain('检查网络')
  })

  it('passes timeout and disconnected errors through the admin interceptor', async () => {
    vi.stubGlobal('localStorage', { getItem: () => null })
    for (const code of ['ECONNABORTED', 'ERR_NETWORK']) {
      await expect(adminApi.get('/api/admin/stats', {
        adapter: async () => { throw new AxiosError('simulated', code) },
      })).rejects.toMatchObject({ kind: code === 'ECONNABORTED' ? 'timeout' : 'network' })
    }
  })
})
