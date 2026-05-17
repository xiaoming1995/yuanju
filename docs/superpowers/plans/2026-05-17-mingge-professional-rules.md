# 命格取格规则·命师口径 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 整体重写 `backend/pkg/bazi/mingge.go::DetectMingGe`，按命师 6 条隐性规则取格。验证集 12 case 全过。

**Architecture:** 抛弃七优先级表。新算法收集四柱透干候选，按通根强度（月支主气=6 → 余气=4 → 他支主气=3 → 余气=1）评分，叠加 4 条隐性规则（月禄/月刃 special、剔比劫、食伤同透取伤、地支气势全土立财）。Spec：`docs/superpowers/specs/2026-05-17-mingge-professional-rules-design.md`

**Tech Stack:** Go 1.25，`testing` 包，`go test ./pkg/bazi/...`

---

## File Structure

| 文件 | 改动类型 | 说明 |
|---|---|---|
| `backend/pkg/bazi/mingge.go` | 整体重写 `DetectMingGe`；新增 `linGuanZhi`、`diWangZhi`、`yangGans`、`rootStrength`；保留 `zhiHideGanFull`、`wuxingMainGan`、`ganWuxingMap`、`wuxingKe`、`isKeWuxing`、`shiShenToGeName`、`minggeDescDict`；删除旧 `gansContains`/`countGanInGans`/`wuxingScore`；新增 `minggeDescDict["月禄格"]` 条目 |
| `backend/pkg/bazi/mingge_test.go` | 新建（main 当前无此文件）| 12 case 表驱动测试 + helper 单测 |

`mingge.go` 现 295 行，重写后预计 ~250-280 行（删除大段七优先级逻辑，新增更紧凑的 6 规则逻辑）。

---

## Task 0：分支准备

**Files:**
- 工作目录创建分支

- [ ] **Step 1: 确认 main 干净**

```bash
cd /Users/liujiming/web/yuanju
git status --short
git branch --show-current
git log --oneline -1
```

Expected: 工作树净；分支 `main`；HEAD 是已 push 的 spec/plan/deprecation commits（即将提交本 plan）。

- [ ] **Step 2: 创建并切换分支**

```bash
git checkout -b feat/mingge-professional-rules
```

Expected: `Switched to a new branch 'feat/mingge-professional-rules'`

---

## Task 1：新增 helper 表 + 函数（先建底盘、不动 DetectMingGe）

**Files:**
- Modify: `backend/pkg/bazi/mingge.go`（在文件顶部 var 区新增）
- Create: `backend/pkg/bazi/mingge_test.go`

- [ ] **Step 1: 在 mingge.go 文件级 var 区新增 3 个表**

打开 `/Users/liujiming/web/yuanju/backend/pkg/bazi/mingge.go`，找到 `wuxingKe` map 声明的结束位置（即 `}` 后），在它之后插入：

```go

// linGuanZhi 日干 → 临官地支（月禄格判定用）
var linGuanZhi = map[string]string{
	"甲": "寅",
	"乙": "卯",
	"丙": "巳",
	"丁": "午",
	"戊": "巳",
	"己": "午",
	"庚": "申",
	"辛": "酉",
	"壬": "亥",
	"癸": "子",
}

// diWangZhi 日干 → 帝旺地支（月刃格判定用，仅阳干）
var diWangZhi = map[string]string{
	"甲": "卯",
	"丙": "午",
	"戊": "午",
	"庚": "酉",
	"壬": "子",
}

// yangGans 阳干集合（月刃格判定用）
var yangGans = map[string]bool{
	"甲": true,
	"丙": true,
	"戊": true,
	"庚": true,
	"壬": true,
}
```

- [ ] **Step 2: 在 mingge.go 顶部 minggeDescDict 中新增 "月禄格" 条目**

在 `minggeDescDict` map 内（已有"建禄格"条目）追加一条：

