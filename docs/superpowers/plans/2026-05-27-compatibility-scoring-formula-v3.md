# 合盘评分公式 v3 实现 Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 把合盘评分引擎从「11 信号源 × 4 维度 × evidence 加权 + 来源贡献封顶」整体替换为「合属相 50 + 合纳音 20 + 合日柱 10 + 合八字 20 = 100 分」的纯加分制，4 维度字段从 `attraction/stability/communication/practicality` 改名为 `zodiac/nayin/day_pillar/eight_chars`，`analysis_version` 由 `"v2"` 升至 `"v3"`，旧记录保留并通过 frontend version 分支兼容渲染。

**Architecture:** 后端 `backend/pkg/bazi/compatibility.go`（现 1510 行）拆为 5 个文件（主入口 + 评分 + 纳音 + evidence + 咨询评估），每文件 ≤ 300 行；service / repository 字段同步；frontend 类型 + `CompatibilityResultPage` / `CompatibilityHistoryPage` 按 version 分两路渲染。AI prompt 模板措辞与 4 模块语境对齐。OpenSpec 新建 change 目录并把旧两份 spec 归档。

**Tech Stack:** Go 1.x（`testing` 标准库），PostgreSQL（JSONB 字段不需 DDL 改），React 18 + TypeScript + Vite（纯 CSS Variables），`github.com/6tail/lunar-go`（已提供纳音字符串）。

**Spec 来源：** [`docs/superpowers/specs/2026-05-27-compatibility-scoring-formula-v2-design.md`](../specs/2026-05-27-compatibility-scoring-formula-v2-design.md)

---

## 文件结构

**Backend - 替换/新增**

| 路径 | 角色 | 行数预估 |
|---|---|---|
| `backend/pkg/bazi/compatibility.go` | 公开类型 + `AnalyzeCompatibility` 主入口（旧文件清空重写） | ~150 |
| `backend/pkg/bazi/compatibility_nayin.go` | 六十甲子纳音五行查表 | ~100 |
| `backend/pkg/bazi/compatibility_scoring.go` | 4 模块打分函数 + 支合判定 | ~280 |
| `backend/pkg/bazi/compatibility_evidence.go` | 命中→evidence / score_explanations / summary_tags | ~200 |
| `backend/pkg/bazi/compatibility_assessment.go` | relationshipType / decision_advice / duration / stage_risks / strategy / claim_evidence_links | ~280 |
| `backend/pkg/bazi/compatibility_test.go` | 重写（保留 `makeCompatNatal` 风格） | ~450 |
| `backend/pkg/bazi/compatibility_nayin_test.go` | 60 甲子纳音查表测试 | ~80 |
| `backend/pkg/bazi/compatibility_scoring_test.go` | 4 模块单元测试 | ~250 |
| `backend/internal/model/compatibility.go` | `CompatibilityDimensionScores` 字段重命名 + `CompatibilityReading.OverallScore` 新增 | 改 ~10 行 |
| `backend/internal/repository/compatibility_repository.go` | 无逻辑变动（version 由 service 传入） | 0 |
| `backend/internal/service/compatibility_service.go` | `compatibilityAnalysisVersion = "v3"` + DimensionScores 映射改字段名 + `OverallScore` 透传 | 改 ~15 行 |
| `backend/pkg/prompt/canonical_compatibility.go` | prompt 模板措辞与 4 模块对齐 | 改 ~30 行 |
| `backend/pkg/database/migrations/00011_compatibility_v3_analysis.sql` | 新增（仅 COMMENT） | ~10 |

**Frontend - 修改**

| 路径 | 角色 |
|---|---|
| `frontend/src/lib/api.ts` | `CompatibilityDimensionScores` 类型重命名（联合 v1/v2/v3）+ `overall_score` 新增 |
| `frontend/src/pages/CompatibilityResultPage.tsx` | 顶层 `version === 'v3'` 分支渲染新 `ScoreOverviewV3`；legacy 路径不动 |
| `frontend/src/pages/CompatibilityHistoryPage.tsx` | 列表卡片按 version 切两路 |

**OpenSpec - 新增**

| 路径 |
|---|
| `openspec/changes/compatibility-scoring-formula-v3/proposal.md` |
| `openspec/changes/compatibility-scoring-formula-v3/design.md` |
| `openspec/changes/compatibility-scoring-formula-v3/tasks.md` |
| `openspec/changes/compatibility-scoring-formula-v3/specs/compatibility-scoring-formula/spec.md` |

---

## 任务总览（25 个）

```
Phase A: Backend 算法核心（拆分 + TDD）          Task 1–14
Phase B: 类型 + Service 层迁移                   Task 15–17
Phase C: Migration + AI Prompt                   Task 18–19
Phase D: Frontend version 分支                    Task 20–23
Phase E: OpenSpec change                         Task 24–25
```

---

## Phase A：Backend 算法核心

### Task 1: 纳音表与查表函数

**Files:**
- Create: `backend/pkg/bazi/compatibility_nayin.go`
- Create: `backend/pkg/bazi/compatibility_nayin_test.go`

- [ ] **Step 1: 写失败测试**

把 60 甲子的纳音五行映射全表覆盖（每条 ganzhi → 五行 单字符）：

```go
package bazi

import "testing"

func TestNayinElement_AllSixtyGanzhi(t *testing.T) {
	cases := []struct {
		ganzhi string
		want   string
	}{
		{"甲子", "金"}, {"乙丑", "金"}, // 海中金
		{"丙寅", "火"}, {"丁卯", "火"}, // 炉中火
		{"戊辰", "木"}, {"己巳", "木"}, // 大林木
		{"庚午", "土"}, {"辛未", "土"}, // 路旁土
		{"壬申", "金"}, {"癸酉", "金"}, // 剑锋金
		{"甲戌", "火"}, {"乙亥", "火"}, // 山头火
		{"丙子", "水"}, {"丁丑", "水"}, // 涧下水
		{"戊寅", "土"}, {"己卯", "土"}, // 城头土
		{"庚辰", "金"}, {"辛巳", "金"}, // 白蜡金
		{"壬午", "木"}, {"癸未", "木"}, // 杨柳木
		{"甲申", "水"}, {"乙酉", "水"}, // 泉中水
		{"丙戌", "土"}, {"丁亥", "土"}, // 屋上土
		{"戊子", "火"}, {"己丑", "火"}, // 霹雳火
		{"庚寅", "木"}, {"辛卯", "木"}, // 松柏木
		{"壬辰", "水"}, {"癸巳", "水"}, // 长流水
		{"甲午", "金"}, {"乙未", "金"}, // 沙中金
		{"丙申", "火"}, {"丁酉", "火"}, // 山下火
		{"戊戌", "木"}, {"己亥", "木"}, // 平地木
		{"庚子", "土"}, {"辛丑", "土"}, // 壁上土
		{"壬寅", "金"}, {"癸卯", "金"}, // 金箔金
		{"甲辰", "火"}, {"乙巳", "火"}, // 覆灯火
		{"丙午", "水"}, {"丁未", "水"}, // 天河水
		{"戊申", "土"}, {"己酉", "土"}, // 大驿土
		{"庚戌", "金"}, {"辛亥", "金"}, // 钗钏金
		{"壬子", "木"}, {"癸丑", "木"}, // 桑柘木
		{"甲寅", "水"}, {"乙卯", "水"}, // 大溪水
		{"丙辰", "土"}, {"丁巳", "土"}, // 沙中土
		{"戊午", "火"}, {"己未", "火"}, // 天上火
		{"庚申", "木"}, {"辛酉", "木"}, // 石榴木
		{"壬戌", "水"}, {"癸亥", "水"}, // 大海水
	}
	for _, tc := range cases {
		got := nayinElement(tc.ganzhi)
		if got != tc.want {
			t.Errorf("nayinElement(%q) = %q, want %q", tc.ganzhi, got, tc.want)
		}
	}
}

func TestNayinElement_Unknown_ReturnsEmpty(t *testing.T) {
	if got := nayinElement(""); got != "" {
		t.Errorf("empty input: got %q, want \"\"", got)
	}
	if got := nayinElement("XX"); got != "" {
		t.Errorf("unknown ganzhi: got %q, want \"\"", got)
	}
}
```

- [ ] **Step 2: 运行测试确认失败**

Run: `cd backend && go test ./pkg/bazi/ -run TestNayinElement -v`
Expected: FAIL with `undefined: nayinElement`

- [ ] **Step 3: 实现 `compatibility_nayin.go`**

```go
package bazi

// 六十甲子纳音五行表。Key 为天干地支组合（如 "甲子"），Value 为纳音对应的五行单字符。
// 数据来源：传统《六十甲子纳音表》，详情见 docs/superpowers/specs/2026-05-27-compatibility-scoring-formula-v2-design.md §3.2。
var nayinElementTable = map[string]string{
	"甲子": "金", "乙丑": "金", // 海中金
	"丙寅": "火", "丁卯": "火", // 炉中火
	"戊辰": "木", "己巳": "木", // 大林木
	"庚午": "土", "辛未": "土", // 路旁土
	"壬申": "金", "癸酉": "金", // 剑锋金
	"甲戌": "火", "乙亥": "火", // 山头火
	"丙子": "水", "丁丑": "水", // 涧下水
	"戊寅": "土", "己卯": "土", // 城头土
	"庚辰": "金", "辛巳": "金", // 白蜡金
	"壬午": "木", "癸未": "木", // 杨柳木
	"甲申": "水", "乙酉": "水", // 泉中水
	"丙戌": "土", "丁亥": "土", // 屋上土
	"戊子": "火", "己丑": "火", // 霹雳火
	"庚寅": "木", "辛卯": "木", // 松柏木
	"壬辰": "水", "癸巳": "水", // 长流水
	"甲午": "金", "乙未": "金", // 沙中金
	"丙申": "火", "丁酉": "火", // 山下火
	"戊戌": "木", "己亥": "木", // 平地木
	"庚子": "土", "辛丑": "土", // 壁上土
	"壬寅": "金", "癸卯": "金", // 金箔金
	"甲辰": "火", "乙巳": "火", // 覆灯火
	"丙午": "水", "丁未": "水", // 天河水
	"戊申": "土", "己酉": "土", // 大驿土
	"庚戌": "金", "辛亥": "金", // 钗钏金
	"壬子": "木", "癸丑": "木", // 桑柘木
	"甲寅": "水", "乙卯": "水", // 大溪水
	"丙辰": "土", "丁巳": "土", // 沙中土
	"戊午": "火", "己未": "火", // 天上火
	"庚申": "木", "辛酉": "木", // 石榴木
	"壬戌": "水", "癸亥": "水", // 大海水
}

// nayinElement 返回 ganzhi（如 "甲子"）对应的纳音五行单字符（如 "金"）。
// 输入不在 60 甲子表中时返回空字符串。
func nayinElement(ganzhi string) string {
	return nayinElementTable[ganzhi]
}

// nayinRelation 描述两个纳音五行之间的关系：
//   "sheng" - 相生
//   "same"  - 相同
//   "ke"    - 相克
//   ""      - 输入有空字符串或未知
func nayinRelation(a, b string) string {
	if a == "" || b == "" {
		return ""
	}
	if a == b {
		return "same"
	}
	if wxSheng[a] == b || wxSheng[b] == a {
		return "sheng"
	}
	if wxKe[a] == b || wxKe[b] == a {
		return "ke"
	}
	return ""
}
```

- [ ] **Step 4: 运行测试确认通过**

Run: `cd backend && go test ./pkg/bazi/ -run TestNayinElement -v`
Expected: PASS（60 项全部命中）

- [ ] **Step 5: 加 `nayinRelation` 测试并跑通**

追加到 `compatibility_nayin_test.go`：

```go
func TestNayinRelation_Cases(t *testing.T) {
	cases := []struct {
		a, b string
		want string
	}{
		{"金", "水", "sheng"}, // 金生水
		{"水", "金", "sheng"}, // 反向
		{"火", "土", "sheng"}, // 火生土
		{"金", "金", "same"},
		{"金", "木", "ke"}, // 金克木
		{"木", "金", "ke"},
		{"", "金", ""},
		{"金", "", ""},
	}
	for _, tc := range cases {
		if got := nayinRelation(tc.a, tc.b); got != tc.want {
			t.Errorf("nayinRelation(%q,%q) = %q, want %q", tc.a, tc.b, got, tc.want)
		}
	}
}
```

Run: `cd backend && go test ./pkg/bazi/ -run TestNayin -v`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add backend/pkg/bazi/compatibility_nayin.go backend/pkg/bazi/compatibility_nayin_test.go
git commit -m "feat(compatibility): add nayin element lookup table

60-ganzhi → 五行 mapping plus nayinRelation helper for the v3 scoring formula.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>"
```

---

### Task 2: 支合共享判定函数 `branchCompatible`

**Files:**
- Create: `backend/pkg/bazi/compatibility_scoring.go`
- Create: `backend/pkg/bazi/compatibility_scoring_test.go`

- [ ] **Step 1: 写失败测试**

```go
package bazi

import "testing"

func TestBranchCompatible_Liuhe(t *testing.T) {
	pairs := [][2]string{
		{"子", "丑"}, {"寅", "亥"}, {"卯", "戌"},
		{"辰", "酉"}, {"巳", "申"}, {"午", "未"},
	}
	for _, p := range pairs {
		if !branchCompatible(p[0], p[1]) {
			t.Errorf("liuhe %s/%s should be compatible", p[0], p[1])
		}
		if !branchCompatible(p[1], p[0]) {
			t.Errorf("liuhe %s/%s (reverse) should be compatible", p[1], p[0])
		}
	}
}

func TestBranchCompatible_Sanhe(t *testing.T) {
	// 申子辰水局 — 三选二
	for _, p := range [][2]string{{"申", "子"}, {"子", "辰"}, {"申", "辰"}} {
		if !branchCompatible(p[0], p[1]) {
			t.Errorf("sanhe %s/%s should be compatible", p[0], p[1])
		}
	}
}

func TestBranchCompatible_SameBranch_NotCompatible(t *testing.T) {
	for _, b := range []string{"子", "丑", "寅", "卯", "辰", "巳", "午", "未", "申", "酉", "戌", "亥"} {
		if branchCompatible(b, b) {
			t.Errorf("same branch %s/%s should NOT be compatible (sanhe requires different)", b, b)
		}
	}
}

func TestBranchCompatible_ChongHaiXing_NotCompatible(t *testing.T) {
	cases := [][2]string{
		{"子", "午"}, {"丑", "未"}, {"寅", "申"}, // 六冲
		{"子", "未"}, {"丑", "午"}, // 六害
		{"子", "卯"}, // 相刑
	}
	for _, p := range cases {
		if branchCompatible(p[0], p[1]) {
			t.Errorf("chong/hai/xing %s/%s should NOT be compatible", p[0], p[1])
		}
	}
}

func TestBranchCompatible_Empty_NotCompatible(t *testing.T) {
	if branchCompatible("", "子") || branchCompatible("子", "") {
		t.Error("empty branch should not be compatible")
	}
}
```

- [ ] **Step 2: 运行测试确认失败**

Run: `cd backend && go test ./pkg/bazi/ -run TestBranchCompatible -v`
Expected: FAIL with `undefined: branchCompatible`

- [ ] **Step 3: 实现 `branchCompatible`**

把以下代码加到新创建的 `backend/pkg/bazi/compatibility_scoring.go`：

```go
package bazi

// branchCompatible 判定两个地支是否构成「支合」：
//   - 六合（子丑/寅亥/卯戌/辰酉/巳申/午未）
//   - 三合（申子辰/亥卯未/巳酉丑/寅午戌 中任选两支不相等者，即「半三合」）
// 同支返回 false（自刑/单纯重复均不算合）；空字符串返回 false。
func branchCompatible(a, b string) bool {
	if a == "" || b == "" || a == b {
		return false
	}
	if sixHe[a] == b {
		return true
	}
	for _, group := range sanheGroups {
		hasA, hasB := false, false
		for _, z := range group {
			if z == a {
				hasA = true
			}
			if z == b {
				hasB = true
			}
		}
		if hasA && hasB {
			return true
		}
	}
	return false
}

// branchCompatibilityKind 返回支合的类型，用于 evidence 文案区分：
//   "liuhe" - 六合
//   "sanhe" - 三合（半三合）
//   ""      - 不合
func branchCompatibilityKind(a, b string) string {
	if a == "" || b == "" || a == b {
		return ""
	}
	if sixHe[a] == b {
		return "liuhe"
	}
	for _, group := range sanheGroups {
		hasA, hasB := false, false
		for _, z := range group {
			if z == a {
				hasA = true
			}
			if z == b {
				hasB = true
			}
		}
		if hasA && hasB {
			return "sanhe"
		}
	}
	return ""
}

// sanheGroupName 返回包含 a 与 b 的三合局名（如 "申子辰" 水局），未命中返回空。
func sanheGroupName(a, b string) string {
	if a == "" || b == "" || a == b {
		return ""
	}
	for _, group := range sanheGroups {
		hasA, hasB := false, false
		for _, z := range group {
			if z == a {
				hasA = true
			}
			if z == b {
				hasB = true
			}
		}
		if hasA && hasB {
			return group[0] + group[1] + group[2]
		}
	}
	return ""
}
```

- [ ] **Step 4: 运行测试确认通过**

Run: `cd backend && go test ./pkg/bazi/ -run TestBranchCompatible -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add backend/pkg/bazi/compatibility_scoring.go backend/pkg/bazi/compatibility_scoring_test.go
git commit -m "feat(compatibility): add branchCompatible shared judgment

