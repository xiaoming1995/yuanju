# 导出图片/PDF 隐藏 用神/忌神 Spec

**日期：** 2026-05-17
**作者：** Claude + 用户
**状态：** 已批准，待实施

---

## 1. 背景

ResultPage 当前的两条导出链路（分享图片 / PDF）都把用户的 `用神` 和 `忌神` 五行值显式展示出来：

- **分享图片**（`ShareCard.tsx`）：底部两个 badge `喜用神：X` / `忌神：Y`
- **PDF**（`PrintLayout.tsx`）：
  - 顶部 header 同样的两个 badge
  - 末尾 "附 · 术语释义" 模块里的 `用神` 和 `忌神` 两条词条解释

用户希望对外可分享/下载的产物里**完全不出现** "用神"、"忌神" 的具体值和术语解释。

## 2. 目标

让分享图片和 PDF 都不再展示用户的用神/忌神具体值，也不再在 PDF 末尾术语表里出现这两个词条。

## 3. 非目标

- **不动 ResultPage 页面上的展示**：上方 `喜用：X / 忌：Y` 标签、`YongshenBadge` 卡、`MingpanAvatar` 等所有页面内显示**保留不变**。用户仍能在自己界面里看到这些信息。
- **不过滤 AI 解读正文文本**：`AIReport.content` 里 AI 自由叙述若提到 "用神"/"忌神"，由于属于自然语言不可机械剥离，本次不动。
- 不改后端、数据库、API 契约。

## 4. 改动清单

| 文件 | 操作 | 位置（当前行号，仅指示） |
|------|------|---------------------------|
| `frontend/src/components/ShareCard.tsx` | 删除整个 `{(yongshen \|\| jishen) && (...)}` JSX 块（喜用神/忌神 badges） | 203-235 |
| `frontend/src/components/ShareCard.tsx` | 删除 props 接口字段 `yongshen: string` `jishen: string` | 79-80 |
| `frontend/src/components/ShareCard.tsx` | 从 props 解构里移除 `yongshen, jishen` | 90 |
| `frontend/src/components/PrintLayout.tsx` | 删除 header 的 `喜用神：` / `忌神：` badges JSX 块 | 207-220 |
| `frontend/src/components/PrintLayout.tsx` | 从 "术语释义" 数组里删除 `用神` 和 `忌神` 两条 | 653-654 |
| `frontend/src/components/PrintLayout.tsx` | 删除 props 接口字段 `yongshen: string; jishen: string` | 66 |
| `frontend/src/components/PrintLayout.tsx` | 从 props 解构里移除 `yongshen, jishen` | 102 |
| `frontend/src/pages/ResultPage.tsx` | 从 `<ShareCard ... />` 调用中移除 `yongshen={...} jishen={...}` 两行 | 1179-1180 |
| `frontend/src/pages/ResultPage.tsx` | 从 `<PrintLayout ... />` 调用中移除 `yongshen={...} jishen={...}` 两行 | 1271-1272 |

`ResultPage.tsx` 其它 `yongshen` / `jishen` 用法（如本地 `const yongshen = structured.yongshen || ...`、`YongshenBadge`、`MingpanAvatar`、上方 `喜用：X` 标签、`structured.yongshen || ...为线索` 文案）**全部保留**。

## 5. 验收

1. 在 ResultPage 点击 "保存分享图" 生成图片 → 图片底部**不再有** `喜用神：X` 和 `忌神：Y` 两个 badge。
2. 在 ResultPage 点击 "导出 PDF" 生成 PDF →
   - 顶部 header **不再有** `喜用神：` / `忌 神：` 两个 badge
   - 末尾 "附 · 术语释义" 模块剩 **6 条**词条：日主、十神、调候、格局、大运、流年
3. ResultPage 网页 UI 自身**无任何视觉变化**（上方 `喜用：X 忌：Y` 标签依旧、`YongshenBadge` 结构卡依旧、`MingpanAvatar` 依旧）。
4. TypeScript 编译通过；ESLint 不报 unused-vars / unused-imports。
5. 新增的 mjs 测试 5/5 全绿。

## 6. 测试

新增 `frontend/tests/yongshen-jishen-hidden-in-export.test.mjs`，静态正则断言（依赖项目现有 `node --test` 约定）：

- ShareCard.tsx 源码不再含 `喜用神：` 或 `忌神：` 字符串
- ShareCard.tsx 的 props 接口不再含 `yongshen:` 或 `jishen:` 字段声明
- PrintLayout.tsx 源码不再含 `喜用神：` 或 `忌 神：` badge 字符串（注意 PDF header 中间有空格）
- PrintLayout.tsx 的术语数组不再含 `term: '用神'` 或 `term: '忌神'`
- PrintLayout.tsx 的 props 接口不再含 `yongshen` / `jishen` 字段

## 7. 风险与回退

- **风险低**：纯前端删除，无数据契约变更，无 DDL。
- **回退**：单 PR revert 即可。
- **可能漏点**：如果 AI 报告 prompt 后续被改为在 `structured.advice` 等结构化字段里也输出用神/忌神值，可能再次泄漏 — 但当前实现里这些字段只是文本，PrintLayout 渲染 structured 时不会专门为 yongshen/jishen 单独显示（在改完 props 后，PrintLayout 内部不再有任何 yongshen/jishen 引用）。
