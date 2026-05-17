# 命格定格规则修正（方案 B）Implementation Plan

> ## ⚠️ 状态：已实施但已废弃（SUPERSEDED）
>
> 此 plan 已被完整执行（5 个 commit 在 feat/mingge-rule-fix-b 分支上）。但实施后用 12 个专业命师标注的真实盘对照，发现方向偏离 —— 对齐率仅 5/12。分支已删除，未合并 main。
>
> **接替本 plan**：[`2026-05-17-mingge-professional-rules.md`](./2026-05-17-mingge-professional-rules.md) — 整体重写 `DetectMingGe`，实现命师 6 条隐性规则。
>
> 本文件保留以记录失败的实施路径。

---

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 修复 `backend/pkg/bazi/mingge.go::DetectMingGe` 中 4 处与标准七优先级取格法的偏差（第 1 / 3 / 6 / 7 层），新增 `mingge_test.go` 覆盖测试。

**Architecture:** 单文件算法纠错。不改公共接口、不改 DB、不影响前端。新增包内常量 / helper（五行相克表、克日干判断、凶吉神顺序、平局决胜）。Spec：`docs/superpowers/specs/2026-05-17-mingge-rule-fix-design.md`

**Tech Stack:** Go 1.21+，`testing` 包，`go test ./pkg/bazi/...`

---

## File Structure

| 文件 | 改动类型 | 说明 |
|---|---|---|
| `backend/pkg/bazi/mingge.go` | Modify | 替换第 1 / 3 / 6 / 7 层实现 + 新增 5 个文件级常量 / helper |
| `backend/pkg/bazi/mingge_test.go` | Create | 覆盖 12 个 case 的表驱动测试 |

`mingge.go` 当前 295 行，预计修改后 ~360 行（远低于 500 行限制）。

---

## Task 0：分支准备

**Files:**
- Modify: 工作目录（创建分支）

- [ ] **Step 1: 确认在 main 干净状态**

```bash
cd /Users/liujiming/web/yuanju
git status --short
git branch --show-current
```

Expected: 无 staged/unstaged 改动；当前分支 `main`；HEAD = `6925b41`（spec commit）。

- [ ] **Step 2: 创建并切换到新分支**

```bash
git checkout -b feat/mingge-rule-fix-b
```

Expected: `Switched to a new branch 'feat/mingge-rule-fix-b'`

---

## Task 1：第 1 层修复（月支三气透月干）+ 测试脚手架

**Files:**
- Create: `backend/pkg/bazi/mingge_test.go`
- Modify: `backend/pkg/bazi/mingge.go:166-173`

- [ ] **Step 1: 创建测试文件（含 C1 / C2 / C3）**

`backend/pkg/bazi/mingge_test.go`:

```go
package bazi

import "testing"

func TestDetectMingGe(t *testing.T) {
	tests := []struct {
		name   string
		result *BaziResult
		wantGe string
	}{
		// ── 第 1 层：月支三气透月干 ────────────────────────
		{
			name: "C1 月支主气透月干 → 建禄格",
			result: &BaziResult{
				YearGan: "辛", YearZhi: "酉",
				MonthGan: "甲", MonthZhi: "寅",
				DayGan: "甲", DayZhi: "子",
				HourGan: "辛", HourZhi: "酉",
			},
			wantGe: "建禄格",
		},
		{
			name: "C2 月支中气透月干 → 七杀格（修 ①）",
			result: &BaziResult{
				YearGan: "辛", YearZhi: "酉",
				MonthGan: "丙", MonthZhi: "寅",
				DayGan: "庚", DayZhi: "子",
				HourGan: "辛", HourZhi: "酉",
			},
			wantGe: "七杀格",
		},
		{
			name: "C3 月支余气透月干 → 偏财格（修 ①）",
			result: &BaziResult{
				YearGan: "辛", YearZhi: "酉",
				MonthGan: "戊", MonthZhi: "寅",
				DayGan: "甲", DayZhi: "子",
				HourGan: "辛", HourZhi: "酉",
			},
			wantGe: "偏财格",
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

- [ ] **Step 2: 跑测试看预期失败**

```bash
cd /Users/liujiming/web/yuanju/backend
go test ./pkg/bazi/ -run TestDetectMingGe -v
```

Expected:
- `C1 月支主气透月干 → 建禄格`: PASS（当前实现支持主气透月干）
- `C2 月支中气透月干 → 七杀格（修 ①）`: **FAIL**（当前实现只查主气，会走到第 4 层，立月刃格）
- `C3 月支余气透月干 → 偏财格（修 ①）`: **FAIL**（同上，当前实现立建禄格）

- [ ] **Step 3: 修改 mingge.go 第 1 层 —— 月支三气循环**

打开 `backend/pkg/bazi/mingge.go`，替换第 166-173 行：

```go
	// ── 优先级 1：月支主气透月干（月柱自透）────────────────────
	if len(monthHideGans) > 0 {
		mainGan := monthHideGans[0] // 主气
		if mainGan == monthGan {
			ss := GetShiShen(dayGan, mainGan)
			geName := shiShenToGeName(ss)
			return geName, minggeDescDict[geName]
		}
	}
