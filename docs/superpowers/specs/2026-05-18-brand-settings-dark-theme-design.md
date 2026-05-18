# Brand Settings 深色主题对齐 — Design Spec

**Date:** 2026-05-18
**Status:** Approved (pending user review of this written spec)
**Scope:** UI 视觉一致性（仅样式层）

---

## 1. 背景与问题

`/settings/brand` 设置页（导出品牌定制化模块）于 2026-05-18 落地。
当时 `BrandSettingsPage.css` 与 `LogoCropModal.css` 沿用了 PDF / PNG 导出产物的浅色奶油 + 米黄色板（`#fdf9f2 / #e0cca0 / #2a1a0a / #fff`），因为页面同时承载一个"导出预览样张" `BrandPreviewCard`，作者顺手把 UI 外壳也涂成了同样的浅色。

但全站主题是**深色 + 金色点缀**（`--bg-base: #0d0f14 / --bg-card: #1a1f2e / --text-primary: #e8e4d8 / --text-accent: #c9a84c`），其他页面（ProfilePage、HistoryPage、ResultPage 屏显部分、CompatibilityPage）均按此 token。

结果：
- `/settings/brand` 整页是"夹在黑色页面里的奶油浮岛"，视觉割裂
- 弹出的 `LogoCropModal` 也是奶油底，二次割裂
- 与个人中心、历史、合盘等所有相邻页面色调矛盾，用户感知到"这是另一个产品"

## 2. 目标

把 Brand Settings 模块的**外壳**全部并入主题深色系，**仅保留**中央 `BrandPreviewCard` 的浅色 — 因为它是导出实物（PNG/PDF）的所见即所得样张，需要保留浅色调以忠实展示。

非目标：
- 不改任何组件结构、不动 TSX
- 不动 `BrandPreviewCard` 的视觉
- 不重新设计交互流程
- 不动 ShareCard / PrintLayout（这两个本来就是导出产物，已是浅色）

## 3. 文件改动清单

| 文件 | 改动 |
|------|------|
| `frontend/src/pages/BrandSettingsPage.css` | 全面用全局 token 重写所有局部颜色 |
| `frontend/src/components/LogoCropModal.css` | 全面用全局 token 重写；移除局部按钮样式（按钮交给全局 `.btn`） |
| `frontend/src/pages/BrandSettingsPage.tsx` | 仅在 logo 上传/删除按钮上加 `className="btn btn-ghost btn-sm"`（其它结构与 class 名不变） |
| `frontend/src/components/LogoCropModal.tsx` | 仅把取消按钮换成 `className="btn btn-ghost"`、确认按钮换成 `className="btn btn-primary"`（其它结构不变） |
| `frontend/src/components/BrandPreviewCard.tsx` | 不改（故意保留浅色） |

## 4. Token 映射规则

页面级（`BrandSettingsPage.css`）：

| 类别 | 旧值（硬编码） | 新值（token） |
|------|---------------|---------------|
| section 卡片底色 | `#fdf9f2` | `var(--bg-card)` |
| 卡片 border | `#e0cca0` | `var(--border-default)` |
| 卡片圆角 | `8px` | `var(--radius-md)` |
| h1 / h2 主文字 | `#2a1a0a` | `var(--text-primary)` |
| 描述文字 | `#5a3a1a` | `var(--text-secondary)` |
| 占位 / 计数 / 小字 | `#999` / `#aaa` | `var(--text-muted)` |
| 输入框底色 | `#fff` | `var(--bg-elevated)` |
| 输入框 disabled 底色 | `#f5efe3` | `var(--bg-card)` |
| 输入框字色 | （继承） | `var(--text-primary)` |
| 输入框 disabled 字色 | `#999` | `var(--text-muted)` |
| 输入框 border | `#e0cca0` | `var(--border-default)` |
| 输入框 focus border | （无） | `var(--border-accent)` |
| 返回按钮 border | `#e0cca0` | `var(--border-default)` |
| 返回按钮 字 | `#5a3a1a` | `var(--text-secondary)` |
| 单选项 label 字 | `#2a1a0a` | `var(--text-primary)` |
| 错误条 bg / border / text | `#fdf0f0` / `#c0392b` / `#c0392b` | `rgba(192,57,43,0.12)` / `rgba(192,57,43,0.4)` / `#e87171` |
| 成功条 bg / border / text | `#ecfbef` / `#6cbf7a` / `#2d6b3a` | `rgba(108,191,122,0.12)` / `rgba(108,191,122,0.4)` / `#7dd87f` |
| 未保存条 bg / border / text | `#fffbe6` / `#d4b896` / `#7a5c2e` | `rgba(212,184,150,0.10)` / `var(--border-accent)` / `var(--text-accent)` |
| Logo 预览框 border（虚线） | `#e0cca0` 1px dashed | `var(--border-default)` 1px dashed |
| Logo 预览框 bg | `#fff` | `var(--bg-elevated)` |
| Logo 上传/删除按钮 | 局部 `bg:#fdf9f2 border:#e0cca0` | **删除局部样式**，按钮元素改用 `className="btn btn-ghost btn-sm"`（TSX 配合一处微改） |
| 底部 tip / 底部链接 | `#aaa` / `var(--text-muted)` | `var(--text-muted)`（保持） |

