# 合盘分享卡内容补全设计

**Date:** 2026-05-28
**Status:** Draft
**Author:** 刘明的MacBook (with Claude Opus 4.7)

## 1. 背景与目标

上一次合盘导出 feature 中 `CompatibilityShareCard` 已经从「C 详版 3 条证据」扩到 9 段（commit `e7abe35`）。用户继续反馈"图片展示的内容并不是详细的，需要补全"——合盘后端产出的关键数据中仍有大量信息没用上：六章节 AI 解读、关系诊断 3 条核心发现、关系策略 4 维、四维评分的"为什么"。本次设计把这 4 项一次补齐，让分享卡承担起"报告封面图"的完整角色。

## 2. 范围

### In scope
- 在 `CompatibilityShareCard.tsx` 的 9 段基础上**就近合并**新增 4 项内容：
  1. §4 四维评分进度条下方新增「正：X · 负：Y」一行
  2. §6 新增「关系诊断 · 核心 3 发现」（紧贴决策面板下）
  3. §10 新增「关系策略 4 条」（沟通 / 冲突 / 现实 / 边界）
  4. §11 新增「命理解读摆要」（六章节首句）
- 在 ShareCard Props 重新加入 `structured: CompatibilityStructuredReport | null`（上次 T1 cleanup 删过，本次确实需要用）
- 页面集成处把 `structured={structuredReport}` 传给 ShareCard

### Out of scope
- PrintLayout 的 PDF 内容变动（PDF 已经包含完整六章节，不需要再加）
- 后端 API 改动 / 新增字段（所有数据现有 `CompatibilityStructuredReport` 都有）
- 单盘 ShareCard 改动
- AI Prompt 修改（不要让 AI 额外生成 `digest` 短摘要字段）
- 分享卡尺寸切换（仍固定 400px 宽）
- 拆"精简版 vs 详细版"双模式（用户选「就近合并」单版即可）

## 3. 设计

### 3.1 新布局（13 段）

```
1.  Header                                 — 不变
2.  双盘四柱                                — 不变
3.  综合契合度大数字                        — 不变
4.  ◇ 四维 (升级)                          — 每条 bar 下方加「正：X · 负：Y」一行
5.  ◇ 关系决策                             — 不变（verdict + 5 行）
6.  ◇ 关系诊断 · 核心 3 发现 (NEW)         — 紧贴 §5
7.  ◇ 核心证据 (6 条)                      — 不变
8.  ◇ 阶段风险                             — 不变
9.  ⚠ 避免事项 (top 2)                     — 不变
10. ◇ 关系策略 (NEW)                       — 沟通 / 冲突 / 现实 / 边界 4 行
11. ◇ 命理解读摆要 (NEW)                   — 六章节首句
12. brand footer                            — 不变
```

预估高度: 当前 ~1300px → 新版 ~1700-2100px (取决于章节内容)。PNG 文件 ~500KB → ~700-900KB。

### 3.2 数据流

```
detail.reading
  ├─ score_explanations[]       → §4 ⤷ NEW 每条 bar 下方
  └─ (其它字段同前)

detail.latest_report.content_structured (新加传入 ShareCard)
  ├─ dimensions[]                → §11 NEW 摘要：取每章首句
  ├─ relationship_diagnosis
  │   └─ top_findings[]          → §6 NEW
  └─ relationship_strategy
      ├─ communication           ┐
      ├─ conflict                ├─→ §10 NEW
      ├─ reality                 │
      └─ boundary                ┘
```

### 3.3 派生逻辑

```ts
// §4 score_explanations 按维度索引
const explByDim = Object.fromEntries(
  (reading.score_explanations || []).map(e => [e.dimension, e])
)

// §6 top_findings 截前 3（每条是 { text, evidence_keys }，渲染时取 .text）
const topFindings = (structured?.relationship_diagnosis?.top_findings || []).slice(0, 3)

// §10 strategy（任一字段非空才渲染整段）
const strategy = structured?.relationship_strategy
const strategyEntries = strategy ? [
  { key: 'communication', label: '沟通', value: strategy.communication },
  { key: 'conflict',      label: '冲突', value: strategy.conflict },
  { key: 'reality',       label: '现实', value: strategy.reality },
  { key: 'boundary',      label: '边界', value: strategy.boundary },
].filter(e => e.value?.trim()) : []

// §11 六章节首句
function firstSentence(content: string): string {
  const cleaned = cleanReportText(content)
  const match = cleaned.match(/^[^。！？]+[。！？]?/)
  return (match?.[0] || cleaned).slice(0, 60).trim()
}
const dimDigests = (structured?.dimensions || []).map(d => ({
  key: d.key,
  title: d.title,
  digest: firstSentence(d.content),
}))
```

### 3.4 Props 接口变更

```ts
export interface CompatibilityShareCardProps {
  reading: CompatibilityReading
  participants: CompatibilityParticipant[]
  evidences: CompatibilityEvidence[]
  decision: DecisionDashboardData
  stageRisks: CompatibilityStageRisk[]
  structured: CompatibilityStructuredReport | null   // RE-ADD（上次 cleanup 删过）
  brand?: ExportBrand | null
}
```

页面集成：`<CompatibilityShareCard structured={structuredReport} ... />`，`structuredReport` 在 `CompatibilityResultPage.tsx:936` 已有。

