## 1. 现状确认与基线

- [x] 1.1 全 `frontend/src` grep `.compatibility-section-header`、`.compatibility-section-header--stacked`、`.compatibility-section-title`、`.compatibility-section-desc`，列出所有引用点，确认仅遗留 `ScoreOverview`（关系速览块）使用——若有 compat 之外的引用，记录并据 D5 收窄删除范围。
- [x] 1.2 跑一遍既有前端测试，记录基线（pass/fail 数与已知预存失败项），便于改后对比。

## 2. PersonalityFit 收敛为裸 SECTION（D2）

- [x] 2.1 在 `PersonalityFit.css` 移除 `.compat-da-personality` 的卡片皮肤（`border-left`、卡片 `background`、`padding:16px`）。
- [x] 2.2 顶层接入基准 SECTION 外层规则：`padding: 0 var(--section-padding-mobile)`、`margin-bottom: var(--section-gap-mobile)`、`scroll-margin-top: var(--sticky-h)`，并在 desktop 媒体查询对应 `--section-padding-desktop`/`--section-gap-desktop`/`--sticky-h-desktop`（对齐 `SectionBasicCharts.css`）。
- [x] 2.3 如顶层 className 仍是卡片类，调整 `PersonalityFit.tsx` 顶层 className 使其与基准 SECTION 结构一致（仅 className，不改 DOM 树/props/渲染逻辑）。
- [x] 2.4 校验内部画像内容（A/B 卡、维度行、差异对照）间距仍合理，仅清理因去卡片 padding 产生的边距塌缩。

## 3. DeepReportNarrative 接入 section 间距（D3，保留卡片）

- [x] 3.1 在 `DeepReportNarrative.css` 顶层增加 `margin-top: var(--section-gap-mobile)` 与左右 `margin: 0 var(--section-padding-mobile)`（desktop 媒体查询对应 desktop token），与 `EvidenceDrawer.css` 的卡片定位方式一致；卡片皮肤（`border-left`/`background`/内 padding）保留不动。

## 4. ScoreOverview 字号收敛（D4）

- [x] 4.1 在遗留 `ScoreOverview`（关系速览块）将 `关系速览` 标题字号迁到 `--fs-subsection-title`（保留 `serif` 与 `<h2>` 语义层级，仅字号）；子说明对齐相邻子块说明（`--fs-caption` 或保留现值，依实现时周边一致性定）。
- [x] 4.2 移除该块对 `.compatibility-section-header(--stacked)` / `.compatibility-section-title` / `.compatibility-section-desc` 的依赖，改用本次定义的子标题样式或既有子块样式。

## 5. 孤儿清理（D5）

- [x] 5.1 重跑 1.1 的 grep，确认 `.compatibility-section-header`、`--stacked`、`.compatibility-section-title`、`.compatibility-section-desc` 在 `frontend/src` 已零引用。
- [x] 5.2 零引用确认后，从 `CompatibilityResultPage.css` 删除这四条孤儿类；若仍有 compat 之外引用则保留并在交付说明里指出。

## 5b. verdict 摘要条对齐内容列（实现中发现的同类离群者）

- [x] 5b.1 `CompatibilityStickyHeader.css`：将横向 `padding: 0 16px`/`0 24px` 改为 `margin: 0 var(--section-padding-mobile/desktop)`（面板缩进到内容列），文字位置不变仍与 SECTION 内容左对齐，消除面板左右探出；`margin-bottom` 并入 margin 简写。
- [x] 5b.2 `CompatibilityResultPage.css` 的 `.compat-export-actions`：增加 `padding: 0 var(--section-padding-mobile)` + desktop `var(--section-padding-desktop)`，使右对齐导出按钮的右边缘与内容列对齐（box-sizing 已全局 border-box，移动端 `width:100%` 不溢出）。

## 6. 验证

- [x] 6.1 `tsc -b` 与 `eslint`（受影响文件）无错误/无孤儿引用。
- [x] 6.2 更新/新增静态断言：`PersonalityFit.css` 不再含 `border-left`/卡片 `padding`、已含 `--section-gap`/`--section-padding`/`scroll-margin-top`；`DeepReportNarrative.css` 已含 section margin token；`ScoreOverview` 已无旧字号类、`关系速览` 用 `--fs-subsection-title`；page CSS 已无孤儿类。修正任何断言旧样式的既有测试（仅改受本次影响的断言）。
- [x] 6.3 真实渲染一条 reading 肉眼确认：01–04 四段间距节奏一致、性格块不再像卡片且与上下段对齐、AI 深度解读与命理证据之间有正常 section-gap、无标题层级倒挂。legacy 速览块若 V3 数据下不渲染，构造/选取一条 legacy 评分 reading 或以源码/CSS 断言替代。
- [x] 6.4 全量前端测试对比 6.1 基线，无新增失败。
