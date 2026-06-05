# 夫妻宫匹配 (Spouse Palace Match) Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在合盘报告里新增一节「夫妻宫匹配」——用一方命盘的配偶星 + 夫妻宫藏干推出「TA 理想的另一半画像」，再拿对方真人去比，给出高/中/低匹配档与文字说明（双向），**不改 0–100 总分**。

**Architecture:** 路线一（薄 Go + LLM 写）。Go 只补一个纯函数 `DetectSpouseStarSignal`（按性别定位配偶星：男财女官杀），把结构化信号拼进每人命盘摘要；画像解读与高/中/低判定交给 LLM。双方夫妻宫的合/冲/克/刑/害沿用已有的 day_pillar 维度 + negative 证据当锚点，不新做。新增 LLM 输出字段 `spouse_palace_match` 由 model 结构体承接、随报告 JSON 持久化。

**Tech Stack:** Go 1.25（module `yuanju`），`pkg/bazi` 纯函数层、`internal/service` 编排层、`internal/model` DTO、`pkg/prompt` 提示词 canonical。测试用 `go test`。

**与 spec 的一处刻意简化：** spec §4.1 列了 `StrengthLabel`（复用日主旺衰）。本计划**不在配偶信号块里重复旺衰**——因为每人摘要串里已经有「旺衰=…」紧邻其后，再写一遍属冗余（CLAUDE.md §2/§5.1）。LLM 仍能从同一行读到旺衰。其余字段全部按 spec 实现。

**部署提醒（与上次一致）：** Go 改动随发版生效；**prompt 改动需后台手动「采用出厂新版」**（`SyncCanonical` 只补种不覆盖）。

---

## File Structure

| 文件 | 职责 | 动作 |
|---|---|---|
| `backend/pkg/bazi/compatibility_spouse.go` | `SpouseStarSignal` 结构 + `DetectSpouseStarSignal` 纯函数（配偶星定位） | Create |
| `backend/pkg/bazi/compatibility_spouse_test.go` | 上述函数的表驱动单测 | Create |
| `backend/internal/service/compatibility_service.go` | 把配偶信号拼进 `compatibilityParticipantSummary` 输出；新增 `spousePortraitSignalText` 格式化 helper | Modify |
| `backend/internal/service/compatibility_service_test.go` | helper 三分支测试 + 摘要串集成测试 + `spouse_palace_match` 解析测试 | Modify |
| `backend/internal/model/compatibility.go` | 新增 `CompatibilitySpousePalaceSide` / `CompatibilitySpousePalaceMatch` 结构 + `CompatibilityStructuredReport.SpousePalaceMatch` 字段 | Modify |
| `backend/pkg/prompt/canonical_compatibility.go` | 版本 `-5`→`-6`；新增 `spouse_palace_match` 输出 schema + 生成约束 | Modify |
| `backend/pkg/prompt/canonical_test.go` | 版本 pin `-5`→`-6`；断言提示词含 `spouse_palace_match` | Modify |

所有 `go test` 命令在 `backend/` 目录下执行（module root）。

---

## Task 1: 配偶星定位纯函数（pkg/bazi）

**Files:**
- Create: `backend/pkg/bazi/compatibility_spouse.go`
- Test: `backend/pkg/bazi/compatibility_spouse_test.go`

**背景（实现者必读）：** `BaziResult`（`backend/pkg/bazi/engine.go:24`）已含本任务需要的全部字段：`Gender`（取值 `"male"`/`"female"`，见 `report_service.go` 中 `r.Gender == "male"` 判断）、`DayGan`、四柱天干十神 `YearGanShiShen`/`MonthGanShiShen`/`DayGanShiShen`/`HourGanShiShen`（`string`）、四柱地支藏干十神 `YearZhiShiShen`/`MonthZhiShiShen`/`DayZhiShiShen`/`HourZhiShiShen`（`[]string`，即该地支各藏干对日主的十神）。男命配偶星 = 财星（正财/偏财），女命 = 官杀（正官/七杀）。夫妻宫 = 日支。本函数只陈述客观事实，不做性格解读。

