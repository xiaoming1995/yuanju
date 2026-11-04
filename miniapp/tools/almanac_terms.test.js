const assert = require('node:assert/strict')
const fs = require('node:fs')
const path = require('node:path')
const test = require('node:test')
const { execFileSync } = require('node:child_process')
const { meanings, getAlmanacTermMeaning, showAlmanacTerm } = require('../data/almanac_terms')

test('every lunar-go v1.4.6 day almanac label has one nonempty explanation', () => {
  const moduleCache = execFileSync('go', ['env', 'GOMODCACHE'], { encoding: 'utf8' }).trim()
  const source = fs.readFileSync(path.join(moduleCache, 'github.com/6tail/lunar-go@v1.4.6/LunarUtil/LunarUtil.go'), 'utf8')
  const match = source.match(/var yiJi = \[\]string\{([^}]*)\}/)
  assert.ok(match, 'lunar-go day almanac label list is present')
  const labels = [...match[1].matchAll(/"([^"]+)"/g)].map(item => item[1])
  assert.equal(labels.length, 142)
  assert.deepEqual(Object.keys(meanings).sort(), labels.sort())
  for (const label of labels) assert.ok(meanings[label].trim().length >= 3, label)
})

test('exact lookup and unknown labels do not manufacture meanings', () => {
  assert.match(getAlmanacTermMeaning('纳采'), /提亲/)
  assert.match(getAlmanacTermMeaning('破屋坏垣'), /旧屋和围墙/)
  assert.match(getAlmanacTermMeaning('新增原词'), /释义待补充/)
})

test('tap explanation preserves the original yi or ji context', () => {
  const calls = []
  global.wx = { showModal(options) { calls.push(options) } }
  showAlmanacTerm({ currentTarget: { dataset: { term: '开市', kind: '忌' } } })
  assert.equal(calls.length, 1)
  assert.equal(calls[0].title, '忌 · 开市')
  assert.match(calls[0].content, /通用黄历日期资料/)
  assert.match(calls[0].content, /投资交易时机/)
  assert.equal(calls[0].showCancel, false)
})

test('all three real almanac lists bind each original label to the shared explanation', () => {
  const pages = [
    ['home', 'bindtap'], ['personal-day', 'bindtap'], ['moving-date', 'catchtap'],
  ]
  for (const [name, binding] of pages) {
    const base = path.join(__dirname, '../pages', name, 'index')
    const markup = fs.readFileSync(`${base}.wxml`, 'utf8')
    const code = fs.readFileSync(`${base}.js`, 'utf8')
    assert.equal((markup.match(new RegExp(`${binding}="showAlmanacTerm"`, 'g')) || []).length, 2, name)
    assert.match(markup, /data-kind="宜"/)
    assert.match(markup, /data-kind="忌"/)
    assert.match(code, /require\('\.\.\/\.\.\/data\/almanac_terms'\)/)
  }
})
