# 合盘分享卡内容补全 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 给 `CompatibilityShareCard` 增加 4 个内容块（四维评分理由 / 关系诊断 3 发现 / 关系策略 4 维 / 命理解读六章节摘要），让分享卡承担"报告封面图"角色。

**Architecture:** 修改 `CompatibilityShareCard.tsx` 加 `structured` prop 与 4 段新 JSX，修改 `CompatibilityShareCard.css` 加对应样式，修改 `CompatibilityResultPage.tsx` 传 `structured={structuredReport}`。无新组件、无新依赖、无后端改动。

**Tech Stack:** React 19 + TypeScript strict + Vite。无前端测试框架——验证靠 `tsc -b && vite build` (`npm run build`) + `npm run lint` + 手工 smoke。

**Spec:** `docs/superpowers/specs/2026-05-28-compatibility-share-card-detail-design.md`

**Branch policy:** 在 `main` 上直接迭代（项目惯例）。每个 task 一个 commit。所有 commit 必须含 `Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>` trailer。

---

## File Structure

```
修改（3 个）:
  frontend/src/components/CompatibilityShareCard.tsx  (+~80 行：prop + 派生 + 4 段 JSX)
  frontend/src/components/CompatibilityShareCard.css  (+~80 行：新增 6 类样式)
  frontend/src/pages/CompatibilityResultPage.tsx       (+1 行：structured={structuredReport})
```

ESLint 注意：`no-unused-vars: error` 项目策略，T1 必须把 ShareCard 的 prop 改动**与**页面传参改动**合并到一个 commit**，否则任一文件单独提交会 lint 失败（同上次 T3+T4 经验）。

---

## Pre-flight checks

- [ ] **A. 工作区干净**
  Run: `cd /Users/liujiming/web/yuanju && git status`
  Expected: clean tree on `main`. (注：可能有 untracked `docs/superpowers/plans/2026-05-28-chart-archive-naming-at-creation.md` 是上次会话遗留，不相关，不处理。)

- [ ] **B. 类型确认**
  Run: `grep -nE "^export interface (CompatibilityRelationshipDiagnosis|CompatibilityRelationshipStrategy|CompatibilityFinding|CompatibilityScoreExplanation)\b" /Users/liujiming/web/yuanju/frontend/src/lib/api.ts`
  Expected: 4 hits.

- [ ] **C. cleanReportText 导出**
  Run: `grep -n "export function cleanReportText" /Users/liujiming/web/yuanju/frontend/src/lib/reportText.ts`
  Expected: 1 hit (line 7).

- [ ] **D. 读 spec**
  Read: `docs/superpowers/specs/2026-05-28-compatibility-share-card-detail-design.md`

---

## Task 1: ShareCard 组件改造 + 页面传参（合并提交）

**Files:**
- Modify: `frontend/src/components/CompatibilityShareCard.tsx`
- Modify: `frontend/src/pages/CompatibilityResultPage.tsx`

**Spec 引用:** §3.1（布局）、§3.3（派生）、§3.4（Props）、§4（错误处理）

合并提交原因：ShareCard 的 `structured` prop 在接口里是必填的；若 T1 只改 ShareCard 但不在调用处传值，TypeScript 报 missing-prop 错误，lint+build 红。

- [ ] **Step 1: 编辑 `frontend/src/components/CompatibilityShareCard.tsx`**

在文件顶部找到 `import type { ... }` 块（line 2-10），追加 `CompatibilityStructuredReport`：

```ts
import type {
  CompatibilityEvidence,
  CompatibilityParticipant,
  CompatibilityReading,
  CompatibilityStageRisk,
  CompatibilityStructuredReport,
  ExportBrand,
} from '../lib/api'
```

同时追加 `cleanReportText` 导入（紧跟 brandText 导入之后）：

```ts
import { cleanReportText } from '../lib/reportText'
```

- [ ] **Step 2: 顶部常量区追加 `STRATEGY_LABEL` 和 `firstSentence` 工具函数**

在 `EVIDENCE_SOURCE_LABEL` 常量定义之后（line 70 后）追加：

