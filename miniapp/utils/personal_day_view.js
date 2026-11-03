function personalHints(day) {
  if (!day || !Array.isArray(day.activity_guidance)) return []
  return day.activity_guidance.filter(item => Array.isArray(item.sources) &&
    item.sources.includes('personal_chart') && ['candidate', 'review'].includes(item.level))
}

function personalDirectionRows(day) {
  const labels = { favorable: '喜用', adverse: '忌神', neutral: '未落入喜忌' }
  if (!day || !labels[day.stem_direction] || !labels[day.branch_direction]) return []
  return [
    { name: `日干${day.day_gan}（${day.stem_element}）`, direction: labels[day.stem_direction] },
    { name: `日支${day.day_zhi}（${day.branch_element}主气）`, direction: labels[day.branch_direction] },
  ]
}

module.exports = { personalHints, personalDirectionRows }
