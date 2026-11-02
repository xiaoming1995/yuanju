const assert = require('node:assert/strict')
const { readFileSync } = require('node:fs')
const { join } = require('node:path')
const test = require('node:test')
const { runInNewContext } = require('node:vm')

const source = readFileSync(join(__dirname, '../config.js'), 'utf8')

function apiBaseUrl(wx) {
  const module = { exports: {} }
  runInNewContext(source, { module, wx })
  return module.exports.apiBaseUrl
}

test('development uses local backend on simulator and LAN backend on phone', () => {
  const develop = platform => ({
    getAccountInfoSync: () => ({ miniProgram: { envVersion: 'develop' } }),
    getDeviceInfo: () => ({ platform }),
  })
  assert.equal(apiBaseUrl(develop('devtools')), 'http://127.0.0.1:9002')
  assert.equal(apiBaseUrl(develop('ios')), 'http://192.168.101.24:9002')
})

test('trial and release always use public HTTPS, including in developer tools', () => {
  for (const envVersion of ['trial', 'release']) {
    const wx = {
      getAccountInfoSync: () => ({ miniProgram: { envVersion } }),
      getDeviceInfo: () => ({ platform: 'devtools' }),
    }
    assert.equal(apiBaseUrl(wx), 'https://www.guanshilu.cn')
  }
})

test('unknown phone environment defaults to public HTTPS', () => {
  assert.equal(apiBaseUrl({ getDeviceInfo: () => ({ platform: 'android' }) }), 'https://www.guanshilu.cn')
  assert.equal(apiBaseUrl({ getAccountInfoSync: () => { throw new Error('unavailable') } }), 'https://www.guanshilu.cn')
})
