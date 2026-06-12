# CSS 样式不一致建议清单

> Date: 2026-06-12
> 来源: docs/superpowers/specs/2026-06-12-css-token-alignment-design.md 的 Task 4
> 性质: 以下每条都会**改变渲染结果**（归一化），需逐项人工决策后才执行。
> 勾选方式：把同意执行的条目标 [x]，后续按勾选项实施。
> 说明: 行号为本文撰写时的快照，执行时请重新 grep 定位（行号会随改动漂移）。
> 排除范围: index.css（token 定义源）、ShareCard、PrintLayout（导出浅色主题，另议）。

---

## 1. 近似金色合并候选

标准金 = `--wu-jin: #c9a84c`。以下色值视觉接近但不等于标准金，归一会轻微改变颜色。

| 色值 | 次数 | 与标准金关系 | 位置 | 建议 |
|------|------|------|------|------|
| `#c9a227` | 12 | 偏绿/暗的金，AdminLayout 专用 | components/AdminLayout.css:40,77,108,109,153,158,231,243,288,318,376（color/border/background 混用） | [ ] 后台整套金。要么整体归一到 `--wu-jin`，要么承认后台是独立视觉体系、保留并记录 |
| `#e0cca0` | 1 | 浅金/米金 | pages/ResultPage.css:2048（`border-bottom: 1px solid #e0cca0`） | [ ] 单点浅金边框，建议归一到 `--border-accent` 或 `--wu-jin`（按是否需要浅色调决定） |
| `#e0be75` | 4 | 浅金 | 见下方 grep（components/pages 散落） | [ ] 浅金一族，评估是否并入 `--wu-jin` 或新增 `--wu-jin-light` |
| `#c9a45b` | 2 | 极接近 `#c9a84c` | 见下方 grep | [ ] 几乎等于标准金，优先归一到 `--wu-jin` |
| `#f4c65f` `#e0b530` `#b8952a` `#a8872a` `#d88b2f` `#5a3a1a` | 各 1 | 金色渐变/描边族 | 多在渐变/阴影上下文 | [ ] 多为渐变端点，归一会破坏渐变层次，谨慎；逐个确认是否独立保留 |

重新定位命令：
```bash
cd frontend/src
grep -rin '#e0be75\|#c9a45b\|#c9a227\|#e0cca0' --include='*.css' . | grep -v 'index.css\|ShareCard\|PrintLayout'
```

注：`#c9a227` 全部集中在 AdminLayout，是一套自洽的后台金，与前台标准金分属两个界面体系——这是「整体决策」而非「逐点替换」，故未在 Task 1 自动替换。

---

## 2. 语义错位的颜色复用（Task 1 跳过的 8 处）

这些色值的 RGBA 与某个 token **数值相等**，但**用途/语义不同**（边框 vs 光晕 vs 背景）。直接套用同名 token 会让语义混乱，需先决定是否新增语义 token。

- `border: 1px solid rgba(201, 168, 76, 0.2)`（值 = `--primary-glow`，但用作**边框**；建议归一到 `--border-accent`(0.3) 或新增 `--border-accent-soft`(0.2)）：
  - [ ] components/compatibility/deep-analysis/DeepReportNarrative.css:50
  - [ ] pages/CompatibilityHistoryPage.css:210
  - [ ] pages/CompatibilityPage.css:149
  - [ ] pages/ResultPage.css:779
  - [ ] pages/ResultPage.css:1144
- [ ] pages/HomePage.css:176 — `filter: drop-shadow(0 0 8px rgba(201,168,76,0.3))`（值 = `--border-accent`，但语义是**光晕**，应是 glow 类 token）
- [ ] pages/ResultPage.css:41 — `text-shadow: 0 0 30px rgba(201,168,76,0.3)`（同上，光晕语义借用了 border 数值）
- [ ] pages/ResultPage.css:530 — `background: rgba(255, 255, 255, 0.06)`（值 = `--border-subtle`，但用作**背景**；建议新增 `--surface-faint` 之类，而非套 border token）

建议方向：先补两个语义 token（`--border-accent-soft: rgba(201,168,76,0.2)`、`--surface-faint: rgba(255,255,255,0.06)`，并视情况补一个 glow token），再让上述位置引用语义正确的 token，避免「数值对了语义错了」。

重新定位命令：
```bash
cd frontend/src
grep -rn 'rgba(201, 168, 76, 0.2)\|rgba(201,168,76,0.3)\|rgba(255, 255, 255, 0.06)' --include='*.css' . | grep -v 'index.css\|ShareCard\|PrintLayout'
```

---

## 3. 字号档位收敛

现有 token 档：`--font-size` 系列 32/22/16/15/12px（外加 section-title/card-title/body/caption 语义名）。实际 CSS 中的完整分布：

