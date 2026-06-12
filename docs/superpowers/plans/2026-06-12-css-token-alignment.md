# CSS Token 对齐（严格不变样）实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在像素级不改变视觉的前提下，把 frontend/src 组件 CSS 中与 index.css 设计变量值完全相同的写死值替换为 `var()` 引用，并产出样式不一致建议清单。

**Architecture:** 纯 CSS 字面值替换，无逻辑改动。按类别分三轮（颜色 → 字号 → 圆角）独立提交，每轮以 build/lint + grep 计数验证。语义判断已在本计划中逐处给出，执行者照表替换即可。

**Tech Stack:** React + Vite 前端，纯 CSS（无预处理器）。验证命令：`cd frontend && npm run lint && npm run build`。

**Spec:** `docs/superpowers/specs/2026-06-12-css-token-alignment-design.md`

**关键规则（来自 spec §4）：**
1. 字面值与变量定义值完全一致才替换（忽略大小写/空格）。
2. 一值多变量按语义选（决策已写在下方各表）。
3. 语义命名变量（`--font-size-*` 等）还需使用场景匹配，拿不准不动。
4. 跳过 `CompatibilityShareCard.css`、`CompatibilityPrintLayout.css`（导出用浅色主题）和 `index.css`。

**所有文件路径相对 `frontend/src/`。行号为计划编写时的快照，执行时以"行号附近 + 原值匹配"定位，若对不上先 grep 确认。**

---

### Task 0: 基线验证

**Files:** 无修改

- [ ] **Step 1: 确认工作区干净**

Run: `cd /Users/liujiming/web/yuanju && git status --porcelain`
Expected: 无输出（如有未提交改动，停下来问用户）

- [ ] **Step 2: 基线 lint + build**

Run: `cd /Users/liujiming/web/yuanju/frontend && npm run lint && npm run build`
Expected: 两者均通过。如基线就失败，停下来报告，不开始替换。

---

### Task 1: 颜色轮

**Files:**
- Modify: `components/YongshenBadge.css`, `components/PolishedPanel.css`, `components/AdminLayout.css`, `components/compatibility/DayPillarPortrait.css`, `pages/AuthPage.css`, `pages/HomePage.css`, `pages/ResultPage.css`

- [ ] **Step 1: 替换 hex 颜色（14 处，逐行照表）**

| 位置 | 原值 | 替换为 | 语义依据 |
|---|---|---|---|
| `components/YongshenBadge.css:38` | `color: #4caf7d;` | `color: var(--status-success);` | 喜神标签=正面状态色，非五行木 |
| `components/YongshenBadge.css:44` | `color: #e05c4b;` | `color: var(--status-danger);` | 忌神标签=负面状态色 |
| `components/PolishedPanel.css:69` | `color: #0d0f14;` | `color: var(--bg-base);` | 金底按钮上的深色文字 |
| `components/AdminLayout.css:183` | `background: #1a1f2e;` | `background: var(--bg-card);` | 卡片底色 |
| `components/AdminLayout.css:198` | `background: #1a1f2e;` | `background: var(--bg-card);` | 表格行 hover 底色 |
| `pages/AuthPage.css:68` | `border-top-color: #0d0f14;` | `border-top-color: var(--bg-base);` | 金底 spinner 的深色段 |
| `pages/HomePage.css:138` | `border-top-color: #0d0f14;` | `border-top-color: var(--bg-base);` | 同上 |
| `pages/ResultPage.css:540` | `var(--wu-jin, #C9A84C)` | `var(--wu-jin)` | 冗余 fallback（变量必定已定义） |
| `pages/ResultPage.css:599` | `var(--bg-secondary, #1a1f2e)` | `var(--bg-card)` | **修复**：`--bg-secondary` 从未定义，实际靠 fallback 渲染；`--bg-card` 值即 `#1a1f2e` |
| `pages/ResultPage.css:633` | `background: #C9A84C;` | `background: var(--wu-jin);` | 神煞"吉"点=品牌金 |
| `pages/ResultPage.css:652` | `color: #C9A84C;` | `color: var(--wu-jin);` | 神煞"吉"徽章文字，与 633 同语义 |
| `pages/ResultPage.css:914` | `color: #0d0f14;` | `color: var(--bg-base);` | 金底元素深色文字 |
| `pages/ResultPage.css:1264` | `color: #0d0f14;` | `color: var(--bg-base);` | 同上 |
| `pages/ResultPage.css:1733` | `color: #0d0f14;` | `color: var(--bg-base);` | 金底激活 tab 深色文字 |

