# Compatibility Scoring v3.1 — 阈值放宽 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 把 v3 严格的「六合/三合」单一闸门评分扩展为「上/中/下/0」分级，让"五行同（双生）"和"五行相生"也能贡献部分加分，避免大量真实八字得 0/100。

**Architecture:** 在 `backend/pkg/bazi/compatibility_scoring.go` 新增两个工具函数（`branchSameElement` / `branchShengElement`，复用 `event_signals.go` 的 `zhiWuxing` + `wxSheng` 表），修改 `scoreZodiac`（三级）和 `scoreDayPillar`（四级）。`scoreEightChars` 调用更新后的 `scoreDayPillar`，归一化公式不变。版本号 `v3` → `v3.1`。

**Tech Stack:** Go (后端测试 + 算法) / TypeScript+React (前端 version 分支)

**Spec:** `docs/superpowers/specs/2026-05-28-compatibility-scoring-v3.1-relax-design.md`

---

## File Structure

| 文件 | 角色 |
|---|---|
| `backend/pkg/bazi/compatibility_scoring.go` | 新增 helper + 修改两个评分函数 |
| `backend/pkg/bazi/compatibility_scoring_test.go` | 新增 TDD 测试 + 修订一处旧断言 |
| `backend/pkg/bazi/compatibility_evidence.go` | 扩展 evidence 生成（双生 / 相生 / 下档 3）|
| `backend/internal/service/compatibility_service.go` | 版本常量 `v3` → `v3.1` |
| `backend/pkg/prompt/canonical_compatibility.go` | 同步 AI prompt 中的算法描述 |
| `frontend/src/lib/api.ts` | 联合类型加 `'v3.1'` |
| `frontend/src/pages/CompatibilityResultPage.tsx` | 版本分支同时识别 `v3` / `v3.1` |
| `frontend/src/pages/CompatibilityHistoryPage.tsx` | 同上 |
| `openspec/changes/compatibility-scoring-formula-v3-1-relax/` | 新增 openspec change 4 个文件 |

---

## Task 1: 新增 branch element helpers + 测试

**Files:**
- Modify: `backend/pkg/bazi/compatibility_scoring.go`（在文件末尾追加）
- Modify: `backend/pkg/bazi/compatibility_scoring_test.go`（追加测试）

- [ ] **Step 1: 写失败测试**

把以下内容追加到 `backend/pkg/bazi/compatibility_scoring_test.go` 末尾：

```go
func TestBranchSameElement_TrueCases(t *testing.T) {
	cases := [][2]string{
		{"亥", "子"}, {"子", "亥"}, // 水
		{"寅", "卯"}, {"卯", "寅"}, // 木
		{"巳", "午"}, {"午", "巳"}, // 火
		{"申", "酉"}, {"酉", "申"}, // 金
		{"辰", "戌"}, {"丑", "未"}, {"辰", "丑"}, {"戌", "未"}, // 土
	}
	for _, p := range cases {
		if !branchSameElement(p[0], p[1]) {
			t.Errorf("branchSameElement(%q,%q) = false, want true", p[0], p[1])
		}
	}
}

func TestBranchSameElement_FalseCases(t *testing.T) {
	cases := [][2]string{
		{"子", "子"},   // 同支
		{"子", "寅"},   // 水生木（不同行）
		{"子", "丑"},   // 不同行（水/土）
		{"", "子"}, {"子", ""},
	}
	for _, p := range cases {
		if branchSameElement(p[0], p[1]) {
			t.Errorf("branchSameElement(%q,%q) = true, want false", p[0], p[1])
		}
	}
}

func TestBranchShengElement_TrueCases(t *testing.T) {
	cases := [][2]string{
		{"子", "寅"}, {"寅", "子"}, // 水生木
		{"寅", "巳"}, {"巳", "寅"}, // 木生火
		{"巳", "辰"}, {"辰", "巳"}, // 火生土
		{"辰", "申"}, {"申", "辰"}, // 土生金
		{"申", "亥"}, {"亥", "申"}, // 金生水
	}
	for _, p := range cases {
		if !branchShengElement(p[0], p[1]) {
			t.Errorf("branchShengElement(%q,%q) = false, want true", p[0], p[1])
		}
	}
}

func TestBranchShengElement_FalseCases(t *testing.T) {
	cases := [][2]string{
		{"子", "子"},   // 同支
		{"亥", "子"},   // 同行水（不是相生）
		{"子", "未"},   // 水土相克
		{"子", "辰"},   // 水土相克
		{"子", "午"},   // 水火相克
		{"", "子"}, {"子", ""},
	}
	for _, p := range cases {
		if branchShengElement(p[0], p[1]) {
			t.Errorf("branchShengElement(%q,%q) = true, want false", p[0], p[1])
		}
	}
}
```

- [ ] **Step 2: 运行测试，确认失败**

```bash
cd /Users/liujiming/web/yuanju/backend && go test ./pkg/bazi/ -run "TestBranchSameElement|TestBranchShengElement" -v
```

Expected: FAIL，错误信息含 `undefined: branchSameElement` / `undefined: branchShengElement`。

- [ ] **Step 3: 在 compatibility_scoring.go 末尾追加 helper**

```go
// branchSameElement 判定两个地支非空、不相等、且五行相同（"双生"）。
// 复用 event_signals.go 的 zhiWuxing 映射。
func branchSameElement(a, b string) bool {
	if a == "" || b == "" || a == b {
		return false
	}
	wxA, wxB := zhiWuxing[a], zhiWuxing[b]
	if wxA == "" || wxB == "" {
		return false
	}
	return wxA == wxB
}

// branchShengElement 判定两个地支非空、不相等、且五行存在 相生 关系（任一方向）。
func branchShengElement(a, b string) bool {
	if a == "" || b == "" || a == b {
		return false
	}
	wxA, wxB := zhiWuxing[a], zhiWuxing[b]
	if wxA == "" || wxB == "" {
		return false
	}
	if wxA == wxB {
		return false
	}
	return wxSheng[wxA] == wxB || wxSheng[wxB] == wxA
}
```

- [ ] **Step 4: 运行测试，确认通过**

```bash
cd /Users/liujiming/web/yuanju/backend && go test ./pkg/bazi/ -run "TestBranchSameElement|TestBranchShengElement" -v
```

Expected: PASS（4 个测试函数全过）。

- [ ] **Step 5: 提交**