Centralize liuhe/sanhe(半) detection in branchCompatible plus kind/group helpers
for the v3 scoring formula.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>"
```

---

### Task 3: 合属相打分函数 `scoreZodiac`

**Files:**
- Modify: `backend/pkg/bazi/compatibility_scoring.go`
- Modify: `backend/pkg/bazi/compatibility_scoring_test.go`

- [ ] **Step 1: 写失败测试**

```go
func TestScoreZodiac_Liuhe_Returns50(t *testing.T) {
	if got := scoreZodiac("子", "丑"); got != 50 {
		t.Errorf("子丑 liuhe: got %d, want 50", got)
	}
}

func TestScoreZodiac_Sanhe_Returns50(t *testing.T) {
	if got := scoreZodiac("申", "子"); got != 50 {
		t.Errorf("申子 半三合: got %d, want 50", got)
	}
	if got := scoreZodiac("子", "辰"); got != 50 {
		t.Errorf("子辰 半三合: got %d, want 50", got)
	}
}

func TestScoreZodiac_NoHit_Returns0(t *testing.T) {
	cases := [][2]string{
		{"子", "午"}, // 六冲
		{"子", "未"}, // 六害
		{"子", "卯"}, // 相刑
		{"子", "子"}, // 同支（自刑）
		{"寅", "卯"}, // 双生（五行同），不命中
	}
	for _, p := range cases {
		if got := scoreZodiac(p[0], p[1]); got != 0 {
			t.Errorf("scoreZodiac(%q,%q) = %d, want 0", p[0], p[1], got)
		}
	}
}

func TestScoreZodiac_Empty_Returns0(t *testing.T) {
	if scoreZodiac("", "子") != 0 || scoreZodiac("子", "") != 0 {
		t.Error("empty branch should score 0")
	}
}
```

- [ ] **Step 2: 运行测试确认失败**

Run: `cd backend && go test ./pkg/bazi/ -run TestScoreZodiac -v`
Expected: FAIL with `undefined: scoreZodiac`

- [ ] **Step 3: 实现 `scoreZodiac`**

追加到 `backend/pkg/bazi/compatibility_scoring.go`：

```go
// scoreZodiac 计算「合属相」模块得分（满分 50）。
// 输入为两人年支；命中六合或三合（含半三合）即得 50，否则 0。
func scoreZodiac(yearZhiA, yearZhiB string) int {
	if branchCompatible(yearZhiA, yearZhiB) {
		return 50
	}
	return 0
}
```

- [ ] **Step 4: 运行测试确认通过**

Run: `cd backend && go test ./pkg/bazi/ -run TestScoreZodiac -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add backend/pkg/bazi/compatibility_scoring.go backend/pkg/bazi/compatibility_scoring_test.go
git commit -m "feat(compatibility): add scoreZodiac (year-zhi 50-point module)

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>"
```

---

### Task 4: 合纳音打分函数 `scoreNayin`

**Files:**
- Modify: `backend/pkg/bazi/compatibility_scoring.go`
- Modify: `backend/pkg/bazi/compatibility_scoring_test.go`

- [ ] **Step 1: 写失败测试**

```go
func TestScoreNayin_Sheng_Returns20(t *testing.T) {
	// 甲子(金) vs 壬午(木) — 金克木，不算分；改用 甲子(金) vs 庚午(土) — 土生金
	if got := scoreNayin("甲子", "庚午"); got != 20 {
		t.Errorf("甲子(金)+庚午(土) 相生: got %d, want 20", got)
	}
}

func TestScoreNayin_Same_Returns20(t *testing.T) {
	if got := scoreNayin("甲子", "乙丑"); got != 20 {
		t.Errorf("甲子(金)+乙丑(金) 同金: got %d, want 20", got)
	}
}

func TestScoreNayin_Ke_Returns0(t *testing.T) {
	// 甲子(金) vs 戊辰(木) — 金克木
	if got := scoreNayin("甲子", "戊辰"); got != 0 {
		t.Errorf("甲子(金)+戊辰(木) 相克: got %d, want 0", got)
	}
}

func TestScoreNayin_Empty_Returns0(t *testing.T) {
	if scoreNayin("", "甲子") != 0 || scoreNayin("甲子", "") != 0 || scoreNayin("XX", "YY") != 0 {
		t.Error("empty / unknown ganzhi should score 0")
	}
}
```

- [ ] **Step 2: 运行测试确认失败**

Run: `cd backend && go test ./pkg/bazi/ -run TestScoreNayin -v`
Expected: FAIL with `undefined: scoreNayin`

- [ ] **Step 3: 实现 `scoreNayin`**

```go
// scoreNayin 计算「合纳音」模块得分（满分 20）。
// 输入为两人年柱干支字符串（如 "甲子"）；
// 纳音五行 相生 或 相同 即得 20，相克或无法识别返回 0。
func scoreNayin(yearGanZhiA, yearGanZhiB string) int {
	wxA := nayinElement(yearGanZhiA)
	wxB := nayinElement(yearGanZhiB)
	switch nayinRelation(wxA, wxB) {
	case "sheng", "same":
		return 20
	default:
		return 0
	}
}
```

- [ ] **Step 4: 运行测试确认通过**

Run: `cd backend && go test ./pkg/bazi/ -run TestScoreNayin -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add backend/pkg/bazi/compatibility_scoring.go backend/pkg/bazi/compatibility_scoring_test.go
git commit -m "feat(compatibility): add scoreNayin (nayin 20-point module)

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>"
```

---

### Task 5: 合日柱打分函数 `scoreDayPillar`

**Files:**
- Modify: `backend/pkg/bazi/compatibility_scoring.go`
- Modify: `backend/pkg/bazi/compatibility_scoring_test.go`

- [ ] **Step 1: 写失败测试**

```go
func TestScoreDayPillar_UpperTier_GanHe(t *testing.T) {
	// 甲子 / 己丑：天干五合 甲己 + 地支六合 子丑 → 上档 10
	got := scoreDayPillar("甲", "子", "己", "丑")
	if got != 10 {
		t.Errorf("甲子/己丑 上档: got %d, want 10", got)
	}
}

func TestScoreDayPillar_UpperTier_GanSheng(t *testing.T) {
	// 甲子 / 丁丑：天干相生（甲木生丁火）+ 地支六合 → 上档 10
	got := scoreDayPillar("甲", "子", "丁", "丑")
	if got != 10 {
		t.Errorf("甲子/丁丑 上档: got %d, want 10", got)
	}
}

func TestScoreDayPillar_LowerTier_GanTong(t *testing.T) {
	// 甲子 / 乙丑：天干相同（甲乙木）+ 地支六合 → 下档 5
	got := scoreDayPillar("甲", "子", "乙", "丑")
	if got != 5 {
		t.Errorf("甲子/乙丑 下档(干同): got %d, want 5", got)
	}
}

func TestScoreDayPillar_LowerTier_GanKe(t *testing.T) {
	// 甲子 / 戊丑：天干相克（甲克戊）+ 地支六合 → 下档 5
	got := scoreDayPillar("甲", "子", "戊", "丑")
	if got != 5 {
		t.Errorf("甲子/戊丑 下档(干克): got %d, want 5", got)
	}
}

func TestScoreDayPillar_ZhiNotCompatible_Returns0(t *testing.T) {
	// 甲子 / 己未：干合甲己 但支不合（子未六害） → 0
	got := scoreDayPillar("甲", "子", "己", "未")
	if got != 0 {
		t.Errorf("甲子/己未 支不合: got %d, want 0", got)
	}
}

func TestScoreDayPillar_SanheZhi(t *testing.T) {
	// 甲子 / 己辰：干合 + 支为子辰半三合（水局） → 上档 10
	got := scoreDayPillar("甲", "子", "己", "辰")
	if got != 10 {
		t.Errorf("甲子/己辰 上档(三合支): got %d, want 10", got)
	}
}
```

- [ ] **Step 2: 运行测试确认失败**

Run: `cd backend && go test ./pkg/bazi/ -run TestScoreDayPillar -v`
Expected: FAIL with `undefined: scoreDayPillar`

- [ ] **Step 3: 实现 `scoreDayPillar`**

```go
// scoreDayPillar 计算「合日柱」模块得分（满分 10）。
// 上档 (10)：日支合 + (干五合 OR 干五行相生)
// 下档 (5) ：日支合 + (干五行相同 OR 干相克 OR 干无关)
// 不命中(0)：日支不合，无论干如何
func scoreDayPillar(dayGanA, dayZhiA, dayGanB, dayZhiB string) int {
	if !branchCompatible(dayZhiA, dayZhiB) {
		return 0
	}
	if ganUpperTier(dayGanA, dayGanB) {
		return 10
	}
	return 5
}

// ganUpperTier 判定两天干是否构成「合日柱上档」的强化条件：
//   - 天干五合（甲己/乙庚/丙辛/丁壬/戊癸）
//   - 天干五行相生
// 干相同 / 干相克 / 干无关 一律返回 false（落到下档 5 分）。
func ganUpperTier(a, b string) bool {
	if a == "" || b == "" {
		return false
	}
	if _, ok := ganWuhe[[2]string{a, b}]; ok {
		return true
	}
	wxA := ganWuxing[a]
	wxB := ganWuxing[b]
	if wxA == "" || wxB == "" {
		return false
	}
	if wxA == wxB {
		return false // 同行 → 下档（不算上档）
	}
	if wxSheng[wxA] == wxB || wxSheng[wxB] == wxA {
		return true
	}
	return false
}
```

- [ ] **Step 4: 运行测试确认通过**

Run: `cd backend && go test ./pkg/bazi/ -run TestScoreDayPillar -v`
Expected: PASS（6 个 case 全部）

- [ ] **Step 5: Commit**

```bash
git add backend/pkg/bazi/compatibility_scoring.go backend/pkg/bazi/compatibility_scoring_test.go
git commit -m "feat(compatibility): add scoreDayPillar with upper/lower tier

支合 prerequisite + gan upper-tier (五合/相生) → 10, otherwise → 5.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>"
```

---

### Task 6: 合八字打分函数 `scoreEightChars`

**Files:**
- Modify: `backend/pkg/bazi/compatibility_scoring.go`
- Modify: `backend/pkg/bazi/compatibility_scoring_test.go`

- [ ] **Step 1: 写失败测试**

```go
func TestScoreEightChars_AllUpper_Returns20(t *testing.T) {
	// 三柱全上档 → sum=30 → (60+1)/3=20
	got := scoreEightChars(
		"甲", "子", "己", "丑", // 年：上档
		"甲", "子", "己", "丑", // 月：上档
		"甲", "子", "己", "丑", // 时：上档
	)
	if got != 20 {
		t.Errorf("三柱全上档: got %d, want 20", got)
	}
}

func TestScoreEightChars_AllLower_Returns10(t *testing.T) {
	// 三柱全下档 (5+5+5=15) → (30+1)/3=10
	got := scoreEightChars(
		"甲", "子", "乙", "丑",
		"甲", "子", "乙", "丑",
		"甲", "子", "乙", "丑",
	)
	if got != 10 {
		t.Errorf("三柱全下档: got %d, want 10", got)
	}
}

func TestScoreEightChars_OneUpperOnly_Returns7(t *testing.T) {
	// 一柱上档 + 两柱不合 (10+0+0=10) → (20+1)/3=7
	got := scoreEightChars(
		"甲", "子", "己", "丑", // 年：上档
		"甲", "午", "甲", "午", // 月：同支不合
		"甲", "午", "甲", "午", // 时：同支不合
	)
	if got != 7 {
		t.Errorf("一柱上档 其余不合: got %d, want 7", got)
	}
}

func TestScoreEightChars_NothingHits_Returns0(t *testing.T) {
	got := scoreEightChars(
		"甲", "午", "甲", "午",
		"甲", "午", "甲", "午",
		"甲", "午", "甲", "午",
	)
	if got != 0 {
		t.Errorf("全不命中: got %d, want 0", got)
	}
}

func TestScoreEightChars_RoundingTable(t *testing.T) {
	cases := []struct {
		sum  int
		want int
	}{{0, 0}, {5, 3}, {10, 7}, {15, 10}, {20, 13}, {25, 17}, {30, 20}}
	for _, tc := range cases {
		if got := normalizeEightCharsSum(tc.sum); got != tc.want {
			t.Errorf("normalizeEightCharsSum(%d) = %d, want %d", tc.sum, got, tc.want)
		}
	}
}
```

- [ ] **Step 2: 运行测试确认失败**

Run: `cd backend && go test ./pkg/bazi/ -run TestScoreEightChars -v`
Expected: FAIL with `undefined: scoreEightChars`

- [ ] **Step 3: 实现 `scoreEightChars` 与归一化函数**

```go
// scoreEightChars 计算「合八字」模块得分（满分 20）。
// 输入为年/月/时三柱（不含日柱）双方的干支。每柱独立按 scoreDayPillar 规则得 0/5/10。
// 三柱总和 ∈ [0,30]，归一化到 [0,20]：(sum*2 + 1) / 3（整数四舍五入）。
func scoreEightChars(
	yearGanA, yearZhiA, yearGanB, yearZhiB string,
	monthGanA, monthZhiA, monthGanB, monthZhiB string,
	hourGanA, hourZhiA, hourGanB, hourZhiB string,
) int {
	y := scoreDayPillar(yearGanA, yearZhiA, yearGanB, yearZhiB)
	m := scoreDayPillar(monthGanA, monthZhiA, monthGanB, monthZhiB)
	h := scoreDayPillar(hourGanA, hourZhiA, hourGanB, hourZhiB)
	return normalizeEightCharsSum(y + m + h)
}

// normalizeEightCharsSum 把三柱总和（0..30）归一化到 [0,20]，整数四舍五入。
func normalizeEightCharsSum(sum int) int {
	if sum <= 0 {
		return 0
	}
	if sum >= 30 {
		return 20
	}
	return (sum*2 + 1) / 3
}
```

- [ ] **Step 4: 运行测试确认通过**

Run: `cd backend && go test ./pkg/bazi/ -run TestScoreEightChars -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add backend/pkg/bazi/compatibility_scoring.go backend/pkg/bazi/compatibility_scoring_test.go
git commit -m "feat(compatibility): add scoreEightChars (year+month+hour 20-pt)

Three pillar pairs scored by scoreDayPillar rule, summed and rounded
to [0,20] via (sum*2+1)/3.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>"
```

---

### Task 7: evidence 列表与 score_explanations 生成器

**Files:**
- Create: `backend/pkg/bazi/compatibility_evidence.go`

- [ ] **Step 1: 写失败测试（在主测试文件中）**

把这些 case 写到现有 `backend/pkg/bazi/compatibility_test.go` 末尾（暂保留旧文件，下面 Task 14 整体重写）：

```go
func TestBuildEvidences_ZodiacLiuhe(t *testing.T) {
	a := makeCompatNatal("甲子", "丙寅", "壬午", "丁未", "male")
	b := makeCompatNatal("己丑", "戊辰", "庚申", "辛酉", "female")
	ev := buildCompatibilityEvidencesV3(a, b)
	found := false
	for _, item := range ev {
		if item.EvidenceKey == "zodiac_liuhe" {
			if item.Weight != 50 || item.Dimension != "zodiac" || item.Polarity != "positive" {
				t.Errorf("zodiac_liuhe: bad shape %+v", item)
			}
			found = true
		}
	}
	if !found {
		t.Error("expected zodiac_liuhe evidence for 子/丑 pair")
	}
}

func TestBuildEvidences_AllHits_Count6(t *testing.T) {
	// 构造一组让 4 模块全部命中、且 eight_chars 三柱全命中的盘
	a := makeCompatNatal("甲子", "甲子", "甲子", "甲子", "male")
	b := makeCompatNatal("己丑", "己丑", "己丑", "己丑", "female")
	ev := buildCompatibilityEvidencesV3(a, b)
	// zodiac + nayin + day_pillar + eight_chars(year/month/hour) = 1+1+1+3 = 6
	if len(ev) != 6 {
		t.Errorf("all-hit case: got %d evidences, want 6", len(ev))
	}
}
```

- [ ] **Step 2: 运行测试确认失败**

Run: `cd backend && go test ./pkg/bazi/ -run TestBuildEvidences -v`
Expected: FAIL with `undefined: buildCompatibilityEvidencesV3`

- [ ] **Step 3: 实现 `compatibility_evidence.go`**

```go
package bazi

import "fmt"

// buildCompatibilityEvidencesV3 把 4 模块的命中按设计文档 §4.2 转为 evidence 列表。
// 仅产出"positive"性 evidence（纯加分制无 negative）；4 模块最多 6 条（1+1+1+3）。
func buildCompatibilityEvidencesV3(a, b *BaziResult) []CompatibilityEvidence {
	out := make([]CompatibilityEvidence, 0, 6)
	out = append(out, zodiacEvidence(a, b)...)
	out = append(out, nayinEvidence(a, b)...)
	out = append(out, dayPillarEvidence(a, b)...)
	out = append(out, eightCharsEvidence(a, b)...)
	return out
}

