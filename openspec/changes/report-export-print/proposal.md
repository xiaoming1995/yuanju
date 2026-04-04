## Why

命理解读报告是缘聚平台的核心输出物，但目前用户只能在浏览器页面上查看，无法将完整的命盘数据与 AI 解读以文档形式保存或分享。增加「导出报告」功能，让用户可以一键将命盘 + AI 解读导出为 PDF，可显著提升报告的传播价值和用户留存。

## What Changes

- 在 `ResultPage.tsx` 报告卡片区域新增「📄 导出报告」按钮，点击调用 `window.print()`
- 在 `ResultPage.css` 新增 `@media print` 打印样式块，覆盖：
  - **隐藏**：顶部导航栏、「重新起盘」/「查看历史记录」底部按钮、精简/专业切换按钮、「生成 AI 解读」按钮、骨架屏加载态
  - **重置**：将深色主题 CSS Variables 替换为白底黑字（确保深色模式下打印可读）
  - **分页控制**：每个区块卡片设 `page-break-inside: avoid`，避免内容被切断
  - **品牌页脚**：通过 CSS `::after` 注入「缘聚命理 · 本报告由 AI 辅助生成，仅供参考」
- 无新外部依赖，无后端改动

## Capabilities

### New Capabilities

- `report-print-export`: 用户可在命盘结果页点击「导出报告」按钮，通过浏览器原生打印对话框将完整命盘（四柱）及 AI 命理解读导出为 PDF 文件，打印样式专门针对纸张版面优化

### Modified Capabilities

（无现有规格行为变更）

## Impact

- **前端**：`frontend/src/pages/ResultPage.tsx`（新增按钮）、`frontend/src/pages/ResultPage.css`（新增 @media print 样式）
- **无后端改动**，无新依赖
- **兼容性**：所有现代浏览器均支持 `window.print()`；深色主题用户打印时颜色自动重置为白底黑字