```bash
cd /Users/liujiming/web/yuanju && git add backend/pkg/bazi/compatibility_scoring.go backend/pkg/bazi/compatibility_scoring_test.go
git commit -m "$(cat <<'EOF'
feat(compatibility): add branchSameElement / branchShengElement helpers (v3.1)

Reuse event_signals.go zhiWuxing + wxSheng tables. Used for v3.1
mid/lower-tier scoring in zodiac and day_pillar modules.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

## Task 2: scoreZodiac 改三级（50/30/20/0）

**Files:**
- Modify: `backend/pkg/bazi/compatibility_scoring.go:67-72`
- Modify: `backend/pkg/bazi/compatibility_scoring_test.go`（追加 + 修订）

- [ ] **Step 1: 修订旧测试 + 追加新测试**

打开 `backend/pkg/bazi/compatibility_scoring_test.go`，找到 `TestScoreZodiac_NoHit_Returns0`（约 line 108）。**把 `{"寅", "卯"}` 这一行从 cases 中删除**——v3.1 中寅卯（双生木）应当得 30，不再是 0。修改后的函数应当只保留：

```go
func TestScoreZodiac_NoHit_Returns0(t *testing.T) {
	cases := [][2]string{
		{"子", "午"}, // 六冲
		{"子", "未"}, // 六害
		{"子", "卯"}, // 相刑
		{"子", "子"}, // 同支（自刑）
	}
	for _, p := range cases {
		if got := scoreZodiac(p[0], p[1]); got != 0 {
			t.Errorf("scoreZodiac(%q,%q) = %d, want 0", p[0], p[1], got)
		}
	}
}
```

然后在文件末尾追加：

```go
func TestScoreZodiac_SameElement_Returns30(t *testing.T) {
	// 五行同（双生）：上档(六合/三合)不命中 → 中档 30
	cases := [][2]string{
		{"亥", "子"}, // 水
		{"寅", "卯"}, // 木
		{"巳", "午"}, // 火
		{"申", "酉"}, // 金
		{"辰", "戌"}, // 土，同时也是六冲——按 v3.1 优先级判中档 30
		{"丑", "未"}, // 土，同时也是六冲——同上
	}
	for _, p := range cases {
		if got := scoreZodiac(p[0], p[1]); got != 30 {
			t.Errorf("scoreZodiac(%q,%q) = %d, want 30", p[0], p[1], got)
		}
	}
}

func TestScoreZodiac_Sheng_Returns20(t *testing.T) {
	// 五行相生：上档(六合/三合) + 中档(同行) 均不命中 → 下档 20
	cases := [][2]string{
		{"子", "寅"}, // 水生木
		{"寅", "巳"}, // 木生火
		{"巳", "辰"}, // 火生土
		{"辰", "申"}, // 土生金
		{"申", "亥"}, // 金生水
	}
	for _, p := range cases {
		if got := scoreZodiac(p[0], p[1]); got != 20 {
			t.Errorf("scoreZodiac(%q,%q) = %d, want 20", p[0], p[1], got)
		}
	}
}

func TestScoreZodiac_Priority_LiuheBeatsSameElement(t *testing.T) {
	// 子丑：六合（土水？ 实际是六合，五行不同行）→ 50
	if got := scoreZodiac("子", "丑"); got != 50 {
		t.Errorf("scoreZodiac(子,丑) = %d, want 50 (六合优先)", got)
	}
	// 巳申：六合，金水化水（五行不同）→ 50
	if got := scoreZodiac("巳", "申"); got != 50 {
		t.Errorf("scoreZodiac(巳,申) = %d, want 50 (六合优先)", got)
	}
}
```

- [ ] **Step 2: 运行测试，确认失败**

```bash
cd /Users/liujiming/web/yuanju/backend && go test ./pkg/bazi/ -run "TestScoreZodiac" -v
```

Expected: 旧版 `scoreZodiac` 命中 50 否则 0，新增的 `TestScoreZodiac_SameElement_Returns30` / `TestScoreZodiac_Sheng_Returns20` 会 FAIL（返回 0 而非 30/20）。

- [ ] **Step 3: 改 scoreZodiac**

打开 `backend/pkg/bazi/compatibility_scoring.go` 找到 line 65-72：

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

改为：

```go
// scoreZodiac 计算「合属相」模块得分（满分 50，v3.1 三级）。
// 上档 50：年支六合或三合（含半三合）
// 中档 30：年支五行相同（双生）；即便同时构成六冲（辰戌/丑未）也按中档计分，与「纯加分制」原则一致
// 下档 20：年支五行相生（任一方向）
// 0    ：相克 / 相冲（不同行的子午等）/ 相穿 / 相害 / 自刑 / 同支
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

- [ ] **Step 4: 运行测试，确认通过**

```bash
cd /Users/liujiming/web/yuanju/backend && go test ./pkg/bazi/ -run "TestScoreZodiac" -v
```

Expected: PASS（含修订后的 `_NoHit_Returns0` 与新增的两个三十/二十分用例与优先级用例）。

- [ ] **Step 5: 提交**

```bash
cd /Users/liujiming/web/yuanju && git add backend/pkg/bazi/compatibility_scoring.go backend/pkg/bazi/compatibility_scoring_test.go
git commit -m "$(cat <<'EOF'
feat(compatibility): scoreZodiac three-tier 50/30/20/0 (v3.1)

Same-element (双生) → 30, same-element-but-chong (辰戌/丑未) also 30
per纯加分制 principle. Sheng → 20. Liuhe/sanhe still 50.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

## Task 3: scoreDayPillar 添加下档 3

**Files:**
- Modify: `backend/pkg/bazi/compatibility_scoring.go:88-100`
- Modify: `backend/pkg/bazi/compatibility_scoring_test.go`

- [ ] **Step 1: 追加测试**

在 `compatibility_scoring_test.go` 末尾追加：

```go
func TestScoreDayPillar_LowerTier3_SameElement(t *testing.T) {
	// 日支双生（亥子同水），干任意 → 3
	if got := scoreDayPillar("甲", "亥", "丙", "子"); got != 3 {
		t.Errorf("甲亥/丙子 双生日支: got %d, want 3", got)
	}
	// 干同（甲乙同木）也不影响下档分
	if got := scoreDayPillar("甲", "亥", "乙", "子"); got != 3 {
		t.Errorf("甲亥/乙子 双生日支(干同): got %d, want 3", got)
	}
}