```go
	"月禄格": "月禄格即月支临官，日主之气聚于月令，身强有力。月禄格之人自立心强，独立自主，做事靠自身努力，不依赖他人。月禄格命主意志坚定，韧性十足，适合自主创业或独立执业。月禄格不喜比劫过多争财，需有官杀来疏导旺气。身强用官杀或财星，则功名利禄皆可期；整体而言是奋斗型命格，努力必有回报。",
```

（释义文字与"建禄格"等价 —— 二者传统上为同义命名）

- [ ] **Step 3: 在 mingge.go 文件底部（DetectMingGe 之后或之前不重要，函数级即可）新增 `rootStrength` 函数**

```go
// rootStrength 计算 gan 在 r 四柱地支藏干中能找到的最强单根分
//
// 评分表：
//   月支主气=6, 月支中气=5, 月支余气=4
//   他支主气=3, 他支中气=2, 他支余气=1
//   无根=0
//
// 返回 4 柱中找到的最大分。如果 gan 同时在月支和他支扎根，取月支根（分高）。
func rootStrength(gan string, r *BaziResult) int {
	posZhis := []struct {
		isMonth bool
		zhi     string
	}{
		{true, r.MonthZhi},
		{false, r.YearZhi},
		{false, r.DayZhi},
		{false, r.HourZhi},
	}

	best := 0
	for _, pz := range posZhis {
		hgs := zhiHideGanFull[pz.zhi]
		for i, hg := range hgs {
			if hg != gan {
				continue
			}
			var s int
			if pz.isMonth {
				s = 6 - i // 主气=6, 中气=5, 余气=4
			} else {
				s = 3 - i // 主气=3, 中气=2, 余气=1
			}
			if s > best {
				best = s
			}
		}
	}
	return best
}
```

- [ ] **Step 4: 创建 `mingge_test.go` 仅含 helper 单测，先验证底盘正确**

写入 `/Users/liujiming/web/yuanju/backend/pkg/bazi/mingge_test.go`：

```go
package bazi

import "testing"

func TestLinGuanZhi(t *testing.T) {
	cases := map[string]string{
		"甲": "寅", "乙": "卯", "丙": "巳", "丁": "午",
		"戊": "巳", "己": "午", "庚": "申", "辛": "酉",
		"壬": "亥", "癸": "子",
	}
	for gan, want := range cases {
		if got := linGuanZhi[gan]; got != want {
			t.Errorf("linGuanZhi[%q] = %q, want %q", gan, got, want)
		}
	}
}

func TestDiWangZhi_OnlyYangGans(t *testing.T) {
	cases := map[string]string{
		"甲": "卯", "丙": "午", "戊": "午", "庚": "酉", "壬": "子",
	}
	for gan, want := range cases {
		if got := diWangZhi[gan]; got != want {
			t.Errorf("diWangZhi[%q] = %q, want %q", gan, got, want)
		}
	}
	// 阴干不应出现在表中
	for _, yin := range []string{"乙", "丁", "己", "辛", "癸"} {
		if _, ok := diWangZhi[yin]; ok {
			t.Errorf("diWangZhi should not contain 阴干 %q", yin)
		}
	}
}

func TestRootStrength_MonthBranchMainQi(t *testing.T) {
	// 月支寅, 主气甲 → 甲在月支主气, 根强度 = 6
	r := &BaziResult{
		YearGan: "辛", YearZhi: "酉",
		MonthGan: "丙", MonthZhi: "寅",
		DayGan: "庚", DayZhi: "子",
		HourGan: "辛", HourZhi: "酉",
	}
	if got := rootStrength("甲", r); got != 6 {
		t.Errorf("rootStrength(甲) = %d, want 6 (月支主气)", got)
	}
	// 丙 在月支寅中气 → 5
	if got := rootStrength("丙", r); got != 5 {
		t.Errorf("rootStrength(丙) = %d, want 5 (月支中气)", got)
	}
	// 戊 在月支寅余气 → 4
	if got := rootStrength("戊", r); got != 4 {
		t.Errorf("rootStrength(戊) = %d, want 4 (月支余气)", got)
	}
	// 辛 在年支/时支酉主气 → 3 (取最强的他支主气)
	if got := rootStrength("辛", r); got != 3 {
		t.Errorf("rootStrength(辛) = %d, want 3 (他支主气)", got)
	}
	// 癸 在日支子主气 → 3
	if got := rootStrength("癸", r); got != 3 {
		t.Errorf("rootStrength(癸) = %d, want 3", got)
	}
	// 庚 = 月干本身但月支寅不藏庚, 其他支也不藏 → 0
	if got := rootStrength("庚", r); got != 0 {
		t.Errorf("rootStrength(庚) = %d, want 0 (无根)", got)
	}
}
```