- [ ] **Step 1: 写失败测试**

创建 `backend/pkg/bazi/compatibility_spouse_test.go`：

```go
package bazi

import (
	"reflect"
	"testing"
)

func TestDetectSpouseStarSignal(t *testing.T) {
	cases := []struct {
		name string
		in   *BaziResult
		want SpouseStarSignal
	}{
		{
			name: "男命：正财透月干、偏财藏年支，日支藏干非财",
			in: &BaziResult{
				Gender: "male", DayGan: "甲",
				MonthGanShiShen: "正财",
				YearZhiShiShen:  []string{"偏财"},
				DayZhiShiShen:   []string{"伤官"},
			},
			want: SpouseStarSignal{
				Available: true, Present: true, Category: "财星",
				StarNames:              []string{"正财", "偏财"},
				Positions:              []string{"月干(透)", "年支(藏)"},
				Visible:                true,
				InSpousePalace:         false,
				DayBranchHiddenShiShen: []string{"伤官"},
			},
		},
		{
			name: "男命：财星坐日支（入夫妻宫，仅藏不透）",
			in: &BaziResult{
				Gender: "male", DayGan: "甲",
				DayZhiShiShen: []string{"正财"},
			},
			want: SpouseStarSignal{
				Available: true, Present: true, Category: "财星",
				StarNames:              []string{"正财"},
				Positions:              []string{"日支(藏)"},
				Visible:                false,
				InSpousePalace:         true,
				DayBranchHiddenShiShen: []string{"正财"},
			},
		},
		{
			name: "女命：正官透时干",
			in: &BaziResult{
				Gender: "female", DayGan: "甲",
				HourGanShiShen: "正官",
				DayZhiShiShen:  []string{"比肩"},
			},
			want: SpouseStarSignal{
				Available: true, Present: true, Category: "官杀",
				StarNames:              []string{"正官"},
				Positions:              []string{"时干(透)"},
				Visible:                true,
				InSpousePalace:         false,
				DayBranchHiddenShiShen: []string{"比肩"},
			},
		},
		{
			name: "男命：配偶星不现（命中无财）",
			in: &BaziResult{
				Gender: "male", DayGan: "甲",
				YearGanShiShen: "比肩",
				DayZhiShiShen:  []string{"劫财"},
			},
			want: SpouseStarSignal{
				Available: true, Present: false, Category: "财星",
				DayBranchHiddenShiShen: []string{"劫财"},
			},
		},
		{
			name: "性别缺失 → 不可用",
			in:   &BaziResult{Gender: "", DayGan: "甲"},
			want: SpouseStarSignal{Available: false},
		},
		{
			name: "缺日柱 → 不可用",
			in:   &BaziResult{Gender: "male", DayGan: ""},
			want: SpouseStarSignal{},
		},
		{
			name: "nil → 不可用",
			in:   nil,
			want: SpouseStarSignal{},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := DetectSpouseStarSignal(tc.in)
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("DetectSpouseStarSignal()\n got=%+v\nwant=%+v", got, tc.want)
			}
		})
	}
}
```

- [ ] **Step 2: 跑测试确认失败**

Run: `go test ./pkg/bazi/ -run TestDetectSpouseStarSignal -v`
Expected: 编译失败 / `undefined: SpouseStarSignal` 与 `undefined: DetectSpouseStarSignal`。

- [ ] **Step 3: 写实现**

创建 `backend/pkg/bazi/compatibility_spouse.go`：