func TestScoreDayPillar_LowerTier3_Sheng(t *testing.T) {
	// 日支五行相生（子→寅 水生木），干任意 → 3
	if got := scoreDayPillar("甲", "子", "丙", "寅"); got != 3 {
		t.Errorf("甲子/丙寅 水生木日支: got %d, want 3", got)
	}
}

func TestScoreDayPillar_Ke_Returns0(t *testing.T) {
	// 日支相克 / 相冲 / 相害 → 0
	if got := scoreDayPillar("甲", "子", "戊", "未"); got != 0 {
		t.Errorf("甲子/戊未 日支相害: got %d, want 0", got)
	}
	if got := scoreDayPillar("甲", "子", "戊", "辰"); got != 0 {
		t.Errorf("甲子/戊辰 日支(水土相克): got %d, want 0", got)
	}
}
```

- [ ] **Step 2: 运行测试，确认失败**

```bash
cd /Users/liujiming/web/yuanju/backend && go test ./pkg/bazi/ -run "TestScoreDayPillar_LowerTier3|TestScoreDayPillar_Ke_Returns0" -v
```

Expected: FAIL（当前 `scoreDayPillar` 在支不合时一律返回 0，双生/相生情况都是 0 而非 3）。

- [ ] **Step 3: 改 scoreDayPillar**

打开 `backend/pkg/bazi/compatibility_scoring.go` 找到 line 88-100：

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
```

改为：

```go
// scoreDayPillar 计算「合日柱」模块得分（满分 10，v3.1 四级）。
// 上档 10：日支六合/三合 + (干五合 OR 干五行相生)
// 中档 5 ：日支六合/三合 + (干同/克/无关)
// 下档 3 ：日支五行同(双生) OR 五行相生（干任意，不再细分）
// 0     ：日支相克 / 相冲 / 相穿 / 相害 / 自刑 / 同支
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

- [ ] **Step 4: 运行测试，确认通过 + 回归全部 day_pillar / eight_chars 测试**

```bash
cd /Users/liujiming/web/yuanju/backend && go test ./pkg/bazi/ -run "TestScoreDayPillar|TestScoreEightChars" -v
```

Expected: PASS（含原有 `TestScoreEightChars_RoundingTable` 等表驱动测试；新规则下旧用例值不变）。

- [ ] **Step 5: 提交**

```bash
cd /Users/liujiming/web/yuanju && git add backend/pkg/bazi/compatibility_scoring.go backend/pkg/bazi/compatibility_scoring_test.go
git commit -m "$(cat <<'EOF'
feat(compatibility): scoreDayPillar add lower-tier 3 for branch same/sheng (v3.1)

Tier ladder: 10 (zhi liuhe/sanhe + gan upper) > 5 (zhi liuhe/sanhe) >
3 (zhi same-element or sheng) > 0. scoreEightChars body unchanged.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

## Task 4: 真实案例端到端回归测试

**Files:**
- Modify: `backend/pkg/bazi/compatibility_scoring_test.go`

- [ ] **Step 1: 在 test 文件末尾追加端到端测试**

```go
func TestRealCase_1996_1995_FourModules(t *testing.T) {
	// Person A 1996-02-08 20时: 丙子 / 庚寅 / 乙亥 / 丙戌
	// Person B 1995-02-02 16时（立春前回退）: 甲戌 / 丁丑 / 甲子 / 壬申
	yearGanA, yearZhiA := "丙", "子"
	monthGanA, monthZhiA := "庚", "寅"
	dayGanA, dayZhiA := "乙", "亥"
	hourGanA, hourZhiA := "丙", "戌"

	yearGanB, yearZhiB := "甲", "戌"
	monthGanB, monthZhiB := "丁", "丑"
	dayGanB, dayZhiB := "甲", "子"
	hourGanB, hourZhiB := "壬", "申"

	// 合属相: 年支 子vs戌 (水vs土，土克水) → 0
	if got := scoreZodiac(yearZhiA, yearZhiB); got != 0 {
		t.Errorf("zodiac: got %d, want 0", got)
	}

	// 合纳音: 涧下水 vs 山头火，水克火 → 0
	if got := scoreNayin(yearGanA+yearZhiA, yearGanB+yearZhiB); got != 0 {
		t.Errorf("nayin: got %d, want 0", got)
	}

	// 合日柱: 日支 亥vs子（双生水），干 乙vs甲（同木）→ 下档 3
	if got := scoreDayPillar(dayGanA, dayZhiA, dayGanB, dayZhiB); got != 3 {
		t.Errorf("day_pillar: got %d, want 3", got)
	}

	// 合八字三柱:
	//   年 子vs戌 土克水 → 0
	//   月 寅vs丑 木被土克 → 0
	//   时 戌vs申 土生金 → 3
	//   sum=3 → 归一化 (3*2+1)/3 = 7/3 = 2
	got := scoreEightChars(
		yearGanA, yearZhiA, yearGanB, yearZhiB,
		monthGanA, monthZhiA, monthGanB, monthZhiB,
		hourGanA, hourZhiA, hourGanB, hourZhiB,
	)
	if got != 2 {
		t.Errorf("eight_chars: got %d, want 2", got)
	}

	// 总分预期 5/100 (0+0+3+2)
	total := scoreZodiac(yearZhiA, yearZhiB) +
		scoreNayin(yearGanA+yearZhiA, yearGanB+yearZhiB) +
		scoreDayPillar(dayGanA, dayZhiA, dayGanB, dayZhiB) +
		got
	if total != 5 {
		t.Errorf("total: got %d, want 5", total)
	}
}
```

- [ ] **Step 2: 运行**

```bash
cd /Users/liujiming/web/yuanju/backend && go test ./pkg/bazi/ -run "TestRealCase_1996_1995_FourModules" -v
```

Expected: PASS。

- [ ] **Step 3: 提交**

