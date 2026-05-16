// AI 报告文本清理与拆段共享工具
//
// AI 输出的章节正文/分析段经常带 Markdown 加粗 / 斜体符号（** __ *）以及
// 用 \n\n 模拟段落分隔。我们 screen 与 PDF 两条渲染路径都需要把这些标记
// 去掉，并按段落拆开。集中在这里，避免两处实现漂移。

export function cleanReportText(s: string | undefined | null): string {
  if (!s) return ''
  return s
    .replace(/\*\*/g, '')   // 加粗 **
    .replace(/__/g, '')     // 加粗 __
    .replace(/(?<!\w)\*(?!\s)|(?<!\s)\*(?!\w)/g, '')  // 单 * 斜体（保守，不误删数学符号）
    .trim()
}

/** 把已清理的文本按 \n\n 拆成多段，过滤空段。 */
export function splitParagraphs(s: string | undefined | null): string[] {
  return cleanReportText(s)
    .split(/\n{2,}/)
    .map(p => p.trim())
    .filter(Boolean)
}
