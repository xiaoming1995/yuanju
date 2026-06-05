# 夫妻宫匹配 · 前端渲染 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在合盘报告「双方性格画像与差异」下方新增一块「夫妻宫匹配」卡片，把后端已产出的 `spouse_palace_match` 渲染出来（左右两列 + 高/中/低契合度徽章 + 一句话总括）。

**Architecture:** 纯前端只读展示。加 TS 类型 → 新建 `SpousePalaceMatch.tsx/.css`（复用现有 section/grid/card 样式与 duration 的高/中/低配色）→ 在 `DeepReportNarrative` 渲染并把双方展示名从页面透传下来。不动后端、不动其他区块。

**Tech Stack:** React + TypeScript + Vite，纯 CSS（CSS 变量，无 Tailwind/CSS-modules）。所有命令在 `frontend/` 目录执行。

**测试说明（重要）：** 前端**没有组件测试框架**（package.json 无 vitest/jest，无任何 `*.test.tsx`）。按 spec §6 的 fallback，本计划每步验证 = **类型检查 + 构建 + lint**（`npm run build` 内含 `tsc -b`；`npm run lint`），最后一步做**人工可视化验证**。不引入测试框架（YAGNI，超范围）。

**部署提醒：** 这是 docker 跑的前端容器（`yuanju-frontend`，:3000/或 vite :5200）。改完代码后该容器若不是挂载源码热更新，需要**重建前端镜像/重启前端容器**才能看到效果。

---

## File Structure

| 文件 | 职责 | 动作 |
|---|---|---|
| `frontend/src/lib/api.ts` | 加 `CompatibilitySpousePalaceSide` / `CompatibilitySpousePalaceMatch` 类型 + `CompatibilityStructuredReport.spouse_palace_match` 可选字段 | Modify |
| `frontend/src/components/compatibility/deep-analysis/SpousePalaceMatch.tsx` | 新组件：左右两列 side 卡片 + 徽章 + summary，含空值兜底 | Create |
| `frontend/src/components/compatibility/deep-analysis/SpousePalaceMatch.css` | 卡片/网格/徽章样式（复用现有变量与配色） | Create |
| `frontend/src/components/compatibility/deep-analysis/DeepReportNarrative.tsx` | 加 `selfName`/`partnerName` props，在性格画像后渲染新组件 | Modify |
| `frontend/src/pages/CompatibilityResultPage.tsx` | 给 `DeepReportNarrative` 传 `selfName`/`partnerName` | Modify |

---

## Task 1: 加 TypeScript 类型

**Files:**
- Modify: `frontend/src/lib/api.ts`（`CompatibilityStructuredReport` 接口约 451–465；新接口加在它附近的其它 `Compatibility*` 接口旁）

**背景：** 前端 TS 类型用 snake_case 镜像后端 JSON，可选字段用 `?:`。后端 `spouse_palace_match` 形状：`self`/`partner` 各含 `ideal_portrait`(string)、`match_level`('high'|'medium'|'low'，缺性别时可能为空串)、`fit_points`/`gap_points`/`evidence_keys`(string[])，外加顶层 `summary`(string)。

- [ ] **Step 1: 加两个接口**

在 `frontend/src/lib/api.ts` 中，`CompatibilityStructuredReport` 接口**之前**（或紧邻它的 `Compatibility*` 接口群里），加入：

```ts
export interface CompatibilitySpousePalaceSide {
  ideal_portrait: string
  match_level: 'high' | 'medium' | 'low' | ''
  fit_points: string[]
  gap_points: string[]
  evidence_keys: string[]
}

export interface CompatibilitySpousePalaceMatch {
  self: CompatibilitySpousePalaceSide
  partner: CompatibilitySpousePalaceSide
  summary: string
}
```

- [ ] **Step 2: 给报告接口加字段**

在 `CompatibilityStructuredReport` 接口里，`personality_comparison?: ... | null` 那一行**之后**，加：

```ts
  spouse_palace_match?: CompatibilitySpousePalaceMatch | null
```

- [ ] **Step 3: 类型检查通过**

Run: `npm run build`
Expected: 构建成功（`tsc -b` 无类型错误）。若只想快速查类型，可跑 `npx tsc -b --noEmit` 也应无报错。

- [ ] **Step 4: 提交**

```bash
git add frontend/src/lib/api.ts
git commit -m "feat(compat-ui): add spouse_palace_match types"
```

---

## Task 2: 新建 SpousePalaceMatch 组件 + 样式

**Files:**
- Create: `frontend/src/components/compatibility/deep-analysis/SpousePalaceMatch.tsx`
- Create: `frontend/src/components/compatibility/deep-analysis/SpousePalaceMatch.css`

**背景：** 风格对齐同目录 `PersonalityComparison.tsx`（外层 `compatibility-report-section`、标题 `serif compatibility-report-title`、两列 grid）。`fit_points`/`gap_points` 是**纯字符串数组**（不是 `{title,detail}`），所以不能复用 `PersonalityComparison`。徽章高/中/低配色复用 `ActionPlan7d30d.css` 的绿/金/红（high=`--wu-mu`、medium=`--wu-jin`、low=`--wu-huo`）。空值兜底见 spec §5：`match` 空→返回 null；某侧 `match_level` 空串→不显徽章；fit/gap 空数组→不渲染该小标题。