```

替换为：

```go
	// ── 优先级 1：月支主 / 中 / 余气 任一透月干（月柱自透）────────
	//   命中即停 → 隐含「主气 > 中气 > 余气」决胜（藏干表已按主→中→余排序）
	if len(monthHideGans) > 0 {
		for _, hg := range monthHideGans {
			if hg == monthGan {
				ss := GetShiShen(dayGan, hg)
				geName := shiShenToGeName(ss)
				return geName, minggeDescDict[geName]
			}
		}
	}
```

- [ ] **Step 4: 再跑测试，应全绿**

```bash
go test ./pkg/bazi/ -run TestDetectMingGe -v
```

Expected: C1 / C2 / C3 全部 `--- PASS`。

- [ ] **Step 5: 跑包内全部测试，确保无回归**

```bash
go test ./pkg/bazi/...
```

Expected: `ok  yuanju/pkg/bazi`。如果有其它历史测试用到 `DetectMingGe` 且因为本次纠错而改变了输出，那是**期望中的纠错**，需逐 case 评估；若是其它模块测试挂了则是回归。

- [ ] **Step 6: 提交**

```bash
git add backend/pkg/bazi/mingge.go backend/pkg/bazi/mingge_test.go
git commit -m "$(cat <<'EOF'
fix(bazi/mingge): layer 1 includes 主气/中气/余气 transparency

Previously only month-branch 主气 透 month-stem could form the 月柱自透
case (priority 1). Now 主气/中气/余气 are all checked — first match wins,
which naturally enforces 主气 > 中气 > 余气 precedence via
zhiHideGanFull's main→middle→余 ordering.

Adds mingge_test.go with table-driven cases C1/C2/C3 covering the three
transparency paths.

Spec: docs/superpowers/specs/2026-05-17-mingge-rule-fix-design.md
EOF
)"
```

Expected: 1 commit created, working tree clean.

---

## Task 2：第 3 层修复（他支三气透月干）

**Files:**
- Modify: `backend/pkg/bazi/mingge_test.go`（追加 C4 / C5）
- Modify: `backend/pkg/bazi/mingge.go:207-218`

- [ ] **Step 1: 追加 C4 / C5 到测试 slice**

在 `mingge_test.go` 的 `tests := []struct{...}{...}` slice 里，C3 之后追加：

```go
		// ── 第 3 层：他支三气透月干 ────────────────────────
		{
			name: "C4 他支主气透月干 → 七杀格（基线）",
			result: &BaziResult{
				YearGan: "庚", YearZhi: "寅",
				MonthGan: "甲", MonthZhi: "未",
				DayGan: "戊", DayZhi: "申",
				HourGan: "辛", HourZhi: "酉",
			},
			wantGe: "七杀格",
		},
		{
			name: "C5 他支中气透月干 → 正官格（修 ③）",
			result: &BaziResult{
				YearGan: "辛", YearZhi: "寅",
				MonthGan: "丙", MonthZhi: "子",
				DayGan: "辛", DayZhi: "酉",
				HourGan: "辛", HourZhi: "未",
			},
			wantGe: "正官格",
		},
