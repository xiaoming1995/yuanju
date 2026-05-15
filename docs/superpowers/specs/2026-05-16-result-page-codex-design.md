# 缘聚 ResultPage 重设计 · 古书章卷（Codex Shell）

**日期**：2026-05-16
**范围**：仅 ResultPage（`/result`、`/history/:id`），移动端优先
**前置**：等 `bazi-ten-god-relation-matrix`、`replicate-dayun-timeline-design` 两个 in-flight change 合并完成后再启动实现

## 背景

ResultPage 是缘聚的核心 UX 表面（用户唯一会反复来读的页面）。当前实现 `frontend/src/pages/ResultPage.tsx` 已膨胀到 1200 行 + 配套 38K 单文件 CSS。审计在 390×844 viewport 上发现两类高严重度问题：

1. **密集数据网格被压扁不可读**：四柱表 10×4 在 390px 宽下文字被压到 9-11px；大运/流年 10 卡被强压成 2 列 5 行；十神矩阵 30px 衬线在窄列里发飘。
2. **超长滚动 + 缺导航**：单页要滚 6-8 屏；无章节锚点 / TOC / 回顶；H2 / H3 / `.section-title` 字号 18/20/22 混用，层级不清。

具体审计明细参见同次 brainstorm 会话记录。

## 目标

- 移动端首次进入即"古书翻阅感"，密集数据可一眼读懂
- 减少滚动疲劳，章节边界明显，能跳能回
- 标题层级和触控目标全局一致，不再有 9px 字、20px 触控
- 视觉上让用户立刻感到"和之前完全不是一个产品"

## 非目标

- 不改其他页面（HomePage / HistoryPage / PastEventsPage / Compatibility 全套 / Profile）
- 不改 AI prompt、八字算法、后端接口
- 不动 `PrintLayout` 输出格式，必须继续打印正常
- 不做跨页 design system 统一（Theme D/E/F 留作后续 change）
- 不重构 `PastEventsPage`（Theme C 留作后续 change）

## 设计决策

### 1 · 信息架构：5 章·单字

| 章 | 包含 |
|---|---|
| **命** | 四柱档案（年/月/日/时）+ 十神矩阵 + 神煞汇总 + 命格徽 |
| **性** | 五行雷达 + 用神 + 调候 + 喜忌色板 + 格局 |
| **运** | 大运十段 timeline + 当前大运的十神来源 / 起讫 / 观察 |
| **年** | 选中大运下的流年十段 + 流月 drawer 入口 + 过往年事入口 |
| **述** | AI 报告（流式 + 章节折叠）+ 词典 + 工具栏（导出/分享/打印）|

章名故意用单字。理由：手指拨 4 次走完全部、章名贴古典文献目录风格、长度天然对齐。

### 2 · 导航 + 手势 + 标题层级

**整体结构**：

```
┌─────────────────────────────┐
│  Navbar (existing, 64px)     │ 不动
├─────────────────────────────┤
│  命 ‧ 性 ‧ 运 ‧ 年 ‧ 述      │ 章节胶囊条 sticky 44px
│  ━━                          │ 当前章下划线，点击跳章
├─────────────────────────────┤
│  第一章                       │
│  命                            │ H1 24px serif + 古典分割线
│  ─────                       │
│  [章内内容垂直滚动]            │
│  ──→ 滑向《性》  ⌫           │ 章末微提示，3s 后渐隐
└─────────────────────────────┘
         ▲ BottomNav (existing, 60px)
                       ⊙ 浮动回顶（章内滚 > 1 屏出现）
```

**技术路线**：CSS `scroll-snap-type: x mandatory` 实现横滑章节切换。理由：

- 满足 CLAUDE.md「无 UI 框架」约束（不引 Swiper / Embla 等）
- 原生惯性 / 回弹 / 边界 / 加速度由浏览器接管，0 JS 处理手势
- 每章是 `flex: 0 0 100vw` 的子元素，自带 `overflow-y: auto` 处理章内滚动
- 章切换状态用 `IntersectionObserver` 监听，更新章节胶囊高亮 + URL hash

**手势规则**：

| 手势 | 行为 |
|---|---|
| 水平左右滑 | 切章（snap 强制对齐） |
| 垂直上下滑 | 章内滚动，不影响章切换 |
| 点胶囊条章名 | 跳章（`scrollTo({behavior:'smooth'})`） |
| 模态/抽屉打开 | 横滑被遮罩拦截，不切章 |
| 回顶按钮 | 仅滚动当前章到顶 |

**标题层级（全局统一）**：

