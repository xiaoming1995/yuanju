## 1. 导出按钮

- [x] 1.1 在 `ResultPage.tsx` 的 AI 报告区 `<h2>` 标题行右侧，新增「📄 导出报告」按钮（仅当 `report` 存在时渲染）
- [x] 1.2 按钮 `onClick` 调用 `window.print()`，class 使用已有的 `btn btn-ghost btn-sm` 样式

## 2. 打印颜色与全局样式重置

- [x] 2.1 在 `ResultPage.css` 末尾新增 `@media print` 块
- [x] 2.2 在 `@media print` 内覆盖根级 CSS Variables：`--bg-main: #fff`、`--bg-surface: #f9f8f6`、`--text-color: #111`、`--text-muted: #555`、`--border-color: #ddd`、`--accent-color: #8b6e4e`
- [x] 2.3 设置 `body { background: #fff; color: #111; }` 确保白底黑字

## 3. 隐藏交互元素

- [x] 3.1 在 `@media print` 中将以下元素设为 `display: none !important`：
  - `.navbar`（顶部导航）
  - `.result-footer`（底部按钮区：重新起盘 / 查看历史）
  - `.report-mode-switcher`（精简/专业切换按钮组）
  - `.report-cta`（生成 AI 解读 CTA 区域）
  - `.guest-banner`（游客引导注册条）
  - `#export-report-btn`（导出按钮自身，确保不出现在打印版中）
- [x] 3.2 设置 `.animate-fade-up { animation: none; opacity: 1; }` 消除打印时动画残留

## 4. 分页控制

- [x] 4.1 对所有 `.card` 元素设置 `page-break-inside: avoid; break-inside: avoid`
- [x] 4.2 对 `.report-block` 设置 `page-break-inside: avoid; break-inside: avoid`
- [x] 4.3 对 `.dayun-section` 设置 `page-break-before: always`（大运时间轴内容较长，独立成页）

## 5. 品牌页脚

- [x] 5.1 在 `@media print` 中为 `.result-page::after` 添加伪元素样式：
  - `content: "缘聚命理 · 本报告由 AI 辅助生成，仅供参考 · yuanju.com"`
  - `display: block; text-align: center; font-size: 11px; color: #999; margin-top: 32px; border-top: 1px solid #eee; padding-top: 12px;`

## 6. 验证

- [x] 6.1 在有 AI 报告状态下，确认「导出报告」按钮显示在标题行右侧
- [x] 6.2 在无 AI 报告状态下，确认按钮不显示
- [x] 6.3 点击按钮，浏览器弹出打印对话框
- [x] 6.4 在打印预览中确认：导航栏/按钮不可见，白底黑字，内容完整
- [x] 6.5 在深色主题下触发打印，确认颜色正确重置为白底黑字
- [x] 6.6 确认页面底部出现品牌页脚文字
