# PDF 命理解读内容优化 · 设计

**日期**：2026-05-16
**范围**：`frontend/src/components/PrintLayout.tsx` + 抽出共享文本清理工具
**目的**：修两个 functional bug、补一个内容缺失、加一个工具书附录

## 背景

PDF 导出走 `PrintLayout.tsx`（445 行）。命理解读 section 直接渲染 `chapter.detail || chapter.brief`，未清理 markdown、未拆段落，没读新加的 `analysis.advice` 字段，没有术语词典帮普通读者理解专业名词，大运表里看不出当前大运是哪段。

screen 端已经在前几次 commit 修了 `**` 字符与段落分隔（0546cc7、3bbf98d）、加了 `analysis.advice`（0058f1b），PDF 这边漏了。

## 范围（4 项）

### P1 + P2 · 共享文本清理 + 段落渲染

- 把 `cleanReportText`（当前在 ResultPage.tsx 顶层私有）抽到共享文件 `frontend/src/lib/reportText.ts`，导出函数
- ResultPage.tsx 改用 import 不再自带定义
- PrintLayout.tsx 在所有渲染 AI 文本的地方（`analysis.logic`、`chapter.detail`/`chapter.brief`、`analysis.advice`）都用 `cleanReportText` 处理，并 `split(/\n{2,}/).filter(Boolean).map((para, idx) => <p key={idx}>{para}</p>)`

### P3 · 「行动建议」 PDF block

- 位置：「命局分析总览」之后、章节列表之前
- 数据源：`structured.analysis.advice`（无则不渲染）
- 样式：复用「命局分析总览」block 的视觉模板（lightBg + borderColor），标题 `▍ 行 动 建 议`
- 字号 12px serif、行距 1.85

### P4 · 末页术语词典

- 位置：「大运总览」之后、落款之前，`pageBreakBefore: 'always'` 独占新页或末半页
- 章节标题：`附 · 术 语 释 义`
- 词条数：8（现有 4：用神、忌神、格局、大运 + 新增 4：日主、十神、调候、流年）
- 布局：2 列网格（grid-template-columns: 1fr 1fr，gap: 10px 20px）
- 单条样式：左侧 term 11px gold + 右侧 desc 10px darkBrown，每条 padding 6px

### P5 · 大运总览表「本运 · 本年」高亮

- 计算：
  - `currentYear = new Date().getFullYear()`
  - `currentDayunIndex = dayun.findIndex(d => currentYear >= d.start_year && currentYear <= d.end_year)`
  - 当前流年干支：通过 `dayun[currentDayunIndex].liu_nian.find(ln => ln.year === currentYear)?.gan_zhi`（如果 PrintLayout 收到的 dayun 数据带 liu_nian）
- 在当前 dayun 行：
  - `<tr>` 背景换成 `#fff8e0`（淡金）
  - 「段」列前加朱红 `●` 圆点
  - 「起止年」列在原文本下方加一行小字 `本年 · {ganZhi}({currentYear})`

## 文件结构

```
新增
  frontend/src/lib/reportText.ts          导出 cleanReportText 共享函数

修改
  frontend/src/pages/ResultPage.tsx       import cleanReportText，删除本地定义
  frontend/src/components/PrintLayout.tsx 应用 cleanReportText + 段落拆分 + advice block + glossary + dayun 高亮
```

## 不在范围

- 章节内容字体（sans-serif → serif）统一（P6）
- 章节标题视觉权重提升（P7）
- 报告目录页（P8）
- 任何非 PDF 的改动
- 词典词条文案重写（沿用 ResultPage 现有的 REPORT_TERMS 描述）
- 大运表本身的列结构（不增不减列）

## 风险与缓解

| 风险 | 缓解 |
|---|---|
| cleanReportText 抽出后 ResultPage import 路径错误 | TypeScript `tsc -b --force` 强制检查 |
| PDF 段落变多导致页数溢出 | 现有 `pageBreakInside: 'avoid'` 保持，新加 advice block 也用同款 |
| 词典在小屏打印（A5）布局塌 | 2 列网格 + max-width，min 字号 9px |
| 当前大运高亮颜色 `#fff8e0` 黑白打印丢色 | 加红点 `●` 作为视觉锚，色丢了形还在 |

## 验收

- `tsc -b --force` 0 errors
- 所有现有 tests 通过（55+）
- 打印预览（Ctrl+P）：
  - 章节正文无 `**` 字符
  - 大运章三段（上一步/当前/下一步）有视觉空行
  - 「行动建议」block 出现在分析总览下方
  - 大运总览表当前大运行有背景色 + ● + 本年小字
  - 末页见「附 · 术语释义」8 条，2 列布局
- 黑白打印：现金/红色丢失但红点 ● 与表格分隔仍清晰

## 实现路径

由于改动量小（单 commit ~80-120 行差异）且全在一个文件 + 一个新文件，**跳过 writing-plans 的 15-task 分解**，直接做。验证 + commit + push 到 main。如果实施中发现意外复杂度（>200 行 / 多文件），再回头补 plan。