```
H1（章名）      24px  serif   weight 600   + 古典分割线
H2（子节）      18px  serif   weight 600   + 1px hairline underline
H3（卡片）      15px  sans    weight 600
正文           14px  sans    line-height 1.7
辅助小字        12px  sans    color: --text-secondary
```

落地时把 `.section-title` / `h2` / `h3` 全局 map 到这套尺度，清掉 18/20/22 混用。

**深链 / 后退 / 打印**：

- URL：`/result#ming` / `#xing` / `#yun` / `#nian` / `#shu`（用拼音以保证分享出去的 URL 编码干净；章节胶囊条仍显示汉字）。当前章 ↔ hash 双向同步：进入页面读 hash 决定初始章；手势 settle 后 `history.pushState` 更新 hash
- 浏览器后退：章切用 `history.pushState` 入栈 → 后退键先在章间回退，到第一章再退出页面（书签式行为）。手势 settle 后再 push，避免快速滑动塞历史栈
- 打印：`@media print` 下移除 snap，5 章按顺序铺成长页面 + `page-break-before: always`
- 桌面 (≥1024px) **fallback：线性滚动 + 章节锚点条**（不做横滑也不做侧栏 TOC）

**章末微提示**：低调快闪式 —— 滚到当前章底部出现淡色 `──→ 滑向《下章名》`，3 秒后渐隐。

### 3 · 章内内容模式（解决密集数据网格）

抽出一个共享组件 `<SnapStrip>` 覆盖所有"多项数据"场景，保证手感统一：

```
┌─────────────────────────────┐
│  子标题       [全表/卡片] ◐  │ 章内 H2 + 视图切换
├─────────────────────────────┤
│  ┌────┐ ┌────┐ ┌────┐ →     │
│  │卡 1│ │卡 2│ │卡 3│  ...   │ 横滑、snap 单项
│  │    │ │    │ │    │       │ 左右邻居露 16px
│  └────┘ └────┘ └────┘       │
│   · · ● · · · ·             │ 当前点放大
│                              │
│  ▽ 当前项详情区               │ 跟随选中
└─────────────────────────────┘
```

实现：`overflow-x:auto + scroll-snap-type:x mandatory`，每张卡 `flex:0 0 calc(100% - 32px)`。当前项用 `IntersectionObserver` 侦测。视图切换默认卡片，可切「全表」（同样 snap 的传统横向溢出表格），切换偏好存 `localStorage`。

**五章具体形态**：

| 章 | 主体 | SnapStrip 用法 | 章内继续展开 |
|---|---|---|---|
| **命** | 四柱 + 十神矩阵 | 4 柱大卡 carousel（年/月/日/时）— 单柱大字：60px 衬线干/支、十神、藏干、神煞 | 命格徽 + 神煞汇总（标签 14px / 8px padding，44pt 触控） |
| **性** | 五行 / 用神 | **不用** SnapStrip — 五行雷达全宽展示，下接用神/调候/格局垂直堆叠卡 | 短，单屏可见 |
| **运** | 大运十段 | 10 个大运卡 SnapStrip，默认 snap 到当前大运 | 当前大运的十神来源、起讫、本运观察垂直展开 |
| **年** | 流年 | 10 个流年卡 SnapStrip（属选中大运），默认 snap 到当前流年 | 流月 drawer 入口 + 过往年事入口 |
| **述** | AI 报告 | **不用** SnapStrip — serif 古书排版 + 章节折叠 | 词典从 4 列降到 2 列 + 工具栏（导出/分享/打印） |

**关键小决定**：

- **流式 AI 排版**：从 `monospace + pre-wrap` 改成 `serif + line-height 1.85 + 段间距 1em`；光标用 `▍` 字符 600ms 闪烁；每段渲染完淡入
- **神煞 tag 触控**：9px / 2-3px padding → 12px / 6-8px padding；包成 `<button>`；可点的加微下划线提示
- **章间状态联动**：在「运」选某段大运 → 切到「年」自动显示该大运下的流年（state 走 ResultPage 顶层，不进 URL）
- **回顶按钮**：右下浮动，位于 BottomNav 之上、章末提示之下；仅滚动当前章
- **LiuYueDrawer**：宽度 `100vw → min(85vw, 360px)`，露出左 15vw 用于关闭手势 + 视觉提示

**桌面端 (≥1024px) fallback**：不应用 SnapStrip，4 柱 / 10 大运 / 10 流年直接平铺。SnapStrip 是手机适配，桌面屏宽足够。

### 4 · 反馈状态