func zodiacEvidence(a, b *BaziResult) []CompatibilityEvidence {
	kind := branchCompatibilityKind(a.YearZhi, b.YearZhi)
	switch kind {
	case "liuhe":
		return []CompatibilityEvidence{{
			EvidenceKey: "zodiac_liuhe",
			Dimension:   "zodiac",
			Type:        "年支六合",
			Polarity:    "positive",
			Source:      "zodiac",
			Title:       "年支六合",
			Detail:      fmt.Sprintf("双方年支 %s/%s 构成六合，属相基础线吸引力强。", a.YearZhi, b.YearZhi),
			Weight:      50,
		}}
	case "sanhe":
		group := sanheGroupName(a.YearZhi, b.YearZhi)
		return []CompatibilityEvidence{{
			EvidenceKey: "zodiac_sanhe",
			Dimension:   "zodiac",
			Type:        "年支三合",
			Polarity:    "positive",
			Source:      "zodiac",
			Title:       "年支三合",
			Detail:      fmt.Sprintf("双方年支 %s/%s 同属 %s 三合局，气场协同。", a.YearZhi, b.YearZhi, group),
			Weight:      50,
		}}
	}
	return nil
}

func nayinEvidence(a, b *BaziResult) []CompatibilityEvidence {
	gzA := a.YearGan + a.YearZhi
	gzB := b.YearGan + b.YearZhi
	wxA := nayinElement(gzA)
	wxB := nayinElement(gzB)
	switch nayinRelation(wxA, wxB) {
	case "sheng":
		return []CompatibilityEvidence{{
			EvidenceKey: "nayin_sheng",
			Dimension:   "nayin",
			Type:        "纳音相生",
			Polarity:    "positive",
			Source:      "nayin",
			Title:       "纳音相生",
			Detail:      fmt.Sprintf("%s 与 %s 纳音五行相生，资源 / 情绪流动顺。", wxA, wxB),
			Weight:      20,
		}}
	case "same":
		return []CompatibilityEvidence{{
			EvidenceKey: "nayin_same",
			Dimension:   "nayin",
			Type:        "纳音相同",
			Polarity:    "positive",
			Source:      "nayin",
			Title:       "纳音同气",
			Detail:      fmt.Sprintf("双方纳音同为 %s，本质同气。", wxA),
			Weight:      20,
		}}
	}
	return nil
}

func dayPillarEvidence(a, b *BaziResult) []CompatibilityEvidence {
	if !branchCompatible(a.DayZhi, b.DayZhi) {
		return nil
	}
	if ganUpperTier(a.DayGan, b.DayGan) {
		return []CompatibilityEvidence{{
			EvidenceKey: "day_pillar_upper",
			Dimension:   "day_pillar",
			Type:        "日柱上档",
			Polarity:    "positive",
			Source:      "day_pillar",
			Title:       "日柱上档",
			Detail: fmt.Sprintf(
				"日柱 %s%s/%s%s 地支合且天干强化（五合 / 相生），亲密层结构稳。",
				a.DayGan, a.DayZhi, b.DayGan, b.DayZhi,
			),
			Weight: 10,
		}}
	}
	return []CompatibilityEvidence{{
		EvidenceKey: "day_pillar_lower",
		Dimension:   "day_pillar",
		Type:        "日柱下档",
		Polarity:    "positive",
		Source:      "day_pillar",
		Title:       "日柱次吉",
		Detail: fmt.Sprintf(
			"日柱 %s%s/%s%s 地支合，天干仅相同 / 克 / 无关，亲密层有基础但未达上吉。",
			a.DayGan, a.DayZhi, b.DayGan, b.DayZhi,
		),
		Weight: 5,
	}}
}

func eightCharsEvidence(a, b *BaziResult) []CompatibilityEvidence {
	out := make([]CompatibilityEvidence, 0, 3)
	type pillar struct {
		name    string
		label   string
		ganA    string
		zhiA    string
		ganB    string
		zhiB    string
	}
	pillars := []pillar{
		{"year", "年柱", a.YearGan, a.YearZhi, b.YearGan, b.YearZhi},
		{"month", "月柱", a.MonthGan, a.MonthZhi, b.MonthGan, b.MonthZhi},
		{"hour", "时柱", a.HourGan, a.HourZhi, b.HourGan, b.HourZhi},
	}
	for _, p := range pillars {
		s := scoreDayPillar(p.ganA, p.zhiA, p.ganB, p.zhiB)
		if s == 0 {
			continue
		}
		tier := "lower"
		if s == 10 {
			tier = "upper"
		}
		out = append(out, CompatibilityEvidence{
			EvidenceKey: "eight_chars_" + p.name + "_" + tier,
			Dimension:   "eight_chars",
			Type:        p.label + "对" + (map[string]string{"upper": "上档", "lower": "下档"})[tier],
			Polarity:    "positive",
			Source:      "eight_chars",
			Title:       p.label + "对" + (map[string]string{"upper": "上档", "lower": "下档"})[tier],
			Detail: fmt.Sprintf(
				"%s %s%s/%s%s 命中%s（贡献 %d）。",
				p.label, p.ganA, p.zhiA, p.ganB, p.zhiB,
				(map[string]string{"upper": "上档", "lower": "下档"})[tier], s,
			),
			Weight: s,
		})
	}
	return out
}
```

- [ ] **Step 4: 运行测试确认通过**

Run: `cd backend && go test ./pkg/bazi/ -run TestBuildEvidences -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add backend/pkg/bazi/compatibility_evidence.go backend/pkg/bazi/compatibility_test.go
git commit -m "feat(compatibility): generate evidences from 4-module hits

Each module hit produces one CompatibilityEvidence (zodiac/nayin/day_pillar
positive-only; eight_chars 0-3 per pillar). Up to 6 evidences total.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>"
```

---

### Task 8: score_explanations 生成器

**Files:**
- Modify: `backend/pkg/bazi/compatibility_evidence.go`
- Modify: `backend/pkg/bazi/compatibility_test.go`

- [ ] **Step 1: 写失败测试**

```go
func TestBuildScoreExplanationsV3_FourEntries(t *testing.T) {
	a := makeCompatNatal("甲子", "丙寅", "壬午", "丁未", "male")
	b := makeCompatNatal("己丑", "戊辰", "庚申", "辛酉", "female")
	ev := buildCompatibilityEvidencesV3(a, b)
	exps := buildScoreExplanationsV3(a, b, ev)
	if len(exps) != 4 {
		t.Fatalf("expected exactly 4 explanations (one per module), got %d", len(exps))
	}
	dims := map[string]bool{}
	for _, e := range exps {
		dims[e.Dimension] = true
		if e.NegativeFactor != "" || len(e.NegativeEvidenceKeys) != 0 {
			t.Errorf("v3 should never set negative factors, got %+v", e)
		}
	}
	for _, d := range []string{"zodiac", "nayin", "day_pillar", "eight_chars"} {
		if !dims[d] {
			t.Errorf("missing dimension %q", d)
		}
	}
}

func TestBuildScoreExplanationsV3_UnHitModule_HasSummary(t *testing.T) {
	a := makeCompatNatal("甲午", "丙寅", "壬午", "丁未", "male")
	b := makeCompatNatal("乙未", "戊辰", "庚申", "辛酉", "female") // 午未六合，但故意让其他模块测试
	ev := buildCompatibilityEvidencesV3(a, b)
	exps := buildScoreExplanationsV3(a, b, ev)
	for _, e := range exps {
		if e.Summary == "" {
			t.Errorf("dimension %q has empty summary", e.Dimension)
		}
	}
}
```

- [ ] **Step 2: 运行测试确认失败**

Run: `cd backend && go test ./pkg/bazi/ -run TestBuildScoreExplanations -v`
Expected: FAIL with `undefined: buildScoreExplanationsV3`

- [ ] **Step 3: 实现**

追加到 `backend/pkg/bazi/compatibility_evidence.go`：

```go
// buildScoreExplanationsV3 按 4 模块各出一条解释（zodiac/nayin/day_pillar/eight_chars）。
// 纯加分制下 NegativeFactor / NegativeEvidenceKeys 永远为空。
func buildScoreExplanationsV3(a, b *BaziResult, evidences []CompatibilityEvidence) []CompatibilityScoreExplanation {
	dimensions := []string{"zodiac", "nayin", "day_pillar", "eight_chars"}
	out := make([]CompatibilityScoreExplanation, 0, 4)
	for _, dim := range dimensions {
		hit := findEvidenceByDimension(evidences, dim)
		exp := CompatibilityScoreExplanation{Dimension: dim}
		if hit != nil {
			exp.PositiveFactor = hit.Title
			exp.PositiveEvidenceKeys = []string{hit.EvidenceKey}
		}
		exp.Summary = scoreExplanationSummaryV3(dim, hit, a, b)
		out = append(out, exp)
	}
	return out
}

func findEvidenceByDimension(evidences []CompatibilityEvidence, dim string) *CompatibilityEvidence {
	for i := range evidences {
		if evidences[i].Dimension == dim {
			return &evidences[i]
		}
	}
	return nil
}

func scoreExplanationSummaryV3(dim string, hit *CompatibilityEvidence, a, b *BaziResult) string {
	switch dim {
	case "zodiac":
		if hit == nil {
			return fmt.Sprintf("双方年支 %s/%s 无六合 / 三合，属相层级无加成。", a.YearZhi, b.YearZhi)
		}
		if hit.EvidenceKey == "zodiac_liuhe" {
			return fmt.Sprintf("双方属相 %s/%s 构成六合，关系基础线吸引力强。", a.YearZhi, b.YearZhi)
		}
		return fmt.Sprintf("双方属相 %s/%s 同属 %s 三合局，气场协同。",
			a.YearZhi, b.YearZhi, sanheGroupName(a.YearZhi, b.YearZhi))
	case "nayin":
		wxA := nayinElement(a.YearGan + a.YearZhi)
		wxB := nayinElement(b.YearGan + b.YearZhi)
		if hit == nil {
			return fmt.Sprintf("%s 与 %s 纳音五行相克，纳音层无加分。", wxA, wxB)
		}
		if hit.EvidenceKey == "nayin_sheng" {
			return fmt.Sprintf("%s 与 %s 纳音五行相生，资源 / 情绪流动顺。", wxA, wxB)
		}
		return fmt.Sprintf("双方纳音同为 %s，本质同气。", wxA)
	case "day_pillar":
		if hit == nil {
			return fmt.Sprintf("日柱 %s%s/%s%s 地支不合，亲密层无加成。",
				a.DayGan, a.DayZhi, b.DayGan, b.DayZhi)
		}
		if hit.EvidenceKey == "day_pillar_upper" {
			return fmt.Sprintf("日柱 %s%s/%s%s 地支合且天干五合 / 相生，亲密层结构稳。",
				a.DayGan, a.DayZhi, b.DayGan, b.DayZhi)
		}
		return fmt.Sprintf("日柱 %s%s/%s%s 地支合，天干仅相同 / 克 / 无关，亲密层有基础但未达上吉。",
			a.DayGan, a.DayZhi, b.DayGan, b.DayZhi)
	case "eight_chars":
		// 八字模块可能 0–3 柱命中，单条 evidence 无法表达命中数，故直接重新计算柱位命中。
		return eightCharsSummary(a, b)
	}
	return ""
}

// eightCharsSummary 单独计算 八字 命中柱数并生成 summary。
func eightCharsSummary(a, b *BaziResult) string {
	type p struct{ name, label, ga, za, gb, zb string }
	pillars := []p{
		{"year", "年柱", a.YearGan, a.YearZhi, b.YearGan, b.YearZhi},
		{"month", "月柱", a.MonthGan, a.MonthZhi, b.MonthGan, b.MonthZhi},
		{"hour", "时柱", a.HourGan, a.HourZhi, b.HourGan, b.HourZhi},
	}
	hits := 0
	var soloLabel string
	for _, pp := range pillars {
		if scoreDayPillar(pp.ga, pp.za, pp.gb, pp.zb) > 0 {
			hits++
			soloLabel = pp.label
		}
	}
	switch hits {
	case 0:
		return "年 / 月 / 时三柱均无合，外围层无加成。"
	case 1:
		return fmt.Sprintf("三柱中仅 %s 合，外围层支撑薄弱。", soloLabel)
	default:
		return fmt.Sprintf("年 / 月 / 时三柱中有 %d 柱合，外围层有支撑。", hits)
	}
}
```

**注意**：实现里 `scoreExplanationSummaryV3` 的 `eight_chars` 分支不依赖 `hit`，直接调用 `eightCharsSummary`，这是因为 8 chars 可能有 0–3 个命中，单条 evidence 无法表达"命中数"语境。

- [ ] **Step 4: 运行测试确认通过**

Run: `cd backend && go test ./pkg/bazi/ -run TestBuildScoreExplanations -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add backend/pkg/bazi/compatibility_evidence.go backend/pkg/bazi/compatibility_test.go
git commit -m "feat(compatibility): generate score explanations per module

One explanation per zodiac/nayin/day_pillar/eight_chars; negative factors
always empty (pure additive). Templates match design doc §4.3.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>"
```

---

### Task 9: summary_tags 生成器

**Files:**
- Modify: `backend/pkg/bazi/compatibility_evidence.go`
- Modify: `backend/pkg/bazi/compatibility_test.go`

- [ ] **Step 1: 写失败测试**

```go
func TestBuildSummaryTagsV3_AllHits(t *testing.T) {
	// 4 模块全命中 + 总分 100 → "上吉合盘" 必出
	got := buildSummaryTagsV3(CompatibilityDimensionScores{
		Zodiac: 50, Nayin: 20, DayPillar: 10, EightChars: 20,
	}, 100)
	if !containsString(got, "上吉合盘") {
		t.Errorf("expected 上吉合盘 tag, got %v", got)
	}
	if !containsString(got, "属相相合") {
		t.Errorf("expected 属相相合 tag, got %v", got)
	}
}

func TestBuildSummaryTagsV3_AllMiss(t *testing.T) {
	got := buildSummaryTagsV3(CompatibilityDimensionScores{}, 0)
	if !containsString(got, "合盘无加成") {
		t.Errorf("expected 合盘无加成 tag, got %v", got)
	}
}

func TestBuildSummaryTagsV3_MaxFour(t *testing.T) {
	got := buildSummaryTagsV3(CompatibilityDimensionScores{
		Zodiac: 50, Nayin: 20, DayPillar: 10, EightChars: 20,
	}, 100)
	if len(got) > 4 {
		t.Errorf("tags exceeded 4: %v", got)
	}
}

func containsString(slice []string, s string) bool {
	for _, x := range slice {
		if x == s {
			return true
		}
	}
	return false
}
```

- [ ] **Step 2: 运行测试确认失败**

Run: `cd backend && go test ./pkg/bazi/ -run TestBuildSummaryTagsV3 -v`
Expected: FAIL with `undefined: buildSummaryTagsV3`（`CompatibilityDimensionScores` 字段名 `Zodiac` 等暂时也未定义，会一并报错——下面 Task 15 才正式重命名）

**临时绕过**：本任务前置依赖 Task 15（字段重命名）。**调整执行顺序：把 Task 15 提前到这里之前**。或者先把测试断言改为旧字段，待 Task 15 后回填。

> **执行建议**：跳过本任务进入 Task 10–14，所有 Phase A 任务里凡涉及 `Zodiac/Nayin/DayPillar/EightChars` 字段的测试，先用 Task 15 之后再回填断言。Phase A 末尾的 Task 14 是入口收口，应在 Task 15 之后再做。
>
> 正式执行 order：1, 2, 3, 4, 5, 6, 7, 8（仅 scoreExplanations，不涉及类型字段名）, 15（类型重命名）, 9（summary_tags）, 10–14, 16–17, 18+。

- [ ] **Step 3: 实现 `buildSummaryTagsV3`**（前置：Task 15 已完成字段重命名）

追加到 `backend/pkg/bazi/compatibility_evidence.go`：

```go
// buildSummaryTagsV3 按 design §4.4 规则生成最多 4 条 tag。
func buildSummaryTagsV3(scores CompatibilityDimensionScores, total int) []string {
	tags := make([]string, 0, 4)
	if scores.Zodiac >= 50 {
		tags = append(tags, "属相相合")
	}
	if scores.Nayin >= 20 {
		tags = append(tags, "纳音同气")
	}
	if scores.DayPillar == 10 {
		tags = append(tags, "日柱上吉")
	} else if scores.DayPillar == 5 {
		tags = append(tags, "日柱次吉")
	}
	if scores.EightChars >= 14 {
		tags = append(tags, "八字承接好")
	}
	if total >= 80 {
		tags = append(tags, "上吉合盘")
	}
	if total < 60 && scores.Zodiac == 0 && scores.Nayin == 0 &&
		scores.DayPillar == 0 && scores.EightChars == 0 {
		tags = append(tags, "合盘无加成")
	}
	if len(tags) > 4 {
		tags = tags[:4]
	}
	return tags
}
```

- [ ] **Step 4: 运行测试确认通过**

Run: `cd backend && go test ./pkg/bazi/ -run TestBuildSummaryTagsV3 -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add backend/pkg/bazi/compatibility_evidence.go backend/pkg/bazi/compatibility_test.go
git commit -m "feat(compatibility): build summary tags from v3 module scores

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>"
```

---

### Task 10: relationshipType 分类 + decision_advice

**Files:**
- Create: `backend/pkg/bazi/compatibility_assessment.go`
- Modify: `backend/pkg/bazi/compatibility_test.go`

- [ ] **Step 1: 写失败测试**（前置 Task 15 字段名）

```go
func TestClassifyRelationshipType_AllBranches(t *testing.T) {
	cases := []struct {
		total, zodiac, dayPillar, eightChars int
		want                                  string
	}{
		{85, 50, 10, 20, "高契合型"},
		{75, 50, 10, 0, "亲密层稳固型"},
		{75, 50, 0, 0, "属相吸引型"},
		{55, 0, 10, 0, "亲密外围支撑型"},
		{55, 0, 0, 14, "亲密外围支撑型"},
		{30, 0, 0, 0, "合盘无加成"},
	}
	for _, tc := range cases {
		got := classifyRelationshipTypeV3(tc.total, CompatibilityDimensionScores{
			Zodiac: tc.zodiac, DayPillar: tc.dayPillar, EightChars: tc.eightChars,
		})
		if got != tc.want {
			t.Errorf("total=%d zodiac=%d day=%d 8chars=%d → got %q, want %q",
				tc.total, tc.zodiac, tc.dayPillar, tc.eightChars, got, tc.want)
		}
	}
}

