## Why

合盘结果页已有一套 section 级设计 token（`--section-gap-*`、`--section-padding-*`、`--subsection-gap`、`--fs-*`），三个基准 SECTION（基础盘 / 是否合 / 深度分析）已全量采用；但后期加入或升级的三个块从未对齐，导致用户感知到"间距随机、风格不统一"：

- `PersonalityFit`（性格画像，上次重排时从子块升为顶级 SECTION）仍保留旧的子块卡片皮肤（`border-left:3px` + `padding:16px`），且未接 `section-gap`/`section-padding`/`scroll-margin-top`，与上下基准段节奏脱节——而它位于叙事主线中段（SECTION 02），离群最刺眼。
- `DeepReportNarrative`（AI 深度解读）未接 section 间距 token，与其上方的命理证据贴得过近。
- `ScoreOverview` 仍在使用旧字号系统（`.compatibility-section-title` 24px），与其它已迁移到 `--fs-section-title`（22/26px）的段落标题大小不一致；旧字号类是项目内残留的第二套排版系统（漂移源）。

这是 token 漂移（drift），不是结构混乱——修复方式是收敛到既有 token，并为"卡片 vs 裸段"确立一条可解释的信息层级规则。

## What Changes

- **确立信息层级规则**：核心叙事段（公式生成、构成主线：基础盘 / 性格 / 是否合 / 深度分析）= 裸 SECTION（头部 + 内容，无卡片边框）；附属/交互层（命理证据抽屉、AI 深度解读）= 卡片，主动"框起来"以标示其可选/二级性质。
- `PersonalityFit` 去掉卡片皮肤（`compat-da-personality` 的 `border-left`/`padding:16px`），接上 `padding: 0 var(--section-padding-*)`、`margin-bottom: var(--section-gap-*)`、`scroll-margin-top`，成为基准三段的真正同级裸 SECTION。这是**可见的视觉变化**（性格块边框消失、与上下段对齐）。
- `DeepReportNarrative` 接上 section 间距 token（顶部 `section-gap`、左右 `section-padding`），**卡片外观保留**（它是可选的 AI 增强层）。
- `ScoreOverview` 标题从旧字号系统迁移到 `--fs-*` token（迁移目标 SECTION 级或子标题级，依其在 SectionVerdict 中的实际层级在 design 阶段确认）。
- **孤儿清理**：当 `ScoreOverview` 迁移后再无组件引用旧的 `.compatibility-section-header` / `.compatibility-section-title`，从 `CompatibilityResultPage.css` 删除这两条死类，永久消除双字号系统。
- **不做**：不新增 token、不改任何 DOM 结构 / 组件 props、不动评分算法 / 数据来源 / AI 生成行为；不把基准三段包成卡片。

## Capabilities

### New Capabilities
- `compatibility-result-visual-consistency`: 合盘结果页顶级模块的间距/排版一致性约束——核心叙事段统一采用 section 级 token 并呈现为裸 SECTION，附属/交互层呈现为卡片；全页仅存在一套字号 token 系统。

### Modified Capabilities
<!-- 无：本次仅为视觉收敛，不改变任何已成文的功能性需求（顺序、数据、评分、证据展示行为均不变）。 -->

## Impact

- 仅前端样式：`frontend/src/components/compatibility/deep-analysis/PersonalityFit.css`、`frontend/src/components/compatibility/deep-analysis/DeepReportNarrative.css`、`frontend/src/components/compatibility/ScoreOverview.css`、`frontend/src/pages/CompatibilityResultPage.css`（删除孤儿类）。
- 可能涉及 `PersonalityFit.tsx` 的 className 调整（移除卡片类，对齐基准 SECTION 的 className 结构）；不改组件 props 与渲染逻辑。
- 无后端、无数据迁移、无 API 变更。
- 风险：性格块由卡片改为裸段是用户可见的视觉变化（预期内）；ScoreOverview 字号迁移可能轻微改变标题尺寸（向既有 token 看齐，属目标）。