- [ ] **Step 2: 替换 rgba 颜色（仅语义匹配的 8 处）**

| 位置 | 原值 | 替换为 | 语义依据 |
|---|---|---|---|
| `components/YongshenBadge.css:92` | `var(--tag-color, rgba(255,255,255,0.1))` | `var(--tag-color, var(--border-default))` | border 用途；同文件 133 行已有此写法先例 |
| `components/compatibility/DayPillarPortrait.css:56` | `border: 1px solid rgba(201, 168, 76, 0.3);` | `border: 1px solid var(--border-accent);` | 金色强调边框 |
| `pages/ResultPage.css:654` | `border: 1px solid rgba(201, 168, 76, 0.3);` | `border: 1px solid var(--border-accent);` | 同上 |
| `pages/ResultPage.css:1291` | `border: 1px solid rgba(201, 168, 76, 0.3);` | `border: 1px solid var(--border-accent);` | 同上 |
| `pages/ResultPage.css:664` | `border: 1px solid rgba(255, 255, 255, 0.1);` | `border: 1px solid var(--border-default);` | 默认边框 |
| `pages/ResultPage.css:669` | `border: 1px solid rgba(255, 255, 255, 0.1);` | `border: 1px solid var(--border-default);` | 同上 |
| `pages/ResultPage.css:1666` | `box-shadow: 0 0 20px rgba(201, 168, 76, 0.2);` | `box-shadow: 0 0 20px var(--primary-glow);` | 金色光晕，正是 glow 语义 |
| `pages/ResultPage.css:1697` | `drop-shadow(0 0 2px rgba(201,168,76,0.2))` | `drop-shadow(0 0 2px var(--primary-glow))` | 同上 |

**以下值虽相等但语义不匹配，按 spec 规则 3 跳过（Task 4 记入建议清单，原文照录）：**

| 位置 | 原值 | 跳过原因 |
|---|---|---|
| `components/compatibility/deep-analysis/DeepReportNarrative.css:50` | `border: 1px solid rgba(201, 168, 76, 0.2)` | 值=--primary-glow 但用作边框；建议清单提议归一到 --border-accent(0.3) |
| `pages/CompatibilityHistoryPage.css:210` | 同上 | 同上 |
| `pages/CompatibilityPage.css:149` | 同上 | 同上 |
| `pages/ResultPage.css:779` | 同上 | 同上 |
| `pages/ResultPage.css:1144` | 同上 | 同上 |
| `pages/HomePage.css:176` | `drop-shadow(0 0 8px rgba(201,168,76,0.3))` | 值=--border-accent 但用作光晕 |
| `pages/ResultPage.css:41` | `text-shadow: 0 0 30px rgba(201,168,76,0.3)` | 同上 |
| `pages/ResultPage.css:530` | `background: rgba(255, 255, 255, 0.06)` | 值=--border-subtle 但用作背景 |

- [ ] **Step 3: 验证替换计数归零/到位**

Run:
```bash
cd /Users/liujiming/web/yuanju/frontend/src
grep -rin '#c9a84c\|#4caf7d\|#e05c4b\|#0d0f14\|#1a1f2e' --include='*.css' . | grep -v 'index.css\|ShareCard\|PrintLayout'
```
Expected: 无输出（14 处 hex 已全部替换）。