func TestBuildDecisionAdviceV3_AllBranches(t *testing.T) {
	cases := []struct {
		total, hitsCount int
		recommendation   string
		verdict          string
		confidence       string
	}{
		{85, 4, "continue", "适合继续推进", "high"},
		{70, 2, "observe", "建议谨慎观察", "medium"},
		{50, 1, "caution", "不宜过早重投入", "medium"},
		{40, 0, "caution", "不宜过早重投入", "low"},
	}
	for _, tc := range cases {
		adv := buildDecisionAdviceV3(tc.total, tc.hitsCount)
		if adv.Recommendation != tc.recommendation || adv.Verdict != tc.verdict || adv.Confidence != tc.confidence {
			t.Errorf("total=%d hits=%d: got rec=%q verdict=%q conf=%q, want %q/%q/%q",
				tc.total, tc.hitsCount,
				adv.Recommendation, adv.Verdict, adv.Confidence,
				tc.recommendation, tc.verdict, tc.confidence)
		}
	}
}
```

注：`buildDecisionAdviceV3` 返回结构体 `decisionAdviceV3`（内部类型），含 `Recommendation / Verdict / Confidence / Conditions / DoNext / Avoid`。

- [ ] **Step 2: 运行测试确认失败**

Run: `cd backend && go test ./pkg/bazi/ -run "TestClassifyRelationshipType|TestBuildDecisionAdviceV3" -v`
Expected: FAIL with `undefined: classifyRelationshipTypeV3 / buildDecisionAdviceV3`

- [ ] **Step 3: 实现 `compatibility_assessment.go` 第 1 段**

```go
package bazi

// classifyRelationshipTypeV3 按 design §5.1 优先级链匹配关系类型。
// 短路求值：第一个 true 即返回。
func classifyRelationshipTypeV3(total int, s CompatibilityDimensionScores) string {
	switch {
	case total >= 80:
		return "高契合型"
	case s.Zodiac == 50 && s.DayPillar >= 5:
		return "亲密层稳固型"
	case s.Zodiac == 50:
		return "属相吸引型"
	case s.DayPillar >= 5 || s.EightChars >= 14:
		return "亲密外围支撑型"
	default:
		return "合盘无加成"
	}
}

type decisionAdviceV3 struct {
	Recommendation string
	Verdict        string
	Confidence     string
	Conditions     []string
	DoNext         []string
	Avoid          []string
}

// buildDecisionAdviceV3 按 design §5.2 三档生成建议。hitsCount = 4 模块中分>0 的个数。
func buildDecisionAdviceV3(total, hitsCount int) decisionAdviceV3 {
	var rec, verdict string
	switch {
	case total >= 80:
		rec = "continue"
		verdict = "适合继续推进"
	case total >= 60:
		rec = "observe"
		verdict = "建议谨慎观察"
	default:
		rec = "caution"
		verdict = "不宜过早重投入"
	}
	var confidence string
	switch {
	case hitsCount >= 3:
		confidence = "high"
	case hitsCount >= 1:
		confidence = "medium"
	default:
		confidence = "low"
	}
	conditions, doNext, avoid := decisionAdviceTextsV3(rec)
	return decisionAdviceV3{
		Recommendation: rec, Verdict: verdict, Confidence: confidence,
		Conditions: conditions, DoNext: doNext, Avoid: avoid,
	}
}

// decisionAdviceTextsV3 三档 conditions/do_next/avoid 文案模板（设计 §5.2 末段）。
func decisionAdviceTextsV3(recommendation string) (conditions, doNext, avoid []string) {
	switch recommendation {
	case "continue":
		return []string{
				"维持现有沟通节奏与现实安排",
				"在关键决策上保持双方同步",
			},
			[]string{
				"把长期承接的关键议题（住、责任分工）逐项落地",
				"用具体行为而非情绪强度判断关系稳定性",
			},
			[]string{
				"误以为 4 模块全命中就免维护，关系仍需经营",
				"用合盘结果替代日常沟通的具体内容",
			}
	case "observe":
		return []string{
				"在一到两个月内验证沟通节奏是否稳定",
				"把容易争执的话题具体化处理",
			},
			[]string{
				"先观察冲突后双方修复能力",
				"把短期吸引点和长期承接点分开评估",
			},
			[]string{
				"在关系规则未稳定前过早绑定重大决定",
				"用单一模块的结果（如属相相合）替代全局判断",
			}
	default:
		return []string{
				"先稳定个人节奏，再考虑重投入",
				"避免在缺少支点的阶段做长期承诺",
			},
			[]string{
				"用 1–3 件具体生活议题观察对方现实承接能力",
				"建立可暂停的关系边界",
			},
			[]string{
				"用『感觉』替代『结构证据』推动关系升级",
				"忽略合盘提示的弱支点强行投入",
			}
	}
}
```

- [ ] **Step 4: 运行测试确认通过**

Run: `cd backend && go test ./pkg/bazi/ -run "TestClassifyRelationshipType|TestBuildDecisionAdviceV3" -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add backend/pkg/bazi/compatibility_assessment.go backend/pkg/bazi/compatibility_test.go
git commit -m "feat(compatibility): classify relationship type & decision advice

Priority chain per design §5.1; three-tier verdict/recommendation/confidence
per §5.2 driven by total score + hits count.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>"
```

---

### Task 11: duration_assessment 三窗口评估

**Files:**
- Modify: `backend/pkg/bazi/compatibility_assessment.go`
- Modify: `backend/pkg/bazi/compatibility_test.go`

- [ ] **Step 1: 写失败测试**

```go
func TestBuildDurationAssessmentV3_Branches(t *testing.T) {
	cases := []struct {
		name                                       string
		zodiac, nayin, dayPillar, eightChars       int
		wantShort, wantMid, wantLong               CompatibilityDurationLevel
	}{
		{"all high",
			50, 20, 10, 20,
			CompatibilityDurationHigh, CompatibilityDurationHigh, CompatibilityDurationHigh},
		{"zodiac+nayin only",
			50, 20, 0, 0,
			CompatibilityDurationHigh, CompatibilityDurationLow, CompatibilityDurationLow},
		{"day_pillar lower with zodiac",
			50, 0, 5, 0,
			CompatibilityDurationMedium, CompatibilityDurationHigh, CompatibilityDurationLow},
		{"eight_chars strong only",
			0, 0, 0, 17,
			CompatibilityDurationLow, CompatibilityDurationLow, CompatibilityDurationMedium},
		{"all miss",
			0, 0, 0, 0,
			CompatibilityDurationLow, CompatibilityDurationLow, CompatibilityDurationLow},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			da := buildDurationAssessmentV3(CompatibilityDimensionScores{
				Zodiac: tc.zodiac, Nayin: tc.nayin,
				DayPillar: tc.dayPillar, EightChars: tc.eightChars,
			})
			if da.Windows.ThreeMonths.Level != tc.wantShort {
				t.Errorf("short: got %q, want %q", da.Windows.ThreeMonths.Level, tc.wantShort)
			}
			if da.Windows.OneYear.Level != tc.wantMid {
				t.Errorf("mid: got %q, want %q", da.Windows.OneYear.Level, tc.wantMid)
			}
			if da.Windows.TwoYearsPlus.Level != tc.wantLong {
				t.Errorf("long: got %q, want %q", da.Windows.TwoYearsPlus.Level, tc.wantLong)
			}
		})
	}
}
```

- [ ] **Step 2: 运行测试确认失败**

Run: `cd backend && go test ./pkg/bazi/ -run TestBuildDurationAssessmentV3 -v`
Expected: FAIL with `undefined: buildDurationAssessmentV3`

- [ ] **Step 3: 实现**

追加到 `backend/pkg/bazi/compatibility_assessment.go`：

```go
// buildDurationAssessmentV3 按 design §5.3 三窗口阈值生成评估。
func buildDurationAssessmentV3(s CompatibilityDimensionScores) CompatibilityDurationAssessment {
	short := shortWindowLevel(s)
	mid := midWindowLevel(s)
	long := longWindowLevel(s)
	return CompatibilityDurationAssessment{
		OverallBand: durationBandV3(long),
		Windows: CompatibilityDurationWindows{
			ThreeMonths:  CompatibilityDurationWindow{Level: short},
			OneYear:      CompatibilityDurationWindow{Level: mid},
			TwoYearsPlus: CompatibilityDurationWindow{Level: long},
		},
		Summary: durationSummaryV3(short, long),
		Reasons: nil, // 由调用方注入 evidence reasons
	}
}

func shortWindowLevel(s CompatibilityDimensionScores) CompatibilityDurationLevel {
	switch {
	case s.Zodiac == 50 && s.Nayin == 20:
		return CompatibilityDurationHigh
	case s.Zodiac == 50 || s.Nayin == 20:
		return CompatibilityDurationMedium
	default:
		return CompatibilityDurationLow
	}
}

func midWindowLevel(s CompatibilityDimensionScores) CompatibilityDurationLevel {
	switch {
	case s.DayPillar == 10 || (s.DayPillar == 5 && s.Zodiac == 50):
		return CompatibilityDurationHigh
	case s.DayPillar >= 5:
		return CompatibilityDurationMedium
	default:
		return CompatibilityDurationLow
	}
}

func longWindowLevel(s CompatibilityDimensionScores) CompatibilityDurationLevel {
	switch {
	case s.EightChars >= 14 && s.DayPillar >= 5:
		return CompatibilityDurationHigh
	case s.EightChars >= 7 || s.DayPillar == 10:
		return CompatibilityDurationMedium
	default:
		return CompatibilityDurationLow
	}
}

func durationBandV3(longLevel CompatibilityDurationLevel) string {
	switch longLevel {
	case CompatibilityDurationHigh:
		return "long_term"
	case CompatibilityDurationMedium:
		return "medium_term"
	default:
		return "short_term"
	}
}

func durationSummaryV3(short, long CompatibilityDurationLevel) string {
	switch {
	case short == CompatibilityDurationHigh && long == CompatibilityDurationHigh:
		return "属相与纳音支撑短期吸引，日柱与八字承接长期稳定，关系发展通道顺畅。"
	case short == CompatibilityDurationHigh && long == CompatibilityDurationLow:
		return "短期靠近感强，但长期承接薄弱——关系更像「先热后难」型，需要把短期热度引导到现实安排。"
	case long == CompatibilityDurationHigh:
		return "短期需要时间培养亲近感，长期承接稳——关系更适合慢热经营。"
	default:
		return "这段关系的维持性更依赖阶段中的现实磨合而非最初的命盘指向。"
	}
}

// durationReasonsFromEvidence 把前 3 条命中 evidence 转为 Reason 字符串。
func durationReasonsFromEvidence(evidences []CompatibilityEvidence) []string {
	out := make([]string, 0, 3)
	for _, ev := range evidences {
		if ev.Polarity != "positive" {
			continue
		}
		out = append(out, ev.Title+"："+ev.Detail)
		if len(out) >= 3 {
			break
		}
	}
	if len(out) == 0 {
		out = append(out, "当前盘面未触发任何加分模块，建议结合现实相处判断维持性。")
	}
	return out
}
```

- [ ] **Step 4: 运行测试确认通过**

Run: `cd backend && go test ./pkg/bazi/ -run TestBuildDurationAssessmentV3 -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add backend/pkg/bazi/compatibility_assessment.go backend/pkg/bazi/compatibility_test.go
git commit -m "feat(compatibility): three-window duration assessment

3-month / 1-year / 2-year+ levels per design §5.3, with band and summary
templates. Reasons populated from positive evidence by caller.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>"
```

---

### Task 12: stage_risks 与 relationship_strategy

**Files:**
- Modify: `backend/pkg/bazi/compatibility_assessment.go`
- Modify: `backend/pkg/bazi/compatibility_test.go`

- [ ] **Step 1: 写失败测试**

```go
func TestBuildStageRisksV3_AllWindowsHaveLevelAndKeys(t *testing.T) {
	evidences := []CompatibilityEvidence{
		{EvidenceKey: "zodiac_liuhe", Dimension: "zodiac"},
		{EvidenceKey: "nayin_sheng", Dimension: "nayin"},
		{EvidenceKey: "day_pillar_upper", Dimension: "day_pillar"},
		{EvidenceKey: "eight_chars_year_upper", Dimension: "eight_chars"},
	}
	duration := CompatibilityDurationAssessment{
		Windows: CompatibilityDurationWindows{
			ThreeMonths:  CompatibilityDurationWindow{Level: CompatibilityDurationHigh},
			OneYear:      CompatibilityDurationWindow{Level: CompatibilityDurationMedium},
			TwoYearsPlus: CompatibilityDurationWindow{Level: CompatibilityDurationLow},
		},
	}
	risks := buildStageRisksV3(duration, evidences)
	if len(risks) != 3 {
		t.Fatalf("expected 3 windows, got %d", len(risks))
	}
	for _, r := range risks {
		if r.RiskLevel == "" || r.MainRisk == "" || r.Trigger == "" || r.Advice == "" {
			t.Errorf("incomplete risk: %+v", r)
		}
	}
	// 验证 evidence key 按窗口正确挂载
	if !containsString(risks[0].EvidenceKeys, "zodiac_liuhe") &&
		!containsString(risks[0].EvidenceKeys, "nayin_sheng") {
		t.Error("3-month window should reference zodiac/nayin evidence")
	}
	if !containsString(risks[1].EvidenceKeys, "day_pillar_upper") {
		t.Error("1-year window should reference day_pillar evidence")
	}
	if !containsString(risks[2].EvidenceKeys, "eight_chars_year_upper") {
		t.Error("2-year+ window should reference eight_chars evidence")
	}
}

func TestBuildRelationshipStrategyV3_ThreeTiers(t *testing.T) {
	for _, rec := range []string{"continue", "observe", "caution"} {
		s := buildRelationshipStrategyV3(rec)
		if s.Communication == "" || s.Conflict == "" || s.Reality == "" || s.Boundary == "" {
			t.Errorf("recommendation %q produced empty strategy: %+v", rec, s)
		}
	}
}
```

- [ ] **Step 2: 运行测试确认失败**

Run: `cd backend && go test ./pkg/bazi/ -run "TestBuildStageRisksV3|TestBuildRelationshipStrategyV3" -v`
Expected: FAIL

- [ ] **Step 3: 实现**

追加到 `backend/pkg/bazi/compatibility_assessment.go`：

```go
// buildStageRisksV3 生成 3 个窗口的风险描述。文案按 (窗口×level) 9 套模板。
func buildStageRisksV3(duration CompatibilityDurationAssessment, evidences []CompatibilityEvidence) []CompatibilityStageRisk {
	zodiacKeys := evidenceKeysByDimension(evidences, "zodiac", "nayin")
	dayKeys := evidenceKeysByDimension(evidences, "day_pillar")
	eightKeys := evidenceKeysByDimension(evidences, "eight_chars")
	return []CompatibilityStageRisk{
		stageRiskV3("three_months", duration.Windows.ThreeMonths.Level, zodiacKeys),
		stageRiskV3("one_year", duration.Windows.OneYear.Level, dayKeys),
		stageRiskV3("two_years_plus", duration.Windows.TwoYearsPlus.Level, eightKeys),
	}
}

func stageRiskV3(window string, level CompatibilityDurationLevel, evidenceKeys []string) CompatibilityStageRisk {
	main, trigger, advice := stageRiskTextV3(window, level)
	return CompatibilityStageRisk{
		Window:       window,
		RiskLevel:    string(level),
		MainRisk:     main,
		Trigger:      trigger,
		Advice:       advice,
		EvidenceKeys: evidenceKeys,
	}
}

