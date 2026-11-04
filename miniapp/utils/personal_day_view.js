function personalThemes(day) {
  if (!day || !Array.isArray(day.theme_guidance)) return []
  return day.theme_guidance.filter(item => ['noble', 'finance', 'romance', 'wellbeing'].includes(item.id) &&
    Array.isArray(item.sources) && item.sources.includes('personal_chart') &&
    ['candidate', 'review', 'neutral'].includes(item.level) && Array.isArray(item.evidence) && item.evidence.length)
}

function personalDirectionRows(day) {
  const labels = { favorable: '喜用', adverse: '忌神', neutral: '未落入喜忌' }
  if (!day || !labels[day.stem_direction] || !labels[day.branch_direction]) return []
  return [
    { name: `日干${day.day_gan}（${day.stem_element}）`, direction: labels[day.stem_direction] },
    { name: `日支${day.day_zhi}（${day.branch_element}主气）`, direction: labels[day.branch_direction] },
  ]
}

module.exports = { personalThemes, personalDirectionRows }