- [ ] **Step 5: 跑 helper 单测**

```bash
cd /Users/liujiming/web/yuanju/backend
go test ./pkg/bazi/ -run "TestLinGuanZhi|TestDiWangZhi|TestRootStrength" -v
```

Expected: 3 个测试全 PASS。

- [ ] **Step 6: 跑包内全部测试确认无回归**

```bash
go test ./pkg/bazi/...
```

Expected: `ok`。注意：此时旧 `DetectMingGe` 仍然完整，所有调用方应不受影响。

- [ ] **Step 7: 提交**

```bash
cd /Users/liujiming/web/yuanju
git add backend/pkg/bazi/mingge.go backend/pkg/bazi/mingge_test.go
git commit -m "$(cat <<'EOF'
feat(bazi/mingge): add helper tables and rootStrength for new algorithm

Adds the foundational data structures for the command-rule mingge rewrite:
- linGuanZhi: day-stem → 临官地支 (for 月禄格 detection)
- diWangZhi: day-stem → 帝旺地支 (阳干 only; for 月刃格 detection)
- yangGans: set of 阳干 (for 月刃格 gating)
- rootStrength(gan, r): scores 通根 strength
    月支主气=6, 中气=5, 余气=4
    他支主气=3, 中气=2, 余气=1
    无根=0
- minggeDescDict["月禄格"]: traditional synonym for 建禄格

The old DetectMingGe is left untouched. Next commit replaces it.

Spec: docs/superpowers/specs/2026-05-17-mingge-professional-rules-design.md
EOF
)"
```

---

## Task 2：写 12 case 测试（TDD RED）

**Files:**
- Modify: `backend/pkg/bazi/mingge_test.go`（追加表驱动测试函数）

- [ ] **Step 1: 在 mingge_test.go 文件末尾追加测试函数**

