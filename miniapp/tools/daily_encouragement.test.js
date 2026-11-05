const assert = require('node:assert/strict')
const test = require('node:test')
const { dailyEncouragement } = require('../utils/daily_encouragement')

const day = { date: '2026-11-05', evidence: ['当日日干与本人命盘喜用同向。'] }
const flatThemes = ['noble', 'finance', 'romance', 'wellbeing'].map(id => ({ id, level: 'neutral' }))
const candidateThemes = flatThemes.map(item => item.id === 'noble' ? { ...item, level: 'candidate' } : item)
const reviewThemes = candidateThemes.map(item => item.id === 'finance' ? { ...item, level: 'review' } : item)

test('supportive guidance is encouraging without promising an outcome', () => {
  const result = dailyEncouragement({ ...day, status: 'supportive' }, { gender: 'male' }, candidateThemes)
  assert.equal(result.text, '靓仔，今天节奏挺顺。想推进的事，可以主动迈一步。')
  assert.equal(result.tone, 'supportive')
  assert.doesNotMatch(result.text, /直接冲|必|保证|投资|买入/)
})

test('mixed and caution states keep a measured tone', () => {
  const mixed = dailyEncouragement({ ...day, status: 'mixed' }, { gender: 'female' }, candidateThemes)
  const caution = dailyEncouragement({ ...day, status: 'caution' }, { gender: 'female' }, candidateThemes)
  assert.match(mixed.text, /^靓女，今天按自己的节奏来/)
  assert.match(caution.text, /^靓女，今天先稳一稳/)
})

test('one adverse direction does not become a positive verdict', () => {
  const result = dailyEncouragement({ ...day, status: 'undetermined', stem_direction: 'adverse', branch_direction: 'neutral' }, { gender: 'unknown' }, candidateThemes)
  assert.equal(result.tone, 'caution')
  assert.match(result.text, /^朋友，今天先稳一稳/)
})

test('one favorable direction without an adverse direction gets a light positive greeting', () => {
  for (const directions of [
    { stem_direction: 'favorable', branch_direction: 'neutral' },
    { stem_direction: 'neutral', branch_direction: 'favorable' },
  ]) {
    const result = dailyEncouragement({ ...day, status: 'undetermined', ...directions }, { gender: 'male' }, candidateThemes)
    assert.equal(result.tone, 'supportive')
    assert.equal(result.text, '靓仔，今天有点顺风。想推进的事，可以从容迈一步。')
  }
  const neutral = dailyEncouragement({ ...day, status: 'undetermined', stem_direction: 'neutral', branch_direction: 'neutral' }, { gender: 'male' }, candidateThemes)
  assert.equal(neutral.tone, 'neutral')
  const incomplete = dailyEncouragement({ ...day, status: 'undetermined', stem_direction: 'favorable' }, { gender: 'male' }, candidateThemes)
  assert.equal(incomplete.tone, 'neutral')
})

test('four flat themes override a favorable stem or even a supportive relation', () => {
  for (const result of [
    dailyEncouragement({ ...day, status: 'undetermined', stem_direction: 'favorable', branch_direction: 'neutral' }, { gender: 'male' }, flatThemes),
    dailyEncouragement({ ...day, status: 'supportive' }, { gender: 'male' }, flatThemes),
  ]) {
    assert.equal(result.tone, 'neutral')
    assert.equal(result.text, '靓仔，今天适合稳稳发挥。照着计划走，手头的事一件件来。')
  }
})

test('a cautious theme blocks a positive greeting', () => {
  const result = dailyEncouragement({ ...day, status: 'supportive' }, { gender: 'female' }, reviewThemes)
  assert.equal(result.tone, 'caution')
})

test('missing evidence and unsupported states produce no personal greeting', () => {
  assert.equal(dailyEncouragement({ ...day, status: 'insufficient' }, { gender: 'male' }, candidateThemes), null)
  assert.equal(dailyEncouragement({ ...day, status: 'new-status' }, { gender: 'male' }, candidateThemes), null)
  assert.equal(dailyEncouragement({ ...day, status: 'supportive', evidence: [] }, { gender: 'male' }, candidateThemes), null)
  assert.equal(dailyEncouragement({ ...day, status: 'supportive' }, null, candidateThemes), null)
  assert.equal(dailyEncouragement({ ...day, status: 'supportive' }, { gender: 'male' }, []), null)
  assert.equal(dailyEncouragement({ ...day, status: 'supportive' }, { gender: 'male' }, [null, ...flatThemes.slice(1)]), null)
})
