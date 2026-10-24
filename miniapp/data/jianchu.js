const jianChuNames = ['建', '除', '满', '平', '定', '执', '破', '危', '成', '收', '开', '闭']

// Short display explanations only. A single Jianchu value is not an activity verdict.
const meanings = {
  建: '起始、建立之象。',
  除: '清理、去旧之象。',
  满: '充盈、完备之象。',
  平: '平衡、调和之象。',
  定: '安定、确定之象。',
  执: '持守、执行之象。',
  破: '破旧、变动之象。',
  危: '谨慎、留有余地之象。',
  成: '完成、达成之象。',
  收: '收纳、收束之象。',
  开: '开启、通达之象。',
  闭: '闭合、休整之象。',
}

function getJianChuInfo(name) {
  if (!Object.prototype.hasOwnProperty.call(meanings, name)) {
    return { name: '', label: '暂无建除数据', meaning: '' }
  }
  return { name, label: `${name}日`, meaning: meanings[name] }
}

module.exports = { jianChuNames, getJianChuInfo }