弹窗级（`LogoCropModal.css`）：

| 类别 | 旧值 | 新值 |
|------|------|------|
| Overlay | `rgba(20,12,6,0.72)`（暖色蒙层） | `rgba(0,0,0,0.6)`（中性蒙层） |
| Modal 面板 bg | `#fdf9f2` | `var(--bg-card)` |
| Modal border | `#e0cca0` | `var(--border-default)` |
| Modal box-shadow | `0 12px 40px rgba(0,0,0,0.3)` | 保留 |
| Modal 标题 | `#2a1a0a` | `var(--text-primary)` |
| canvas 容器 bg | `#1a1208`（深棕） | `var(--bg-base)`（与全站统一） |
| 缩放滑块 label 字 | `#5a3a1a` | `var(--text-secondary)` |
| 动图提示文字 | `#999` | `var(--text-muted)` |
| 取消按钮 | 自定 ghost 样式 | **删除局部样式**，改用 `className="btn btn-ghost"` |
| 确认按钮 | 自定金色渐变 | **删除局部样式**，改用 `className="btn btn-primary"` |

## 5. 行为不变量

- 所有现有 class 名保持不变；TSX 仅在 4 个按钮上**追加**全局 `.btn` class，不删除原有 class，不重排 JSX
- LogoCropModal 的 Cropper 第三方库交互不动
- 上传 / 裁剪 / 保存 / 删除 / 重置 / 成功提示自动消退（2.5s）等全部行为不变
- 不影响 `BrandPreviewCard` 内容（依然奶油白 + 暖色文字）
- 不影响 ShareCard 与 PrintLayout（导出产物）的渲染

## 6. 测试策略

**已有的静态正则测试** `frontend/tests/brand-settings.test.mjs` 只断言文件存在 + 行为路径，与样式无关，因此不会因样式改动而失败 — 不需修改。

**自动化检查：**
1. `cd frontend && npm run lint` — 干净
2. `cd frontend && npm run build` — TypeScript 通过
3. `cd frontend && node --test tests/brand-settings.test.mjs` — 通过

**手动验收**（无法被自动化覆盖的纯视觉）：
1. `/settings/brand` 整页背景与卡片均为深色，与 `/profile` 视觉一致
2. 顶部"返回 / 导出品牌设置"标题区与 BottomNav / Navbar 色调连续
3. 输入框 focus 时金色高亮
4. 三个状态条（错误 / 成功 / 未保存）在深色下都可读
5. Logo 上传按钮 hover 与全局其他 ghost 按钮一致
6. 点击"上传" → 裁剪 Modal 是深色 + 金色"确认"按钮
7. 中央 BrandPreviewCard 保留奶油色样张（不变）
8. 保存成功的绿色提示在深色下不刺眼
9. 重置默认按钮、保存按钮、删除按钮均使用全局 `.btn` 风格

## 7. 风险与回滚

**风险：**
- TSX 文件中两处微小改动（`brand-logo-actions button` → 加 `className="btn btn-ghost btn-sm"`，Modal 内两个按钮换 className）— 这是仅有的非纯 CSS 改动，需小心。
- 全局 `.btn` 在不同尺寸下的视觉重量可能与原局部按钮有差，需要在手动验收时确认。

**回滚：** 这是纯样式 PR，git revert 单个 commit 即可。

## 8. 不在本次范围

- 不重新设计 BrandPreviewCard
- 不调整 layout / 不重排 section
- 不新增功能、不增加新字段
- 不改 backend、不改数据库、不改 API
- 不动其他页面的样式
