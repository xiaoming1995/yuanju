# 合盘「夫妻宫匹配」前端渲染设计文档

日期：2026-06-06
状态：已通过 brainstorming 评审，待用户复核

## 1. 目标

后端已产出并入库 `spouse_palace_match` 字段（见 `2026-06-06-spouse-palace-match-design.md`），但前端不渲染它，用户在合盘报告页看不到。本次给前端加一块「夫妻宫匹配」卡片，把这份数据展示出来。

**纯前端展示**，不动后端、不动其他区块、不加 i18n（项目本就硬编码中文）。

## 2. 数据形状（后端已定，前端只读）

`latest_report.content_structured` 已在后端解析为对象返回，前端直接拿对象（无需再 `JSON.parse`）。新增字段：

```ts
spouse_palace_match?: {
  self:    SpousePalaceSide   // 用 A 的夫妻宫推理想另一半，拿 B 真人比
  partner: SpousePalaceSide   // 反方向
  summary: string             // 一句话总括双向
} | null

interface SpousePalaceSide {
  ideal_portrait: string        // 该方命里理想/容易吸引的另一半画像（段落）
  match_level: 'high' | 'medium' | 'low' | ''  // 契合度；缺性别时可能为空
  fit_points: string[]          // 对方哪里对上了（纯字符串）
  gap_points: string[]          // 对方哪里有差距（纯字符串）
  evidence_keys: string[]       // 内部证据键，前端不展示
}
```

注意：`fit_points`/`gap_points` 是**纯字符串数组**，与 `personality_comparison` 的 `{title, detail}` 结构不同，因此**不能复用** `PersonalityComparison` 组件，需新建。

## 3. 关键决策（brainstorming 已确认）

| 决策点 | 结论 |
|---|---|
| 位置 | 放在「双方性格画像与差异」(`PersonalityComparison`) 卡片**正下方** |
| 布局 | 左右两列并排（A 一列 / B 一列），同现有 portrait-grid 风格；窄屏自动叠成上下 |
| 契合度文案 | 高 / 中 / 低 |
| 契合度配色 | 复用 duration 那套：高=绿(`--wu-mu`)、中=金(`--wu-jin`)、低=红(`--wu-huo`) |
| evidence_keys | **不展示**（内部键，用户看着是乱码） |
| 空值兜底 | 见 §5 |
| 后端 / 其他区块 | 不动 |

## 4. 架构与改动点

### 4.1 类型（`frontend/src/lib/api.ts`）

- 新增 `CompatibilitySpousePalaceSide`、`CompatibilitySpousePalaceMatch` 接口（形状见 §2）。
- `CompatibilityStructuredReport` 加可选字段：
  ```ts
  spouse_palace_match?: CompatibilitySpousePalaceMatch | null
  ```
  与现有 `personality_comparison?: ... | null` 同款（snake_case，可选）。

### 4.2 新组件（`frontend/src/components/compatibility/deep-analysis/SpousePalaceMatch.tsx` + `.css`）

- 入参：`{ match?: CompatibilitySpousePalaceMatch | null }`。
- 顶部标题：`夫妻宫匹配`（`serif compatibility-report-title`，与其他 section 标题一致）。
- 两列网格（复用 `compatibility-portrait-grid` 同款 2 列 + 窄屏断点）：每列一张 side 卡片。
- side 卡片内容（A 列 / B 列对称）：
  - 头部：`{名字}理想的另一半` + 契合度徽章（高/中/低，配色见 §3）。
  - `ideal_portrait` 段落。
  - `fit_points`：标「对上了」，逐条列出（✓ 前缀）。
  - `gap_points`：标「有差距」，逐条列出（✗ 前缀）。
- 底部：`summary` 一句话（整块通栏，置于两列之下）。
- 名字来源：A=self 列、B=partner 列。展示名取报告/详情里已有的 self/partner display name（`CompatibilityResultPage` 已能拿到参与者名；若组件层拿不到，则用「A / B」或「你 / 对方」兜底——实现时按页面已有数据决定，优先用真实展示名）。

徽章 label/class 映射（与现有 duration 约定对齐）：
```ts
const matchLevelText = { high: '高', medium: '中', low: '低' }
const matchLevelClass = {
  high: 'spouse-match-badge--high',
  medium: 'spouse-match-badge--medium',
  low: 'spouse-match-badge--low',
}
```
CSS 配色沿用 `ActionPlan7d30d.css` 的绿/金/红（high→`--wu-mu`、medium→`--wu-jin`、low→`--wu-huo`），底色用对应 `rgba(...,0.12)`。

### 4.3 挂载（`frontend/src/components/compatibility/deep-analysis/DeepReportNarrative.tsx`）

- import 新组件。
- 在 `<PersonalityComparison .../>` 渲染**之后**插入 `<SpousePalaceMatch match={structuredReport.spouse_palace_match} />`。

## 5. 空值与边界

- `spouse_palace_match` 为 `null`/`undefined`（旧报告）→ 组件返回 `null`，整块不渲染（对齐 `PersonalityComparison` 的 `if (!comparison) return null` 写法）。
- `self` 与 `partner` 都为空 → 同样不渲染。
- 某一侧 `match_level` 为空串（缺性别分支）→ 该侧**不显示徽章**，只显示 `ideal_portrait`（其文案后端已写明「缺性别，无法定配偶星」），`fit_points`/`gap_points` 可能为空数组 → 对应列表不渲染。
- `fit_points`/`gap_points` 为空数组 → 不渲染该小标题。

## 6. 测试

前端项目的测试约定需先确认（README/package.json 里的 test 脚本）。按现有约定：
- 若有组件测试框架（如 vitest + testing-library）：为 `SpousePalaceMatch` 写渲染测试——
  - 有完整数据 → 渲染标题、两列、两个徽章（高/低）、画像文字、fit/gap 条目、summary。
  - `match=null` → 渲染为空（组件返回 null）。
  - 某侧 `match_level=''` → 该侧无徽章但有画像。
- 若前端无组件测试基建：以 `tsc` 类型检查 + `npm run build` 通过为验证底线，并在实现计划里说明无单测的取舍。

## 7. 成功标准

1. 前端类型检查 / 构建通过（`tsc` + build）。
2. 打开一份**含** `spouse_palace_match` 的报告（如 reading `5a6073e0-547b-4df2-aebc-b53670a49963`）→ 性格画像下方出现「夫妻宫匹配」卡片，双列画像 + 高/中/低徽章 + summary 正常显示。
3. 打开一份**旧报告**（无该字段）→ 不出现该卡片，页面无报错。
4. 窄屏（手机宽度）下两列叠成上下，不溢出。

## 8. 明确不做（YAGNI）

- 不改后端、不动其他报告区块。
- 不展示 `evidence_keys`。
- 不加 i18n / 多语言。
- 不做单独的「双向总契合档」（数据只有两个方向各自的 match_level，不臆造合并值）。