```bash
grep -rinE 'rgba\(201, ?168, ?76, ?0\.3\)|rgba\(255, ?255, ?255, ?0\.1\)\;|rgba\(201, ?168, ?76, ?0\.2\)' --include='*.css' . | grep -v 'index.css\|ShareCard\|PrintLayout'
```
Expected: 仅剩 Step 2 跳过表中列出的 8 处。

- [ ] **Step 4: lint + build**

Run: `cd /Users/liujiming/web/yuanju/frontend && npm run lint && npm run build`
Expected: 均通过。

- [ ] **Step 5: Commit**

```bash
cd /Users/liujiming/web/yuanju
git add frontend/src
git commit -m "refactor(css): replace exact-match hardcoded colors with design tokens

Pixel-identical: every literal equals the token's defined value.
Also fixes var(--bg-secondary) referencing an undefined token.

Co-Authored-By: Claude Fable 5 <noreply@anthropic.com>"
```

---

### Task 2: 字号轮

**Files:**
- Modify: 候选分布于全部非排除 CSS 文件，以 grep 结果为准

仅 5 个字号值有对应变量，且变量是语义命名，**必须场景匹配才换**：

| 值 | 变量 | 允许替换的场景 | 典型不替换场景 |
|---|---|---|---|
| 32px | `--font-size-page-title` | 页面唯一主标题（h1 级，如 `.page-title`、hero 标题） | 大号数字展示（如评分） |
| 22px | `--font-size-section-title` | 区块标题（h2 级，选择器含 section/title 且为区块级） | 大徽章、数值 |
| 16px | `--font-size-card-title` | 卡片标题 | 正文偏大字号、输入框文字 |
| 15px | `--font-size-body` | 正文段落 | 按钮文字、表单控件 |
| 12px | `--font-size-caption` | 辅助说明/时间戳/meta 文字 | 徽章、标签、图表刻度 |

- [ ] **Step 1: 生成候选清单**

Run:
```bash
cd /Users/liujiming/web/yuanju/frontend/src
grep -rn 'font-size: 32px\|font-size: 22px\|font-size: 16px\|font-size: 15px\|font-size: 12px' --include='*.css' . | grep -v 'index.css\|ShareCard\|PrintLayout'
```
Expected: 约 105 行候选（32px×3、22px×12、16px×18、15px×14、12px×58）。

- [ ] **Step 2: 逐行判断并替换**

对每个候选行，读其选择器名和上下 5 行判断场景。判断标准：选择器名能明确对应上表"允许场景"（如 `.history-page-title`、`.section-title`、`.card-meta`）才替换；含 badge/btn/tag/chip/tab/input/score/num 等字样或拿不准的一律跳过。把每处"替换/跳过"决定记录到临时文件 `/tmp/fontsize-decisions.txt`（格式：`文件:行号 | 值 | 替换为/跳过 | 一句话原因`），供 Task 4 使用。

预期：12px 的 58 处大部分是徽章/标签会跳过；总替换量预计 20–40 处。这是正常结果，不要为了凑数放宽标准。

- [ ] **Step 3: 验证（抽样核对 + 构建）**

Run: `git -C /Users/liujiming/web/yuanju diff --stat && git -C /Users/liujiming/web/yuanju diff | grep '^[+-].*font-size' | head -40`
核对：每处 `+` 行的变量档位值与 `-` 行字面值一致（32/22/16/15/12 对应无误）。

Run: `cd /Users/liujiming/web/yuanju/frontend && npm run lint && npm run build`
Expected: 均通过。

- [ ] **Step 4: Commit**

```bash
cd /Users/liujiming/web/yuanju
git add frontend/src
git commit -m "refactor(css): use type-scale tokens where value and semantics both match

Conservative pass: only selectors clearly matching the token's semantic
(page/section/card title, body, caption) were converted.

Co-Authored-By: Claude Fable 5 <noreply@anthropic.com>"
```

