# PDF 打印布局优化

## 问题

当前 PDF 导出（`window.print()`）存在以下问题：
1. `@media print` CSS 覆写的变量名与 app 实际使用的变量名不匹配，导致暗色背景直接打印成黑色
2. `.bottom-nav`、`<ShareCard>` DOM 等元素未在打印时隐藏，出现在 PDF 中
3. 品牌落款文案仍为旧版"AI 辅助生成"
4. 无独立的打印视图，打印结果依赖屏幕布局，格式难以控制

## 方案

新增 `<PrintLayout>` 组件，专为打印设计：
- `@media screen` 时 `display: none`（不影响屏幕视图）
- `@media print` 时显示，同时将屏幕内容 `.screen-only` 全部隐藏
- 白底黑字，使用 inline styles 避免 CSS 变量继承问题
- 内容顺序：品牌头部 → 四柱 → 喜用神/忌神 → 神煞（内联注解）→ 命理解读章节 → 大运竖向列表 → 品牌落款

## 设计决策

| 项目 | 决定 |
|------|------|
| 大运格式 | 竖向列表（每大运一行） |
| 无报告时 | 正常导出，报告区显示"命理解读尚未生成"占位文字 |
| 神煞标注 | 内联展开注解说明 |
| 五行雷达图 | 不展示 |
| 调候用神 | 不展示 |

## 范围

- 仅修改用户端（管理后台不涉及）
- 新增 `frontend/src/components/PrintLayout.tsx`
- 修改 `frontend/src/pages/ResultPage.tsx`（引入组件、添加 `.screen-only`）
- 修改 `frontend/src/pages/ResultPage.css`（清理旧 print CSS，添加新规则）