```ts
const STRATEGY_LABEL: Record<string, string> = {
  communication: '沟通',
  conflict: '冲突',
  reality: '现实',
  boundary: '边界',
}

function firstSentence(content: string): string {
  const cleaned = cleanReportText(content)
  const match = cleaned.match(/^[^。！？]+[。！？]?/)
  return (match?.[0] || cleaned).slice(0, 60).trim()
}
```

- [ ] **Step 3: Props 接口增加 `structured`**

定位 `CompatibilityShareCardProps`（line 139-146），改为：

```ts
export interface CompatibilityShareCardProps {
  reading: CompatibilityReading
  participants: CompatibilityParticipant[]
  evidences: CompatibilityEvidence[]
  decision: DecisionDashboardData
  stageRisks: CompatibilityStageRisk[]
  structured: CompatibilityStructuredReport | null
  brand?: ExportBrand | null
}
```

- [ ] **Step 4: 函数体改 destructure 加 `structured`，派生新数据**

定位 line 149 的 destructure，改为：

```ts
const { reading, participants, evidences, decision, stageRisks, structured, brand } = props
```

在现有的 `dimEntries` 派生之后（约 line 174），追加这些派生：

```ts
const explByDim = Object.fromEntries(
  (reading.score_explanations || []).map(e => [e.dimension, e])
)
const topFindings = (structured?.relationship_diagnosis?.top_findings || []).slice(0, 3)
const strategy = structured?.relationship_strategy
const strategyEntries = strategy ? [
  { key: 'communication', label: STRATEGY_LABEL.communication, value: strategy.communication },
  { key: 'conflict',      label: STRATEGY_LABEL.conflict,      value: strategy.conflict },
  { key: 'reality',       label: STRATEGY_LABEL.reality,       value: strategy.reality },
  { key: 'boundary',      label: STRATEGY_LABEL.boundary,      value: strategy.boundary },
].filter(e => e.value?.trim()) : []
const dimDigests = (structured?.dimensions || []).map(d => ({
  key: d.key,
  title: d.title,
  digest: firstSentence(d.content),
}))
```

- [ ] **Step 5: §4 四维 section 改造——bar 下方加「正：X · 负：Y」一行**

定位 `<section className="compat-share-dims">` 块（line 203-216），把里面的 `dimEntries.map(...)` 部分替换为：

```tsx
{dimEntries.map(d => {
  const expl = explByDim[d.key]
  const pos = expl?.positive_factor?.trim()
  const neg = expl?.negative_factor?.trim()
  const showWhy = Boolean(pos || neg)
  return (
    <div key={d.key} className="compat-share-dim-block">
      <div className="compat-share-dim-row">
        <span className="compat-share-dim-lbl">{d.label}</span>
        <div className="compat-share-dim-bar">
          <i style={{ width: `${Math.max(0, Math.min(100, d.value))}%` }} />
        </div>
        <span className="compat-share-dim-val">{d.value}</span>
      </div>
      {showWhy && (
        <div className="compat-share-dim-why">
          {pos && <span className="compat-share-dim-why-pos">正：{pos}</span>}
          {pos && neg && <span className="compat-share-dim-why-sep"> · </span>}
          {neg && <span className="compat-share-dim-why-neg">负：{neg}</span>}
        </div>
      )}
    </div>
  )
})}
```

注意：包了一层 `<div className="compat-share-dim-block">` 把 dim-row 和 dim-why 合并，便于 margin 控制。

- [ ] **Step 6: §6 关系诊断 · 核心 3 发现（新增）**

定位 §5 关系决策 section 的结束 `</section>`（约 line 226），紧接其后新增：

```tsx
{topFindings.length > 0 && (
  <section className="compat-share-findings">
    <h3 className="compat-share-section-h">◇ 关系诊断 · 核心发现</h3>
    {topFindings.map((f, i) => (
      <div key={i} className="compat-share-finding-item">· {f.text}</div>
    ))}
  </section>
)}
```

- [ ] **Step 7: §10 关系策略（新增）**

定位 §8 避免 section 的结束 `</section>`（约 line 257），紧接其后新增：