```go
package bazi

// SpouseStarSignal 描述一个人命盘中「配偶星 + 夫妻宫」的结构化信号，
// 供 LLM 推导「TA 命里理想/容易吸引的另一半画像」。本结构只陈述客观事实，不做性格解读。
type SpouseStarSignal struct {
	Available              bool     // 性别可识别（能定配偶星类别）；false → 上层跳过本节
	Present                bool     // 命盘中存在配偶星
	Category               string   // "财星"(男) / "官杀"(女)；Available=false 时为空
	StarNames              []string // 命中出现的具体十神：正财/偏财 或 正官/七杀（去重保序）
	Positions              []string // 配偶星位置，如 "月干(透)"、"日支(藏)"
	Visible                bool     // 配偶星是否透于天干
	InSpousePalace         bool     // 配偶星是否坐日支（夫妻宫）
	DayBranchHiddenShiShen []string // 日支藏干各自的十神（夫妻宫画像主料）
}

// DetectSpouseStarSignal 按性别定位配偶星（男看财星、女看官杀），扫描四柱天干（透）
// 与地支藏干（藏）的十神，并附上日支藏干十神。性别缺失或盘异常时返回 Available=false。
func DetectSpouseStarSignal(r *BaziResult) SpouseStarSignal {
	if r == nil || r.DayGan == "" {
		return SpouseStarSignal{}
	}

	var category string
	targets := map[string]bool{}
	switch r.Gender {
	case "male":
		category = "财星"
		targets["正财"] = true
		targets["偏财"] = true
	case "female":
		category = "官杀"
		targets["正官"] = true
		targets["七杀"] = true
	default:
		return SpouseStarSignal{Available: false}
	}

	sig := SpouseStarSignal{Available: true, Category: category}
	sig.DayBranchHiddenShiShen = append([]string(nil), r.DayZhiShiShen...)

	seen := map[string]bool{}
	addStar := func(name string) {
		if !seen[name] {
			seen[name] = true
			sig.StarNames = append(sig.StarNames, name)
		}
	}

	// 透干：四柱天干十神
	ganPillars := []struct {
		label string
		ss    string
	}{
		{"年干", r.YearGanShiShen},
		{"月干", r.MonthGanShiShen},
		{"日干", r.DayGanShiShen},
		{"时干", r.HourGanShiShen},
	}
	for _, p := range ganPillars {
		if targets[p.ss] {
			sig.Present = true
			sig.Visible = true
			sig.Positions = append(sig.Positions, p.label+"(透)")
			addStar(p.ss)
		}
	}

	// 藏干：四柱地支藏干十神
	zhiPillars := []struct {
		label    string
		isDayZhi bool
		ss       []string
	}{
		{"年支", false, r.YearZhiShiShen},
		{"月支", false, r.MonthZhiShiShen},
		{"日支", true, r.DayZhiShiShen},
		{"时支", false, r.HourZhiShiShen},
	}
	for _, p := range zhiPillars {
		for _, s := range p.ss {
			if targets[s] {
				sig.Present = true
				sig.Positions = append(sig.Positions, p.label+"(藏)")
				if p.isDayZhi {
					sig.InSpousePalace = true
				}
				addStar(s)
			}
		}
	}

	return sig
}
```

注意：`append([]string(nil), nil...)` 在无元素时仍返回 `nil`，因此「配偶星不现」用例的 `StarNames`/`Positions` 保持 `nil`，与测试的 `want` 零值切片一致（`reflect.DeepEqual(nil slice, nil slice)` 为真）。

- [ ] **Step 4: 跑测试确认通过**

Run: `go test ./pkg/bazi/ -run TestDetectSpouseStarSignal -v`
Expected: PASS（7 个子用例全过）。

- [ ] **Step 5: 提交**

```bash
git add backend/pkg/bazi/compatibility_spouse.go backend/pkg/bazi/compatibility_spouse_test.go
git commit -m "feat(bazi): add spouse-star locator (配偶星定位) for spouse-palace match"
```

---

## Task 2: 把配偶信号拼进命盘摘要（service）

**Files:**
- Modify: `backend/internal/service/compatibility_service.go`（`compatibilityParticipantSummary` 约 591–622；新增 helper）
- Test: `backend/internal/service/compatibility_service_test.go`

**背景：** `compatibilityParticipantSummary(p *model.CompatibilityParticipant) (string, error)` 从 `p.ChartSnapshot` 反序列化出 `bazi.BaziResult`，拼成每人摘要串，最终流入 prompt 的 `{{.SelfChartSummary}}`/`{{.PartnerChartSummary}}`（`compatibility_service.go:540-568`）。摘要末尾追加一句「配偶画像信号」即可被 LLM 读到，无需新增模板变量。`zeroDash`（同文件 633）把空串显示为 `—`。快照里 `Gender` 可能为空，需从 `p.BirthProfile.Gender` 兜底。

