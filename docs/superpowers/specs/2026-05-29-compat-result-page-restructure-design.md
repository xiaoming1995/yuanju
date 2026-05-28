# 合盘结果页排版重构设计

> **Status:** Draft · brainstorming output
> **Author:** Claude (Opus 4.7) + 用户
> **Date:** 2026-05-29
> **Related artifacts:** `frontend/src/pages/CompatibilityResultPage.tsx`, `frontend/src/pages/CompatibilityResultPage.css`, `frontend/src/components/CompatibilityShareCard.tsx`, `frontend/src/components/CompatibilityPrintLayout.tsx`

## 1. 背景与问题

合盘结果页（`CompatibilityResultPage`）当前 1289 行 TSX + 1457 行 CSS，纵向堆叠 11 个独立 `card` 段落（决策仪表盘 → 阅读地图 → 为什么这么判断 → 性格相处画像 → 7/30 天验证 → 关系策略 → 量化分数 → 关键判断依据 → AI 深度报告 → 专业命盘细节折叠区）。

用户反馈"排版有点乱不整洁，看起来不舒服"。三类根因诊断（用户确认主要是 ① 和 ③）：

| # | 根因 | 现状证据 |
|---|------|---------|
| ① | **信息过载 / 重复展开** | "决策仪表盘"已下结论 → 后续"为什么这么判断 / 性格画像 / 7-30 天验证 / 关系策略"又换角度讲一遍 |
| ② | **视觉拥挤 / 密度过高** | 每个 card 节奏一致（24px 标题 + 13px 描述 + 内部网格），4 层灰阶 + 金色滥用 |
| ③ | **缺主次 / 找不到主线** | 10+ 卡片视觉权重相同，"阅读地图"本身又是一个 card，未起到导航作用 |

用户给出的 mental model（"老中医"叙事节奏）：**双方基础盘 → 是否合 → AI 深度分析**。

**关键发现**：分享卡（`CompatibilityShareCard.tsx`）和 PDF（`CompatibilityPrintLayout.tsx`）**已经**按"双方双列 → 综合契合度 → verdict → 正文"排版。**只有主结果页跑偏**。本次重构是让页面对齐导出，而不是设计一种新结构。

目标：让用户进入合盘结果页时，**第一屏看到双方是谁、第二屏看到是否合、第三屏开始看 AI 解读**；任何时候 sticky 摘要给出"姓名 × 姓名 · 分数 · verdict"导航锚。

## 2. 范围

### 2.1 In scope

- `CompatibilityResultPage.tsx` 重构（11 段 → 3 段 + 1 sticky + 1 抽屉）
- `CompatibilityResultPage.css` 拆分（页面级布局 / token / 容器规则）
- 新增组件目录 `frontend/src/components/compatibility/` 及其下的 7+ 个新组件文件
- `.compatibility-result-page` 作用域新增 14 个 CSS 变量（间距 / 字号 / sticky / 容器）
- 删除 `ResultReadingMap`（被 sticky 取代）

### 2.2 Out of scope

- 后端 / 数据模型 / API（任何改动）
- `CompatibilityShareCard.tsx` / `CompatibilityShareCard.css` 渲染逻辑
- `CompatibilityPrintLayout.tsx` / `CompatibilityPrintLayout.css` 渲染逻辑
- `CompatibilityPage.tsx`（录入页）
- `CompatibilityHistoryPage.tsx`（历史页）
- 新 UI 库 / 新字体 / 新图标 / 动效（fade / slide / shimmer 等）

## 3. 设计方案

### 3.1 信息架构

11 段 → 3 段 + 1 sticky + 1 抽屉的完整映射：

| 现状段 | 去向 | 备注 |
|--------|------|------|
| ② 决策仪表盘 · verdict | **Sticky 摘要** | "姓名 × 姓名 · 分数 · verdict"，始终可见 |
| ⑪ 双方四柱（折叠底部） | **段 ① 双方基础盘** | 从最底提升到顶 |
| ⑧ 量化分数 + ④ findings + ② verdict 详情 | **段 ② 是否合** | 合并为判断核心 |
| ⑤ 性格画像 + ⑥ 7-30 天验证 + ⑦ 关系策略 + ② 下一步/避免 + ⑩ AI 长文 | **段 ③ AI 深度分析** | 5 个子段，默认全展开（`<details open>`） |
| ⑨ 关键判断依据 + ⑪ 结构化证据组 | **▼ 证据抽屉** | 默认展开（`<details open>`） |
| ③ 阅读地图 | **删除** | 被 sticky 取代 |

