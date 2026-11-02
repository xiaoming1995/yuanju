function personalHints(day) {
  if (!day || !Array.isArray(day.activity_guidance)) return []
  return day.activity_guidance.filter(item => Array.isArray(item.sources) &&
    item.sources.includes('personal_chart') && ['candidate', 'review'].includes(item.level))
}

module.exports = { personalHints }