```

- [ ] **Step 2: 跑测试**

```bash
go test ./pkg/bazi/ -run TestDetectMingGe -v
```

Expected:
- C4 `→ 七杀格（基线）`: PASS（当前实现的第 3 层已能命中他支主气）
- C5 `→ 正官格（修 ③）`: **FAIL**（当前实现立建禄格，因为走到第 4 层取辛打分）

- [ ] **Step 3: 修改 mingge.go 第 3 层 —— 跨支按主中余气分层**

替换第 207-218 行（决策点 A：先扫所有支的主气、再中气、再余气）：

```go
	// ── 优先级 3：其它三柱地支主气透月干 ────────────────────────
	for _, oz := range otherZhis {
		hgs := zhiHideGanFull[oz]
		if len(hgs) == 0 {
			continue
		}
		mainGan := hgs[0]
		if mainGan == monthGan {
			ss := GetShiShen(dayGan, mainGan)
			geName := shiShenToGeName(ss)
			return geName, minggeDescDict[geName]
		}
	}
```

替换为：

```go
	// ── 优先级 3：其它三柱地支主 / 中 / 余气透月干 ────────────────
	//   按 主气 → 中气 → 余气 跨支扫描（先把三支的主气都扫完，再扫中气，最后余气）
	//   所有支最多 3 个藏干，外层固定 depth < 3
	for depth := 0; depth < 3; depth++ {
		for _, oz := range otherZhis {
			hgs := zhiHideGanFull[oz]
			if depth >= len(hgs) {
				continue
			}
			if hgs[depth] == monthGan {
				ss := GetShiShen(dayGan, hgs[depth])
				geName := shiShenToGeName(ss)
				return geName, minggeDescDict[geName]
			}
		}
	}
```

- [ ] **Step 4: 跑测试应全绿**

```bash
go test ./pkg/bazi/ -run TestDetectMingGe -v
```

Expected: C1 / C2 / C3 / C4 / C5 全部 PASS。

- [ ] **Step 5: 跑包内全部测试**

```bash
go test ./pkg/bazi/...
```

Expected: `ok`。

- [ ] **Step 6: 提交**

```bash
git add backend/pkg/bazi/mingge.go backend/pkg/bazi/mingge_test.go
git commit -m "$(cat <<'EOF'
fix(bazi/mingge): layer 3 scans 主/中/余气 across other branches

Layer 3 (他支透月干) previously only checked each other branch's 主气.
Now scans 主气 across all three branches first, then 中气 across all
three, then 余气 — implementing decision A (主气 > 中气 > 余气 priority
spans across branches).

Adds C4 (baseline 主气 match) and C5 (中气 match — the new fix point).

Spec: docs/superpowers/specs/2026-05-17-mingge-rule-fix-design.md
EOF
)"
```

---

## Task 3：第 6 层修复（合化局须克日主才立格）

**Files:**
- Modify: `backend/pkg/bazi/mingge_test.go`（追加 C6 / C7 / C8）
- Modify: `backend/pkg/bazi/mingge.go`：
  - 文件级新增 `wuxingKe` map + `isKeWuxing` helper
  - 第 256-262 行加克日主过滤

- [ ] **Step 1: 追加 C6 / C7 / C8**

在 mingge_test.go tests slice 中追加：

```go
		// ── 第 6 层：合化局须克日主才立格 ────────────────────
		{
			name: "C6 巳酉丑三合金 + 甲日 → 七杀格（克日主，立格）",
			result: &BaziResult{
				YearGan: "乙", YearZhi: "巳",
				MonthGan: "壬", MonthZhi: "酉",
				DayGan: "甲", DayZhi: "丑",
				HourGan: "乙", HourZhi: "午",
			},
			wantGe: "七杀格",
		},
		{
			name: "C7 申子辰三合水 + 甲日 → 杂气格（水生木不克，过滤）（修 ⑥）",
			result: &BaziResult{
				YearGan: "丙", YearZhi: "申",
				MonthGan: "丙", MonthZhi: "子",
				DayGan: "甲", DayZhi: "辰",
				HourGan: "辛", HourZhi: "午",
			},
			wantGe: "杂气格",
		},
		{
			name: "C8 巳午未三会火 + 甲日 → 杂气格（火被木生不克，过滤）（修 ⑥）",
			result: &BaziResult{
				YearGan: "壬", YearZhi: "巳",
				MonthGan: "壬", MonthZhi: "子",
				DayGan: "甲", DayZhi: "午",
				HourGan: "壬", HourZhi: "未",
			},
			wantGe: "杂气格",
		},