- [ ] **Step 1: 写失败测试**

在 `backend/internal/service/compatibility_service_test.go` 末尾追加：

```go
func TestSpousePortraitSignalText_Branches(t *testing.T) {
	// 可用 + 有配偶星
	present := spousePortraitSignalText("我", bazi.SpouseStarSignal{
		Available: true, Present: true, Category: "财星",
		StarNames:              []string{"正财"},
		Positions:              []string{"月干(透)"},
		Visible:                true,
		DayBranchHiddenShiShen: []string{"伤官"},
	})
	for _, want := range []string{"配偶星(财星)", "正财", "月干(透)", "透干", "夫妻宫(日支)藏干十神="} {
		if !strings.Contains(present, want) {
			t.Errorf("present 分支缺 %q；got: %s", want, present)
		}
	}

	// 可用 + 配偶星不现
	absent := spousePortraitSignalText("我", bazi.SpouseStarSignal{
		Available: true, Present: false, Category: "财星",
		DayBranchHiddenShiShen: []string{"劫财"},
	})
	if !strings.Contains(absent, "配偶星(财星)不现") {
		t.Errorf("absent 分支缺『不现』；got: %s", absent)
	}

	// 不可用（缺性别）
	missing := spousePortraitSignalText("我", bazi.SpouseStarSignal{Available: false})
	if !strings.Contains(missing, "性别缺失，无法定配偶星") {
		t.Errorf("missing 分支缺『性别缺失』；got: %s", missing)
	}
}

func TestCompatibilityParticipantSummary_AppendsSpouseSignal(t *testing.T) {
	snapshot, _ := json.Marshal(bazi.BaziResult{
		YearGan: "甲", YearZhi: "子",
		MonthGan: "辛", MonthZhi: "未",
		DayGan: "甲", DayZhi: "子",
		HourGan: "丁", HourZhi: "卯",
		MonthGanShiShen: "正官", // 注意：本盘用作男命，正官非财；改用月支藏财演示
		DayZhiShiShen:   []string{"正印"},
		MonthZhiShiShen: []string{"正财"},
		Gender:          "male",
	})
	raw := json.RawMessage(snapshot)
	p := &model.CompatibilityParticipant{DisplayName: "我", ChartSnapshot: &raw}

	summary, err := compatibilityParticipantSummary(p)
	if err != nil {
		t.Fatalf("compatibilityParticipantSummary error: %v", err)
	}
	for _, want := range []string{"配偶画像信号", "配偶星(财星)", "正财"} {
		if !strings.Contains(summary, want) {
			t.Errorf("summary 缺 %q；got: %s", want, summary)
		}
	}
}
```

（`strings`、`json`、`bazi`、`model` 在该测试文件已 import；若 `bazi` 未引入，按现有 import 块补 `"yuanju/pkg/bazi"`。）

- [ ] **Step 2: 跑测试确认失败**

Run: `go test ./internal/service/ -run 'TestSpousePortraitSignalText_Branches|TestCompatibilityParticipantSummary_AppendsSpouseSignal' -v`
Expected: 编译失败 / `undefined: spousePortraitSignalText`。

- [ ] **Step 3: 写实现**

在 `backend/internal/service/compatibility_service.go` 新增 helper（放在 `compatibilityParticipantSummary` 之后、`compatibilityStrengthLabels` 之前）：

