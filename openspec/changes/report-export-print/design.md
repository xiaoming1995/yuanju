## Context

缘聚结果页（`ResultPage.tsx`）包含多个数据区块：生辰标题、四柱数据网格、五行雷达图（SVG）、命元特质、专属命理头像（SVG）、大运时间轴、AI 解读报告。页面使用深色主题 CSS Variables。目前没有任何导出能力。

探索阶段评估了两条技术路径后，选择了基于浏览器原生 `window.print()` API + `@media print` 专属样式表的方案，不引入任何新外部依赖。

## Goals / Non-Goals

**Goals:**
- 用户点击按钮即可触发浏览器打印对话框，可直接「另存为 PDF」
- 打印版面为白底黑字，中文内容正常渲染，SVG 图表完整显示
- 隐藏与打印无关的交互元素（导航、按钮等）
- 每个卡片区块避免被分页截断
- 页面底部自动注入品牌页脚

**Non-Goals:**
- 服务器端 PDF 生成（Streaming/Puppeteer 等）
- 图片格式导出（PNG/JPG）
- 引入 `html2canvas` / `jsPDF` 等第三方依赖
- 自定义纸张尺寸或页边距（让浏览器处理）

## Decisions

### D1：选择 `window.print()` 而非 `html2canvas + jsPDF`

当前结果页包含多个复杂 SVG 组件（WuxingRadar、MingpanAvatar）和大量 CSS Variables（深色主题），`html2canvas` 对 SVG 和 CSS Variables 支持均较差，极大概率产生颜色丢失、中文字体回退、SVG 渲染空白等问题。

浏览器原生打印 API 则原生支持 SVG、中文字体、CSS Variables（可在 `@media print` 中覆盖为兼容值），且打印生成的 PDF 质量（矢量文字、分辨率）远高于 `html2canvas` 位图截图方案。

### D2：打印样式覆盖策略

```css
@media print {
  /* 1. 隐藏交互元素 */
  .navbar, .result-footer, .report-mode-switcher,
  .report-cta, .guest-banner { display: none !important; }

  /* 2. 重置颜色系统为白底黑字（兼容深色主题） */
  :root { --bg-main: #fff; --bg-surface: #f9f9f9; --text-color: #111; ... }

  /* 3. 分页控制 */
  .card { page-break-inside: avoid; }
  .report-block { page-break-inside: avoid; }

  /* 4. 品牌页脚 */
  .result-page::after {
    content: "缘聚命理 · 本报告由 AI 辅助生成，仅供参考 · yuanju.com";
    display: block;
    text-align: center;
    font-size: 11px;
    color: #999;
    margin-top: 32px;
    border-top: 1px solid #eee;
    padding-top: 12px;
  }
}
```

### D3：导出按钮位置

放在 AI 报告区标题行右侧（与精简/专业切换按钮同排），仅当 `report` 存在时显示（无报告时不显示导出按钮）。

## Risks / Trade-offs

- **[深色主题切换]** 用户若在深色主题下打印 → Mitigation：@media print 中强制重置所有 CSS Variables 为明色值
- **[分页不完美]** 大运时间轴较长可能仍有断行 → Mitigation：对 DayunTimeline 整体设 `page-break-before: always` 另起一页
- **[浏览器差异]** Safari / Firefox 对某些 @media print CSS 支持略有差异 → Mitigation：避免使用实验性属性，保持 CSS 简单
- **[SVG 打印]** 部分旧版浏览器 SVG 在打印时可能缩放异常 → 暂接受，目标用户主要使用现代浏览器

## Open Questions

- 无，探索阶段已完成所有决策