---

### Task 3: 形状轮（border-radius）

**Files:**
- Modify: `components/TiaohouCard.css`, `components/AdminLayout.css`, `components/compatibility/ParticipantSummaryCard.css`, `components/compatibility/ScoreOverview.css`, `components/compatibility/DayPillarPortrait.css`, `components/compatibility/deep-analysis/ActionPlan7d30d.css`, `pages/HistoryPage.css`, `pages/ResultPage.css`, `pages/CompatibilityHistoryPage.css`

`--radius-sm/md/lg` 是取值类变量（非语义命名），精确匹配即替换。transition 和 box-shadow 类别经统计精确匹配为 0 处，本轮不涉及。

- [ ] **Step 1: 替换 21 处（逐行照表）**

`border-radius: 6px` → `border-radius: var(--radius-sm)`（5 处）：
- `components/AdminLayout.css:100`
- `components/compatibility/DayPillarPortrait.css:57`
- `pages/ResultPage.css:870`
- `pages/ResultPage.css:984`
- `pages/ResultPage.css:1145`

`border-radius: 12px` → `border-radius: var(--radius-md)`（13 处）：
- `components/TiaohouCard.css:4`
- `components/AdminLayout.css:124`
- `components/AdminLayout.css:147`
- `components/compatibility/ParticipantSummaryCard.css:55`
- `components/compatibility/ParticipantSummaryCard.css:88`
- `components/compatibility/ScoreOverview.css:4`
- `components/compatibility/deep-analysis/ActionPlan7d30d.css:171`
- `pages/HistoryPage.css:15`
- `pages/HistoryPage.css:64`
- `pages/HistoryPage.css:88`
- `pages/HistoryPage.css:327`
- `pages/ResultPage.css:1795`
- `pages/CompatibilityHistoryPage.css:154`

`border-radius: 20px` → `border-radius: var(--radius-lg)`（3 处）：
- `components/AdminLayout.css:204`
- `pages/ResultPage.css:87`
- `pages/ResultPage.css:647`

注：`ResultPage.css:647` 是小药丸徽章，20px 实际起"全圆角"作用——值相等照换，但 Task 4 建议清单中提议将来归一为 `--radius-full`（连同 `99px`/`999px` 的 13 处）。

- [ ] **Step 2: 验证计数归零**

Run:
```bash
cd /Users/liujiming/web/yuanju/frontend/src
grep -rn 'border-radius: 6px\|border-radius: 12px\|border-radius: 20px' --include='*.css' . | grep -v 'index.css\|ShareCard\|PrintLayout'
```
Expected: 无输出。

- [ ] **Step 3: lint + build**

Run: `cd /Users/liujiming/web/yuanju/frontend && npm run lint && npm run build`
Expected: 均通过。

- [ ] **Step 4: Commit**

```bash
cd /Users/liujiming/web/yuanju
git add frontend/src
git commit -m "refactor(css): replace exact-match border-radius literals with radius tokens

Co-Authored-By: Claude Fable 5 <noreply@anthropic.com>"
```

---

### Task 4: 建议清单文档

**Files:**
- Create: `docs/superpowers/specs/2026-06-12-css-style-suggestions.md`

- [ ] **Step 1: 收集数据**

依次运行并保存输出：