| 字号 | 出现次数 | 有 token? |
|------|------|------|
| 13px | 105 | ✗ 无档位 |
| 12px | 52 | ✓ caption |
| 14px | 47 | ✗ 无档位 |
| 11px | 32 | ✗ 无档位 |
| 10px | 19 | ✗ 无档位 |
| 18px | 18 | ✗ 无档位 |
| 16px | 14 | ✓ card-title |
| 15px | 13 | ✓ body |
| 20px | 12 | ✗ 无档位 |
| 24px | 8 | ✗ 无档位 |
| 9px | 7 | ✗ 无档位 |
| 28px | 7 | ✗ 无档位 |
| 22px | 5 | ✓ section-title |
| 30px | 4 | ✗ 无档位 |
| 32px | 3 | ✓（最大档） |
| 26/25/17px | 3/2/2 | ✗ |
| 56/44/40px | 各 1 | ✗ |

**无 token 档位的规模**：9/10/11/13/14px 合计约 **203 处**（13px 自己就 105 处，是绝对大头），全部落在 token 缝隙里。这不是逐条替换问题，而是「要不要扩档」的体系决策。

从 `/tmp/fontsize-decisions.txt` 的 SKIPPED 项归纳出几类典型，它们语义清晰、复用度高，是**新增语义档**的主要候选：

- **badge / tag / chip 类**（普遍 12px）：YongshenBadge:128、AdminLayout:205、ScoreOverview:23、EvidenceDrawer:86、SpousePalaceMatch:33、ResultPage:58/1148、ProfilePage:241……
  - [ ] 建议：是否新增 `--font-size-badge`（12px 或 11px），统一徽章/标签视觉。
- **按钮 / 控件类**（16px / 15px / 12px 混杂）：Button.css:63、PolishedPanel:132、HistoryPage:205、ResultPage:1113/1445/1450……
  - [ ] 建议：是否新增 `--font-size-button`，收敛按钮文字。
- **数值展示类（value/data/score）**（12px ~ 32px 跨度大）：AdminLayout:156(32px)、ParticipantSummaryCard:15(22px)、ResultPage:380/489/504/1586、CompatibilityStickyHeader:57……
  - [ ] 建议：数值展示跨度太大，按层级拆 `--font-size-stat-lg / -md` 或维持各自上下文，**不建议强行统一**。
- **label / kicker / overline 类**（普遍 12px）：AdminLayout:164/268、ResultPage:220/296/454/1240、ProfilePage:67、CompatibilityResultPage:88……
  - [ ] 建议：是否新增 `--font-size-overline`（12px，常配 letter-spacing/uppercase）。

**两处被「硬规则」保守回退、但实为 caption 的，可优先勾选：**
- [ ] components/PolishedPanel.css:56 — `.polished-input-meta`，12px，值/语义都符合 caption，仅因选择器含 `input/score` 关键词被跳过 → 直接归一到 `--font-size-caption`
- [ ] components/compatibility/ScoreOverview.css:100 — `.compatibility-quick-score-hint`，12px，同上（含 `score` 关键词） → 直接归一到 `--font-size-caption`

---

## 4. 药丸圆角归一

药丸（pill）圆角散落两套写法：`999px`（9 处）与 `99px`（4 处），且无 token。现有 token 有 `--radius-full: 9999px`。

- [ ] components/compatibility/ScoreOverview.css:22
- [ ] components/compatibility/ScoreOverview.css:120
- [ ] components/compatibility/deep-analysis/SpousePalaceMatch.css:32
- [ ] pages/ProfilePage.css:244
- [ ] components/compatibility/EvidenceDrawer.css:88
- [ ] pages/CompatibilityHistoryPage.css:234
- [ ] pages/CompatibilityHistoryPage.css:298
- [ ] pages/CompatibilityPage.css:315
- [ ] pages/CompatibilityPage.css:344
- [ ] pages/ResultPage.css:811（`99px`）
- [ ] pages/ResultPage.css:912（`99px`）
- [ ] pages/ResultPage.css:1014（`99px`）
- [ ] pages/ResultPage.css:1379（`99px`）

外加一处「语义药丸却用了 radius-lg」：
- [ ] pages/ResultPage.css:647 `.shensha-modal-badge` 现用 `var(--radius-lg)`(20px)，但形态是药丸徽章 → 建议统一到 `--radius-full`（注意：会从 20px 圆角变成全圆，视觉变化较明显，需确认设计意图）

建议：999px/99px 全部归一到 `--radius-full`；`.shensha-modal-badge` 需设计确认后再动。

---

## 5. transition 时长制式

token 的过渡时长是 **150 / 250 / 400ms**（ease）。但组件里普遍用**秒制**，且数值与 token **不相等**：