```tsx
{strategyEntries.length > 0 && (
  <section className="compat-share-strategy">
    <h3 className="compat-share-section-h">◇ 关系策略</h3>
    {strategyEntries.map(e => (
      <div key={e.key} className="compat-share-strategy-row">
        <span className="compat-share-strategy-lbl">{e.label}</span>
        <span className="compat-share-strategy-val">{e.value}</span>
      </div>
    ))}
  </section>
)}
```

- [ ] **Step 8: §11 命理解读摆要（新增）**

紧接 §10 关系策略 section 的结束 `</section>`，在 footer 之前新增：

```tsx
{dimDigests.length > 0 && (
  <section className="compat-share-digests">
    <h3 className="compat-share-section-h">◇ 命理解读摆要</h3>
    {dimDigests.map(d => (
      <div key={d.key} className="compat-share-digest-item">
        <div className="compat-share-digest-title">• {d.title}</div>
        <div className="compat-share-digest-line">{d.digest}</div>
      </div>
    ))}
  </section>
)}
```

- [ ] **Step 9: 编辑 `frontend/src/pages/CompatibilityResultPage.tsx`，传 structured prop**

定位 modal 内的 `<CompatibilityShareCard ref={shareCardRef} ... />` JSX（用 grep 找：`grep -n 'CompatibilityShareCard' frontend/src/pages/CompatibilityResultPage.tsx`）。在 props 列表里追加 `structured={structuredReport}`（`structuredReport` 已在 `CompatibilityResultPage.tsx:936` 派生过，在 scope 里）：

```tsx
<CompatibilityShareCard
  ref={shareCardRef}
  reading={reading}
  participants={detail.participants}
  evidences={detail.evidences}
  decision={decisionDashboard}
  stageRisks={decisionStageRisks}
  structured={structuredReport}
  brand={brand}
/>
```

- [ ] **Step 10: 验证 lint + build**

Run: `cd /Users/liujiming/web/yuanju/frontend && npm run lint && npm run build`
Expected: 全绿。无 unused-vars / missing-prop 错误。

- [ ] **Step 11: Commit (单 commit 覆盖两文件)**