```bash
cd /Users/liujiming/web/yuanju/frontend/src
# 1. 近似金色及其他杂色（按出现次数排序，含文件分布）
grep -rinoE '#[0-9a-f]{3,8}\b' --include='*.css' . | grep -v 'index.css\|ShareCard\|PrintLayout' | awk -F: '{print $NF}' | tr 'A-F' 'a-f' | sort | uniq -c | sort -rn
# 2. 写死字号全分布
grep -rhoE 'font-size:\s*[0-9.]+px' --include='*.css' . | grep -v node_modules | sort | uniq -c | sort -rn
# 3. 药丸圆角（999px/99px → --radius-full 候选）
grep -rn 'border-radius: 999px\|border-radius: 99px' --include='*.css' . | grep -v 'index.css\|ShareCard\|PrintLayout'
# 4. 秒制 transition（与 token 的毫秒制并存）
grep -rhoE 'transition:[^;]+' --include='*.css' . | grep -oE '[0-9.]+s\b' | sort | uniq -c | sort -rn
# 5. TSX inline style 规模
grep -rn 'style={{' --include='*.tsx' . | wc -l
# 6. 页面级横向 padding 是否对齐 --space-page-x(24px)
grep -rn 'padding' --include='*.css' pages/*.css | grep -E 'padding(-left|-right)?:.*(16|20|24|28|32)px' | head -30
```

- [ ] **Step 2: 写文档**

创建 `docs/superpowers/specs/2026-06-12-css-style-suggestions.md`，结构如下（用 Step 1 实际数据填充，每条建议带 `文件:行号` 或 grep 命令；数据已在手，不留 TBD）：

```markdown
# CSS 样式不一致建议清单

> Date: 2026-06-12
> 来源: docs/superpowers/specs/2026-06-12-css-token-alignment-design.md 的 Task 4
> 性质: 以下每条都会**改变渲染结果**（归一化），需逐项人工决策后才执行。

## 1. 近似金色合并候选
（#c9a227 ×12、#c9a96e ×10、#e0cca0 ×7 等与标准金 --wu-jin #c9a84c 的对照表，
 含出现位置；建议统一为 --wu-jin 或确认为有意的层次色后入 token）

## 2. 语义错位的颜色复用（Task 1 跳过的 8 处，原文照录）
（值恰好等于某 token 但用途不符，建议按用途归一，如金色 0.2 边框 → --border-accent）

## 3. 字号档位收敛
（15 档 → 5 档 type scale 的映射建议表 + Task 2 的 /tmp/fontsize-decisions.txt 中跳过项；
 9px/10px/11px 小字号建议明确是否保留为图表专用档）

## 4. 药丸圆角归一
（999px/99px/20px-药丸 → --radius-full，附位置清单）

## 5. transition 时长制式
（组件多用 0.2s/0.3s 秒制，token 为 150/250/400ms 毫秒制；建议归一方向）

## 6. 可疑对齐点
（Step 1 第 6 项数据：各页面横向 padding 与 --space-page-x 的偏差；其他人工发现）

## 7. 后续治理方向
- TSX inline style（约 759 处，引用 2026-06-01 审计）
- 导出类文件（ShareCard/PrintLayout）的浅色主题是否值得独立 token 组

## 附录：一值多变量的语义选择记录
（照录 Task 1 表中的"语义依据"列：#4caf7d→--status-success 而非 --wu-mu、
 #e05c4b→--status-danger 而非 --wu-huo、#C9A84C→--wu-jin 而非 --primary/--text-accent 等）
```

- [ ] **Step 3: 自查文档**

通读一遍：无 TBD/占位符；每条建议有位置数据；第 2 节 8 处与 Task 1 跳过表一致。

- [ ] **Step 4: Commit**

```bash
cd /Users/liujiming/web/yuanju
git add docs/superpowers/specs/2026-06-12-css-style-suggestions.md
git commit -m "docs: css style inconsistency suggestions (normalization candidates)

Co-Authored-By: Claude Fable 5 <noreply@anthropic.com>"
```

---

## 完成标准（对照 spec §6）

- [ ] 三轮替换各自独立提交，`npm run lint && npm run build` 全程绿
- [ ] 每处替换的变量定义值 == 被替换字面值（Task 1/3 照表保证，Task 2 经 diff 核对）
- [ ] 排除文件（ShareCard/PrintLayout/index.css）零改动：`git log --stat` 确认
- [ ] 建议清单文档已提交，覆盖 spec §2 第 3 点全部类别