| 时长 | 出现次数 | 与 token 关系 |
|------|------|------|
| 0.15s | 12 | = 150ms（数值相等，仅单位不同，可安全归一） |
| 0.18s | 11 | ≠ 任何 token（介于 150/250 之间） |
| 0.2s | 6 | ≠ 任何 token（介于 150/250 之间） |
| 0.16s | 4 | ≠ 任何 token |

建议方向（二选一，需决策）：
- [ ] 方案 A（不改动画手感）：扩 token 档，新增 `--transition-fast: 180ms` 等覆盖现有 0.16/0.18/0.2s，再让 CSS 引用 token——零渲染变化，但 token 档变多。
- [ ] 方案 B（统一手感）：把 0.16/0.18/0.2s 全部归一到最近的现有档（多数 → 150ms 或 250ms）——token 干净，但会改变动画时长（变快或变慢），需接受手感变化。
- [ ] 无争议子项：12 处 `0.15s` 数值等于 `--transition-fast`(150ms)，可**先安全归一**（仅换写法不改时长）。

重新定位命令：
```bash
cd frontend/src
grep -rn 'transition:[^;]*0\.\(15\|16\|18\|2\)s' --include='*.css' . | grep -v 'index.css\|ShareCard\|PrintLayout'
```

---

## 6. 可疑对齐点（页面横向 padding 偏离 --space-page-x:24px）

`--space-page-x: 24px` 是页面级横向内边距标准。但页面级容器/卡片普遍用 16px / 20px / 32px，未对齐。挑出**页面级容器**的明显偏差（卡片内边距属组件内距，不强求对齐，下列只列页面/区块级）：

- [ ] pages/HomePage.css:170 `.feature-card { padding: 32px 24px }` — 横向 24px 已对齐，纵向另议
- [ ] pages/HomePage.css:65/160 `.form-section`/`.features-section`（横向 0，纵向 padding）— 与 page-x 无关，可忽略
- [ ] pages/CompatibilityPage.css:43/65/199 `padding: 20px`（区块容器，横向 20px ≠ 24px）
- [ ] pages/ProfilePage.css:11 `padding: 20px`（页面根容器，横向 20px ≠ 24px）
- [ ] pages/HistoryPage.css:319/361 `padding: 20px`、:125 `padding: 48px 32px`（区块横向 20/32px ≠ 24px）
- [ ] pages/CompatibilityHistoryPage.css:146/323 `padding: 20px`、:51 `padding: 48px 32px`（同上）
- [ ] pages/ResultPage.css:1358 `padding: 28px 24px`、:1648 `padding: 40px 20px`、:1607 `padding: 32px 16px`（区块横向 16/20px ≠ 24px）

建议：页面/区块级横向 padding 统一到 `--space-page-x`(24px)；16px 的小卡片内距属组件内部留白，**不**强制对齐。需逐个确认哪些是「页面级容器」（应对齐）vs「卡片内距」（保留）。

完整原始列表见 Step 1 命令 G（输出较杂，含大量卡片内距，已在此筛掉）。

---

## 7. 后续治理方向

- [ ] **TSX inline style**：当前 `style={{` 共 **748** 处（`grep -rn 'style={{' --include='*.tsx' frontend/src | wc -l`）。对照 2026-06-01 UX 审计记录的 759 处，量级一致、略有下降。这些内联样式绕过了 token 体系，是最大的「不可治理面」，建议另立专项逐步迁移到 CSS class + token。
- [ ] **导出类文件独立 token 组**：ShareCard / PrintLayout 是浅色主题（暗色 token 不适用），本次已全程排除。建议评估是否为它们单独建一组浅色 token（如 `--export-*`），而不是散落硬编码。
- [ ] **合并局部字号变量与全局**：pages/CompatibilityResultPage.css:10-15 自建了一套局部 `--fs-*`（`--fs-section-kicker:11px / --fs-section-title:22px / --fs-subsection-title:15px / --fs-body:14px / --fs-caption:12px` 等），其中 `--fs-body:14px ≠` 全局 `--font-size-body:15px`，且被 **5 个合婚子组件**复用（DayPillarPortrait / EvidenceDrawer / SectionVerdict / NextStepsAndAvoid / ScoreOverview）。两套字号体系并存，建议择机合并为一套（Task 2 质量审查发现）。

---

## 附录：一值多变量的语义选择记录（Task 1 已执行替换的选择依据）

当一个色值在 token 表里能匹配到多个变量时，Task 1 按**语义**而非字面选择，记录如下供回溯：

- `#4caf7d` → `--status-success`（喜神 = 正面状态，未选 `--wu-mu`）
- `#e05c4b` → `--status-danger`（忌神 = 负面状态，未选 `--wu-huo`）
- `#C9A84C` → `--wu-jin`（神煞吉点 / 徽章 = 品牌金，未选 `--primary` / `--text-accent`）
- `rgba(255,255,255,0.1)` fallback → `var(--border-default)`（对齐 YongshenBadge.css:133 既有写法）
