function dailyEncouragement(day, chart, themes) {
  if (!day || !chart || !day.date || !Array.isArray(day.evidence) || !day.evidence.length) return null
  if (!Array.isArray(themes) || themes.length !== 4 ||
    themes.some(item => !item || !['noble', 'finance', 'romance', 'wellbeing'].includes(item.id) ||
      !['candidate', 'review', 'neutral'].includes(item.level)) ||
    new Set(themes.map(item => item.id)).size !== 4) return null

  const address = chart.gender === 'male' ? '靓仔' : chart.gender === 'female' ? '靓女' : '朋友'
  const hasReview = themes.some(item => item.level === 'review')
  const hasCandidate = themes.some(item => item.level === 'candidate')
  const flatSentence = '今天适合稳稳发挥。照着计划走，手头的事一件件来。'
  let tone
  let sentence
  switch (day.status) {
    case 'supportive':
      if (hasReview) {
        tone = 'caution'
        sentence = '今天先稳一稳。重要安排多核对一步。'
      } else if (hasCandidate) {
        tone = 'supportive'
        sentence = '今天节奏挺顺。想推进的事，可以主动迈一步。'
      } else {
        tone = 'neutral'
        sentence = flatSentence
      }
      break
    case 'caution':
      tone = 'caution'
      sentence = '今天先稳一稳。留点余地，也能把事情办好。'
      break
    case 'mixed':
      tone = 'neutral'
      sentence = '今天按自己的节奏来。重要安排多核对一步。'
      break
    case 'undetermined':
      if (day.stem_direction === 'adverse' || day.branch_direction === 'adverse' || hasReview) {
        tone = 'caution'
        sentence = '今天先稳一稳。重要安排多核对一步。'
      } else if (hasCandidate && ((day.stem_direction === 'favorable' && day.branch_direction === 'neutral') ||
        (day.stem_direction === 'neutral' && day.branch_direction === 'favorable'))) {
        tone = 'supportive'
        sentence = '今天有点顺风。想推进的事，可以从容迈一步。'
      } else {
        tone = 'neutral'
        sentence = hasCandidate ? '今天按自己的节奏来。想做的事可以从容安排。' : flatSentence
      }
      break
    default:
      return null
  }
  return { tone, text: `${address}，${sentence}` }
}

module.exports = { dailyEncouragement }