### 3.2 组件树

```
frontend/src/
├── pages/
│   └── CompatibilityResultPage.tsx           // 1289 → 目标 <500 行，只负责数据聚合 + 三段编排
└── components/compatibility/
    ├── CompatibilityStickyHeader.tsx         // 新：sticky 摘要
    ├── CompatibilityStickyHeader.css
    ├── SectionBasicCharts.tsx                // 新：段 ① 容器，复用 ParticipantSummaryCard
    ├── SectionBasicCharts.css
    ├── SectionVerdict.tsx                    // 新：段 ② verdict + 总分 + 维度条 + findings
    ├── SectionVerdict.css
    ├── SectionDeepAnalysis.tsx               // 新：段 ③ 容器
    ├── SectionDeepAnalysis.css
    ├── deep-analysis/
    │   ├── PersonalityFit.tsx                // 接管 PersonalityFitPanel 渲染
    │   ├── PersonalityFit.css
    │   ├── ActionPlan7d30d.tsx               // 合并 PersonalityValidationPlanPanel + StageRiskGrid + DurationTaskSummary
    │   ├── ActionPlan7d30d.css
    │   ├── RelationshipStrategy.tsx          // 接管 RelationshipStrategyPanel
    │   ├── RelationshipStrategy.css
    │   ├── NextStepsAndAvoid.tsx             // 接管 DecisionDashboardPanel 的"下一步/避免/核心矛盾"
    │   ├── NextStepsAndAvoid.css
    │   ├── DeepReportNarrative.tsx           // 接管 DeepReportPanel
    │   └── DeepReportNarrative.css
    ├── EvidenceDrawer.tsx                    // 新：底部抽屉
    ├── EvidenceDrawer.css
    ├── ParticipantSummaryCard.tsx            // 从 CompatibilityResultPage.tsx 抽出
    └── ParticipantSummaryCard.css
```

**去向规则**：

| 现有组件（inline in page） | 去向 | 改动 |
|--------------------------|------|------|
| `ParticipantSummaryCard` | 段 ① | 不变，只是移到顶部并改为 2 张并排 |
| `DecisionDashboardPanel` | **拆 2 半** | verdict → sticky；其余 props（dashboard / advice）拆到段 ② 和段 ③ |
| `ResultReadingMap` | **删除** | |
| `DecisionEvidenceSummary` | 段 ② | findings 截到 3-5 条；"查看依据" 链接保留 |
| `PersonalityFitPanel` | 段 ③ 子段 | 外层 `card` 去掉，作为子段 |
| `PersonalityValidationPlanPanel + StageRiskGrid + DurationTaskSummary` | 段 ③ 子段 | 合并为"7/30 天行动" 一个子段 |
| `RelationshipStrategyPanel` | 段 ③ 子段 | 外层 `card` 去掉 |
| `ScoreOverviewV3` / `ScoreOverview` | 段 ② | 不变，但归到"是否合"段内 |
| `EvidenceLinkedClaims` | 抽屉 | 下移；标题改"关键判断依据" |
| `DeepReportPanel` | 段 ③ 子段（最末） | "未生成"时的引导按钮保留；子段标题"AI 长文叙事" |
| `ProfessionalEvidenceGroups` | 抽屉 | 不变 |

### 3.3 数据流

无变化。所有现存 props 接口保留：

- 主页面收集 `detail` / `reading` / `consulting` / `decisionDashboard` / `personalitySummary` / `personalityValidationPlan` 等聚合数据（保持现有 `buildDecisionDashboardData` / `buildPersonalityFitSummary` / `buildPersonalityValidationPlan` 调用不变）
- 主页面把数据透传给 3 个 Section 容器组件
- Section 容器内部把数据按需透传给子组件