```

- [ ] **Step 2: 跑测试**

```bash
go test ./pkg/bazi/ -run TestDetectMingGe -v
```

Expected:
- C6 PASS（当前实现已立七杀格，金克木）
- C7 **FAIL**（当前立偏印格，因不判克日主）
- C8 **FAIL**（当前立食神格）

- [ ] **Step 3: 在 mingge.go 文件级（紧邻 `wuxingMainGan` 之后，第 52 行后）新增**

```go
// wuxingKe 五行相克：木→土、土→水、水→火、火→金、金→木
var wuxingKe = map[string]string{
	"木": "土",
	"土": "水",
	"水": "火",
	"火": "金",
	"金": "木",
}

// isKeWuxing 判断 attacker 五行是否克 defender 五行
func isKeWuxing(attacker, defender string) bool {
	return wuxingKe[attacker] == defender
}
```

定位线索：紧跟在以下既有代码之后插入：

```go
var wuxingMainGan = map[string]string{
	"木": "甲",
	"火": "丙",
	"土": "戊",
	"金": "庚",
	"水": "壬",
}
```

- [ ] **Step 4: 修改第 6 层立格条件**

替换 mingge.go 第 256-262 行（数行号会因前面新增 helper 而下移；按文本定位「优先级 6」）：

```go
	// ── 优先级 6：无透格神，检测三合/三会局 ─────────────────────
	if wx := detectSanHeHui(allZhis); wx != "" {
		if repGan, ok := wuxingMainGan[wx]; ok {
			ss := GetShiShen(dayGan, repGan)
			geName := shiShenToGeName(ss)
			return geName, minggeDescDict[geName]
		}
	}
```

替换为：

```go
	// ── 优先级 6：无格时，克日干的三合 / 三会局立格 ────────────────
	if wx := detectSanHeHui(allZhis); wx != "" {
		dayWx := ganWuxingMap[dayGan]
		if isKeWuxing(wx, dayWx) {
			if repGan, ok := wuxingMainGan[wx]; ok {
				ss := GetShiShen(dayGan, repGan)
				geName := shiShenToGeName(ss)
				return geName, minggeDescDict[geName]
			}
		}
	}
```

- [ ] **Step 5: 跑测试应全绿**

```bash
go test ./pkg/bazi/ -run TestDetectMingGe -v
```

Expected: C1-C8 全部 PASS。

- [ ] **Step 6: 跑包内全部测试**

```bash
go test ./pkg/bazi/...
```

Expected: `ok`。

- [ ] **Step 7: 提交**

```bash
git add backend/pkg/bazi/mingge.go backend/pkg/bazi/mingge_test.go
git commit -m "$(cat <<'EOF'
fix(bazi/mingge): layer 6 only forms 格 when 合化 clashes day master

Previously any 三合 / 三会 局 would form a 格 using the 阳干 representative
of the merged 五行 — even when that 五行 生 or 同 the day master. This
flooded benign water-locality charts with 偏印格, etc.