### 3.5 视觉规范

| 项 | 取值 |
|---|---|
| §4「正」颜色 | `#2d5a3d` 绿 |
| §4「负」颜色 | `#8b1a1a` 红 |
| §4 因子行字号 | 10px (主体 12px 略小一档) |
| §4 因子行 margin | 进度条下方 2px 顶距 |
| §6 finding 前缀 | `· ` (与避免事项 §9 风格保持) |
| §10 strategy label | 浅褐 `#7a5c3a`，加粗，min-width 40px |
| §10 strategy value | 主色 `#4a3728` |
| §11 章节标题 | 古铜金 `#7a5c3a`，加粗，11px |
| §11 章节首句 | 主色 `#4a3728`，10px |
| §11 章节项 margin | 上下各 3px |

### 3.6 CSS 类名

新增 6 个类名（沿用 `.compat-share-*` 前缀）：

```css
.compat-share-dim-why       /* §4 ⤷ 一行容器 */
.compat-share-dim-why-pos   /* 正向 span */
.compat-share-dim-why-neg   /* 负向 span */
.compat-share-findings      /* §6 容器 */
.compat-share-finding-item  /* §6 单条 */
.compat-share-strategy      /* §10 容器 */
.compat-share-strategy-row  /* §10 单行 */
.compat-share-digests       /* §11 容器 */
.compat-share-digest-item   /* §11 单项（标题 + 首句两行） */
```

## 4. 错误处理 / 数据缺失矩阵

| 状态 | 处理 |
|---|---|
| `structured == null` | 按钮 disabled 不触发；若强渲染则 §6/§10/§11 全跳过 |
| `dimensions` 为空 | §11 不渲染 |
| `relationship_diagnosis == null` 或 `top_findings` 为空 | §6 不渲染 |
| `relationship_strategy == null` 或 4 项全空 | §10 不渲染 |
| strategy 任一项为空 | 跳过该行，渲染其他非空项 |
| `top_findings.length < 3` | 按实际数量渲染（1 或 2 也行） |
| 章节 `content` 无句号 | `firstSentence` fallback：`cleaned.slice(0, 60)` |
| `score_explanations == null` 或为空 | §4 ⤷ 行不渲染（bar 仍正常） |
| 某维度 `positive_factor` 或 `negative_factor` 一个为空 | 只渲染非空那一面，分隔符 `·` 省略 |
| 两 factor 都为空 | 该 bar 下方 ⤷ 行整体跳过 |
| v3 评分模式（zodiac/nayin/day_pillar/eight_chars） | `score_explanations.dimension` 不匹配 v3 维度键，所有 ⤷ 行跳过 |

**核心不变量**：双盘 + 综合分 + verdict + footer 在任何数据下都能渲染。

## 5. 验证标准（实施后逐条验）

### 数据呈现 (8 条)
1. §4 每条评分进度条下方有「正：X · 负：Y」一行，颜色绿/红区分
2. v3 模式下，§4 下方不出现该行
3. §6 「核心 3 发现」展示前 3 条，每条带 `·` 前缀
4. §10 「关系策略」4 行：沟通 / 冲突 / 现实 / 边界
5. §11 章节摘要展示 `structured.dimensions` 每章标题 + 首句（≤60 字）
6. §11 首句截取：第一个 。/！/？ 之前的内容
7. §4.3 矩阵任一缺失场景下 section 不出现，不崩溃
8. ShareCard 最小数据下结构稳定

### 视觉 (4 条)
9. 新增 4 段与原 9 段视觉风格统一（国风纸色、`◇`/`⚠` 前缀、12px 主字号）
10. §6 / §10 / §11 之间有视觉分隔
11. §4 ⤷ 行 10px，正绿负红
12. 卡片宽度仍精确 400px

### 导出完整性 (3 条)
13. PNG 包含所有 13 段
14. PNG 高度 ≥ 1700px
15. iOS Web Share / Android download / 桌面 toPng 三端都能完整捕获

### 回归 (3 条)
16. 原有 9 段位置和样式不变
17. 单盘 ResultPage 的 ShareCard 完全不动
18. `npm run lint && npm run build` 全绿

## 6. 文件清单

```
修改（2 个）:
  frontend/src/components/CompatibilityShareCard.tsx     (+~80 行)
  frontend/src/components/CompatibilityShareCard.css     (+~60 行)
修改（1 个）:
  frontend/src/pages/CompatibilityResultPage.tsx         (+1 行: structured 传参)
```

## 7. 关键决策摘要

| 决策 | 选项 | 取值 | 理由 |
|---|---|---|---|
| 补哪些内容 | 4 项中选 / 1-2 项 / 全选 | **全选 4 项** | 用户明确要求 |
| 排版策略 | 就近合并 / 末尾追加 / 拆双版 | **就近合并** | 信息按逻辑联系排序 |
| 六章节摘要 | 首句 / 字数截 / AI 新字段 | **首句** | 无需后端 + AI 习惯总括句在首 |
| 四维「为什么」格式 | 一行/两行/行尾同行 | **一行「正：X · 负：Y」** | 信息密度高 |
| structured prop | 加回 / 走 ref | **加回 Props** | 上次 cleanup 是误删，本次确实要用 |

## 8. 不引入的依赖
- 后端 API / Prompt 改动
- 新组件 / 抽象层
- 新的视觉模式切换