- [ ] **Step 1: 写组件**

创建 `frontend/src/components/compatibility/deep-analysis/SpousePalaceMatch.tsx`：

```tsx
import './SpousePalaceMatch.css'
import type {
  CompatibilitySpousePalaceMatch,
  CompatibilitySpousePalaceSide,
} from '../../../lib/api'

// 契合度 label 由前端按 level 固定映射（防 LLM 标签漂移）
const MATCH_LEVEL_TEXT: Record<string, string> = {
  high: '高',
  medium: '中',
  low: '低',
}

function SideCard({ name, side }: { name: string; side?: CompatibilitySpousePalaceSide }) {
  if (!side) return null
  const fit = Array.isArray(side.fit_points) ? side.fit_points.filter(Boolean) : []
  const gap = Array.isArray(side.gap_points) ? side.gap_points.filter(Boolean) : []
  const levelText = MATCH_LEVEL_TEXT[side.match_level]
  return (
    <div className="spouse-match-card">
      <div className="spouse-match-card-head">
        <span className="spouse-match-card-title">{name}理想的另一半</span>
        {levelText && (
          <span className={`spouse-match-badge spouse-match-badge--${side.match_level}`}>契合 {levelText}</span>
        )}
      </div>
      {side.ideal_portrait && <p className="spouse-match-portrait">{side.ideal_portrait}</p>}
      {fit.length > 0 && (
        <div className="spouse-match-list">
          <div className="spouse-match-list-title">对上了</div>
          {fit.map((text, index) => (
            <p key={`fit-${index}`} className="spouse-match-point">✓ {text}</p>
          ))}
        </div>
      )}
      {gap.length > 0 && (
        <div className="spouse-match-list">
          <div className="spouse-match-list-title">有差距</div>
          {gap.map((text, index) => (
            <p key={`gap-${index}`} className="spouse-match-point">✗ {text}</p>
          ))}
        </div>
      )}
    </div>
  )
}

export default function SpousePalaceMatch({
  match,
  selfName,
  partnerName,
}: {
  match?: CompatibilitySpousePalaceMatch | null
  selfName: string
  partnerName: string
}) {
  if (!match || (!match.self && !match.partner)) return null
  return (
    <div className="compatibility-report-section spouse-match">
      <div className="serif compatibility-report-title">夫妻宫匹配</div>
      <div className="spouse-match-grid">
        <SideCard name={selfName} side={match.self} />
        <SideCard name={partnerName} side={match.partner} />
      </div>
      {match.summary && <p className="spouse-match-summary">{match.summary}</p>}
    </div>
  )
}
```

- [ ] **Step 2: 写样式**

创建 `frontend/src/components/compatibility/deep-analysis/SpousePalaceMatch.css`：

```css
.spouse-match .spouse-match-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
  margin-bottom: 14px;
}

.spouse-match-card {
  min-width: 0;
  padding: 14px;
  border-radius: 8px;
  border: 1px solid var(--border-subtle);
  background: rgba(255, 255, 255, 0.04);
}

.spouse-match-card-head {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 10px;
}

.spouse-match-card-title {
  color: var(--wu-jin);
  font-size: 14px;
  line-height: 1.5;
}

.spouse-match-badge {
  margin-left: auto;
  padding: 2px 10px;
  border-radius: 999px;
  font-size: 12px;
  white-space: nowrap;
}

.spouse-match-badge--high {
  background: rgba(76, 175, 125, 0.12);
  color: var(--wu-mu);
}

.spouse-match-badge--medium {
  background: rgba(201, 168, 76, 0.12);
  color: var(--wu-jin);
}

.spouse-match-badge--low {
  background: rgba(224, 92, 75, 0.12);
  color: var(--wu-huo);
}

.spouse-match-portrait {
  margin: 0 0 10px;
  color: var(--text-secondary);
  font-size: 13px;
  line-height: 1.65;
}

.spouse-match-list + .spouse-match-list {
  margin-top: 10px;
}

.spouse-match-list-title {
  color: var(--wu-jin);
  font-size: 13px;
  margin-bottom: 6px;
}

.spouse-match-point {
  margin: 4px 0 0;
  color: var(--text-secondary);
  font-size: 13px;
  line-height: 1.6;
}

.spouse-match-summary {
  margin: 4px 0 0;
  color: var(--text-primary);
  font-size: 13px;
  line-height: 1.65;
}

@media (max-width: 768px) {
  .spouse-match .spouse-match-grid {
    grid-template-columns: 1fr;
  }
}
```

- [ ] **Step 3: 类型检查 + lint 通过**

Run: `npm run build && npm run lint`
Expected: 构建成功、无类型错误；eslint 对新文件无错误（组件此时尚未被引用，但应能独立编译——若 `tsc -b` 因「未使用」报错，不会，因为 export default 被视为使用；若 lint 报 unused，忽略仅在它被引用后消失的情形，正常不会）。