```bash
cd /Users/liujiming/web/yuanju && git add backend/pkg/bazi/compatibility_scoring_test.go
git commit -m "$(cat <<'EOF'
test(compatibility): v3.1 real-case regression — 1996-02-08 vs 1995-02-02 → 5/100

Locks in the v3.1 expected output for the user-reported pair:
zodiac=0, nayin=0, day_pillar=3, eight_chars=2, total=5.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

## Task 5: 扩展 evidence 生成（双生 / 相生 / 下档 3）

**Files:**
- Modify: `backend/pkg/bazi/compatibility_evidence.go`（zodiacEvidence、dayPillarEvidence、eightCharsEvidence、scoreExplanationSummaryV3）

- [ ] **Step 1: 追加测试**

在 `compatibility_scoring_test.go` 末尾追加：

```go
func TestZodiacEvidence_SameElement_Hit30(t *testing.T) {
	a := &BaziResult{YearGan: "甲", YearZhi: "亥"}
	b := &BaziResult{YearGan: "丙", YearZhi: "子"}
	ev := zodiacEvidence(a, b)
	if len(ev) != 1 {
		t.Fatalf("got %d evidences, want 1", len(ev))
	}
	if ev[0].EvidenceKey != "zodiac_same_element" {
		t.Errorf("got key %q, want zodiac_same_element", ev[0].EvidenceKey)
	}
	if ev[0].Weight != 30 {
		t.Errorf("got weight %d, want 30", ev[0].Weight)
	}
}

func TestZodiacEvidence_Sheng_Hit20(t *testing.T) {
	a := &BaziResult{YearGan: "甲", YearZhi: "子"}
	b := &BaziResult{YearGan: "丙", YearZhi: "寅"}
	ev := zodiacEvidence(a, b)
	if len(ev) != 1 || ev[0].EvidenceKey != "zodiac_sheng" || ev[0].Weight != 20 {
		t.Errorf("zodiac sheng evidence wrong: %+v", ev)
	}
}

func TestDayPillarEvidence_LowerTier3(t *testing.T) {
	a := &BaziResult{DayGan: "乙", DayZhi: "亥"}
	b := &BaziResult{DayGan: "甲", DayZhi: "子"}
	ev := dayPillarEvidence(a, b)
	if len(ev) != 1 {
		t.Fatalf("got %d, want 1", len(ev))
	}
	if ev[0].EvidenceKey != "day_pillar_safe" || ev[0].Weight != 3 {
		t.Errorf("day_pillar lower-3 evidence wrong: %+v", ev[0])
	}
}

