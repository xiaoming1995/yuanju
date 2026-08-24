import { renderToStaticMarkup } from 'react-dom/server'
import { describe, expect, it } from 'vitest'
import ShareCard from './ShareCard'
import type { ShareCardProps } from './ShareCard'

const baseProps: ShareCardProps = {
  birthYear: 1995,
  birthMonth: 10,
  birthDay: 12,
  birthHour: 12,
  gender: 'male',
  yearGan: '乙',
  yearZhi: '亥',
  monthGan: '丙',
  monthZhi: '戌',
  dayGan: '丁',
  dayZhi: '丑',
  hourGan: '丙',
  hourZhi: '午',
  yearGanWx: '木',
  yearZhiWx: '水',
  monthGanWx: '火',
  monthZhiWx: '土',
  dayGanWx: '火',
  dayZhiWx: '土',
  hourGanWx: '火',
  hourZhiWx: '火',
  structured: {
    yongshen: '水',
    jishen: '火',
    analysis: { logic: '', summary: '' },
    chapters: [],
  },
}

describe('ShareCard past-events section', () => {
  it('没有已生成过往内容时不显示过往模块', () => {
    const html = renderToStaticMarkup(<ShareCard {...baseProps} pastEventsExportSegments={[]} />)

    expect(html).not.toContain('过 往 年 运 回 看')
  })

  it('显示已生成过往内容且不带依据明细', () => {
    const html = renderToStaticMarkup(
      <ShareCard
        {...baseProps}
        pastEventsExportSegments={[
          {
            dayun_index: 1,
            gan_zhi: '乙酉',
            start_age: 2,
            end_age: 11,
            start_year: 1996,
            end_year: 2005,
            themes: ['学业贵人加持'],
            summary: '本步大运已有总结。',
            years: [
              {
                year: 1996,
                age: 2,
                gan_zhi: '丙子',
                narrative: '这一年有明显环境变化。',
                signals: ['迁变'],
                evidence_summary: ['专业证据不应进入分享图'],
              },
            ],
          },
        ]}
      />,
    )

    expect(html).toContain('过 往 年 运 回 看')
    expect(html).toContain('本步大运已有总结。')
    expect(html).toContain('这一年有明显环境变化。')
    expect(html).not.toContain('专业证据不应进入分享图')
  })
})