| 状态 | 处理 |
|---|---|
| 章节首次进入加载 | 章内骨架（4 柱占位、10 块大运占位等），不再用 generic spinner |
| AI 报告流式 | serif 字体 + line-height 1.85；段落整段 fade-in；`▍` 衬线光标 600ms 闪烁 |
| 章节空数据 | 古典风格提示卡（如「此盘暂无神煞标注」），不留空白 |
| 章节请求失败 | 章内 inline 错误卡（图标 + 信息 + 重试），不让整页崩 |
| 流式中断 | 重试按钮直接重发该次流；状态恢复到中断前 |

### 5 · 文件结构落点

```
frontend/src/components/
  CodexShell.tsx            最外层翻页容器 + 胶囊条 + 章节状态 + URL hash sync
  CodexShell.css
  SnapStrip.tsx             共享横滑组件
  SnapStrip.css
  result-chapters/
    Ming.tsx                命章
    Ming.css
    Xing.tsx                性章
    Xing.css
    Yun.tsx                 运章
    Yun.css
    Nian.tsx                年章
    Nian.css
    Shu.tsx                 述章
    Shu.css

frontend/src/pages/
  ResultPage.tsx            从 1200 行降到 ~300 行（仅外层数据加载 + 路由 + CodexShell 装配）
  ResultPage.css            从 38K 降到 ~10K（仅 page shell 级样式）

frontend/tests/
  result-codex-shell.test.mjs      章节存在 / 胶囊条 / hash 同步 / 章切换
  snap-strip-behavior.test.mjs     SnapStrip 单组件横滑 / snap / 指示点 / 视图切换
  result-codex-desktop.test.mjs    ≥1024px 走线性 fallback
  mobilePageQaMatrix.mjs           扩展 ResultPage 检查项
```

## 风险 + 缓解

| 风险 | 缓解 |
|---|---|
| iOS Safari `scroll-snap` 抖动 | feature-detect 失败时降级线性滚动；加 `-webkit-overflow-scrolling: touch` |
| LiuYueDrawer 右滑同向冲突 | drawer 打开时 CodexShell 容器 `pointer-events: none`；drawer 宽度从 100vw 改 `min(85vw, 360px)` |
| 打印格式破坏 | `@media print` 下章节铺平 + 每章 `page-break-before`；PrintLayout 不动；新增打印测试 |
| 1200 行 ResultPage 拆 5 个章组件 | 先「搬」后「优化」：第一步仅搬迁现有 JSX 到对应 chapter 组件，不动样式；后续 PR 才动样式 |
| in-flight 两个 change 冲突 | 见「实现顺序」 |

## 测试 / 验收

**自动化**：
- 新增 4 个 e2e/UX 测试文件（位置见上）
- 扩展 `mobilePageQaMatrix.mjs` 加：胶囊条吸顶检测、SnapStrip 无溢出、回顶按钮存在

**人工验收**：
- 真机：iPhone 12/13、iPhone X/11、Android 360px 各跑一遍
- 触发：流式正常、流式中断 + 重试、5 章全部数据 / 部分空 / 部分失败
- 桌面：1024 / 1440 各看 fallback
- 打印：导出 PDF 不变形

## 实现顺序

1. **先合并**：`bazi-ten-god-relation-matrix` + `replicate-dayun-timeline-design` 两个 in-flight change 落地
2. **再起新 change**：`redesign-result-page-codex-shell`（OpenSpec 方式）
3. **该 change 内的 task 拆分**（具体细节交给后续 writing-plans 阶段）：
   - 构建 `CodexShell` + `SnapStrip` 基础组件 + 测试
   - 拆 ResultPage 渲染到 5 个 chapter 组件（不动样式，纯搬迁）
   - 章 1 / 2 / 3 / 4 / 5 各自的内部 SnapStrip / 视图切换 / 触控 / 视觉
   - 标题层级全局收口
   - 流式排版 + 反馈状态 + 章节骨架
   - 桌面 fallback
   - 打印验证
   - 真机 QA

## 验收即"完成"的标准

- 视觉：在 390×844 真机上首屏不需要捏放就能读所有四柱信息；切章像翻书
- 功能：5 章可横滑、可点跳、可深链；流式 AI 排版古雅、可重试；打印不破
- 代码：ResultPage.tsx ≤ 400 行；全部新组件单独可测；既有 12 个 UX 测试不退化

## 不在本次范围

- HomePage / HistoryPage / PastEventsPage / Compatibility 三页 / Profile 的 UX 改动
- 跨页 design system 统一（断点 / hover 守卫 / 加载骨架库）
- 触控目标 / 神煞 tag 之外的可点元素的 a11y 复审
- 桌面侧栏 TOC 实现
- 视觉风格大改（配色 / 字体族 / 间距 token） —— 沿用现有 CSS 变量
