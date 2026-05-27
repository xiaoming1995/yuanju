# Compatibility Scoring v3.1 — 阈值放宽（双生 / 相生 加分）设计

## 背景

v3 评分使用「传统命理 100 分制」四模块加分式：合属相 50 + 合纳音 20 + 合日柱 10 + 合八字 20。
- 合属相、合日柱、合八字 三个模块以「六合 / 三合」为单一闸门
- 合纳音 以「相生 / 相同」为闸门
- 阈值非常严格：地支两两不同组合 66 对中，仅 18 对（27%）能拿合属相满分

实测两八字 (1996-02-08 20时) vs (1995-02-02 16时) 得 0/100，所有模块零分：

| 模块 | 输入 | 命中条件 | 结果 |
|---|---|---|---|
| 合属相 | 年支 子 vs 戌 | 六合/三合 | 不命中 → 0 |
| 合纳音 | 涧下水 vs 山头火 | 相生/相同 | 水克火 → 0 |
| 合日柱 | 日支 亥 vs 子 | 六合/三合 | 亥子双生水但非六合三合 → 0 |
| 合八字 | 年/月/时三柱均不合 | 六合/三合 | 全部 → 0 |

其中日支 亥-子 同水「双生」、月支 寅-丑（土生木被克但反向：丑土被寅木克）等关系在传统命理里并非「无关」，但 v3 spec line 30-32 明文写「双生即便五行同也是 0」。

结果：v3 在实际使用中过于稀疏，大量真实八字对落到「0 分 + 空 evidence 列表」，用户难以理解。

## 目标

- 把「五行相同（双生）」与「五行相生」纳入合属相、合日柱、合八字 三个模块的中/低档加分
- 保持四模块加分制框架与各模块分值上限不变（50 / 20 / 10 / 20）
- 不动数据库 schema、不影响 v3 旧记录、不破坏 v1/v2 历史渲染
- 升级 `analysis_version`: `"v3"` → `"v3.1"`
- 评分函数与 spec 同步修订，保证规约与实现一致

## 非目标

- 不调整合纳音模块（已含相生/相同/相克全覆盖）
- 不引入负分、不引入相冲/相穿/相害的惩罚（保持纯加分制）
- 不更换归一化公式
- 不重算 v3 历史记录
- 不影响 v1/v2 渲染路径

## 评分规则变更

### 合属相 (满分 50)

| 等级 | 命中条件 | 得分 |
|---|---|---|
| 上档 | 六合 (子丑/寅亥/卯戌/辰酉/巳申/午未) **或** 三合（含半三合：申子辰/亥卯未/巳酉丑/寅午戌 任两支） | 50 |
| 中档 | 五行相同（双生：亥子水/寅卯木/巳午火/申酉金/辰戌丑未土两两） | 30 |
| 下档 | 五行相生（一支生另一支，例：子→寅 水生木） | 20 |
| 0    | 相克 / 相冲 / 相穿 / 相害 / 自刑 等其他情形 | 0 |

**判定优先级**：上档 > 中档 > 下档 > 0。同支（自刑/重复）按现行 spec 仍返回 0。

**冲/穿/害与五行同的冲突处理**：辰戌、丑未 既是五行同土（中档候选），也是六冲；按本规则它们判为**中档 30**，相冲不扣分。这与 v3「no negative contributions」原则保持一致：负面关系一律不扣分，正面关系择最高档计分。同样：寅申（金克木）既是冲也是克 → 0；巳亥（水克火）既是冲也是克 → 0；卯酉（金克木）→ 0；子午（水克火）→ 0。

### 合纳音 (满分 20)

**不变**。已实现「相生 / 相同 → 20，相克 / 无关 → 0」。

### 合日柱 (满分 10)

| 等级 | 命中条件 | 得分 |
|---|---|---|
| 上档 | 日支「六合/三合」 **且** 日干「五合 或 五行相生」 | 10 |
| 中档 | 日支「六合/三合」（日干同/克/无关均可） | 5 |
| 下档 | 日支「五行同 或 五行相生」（日干任意） | 3 |
| 0    | 日支相克 / 相冲 / 相穿 / 无关 | 0 |

**判定优先级**：上档 > 中档 > 下档 > 0。

### 合八字 (满分 20)

