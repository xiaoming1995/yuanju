// 六十甲子 · 日柱速写（静态附加层）。按日柱干支查一段性格/相处风格速写。
// 与 compatibilityPersonality.ts / 合盘评分 / LLM 完全解耦，仅供展示。
// 调性：双关微辣、不露骨。文案逐批撰写，见实现计划 Task 4。

export type DayPillarPortrait = {
  tag: string // 4-6 字定性钩子
  text: string // 速写正文（2-3 句）
}

const DAY_PILLAR_PORTRAITS: Record<string, DayPillarPortrait> = {
  甲子: {
    tag: '有耐心的狼',
    text: '白天一身正气、规矩得能领面锦旗，关了灯完全两个人。子水是沐浴之地——外头道貌岸然、里头野。慢热却记仇式专一，是只“有耐心的狼”，不动则已，一动就盯死你，耐力还出奇地好。',
  },
  乙丑: {
    tag: '冷面撩人',
    text: '长发一甩、腰肢一摆，天生勾人。丑是官杀库，女命坐下藏着一整支“后备队”——不是她招蜂引蝶，是异性自己排队上门；偏财一掺，浪漫起来不计成本。男命则是闷声撩人的惯犯，话不多、手不闲，冷着脸就把人带走了。',
  },
  壬申: {
    tag: '供大于求',
    text: '自坐长生，一眼活泉，越掏越有、越用越旺。精力、情绪、还有别的，统统源源不断，是“供大于求”型选手——鲜活、耐折腾，就是太满，得配个接得住的。',
  },
}

// 查表：命中返回速写，未知干支返回 undefined（旧数据/缺字段时由调用方跳过渲染）
export function getDayPillarPortrait(dayGan: string, dayZhi: string): DayPillarPortrait | undefined {
  return DAY_PILLAR_PORTRAITS[`${dayGan}${dayZhi}`]
}