func stageRiskTextV3(window string, level CompatibilityDurationLevel) (main, trigger, advice string) {
	switch window {
	case "three_months":
		switch level {
		case CompatibilityDurationHigh:
			return "靠近感强但节奏需要校准", "对方推进速度与你不同步时", "保持轻量频繁互动，不急于规则化关系。"
		case CompatibilityDurationMedium:
			return "短期吸引点有限", "缺乏话题或场景持续输入时", "刻意制造共同体验，避免单方追逐。"
		default:
			return "短期吸引基础薄弱", "热度退去后缺少留存点", "用现实生活节奏检验是否值得继续投入。"
		}
	case "one_year":
		switch level {
		case CompatibilityDurationHigh:
			return "亲密层稳固但仍需经营", "生活节奏被外部压力打乱时", "建立稳定的冲突修复机制。"
		case CompatibilityDurationMedium:
			return "亲密层有支撑但易波动", "情绪强度替代具体沟通时", "把分歧拆成具体事项，不情绪化判断关系本身。"
		default:
			return "亲密层缺乏天然契合", "对方亲密表达与你期待错位时", "先观察互相调整的意愿，再做长期承诺。"
		}
	default:
		switch level {
		case CompatibilityDurationHigh:
			return "长期稳定基础好", "责任分工与资源投入需要落地时", "建立可持续的责任分工与共同计划。"
		case CompatibilityDurationMedium:
			return "长期承接强度中等", "现实压力（住、家庭、收入）进入关系时", "用阶段性目标替代『未来无限期』式承诺。"
		default:
			return "长期承接薄弱", "需要共同处理重大现实议题时", "在做长期承诺前重新评估关系结构。"
		}
	}
}

func evidenceKeysByDimension(evidences []CompatibilityEvidence, dims ...string) []string {
	set := map[string]bool{}
	for _, d := range dims {
		set[d] = true
	}
	out := make([]string, 0, 2)
	for _, ev := range evidences {
		if set[ev.Dimension] && ev.EvidenceKey != "" {
			out = append(out, ev.EvidenceKey)
		}
	}
	return out
}

// buildRelationshipStrategyV3 按 recommendation 三档切换 12 句策略模板（4 句 × 3 档）。
func buildRelationshipStrategyV3(recommendation string) CompatibilityRelationshipStrategy {
	switch recommendation {
	case "continue":
		return CompatibilityRelationshipStrategy{
			Communication: "重要议题用明确约定替代情绪试探。",
			Conflict:      "冲突先暂停升级，再回到具体事件与责任分工。",
			Reality:       "长期计划拆成可验证的小步骤，逐项落地。",
			Boundary:      "保持双方个人节奏，避免过早形成单向依赖。",
		}
	case "observe":
		return CompatibilityRelationshipStrategy{
			Communication: "重要话题做到事先沟通规则，再讨论内容。",
			Conflict:      "争执后给彼此 24 小时冷却，再回到事实层处理。",
			Reality:       "用 1–2 个生活议题（出行、家庭联系）观察现实承接能力。",
			Boundary:      "在关系规则未稳定前，避免重大物质或时间投入。",
		}
	default:
		return CompatibilityRelationshipStrategy{
			Communication: "用具体行为而非情绪强度作为判断锚点。",
			Conflict:      "冲突后先评估是否值得继续修复，再决定行动。",
			Reality:       "把共同决策的频率与强度降到最低，先稳定个人节奏。",
			Boundary:      "明确可暂停 / 可退出的关系边界，避免被动滑入承诺。",
		}
	}
}
```

- [ ] **Step 4: 运行测试确认通过**

Run: `cd backend && go test ./pkg/bazi/ -run "TestBuildStageRisksV3|TestBuildRelationshipStrategyV3" -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add backend/pkg/bazi/compatibility_assessment.go backend/pkg/bazi/compatibility_test.go
git commit -m "feat(compatibility): stage risks & relationship strategy templates

9-template stage risk text (3 windows × 3 levels) per design §5.4; 12-line
strategy (4 lines × 3 recommendation tiers) per §5.5.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>"
```

---

### Task 13: claim_evidence_links + 主入口拼装

**Files:**
- Modify: `backend/pkg/bazi/compatibility_assessment.go`
- Modify: `backend/pkg/bazi/compatibility_test.go`

- [ ] **Step 1: 写失败测试**

```go
func TestBuildClaimEvidenceLinksV3_FromHits(t *testing.T) {
	evidences := []CompatibilityEvidence{
		{EvidenceKey: "zodiac_liuhe"},
		{EvidenceKey: "day_pillar_upper"},
	}
	links := buildClaimEvidenceLinksV3("适合继续推进", evidences)
	if len(links) != 1 {
		t.Fatalf("expected 1 link, got %d", len(links))
	}
	if links[0].ClaimID != "relationship_main_judgement" {
		t.Errorf("bad ClaimID: %q", links[0].ClaimID)
	}
	if len(links[0].EvidenceKeys) != 2 {
		t.Errorf("expected 2 evidence keys, got %v", links[0].EvidenceKeys)
	}
}

func TestBuildClaimEvidenceLinksV3_NoEvidence_EmptyResult(t *testing.T) {
	links := buildClaimEvidenceLinksV3("不宜过早重投入", nil)
	if len(links) != 0 {
		t.Errorf("no evidence should produce no link, got %d", len(links))
	}
}

func TestBuildConsultingAssessmentV3_Integration(t *testing.T) {
	scores := CompatibilityDimensionScores{Zodiac: 50, Nayin: 20, DayPillar: 10, EightChars: 17}
	total := 97
	hits := 4
	evidences := []CompatibilityEvidence{
		{EvidenceKey: "zodiac_liuhe", Dimension: "zodiac"},
		{EvidenceKey: "nayin_sheng", Dimension: "nayin"},
		{EvidenceKey: "day_pillar_upper", Dimension: "day_pillar"},
		{EvidenceKey: "eight_chars_year_upper", Dimension: "eight_chars"},
	}
	duration := buildDurationAssessmentV3(scores)
	got := buildConsultingAssessmentV3(total, hits, scores, evidences, duration)
	if got.RelationshipDiagnosis.RelationshipType != "高契合型" {
		t.Errorf("bad relationship type: %q", got.RelationshipDiagnosis.RelationshipType)
	}
	if got.DecisionAdvice.Recommendation != "continue" {
		t.Errorf("bad recommendation: %q", got.DecisionAdvice.Recommendation)
	}
	if len(got.StageRisks) != 3 {
		t.Errorf("expected 3 stage risks, got %d", len(got.StageRisks))
	}
	if got.RelationshipStrategy.Communication == "" {
		t.Error("missing strategy")
	}
}
```

- [ ] **Step 2: 运行测试确认失败**

Run: `cd backend && go test ./pkg/bazi/ -run "TestBuildClaimEvidenceLinksV3|TestBuildConsultingAssessmentV3" -v`
Expected: FAIL

- [ ] **Step 3: 实现**

追加到 `backend/pkg/bazi/compatibility_assessment.go`：

```go
// buildClaimEvidenceLinksV3 把"main judgement" claim 与所有命中 evidence_key 关联。
// evidences 为空则返回空切片。
func buildClaimEvidenceLinksV3(verdict string, evidences []CompatibilityEvidence) []CompatibilityClaimEvidenceLink {
	keys := make([]string, 0, len(evidences))
	for _, ev := range evidences {
		if ev.EvidenceKey != "" {
			keys = append(keys, ev.EvidenceKey)
		}
	}
	if len(keys) == 0 {
		return nil
	}
	return []CompatibilityClaimEvidenceLink{{
		ClaimID:      "relationship_main_judgement",
		Claim:        verdict,
		EvidenceKeys: keys,
		Reasoning:    "主要判断综合 4 模块（合属相 / 合纳音 / 合日柱 / 合八字）的命中结果与总分。",
		Caveat:       "合盘表达的是关系倾向，现实选择与相处方式会改变结果表现。",
	}}
}

// buildConsultingAssessmentV3 把所有 consulting 字段拼装为完整结构。
func buildConsultingAssessmentV3(
	total, hitsCount int,
	scores CompatibilityDimensionScores,
	evidences []CompatibilityEvidence,
	duration CompatibilityDurationAssessment,
) CompatibilityConsultingAssessment {
	adv := buildDecisionAdviceV3(total, hitsCount)
	relType := classifyRelationshipTypeV3(total, scores)
	stageRisks := buildStageRisksV3(duration, evidences)
	strategy := buildRelationshipStrategyV3(adv.Recommendation)
	links := buildClaimEvidenceLinksV3(adv.Verdict, evidences)
	topKeys := make([]string, 0, len(evidences))
	for _, ev := range evidences {
		if ev.EvidenceKey != "" {
			topKeys = append(topKeys, ev.EvidenceKey)
			if len(topKeys) >= 3 {
				break
			}
		}
	}
	return CompatibilityConsultingAssessment{
		RelationshipDiagnosis: CompatibilityRelationshipDiagnosis{
			RelationshipType: relType,
			Verdict:          adv.Verdict,
			Summary:          relationshipDiagnosisSummaryV3(relType, adv.Recommendation),
			TopFindings: []CompatibilityFinding{{
				Text:         summaryFindingTextV3(adv.Recommendation),
				EvidenceKeys: topKeys,
			}},
		},
		DecisionAdvice: CompatibilityDecisionAdvice{
			Recommendation: adv.Recommendation,
			Confidence:     adv.Confidence,
			Conditions:     adv.Conditions,
			DoNext:         adv.DoNext,
			Avoid:          adv.Avoid,
		},
		StageRisks:           stageRisks,
		RelationshipStrategy: strategy,
		ClaimEvidenceLinks:   links,
	}
}

func relationshipDiagnosisSummaryV3(relType, recommendation string) string {
	switch recommendation {
	case "continue":
		return relType + "：4 模块共同支撑，关系具备继续推进的基础。"
	case "observe":
		return relType + "：合盘提示有支点也有缺口，需要边推进边观察。"
	default:
		return relType + "：合盘加分有限，长期稳定需先通过现实相处验证。"
	}
}

func summaryFindingTextV3(recommendation string) string {
	switch recommendation {
	case "continue":
		return "4 模块命中较多，关系优势具备结构性支撑。"
	case "observe":
		return "命中与未命中并存，重点观察缺口模块对应的相处场景。"
	default:
		return "命中模块稀少，建议把合盘结论与现实接触配合判断。"
	}
}
```

- [ ] **Step 4: 运行测试确认通过**

Run: `cd backend && go test ./pkg/bazi/ -run "TestBuildClaimEvidenceLinksV3|TestBuildConsultingAssessmentV3" -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add backend/pkg/bazi/compatibility_assessment.go backend/pkg/bazi/compatibility_test.go
git commit -m "feat(compatibility): claim-evidence links & consulting assembly

Wire up the full consulting_assessment shape from total/hits/scores/
evidences/duration. relationship_main_judgement claim references all hit
evidence keys.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>"
```

---

### Task 14: `AnalyzeCompatibility` 主入口重写 + 旧代码删除

**Files:**
- Modify: `backend/pkg/bazi/compatibility.go`（整体重写）
- Modify: `backend/pkg/bazi/compatibility_test.go`（保留 `makeCompatNatal` 工具，删除旧 assertion）

- [ ] **Step 1: 先读旧入口**

Run: `wc -l backend/pkg/bazi/compatibility.go`
Expected: 当前 1510 行（重写后将 ~150 行）

- [ ] **Step 2: 写集成测试（先于实现）**

替换 `backend/pkg/bazi/compatibility_test.go` 全部内容：

```go
package bazi

import (
	"strings"
	"testing"
)

func makeCompatNatal(yearGZ, monthGZ, dayGZ, hourGZ, gender string) *BaziResult {
	yr := []rune(yearGZ)
	mr := []rune(monthGZ)
	dr := []rune(dayGZ)
	hr := []rune(hourGZ)
	return &BaziResult{
		YearGan: string(yr[0]), YearZhi: string(yr[1]),
		MonthGan: string(mr[0]), MonthZhi: string(mr[1]),
		DayGan: string(dr[0]), DayZhi: string(dr[1]),
		HourGan: string(hr[0]), HourZhi: string(hr[1]),
		YearGanWuxing:  ganWuxing[string(yr[0])],
		MonthGanWuxing: ganWuxing[string(mr[0])],
		DayGanWuxing:   ganWuxing[string(dr[0])],
		HourGanWuxing:  ganWuxing[string(hr[0])],
		YearZhiWuxing:  zhiWuxing[string(yr[1])],
		MonthZhiWuxing: zhiWuxing[string(mr[1])],
		DayZhiWuxing:   zhiWuxing[string(dr[1])],
		HourZhiWuxing:  zhiWuxing[string(hr[1])],
		YearHideGan:    []string{string(yr[0])},
		MonthHideGan:   []string{string(mr[0])},
		DayHideGan:     []string{string(dr[0])},
		HourHideGan:    []string{string(hr[0])},
		Gender:         gender,
	}
}

func TestAnalyzeCompatibility_ReturnsV3CoreShape(t *testing.T) {
	a := makeCompatNatal("甲子", "丙寅", "甲子", "丁卯", "male")
	b := makeCompatNatal("己丑", "戊辰", "己丑", "庚申", "female")
	got := AnalyzeCompatibility(a, b)
	if got.OverallLevel == "" {
		t.Fatal("expected overall level")
	}
	if got.OverallScore < 0 || got.OverallScore > 100 {
		t.Fatalf("overall_score out of range: %d", got.OverallScore)
	}
	if got.DimensionScores.Zodiac < 0 || got.DimensionScores.Zodiac > 50 {
		t.Errorf("zodiac out of range: %d", got.DimensionScores.Zodiac)
	}
	if got.DimensionScores.Nayin < 0 || got.DimensionScores.Nayin > 20 {
		t.Errorf("nayin out of range: %d", got.DimensionScores.Nayin)
	}
	if got.DimensionScores.DayPillar < 0 || got.DimensionScores.DayPillar > 10 {
		t.Errorf("day_pillar out of range: %d", got.DimensionScores.DayPillar)
	}
	if got.DimensionScores.EightChars < 0 || got.DimensionScores.EightChars > 20 {
		t.Errorf("eight_chars out of range: %d", got.DimensionScores.EightChars)
	}
	wantTotal := got.DimensionScores.Zodiac + got.DimensionScores.Nayin +
		got.DimensionScores.DayPillar + got.DimensionScores.EightChars
	if got.OverallScore != wantTotal {
		t.Errorf("overall_score %d != sum of modules %d", got.OverallScore, wantTotal)
	}
	if got.ConsultingAssessment.RelationshipDiagnosis.RelationshipType == "" {
		t.Error("missing relationship type")
	}
}

func TestAnalyzeCompatibility_PerfectHit(t *testing.T) {
	// 双方完全同盘且互合 → 4 模块满分尝试
	a := makeCompatNatal("甲子", "甲子", "甲子", "甲子", "male")
	b := makeCompatNatal("己丑", "己丑", "己丑", "己丑", "female")
	got := AnalyzeCompatibility(a, b)
	if got.DimensionScores.Zodiac != 50 {
		t.Errorf("zodiac: got %d, want 50", got.DimensionScores.Zodiac)
	}
	if got.DimensionScores.Nayin != 20 {
		t.Errorf("nayin: got %d, want 20", got.DimensionScores.Nayin)
	}
	if got.DimensionScores.DayPillar != 10 {
		t.Errorf("day_pillar: got %d, want 10", got.DimensionScores.DayPillar)
	}
	if got.DimensionScores.EightChars != 20 {
		t.Errorf("eight_chars: got %d, want 20", got.DimensionScores.EightChars)
	}
	if got.OverallScore != 100 {
		t.Errorf("overall_score: got %d, want 100", got.OverallScore)
	}
	if got.OverallLevel != CompatibilityHigh {
		t.Errorf("overall_level: got %q, want %q", got.OverallLevel, CompatibilityHigh)
	}
}

func TestAnalyzeCompatibility_AllMiss(t *testing.T) {
	a := makeCompatNatal("甲午", "甲午", "甲午", "甲午", "male")
	b := makeCompatNatal("庚子", "庚子", "庚子", "庚子", "female")
	got := AnalyzeCompatibility(a, b)
	// 子午六冲（不算分）+ 纳音 海中金 vs 壁上土：土生金 → 仍 +20
	// 这里只断言总分上限：合属相必为 0
	if got.DimensionScores.Zodiac != 0 {
		t.Errorf("子午冲: zodiac should be 0, got %d", got.DimensionScores.Zodiac)
	}
	if got.OverallScore < 0 || got.OverallScore > 100 {
		t.Errorf("overall_score out of range: %d", got.OverallScore)
	}
}

func TestAnalyzeCompatibility_OverallLevelThresholds(t *testing.T) {
	cases := []struct {
		total int
		want  CompatibilityLevel
	}{
		{100, CompatibilityHigh},
		{80, CompatibilityHigh},
		{79, CompatibilityMedium},
		{60, CompatibilityMedium},
		{59, CompatibilityLow},
		{0, CompatibilityLow},
	}
	for _, tc := range cases {
		got := overallLevelFromScoreV3(tc.total)
		if got != tc.want {
			t.Errorf("score %d: got %q, want %q", tc.total, got, tc.want)
		}
	}
}