```bash
cd /Users/liujiming/web/yuanju
git add frontend/src/components/CompatibilityShareCard.tsx frontend/src/pages/CompatibilityResultPage.tsx
git commit -m "$(cat <<'EOF'
feat(compat-export): expand share card to 13 sections (logic)

Per spec §3.3, add four new content blocks via 就近合并 layout:

- §4 each dim bar gets a "正：X · 负：Y" subline from
  reading.score_explanations (indexed by dimension)
- §6 NEW 关系诊断·核心发现 from diagnosis.top_findings[].text
- §10 NEW 关系策略 from relationship_strategy (沟通/冲突/现实/边界)
- §11 NEW 命理解读摆要 from structured.dimensions[].content first
  sentence via firstSentence() — regex matches up to first 。/！/？
  punctuation, fallback to slice(0, 60)

Re-add structured: CompatibilityStructuredReport | null to
ShareCardProps (was removed in edfb406 once verdict moved to
decision.verdict; now genuinely needed for new sections).
CompatibilityResultPage passes structuredReport through. CSS
in next commit.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

## Task 2: ShareCard CSS for new sections

**Files:**
- Modify: `frontend/src/components/CompatibilityShareCard.css`

**Spec 引用:** §3.5（视觉规范）、§3.6（类名）

- [ ] **Step 1: 在 css 文件末尾（在 `.compat-share-watermark` 规则之前）追加新规则**

打开 `frontend/src/components/CompatibilityShareCard.css`，定位 `.compat-share-watermark` 规则的开始（用 grep 找）。在它**之前**追加：

```css
/* §4 ⤷ 四维「为什么」一行 */
.compat-share-dim-block {
  margin: 4px 0;
}
.compat-share-dim-why {
  font-size: 10px;
  line-height: 1.5;
  margin-top: 2px;
  padding-left: 44px;   /* 对齐到 bar 左边，跨越 label 列 (36px) + gap (8px) */
}
.compat-share-dim-why-pos { color: #2d5a3d; }
.compat-share-dim-why-neg { color: #8b1a1a; }
.compat-share-dim-why-sep { color: #9b815c; }

/* §6 关系诊断·核心发现 */
.compat-share-findings {
  position: relative;
  z-index: 1;
}
.compat-share-finding-item {
  font-size: 11px;
  line-height: 1.6;
  margin: 3px 0;
  padding: 4px 8px;
  background: #fbf3e0;
  border-left: 2px solid #7a5c3a;
  border-radius: 0 4px 4px 0;
  color: #4a3728;
}

/* §10 关系策略 */
.compat-share-strategy {
  position: relative;
  z-index: 1;
}
.compat-share-strategy-row {
  display: grid;
  grid-template-columns: 40px 1fr;
  gap: 6px;
  font-size: 11px;
  line-height: 1.6;
  margin: 3px 0;
  color: #4a3728;
}
.compat-share-strategy-lbl {
  color: #7a5c3a;
  font-weight: 700;
  text-align: center;
  background: #efe0bc;
  border-radius: 3px;
  letter-spacing: 1px;
}
.compat-share-strategy-val {
  color: #4a3728;
}

/* §11 命理解读摆要 */
.compat-share-digests {
  position: relative;
  z-index: 1;
}
.compat-share-digest-item {
  margin: 3px 0 5px;
}
.compat-share-digest-title {
  font-size: 11px;
  font-weight: 700;
  color: #7a5c3a;
  font-family: "Noto Serif SC", serif;
}
.compat-share-digest-line {
  font-size: 10px;
  line-height: 1.6;
  color: #4a3728;
  padding-left: 10px;
}
```

- [ ] **Step 2: 检查 dim-row 旧 margin 是否需要移除**

在文件中查找 `.compat-share-dim-row` 规则。如果该规则上有 `margin: 4px 0`（或类似），**改成 `margin: 0`**（因为现在 .compat-share-dim-block 提供 margin，dim-row 不应重复提供）。

Run: `grep -A4 '\.compat-share-dim-row' /Users/liujiming/web/yuanju/frontend/src/components/CompatibilityShareCard.css`

如果看到 margin 在 `.compat-share-dim-row` 上，用 Edit 工具改为 `margin: 0`。如果没有 margin，跳过这一 step。

- [ ] **Step 3: 验证 lint + build**

Run: `cd /Users/liujiming/web/yuanju/frontend && npm run lint && npm run build`
Expected: 全绿。

- [ ] **Step 4: Commit**

```bash
cd /Users/liujiming/web/yuanju
git add frontend/src/components/CompatibilityShareCard.css
git commit -m "$(cat <<'EOF'
style(compat-export): CSS for share card 4 new sections

Six new class rules per spec §3.6:
- .compat-share-dim-block + .compat-share-dim-why{-pos,-neg,-sep}
  for §4 dim factor subline (10px, 正绿 #2d5a3d / 负红 #8b1a1a)
- .compat-share-findings + .compat-share-finding-item for §6
  diagnosis findings with left border accent
- .compat-share-strategy + .compat-share-strategy-row/-lbl/-val
  for §10 4-row strategy with chip-style label column
- .compat-share-digests + .compat-share-digest-item/-title/-line
  for §11 six-chapter digest list

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

## Task 3: Manual smoke per spec §5 validation

**Files:** none (verification only)

**Spec 引用:** §5 18-point validation

- [ ] **Step 1: Dev server 已运行**

Run: `curl -s -o /dev/null -w "%{http_code}\n" http://localhost:5200/`
Expected: 200. 如果不是 200，启动：`cd /Users/liujiming/web/yuanju/frontend && npm run dev`

- [ ] **Step 2: 浏览器打开合盘结果详情页**

- 用户已有一份生成完 AI 报告的合盘（structuredReport 非空）
- 进入 `/compatibility/result/:id` 页

- [ ] **Step 3: 点「分享图片」打开 modal，逐条验证 spec §5 18 点**

**数据呈现 (8 条):**
1. §4 每条评分进度条下方有「正：X · 负：Y」一行，正绿负红
2. v3 评分模式下 §4 ⤷ 行不出现（只有 v2 attraction/stability/communication/practicality 才有 explanations）
3. §6 「核心发现」展示 `top_findings` 前 3 条，每条带 `·` 前缀
4. §10 「关系策略」4 行：沟通 / 冲突 / 现实 / 边界，每行 chip 标签 + 内容
5. §11 「命理解读摆要」展示 dimensions 章节标题 + 首句
6. §11 首句长度 ≤60 字，且断在 。/！/？ 之前
7. 缺失字段时 section 不出现（修改 backend response 暴露 / 直接 DevTools 改 detail.latest_report.content_structured 测试）
8. ShareCard 在最小数据下结构稳定

**视觉 (4 条):**
9. 新增 4 段与原 9 段风格统一
10. §6/§10/§11 间有视觉分隔
11. §4 ⤷ 行 10px，正绿 `#2d5a3d` 负红 `#8b1a1a`
12. 卡片宽度仍 400px

**导出 (3 条):**
13. 点「保存到本地」下载的 PNG 包含所有 13 段
14. PNG 高度 ≥1700px（macOS 上右键查看属性 / Windows 上属性面板查）
15. iOS Safari 「保存/分享」 / Android 下载 / 桌面 toPng 三端测试至少一端

**回归 (3 条):**
16. 原 9 段位置和样式不变
17. 单盘 ResultPage 的 ShareCard / PDF 完全不动
18. `npm run lint && npm run build` 全绿

- [ ] **Step 4: 失败处理**

任一条不通过：记录现象 → 回到对应 task 的对应 step 修复 → 重新跑 §5。

- [ ] **Step 5: verify commit**

全 18 条通过后：

```bash
cd /Users/liujiming/web/yuanju
git commit --allow-empty -m "$(cat <<'EOF'
verify(compat-export): share-card 13 sections smoke per §5 passed

All 18 verification points from
docs/superpowers/specs/2026-05-28-compatibility-share-card-detail-design.md §5
walked through on local dev server. Four new content blocks render
in modal preview and PNG download; original 9 sections unchanged;
single-chart ShareCard unaffected.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

## Final review

After T3 commit, dispatch final code-reviewer over `HEAD~3..HEAD` (the 3 commits of this feature: T1 + T2 + T3 verify).

---

## Self-review pass

**Spec coverage (§ → task):**
- §1 背景 → covered by plan goal
- §2 范围 in scope (4 项内容补全) → T1 step 5-8 cover all 4
- §2 out of scope (后端 / PDF / 单盘 / 双版) → none touched
- §3.1 13 段布局 → T1 step 5-8 全部 cover
- §3.2 数据流 → T1 derivation block (step 4)
- §3.3 派生代码 → T1 step 4
- §3.4 Props 接口 → T1 step 3
- §3.5 视觉规范 → T2 CSS
- §3.6 类名 → T2 CSS (6 类名 + 子类名)
- §4 错误处理矩阵 → 实现自动覆盖：optional chain + `.filter(e => e.value?.trim())` + `?.text` 在 JSX 内自动跳过
- §5 验证 → T3
- §6 文件清单 → File Structure section
- §7 决策摘要 → 无需实现
- §8 不引入的依赖 → none introduced

**Type consistency:**
- `firstSentence(content: string)` defined T1 step 2, used T1 step 4 ✓
- `explByDim` shape `Record<string, CompatibilityScoreExplanation>` T1 step 4, looked up by `d.key` in step 5 ✓
- `topFindings` element shape `CompatibilityFinding` (with `.text`) T1 step 4, accessed as `f.text` in step 6 ✓
- `strategyEntries` shape `{ key, label, value }` T1 step 4, mapped in step 7 with `e.key/e.label/e.value` ✓
- `dimDigests` shape `{ key, title, digest }` T1 step 4, mapped in step 8 with `d.key/d.title/d.digest` ✓

**Placeholder scan:**
- No "TBD" / "TODO" / "implement later" ✓
- All code blocks complete ✓
- All commands explicit with expected output ✓