Adds wuxingKe + isKeWuxing. Layer 6 now requires the merged 五行 to 克
the day master's 五行 before立格 — i.e. only 七杀格 / 偏印格 / 偏财格
shaped scenarios survive (which is the spec's "克日干" gate).

C6 covers the positive path (金 克 甲 → 七杀格); C7/C8 cover the new
filter (水 生 甲 / 火 同 甲 → 杂气格).

Spec: docs/superpowers/specs/2026-05-17-mingge-rule-fix-design.md
EOF
)"
```

---

## Task 4：第 7 层修复（吉凶神决胜）

**Files:**
- Modify: `backend/pkg/bazi/mingge_test.go`（追加 C9 / C10 / C11）
- Modify: `backend/pkg/bazi/mingge.go`：
  - 文件级新增 `xiongShenSet` / `xiongShenOrder` / `jiShenOrder` / `tieBreak`
  - 第 7 层逻辑替换

- [ ] **Step 1: 追加 C9 / C10 / C11**

```go
		// ── 第 7 层：吉凶神决胜 ────────────────────────────
		{
			name: "C9 比肩 4 + 七杀 4 → 七杀格（凶优先）（修 ④）",
			result: &BaziResult{
				YearGan: "甲", YearZhi: "申",
				MonthGan: "甲", MonthZhi: "申",
				DayGan: "甲", DayZhi: "申",
				HourGan: "甲", HourZhi: "申",
			},
			wantGe: "七杀格",
		},
		{
			name: "C10 比肩 4 + 伤官 4 → 伤官格（凶优先）（修 ④）",
			result: &BaziResult{
				YearGan: "甲", YearZhi: "午",
				MonthGan: "甲", MonthZhi: "午",
				DayGan: "甲", DayZhi: "午",
				HourGan: "甲", HourZhi: "午",
			},
			wantGe: "伤官格",
		},
		{
			name: "C11 比肩 4 + 食神 4 → 食神格（吉神内部决胜）",
			result: &BaziResult{
				YearGan: "甲", YearZhi: "巳",
				MonthGan: "甲", MonthZhi: "巳",
				DayGan: "甲", DayZhi: "巳",
				HourGan: "甲", HourZhi: "巳",
			},
			wantGe: "食神格",
		},
```

- [ ] **Step 2: 跑测试**

```bash
go test ./pkg/bazi/ -run TestDetectMingGe -v
```

Expected: C9 / C10 / C11 三个新 case 中至少一部分 **FAIL**。原因：当前实现用 `map` 遍历取 max — 频次并列时不稳定，可能错选吉神（C9/C10 期望凶神）或错选 比肩（C11 期望食神）。

> 注意：当前实现的不稳定性意味着 RED 状态本身可能 flaky。若 C9-C11 全部碰巧通过，仍需要进入下一步，因为修复后会保证**稳定**输出（这是测试的最终目的）。

- [ ] **Step 3: 在 mingge.go 文件级新增（紧跟在 Step 3.3 新增的 `isKeWuxing` 后插入）**

```go
// xiongShenSet 凶神集合（严格 4 凶口径）
var xiongShenSet = map[string]bool{
	"七杀": true,
	"伤官": true,
	"偏印": true,
	"劫财": true,
}

// xiongShenOrder 凶神同频时的内部决胜顺序：七杀 > 伤官 > 偏印 > 劫财
var xiongShenOrder = []string{"七杀", "伤官", "偏印", "劫财"}

// jiShenOrder 吉神同频时的内部决胜顺序：正官 > 正印 > 正财 > 食神 > 偏财 > 比肩
var jiShenOrder = []string{"正官", "正印", "正财", "食神", "偏财", "比肩"}

// tieBreak 在频次并列的十神候选中按 凶神顺序 → 吉神顺序 决胜
func tieBreak(tiedShiShens []string) string {
	for _, x := range xiongShenOrder {
		for _, s := range tiedShiShens {
			if s == x {
				return s
			}
		}
	}
	for _, j := range jiShenOrder {
		for _, s := range tiedShiShens {
			if s == j {
				return s
			}
		}
	}
	if len(tiedShiShens) > 0 {
		return tiedShiShens[0]
	}
	return ""
}
```

- [ ] **Step 4: 替换第 7 层逻辑**

定位 mingge.go 中的「优先级 7」注释段。替换以下原代码：

```go
	// ── 优先级 7：统计全局干支十神频率，满4个以上取最高频者 ────────
	// 统计所有天干 + 各地支主气中各十神出现次数
	allSources := append([]string{}, allGans...)
	for _, z := range allZhis {
		hgs := zhiHideGanFull[z]
		if len(hgs) > 0 {
			allSources = append(allSources, hgs[0]) // 主气
		}
	}
	shishenCount := make(map[string]int)
	for _, g := range allSources {
		ss := GetShiShen(dayGan, g)
		if ss != "" {
			shishenCount[ss]++
		}
	}
	bestSS := ""
	bestCnt := 0
	for ss, cnt := range shishenCount {
		if cnt >= 4 && cnt > bestCnt {
			bestCnt = cnt
			bestSS = ss
		}
	}
	if bestSS != "" {
		geName := shiShenToGeName(bestSS)
		return geName, minggeDescDict[geName]
	}