func TestAnalyzeCompatibility_EvidenceKeyAndShape(t *testing.T) {
	a := makeCompatNatal("甲子", "丙寅", "甲子", "丁卯", "male")
	b := makeCompatNatal("己丑", "戊辰", "己丑", "庚申", "female")
	got := AnalyzeCompatibility(a, b)
	for _, ev := range got.Evidences {
		if ev.Polarity != "positive" {
			t.Errorf("v3 evidence should always be positive, got %q", ev.Polarity)
		}
		if ev.EvidenceKey == "" {
			t.Errorf("evidence missing key: %+v", ev)
		}
		if !strings.HasPrefix(ev.EvidenceKey, ev.Dimension+"_") {
			t.Errorf("evidence_key %q does not start with dimension %q", ev.EvidenceKey, ev.Dimension)
		}
	}
}
```

- [ ] **Step 3: 运行测试确认失败**

Run: `cd backend && go test ./pkg/bazi/ -run TestAnalyzeCompatibility -v`
Expected: FAIL（旧 `AnalyzeCompatibility` 返回的字段名是 `Attraction` 等，类型不匹配；且 `overallLevelFromScoreV3` 未定义）

- [ ] **Step 4: 整体替换 `compatibility.go`**

把 `backend/pkg/bazi/compatibility.go` 整体内容替换为：

```go
package bazi

// Compatibility scoring v3 — 100-point classical formula
// (zodiac 50 / nayin 20 / day_pillar 10 / eight_chars 20).
// Design: docs/superpowers/specs/2026-05-27-compatibility-scoring-formula-v2-design.md

type CompatibilityLevel string

const (
	CompatibilityHigh   CompatibilityLevel = "high"
	CompatibilityMedium CompatibilityLevel = "medium"
	CompatibilityLow    CompatibilityLevel = "low"
)

type CompatibilityDurationLevel string

const (
	CompatibilityDurationHigh   CompatibilityDurationLevel = "high"
	CompatibilityDurationMedium CompatibilityDurationLevel = "medium"
	CompatibilityDurationLow    CompatibilityDurationLevel = "low"
)

type CompatibilityDimensionScores struct {
	Zodiac     int `json:"zodiac"`
	Nayin      int `json:"nayin"`
	DayPillar  int `json:"day_pillar"`
	EightChars int `json:"eight_chars"`
}

type CompatibilityEvidence struct {
	EvidenceKey    string   `json:"evidence_key"`
	Dimension      string   `json:"dimension"`
	Type           string   `json:"type"`
	Polarity       string   `json:"polarity"`
	Source         string   `json:"source"`
	Perspective    string   `json:"perspective,omitempty"`
	Actor          string   `json:"actor,omitempty"`
	Target         string   `json:"target,omitempty"`
	RelatedSources []string `json:"related_sources,omitempty"`
	Title          string   `json:"title"`
	Detail         string   `json:"detail"`
	Weight         int      `json:"weight"`
}

type CompatibilityScoreExplanation struct {
	Dimension            string   `json:"dimension"`
	PositiveFactor       string   `json:"positive_factor,omitempty"`
	NegativeFactor       string   `json:"negative_factor,omitempty"`
	PositiveEvidenceKeys []string `json:"positive_evidence_keys,omitempty"`
	NegativeEvidenceKeys []string `json:"negative_evidence_keys,omitempty"`
	Summary              string   `json:"summary"`
}

type CompatibilityFinding struct {
	Text         string   `json:"text"`
	EvidenceKeys []string `json:"evidence_keys"`
}

type CompatibilityRelationshipDiagnosis struct {
	RelationshipType string                 `json:"relationship_type"`
	Verdict          string                 `json:"verdict"`
	Summary          string                 `json:"summary"`
	TopFindings      []CompatibilityFinding `json:"top_findings"`
}

type CompatibilityDecisionAdvice struct {
	Recommendation string   `json:"recommendation"`
	Confidence     string   `json:"confidence"`
	Conditions     []string `json:"conditions"`
	DoNext         []string `json:"do_next"`
	Avoid          []string `json:"avoid"`
}

type CompatibilityStageRisk struct {
	Window       string   `json:"window"`
	RiskLevel    string   `json:"risk_level"`
	MainRisk     string   `json:"main_risk"`
	Trigger      string   `json:"trigger"`
	Advice       string   `json:"advice"`
	EvidenceKeys []string `json:"evidence_keys"`
}

type CompatibilityRelationshipStrategy struct {
	Communication string `json:"communication"`
	Conflict      string `json:"conflict"`
	Reality       string `json:"reality"`
	Boundary      string `json:"boundary"`
}

type CompatibilityClaimEvidenceLink struct {
	ClaimID      string   `json:"claim_id"`
	Claim        string   `json:"claim"`
	EvidenceKeys []string `json:"evidence_keys"`
	Reasoning    string   `json:"reasoning"`
	Caveat       string   `json:"caveat"`
}

type CompatibilityConsultingAssessment struct {
	RelationshipDiagnosis CompatibilityRelationshipDiagnosis `json:"relationship_diagnosis"`
	DecisionAdvice        CompatibilityDecisionAdvice        `json:"decision_advice"`
	StageRisks            []CompatibilityStageRisk           `json:"stage_risks"`
	RelationshipStrategy  CompatibilityRelationshipStrategy  `json:"relationship_strategy"`
	ClaimEvidenceLinks    []CompatibilityClaimEvidenceLink   `json:"claim_evidence_links"`
}

type CompatibilityDurationWindow struct {
	Level CompatibilityDurationLevel `json:"level"`
}

type CompatibilityDurationWindows struct {
	ThreeMonths  CompatibilityDurationWindow `json:"three_months"`
	OneYear      CompatibilityDurationWindow `json:"one_year"`
	TwoYearsPlus CompatibilityDurationWindow `json:"two_years_plus"`
}

type CompatibilityDurationAssessment struct {
	OverallBand string                       `json:"overall_band"`
	Windows     CompatibilityDurationWindows `json:"windows"`
	Summary     string                       `json:"summary"`
	Reasons     []string                     `json:"reasons"`
}

type CompatibilityAnalysis struct {
	OverallScore         int                               `json:"overall_score"`
	OverallLevel         CompatibilityLevel                `json:"overall_level"`
	DimensionScores      CompatibilityDimensionScores      `json:"dimension_scores"`
	Evidences            []CompatibilityEvidence           `json:"evidences"`
	ScoreExplanations    []CompatibilityScoreExplanation   `json:"score_explanations"`
	SummaryTags          []string                          `json:"summary_tags"`
	DurationAssessment   CompatibilityDurationAssessment   `json:"duration_assessment"`
	ConsultingAssessment CompatibilityConsultingAssessment `json:"consulting_assessment"`
}

// AnalyzeCompatibility 是合盘评分 v3 的公开入口。
// 计算 4 模块得分（合属相 / 合纳音 / 合日柱 / 合八字），汇总到总分（0–100），
// 并产出 evidence / score_explanations / summary_tags / duration / consulting 全套结构。
func AnalyzeCompatibility(a, b *BaziResult) CompatibilityAnalysis {
	if a == nil || b == nil {
		return CompatibilityAnalysis{
			OverallLevel:    CompatibilityLow,
			DimensionScores: CompatibilityDimensionScores{},
		}
	}
	scores := CompatibilityDimensionScores{
		Zodiac:    scoreZodiac(a.YearZhi, b.YearZhi),
		Nayin:     scoreNayin(a.YearGan+a.YearZhi, b.YearGan+b.YearZhi),
		DayPillar: scoreDayPillar(a.DayGan, a.DayZhi, b.DayGan, b.DayZhi),
		EightChars: scoreEightChars(
			a.YearGan, a.YearZhi, b.YearGan, b.YearZhi,
			a.MonthGan, a.MonthZhi, b.MonthGan, b.MonthZhi,
			a.HourGan, a.HourZhi, b.HourGan, b.HourZhi,
		),
	}
	total := scores.Zodiac + scores.Nayin + scores.DayPillar + scores.EightChars
	evidences := buildCompatibilityEvidencesV3(a, b)
	explanations := buildScoreExplanationsV3(a, b, evidences)
	tags := buildSummaryTagsV3(scores, total)
	duration := buildDurationAssessmentV3(scores)
	duration.Reasons = durationReasonsFromEvidence(evidences)
	hits := countHitsV3(scores)
	consulting := buildConsultingAssessmentV3(total, hits, scores, evidences, duration)
	return CompatibilityAnalysis{
		OverallScore:         total,
		OverallLevel:         overallLevelFromScoreV3(total),
		DimensionScores:      scores,
		Evidences:            evidences,
		ScoreExplanations:    explanations,
		SummaryTags:          tags,
		DurationAssessment:   duration,
		ConsultingAssessment: consulting,
	}
}

// overallLevelFromScoreV3 把 0–100 总分映射到 high / medium / low 三档。
func overallLevelFromScoreV3(total int) CompatibilityLevel {
	switch {
	case total >= 80:
		return CompatibilityHigh
	case total >= 60:
		return CompatibilityMedium
	default:
		return CompatibilityLow
	}
}

// countHitsV3 统计 4 模块中 score > 0 的个数（0–4）。
func countHitsV3(s CompatibilityDimensionScores) int {
	n := 0
	if s.Zodiac > 0 {
		n++
	}
	if s.Nayin > 0 {
		n++
	}
	if s.DayPillar > 0 {
		n++
	}
	if s.EightChars > 0 {
		n++
	}
	return n
}
```

- [ ] **Step 5: 编译验证**

Run: `cd backend && go build ./...`
Expected: 编译错误（service/repository 层仍用旧字段 Attraction 等）

**先**集成测试在 bazi 包内可独立通过：

Run: `cd backend && go test ./pkg/bazi/ -v -count=1`
Expected: PASS

如果整体 build 因 service 层报错，先临时跳过 build 验证，转入 Task 16+。

- [ ] **Step 6: Commit**

```bash
git add backend/pkg/bazi/compatibility.go backend/pkg/bazi/compatibility_test.go
git commit -m "feat(compatibility): rewrite AnalyzeCompatibility for v3 formula

Replace 1510-line evidence-weighted engine with 4-module additive scoring.
4 sibling files (nayin/scoring/evidence/assessment) provide the components;
this file holds only public types and the entry point (~190 LOC).

Note: backend service/repository layers still reference legacy fields
and will fail go build — fix follows in subsequent tasks.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>"
```

---

## Phase B：类型 + Service 层迁移

### Task 15: model.CompatibilityDimensionScores 字段重命名 + OverallScore 字段

**Files:**
- Modify: `backend/internal/model/compatibility.go`

**执行时机：** 见 Task 9 顶部说明——本任务应在 Task 1–7 之后、Task 9 之前执行。

- [ ] **Step 1: 修改 model**

把 `backend/internal/model/compatibility.go` 中以下两处改掉：

**(a)** 第 23-28 行：

```go
// 旧
type CompatibilityDimensionScores struct {
	Attraction    int `json:"attraction"`
	Stability     int `json:"stability"`
	Communication int `json:"communication"`
	Practicality  int `json:"practicality"`
}
```

替换为：

```go
type CompatibilityDimensionScores struct {
	Zodiac     int `json:"zodiac"`
	Nayin      int `json:"nayin"`
	DayPillar  int `json:"day_pillar"`
	EightChars int `json:"eight_chars"`
}
```

**(b)** `CompatibilityReading` struct（第 126 行附近）顶部加 `OverallScore` 字段：

```go
type CompatibilityReading struct {
	ID                   string                            `json:"id"`
	UserID               string                            `json:"user_id"`
	RelationshipStage    string                            `json:"relationship_stage"`
	PrimaryQuestion      string                            `json:"primary_question"`
	OverallScore         int                               `json:"overall_score"`   // 新增
	OverallLevel         string                            `json:"overall_level"`
	DimensionScores      CompatibilityDimensionScores      `json:"dimension_scores"`
	// ... 其余字段保持
}
```

同样在 `CompatibilityHistoryItem`（第 215 行附近）：

```go
type CompatibilityHistoryItem struct {
	ID                string                       `json:"id"`
	RelationshipStage string                       `json:"relationship_stage"`
	PrimaryQuestion   string                       `json:"primary_question"`
	OverallScore      int                          `json:"overall_score"`   // 新增
	OverallLevel      string                       `json:"overall_level"`
	DimensionScores   CompatibilityDimensionScores `json:"dimension_scores"`
	SummaryTags       []string                     `json:"summary_tags"`
	SelfName          string                       `json:"self_name"`
	PartnerName       string                       `json:"partner_name"`
	CreatedAt         time.Time                    `json:"created_at"`
}
```

- [ ] **Step 2: 编译试探**

Run: `cd backend && go build ./...`
Expected: 多处编译错误，定位在 `internal/service/compatibility_service.go`、`internal/repository/compatibility_repository.go`、可能还有 `prompt/canonical_compatibility.go` 或测试文件。这些会在 Task 16+ 修复。

- [ ] **Step 3: Commit**

```bash
git add backend/internal/model/compatibility.go
git commit -m "feat(compatibility): rename DimensionScores fields to v3 modules

DimensionScores keys change from attraction/stability/communication/
practicality to zodiac/nayin/day_pillar/eight_chars. Both
CompatibilityReading and CompatibilityHistoryItem get a new OverallScore
field (0–100). Build will fail until service/repository are updated
in Task 16.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>"
```

---

### Task 16: service 层映射更新 + version 升级

**Files:**
- Modify: `backend/internal/service/compatibility_service.go`

- [ ] **Step 1: 替换 4 处**

**(a)** 第 16 行：

```go
const compatibilityAnalysisVersion = "v2"
```

改为：

```go
const compatibilityAnalysisVersion = "v3"
```

**(b)** 第 71-76 行的 `DimensionScores` 映射：

```go
// 旧
model.CompatibilityDimensionScores{
	Attraction:    analysis.DimensionScores.Attraction,
	Stability:     analysis.DimensionScores.Stability,
	Communication: analysis.DimensionScores.Communication,
	Practicality:  analysis.DimensionScores.Practicality,
},
```

改为：

```go
model.CompatibilityDimensionScores{
	Zodiac:     analysis.DimensionScores.Zodiac,
	Nayin:      analysis.DimensionScores.Nayin,
	DayPillar:  analysis.DimensionScores.DayPillar,
	EightChars: analysis.DimensionScores.EightChars,
},
```

**(c)** 把 `analysis.OverallScore` 透传到 `CompatibilityReading.OverallScore`。

查 `repository.CreateCompatibilityReading` 当前签名（无 OverallScore 参数）。**需要扩展签名**——在 service 调用后单独 set 字段或更新 repository 签名。

最简方案：在 service 层 `CreateCompatibilityReading` 调用后追加一行：

```go
reading.OverallScore = analysis.OverallScore
```

这样不改 repository signature，OverallScore 暂不持久化（DB 列也没建）。但前端类型已含此字段，会从 service 响应里读取。

> **持久化考虑：** OverallScore 没有列；若用户后续刷新历史列表，OverallScore 会丢失。先**接受这一限制**并在 Task 17 把 OverallScore 持久化到 DB。

- [ ] **Step 2: 编译**

Run: `cd backend && go build ./...`
Expected: 仍有错误，定位在 `prompt/canonical_compatibility.go`（template 引用旧字段）或测试。继续修复。

如果仍卡在 service 测试，把现有 `compatibility_service_test.go` 中的字段断言（如有 Attraction/Stability）也改为新字段名。

- [ ] **Step 3: 跑 bazi 包测试**

Run: `cd backend && go test ./pkg/bazi/ -v -count=1`
Expected: PASS（独立模块测试已通）

- [ ] **Step 4: Commit**

```bash
git add backend/internal/service/compatibility_service.go
git commit -m "feat(compatibility): wire service to v3 fields + bump analysis_version

analysis_version constant goes "v2" → "v3"; DimensionScores mapping uses
zodiac/nayin/day_pillar/eight_chars; OverallScore transient-set on the
returned reading (persistence in next task).

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>"
```

---

### Task 17: OverallScore 持久化到 DB

**Files:**
- Modify: `backend/pkg/database/migrations/00011_compatibility_v3_analysis.sql`（创建）
- Modify: `backend/internal/repository/compatibility_repository.go`
- Modify: `backend/internal/service/compatibility_service.go`

- [ ] **Step 1: 创建 migration**

Create `backend/pkg/database/migrations/00011_compatibility_v3_analysis.sql`：

```sql
-- Compatibility v3: rename rationale comment + add overall_score column.
-- The new 100-point scoring formula stores per-module breakdown in the
-- existing dimension_scores JSONB column with new keys (zodiac/nayin/
-- day_pillar/eight_chars); analysis_version='v3' marks the new format.

ALTER TABLE compatibility_readings
	ADD COLUMN IF NOT EXISTS overall_score INTEGER NOT NULL DEFAULT 0;

COMMENT ON COLUMN compatibility_readings.analysis_version IS
	'v1/v2 = legacy 4-dim evidence-weighted scoring (attraction/stability/communication/practicality); v3 = zodiac/nayin/day_pillar/eight_chars 100-point classical formula';