```go
func TestDetectMingGe_ProfessionalRules(t *testing.T) {
	tests := []struct {
		name   string
		result *BaziResult
		wantGe string
	}{
		// 规则 3：月支中气根强，单候选立格
		{
			name: "C1 男 1995-10-12 11时 → 偏印格",
			result: &BaziResult{
				YearGan: "乙", YearZhi: "亥",
				MonthGan: "丙", MonthZhi: "戌",
				DayGan: "丙", DayZhi: "子",
				HourGan: "甲", HourZhi: "午",
			},
			wantGe: "偏印格",
		},
		{
			name: "C2 男 1996-02-08 20时 → 伤官格",
			result: &BaziResult{
				YearGan: "丙", YearZhi: "子",
				MonthGan: "庚", MonthZhi: "寅",
				DayGan: "乙", DayZhi: "亥",
				HourGan: "丙", HourZhi: "戌",
			},
			wantGe: "伤官格",
		},
		// 规则 4：食伤同透取伤（C3 辛无根但强制立）
		{
			name: "C3 女 1991-02-07 16时 → 伤官格（食伤同透）",
			result: &BaziResult{
				YearGan: "辛", YearZhi: "未",
				MonthGan: "庚", MonthZhi: "寅",
				DayGan: "戊", DayZhi: "申",
				HourGan: "庚", HourZhi: "申",
			},
			wantGe: "伤官格",
		},
		// 规则 3：单候选 + 余气根
		{
			name: "C4 女 1997-12-01 12时 → 偏财格",
			result: &BaziResult{
				YearGan: "丁", YearZhi: "丑",
				MonthGan: "辛", MonthZhi: "亥",
				DayGan: "丁", DayZhi: "丑",
				HourGan: "丙", HourZhi: "午",
			},
			wantGe: "偏财格",
		},
		// 规则 5：单候选无根 → 杂气格
		{
			name: "C5 女 1988-01-18 04时 → 杂气格",
			result: &BaziResult{
				YearGan: "丁", YearZhi: "卯",
				MonthGan: "癸", MonthZhi: "丑",
				DayGan: "壬", DayZhi: "申",
				HourGan: "壬", HourZhi: "寅",
			},
			wantGe: "杂气格",
		},
		{
			name: "C6 女 1996-11-08 08时 → 杂气格",
			result: &BaziResult{
				YearGan: "丙", YearZhi: "子",
				MonthGan: "己", MonthZhi: "亥",
				DayGan: "己", DayZhi: "酉",
				HourGan: "戊", HourZhi: "辰",
			},
			wantGe: "杂气格",
		},
		// 规则 3：多候选按根强度选
		{
			name: "C7 女 1991-12-30 10时 → 正财格",
			result: &BaziResult{
				YearGan: "辛", YearZhi: "未",
				MonthGan: "庚", MonthZhi: "子",
				DayGan: "甲", DayZhi: "戌",
				HourGan: "己", HourZhi: "巳",
			},
			wantGe: "正财格",
		},
		// 规则 5：候选都无根 + 地支气势非财 → 杂气格
		{
			name: "C8 男 1996-12-16 22时 → 杂气格（地支全水 但非财）",
			result: &BaziResult{
				YearGan: "丙", YearZhi: "子",
				MonthGan: "庚", MonthZhi: "子",
				DayGan: "丁", DayZhi: "亥",
				HourGan: "辛", HourZhi: "亥",
			},
			wantGe: "杂气格",
		},
		// 规则 3：他支中气根 > 他支余气根
		{
			name: "C9 男 1995-01-23 16时 → 偏印格",
			result: &BaziResult{
				YearGan: "甲", YearZhi: "戌",
				MonthGan: "丁", MonthZhi: "丑",
				DayGan: "甲", DayZhi: "寅",
				HourGan: "壬", HourZhi: "申",
			},
			wantGe: "偏印格",
		},
		// 规则 3：月支中气根 > 他支中气根
		{
			name: "C10 女 1993-01-16 14时 → 七杀格",
			result: &BaziResult{
				YearGan: "壬", YearZhi: "申",
				MonthGan: "癸", MonthZhi: "丑",
				DayGan: "丁", DayZhi: "酉",
				HourGan: "丁", HourZhi: "未",
			},
			wantGe: "七杀格",
		},
		// 规则 3：月支中气根 > 他支主气根 > 无根
		{
			name: "C11 男 2015-02-02 18时40分 → 偏财格",
			result: &BaziResult{
				YearGan: "甲", YearZhi: "午",
				MonthGan: "丁", MonthZhi: "丑",
				DayGan: "己", DayZhi: "酉",
				HourGan: "癸", HourZhi: "酉",
			},
			wantGe: "偏财格",
		},
		// 规则 6：候选都无根 + 地支气势全土 = 乙日财 → 月支主气戊配乙阴 异性 → 正财格
		{
			name: "C12 男 1985-04-26 19时46分 → 正财格（地支气势全土）",
			result: &BaziResult{
				YearGan: "乙", YearZhi: "丑",
				MonthGan: "庚", MonthZhi: "辰",
				DayGan: "乙", DayZhi: "未",
				HourGan: "丙", HourZhi: "戌",
			},
			wantGe: "正财格",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotGe, _ := DetectMingGe(tt.result)
			if gotGe != tt.wantGe {
				t.Errorf("DetectMingGe got = %q, want %q", gotGe, tt.wantGe)
			}
		})
	}
}
```