- 三柱（年柱、月柱、时柱）各按上述「合日柱」规则得 0 / 3 / 5 / 10
- 三柱之和 sum ∈ [0, 30]
- 归一化函数 `(sum × 2 + 1) / 3` **不变**（整数除法；sum=0→0；sum=3→2；sum=15→10；sum=30→20）

## 实现架构

### 1. helpers（compatibility_scoring.go 新增）

```go
// 复用 event_signals.go 已有的 zhiWuxing / wxSheng

// 返回 true 当两支非空、非同支、五行相同
func branchSameElement(a, b string) bool

// 返回 true 当两支非空、非同支，且 a 与 b 的五行存在 生 关系（任一方向）
func branchShengElement(a, b string) bool
```

### 2. 修改 `scoreZodiac`

```go
func scoreZodiac(yearZhiA, yearZhiB string) int {
    if branchCompatible(yearZhiA, yearZhiB) {
        return 50
    }
    if branchSameElement(yearZhiA, yearZhiB) {
        return 30
    }
    if branchShengElement(yearZhiA, yearZhiB) {
        return 20
    }
    return 0
}
```

### 3. 修改 `scoreDayPillar`

```go
func scoreDayPillar(dayGanA, dayZhiA, dayGanB, dayZhiB string) int {
    if branchCompatible(dayZhiA, dayZhiB) {
        if ganUpperTier(dayGanA, dayGanB) {
            return 10
        }
        return 5
    }
    if branchSameElement(dayZhiA, dayZhiB) || branchShengElement(dayZhiA, dayZhiB) {
        return 3
    }
    return 0
}
```

注：下档 3 分**不区分**干关系，只看支。理由：下档已是"安慰分"，再细分意义不大。

### 4. `scoreEightChars`

无需修改函数体，调用更新后的 `scoreDayPillar` 即可。归一化保持 `(sum × 2 + 1) / 3`。

### 5. Evidence (compatibility_evidence.go)

evidences 列表最多仍 6 条（zodiac/nayin/day_pillar + 3 个 eight_chars）。新增两类 evidence kind：
- `branch_same_element`（双生）：文案「双方年支同属<水/木/火/金/土>，五行相同有亲近感」
- `branch_sheng_element`（相生）：文案「一方<水/木/...>生另一方<木/火/...>，构成五行相生」

合日柱、合八字的下档 evidence 同理生成。

### 6. 版本号

- `backend/internal/service/compatibility_service.go:16` 改为 `const compatibilityAnalysisVersion = "v3.1"`
- `frontend/src/lib/api.ts` 类型联合扩展加入 `'v3.1'`，schema 与 v3 完全兼容
- `frontend/src/pages/CompatibilityHistoryPage.tsx` 历史页渲染分支：`'v3' | 'v3.1'` 共用 ScoreOverviewV3 路径
- `ScoreOverviewV3.tsx` 文案保持泛化（不写死"v3"），不需要新组件

### 7. DB / migration

- 无 schema 变化（`overall_score INTEGER` 列已存在）
- 旧 v3 行**不重算**，`analysis_version` 列保留原值

### 8. AI Prompt (`backend/pkg/prompt/canonical_compatibility.go`)

需检查 prompt 是否包含 v3 评分公式描述。若包含：
- 同步加入 v3.1 三级规则说明
- 修正"双生不计分"等过时表述

若 prompt 仅输出四模块结构化字段（`zodiac_score / nayin_score / day_pillar_score / eight_chars_score`）而不解释算法，则不动。

### 9. Spec 文档

新增 openspec change 文件夹：`openspec/changes/compatibility-scoring-formula-v3-1-relax/`，包含：
- `proposal.md` — 修订理由 + 本案诊断
- `design.md` — 摘自本文档
- `tasks.md` — 实现任务清单
- `specs/compatibility-scoring-formula/spec.md` — **MODIFIED** Requirements

MODIFIED Requirements 涉及的 spec sections：
- **Requirement: Zodiac module** — 改为三级 50/30/20/0；删除"No hit including 双生..."条款
- **Requirement: Day pillar module** — 加入下档 3
- **Requirement: Eight-chars module** — sum 范围说明更新（仍为 0..30，仅 per-pillar 可能取值集变更）
- **Requirement: Analysis version tag** — `"v3"` → `"v3.1"`；说明 v3 旧记录保留

保留 v3 历史 spec，不删除原 change folder。

## 测试