func TestEightCharsEvidence_SafeTier_3(t *testing.T) {
	// 仅时柱 戌 vs 申（土生金）下档 3
	a := &BaziResult{
		YearGan: "丙", YearZhi: "子",
		MonthGan: "庚", MonthZhi: "寅",
		DayGan: "乙", DayZhi: "亥",
		HourGan: "丙", HourZhi: "戌",
	}
	b := &BaziResult{
		YearGan: "甲", YearZhi: "戌",
		MonthGan: "丁", MonthZhi: "丑",
		DayGan: "甲", DayZhi: "子",
		HourGan: "壬", HourZhi: "申",
	}
	ev := eightCharsEvidence(a, b)
	if len(ev) != 1 {
		t.Fatalf("got %d, want 1 (仅时柱命中下档)", len(ev))
	}
	if ev[0].EvidenceKey != "eight_chars_hour_safe" || ev[0].Weight != 3 {
		t.Errorf("hour safe evidence wrong: %+v", ev[0])
	}
}
```

- [ ] **Step 2: 运行，确认失败**

```bash
cd /Users/liujiming/web/yuanju/backend && go test ./pkg/bazi/ -run "TestZodiacEvidence_SameElement|TestZodiacEvidence_Sheng|TestDayPillarEvidence_LowerTier3|TestEightCharsEvidence_SafeTier" -v
```

Expected: FAIL（旧 zodiacEvidence 只识别 liuhe/sanhe，旧 dayPillarEvidence 不识别下档 3，旧 eightCharsEvidence 用 tier=upper/lower 不含 safe）。

- [ ] **Step 3: 改 zodiacEvidence（compatibility_evidence.go:16-44）**

打开 `backend/pkg/bazi/compatibility_evidence.go` 找到 `zodiacEvidence` 函数，替换为：

```go
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
	if branchSameElement(a.YearZhi, b.YearZhi) {
		wx := wxPinyin2CN[zhiWuxing[a.YearZhi]]
		return []CompatibilityEvidence{{
			EvidenceKey: "zodiac_same_element",
			Dimension:   "zodiac",
			Type:        "年支同行",
			Polarity:    "positive",
			Source:      "zodiac",
			Title:       "年支同行",
			Detail:      fmt.Sprintf("双方年支 %s/%s 同属 %s 行（双生），属相层有亲近感。", a.YearZhi, b.YearZhi, wx),
			Weight:      30,
		}}
	}
	if branchShengElement(a.YearZhi, b.YearZhi) {
		wxA := wxPinyin2CN[zhiWuxing[a.YearZhi]]
		wxB := wxPinyin2CN[zhiWuxing[b.YearZhi]]
		return []CompatibilityEvidence{{
			EvidenceKey: "zodiac_sheng",
			Dimension:   "zodiac",
			Type:        "年支相生",
			Polarity:    "positive",
			Source:      "zodiac",
			Title:       "年支相生",
			Detail:      fmt.Sprintf("双方年支 %s/%s 构成 %s/%s 五行相生，属相层有顺承之意。", a.YearZhi, b.YearZhi, wxA, wxB),
			Weight:      20,
		}}
	}
	return nil
}
```

- [ ] **Step 4: 改 dayPillarEvidence（compatibility_evidence.go:78-110）**

替换 `dayPillarEvidence` 全函数为：

```go
func dayPillarEvidence(a, b *BaziResult) []CompatibilityEvidence {
	if branchCompatible(a.DayZhi, b.DayZhi) {
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
			Type:        "日柱次吉",
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
	if branchSameElement(a.DayZhi, b.DayZhi) || branchShengElement(a.DayZhi, b.DayZhi) {
		return []CompatibilityEvidence{{
			EvidenceKey: "day_pillar_safe",
			Dimension:   "day_pillar",
			Type:        "日柱安慰",
			Polarity:    "positive",
			Source:      "day_pillar",
			Title:       "日柱安慰分",
			Detail: fmt.Sprintf(
				"日柱 %s%s/%s%s 地支虽不合，但五行相同或相生，亲密层有微弱亲近感。",
				a.DayGan, a.DayZhi, b.DayGan, b.DayZhi,
			),
			Weight: 3,
		}}
	}
	return nil
}
```

- [ ] **Step 5: 改 eightCharsEvidence（compatibility_evidence.go:112-153）**

替换 `eightCharsEvidence` 全函数为：

```go
func eightCharsEvidence(a, b *BaziResult) []CompatibilityEvidence {
	out := make([]CompatibilityEvidence, 0, 3)
	type pillar struct {
		name  string
		label string
		ganA  string
		zhiA  string
		ganB  string
		zhiB  string
	}
	pillars := []pillar{
		{"year", "年柱", a.YearGan, a.YearZhi, b.YearGan, b.YearZhi},
		{"month", "月柱", a.MonthGan, a.MonthZhi, b.MonthGan, b.MonthZhi},
		{"hour", "时柱", a.HourGan, a.HourZhi, b.HourGan, b.HourZhi},
	}
	tierByScore := map[int]struct {
		key   string
		label string
	}{
		10: {"upper", "上档"},
		5:  {"lower", "下档"},
		3:  {"safe", "安慰分"},
	}
	for _, p := range pillars {
		s := scoreDayPillar(p.ganA, p.zhiA, p.ganB, p.zhiB)
		t, ok := tierByScore[s]
		if !ok {
			continue
		}
		out = append(out, CompatibilityEvidence{
			EvidenceKey: "eight_chars_" + p.name + "_" + t.key,
			Dimension:   "eight_chars",
			Type:        p.label + "对" + t.label,
			Polarity:    "positive",
			Source:      "eight_chars",
			Title:       p.label + "对" + t.label,
			Detail: fmt.Sprintf(
				"%s %s%s/%s%s 命中%s（贡献 %d）。",
				p.label, p.ganA, p.zhiA, p.ganB, p.zhiB, t.label, s,
			),
			Weight: s,
		})
	}
	return out
}
```

- [ ] **Step 6: 改 scoreExplanationSummaryV3 zodiac/day_pillar 分支（compatibility_evidence.go:182-219）**

把 `scoreExplanationSummaryV3` 函数中的 `case "zodiac":` 和 `case "day_pillar":` 两段替换为：

```go
	case "zodiac":
		if hit == nil {
			return fmt.Sprintf("双方年支 %s/%s 无六合 / 三合 / 同行 / 相生，属相层无加成。", a.YearZhi, b.YearZhi)
		}
		switch hit.EvidenceKey {
		case "zodiac_liuhe":
			return fmt.Sprintf("双方属相 %s/%s 构成六合，关系基础线吸引力强。", a.YearZhi, b.YearZhi)
		case "zodiac_sanhe":
			return fmt.Sprintf("双方属相 %s/%s 同属 %s 三合局，气场协同。",
				a.YearZhi, b.YearZhi, sanheGroupName(a.YearZhi, b.YearZhi))
		case "zodiac_same_element":
			return fmt.Sprintf("双方年支 %s/%s 五行同行（双生），属相层有亲近感。", a.YearZhi, b.YearZhi)
		case "zodiac_sheng":
			return fmt.Sprintf("双方年支 %s/%s 五行相生，属相层有顺承之意。", a.YearZhi, b.YearZhi)
		}
		return ""
```

```go
	case "day_pillar":
		if hit == nil {
			return fmt.Sprintf("日柱 %s%s/%s%s 地支不合且五行无亲，亲密层无加成。",
				a.DayGan, a.DayZhi, b.DayGan, b.DayZhi)
		}
		switch hit.EvidenceKey {
		case "day_pillar_upper":
			return fmt.Sprintf("日柱 %s%s/%s%s 地支合且天干五合 / 相生，亲密层结构稳。",
				a.DayGan, a.DayZhi, b.DayGan, b.DayZhi)
		case "day_pillar_lower":
			return fmt.Sprintf("日柱 %s%s/%s%s 地支合，天干仅相同 / 克 / 无关，亲密层有基础但未达上吉。",
				a.DayGan, a.DayZhi, b.DayGan, b.DayZhi)
		case "day_pillar_safe":
			return fmt.Sprintf("日柱 %s%s/%s%s 地支虽不合，但五行同行或相生，亲密层有微弱亲近感。",
				a.DayGan, a.DayZhi, b.DayGan, b.DayZhi)
		}
		return ""
```

注：`case "eight_chars"` 调用的 `eightCharsSummary` 内部判断 `scoreDayPillar(...) > 0` 即算命中，新规则下下档 3 也会被统计为命中，无需修改。`case "nayin":` 完全不变。

- [ ] **Step 7: 运行所有 bazi 包测试**

```bash
cd /Users/liujiming/web/yuanju/backend && go test ./pkg/bazi/ -v
```

Expected: ALL PASS（含新 evidence 测试 + 旧测试 + 真实案例回归）。

- [ ] **Step 8: 提交**

```bash
cd /Users/liujiming/web/yuanju && git add backend/pkg/bazi/compatibility_evidence.go backend/pkg/bazi/compatibility_scoring_test.go
git commit -m "$(cat <<'EOF'
feat(compatibility): evidence + score-explanation for v3.1 new tiers

- zodiacEvidence: emit zodiac_same_element (30) / zodiac_sheng (20)
- dayPillarEvidence: emit day_pillar_safe (3)
- eightCharsEvidence: tierByScore table — upper/lower/safe
- scoreExplanationSummaryV3: zodiac/day_pillar branches updated

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

## Task 6: 升级版本号到 v3.1

**Files:**
- Modify: `backend/internal/service/compatibility_service.go:16`
- Modify: `backend/internal/service/compatibility_service_test.go`（如有版本断言）

- [ ] **Step 1: 检查并修改 service 文件**

```bash
grep -n "compatibilityAnalysisVersion\|\"v3\"" backend/internal/service/compatibility_service.go
```

打开 `backend/internal/service/compatibility_service.go` 找到 line 16：

```go
const compatibilityAnalysisVersion = "v3"
```

改为：

```go
const compatibilityAnalysisVersion = "v3.1"
```

- [ ] **Step 2: 检查 service 测试是否有版本断言**

```bash
grep -n "\"v3\"\|analysis_version" backend/internal/service/compatibility_service_test.go
```

如果有 `== "v3"` 之类的断言，全部改为 `"v3.1"`。如果没有则跳过。

- [ ] **Step 3: 运行 service 测试**

```bash
cd /Users/liujiming/web/yuanju/backend && go test ./internal/service/ -run "Compatibility" -v
```

Expected: PASS。

- [ ] **Step 4: 提交**

```bash
cd /Users/liujiming/web/yuanju && git add backend/internal/service/compatibility_service.go backend/internal/service/compatibility_service_test.go 2>/dev/null
git commit -m "$(cat <<'EOF'
feat(compatibility): bump analysis_version v3 → v3.1

New readings will carry v3.1. Existing v3 records remain unchanged
and continue to render via the v3 frontend branch.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

## Task 7: 同步 AI prompt 中的算法描述

**Files:**
- Modify: `backend/pkg/prompt/canonical_compatibility.go`

- [ ] **Step 1: 修改 prompt 版本号 + 规则描述**

打开 `backend/pkg/prompt/canonical_compatibility.go`：

把 line 5 的 prompt 版本：
```go
		Version:     "v3-question-aware-2",
```
改为：
```go
		Version:     "v3.1-question-aware-1",
```

把 line 28 的注释：
```
四模块分数（JSON，v3 评分公式）：
```
改为：
```
四模块分数（JSON，v3.1 评分公式）：
```

把 line 32-35 评分规则说明：
```
- zodiac（合属相，0–50）：年支六合或三合命中即满分 50，否则 0。
- nayin（合纳音，0–20）：年柱纳音五行相生或相同得 20，相克 0。
- day_pillar（合日柱，0–10）：日支合 + 干合/相生 = 10，日支合 + 其他 = 5，日支不合 = 0。
- eight_chars（合八字，0–20）：年/月/时三柱独立按合日柱规则得 0/5/10，三柱和归一化到 [0,20]。
```
改为：
```
- zodiac（合属相，0–50，v3.1 三级）：年支六合/三合 = 50；五行同行（双生）= 30；五行相生 = 20；相克/相冲/相穿 = 0。
- nayin（合纳音，0–20）：年柱纳音五行相生或相同得 20，相克 0。
- day_pillar（合日柱，0–10，v3.1 四级）：日支六合/三合 + 干合/相生 = 10；日支六合/三合 = 5；日支五行同/相生 = 3；日支相克/相冲 = 0。
- eight_chars（合八字，0–20）：年/月/时三柱独立按合日柱规则得 0/3/5/10，三柱和归一化到 [0,20]。
```

- [ ] **Step 2: 运行 prompt 同步测试（若有）**

```bash
cd /Users/liujiming/web/yuanju/backend && go test ./pkg/prompt/ -v
```

Expected: PASS。如果 `sync_test.go` 用快照比对，会要求更新——按提示更新快照即可。

- [ ] **Step 3: 提交**

```bash
cd /Users/liujiming/web/yuanju && git add backend/pkg/prompt/canonical_compatibility.go
# 如果 sync_test 触发了快照更新，也一并 add
git add backend/pkg/prompt/ 2>/dev/null
git commit -m "$(cat <<'EOF'
feat(compatibility): update AI prompt to describe v3.1 scoring rules

Prompt version v3-question-aware-2 → v3.1-question-aware-1.
Scoring-rules block reflects new tiers (zodiac 50/30/20/0,
day_pillar 10/5/3/0).

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

## Task 8: 前端联合类型扩展 v3.1

**Files:**
- Modify: `frontend/src/lib/api.ts`（line 394 和 446）

- [ ] **Step 1: 修改两处 analysis_version 联合类型**

打开 `frontend/src/lib/api.ts`：

line 394：
```ts
  analysis_version: 'v1' | 'v2' | 'v3'
```
改为：
```ts
  analysis_version: 'v1' | 'v2' | 'v3' | 'v3.1'
```

line 446：同样修改（同名字段在 `CompatibilityHistoryItem`）。

- [ ] **Step 2: 运行前端类型检查**

```bash
cd /Users/liujiming/web/yuanju/frontend && npm run build 2>&1 | head -40
```

Expected: 类型检查通过（或仅显示因 page 文件还没改导致的窄类型 warning——下一任务修复）。如果完全失败请把错误信息粘到下一任务前确认。

- [ ] **Step 3: 提交**

```bash
cd /Users/liujiming/web/yuanju && git add frontend/src/lib/api.ts
git commit -m "$(cat <<'EOF'
feat(compatibility): extend api.ts analysis_version union with 'v3.1'

Covers both CompatibilityReading and CompatibilityHistoryItem.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

## Task 9: 前端 v3 渲染分支同时识别 v3.1

**Files:**
- Modify: `frontend/src/pages/CompatibilityResultPage.tsx:942`
- Modify: `frontend/src/pages/CompatibilityHistoryPage.tsx:108`

- [ ] **Step 1: 修改 CompatibilityResultPage.tsx**

打开 `frontend/src/pages/CompatibilityResultPage.tsx` line 942：
```ts
  const isV3 = reading.analysis_version === 'v3' && isV3DimensionScores(reading.dimension_scores)
```
改为：
```ts
  const isV3 = (reading.analysis_version === 'v3' || reading.analysis_version === 'v3.1') && isV3DimensionScores(reading.dimension_scores)
```

- [ ] **Step 2: 修改 CompatibilityHistoryPage.tsx**

打开 `frontend/src/pages/CompatibilityHistoryPage.tsx` line 108：
```ts
              const isV3 = item.analysis_version === 'v3'
```
改为：
```ts
              const isV3 = item.analysis_version === 'v3' || item.analysis_version === 'v3.1'
```

- [ ] **Step 3: 运行前端类型检查 + 构建**

```bash
cd /Users/liujiming/web/yuanju/frontend && npm run build 2>&1 | tail -20
```

Expected: 构建成功，无 TS 错误。

- [ ] **Step 4: 提交**

```bash
cd /Users/liujiming/web/yuanju && git add frontend/src/pages/CompatibilityResultPage.tsx frontend/src/pages/CompatibilityHistoryPage.tsx
git commit -m "$(cat <<'EOF'
feat(compatibility): frontend v3/v3.1 share ScoreOverviewV3 renderer

History page + result page version branch now recognizes both 'v3'
and 'v3.1'. Schema identical, only scoring tiers differ.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

## Task 10: 新增 openspec change folder

**Files:**
- Create: `openspec/changes/compatibility-scoring-formula-v3-1-relax/proposal.md`
- Create: `openspec/changes/compatibility-scoring-formula-v3-1-relax/design.md`
- Create: `openspec/changes/compatibility-scoring-formula-v3-1-relax/tasks.md`
- Create: `openspec/changes/compatibility-scoring-formula-v3-1-relax/specs/compatibility-scoring-formula/spec.md`

- [ ] **Step 1: 创建 proposal.md**

```bash
mkdir -p openspec/changes/compatibility-scoring-formula-v3-1-relax/specs/compatibility-scoring-formula
```

写入 `openspec/changes/compatibility-scoring-formula-v3-1-relax/proposal.md`：

```markdown
# Compatibility Scoring Formula v3.1 — Relax Threshold

## Why
v3 评分（合属相 50 + 合纳音 20 + 合日柱 10 + 合八字 20）将「六合/三合」与「纳音相生/相同」设为单一闸门，地支双生（如亥子/寅卯）与五行相生（如子→寅）均得 0。统计上约 73% 的地支两两组合在合属相模块得 0；用户在真实案例 (1996-02-08 20时) vs (1995-02-02 16时) 上拿到 0/100 + 空 evidence 列表，无法理解。

## What Changes
- 合属相：单档 50/0 → 三级 50/30/20/0（双生 30、相生 20）
- 合日柱：三级 0/5/10 → 四级 0/3/5/10（日支同行/相生 → 下档 3）
- 合八字：内部调用更新后的 scoreDayPillar；归一化函数不变
- 合纳音：不变
- `analysis_version`: `"v3"` → `"v3.1"`；v3 旧记录保留不重算
- AI prompt 与 evidence 文案同步更新

## Impact
- Affected specs:
  - **MODIFIED:** `compatibility-scoring-formula`（zodiac / day_pillar / eight_chars / analysis_version 四个 Requirement）
- Affected code:
  - backend/pkg/bazi/compatibility_scoring.go
  - backend/pkg/bazi/compatibility_scoring_test.go
  - backend/pkg/bazi/compatibility_evidence.go
  - backend/internal/service/compatibility_service.go
  - backend/pkg/prompt/canonical_compatibility.go
  - frontend/src/lib/api.ts
  - frontend/src/pages/CompatibilityResultPage.tsx
  - frontend/src/pages/CompatibilityHistoryPage.tsx
- DB migration: 无 schema 变化
- 历史记录: v3 行保留 analysis_version='v3' 不重算
```

- [ ] **Step 2: 创建 design.md**

写入 `openspec/changes/compatibility-scoring-formula-v3-1-relax/design.md`：

```markdown
# v3.1 Relax-Threshold Design

详见 `docs/superpowers/specs/2026-05-28-compatibility-scoring-v3.1-relax-design.md`。

本 change 的关键设计点：
1. 同时构成「五行同」与「六冲」的辰戌、丑未仍判中档 30（与「纯加分制」一致）
2. 合日柱下档 3 不区分干关系（下档已是安慰分）
3. 归一化 `(sum × 2 + 1) / 3` 不变，sum 仍 ∈ [0, 30]
4. 新增 helper：`branchSameElement` / `branchShengElement`，复用 `event_signals.go` 的 `zhiWuxing` / `wxSheng`
```

- [ ] **Step 3: 创建 tasks.md**

写入 `openspec/changes/compatibility-scoring-formula-v3-1-relax/tasks.md`：

```markdown
# Tasks

- [x] T1 新增 branchSameElement / branchShengElement helper
- [x] T2 scoreZodiac 改三级 50/30/20/0
- [x] T3 scoreDayPillar 加下档 3
- [x] T4 端到端回归测试 (1996/1995 → 5/100)
- [x] T5 evidence 模块扩展
- [x] T6 analysis_version v3 → v3.1
- [x] T7 AI prompt 算法描述同步
- [x] T8 前端 api.ts 联合类型扩展
- [x] T9 前端 result/history 页面 version 分支同步
- [x] T10 openspec change folder + spec 修订
```

- [ ] **Step 4: 创建 spec.md（MODIFIED Requirements）**

写入 `openspec/changes/compatibility-scoring-formula-v3-1-relax/specs/compatibility-scoring-formula/spec.md`：

```markdown
# compatibility-scoring-formula Specification

## MODIFIED Requirements

### Requirement: Zodiac module (year-zhi liuhe / sanhe + same-element + sheng)
The zodiac module SHALL award 50 points when the two year-zhi form 六合 or 三合 (half-sanhe acceptable); 30 points when the two year-zhi have the same 五行 (双生); 20 points when the two year-zhi 五行 stand in a 生 relation (either direction); otherwise 0.

#### Scenario: Year-zhi liuhe hit
- **WHEN** the year-zhi pair is one of {子丑, 寅亥, 卯戌, 辰酉, 巳申, 午未}
- **THEN** zodiac SHALL be 50

#### Scenario: Year-zhi half-sanhe hit
- **WHEN** the year-zhi pair is two distinct members of one sanhe group {申子辰, 亥卯未, 巳酉丑, 寅午戌}
- **THEN** zodiac SHALL be 50

#### Scenario: Year-zhi same-element (double-life)
- **WHEN** the two year-zhi share the same 五行 (亥子 水、寅卯 木、巳午 火、申酉 金、辰戌丑未 土两两) AND the pair is not 六合/三合
- **THEN** zodiac SHALL be 30
- **AND** this applies even when the pair simultaneously forms a 六冲 (e.g. 辰戌, 丑未) — 纯加分制 with no negative deductions

#### Scenario: Year-zhi sheng
- **WHEN** the two year-zhi 五行 stand in a 生 relation (either direction, e.g. 子→寅 水生木) AND the pair is neither 六合/三合 nor same-element
- **THEN** zodiac SHALL be 20

#### Scenario: No hit (相克 / 相冲 with different 五行 / 相穿 / 相害 / 自刑 / 同支)
- **WHEN** none of the above hits
- **THEN** zodiac SHALL be 0

### Requirement: Day pillar module (four tiers)
The day_pillar module SHALL award:
- 10 (上档 / `day_pillar_upper`) when day-zhi is 六合/三合 AND day-gan is 五合 or 五行相生
- 5 (下档 / `day_pillar_lower`) when day-zhi is 六合/三合 (day-gan 同/克/无关)
- 3 (安慰分 / `day_pillar_safe`) when day-zhi 五行同 (双生) OR 五行相生 (day-gan ignored)
- 0 when day-zhi is 相克 / 相冲 / 相穿 / 相害 / 同支

注：tier 命名 "lower" 保留 v3 原有的 evidence_key 字符串以避免破坏 v3 记录的 evidence 兼容性；v3.1 新增的 3 分档使用独立 key `day_pillar_safe`。

#### Scenario: Upper tier — gan 五合 + zhi 六合/三合
- **WHEN** day-gan pair is in 天干五合 set {甲己, 乙庚, 丙辛, 丁壬, 戊癸} AND day-zhi pair is 六合/三合
- **THEN** day_pillar SHALL be 10

#### Scenario: Upper tier — gan 五行相生 + zhi 六合/三合
- **WHEN** the two day-gan 五行 stand in a 相生 relation (excluding identity) AND day-zhi pair is 六合/三合
- **THEN** day_pillar SHALL be 10

#### Scenario: Lower tier — zhi 六合/三合 alone
- **WHEN** day-zhi pair is 六合/三合 AND no upper-tier 干 condition met
- **THEN** day_pillar SHALL be 5

#### Scenario: Safe tier — zhi same-element or sheng
- **WHEN** day-zhi pair is 五行同 (双生) OR 五行相生 AND not 六合/三合
- **THEN** day_pillar SHALL be 3, regardless of day-gan relation

#### Scenario: zhi 不合（相克 / 相冲 with different 五行 / 同支）
- **WHEN** none of the above scenarios apply
- **THEN** day_pillar SHALL be 0

### Requirement: Eight-chars module (per-pillar 0/3/5/10)
The eight_chars module SHALL score each of the three non-day pillar pairs (year/year, month/month, hour/hour) by the day_pillar rule, sum the three results (0–30) and normalize to [0, 20] via (sum × 2 + 1) / 3 with integer division.

#### Scenario: All three pillars upper tier
- **WHEN** year/month/hour pillar pairs each score 10
- **THEN** eight_chars SHALL be 20

#### Scenario: One pillar safe tier, two not compatible
- **WHEN** sum is 3
- **THEN** eight_chars SHALL be 2

#### Scenario: One pillar upper tier, two not compatible
- **WHEN** sum is 10
- **THEN** eight_chars SHALL be 7

### Requirement: Analysis version tag (v3.1)
The analysis_version field of new CompatibilityReading records written by this engine SHALL be "v3.1"; v1/v2/v3 records remain unchanged and renderable through their respective frontend paths.

#### Scenario: New reading written
- **WHEN** CreateCompatibilityReading runs after this change
- **THEN** the stored row SHALL have analysis_version = "v3.1"

#### Scenario: Legacy v3 record read
- **WHEN** a v3 record is fetched
- **THEN** the API SHALL surface its v3 dimension_scores (zodiac/nayin/day_pillar/eight_chars) as-is
- **AND** overall_score SHALL remain whatever was originally stored
```

- [ ] **Step 5: 提交**

```bash
cd /Users/liujiming/web/yuanju && git add openspec/changes/compatibility-scoring-formula-v3-1-relax
git commit -m "$(cat <<'EOF'
spec(compatibility): openspec change for v3.1 relax-threshold

Adds proposal, design (referring back to docs/superpowers/specs/...),
tasks completion log, and MODIFIED Requirements for
compatibility-scoring-formula spec (zodiac / day_pillar /
eight_chars / analysis_version).

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

## Task 11: 端到端手动验证

**Files:** N/A（仅运行验证）

- [ ] **Step 1: 跑后端全部测试**

```bash
cd /Users/liujiming/web/yuanju/backend && go test ./... 2>&1 | tail -30
```

Expected: 所有包通过；特别注意 `pkg/bazi` 和 `internal/service`。

- [ ] **Step 2: 跑前端构建**

```bash
cd /Users/liujiming/web/yuanju/frontend && npm run build 2>&1 | tail -10
```

Expected: build success，无 TS 错误。

- [ ] **Step 3: 启动 dev server，手测两人案例**

```bash
cd /Users/liujiming/web/yuanju && docker-compose up -d
# 或按项目惯例启动 backend + frontend
```

打开浏览器 → 合盘验证 → 输入 (1996-02-08 20时, 男/女不重要) vs (1995-02-02 16时)。

期望：
- 总分显示 **5/100**
- 合属相 0/50（子戌 土克水）
- 合纳音 0/20（水克火）
- 合日柱 **3/10**（亥子双生水 + 干同木）
- 合八字 **2/20**（仅时柱 戌申 土生金 命中下档 3 → 归一化 2）
- evidence 列表至少 1 条（time pillar safe）
- analysis_version 是 `v3.1`（开发者工具查 API 返回）

- [ ] **Step 4: 验证 v3 旧记录仍可正常渲染**

打开合盘历史页，找一条旧的 v3 记录（如果数据库里有）。期望显示与改动前完全一致（不重算）。

- [ ] **Step 5: 若验证通过 — Done**

不需要提交。

---

## Self-Review（仅供 plan 作者，实施时跳过）

**1. Spec coverage（spec 中的每个 Requirement 是否有对应 task？）**
- ✅ Zodiac three-tier → Task 2 + 修订 evidence (Task 5)
- ✅ Day pillar four-tier → Task 3 + evidence (Task 5)
- ✅ Eight-chars normalization unchanged → 隐含在 Task 3（scoreEightChars 调用更新后的 scoreDayPillar）
- ✅ analysis_version v3.1 → Task 6 + 前端 (Task 8/9)
- ✅ Spec 文档更新 → Task 10
- ✅ Prompt 同步 → Task 7
- ✅ 真实案例 5/100 → Task 4 (单元) + Task 11 (端到端)

**2. Placeholder scan：** 无 TBD / TODO，每个 step 都有具体代码或具体 path。

**3. Type consistency：** 函数命名 `branchSameElement` / `branchShengElement` 在 helpers (Task 1)、scoreZodiac (Task 2)、scoreDayPillar (Task 3)、evidence (Task 5) 中保持一致。EvidenceKey 命名 `zodiac_same_element` / `zodiac_sheng` / `day_pillar_safe` / `eight_chars_*_safe` 在 evidence 函数与 summary 函数 (Task 5) 间一致。版本号 `v3.1` 字符串在 backend / frontend / spec 中一致（带点号）。

---

## 实施方式选择

Plan 完成并保存到 `docs/superpowers/plans/2026-05-28-compatibility-scoring-v3.1-relax.md`。两种执行方式：

**1. Subagent-Driven（推荐）** — 每个 task 派一个新 subagent 实现 + 我两阶段 review

**2. Inline Execution** — 在当前 session 串行执行所有 task + 检查点 review

请选择哪种？