- [ ] **Step 2: 跑测试验证大量 FAIL（RED）**

```bash
cd /Users/liujiming/web/yuanju/backend
go test ./pkg/bazi/ -run TestDetectMingGe_ProfessionalRules -v
```

Expected: 多个 case FAIL（因为 `DetectMingGe` 还是旧的七优先级算法）。具体来说：
- C5 / C6 / C8 / C9 / C12 应该 FAIL（旧算法立月刃/食神/月刃/伤官/月刃等 ≠ 命师标注）
- C3 应该 FAIL（旧立食神格而非伤官）
- C7 应该 FAIL（旧立七杀格 / 建禄格等）
- C11 应该 FAIL（旧立建禄格）

C1 / C2 / C4 / C10 旧算法刚好对齐 → PASS。

具体多少 PASS 多少 FAIL 是看旧算法对每个 case 的输出。**不一定**必须 5/12（之前 PR 实施时的测试），因为本任务运行在 **未经任何修复** 的 main 上的 mingge.go。

⚠️ 如果意外发现某 case 出乎意料 PASS，记下结果，不要 STOP —— 旧算法巧合通过不影响下一步替换算法。

- [ ] **Step 3: 不要 commit FAIL 中间状态**

下一个 task 实施新算法后再一起 commit 测试 + 实现。

---

## Task 3：整体重写 DetectMingGe（核心实现）

**Files:**
- Modify: `backend/pkg/bazi/mingge.go`（替换整个 `DetectMingGe` 函数 + 删除旧 helper）

- [ ] **Step 1: 删除旧的小 helper（仅旧 DetectMingGe 使用）**

在 mingge.go 中找到并 **完全删除** 以下函数和声明（它们只被旧 DetectMingGe 内部用，新算法不需要）：

```go
// shiShenToGeName … (保留, 还要用)
// gansContains, countGanInGans, wuxingScore … 删除
// detectSanHeHui … (保留, 备用)
```

执行删除：
- 删除 `gansContains` 函数（约 8 行）
- 删除 `countGanInGans` 函数（约 10 行）
- 删除 `wuxingScore` 函数（约 10 行）

- [ ] **Step 2: 替换整个 `DetectMingGe` 函数体**

找到现有 `DetectMingGe` 函数（应该是 `func DetectMingGe(r *BaziResult) (name, desc string) { ... }`），**整段**替换为：