- [ ] **Step 4: 提交**

```bash
git add frontend/src/components/compatibility/deep-analysis/SpousePalaceMatch.tsx frontend/src/components/compatibility/deep-analysis/SpousePalaceMatch.css
git commit -m "feat(compat-ui): add SpousePalaceMatch card component"
```

---

## Task 3: 挂载到报告 + 透传双方展示名

**Files:**
- Modify: `frontend/src/components/compatibility/deep-analysis/DeepReportNarrative.tsx`
- Modify: `frontend/src/pages/CompatibilityResultPage.tsx`（`DeepReportNarrative` 调用约 353–362）

**背景：** LLM 生成的 `ideal_portrait` 文案用双方展示名（如「A的配偶星…」），所以列头也要用同一展示名，否则对不上。`DeepReportNarrative` 当前 Props 不含名字，需新增 `selfName`/`partnerName` 两个 prop，由页面传入。页面里 `selfP`/`partnerP` 已在作用域（`CompatibilityResultPage.tsx:181-182`），且同文件 line 326-327 已有 `selfP?.display_name || '我'` / `partnerP?.display_name || '对方'` 的现成写法，照搬即可。

- [ ] **Step 1: 改 DeepReportNarrative —— import + Props + 渲染**

在 `frontend/src/components/compatibility/deep-analysis/DeepReportNarrative.tsx`：

(a) 顶部 import 区，`import PersonalityComparison from './PersonalityComparison'` 之后加：

```tsx
import SpousePalaceMatch from './SpousePalaceMatch'
```

(b) `type Props = { ... }` 里，`onGenerateReport: () => void` 之后加两行：

```tsx
  selfName: string
  partnerName: string
```

(c) 函数签名解构 `export default function DeepReportNarrative({ ... onGenerateReport, }: Props)` 里，把 `onGenerateReport,` 之后补上：

```tsx
  selfName,
  partnerName,
```

(d) JSX 中，`<PersonalityComparison comparison={structuredReport.personality_comparison} />` 这一行**之后**插入：

```tsx
            <SpousePalaceMatch
              match={structuredReport.spouse_palace_match}
              selfName={selfName}
              partnerName={partnerName}
            />
```

- [ ] **Step 2: 改页面 —— 传名字**

在 `frontend/src/pages/CompatibilityResultPage.tsx` 的 `<DeepReportNarrative ... />`（约 353–362）里，`onGenerateReport={handleGenerateReport}` 之后加两行：

```tsx
          selfName={selfP?.display_name || '我'}
          partnerName={partnerP?.display_name || '对方'}
```

- [ ] **Step 3: 构建 + lint 通过**

Run: `npm run build && npm run lint`
Expected: 构建成功、无类型错误、lint 无新错误。（若漏传 `selfName`/`partnerName`，`tsc` 会因缺必填 prop 报错——这就是本步的类型护栏。）

- [ ] **Step 4: 提交**

```bash
git add frontend/src/components/compatibility/deep-analysis/DeepReportNarrative.tsx frontend/src/pages/CompatibilityResultPage.tsx
git commit -m "feat(compat-ui): render SpousePalaceMatch under personality comparison"
```

---

## Task 4: 人工可视化验证

**Files:** 无改动，仅验证。

- [ ] **Step 1: 起前端**

若用 vite dev：`npm run dev`（看终端给出的本地地址，spec 提到 :5200）。若是 docker 前端容器且非热更新，需重建/重启该容器再访问 :3000。

- [ ] **Step 2: 打开含字段的报告**

访问已知含 `spouse_palace_match` 的报告：`/compatibility/5a6073e0-547b-4df2-aebc-b53670a49963`。
Expected：
- 「双方性格画像与差异」卡片**正下方**出现「夫妻宫匹配」卡片。
- 左右两列：左列标题「{self名}理想的另一半」+ 契合度徽章（该报告 self=中、partner=低，颜色金/红）；每列有画像段落、「对上了」「有差距」列表。
- 卡片底部一句话 `summary`。
- 窄屏（浏览器拉窄到手机宽度）两列叠成上下，不溢出。

- [ ] **Step 3: 打开旧报告（无字段）**

打开任意一份在本功能上线前生成的旧合盘报告（或后端 `content_structured` 不含 `spouse_palace_match` 的报告）。
Expected：不出现「夫妻宫匹配」卡片，页面其余部分正常、控制台无报错。

- [ ] **Step 4: 无新提交**（仅验证；如发现问题回到对应 Task 修复）

---

## 验收（对齐 spec §7）

1. `npm run build`（含 `tsc -b`）与 `npm run lint` 通过。
2. 含字段的报告：性格画像下方出现「夫妻宫匹配」卡片，双列画像 + 高/中/低徽章 + summary 正常。
3. 旧报告：不渲染该卡片，无报错。
4. 窄屏两列叠成上下，不溢出。

## 明确不做（YAGNI，对齐 spec §8）

- 不改后端、不动其它报告区块。
- 不展示 `evidence_keys`。
- 不加 i18n、不引入前端测试框架。
- 不做「双向合并总档」。