COMMENT ON COLUMN compatibility_readings.overall_score IS
	'v3 only: total 0–100 score = sum of 4 module scores stored in dimension_scores JSONB. For v1/v2 records the column is 0.';
```

- [ ] **Step 2: 改 repository INSERT / SELECT**

修改 `backend/internal/repository/compatibility_repository.go` 第 11 行附近 `CreateCompatibilityReading` 函数签名与 SQL：

```go
func CreateCompatibilityReading(
	userID, overallLevel string,
	overallScore int,                  // 新增
	scores model.CompatibilityDimensionScores,
	scoreExplanations []model.CompatibilityScoreExplanation,
	duration model.CompatibilityDurationAssessment,
	consulting model.CompatibilityConsultingAssessment,
	summaryTags []string,
	analysisVersion string,
	context model.CompatibilityContext,
) (*model.CompatibilityReading, error) {
	// ...
	err := database.DB.QueryRow(
		`INSERT INTO compatibility_readings (user_id, overall_level, overall_score, dimension_scores, score_explanations, duration_assessment, consulting_assessment, summary_tags, analysis_version, relationship_stage, primary_question)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		 RETURNING id, user_id, relationship_stage, primary_question, overall_score, overall_level, dimension_scores, score_explanations, duration_assessment, consulting_assessment, summary_tags, analysis_version, created_at, updated_at`,
		userID, overallLevel, overallScore, scoresJSON, scoreExplanationsJSON, durationJSON, consultingJSON, tagsJSON, analysisVersion, context.RelationshipStage, context.PrimaryQuestion,
	).Scan(&r.ID, &r.UserID, &r.RelationshipStage, &r.PrimaryQuestion, &r.OverallScore, &r.OverallLevel, &rawScores, &rawScoreExplanations, &rawDuration, &rawConsulting, &rawTags, &r.AnalysisVersion, &r.CreatedAt, &r.UpdatedAt)
	// ...
}
```

同样修改第 96 行附近的 `GetCompatibilityReadingByID` 与 `GetCompatibilityReadingsByUserID` 的 SELECT 子句，把 `overall_score` 加进 column list 和 `Scan(...)` 列表（在 `r.OverallLevel` 前）。

历史列表的 SELECT 也加上 `overall_score`，并在 `CompatibilityHistoryItem` 的 Scan 列表里读取。

- [ ] **Step 3: service 调整调用**

修改 `compatibility_service.go` 中 `CreateCompatibilityReading` 调用：

```go
reading, err := repository.CreateCompatibilityReading(
	userID,
	string(analysis.OverallLevel),
	analysis.OverallScore,           // 新增
	model.CompatibilityDimensionScores{...},
	// ... 其余不变
)
```

删除 Task 16 步骤 1 中追加的 `reading.OverallScore = analysis.OverallScore`（不再需要，已由 DB 回填）。

- [ ] **Step 4: 跑 migration + 编译**

Run: 
```bash
docker-compose exec postgres psql -U yuanju -d yuanju -f /docker-entrypoint-initdb.d/00011_compatibility_v3_analysis.sql 2>&1 || echo "migration file path may differ; alternative: run manually"
```

或者直接挂载后重启服务。

Run: `cd backend && go build ./...`
Expected: 可能有少量编译错误待补——把所有调用 `CreateCompatibilityReading` 的地方都补上 `overallScore` 参数（grep `CreateCompatibilityReading\(` 找全）。

Run: `cd backend && go test ./... -count=1`
Expected: PASS（service 测试可能也需要更新——保留旧 mock 也要补 overallScore）

- [ ] **Step 5: Commit**

```bash
git add backend/pkg/database/migrations/00011_compatibility_v3_analysis.sql backend/internal/repository/compatibility_repository.go backend/internal/service/compatibility_service.go
git commit -m "feat(compatibility): persist overall_score in DB

Migration 00011 adds an INTEGER column with default 0; new INSERT path
writes the v3 total score, legacy v1/v2 records remain at 0. SELECT
queries surface overall_score for the reading and history endpoints.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>"
```

---

## Phase C：Migration + AI Prompt

### Task 18: 检查并修复 `compatibility_handler_test.go` 残留 v2 字段

**Files:**
- Modify: `backend/internal/handler/compatibility_handler_test.go`（按需）

- [ ] **Step 1: 全项目 grep 残留**

Run: `grep -rn "Attraction\|Stability\|Communication\|Practicality" backend/ --include='*.go' | grep -v event_signals | grep -v "// "`

Expected: 列出所有引用旧字段名的位置。逐一处理：
- 测试 fixture 改字段名（参考 v3 模块名）
- 模板字符串改措辞

- [ ] **Step 2: 修复每处**

具体每个 hit 单独修复，确保只动测试或映射代码，不修改 v3 算法逻辑。

- [ ] **Step 3: 全量编译 + 测试**

Run: `cd backend && go build ./... && go test ./... -count=1`
Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add -u backend/
git commit -m "fix(compatibility): clean up legacy 4-dim field references

Tests / fixtures / mappings that still mentioned attraction/stability/
communication/practicality are updated to match v3 module names.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>"
```

---

### Task 19: AI Prompt 改造

**Files:**
- Modify: `backend/pkg/prompt/canonical_compatibility.go`

- [ ] **Step 1: 读现行 prompt**

Run: `wc -l backend/pkg/prompt/canonical_compatibility.go`
Expected: ~150-200 行（视实际）。

打开看 `compatibilityCanonicalContent` const 全文。

- [ ] **Step 2: 改 prompt 文本**

在 `compatibilityCanonicalContent` 中：

**(a)** 把 "四维分数（JSON）：" 段落改写为：

```
四模块分数（JSON，v3 评分公式）：
{{.ScoresJSON}}

评分规则说明：
- zodiac（合属相，0–50）：年支六合或三合命中即满分 50，否则 0。
- nayin（合纳音，0–20）：年柱纳音五行相生或相同得 20，相克 0。
- day_pillar（合日柱，0–10）：日支合 + 干合/相生 = 10，日支合 + 其他 = 5，日支不合 = 0。
- eight_chars（合八字，0–20）：年/月/时三柱独立按合日柱规则得 0/5/10，三柱和归一化到 [0,20]。
- 总分 = 四模块直接相加 ∈ [0,100]：≥80 high；60–79 medium；<60 low。
- 本算法采用「纯加分制」，所有 evidence 的 polarity 均为 positive；不命中的模块得 0 分，不产生 evidence。
```

**(b)** "四维分数解释（JSON…）" 中"维度"措辞 → "模块"。

**(c)** "结构化证据（JSON）" 段保留，但加一句说明：

```
注：所有 evidence 来源仅四种（zodiac / nayin / day_pillar / eight_chars），polarity 永远为 positive。
```

**(d)** 把 prompt JSON 示例中的 "spouse_palace_stability_spouse_palace_chong" 这类旧 evidence_key 改为 v3 风格 key（如 "zodiac_liuhe"）。

**(e)** "证据约束" 列表里把"perspective/actor/target 理解方向性证据"删除（v3 不再使用方向性字段）。

- [ ] **Step 3: prompt 文本测试**

Run: `cd backend && go test ./pkg/prompt/ -count=1`
Expected: 现有 `canonical_test.go` / `sync_test.go` 可能涉及对内容字符串的断言——按需更新断言。

- [ ] **Step 4: Commit**

```bash
git add backend/pkg/prompt/canonical_compatibility.go backend/pkg/prompt/canonical_test.go
git commit -m "feat(compatibility): rewrite AI prompt for v3 4-module scoring

Replace 4-dimension psychology wording with 4-module classical wording;
add explicit scoring rule paragraph; remove perspective/actor/target
clause (v3 evidence has no directional fields).

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>"
```

---

## Phase D：Frontend version 分支

### Task 20: TypeScript 类型重定义

**Files:**
- Modify: `frontend/src/lib/api.ts`

- [ ] **Step 1: 定位现有类型**

Run: `grep -n "CompatibilityDimensionScores\|CompatibilityReading\b" frontend/src/lib/api.ts | head -10`

Expected: 找到 type 定义位置（约第 350-420 行）。

- [ ] **Step 2: 改类型**

在 `frontend/src/lib/api.ts` 中找到 `CompatibilityDimensionScores`，新增一个联合类型并保留 legacy：

```ts
// v1/v2 legacy（保留以兼容历史记录渲染）
export type CompatibilityDimensionScoresLegacy = {
  attraction: number
  stability: number
  communication: number
  practicality: number
}

// v3 新结构
export type CompatibilityDimensionScoresV3 = {
  zodiac: number
  nayin: number
  day_pillar: number
  eight_chars: number
}

// 联合类型 — 由 analysis_version 鉴别
export type CompatibilityDimensionScores =
  | CompatibilityDimensionScoresLegacy
  | CompatibilityDimensionScoresV3
```

在 `CompatibilityReading` 中：

```ts
export type CompatibilityReading = {
  id: string
  // ...
  overall_score: number               // 新增
  overall_level: 'high' | 'medium' | 'low'
  dimension_scores: CompatibilityDimensionScores
  analysis_version: 'v1' | 'v2' | 'v3'
  // ...
}
```

`CompatibilityHistoryItem` 同样加 `overall_score: number` 和 `analysis_version` 字段。

- [ ] **Step 3: 加类型守卫**

在 `api.ts` 同文件追加 helper：

```ts
export function isV3DimensionScores(
  scores: CompatibilityDimensionScores,
): scores is CompatibilityDimensionScoresV3 {
  return 'zodiac' in scores
}
```

- [ ] **Step 4: 类型检查**

Run: `cd frontend && npx tsc --noEmit`
Expected: 可能有多处 TS 错误——使用 `dimension_scores.attraction` 的代码会 unsafe，需要 narrowing。这些会在 Task 21-23 修复。

- [ ] **Step 5: Commit**

```bash
git add frontend/src/lib/api.ts
git commit -m "feat(compatibility): TS types add v3 schema with version-discriminated union

CompatibilityDimensionScores becomes a union of v1/v2 legacy (attraction/
stability/communication/practicality) and v3 (zodiac/nayin/day_pillar/
eight_chars). overall_score added to CompatibilityReading and history
items.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>"
```

---

### Task 21: ScoreOverviewV3 组件

**Files:**
- Modify: `frontend/src/pages/CompatibilityResultPage.tsx`（顶部插入新组件）

- [ ] **Step 1: 找现有 ScoreOverview**

Run: `grep -n "function ScoreOverview" frontend/src/pages/CompatibilityResultPage.tsx`
Expected: 约第 304 行。

- [ ] **Step 2: 在 `ScoreOverview` 上方插入 `ScoreOverviewV3`**

```tsx
const dimensionHintV3: Record<keyof CompatibilityDimensionScoresV3, string> = {
  zodiac: '属相（年支）层：六合/三合 命中即满分 50',
  nayin: '纳音五行：相生/相同 命中即满分 20',
  day_pillar: '日柱（亲密层）：支合 + 干合/生 满分 10',
  eight_chars: '年/月/时三柱：外围承接，最高 20',
}

const dimensionLabelV3: Record<keyof CompatibilityDimensionScoresV3, string> = {
  zodiac: '合属相',
  nayin: '合纳音',
  day_pillar: '合日柱',
  eight_chars: '合八字',
}

const dimensionMaxV3: Record<keyof CompatibilityDimensionScoresV3, number> = {
  zodiac: 50,
  nayin: 20,
  day_pillar: 10,
  eight_chars: 20,
}

function ScoreOverviewV3({
  scores,
  overallScore,
  overallLevel,
}: {
  scores: CompatibilityDimensionScoresV3
  overallScore: number
  overallLevel: 'high' | 'medium' | 'low'
}) {
  const keys: Array<keyof CompatibilityDimensionScoresV3> = [
    'zodiac',
    'nayin',
    'day_pillar',
    'eight_chars',
  ]
  return (
    <section className="compat-score-v3">
      <header className="compat-score-v3__header">
        <span className="compat-score-v3__total">{overallScore}</span>
        <span className="compat-score-v3__unit">/100</span>
        <span className={`compat-score-v3__badge compat-score-v3__badge--${overallLevel}`}>
          {overallLevel === 'high' ? '上吉' : overallLevel === 'medium' ? '中' : '低'}
        </span>
      </header>
      <ul className="compat-score-v3__modules">
        {keys.map((key) => {
          const value = scores[key]
          const max = dimensionMaxV3[key]
          return (
            <li key={key} className="compat-score-v3__module">
              <div className="compat-score-v3__module-row">
                <span className="compat-score-v3__module-label">{dimensionLabelV3[key]}</span>
                <span className="compat-score-v3__module-value">
                  {value}<span className="compat-score-v3__module-max">/{max}</span>
                </span>
              </div>
              <div className="compat-score-v3__module-hint">{dimensionHintV3[key]}</div>
              <div className="compat-score-v3__module-bar">
                <div
                  className="compat-score-v3__module-bar-fill"
                  style={{ width: `${(value / max) * 100}%` }}
                />
              </div>
            </li>
          )
        })}
      </ul>
    </section>
  )
}
```

- [ ] **Step 3: 给组件加 CSS**

把对应的 class 样式加到 `frontend/src/pages/CompatibilityResultPage.css`：

```css
.compat-score-v3 {
  background: var(--color-surface);
  border-radius: 12px;
  padding: 24px;
  margin-bottom: 16px;
}
.compat-score-v3__header {
  display: flex; align-items: baseline; gap: 8px;
  margin-bottom: 16px;
}
.compat-score-v3__total {
  font-size: 56px; font-weight: 700;
  color: var(--color-text-primary);
}
.compat-score-v3__unit {
  font-size: 18px; color: var(--color-text-secondary);
}
.compat-score-v3__badge {
  margin-left: auto;
  padding: 4px 12px;
  border-radius: 999px;
  font-size: 12px;
}
.compat-score-v3__badge--high { background: var(--color-accent); color: white; }
.compat-score-v3__badge--medium { background: var(--color-warning-soft); color: var(--color-warning); }
.compat-score-v3__badge--low { background: var(--color-muted); color: var(--color-text-secondary); }
.compat-score-v3__modules { list-style: none; padding: 0; margin: 0; display: grid; gap: 12px; }
.compat-score-v3__module { padding: 12px; border-radius: 8px; background: var(--color-surface-soft); }
.compat-score-v3__module-row { display: flex; justify-content: space-between; }
.compat-score-v3__module-label { font-weight: 600; }
.compat-score-v3__module-value { font-size: 18px; font-weight: 700; }
.compat-score-v3__module-max { color: var(--color-text-secondary); font-weight: 400; }
.compat-score-v3__module-hint { font-size: 12px; color: var(--color-text-secondary); margin-top: 4px; }
.compat-score-v3__module-bar {
  height: 4px; margin-top: 8px;
  background: var(--color-border); border-radius: 2px; overflow: hidden;
}
.compat-score-v3__module-bar-fill {
  height: 100%; background: var(--color-accent); transition: width 200ms ease;
}
```

- [ ] **Step 4: 启动前端验证渲染**

Run（开新终端）: `cd frontend && npm run dev`
打开浏览器到合盘结果页，验证 v3 组件能渲染（手动构造一个 v3 数据可走 Task 22）。

- [ ] **Step 5: Commit**

```bash
git add frontend/src/pages/CompatibilityResultPage.tsx frontend/src/pages/CompatibilityResultPage.css
git commit -m "feat(compatibility): add ScoreOverviewV3 component & styles

100-point total + 4 module breakdown (zodiac/nayin/day_pillar/eight_chars)
with progress bars; not yet wired into the page — see next task.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>"
```

---

### Task 22: 把 `CompatibilityResultPage` 接入 version 分支

**Files:**
- Modify: `frontend/src/pages/CompatibilityResultPage.tsx`

- [ ] **Step 1: 找当前 ScoreOverview 调用点**

Run: `grep -n "<ScoreOverview" frontend/src/pages/CompatibilityResultPage.tsx`

定位 render 中调用 `<ScoreOverview scores={...} />` 的代码段。

- [ ] **Step 2: 插入 version 分支**

把调用处改成：

```tsx
{reading.analysis_version === 'v3' && isV3DimensionScores(reading.dimension_scores) ? (
  <ScoreOverviewV3
    scores={reading.dimension_scores}
    overallScore={reading.overall_score}
    overallLevel={reading.overall_level}
  />
) : (
  <ScoreOverview scores={reading.dimension_scores as CompatibilityDimensionScoresLegacy} />
)}
```

`isV3DimensionScores` 已在 Task 20 添加。

- [ ] **Step 3: 全文件检查其它使用 `attraction/stability` 的位置**

Run: `grep -n "attraction\|stability\|communication\|practicality" frontend/src/pages/CompatibilityResultPage.tsx`

每处都按 version 守卫包装一层；或者把 legacy 渲染路径中所有引用 narrowing 到 `CompatibilityDimensionScoresLegacy`：

```ts
const legacyScores = reading.dimension_scores as CompatibilityDimensionScoresLegacy
// 之后 legacyScores.attraction 等可用
```

仅在 `analysis_version !== 'v3'` 分支内才使用 legacyScores。

- [ ] **Step 4: 类型检查**

Run: `cd frontend && npx tsc --noEmit`
Expected: PASS（无 TS 错误）

- [ ] **Step 5: 浏览器验证**

启动前端 + 后端（docker-compose up -d 或本地 go run），生成一个新合盘记录，验证 v3 路径渲染正确；切到旧历史记录验证 v1/v2 仍然展示。

- [ ] **Step 6: Commit**

```bash
git add frontend/src/pages/CompatibilityResultPage.tsx
git commit -m "feat(compatibility): wire ScoreOverviewV3 with version dispatch

reading.analysis_version === 'v3' renders the new 100-point overview;
v1/v2 records keep the legacy 4-dim renderer untouched.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>"
```

---

### Task 23: 历史列表 version 分支

**Files:**
- Modify: `frontend/src/pages/CompatibilityHistoryPage.tsx`

- [ ] **Step 1: 找当前历史卡片渲染**

Run: `grep -n "dimension_scores\|history" frontend/src/pages/CompatibilityHistoryPage.tsx | head -10`

定位列表项渲染（每个 history item 显示分数/level/tags 的部分）。

- [ ] **Step 2: 添加 mini score badge**

把列表 item 中 score 显示改为：

```tsx
{item.analysis_version === 'v3' ? (
  <div className="compat-history__score-v3">
    <span className="compat-history__score-v3-value">{item.overall_score}</span>
    <span className="compat-history__score-v3-unit">/100</span>
    <span className={`compat-history__level compat-history__level--${item.overall_level}`}>
      {item.overall_level === 'high' ? '上吉' : item.overall_level === 'medium' ? '中' : '低'}
    </span>
  </div>
) : (
  <div className="compat-history__score-legacy">
    {/* 保留现行 4 维度小图 */}
    {/* ... 现有渲染 */}
  </div>
)}
```

- [ ] **Step 3: 加最小样式**

Append to `frontend/src/pages/CompatibilityHistoryPage.css`：

```css
.compat-history__score-v3 {
  display: flex; align-items: baseline; gap: 4px;
}
.compat-history__score-v3-value {
  font-size: 24px; font-weight: 700;
  color: var(--color-text-primary);
}
.compat-history__score-v3-unit {
  font-size: 12px; color: var(--color-text-secondary);
}
.compat-history__level {
  margin-left: 8px;
  padding: 2px 8px; border-radius: 999px; font-size: 11px;
}
.compat-history__level--high { background: var(--color-accent); color: white; }
.compat-history__level--medium { background: var(--color-warning-soft); color: var(--color-warning); }
.compat-history__level--low { background: var(--color-muted); color: var(--color-text-secondary); }
```

- [ ] **Step 4: 类型检查 + 视觉验证**

Run: `cd frontend && npx tsc --noEmit`
Expected: PASS

Run（开发服务器已起）: 打开 `/compatibility/history`，确认旧记录显示 legacy 4-dim mini badge，新记录显示总分大数字。

- [ ] **Step 5: Commit**

```bash
git add frontend/src/pages/CompatibilityHistoryPage.tsx frontend/src/pages/CompatibilityHistoryPage.css
git commit -m "feat(compatibility): history list dual-renderer for v1/v2/v3

v3 records show overall_score with level badge; legacy keeps the 4-dim
mini visualisation.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>"
```

---

## Phase E：OpenSpec 归档

### Task 24: 新建 OpenSpec change 目录

**Files:**
- Create: `openspec/changes/compatibility-scoring-formula-v3/proposal.md`
- Create: `openspec/changes/compatibility-scoring-formula-v3/design.md`
- Create: `openspec/changes/compatibility-scoring-formula-v3/tasks.md`
- Create: `openspec/changes/compatibility-scoring-formula-v3/specs/compatibility-scoring-formula/spec.md`

- [ ] **Step 1: 写 `proposal.md`**

```markdown
# Compatibility Scoring Formula v3

## Why
合盘模块当前评分算法（11 信号源 × 4 维度 × evidence 加权 + 来源贡献封顶）存在双重计分、十神过粗、五行失衡触发门槛过严、`compatibility.go` 超 1510 行违反 500 行硬约束等问题。用户给出的传统命理 100 分制公式（合属相 50 + 合纳音 20 + 合日柱 10 + 合八字 20）表达更直接，且分模块解释更容易向用户展示。

## What Changes
- 替换 `backend/pkg/bazi/compatibility.go` 全部评分逻辑为「纯加分制」4 模块算法
- `CompatibilityDimensionScores` 字段重命名：`attraction/stability/communication/practicality` → `zodiac/nayin/day_pillar/eight_chars`
- 新增 `overall_score` 字段（INTEGER 列 + JSON 顶层字段）
- `analysis_version` 由 `"v2"` 升至 `"v3"`；v1/v2 旧记录通过前端 version 分支保留渲染
- 旧两份 spec `compatibility-explainable-compatibility-scoring` / `compatibility-depth-signal-engine` 待归档（语义被新公式推翻）
- frontend 历史页 + 结果页 version 分支双渲染

## Impact
- Affected specs:
  - **REMOVED:** `compatibility-explainable-compatibility-scoring`（v3 取消 evidence 加权与来源贡献封顶模型）
  - **REMOVED:** `compatibility-depth-signal-engine`（v3 取消 11 类信号源体系）
  - **ADDED:** `compatibility-scoring-formula`（本 change）
- Affected code:
  - backend/pkg/bazi/compatibility*.go（重写 + 4 个新文件）
  - backend/internal/model/compatibility.go（字段重命名）
  - backend/internal/service/compatibility_service.go（version + 映射）
  - backend/internal/repository/compatibility_repository.go（overall_score 列）
  - backend/pkg/database/migrations/00011_compatibility_v3_analysis.sql（新增）
  - backend/pkg/prompt/canonical_compatibility.go（prompt 改造）
  - frontend/src/lib/api.ts（类型联合）
  - frontend/src/pages/CompatibilityResultPage.tsx & .css（v3 组件 + 分支）
  - frontend/src/pages/CompatibilityHistoryPage.tsx & .css（version 分支）
- DB migration: 新增列 `overall_score INTEGER NOT NULL DEFAULT 0`
- 历史记录：保留不动；前端按 `analysis_version` 切渲染分支
```

- [ ] **Step 2: 写 `design.md`**

把现有设计文档 `docs/superpowers/specs/2026-05-27-compatibility-scoring-formula-v2-design.md` 简化版（去掉文件清单、tasks 等已被本 change 覆盖的部分），保留：

- 算法规范（§3 全部）
- 输出 Schema（§4 字段对照）
- consulting / duration / strategy 改写（§5 总结）

把 §1–2 改成简短 "Why this design" 节，引用本 proposal。具体内容约 100-150 行即可。

- [ ] **Step 3: 写 `tasks.md`**

```markdown
# Tasks

## 1. Backend algorithm core (compatibility_*.go)
- [ ] 1.1 Implement nayin 60-ganzhi lookup table + tests
- [ ] 1.2 Implement branchCompatible (liuhe + sanhe half) + tests
- [ ] 1.3 Implement scoreZodiac / scoreNayin / scoreDayPillar / scoreEightChars + tests
- [ ] 1.4 Implement evidence list + score_explanations + summary_tags generators + tests
- [ ] 1.5 Implement classifyRelationshipType + decision_advice + duration + stage_risks + strategy + tests
- [ ] 1.6 Rewrite AnalyzeCompatibility entry; delete 11 legacy build*Signals helpers
- [ ] 1.7 Integration tests covering perfect-hit / all-miss / threshold cases

## 2. Type & Service & DB
- [ ] 2.1 Rename DimensionScores fields + add OverallScore in model
- [ ] 2.2 Bump compatibilityAnalysisVersion to "v3"
- [ ] 2.3 Add migration 00011 (overall_score column + COMMENT)
- [ ] 2.4 Extend repository CreateCompatibilityReading / SELECTs to handle overall_score
- [ ] 2.5 Verify go build ./... and go test ./... pass

## 3. AI Prompt
- [ ] 3.1 Rewrite canonical_compatibility.go content for 4-module language
- [ ] 3.2 Update canonical_test.go / sync_test.go assertions

## 4. Frontend
- [ ] 4.1 Extend api.ts types with discriminated union v1/v2/v3
- [ ] 4.2 Add ScoreOverviewV3 component + CSS
- [ ] 4.3 Wire CompatibilityResultPage version dispatch
- [ ] 4.4 Wire CompatibilityHistoryPage version dispatch

## 5. OpenSpec
- [ ] 5.1 Validate spec deltas with /opsx validation
- [ ] 5.2 Archive after implementation merges
```

- [ ] **Step 4: 写 `specs/compatibility-scoring-formula/spec.md`**

```markdown
# compatibility-scoring-formula Specification

## ADDED Requirements

### Requirement: Four-module additive scoring
The compatibility engine SHALL compute a total score 0–100 as the sum of four independent module scores: zodiac (0–50), nayin (0–20), day_pillar (0–10), eight_chars (0–20).

#### Scenario: Module score ranges respected
- **WHEN** AnalyzeCompatibility runs on any pair of BaziResult inputs
- **THEN** each module score SHALL fall within its declared range
- **AND** the total score SHALL equal the arithmetic sum of the four module scores
- **AND** the total score SHALL be between 0 and 100 inclusive

#### Scenario: No negative contributions
- **WHEN** any combination of inputs is evaluated
- **THEN** no module SHALL deduct points
- **AND** unfavorable interactions (six-chong, six-hai, xing, ke nayin) SHALL contribute 0 instead of negative values

### Requirement: Zodiac module (year-zhi liuhe / sanhe only)
The zodiac module SHALL award 50 points when the two year-zhi form 六合 or 三合 (half-sanhe acceptable), otherwise 0.

#### Scenario: Year-zhi liuhe hit
- **WHEN** the year-zhi pair is one of {子丑, 寅亥, 卯戌, 辰酉, 巳申, 午未}
- **THEN** zodiac SHALL be 50

#### Scenario: Year-zhi half-sanhe hit
- **WHEN** the year-zhi pair is two distinct members of one sanhe group {申子辰, 亥卯未, 巳酉丑, 寅午戌}
- **THEN** zodiac SHALL be 50

#### Scenario: No hit (including double-生, six-chong, six-hai, xing)
- **WHEN** none of the above hits, even when 五行相生/同 (双生) applies
- **THEN** zodiac SHALL be 0

### Requirement: Nayin module (year-pillar nayin element)
The nayin module SHALL award 20 points when the two year-pillar nayin 五行 are 相生 or 相同, otherwise 0.

#### Scenario: Nayin shengsi or same
- **WHEN** nayin elements have a 生 or 同 relationship
- **THEN** nayin SHALL be 20

#### Scenario: Nayin ke
- **WHEN** nayin elements have a 克 relationship
- **THEN** nayin SHALL be 0

### Requirement: Day pillar module
The day_pillar module SHALL award 10 (upper tier) when the day-zhi pair is 支合 AND the day-gan pair is 干合 or 干相生; 5 (lower tier) when 支合 alone (regardless of 干 relation among same/克/无关); 0 when 支不合.

#### Scenario: Upper tier — gan five-合 + zhi 合
- **WHEN** day-gan pair is in 天干五合 set {甲己, 乙庚, 丙辛, 丁壬, 戊癸} AND day-zhi pair is 支合
- **THEN** day_pillar SHALL be 10

#### Scenario: Upper tier — gan 五行 sheng + zhi 合
- **WHEN** the two day-gan 五行 stand in a 相生 relation (excluding identity) AND day-zhi pair is 支合
- **THEN** day_pillar SHALL be 10

#### Scenario: Lower tier — zhi 合 alone (gan 同/克/无关)
- **WHEN** day-zhi pair is 支合 AND no upper-tier 干 condition met
- **THEN** day_pillar SHALL be 5

#### Scenario: zhi 不合
- **WHEN** day-zhi pair is not 支合
- **THEN** day_pillar SHALL be 0, regardless of any 干 relation

### Requirement: Eight-chars module (year/month/hour aggregation)
The eight_chars module SHALL score each of the three non-day pillar pairs (year/year, month/month, hour/hour) by the day_pillar rule, sum the three results (0–30) and normalize to [0, 20] via (sum × 2 + 1) / 3.

#### Scenario: All three pillars upper tier
- **WHEN** year/month/hour pillar pairs each score 10
- **THEN** eight_chars SHALL be 20

#### Scenario: One pillar upper tier, two not compatible
- **WHEN** sum is 10
- **THEN** eight_chars SHALL be 7

### Requirement: Overall level threshold
The overall_level SHALL map from total score: ≥ 80 → "high"; 60–79 → "medium"; < 60 → "low".

#### Scenario: Boundary at 80
- **WHEN** total score is exactly 80
- **THEN** overall_level SHALL be "high"

#### Scenario: Boundary at 60
- **WHEN** total score is exactly 60
- **THEN** overall_level SHALL be "medium"

#### Scenario: Boundary at 59
- **WHEN** total score is 59
- **THEN** overall_level SHALL be "low"

### Requirement: Evidence list only contains positive hits
The evidences list SHALL contain at most six items: one each for zodiac/nayin/day_pillar hits and up to three for eight_chars (year, month, hour). All evidence polarity SHALL be "positive"; unhit modules contribute no evidence.

#### Scenario: All four modules score positive
- **WHEN** zodiac > 0 AND nayin > 0 AND day_pillar > 0 AND all three eight_chars sub-pillars hit
- **THEN** evidences SHALL contain exactly 6 entries

#### Scenario: No module hits
- **WHEN** all four module scores are 0
- **THEN** evidences SHALL be an empty list

### Requirement: Analysis version tag
The analysis_version field of new CompatibilityReading records written by this engine SHALL be "v3"; v1/v2 records remain unchanged and renderable through the legacy frontend path.

#### Scenario: New reading written
- **WHEN** CreateCompatibilityReading runs after this change
- **THEN** the stored row SHALL have analysis_version = "v3"

#### Scenario: Legacy record read
- **WHEN** a v1 or v2 record is fetched
- **THEN** the API SHALL surface its dimension_scores with the original 4-dim keys (attraction/stability/communication/practicality)
- **AND** overall_score SHALL be 0 for these legacy records
```

- [ ] **Step 5: 校验 OpenSpec**

Run（项目根）: `cd openspec && /opsx-validate compatibility-scoring-formula-v3 --strict` 或 `npx openspec validate`（参考项目脚本）
Expected: 校验通过

- [ ] **Step 6: Commit**

```bash
git add openspec/changes/compatibility-scoring-formula-v3/
git commit -m "spec(compatibility): add v3 scoring-formula change

Captures the new 100-point classical formula as an OpenSpec change with
proposal/design/tasks/specs deltas. Two prior specs (explainable-scoring,
depth-signal-engine) are marked for archive once implementation merges.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>"
```

---

### Task 25: 归档旧 spec

**Files:**
- Move: `openspec/specs/compatibility-explainable-compatibility-scoring/` → archive
- Move: `openspec/specs/compatibility-depth-signal-engine/` → archive

- [ ] **Step 1: 通过 opsx-archive 工作流（推荐）**

如果项目脚本 `/opsx-archive` 支持旧 spec 归档，直接运行：

```bash
# 仅举例，按项目实际命令调整
/opsx-archive compatibility-explainable-compatibility-scoring
/opsx-archive compatibility-depth-signal-engine
```

否则手动：

```bash
git mv openspec/specs/compatibility-explainable-compatibility-scoring/spec.md \
       openspec/changes/archive/2026-05-27-compatibility-explainable-compatibility-scoring/spec.md
git mv openspec/specs/compatibility-depth-signal-engine/spec.md \
       openspec/changes/archive/2026-05-27-compatibility-depth-signal-engine/spec.md
```

> **本 task 仅在前述 24 个 task 全部 merge 后执行**，否则会破坏 OpenSpec 校验。

- [ ] **Step 2: Commit**

```bash
git add openspec/
git commit -m "spec(compatibility): archive legacy explainable-scoring & depth-signal specs

These two specs described the evidence-weighted 4-dimension model that
v3 replaces. Implementation is complete and the v3 change supersedes
them.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>"
```

---

## Self-Review Notes

1. **Task 9 顺序依赖**已在 Task 9 顶部明确：实际执行顺序为 1, 2, 3, 4, 5, 6, 7, 8, **15**, 9, 10, 11, 12, 13, 14, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25。
2. **Task 17 持久化** OverallScore 需要 DB 迁移先跑——执行 plan 时 `docker-compose exec` 命令按项目实际入口调整。
3. **Frontend 启动验证** Task 22/23 要求实际浏览器看 v3 渲染；如果项目无可视化沙箱，可改成手写 fixture 跑 Vitest（项目目前无前端单元测试框架，故选浏览器手测）。
4. **OpenSpec 校验命令** Task 24/25 按项目 `openspec/` 下实际工具链调整（grep 项目 README 或 `.claude/commands/opsx-*` 查具体命令）。

---

## 执行选择

Plan 已写入并即将提交。两种执行模式：

**1. Subagent-Driven（推荐）** - 我每个 task dispatch 一个 fresh subagent，task 间做 review，快速迭代。
**2. Inline Execution** - 用 executing-plans 在当前会话中分批执行，每个 checkpoint 我会暂停让你审查。

完成 plan 提交后请告诉我选哪种。