### compatibility_scoring_test.go 新增

```go
// 三级 zodiac
TestScoreZodiac_LiuheSanhe   → 50  (子丑 / 申子)
TestScoreZodiac_SameElement  → 30  (亥子 / 寅卯 / 巳午 / 申酉 / 辰戌 / 丑未)
TestScoreZodiac_Sheng        → 20  (子寅 水生木 / 寅巳 木生火 等)
TestScoreZodiac_Ke           → 0   (子未 水土相克 / 子戌 同上)
TestScoreZodiac_Chong        → 0   (子午 / 寅申)  // 冲且克 → 0
TestScoreZodiac_Chong_SameElement → 30 (辰戌 / 丑未)  // 冲但同行 → 中档 30

// 日柱下档
TestScoreDayPillar_LowerTier_SameElement  → 3 (日支亥子、干任意)
TestScoreDayPillar_LowerTier_Sheng        → 3 (日支子寅、干任意)
TestScoreDayPillar_Ke                     → 0

// 八字归一化
TestScoreEightChars_SinglePillarLowerTier → sum=3 → 2
TestScoreEightChars_MixedTiers            → sum=18 (10+5+3) → 12

// 本案回归
TestRealCase_1996_1995_Yields_5_of_100
  A: 丙子 / 庚寅 / 乙亥 / 丙戌
  B: 甲戌 / 丁丑 / 甲子 / 壬申
  期望: zodiac=0, nayin=0, day_pillar=3, eight_chars=2, total=5
```

### 旧测试用例

检查现有 `compatibility_scoring_test.go` 是否有依赖「双生 = 0」「相生 = 0」的断言；如有，明确改为新值并在测试名标注 v3.1。

## 影响范围

| 文件 | 变更类型 |
|---|---|
| `backend/pkg/bazi/compatibility_scoring.go` | 修改：scoreZodiac、scoreDayPillar；新增 branchSameElement、branchShengElement |
| `backend/pkg/bazi/compatibility_scoring_test.go` | 新增/修改测试用例 |
| `backend/pkg/bazi/compatibility_evidence.go` | 新增两类 evidence kind 与文案 |
| `backend/pkg/bazi/compatibility_assessment.go` | 检查是否引用「双生 = 0」描述；如有则同步 |
| `backend/internal/service/compatibility_service.go` | 版本常量 v3 → v3.1 |
| `backend/pkg/prompt/canonical_compatibility.go` | 检查算法描述；如有，同步 v3.1 规则 |
| `frontend/src/lib/api.ts` | 联合类型加入 'v3.1' |
| `frontend/src/pages/CompatibilityHistoryPage.tsx` | version 分支同时识别 v3 / v3.1 |
| `frontend/src/pages/CompatibilityResultPage.tsx` | 同上（如果按 version 切组件） |
| `openspec/changes/compatibility-scoring-formula-v3-1-relax/*` | 新增 openspec change |

**未涉及**：
- migrations（无 schema 变化）
- ScoreOverviewV3 组件
- v1/v2 旧渲染路径
- 报告内容（年运、十神等其他算法）

## 风险

1. **spec 与实现不同步**：v3 spec 显式说「双生 = 0」，必须同步修改。在 PR 检查清单里列入。
2. **AI prompt 引用算法描述**：若 prompt 写死了「双生不计分」，新算法输出会与 AI 解读矛盾。需检查并同步。
3. **整数归一化粒度**：sum=3 → `(7)/3 = 2`（int 除法），用户可能期望 2.33。可接受（v3 原本就用这个公式）。
4. **前端版本号识别**：`'v3.1'` 包含点号，需确认 TypeScript 联合类型与历史页 switch 都按字面字符串匹配。

## 验收标准

1. 本案 (1996-02-08 20时) vs (1995-02-02 16时) 评分：
   - zodiac = 0（子戌 土克水）
   - nayin = 0（水火相克）
   - day_pillar = 3（亥子 双生）
   - eight_chars = 2（时支 戌申 土生金 → sum=3 归一化 2）
   - **total = 5/100**
2. 旧 v3 记录（analysis_version='v3'）从历史页打开后渲染与改动前一致
3. 新生成报告 analysis_version 写入 'v3.1'
4. compatibility_scoring_test.go 全部通过
5. spec 文档 `even when 五行相生/同 (双生) applies` 字样已移除