```

替换为：

```go
	// ── 优先级 7：全局十神频次 ≥ 4 → 凶神优先 → 内部固定顺序决胜 ────
	// 统计所有天干 + 各地支主气中各十神出现次数
	allSources := append([]string{}, allGans...)
	for _, z := range allZhis {
		hgs := zhiHideGanFull[z]
		if len(hgs) > 0 {
			allSources = append(allSources, hgs[0]) // 主气
		}
	}
	shishenCount := make(map[string]int)
	for _, g := range allSources {
		ss := GetShiShen(dayGan, g)
		if ss != "" {
			shishenCount[ss]++
		}
	}

	type cand struct {
		ss    string
		cnt   int
		xiong bool
	}
	var cands []cand
	for ss, cnt := range shishenCount {
		if cnt >= 4 {
			cands = append(cands, cand{ss, cnt, xiongShenSet[ss]})
		}
	}
	if len(cands) > 0 {
		// 1. 凶神优先：池里有凶神就只在凶神中选；都吉则全池
		var pool []cand
		for _, c := range cands {
			if c.xiong {
				pool = append(pool, c)
			}
		}
		if len(pool) == 0 {
			pool = cands
		}

		// 2. 按频次取最高
		maxCnt := pool[0].cnt
		for _, c := range pool[1:] {
			if c.cnt > maxCnt {
				maxCnt = c.cnt
			}
		}

		// 3. 频次并列 → 按 xiongShenOrder / jiShenOrder 决胜
		var tied []string
		for _, c := range pool {
			if c.cnt == maxCnt {
				tied = append(tied, c.ss)
			}
		}
		pickedSS := tieBreak(tied)
		if pickedSS != "" {
			geName := shiShenToGeName(pickedSS)
			return geName, minggeDescDict[geName]
		}
	}
```

- [ ] **Step 5: 跑测试并多跑几次验证稳定性**

```bash
go test ./pkg/bazi/ -run TestDetectMingGe -v
go test ./pkg/bazi/ -run TestDetectMingGe -count=5
```

Expected: C1-C11 全部 PASS，多次跑结果一致（修复了原 `map` 遍历的不稳定）。

- [ ] **Step 6: 跑包内全部测试**

```bash
go test ./pkg/bazi/...
```

Expected: `ok`。

- [ ] **Step 7: 提交**

```bash
git add backend/pkg/bazi/mingge.go backend/pkg/bazi/mingge_test.go
git commit -m "$(cat <<'EOF'
fix(bazi/mingge): layer 7 favors 凶神 on tie, with stable tie-break

The previous Layer 7 picked the highest-frequency 十神 with cnt >= 4,
but used a Go map iteration to do so — making the result flaky when
multiple 十神 tied at the same count.

This commit:
- Defines xiongShenSet (strict 4-evil: 七杀/伤官/偏印/劫财)
- When both 凶 and 吉 神 reach >= 4, only the 凶神 pool is considered
- Within the pool, frequency wins; ties resolved by xiongShenOrder
  (七杀 > 伤官 > 偏印 > 劫财) then jiShenOrder
  (正官 > 正印 > 正财 > 食神 > 偏财 > 比肩)
- All decisions are now deterministic

C9 / C10 verify 凶 priority on 比肩-vs-凶 ties; C11 verifies stable
吉神 internal order (食神 > 比肩).

Spec: docs/superpowers/specs/2026-05-17-mingge-rule-fix-design.md
EOF
)"
```

---

## Task 5：兜底 case + 全量回归

**Files:**
- Modify: `backend/pkg/bazi/mingge_test.go`（追加 C12）

- [ ] **Step 1: 追加 C12 兜底测试**

```go
		// ── 兜底：杂气格 ────────────────────────────────────
		{
			name: "C12 各神频次 ≤ 3、无合局、无透干 → 杂气格",
			result: &BaziResult{
				YearGan: "丙", YearZhi: "子",
				MonthGan: "戊", MonthZhi: "卯",
				DayGan: "庚", DayZhi: "午",
				HourGan: "壬", HourZhi: "酉",
			},
			wantGe: "杂气格",
		},