```go
// spousePortraitSignalText 把配偶星信号格式化为一句话，拼进命盘摘要供 LLM 推配偶画像。
func spousePortraitSignalText(name string, sig bazi.SpouseStarSignal) string {
	if !sig.Available {
		return fmt.Sprintf("%s 配偶画像信号：性别缺失，无法定配偶星，本节跳过。", name)
	}
	hidden := zeroDash(strings.Join(sig.DayBranchHiddenShiShen, ","))
	if !sig.Present {
		return fmt.Sprintf("%s 配偶画像信号：配偶星(%s)不现；夫妻宫(日支)藏干十神=%s。", name, sig.Category, hidden)
	}
	visible := "未透干"
	if sig.Visible {
		visible = "透干"
	}
	palace := "未坐夫妻宫"
	if sig.InSpousePalace {
		palace = "坐夫妻宫(日支)"
	}
	return fmt.Sprintf("%s 配偶画像信号：配偶星(%s)=%s；位置=%s；%s；%s；夫妻宫(日支)藏干十神=%s。",
		name, sig.Category,
		zeroDash(strings.Join(sig.StarNames, ",")),
		zeroDash(strings.Join(sig.Positions, ",")),
		visible, palace, hidden)
}
```

再修改 `compatibilityParticipantSummary` 的 `return`（约 610–621），把原本直接 `return fmt.Sprintf(...), nil` 改为先存到变量再追加配偶信号。将：

```go
	return fmt.Sprintf(
		"%s：%s%s·%s%s·%s%s·%s%s；日主=%s；五行=%d木/%d火/%d土/%d金/%d水；十神=%s；命格=%s；旺衰=%s；用神=%s；忌神=%s。",
		p.DisplayName,
		result.YearGan, result.YearZhi,
		result.MonthGan, result.MonthZhi,
		result.DayGan, result.DayZhi,
		result.HourGan, result.HourZhi,
		result.DayGan,
		result.Wuxing.Mu, result.Wuxing.Huo, result.Wuxing.Tu, result.Wuxing.Jin, result.Wuxing.Shui,
		shishen, zeroDash(result.MingGe), strength,
		result.Yongshen, result.Jishen,
	), nil
```

改为：

```go
	base := fmt.Sprintf(
		"%s：%s%s·%s%s·%s%s·%s%s；日主=%s；五行=%d木/%d火/%d土/%d金/%d水；十神=%s；命格=%s；旺衰=%s；用神=%s；忌神=%s。",
		p.DisplayName,
		result.YearGan, result.YearZhi,
		result.MonthGan, result.MonthZhi,
		result.DayGan, result.DayZhi,
		result.HourGan, result.HourZhi,
		result.DayGan,
		result.Wuxing.Mu, result.Wuxing.Huo, result.Wuxing.Tu, result.Wuxing.Jin, result.Wuxing.Shui,
		shishen, zeroDash(result.MingGe), strength,
		result.Yongshen, result.Jishen,
	)

	// 配偶画像信号：快照无性别时从出生档案兜底
	if result.Gender == "" {
		result.Gender = p.BirthProfile.Gender
	}
	sig := bazi.DetectSpouseStarSignal(&result)
	return base + " " + spousePortraitSignalText(p.DisplayName, sig), nil
```

- [ ] **Step 4: 跑测试确认通过**

Run: `go test ./internal/service/ -run 'TestSpousePortraitSignalText_Branches|TestCompatibilityParticipantSummary_AppendsSpouseSignal|TestCompatibilityParticipantSummary' -v`
Expected: PASS，且既有 `TestCompatibilityParticipantSummary_ValidSnapshot`、`TestCompatibilityParticipantSummary_IncludesPersonalitySignals` 仍通过（它们用 `strings.Contains`，追加句子不影响）。

- [ ] **Step 5: 提交**

```bash
git add backend/internal/service/compatibility_service.go backend/internal/service/compatibility_service_test.go
git commit -m "feat(service): append spouse-star signal to compatibility chart summary"
```

---

## Task 3: model 承接 spouse_palace_match 输出

**Files:**
- Modify: `backend/internal/model/compatibility.go`（在 `CompatibilityPersonalityComparison` 之后新增结构；在 `CompatibilityStructuredReport` 加字段，约 203–217）
- Test: `backend/internal/service/compatibility_service_test.go`

**背景：** LLM 返回的报告 JSON 直接 `json.Unmarshal` 进 `model.CompatibilityStructuredReport`（`compatibility_service.go:358`）。新字段用指针 + `omitempty`，旧报告无此字段时解析为 `nil`（与 `PersonalityComparison` 一致），不报错。

