## Context

合盘结果页在 `CompatibilityResultPage.css` 上定义了一套 section 级 token（`--section-gap-mobile/desktop`、`--section-padding-mobile/desktop`、`--subsection-gap`、`--fs-section-kicker/title/title-desktop/subsection-title/body/caption`）。三个基准 SECTION（`SectionBasicCharts`/`SectionVerdict`/`SectionDeepAnalysis`）与 `EvidenceDrawer` 已全量采用；三个块未对齐：

- `PersonalityFit`：上次叙事重排把它从"深度分析"内的子块升级为顶级 SECTION 02，但 CSS 仍是旧子块卡片（`.compat-da-personality { padding:16px; border-left:3px solid var(--wu-jin); }`），未接 `section-gap`/`section-padding`/`scroll-margin-top`。它已经有标准的 `compat-section-kicker`(SECTION 02) + `compat-section-title` 头部。
- `DeepReportNarrative`：`.compat-da-report` 卡片（同样 `border-left:3px`），未接 section 间距 token。
- `ScoreOverview.tsx` 实际导出两个组件：`ScoreOverviewV3`（V3 路径，`compat-score-v3__*`，无旧排版）与遗留 `ScoreOverview`（`关系速览` 速览块，仅 legacy 评分路径渲染）。遗留块用 `.compatibility-section-header--stacked` + `<h2 class="compatibility-section-title">关系速览</h2>`(24px) + `.compatibility-section-desc` —— 这是项目内残留的第二套字号系统。

关键观察：`关系速览` 在结构上是 SECTION 03（是否合）**内部的子标题**，但渲染为 24px `<h2>`，比其父 SECTION 标题（`--fs-section-title` 22/26px）还大——层级倒挂。

## Goals / Non-Goals

**Goals:**
- 用一条可解释的规则消除"卡片 vs 裸段"的随意性。
- 让三个离群块收敛到既有 token，无新增 token。
- 消除双字号系统：迁移后删除孤儿旧类。
- 保持功能/数据/DOM 结构不变，改动可被肉眼验证。

**Non-Goals:**
- 不新增任何 CSS 变量 / token。
- 不改组件 props、渲染逻辑、DOM 树结构（仅允许 `PersonalityFit` 顶层 className 由卡片类换为 SECTION 类）。
- 不把基准三段包成卡片。
- 不动评分算法、数据来源、AI 生成行为、模块顺序。
- 不重构 `ScoreOverviewV3`（它已合规）。

## Decisions

### D1：卡片 vs 裸段按"内容性质"分层，而非按观感
- **裸 SECTION（头部 + 内容，无边框）= 核心叙事段**：基础盘 / 性格 / 是否合 / 深度分析。它们是公式生成、构成阅读主线。
- **卡片（主动框起）= 附属/交互层**：命理证据（抽屉）、AI 深度解读（可选生成、有按钮/loading）。卡片在此承担"这是可选的二级内容"的语义信号。
- *备选*：(a) 全部裸段——但 AI 解读是可选交互块，失去卡片就和正文混同；(b) 全部卡片——基准三段变重、风格突变且改动大。均不取。
- *与既有决定一致*：性格是公式生成（非 AI），归入主线裸段；SECTION 04 标题此前已从"AI 深度分析"改为"深度分析"，真正的 AI 仅在最后那张卡片里——卡片强化该语义。

### D2：PersonalityFit 收敛为基准 SECTION 同级
- 移除 `.compat-da-personality` 的卡片皮肤（`padding:16px`、`border-left`、`background`）。
- 顶层接 `padding: 0 var(--section-padding-mobile)`（desktop 对应）、`margin-bottom: var(--section-gap-*)`、`scroll-margin-top: var(--sticky-h*)`，对齐 `SectionBasicCharts`/`SectionVerdict`/`SectionDeepAnalysis` 的外层规则。
- 头部已是 `compat-section-kicker` + `compat-section-title`，沿用；内部画像内容（A/B 卡、维度行、差异对照）作为 section 内容保留，仅在必要时把外层卡片 padding 改为 section padding。

### D3：DeepReportNarrative 保留卡片 + 接 section 间距
- 卡片皮肤（`border-left:3px`、`background`、内 padding）**保留**。
- 顶层增加 `margin-top: var(--section-gap-*)`、左右 `margin: 0 var(--section-padding-*)`（对齐 `EvidenceDrawer` 的卡片定位方式：用 margin 留出页面边距而非 padding 顶满），使它与上方命理证据之间有正常 section 节奏。

### D4：ScoreOverview（遗留速览块）迁到子标题字号
- `关系速览` 是 SECTION 内子标题 → 迁移目标为 `--fs-subsection-title`（15px），而非 section title；同时修正"子标题大于父标题"的倒挂。
- 标题继续用 `serif` + `<h2>`（保留语义层级），仅字号经由新类/内联引用 `--fs-subsection-title`；`.compatibility-section-desc` 文案迁到 `--fs-caption` 或保留其现值由实现时对齐周边子说明。
- *备选*：迁到 `--fs-section-title`——否决，会把子块抬到与 section 同级，重蹈倒挂。

### D5：孤儿清理
- D4 迁移后，全仓确认再无组件引用 `.compatibility-section-header`、`.compatibility-section-header--stacked`、`.compatibility-section-title`、`.compatibility-section-desc`，则从 `CompatibilityResultPage.css` 删除（CLAUDE.md §3：清理由本次改动产生的孤儿）。
- *前置校验*：删除前对全 `frontend/` 做 grep，确认零引用；若仍有外部引用则不删，仅在本次涉及处迁移。

## Risks / Trade-offs

- [性格块由卡片变裸段是用户可见变化] → 这是目标本身；在交付说明里明确告知，避免被当作回归。
- [遗留 ScoreOverview 仅 legacy 路径渲染，当前 V3 数据下难以肉眼复现] → 验证时构造/选取一条 legacy 评分 reading，或退而求其次做静态 CSS/源码断言确认旧类已无引用、新字号已生效。
- [删错孤儿类导致其它页面失样] → D5 强制全仓 grep 前置校验，零引用才删。
- [`compatibility-section-header--stacked` 等可能被非 compat 页面复用] → 同上，grep 范围为整个 `frontend/src`。

## Open Questions

- 无阻塞性未决项。`关系速览` 子说明（`.compatibility-section-desc`）迁到 `--fs-caption` 还是保留 13px，留待实现时对齐相邻子块说明文字，不影响范围。