主页面仅承担：路由参数读取 → API 调用（不变）→ 数据归一化（不变）→ 三段编排。

### 3.4 Sticky 摘要规则

```css
.compat-sticky-header {
  position: sticky;
  top: 0;
  z-index: 50;
  height: var(--sticky-h);
  background: var(--bg-surface);
  backdrop-filter: blur(8px);
  border-bottom: 1px solid var(--border-subtle);
}
```

- **高度**：移动 48px / 桌面 56px，设 `--sticky-h`，段标题 `scroll-margin-top: var(--sticky-h)`
- **始终可见**：不做"滚动一段才出现"的花式触发
- **内容**：
  - 左：`姓名 × 姓名` + verdict（超长 ellipsis）
  - 右：`总分 / 100`（金色）
- **桌面 ≥ 1024px**：verdict 文本支持点击锚跳到段 ②（`#compat-section-verdict`），鼠标手型
- **导出例外**：`CompatibilityShareCard` 和 `CompatibilityPrintLayout` **不渲染** sticky（它们已有自己的 hero 区）
- **aria**：`aria-label="合盘摘要"`，verdict 文本是 `<a>` 时带 `aria-label="跳到判断详情"`

### 3.5 响应式栅格

| 区域 | ≤ 640px | 641-1023px | ≥ 1024px |
|------|---------|------------|----------|
| 容器 | 16px 边距 | 24px 边距 | `max-width: 900px` 居中 |
| 段 ① 双盘 | 单列纵向 | 双列 | 双列 |
| 段 ② 是否合 | 单列 | 单列 | `verdict+score` 与 `findings` 双列 |
| 段 ③ AI 深度 | 单列 | 单列 | 单列（窄文 ≤ 720px，阅读友好）|
| 段 ③ 子段内部 | 单列 | grid（如 7/30 天阶段卡）| grid |
| 段间距 | `--section-gap-mobile: 24px` | 同 mobile | `--section-gap-desktop: 40px` |
| 段内 padding | `--section-padding-mobile: 16px` | 同 mobile | `--section-padding-desktop: 24px` |

### 3.6 视觉 token 与节奏

#### 新增 14 个 CSS 变量（写入 `.compatibility-result-page` 作用域）

```css
/* 间距节奏 */
--section-gap-mobile: 24px;
--section-gap-desktop: 40px;
--section-padding-mobile: 16px;
--section-padding-desktop: 24px;
--subsection-gap: 16px;

/* 字号层级 */
--fs-section-kicker: 11px;        /* 大写 letter-spacing 2px */
--fs-section-title: 22px;          /* mobile, serif */
--fs-section-title-desktop: 26px;
--fs-subsection-title: 15px;
--fs-body: 14px;
--fs-caption: 12px;

/* sticky 与栅格 */
--sticky-h: 48px;                  /* mobile */
--sticky-h-desktop: 56px;
--container-max: 900px;
```

#### 3 层信息节奏（严格定义）

每个段统一三层结构：

```html
<section>
  <div class="kicker">SECTION 01</div>               <!-- 金色 11px 大写 letter-spacing 2px -->
  <h2 class="section-title serif">段标题</h2>         <!-- serif 22/26px -->
  <div class="subsection-card">                      <!-- 金色左 3px 边线 + bg-card -->
    <div class="subsection-title">子段标题</div>      <!-- 15px -->
    <p class="subsection-body">子段正文</p>           <!-- 14px / text-secondary -->
  </div>
</section>
```

#### 颜色克制规则

- **金色 `--wu-jin (#c9a84c)`** 只用于：sticky 强调 / 段 kicker / 子段卡左边线 / 总分数字。**禁用于**：子段 hover / icon 背景 / 按钮 / tag（用 `--text-secondary` 替代）。
- **背景灰阶降到 3 层**：`--bg-base`（页面）/ `--bg-surface`（sticky）/ `--bg-card`（子段卡）。**废弃**：`--bg-card-hover`、`--bg-elevated` 在合盘页的使用。
- **五行色** 只在段 ① 双方基础盘的"日主 + 五行 badge"出现，其他地方不染色。
- **透明白** （如 `rgba(255,255,255,0.03/0.06/0.08)`）在新页面里统一为 `--border-subtle`。