- [ ] **Step 1: 写失败测试**

在 `backend/internal/service/compatibility_service_test.go` 末尾追加：

```go
func TestCompatibilityStructuredReport_SpousePalaceMatchParsing(t *testing.T) {
	withField := `{"summary":"s","spouse_palace_match":{` +
		`"self":{"ideal_portrait":"A理想","match_level":"medium","fit_points":["温和"],"gap_points":["急躁"],"evidence_keys":["day_pillar_upper"]},` +
		`"partner":{"ideal_portrait":"B理想","match_level":"low","fit_points":[],"gap_points":[],"evidence_keys":[]},` +
		`"summary":"双向偏弱"}}`
	var r1 model.CompatibilityStructuredReport
	if err := json.Unmarshal([]byte(withField), &r1); err != nil {
		t.Fatalf("unmarshal with field: %v", err)
	}
	if r1.SpousePalaceMatch == nil {
		t.Fatal("expected non-nil SpousePalaceMatch")
	}
	if r1.SpousePalaceMatch.Self.MatchLevel != "medium" {
		t.Errorf("self match_level = %q", r1.SpousePalaceMatch.Self.MatchLevel)
	}
	if r1.SpousePalaceMatch.Self.IdealPortrait != "A理想" {
		t.Errorf("self ideal_portrait = %q", r1.SpousePalaceMatch.Self.IdealPortrait)
	}
	if len(r1.SpousePalaceMatch.Self.FitPoints) != 1 {
		t.Errorf("self fit_points len = %d", len(r1.SpousePalaceMatch.Self.FitPoints))
	}
	if r1.SpousePalaceMatch.Summary != "双向偏弱" {
		t.Errorf("summary = %q", r1.SpousePalaceMatch.Summary)
	}

	// 旧报告无该字段 → nil，不报错
	var r2 model.CompatibilityStructuredReport
	if err := json.Unmarshal([]byte(`{"summary":"s"}`), &r2); err != nil {
		t.Fatalf("unmarshal without field: %v", err)
	}
	if r2.SpousePalaceMatch != nil {
		t.Error("expected nil SpousePalaceMatch for legacy report")
	}
}
```

- [ ] **Step 2: 跑测试确认失败**

Run: `go test ./internal/service/ -run TestCompatibilityStructuredReport_SpousePalaceMatchParsing -v`
Expected: 编译失败 / `r1.SpousePalaceMatch undefined`。

- [ ] **Step 3: 写实现**

在 `backend/internal/model/compatibility.go` 的 `CompatibilityPersonalityComparison` 结构（约 196–201）之后新增：

```go
type CompatibilitySpousePalaceSide struct {
	IdealPortrait string   `json:"ideal_portrait"`
	MatchLevel    string   `json:"match_level"`
	FitPoints     []string `json:"fit_points"`
	GapPoints     []string `json:"gap_points"`
	EvidenceKeys  []string `json:"evidence_keys"`
}

type CompatibilitySpousePalaceMatch struct {
	Self    CompatibilitySpousePalaceSide `json:"self"`
	Partner CompatibilitySpousePalaceSide `json:"partner"`
	Summary string                        `json:"summary"`
}
```

并在 `CompatibilityStructuredReport`（约 203–217）的 `PersonalityComparison` 字段之后加一行：

```go
	SpousePalaceMatch     *CompatibilitySpousePalaceMatch     `json:"spouse_palace_match,omitempty"`
```

- [ ] **Step 4: 跑测试确认通过**

Run: `go test ./internal/service/ -run TestCompatibilityStructuredReport_SpousePalaceMatchParsing -v`
Expected: PASS。

- [ ] **Step 5: 提交**

```bash
git add backend/internal/model/compatibility.go backend/internal/service/compatibility_service_test.go
git commit -m "feat(model): parse spouse_palace_match in compatibility structured report"
```

---

## Task 4: 更新 prompt 模板（输出字段 + 约束 + 版本号）