```

- [ ] **Step 2: 跑 mingge 单文件测试**

```bash
go test ./pkg/bazi/ -run TestDetectMingGe -v
```

Expected: C1-C12 全部 PASS。

- [ ] **Step 3: 跑包内全部测试**

```bash
go test ./pkg/bazi/...
```

Expected: `ok`。若有其它 bazi 测试 fail，**逐 case 评估**：
- 若 fail 原因是命格名变了（如某历史 case 期望「偏印格」、现在改判「杂气格」）→ 期望中的纠错，**更新历史测试的期望值**
- 若 fail 与 mingge 无关 → 真正回归，停下来排查

- [ ] **Step 4: 跑全仓库测试做最后一次回归扫描**

```bash
cd /Users/liujiming/web/yuanju/backend
go test ./...
```

Expected: 全部 `ok`。

- [ ] **Step 5: 跑 build 确认编译**

```bash
go build ./...
go vet ./...
```

Expected: 无输出（成功）。

- [ ] **Step 6: 提交**

```bash
git add backend/pkg/bazi/mingge_test.go
git commit -m "$(cat <<'EOF'
test(bazi/mingge): add C12 fallback case for 杂气格

Covers the case where no 透干 catches Layers 1-4, no 合化局 catches
Layer 6, and no 十神 reaches >= 4 in Layer 7 — ensuring the 杂气格
fallback path stays wired up.

Spec: docs/superpowers/specs/2026-05-17-mingge-rule-fix-design.md
EOF
)"
```

---

## Task 6：推上远程并准备 PR

**Files:** 无（仅 git 操作）

- [ ] **Step 1: 跑最终一次完整测试**

```bash
cd /Users/liujiming/web/yuanju/backend
go test ./pkg/bazi/ -count=3 -v
go test ./...
```

Expected: 全 PASS，包括 mingge_test.go 多次跑结果一致。

- [ ] **Step 2: 推送分支到远程**

```bash
cd /Users/liujiming/web/yuanju
git push -u origin feat/mingge-rule-fix-b
```

Expected: 远程新建 branch，push 成功。

- [ ] **Step 3: 报告完成 — 等待人类审阅 / PR 创建**

向人类汇报：
- 分支：`feat/mingge-rule-fix-b`
- 提交数：5（Task 1-5 各一）
- mingge.go 净增约 60-80 行
- mingge_test.go 新增约 180 行（12 个 case）
- 所有测试通过、`go vet` 无 warning

人类决定是否创建 PR、是否合并 main。

---

## Self-Review（写给执行 agent 看的提醒）

实施过程中容易踩的坑：

1. **mingge.go 行号会随每次插入而下移**。Task 3 加 helper 之后，Task 4 加 helper 时的"文件级新增"必须按文本（注释 / 邻近函数）定位，而非行号。
2. **C2 现代码行为是「月刃格」，C3 是「建禄格」** —— 这是当前实现走到第 4 层取 `辛` / `甲` 的结果。RED 阶段要看到这个**特定**错误值；如果看到别的（如杂气格），说明 case 构造有问题，停下重审。
3. **C7 / C8 的修复后期望是「杂气格」** —— 即合化局过滤后必须不能在第 7 层立格（构造时已确保各神频次 ≤ 3）。如果新发现某 case 在第 7 层意外立格，需要调整构造的天干/地支。
4. **C9 / C10 / C11 都用了「4 个相同天干 + 4 个相同地支」**。这种构造**完全屏蔽了 Layer 1-6**（所有藏干仅来自一个地支、与四个相同天干都不相交），保证测试只触发 Layer 7 逻辑。请勿"优化"成更"真实"的八字。
5. **修复后多跑几次 `go test -count=5`** —— 旧实现的 Layer 7 用 `map` 遍历是 nondeterministic。新实现完全 deterministic，多次跑必须输出一致。
6. **不要给 mingge.go 加新的文件级 `unicode/utf8` / `strings` 等 import**。修复用到的所有功能都已在文件内。
7. **提交粒度**：每个 Task 一个 commit。不要把多个 Task 揉到同一 commit。