```go
// DetectMingGe 按命师 6 规则取格
//
// 规则优先级：
//   0. 月禄/月刃 special case
//   1. 收集天干非比劫候选
//   4. 食伤同透 → 强制立伤官格
//   3. 通根强度排序 → 取最强者立格
//   6. 候选都无根 / 无候选 → 地支气势全土且为日干财 → 立财格
//   5. 兜底 → 杂气格
//
// Spec: docs/superpowers/specs/2026-05-17-mingge-professional-rules-design.md
func DetectMingGe(r *BaziResult) (name, desc string) {
	dayGan := r.DayGan

	// ── 规则 0: 月禄/月刃 special case ───────────────────────
	if linGuanZhi[dayGan] == r.MonthZhi {
		return "月禄格", minggeDescDict["月禄格"]
	}
	if yangGans[dayGan] && diWangZhi[dayGan] == r.MonthZhi {
		return "月刃格", minggeDescDict["月刃格"]
	}

	// ── 规则 1: 收集天干非比劫候选 ─────────────────────────
	type cand struct {
		gan     string
		shishen string
	}
	var candidates []cand
	for _, g := range []string{r.YearGan, r.MonthGan, r.HourGan} {
		ss := GetShiShen(dayGan, g)
		if ss == "比肩" || ss == "劫财" || ss == "" {
			continue
		}
		candidates = append(candidates, cand{g, ss})
	}

	// ── 规则 4: 食伤同透 → 强制立伤官格 ────────────────────
	hasFood, hasInjury := false, false
	for _, c := range candidates {
		if c.shishen == "食神" {
			hasFood = true
		}
		if c.shishen == "伤官" {
			hasInjury = true
		}
	}
	if hasFood && hasInjury {
		return "伤官格", minggeDescDict["伤官格"]
	}

	// ── 规则 3: 通根强度排序，取最强 ──────────────────────
	if len(candidates) > 0 {
		bestIdx := 0
		bestRoot := rootStrength(candidates[0].gan, r)
		for i := 1; i < len(candidates); i++ {
			root := rootStrength(candidates[i].gan, r)
			if root > bestRoot {
				bestRoot = root
				bestIdx = i
			}
		}
		if bestRoot > 0 {
			geName := shiShenToGeName(candidates[bestIdx].shishen)
			return geName, minggeDescDict[geName]
		}
	}

	// ── 规则 6: 地支气势全土且为日干财 → 立财格 ──────────
	zhiMains := []string{
		zhiHideGanFull[r.YearZhi][0],
		zhiHideGanFull[r.MonthZhi][0],
		zhiHideGanFull[r.DayZhi][0],
		zhiHideGanFull[r.HourZhi][0],
	}
	firstWx := ganWuxingMap[zhiMains[0]]
	allSameWx := firstWx != ""
	for _, mg := range zhiMains[1:] {
		if ganWuxingMap[mg] != firstWx {
			allSameWx = false
			break
		}
	}
	if allSameWx {
		dayWx := ganWuxingMap[dayGan]
		if isKeWuxing(dayWx, firstWx) {
			monthMainGan := zhiHideGanFull[r.MonthZhi][0]
			ss := GetShiShen(dayGan, monthMainGan)
			geName := shiShenToGeName(ss)
			return geName, minggeDescDict[geName]
		}
	}

	// ── 规则 5: 兜底 ────────────────────────────────────
	return "杂气格", minggeDescDict["杂气格"]
}
```

- [ ] **Step 3: 跑 12 case 测试**

```bash
cd /Users/liujiming/web/yuanju/backend
go test ./pkg/bazi/ -run TestDetectMingGe_ProfessionalRules -v
```

Expected: C1 到 C12 全部 PASS（12/12）。

如果有任何 FAIL，**STOP** 并报告 case 名 + 实际输出 vs 期望。不要私自调整逻辑掩盖。

- [ ] **Step 4: 跑 helper 单测确认未被打破**

```bash
go test ./pkg/bazi/ -run "TestLinGuanZhi|TestDiWangZhi|TestRootStrength" -v
```

Expected: 全 PASS。

- [ ] **Step 5: 跑包内全部测试**

```bash
go test ./pkg/bazi/...
```

Expected: `ok`。

如果有 *其他* mingge 相关测试 FAIL（非本 task 的 12 case 测试），那是命师纠错型变化，**记录变化**但不在本 task 处理。

- [ ] **Step 6: 跑全 backend 回归测试**

```bash
go test ./...
```

Expected: 全 `ok`。

特别检查 `internal/service/report_service_test.go:254`（用 1996-02-08 20时，即 C2，命师标注伤官格）—— 新算法也应输出伤官格。如果该测试期望值不一致，**STOP 并报告**。

- [ ] **Step 7: build + vet 验证**

```bash
go build ./...
go vet ./...
```

Expected: 静默通过。

- [ ] **Step 8: 提交**