**Files:**
- Modify: `backend/pkg/prompt/canonical_compatibility.go`
- Test: `backend/pkg/prompt/canonical_test.go`

**背景：** canonical prompt 是出厂版（`pkg/prompt/canonical_compatibility.go`）。版本号当前 `v3.1-question-aware-5`（line 6）。`drift_test.go`/`sync_test.go`/`canonical_compatibility_famous_couple_test.go` 都用 `def.Content`/`def.Hash` 动态比较，内容改了会自动重算 hash，无需改这三个；**唯一硬编码版本号的是 `canonical_test.go`**。配偶信号已随 `{{.SelfChartSummary}}`/`{{.PartnerChartSummary}}` 进入 prompt（Task 2），故本任务**不加模板变量**，只加输出 schema 与生成约束。

- [ ] **Step 1: 写失败测试**

编辑 `backend/pkg/prompt/canonical_test.go`：

1）把 `TestMustGet_CompatibilityReturnsRegisteredDefinition` 里的版本断言

```go
	if def.Version != "v3.1-question-aware-5" {
		t.Errorf("expected Version v3.1-question-aware-5, got %q", def.Version)
	}
```

改为：

```go
	if def.Version != "v3.1-question-aware-6" {
		t.Errorf("expected Version v3.1-question-aware-6, got %q", def.Version)
	}
```

2）在同一测试的 `for _, want := range []string{"question_focus", ...}` 列表里加入 `"spouse_palace_match"` 与 `"夫妻宫匹配约束"`：

```go
	for _, want := range []string{"question_focus", "decision_advice", "personality_comparison", "表达约束", "{{.PrimaryQuestionLabel}}", "spouse_palace_match", "夫妻宫匹配约束"} {
```

- [ ] **Step 2: 跑测试确认失败**

Run: `go test ./pkg/prompt/ -run TestMustGet_CompatibilityReturnsRegisteredDefinition -v`
Expected: FAIL —— 版本仍是 `-5`，且内容不含 `spouse_palace_match`/`夫妻宫匹配约束`。

- [ ] **Step 3: 写实现**

编辑 `backend/pkg/prompt/canonical_compatibility.go`：

3a）版本号（line 5–6）：把

```go
		Version:     "v3.1-question-aware-5",
		Description: "合盘决策咨询 prompt（含 question_focus / decision_advice / stage_risks / personality_comparison）",
```

改为：

```go
		Version:     "v3.1-question-aware-6",
		Description: "合盘决策咨询 prompt（含 question_focus / decision_advice / stage_risks / personality_comparison / spouse_palace_match）",
```

3b）新增生成约束：在 `输出严格为 JSON：` 这一行**之前**插入一段（紧接「关系经营策略·沟通」block 之后）：

```
夫妻宫匹配约束（spouse_palace_match，务必遵守）：
- 依据每人命盘摘要里的「配偶画像信号」（配偶星 = 男看财星 / 女看官杀，及其位置、透藏、是否坐夫妻宫、日支藏干十神）推出「这个人命里理想 / 容易吸引的另一半画像」，再拿对方真实的 personality_comparison 画像去比，给出像在哪（fit_points）、差在哪（gap_points）。
- self 子块 = 用 A 的配偶画像信号推理想另一半、拿 B 的真实画像比；partner 子块 = 用 B 推、拿 A 比。
- match_level 取 high / medium / low，且必须与已知夫妻宫状态自洽：当 day_pillar 维度或 negative 证据显示双方日支相冲 / 相克 / 相刑 / 相害时，不得给 high；严禁与负面证据相矛盾。
- 若某人「配偶画像信号」为「性别缺失，无法定配偶星」，其对应子块的 ideal_portrait 写明「缺性别，无法定配偶星」、match_level 置空、fit_points / gap_points 为空数组，并在 summary 注明本节因缺性别跳过。
- 若配偶星「不现」，照样给画像但注明「配偶星不显，结论偏轮廓」。
- 文字守表达约束：术语后跟大白话、条件语气不下死命；画像尺度微辣不露骨、不越线。

```

