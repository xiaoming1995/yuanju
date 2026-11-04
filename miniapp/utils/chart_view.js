const tones = { 木: 'wood', 火: 'fire', 土: 'earth', 金: 'metal', 水: 'water' }
const tone = element => tones[element] || 'neutral'
const labels = ['年柱', '月柱', '日柱', '时柱']
const keys = ['year', 'month', 'day', 'hour']

function toHistory(chart) {
  return {
    id: chart.id,
    name: chart.display_name || '我的干支',
    date: `${chart.birth_year}-${String(chart.birth_month).padStart(2, '0')}-${String(chart.birth_day).padStart(2, '0')} ${String(chart.birth_hour).padStart(2, '0')}:00`,
    columns: keys.map(key => ({
      key,
      gan: chart[`${key}_gan`],
      zhi: chart[`${key}_zhi`],
      text: `${chart[`${key}_gan`]}${chart[`${key}_zhi`]}`,
      tone: 'neutral',
    })),
  }
}

function toResult(result, input) {
  const pillars = keys.map((key, index) => ({
    label: labels[index], gan: result[`${key}_gan`], zhi: result[`${key}_zhi`],
    ganWuxing: result[`${key}_gan_wuxing`], zhiWuxing: result[`${key}_zhi_wuxing`],
    ganTone: tone(result[`${key}_gan_wuxing`]), zhiTone: tone(result[`${key}_zhi_wuxing`]),
    ganShiShen: result[`${key}_gan_shishen`], hideGan: result[`${key}_hide_gan`] || [],
    zhiShiShen: result[`${key}_zhi_shishen`] || [], xingYun: result[`${key}_xing_yun`],
    diShi: result[`${key}_di_shi`], xunKong: result[`${key}_xun_kong`],
    naYin: result[`${key}_na_yin`], shenSha: result[`${key}_shen_sha`] || [],
  }))
  const fields = [
    ['主星', 'ganShiShen'], ['天干', 'gan'], ['地支', 'zhi'], ['藏干', 'hideGan'],
    ['副星', 'zhiShiShen'], ['星运', 'xingYun'], ['自坐', 'diShi'], ['空亡', 'xunKong'],
    ['纳音', 'naYin'], ['神煞', 'shenSha'],
  ]
  const basicRows = fields.map(([label, field]) => ({
    label, kind: field === 'gan' || field === 'zhi' ? 'main' : Array.isArray(pillars[0][field]) ? 'list' : 'text',
    cells: pillars.map(p => ({
      pillar: p.label, text: Array.isArray(p[field]) ? p[field].join(' ') : p[field],
      items: Array.isArray(p[field]) ? p[field] : [],
      wuxing: field === 'gan' ? p.ganWuxing : field === 'zhi' ? p.zhiWuxing : '',
      tone: field === 'gan' ? p.ganTone : field === 'zhi' ? p.zhiTone : '',
    })),
  }))
  const stats = result.wuxing || {}
  const wuxingRows = [['木', 'mu'], ['火', 'huo'], ['土', 'tu'], ['金', 'jin'], ['水', 'shui']].map(([label, key]) => ({
    label, tone: tone(label), count: stats[key] || 0,
    percent: stats.total ? Math.round((stats[key] || 0) * 100 / stats.total) : 0,
  }))
  const assessment = result.natal_assessment || {}
  const fuyi = assessment.fuyi || {}
  return {
    input: { dateTime: input.dateText, gender: input.gender }, pillars, basicRows,
    wuxingRows, wuxingTotal: stats.total || 0,
    mingGe: result.ming_ge, mingGeDesc: result.ming_ge_desc,
    stemGuidance: assessment.stem_guidance || { primary_favorable: [], secondary_favorable: [], adverse: [] },
    personalDay: { fuyiYongshen: fuyi.yongshen || result.yongshen, fuyiJishen: fuyi.jishen || result.jishen },
  }
}

module.exports = { toHistory, toResult }