#### 不做的事

- 不重做整体配色（深色金主题保留）
- 不引入新颜色 / 新字体 / 新图标系统
- 不引入动效（fade / slide / shimmer）

### 3.7 折叠默认全展开

用户明确要求：所有可折叠区域使用 `<details open>`，**默认全展开**，保留可手动收起的能力。

适用范围：
- 段 ③ 的 5 个子段
- 底部"证据抽屉"

理由：用户来合盘是想"全看完"，不愿点。但保留 `<details>` 让证据组很长时用户可以主动折叠。

### 3.8 导出一致性

**核心原则：不改 `CompatibilityShareCard` 和 `CompatibilityPrintLayout` 的渲染逻辑。**

它们已经按"双方双列 → 综合契合度 → verdict → 正文"排版。本次新结构是让页面对齐导出，反向兼容性自然成立。

3 处对齐验证：

1. ShareCard 引用的字段（如 `decision.verdict`、`reading.overall_score`、`structured.dimensions` 等）在新结构里没改名
2. `SectionVerdict` 中"findings 3-5 条"的截取规则，和 ShareCard 已有截取规则一致
3. 导出按钮（"分享图片"、"导出 PDF"）位置：当前在容器顶部 `compat-export-actions`，新结构保留在 sticky 摘要下方，不放进 sticky 内部以避免移动端拥挤

### 3.9 测试策略

#### 单元 / smoke 测试

- **v3 reading**：保留一份 fixture，验证 `SectionVerdict` 渲染 `ScoreOverviewV3` 分支
- **legacy reading**：保留一份 fixture，验证 `SectionVerdict` 渲染 `ScoreOverview` 分支
- **缺数据**：`personalitySummary` 为 null 时，`SectionDeepAnalysis` 的 `PersonalityFit` 子段不渲染，但其他子段仍正常

#### 手动 QA checklist

1. 手机（≤ 414px）三段顺序正确，sticky 始终可见
2. 桌面（≥ 1024px）容器居中 900px，段 ① 双列、段 ② 双列布局
3. sticky 摘要点击 verdict 链接，滚动到段 ② 时段标题不被 sticky 遮挡（`scroll-margin-top` 生效）
4. 所有 `<details>` 默认展开
5. v3 用户：sticky 显示 `overall_score`，段 ② 显示 `ScoreOverviewV3`
6. legacy 用户：sticky 显示 `overall_score`（或聚合分数），段 ② 显示 `ScoreOverview`
7. 在新页面打开"分享图片"，弹出的 ShareCard 与改前像素级一致
8. 在新页面打开"导出 PDF"，PDF 输出与改前像素级一致
9. 移动端底部 BottomNav 不被遮挡，证据抽屉滚到底可见

## 4. 实施分批

### 批次 1 · 设计 token + 三段骨架（最低风险）

- 新增 14 个 CSS 变量到 `.compatibility-result-page` 作用域（保持页面局部、不污染 `:root`）
- 在 `CompatibilityResultPage.tsx` 加 3 个空容器 `SectionBasicCharts` / `SectionVerdict` / `SectionDeepAnalysis`（先什么都不渲染）
- 渲染顺序：现有 11 段保留 + 新 3 段并存（用 `import.meta.env.DEV` 或一个本地常量 `ENABLE_NEW_LAYOUT` 控制隐藏新段）

**验收**：原页面表现 0 变化；token 可被新组件引用。

### 批次 2 · 段 ① 双方基础盘 + Sticky 摘要

- 抽 `CompatibilityStickyHeader.tsx` + `SectionBasicCharts.tsx`
- 抽 `ParticipantSummaryCard.tsx`（从 `CompatibilityResultPage.tsx` 当前 839-888 行抽出）到 `components/compatibility/`
- 段 ① 在新结构中渲染 2 张并排
- 删除 `ResultReadingMap` 函数定义和调用
- 所有 `<details>` 改为 `<details open>`