```bash
cd /Users/liujiming/web/yuanju
git add backend/pkg/bazi/mingge.go backend/pkg/bazi/mingge_test.go
git commit -m "$(cat <<'EOF'
feat(bazi/mingge): rewrite DetectMingGe with professional rules

Rewrite the entire DetectMingGe function to use the 6-rule algorithm
documented in the design spec. This replaces the previous 7-priority
algorithm (which had ~5/12 agreement with professional fortune-teller
labels) with one that achieves 12/12 on the validation set.

Algorithm:
- Rule 0: 月禄格 (月支为日干临官) / 月刃格 (阳干日 + 月支为帝旺)
- Rule 1: collect non-bi/jie candidates from year/month/hour stems
- Rule 4: 食 + 伤 both transparent → force 伤官格
- Rule 3: rank candidates by 通根 strength (月支主气=6 ... 余气=4, 他支
  主气=3 ... 余气=1); strongest wins iff it has any root
- Rule 6: if no rooted candidate, but all 4 地支主气 are 同五行 = 日干
  财 → 立 month-主气 → 财格
- Rule 5: fallback → 杂气格

Removes obsolete helpers (gansContains, countGanInGans, wuxingScore).
Keeps detectSanHeHui and zhiHideGanFull for future use.

Spec: docs/superpowers/specs/2026-05-17-mingge-professional-rules-design.md
EOF
)"
```

---

## Task 4：最终验证 + 推送

**Files:** 无（git 操作）

- [ ] **Step 1: 跑最终完整测试**

```bash
cd /Users/liujiming/web/yuanju/backend
go test ./pkg/bazi/ -v -count=3
go test ./...
go vet ./...
go build ./...
```

Expected: 全 PASS 稳定（`count=3` 多次结果一致 —— 新算法 deterministic）。

- [ ] **Step 2: 检查分支状态**

```bash
cd /Users/liujiming/web/yuanju
git log --oneline main..HEAD
git diff main..HEAD --stat -- backend/
```

Expected: 看到 2 个 commit（helper、rewrite），diff 限于 `backend/pkg/bazi/mingge.go` 和 `mingge_test.go`。

- [ ] **Step 3: 推送分支**

```bash
git push -u origin feat/mingge-professional-rules
```

Expected: 远程新建分支。

- [ ] **Step 4: 报告完成**

向人类汇报：
- 分支：`feat/mingge-professional-rules`
- 2 个 commit
- `mingge.go` 净变化 ~XX 行（删除旧逻辑、新增 helper + 重写 DetectMingGe）
- `mingge_test.go` 新建，~250 行（helper 单测 + 12 case 验证）
- 12/12 case 全过

人类决定 PR 创建 / 合并。

---

## Self-Review（执行 agent 必读）

实施过程中可能踩的坑：

1. **不要保留旧 DetectMingGe**：Task 3 Step 2 是 **整段替换**，不是叠加。如果误把新代码追加到旧函数后面，会导致重复定义编译错。

2. **`gansContains` 等 helper 可能被文件外其他代码引用** —— 删之前先 `grep` 确认：
   ```bash
   grep -rn "gansContains\|countGanInGans\|wuxingScore" backend/pkg/bazi/
   ```
   如有其他文件引用，**STOP** 并报告。Spec 假定只 `mingge.go` 内部用。

3. **`minggeDescDict["月禄格"]` 必须先加（Task 1 Step 2）**，否则 Task 3 跑到 C7/C10 走到月禄路径时返回空 desc。但当前 12 case 中没有月禄盘命中，C7 走规则 3、C10 走规则 3 —— 不会触发月禄路径。仍然加，因为生产环境其他盘可能触发。

4. **C5 / C6 / C8 期望 = "杂气格"**：命师口径"不成格" = 算法上的"杂气格"。测试用 "杂气格" 字面匹配，不是"不成格"。

5. **C11 / C12 使用 hour 部分（18 / 19）** —— 不要试图加 minute 字段；`BaziResult` 没有这个字段。

6. **Task 2 RED 阶段不需要 commit** —— 直接进入 Task 3 实施 GREEN。Task 3 一并 commit 测试 + 实现。

7. **多次跑 `-count=3` 确认 deterministic** —— 新算法无 `map` 遍历影响输出（候选用 slice 顺序，规则 3 取最强单点）。

8. **如发现 `report_service_test.go:254` 那个 1996-02-08 20时 + male 的测试期望值不再对齐**，stop 并报告。预期新算法也输出伤官格（与该测试一致）。