3c）新增输出字段：在输出 JSON 骨架里，`"personality_comparison": { ... }` 的闭合 `},`（即 `clash_points` 数组闭合后的 `},`，line 146）与 `"decision_advice": {`（line 147）之间，插入：

```
  "spouse_palace_match": {
    "self": {
      "ideal_portrait": "A 命里理想 / 容易吸引的另一半画像（基于配偶星 + 夫妻宫藏干）",
      "match_level": "high|medium|low",
      "fit_points": ["B 哪里对上了 A 的理想"],
      "gap_points": ["B 哪里和 A 的理想有差距"],
      "evidence_keys": ["支撑该判断的 evidence_key"]
    },
    "partner": {
      "ideal_portrait": "B 命里理想 / 容易吸引的另一半画像",
      "match_level": "high|medium|low",
      "fit_points": ["A 哪里对上了 B 的理想"],
      "gap_points": ["A 哪里和 B 的理想有差距"],
      "evidence_keys": []
    },
    "summary": "一句话总括双向夫妻宫匹配（含缺性别 / 配偶星不现的说明）"
  },
```

实现者注意：用 Edit 工具时以唯一锚点定位——3c 的 `old_string` 取

```
  },
  "decision_advice": {
```

（这是 personality_comparison 闭合 + decision_advice 起始；在该 `old_string` 的 `},` 与 `"decision_advice"` 之间插入新块）。若该两行组合在文件中不唯一，扩大锚点至包含上一行 `    ]` （clash_points 数组闭合）。

- [ ] **Step 4: 跑测试确认通过**

Run: `go test ./pkg/prompt/ -v`
Expected: 全 PASS —— 含 `canonical_test`、`drift_test`、`sync_test`、`canonical_compatibility_famous_couple_test`（后三者读 `def.Content`/`def.Hash` 动态比较，自动适配新内容）。

- [ ] **Step 5: 查无遗漏的旧版本号引用**

Run: `grep -rn "v3.1-question-aware-5" backend/`
Expected: 无输出（确认没有别处仍 pin 旧版本）。若有命中，逐一改为 `-6` 并说明。

- [ ] **Step 6: 提交**

```bash
git add backend/pkg/prompt/canonical_compatibility.go backend/pkg/prompt/canonical_test.go
git commit -m "feat(prompt): add spouse_palace_match output + constraints, bump v3.1-question-aware-6"
```

---

## Task 5: 全量回归 + 评分不变量确认

**Files:** 无改动，仅验证。

- [ ] **Step 1: 跑全量测试**

Run: `go test ./...`
Expected: 全 PASS。**唯一可忽略**：`internal/repository` 中 `TestGetTokenUsageCostByModel_AggregatesGroupedRows` 在无实时 DB 时 panic（nil `*sql.DB`）——与本改动无关（本改动不碰 token_usage）。其余必须全绿。

- [ ] **Step 2: 确认评分未被触碰**

Run: `go test ./pkg/bazi/ -run TestAnalyzeCompatibilitySurfacesNegativesWithoutScoreChange -v`
Expected: PASS —— 触发盘 `OverallScore == 34`、等级 low、`DimensionScores` 不变。本改动新增的是独立纯函数 + 序列化 + prompt + model 字段，从未进入 `AnalyzeCompatibility` 的打分路径，此测试即评分红线的回归保障。

- [ ] **Step 3: 无新提交**（本任务仅验证；若上面任一失败，回到对应 Task 修复后再继续）

---

## 验收（成功标准，对齐 spec §7）

1. `go test ./...` 全绿（token_usage nil-DB 测试除外）。
2. 一对带性别的盘，报告 `spouse_palace_match` 双向出 `ideal_portrait` + `match_level`，且 `match_level` 与夫妻宫合冲状态自洽（相冲不给 high）。
3. 缺性别一方按约定跳过并注明。
4. 总分不变（Task 5 Step 2 保障）。

## 明确不做（YAGNI，对齐 spec §6）

- 不进 0–100 总分、不建「十神→性格」对照表、不引入真易经卦象、不做公式化相似度、不动前端展示组件。