**验收**：手机/桌面 sticky 行为正确；分享卡/PDF 不受影响；删除 ResultReadingMap 后页面其他段表现 0 变化。

### 批次 3 · 段 ② 是否合

- 抽 `SectionVerdict.tsx`
- 拆 `DecisionDashboardPanel`：
  - `verdict` 上移到 sticky（在主页面构造 sticky props 时取用）
  - 总分 + findings 进段 ②
  - 下一步 / 避免 / 核心矛盾 留作 `NextStepsAndAvoid` 子段（批次 4 用）
- `ScoreOverviewV3` + `DecisionEvidenceSummary` 合并到段 ②

**验收**：段 ② 渲染所有原"是否合"信息，没字段丢失；v3 + legacy 两种 reading 都正确渲染。

### 批次 4 · 段 ③ AI 深度 + 证据抽屉 + 清理

- 抽 `SectionDeepAnalysis.tsx` + 5 个子段（`PersonalityFit` / `ActionPlan7d30d` / `RelationshipStrategy` / `NextStepsAndAvoid` / `DeepReportNarrative`）
- 抽 `EvidenceDrawer.tsx`
- 删除旧 inline 函数：`PersonalityFitPanel` / `PersonalityValidationPlanPanel` / `StageRiskGrid` / `DurationTaskSummary` / `RelationshipStrategyPanel` / `DecisionDashboardPanel` / `EvidenceLinkedClaims` / `DeepReportPanel`（迁移完成后）
- 关闭 feature flag，删除旧 11 段渲染路径
- 清理 `CompatibilityResultPage.css`：每段样式拆到对应组件 `.css`，主 CSS 只保留页面级布局和 token

**验收**：`CompatibilityResultPage.tsx` < 500 行；任一新组件文件 < 250 行；CSS 拆分完成；所有手动 QA + 分享/PDF 像素级一致性回归通过。

## 5. 风险与缓解

| 风险 | 缓解 |
|------|------|
| v3 与 legacy 两套评分共存，重构时遗漏分支 | `SectionVerdict` 内部保留 `isV3` 判断；批次 3 单独写 v3 / legacy 两条用例 smoke 测试 |
| PDF 字体/分页变化（html2canvas + jsPDF） | 完全不改 `CompatibilityPrintLayout.tsx` 渲染；只验证字段引用 |
| 移动端 sticky 与 BottomNav 之间内容夹层滚动卡 | 设 `scroll-padding-top: var(--sticky-h)` + 段标题加 `scroll-margin-top: var(--sticky-h)`；锚点跳转回归测试 |
| 折叠默认展开导致首屏滚动距离过长，被误判为"还是乱" | sticky 摘要 + 段落 kicker + 段标题提供导航感；若批次 4 上线后仍反馈，再评估子段标题 sticky-secondary（不在本次范围内）|
| 已有用户的浏览器缓存 + 历史页缩略图引用 | `CompatibilityShareCard` / `CompatibilityPrintLayout` 不动，历史页不受影响 |

## 6. 验收标准（成功的判定）

1. `CompatibilityResultPage.tsx` 行数 < 500
2. 任一新 panel 文件 < 250 行
3. 任意页面 panel 用到的 CSS 不再混用 4 层灰阶
4. 手机 / 桌面 sticky 摘要始终可见
5. PDF 导出在新页面下结果与改前像素级一致
6. 分享图片导出在新页面下结果与改前像素级一致
7. v3 + legacy 两种 reading 都能正确渲染（保留两份 smoke 测试）

## 7. 不在本次范围内

- 后端 / 数据模型 / API（任何改动）
- `CompatibilityShareCard` / `CompatibilityShareCard.css` 渲染逻辑
- `CompatibilityPrintLayout` / `CompatibilityPrintLayout.css` 渲染逻辑
- `CompatibilityPage`（录入页）
- `CompatibilityHistoryPage`（历史页）
- 新 UI 库 / 新字体 / 新图标系统
- 任何动效（fade / slide / shimmer）

---

> 实施计划（拆任务）将由 `writing-plans` skill 在另一份 plan 文档中输出。
