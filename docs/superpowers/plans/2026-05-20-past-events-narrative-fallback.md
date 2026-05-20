# Past-Events Narrative Fallback Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 保证 `ai_dayun_summaries.years` 中每个流年 `narrative` 入库均非空，消除 v3-progressive-compressed 下 10% 的"空白流年"。

**Architecture:** 双层防御 —— (1) `pkg/bazi.RenderYearNarrativeWithFallback` 包装现有 `RenderYearNarrative`，在 AI 输出空或 validator 清空后用 polarity-driven 安全句式兜底；(2) `GenerateDayunSummariesStream` 的 prompt 追加"弱信号年安全措辞"指引，让 AI 更愿意自然填充。

**Tech Stack:** Go 1.21+, 现有 `pkg/bazi` 包 + `internal/service/report_service.go` + `internal/repository/algorithm_version.go`。无新依赖、无 DB migration、无前端改动。

**Reference spec:** `docs/superpowers/specs/2026-05-20-past-events-narrative-fallback-design.md`

---

## File Structure

| File | Action | Responsibility |
|------|--------|----------------|
| `backend/pkg/bazi/event_narrative.go` | Modify (append) | 添加 `RenderYearNarrativeWithFallback` 和 `makeMinimalFallback` 两个函数 |
| `backend/pkg/bazi/event_narrative_test.go` | Modify (append) | 三个新测试：always non-empty / polarity 分支 / keyword safe |
| `backend/internal/service/report_service.go` | Modify (中部) | (a) 抽出 `fillBlankYearNarratives` 私有函数 (b) 在 prompt 模板追加"弱信号年安全措辞"块 |
| `backend/internal/service/report_service_test.go` | Modify (append) | `fillBlankYearNarratives` 的单元测试（空 narrative 走兜底） |
| `backend/internal/repository/algorithm_version.go` | Modify (1 行) | 升版至 `v3.1-narrative-guarded` 并补充版本历史注释 |

---

## Task 1: 在 pkg/bazi 中实现 `RenderYearNarrativeWithFallback`（包装器）

**Files:**
- Modify: `backend/pkg/bazi/event_narrative.go` (append at end of file, 当前 776 行)
- Test: `backend/pkg/bazi/event_narrative_test.go` (append)

- [ ] **Step 1.1: 添加失败的"包装器在空输入时返回非空"测试**

追加到 `backend/pkg/bazi/event_narrative_test.go` 末尾：

```go
func TestRenderYearNarrativeWithFallback_NoSignalsReturnsNonEmpty(t *testing.T) {
	// 0 signals — RenderYearNarrative 返 ""，Fallback 必须兜底
	ys := YearSignals{Year: 2022, Age: 27, GanZhi: "壬寅", DayunGanZhi: "辛丑"}
	got := RenderYearNarrativeWithFallback(ys)
	if got == "" {
		t.Fatal("expected non-empty fallback for no-signals year")
	}
}

func TestRenderYearNarrativeWithFallback_AnchoredYearDelegatesToOriginal(t *testing.T) {
	// 有真实信号的年应原样返回 RenderYearNarrative 的输出（不调兜底）
	ys := YearSignals{
		Year:   2026,
		Age:    31,
		GanZhi: "丙午",
		Signals: []EventSignal{
			{Type: "健康", Evidence: "流年地支午冲日支子，日柱受冲，体力精神有下滑风险", Polarity: PolarityXiong, Source: SourceZhuwei},
		},
	}
	want := RenderYearNarrative(ys)
	if want == "" {
		t.Fatal("test fixture invalid: RenderYearNarrative returned empty for anchored year")
	}
	got := RenderYearNarrativeWithFallback(ys)
	if got != want {
		t.Errorf("expected wrapper to return original narrative; got %q want %q", got, want)
	}
}
```

- [ ] **Step 1.2: 跑测试确认失败**

```bash
cd backend && go test ./pkg/bazi/... -run TestRenderYearNarrativeWithFallback -v
```

期望：`undefined: RenderYearNarrativeWithFallback`（编译失败）

- [ ] **Step 1.3: 实现 `RenderYearNarrativeWithFallback` + `makeMinimalFallback`**

追加到 `backend/pkg/bazi/event_narrative.go` 末尾：

```go
// RenderYearNarrativeWithFallback 是 RenderYearNarrative 的非空包装。
// AI 模式下，当 AI 输出 narrative="" 或 validator 清空 narrative 后调用，
// 保证流年卡片始终有内容，避免前端渲染光秃秃的干支头。
//
// 调用方在确实需要"每年都有内容"时使用；不替代 RenderYearNarrative —
// template 模式仍可保留"<2 句返空"的契约。
func RenderYearNarrativeWithFallback(ys YearSignals) string {
	if s := RenderYearNarrative(ys); s != "" {
		return s
	}
	return makeMinimalFallback(ys)
}

// makeMinimalFallback 按 用神基底 信号的 polarity 选择安全句式。
//
// 严禁包含 ValidateYearNarrative 的 28 个 validatedKeywords
// （用神位/忌神位/伏吟/反吟/神煞名 等）——本兜底文案的目的就是绕开
// 那些需要 evidence 追溯的强术语。
func makeMinimalFallback(ys YearSignals) string {
	polarity := ""
	for _, s := range ys.Signals {
		if s.Source == SourceYongshen {
			polarity = s.Polarity
			break
		}
	}
	switch polarity {
	case PolarityJi:
		return fmt.Sprintf("%d岁%s年，命局基调向吉，本年信号较稀，运势相对平顺，宜稳中求进。", ys.Age, ys.GanZhi)
	case PolarityXiong:
		return fmt.Sprintf("%d岁%s年，命局基调偏凶，本年信号较稀，宜守不宜攻，谨慎处事。", ys.Age, ys.GanZhi)
	default:
		dy := ys.DayunGanZhi
		if dy == "" {
			return fmt.Sprintf("%d岁%s年信号较稀，命局本年无明显波动。", ys.Age, ys.GanZhi)
		}
		return fmt.Sprintf("%d岁%s年信号较稀，运势相对平稳，按本段大运%s方向延展即可。", ys.Age, ys.GanZhi, dy)
	}
}
```

如果 `event_narrative.go` 顶部尚无 `import "fmt"`，在 import 块加入 `"fmt"`。

- [ ] **Step 1.4: 跑测试确认通过**

```bash
cd backend && go test ./pkg/bazi/... -run TestRenderYearNarrativeWithFallback -v
```

期望：`PASS` 两个测试。

- [ ] **Step 1.5: 跑全包测试确认无回归**

```bash
cd backend && go test ./pkg/bazi/... -count=1
```

期望：所有现有 `TestRenderYearNarrative_*` 测试不受影响。

- [ ] **Step 1.6: Commit**

```bash
cd backend && git add pkg/bazi/event_narrative.go pkg/bazi/event_narrative_test.go
git commit -m "feat(bazi): RenderYearNarrativeWithFallback polarity-driven safety net"
```

---

## Task 2: 兜底文案的 polarity 分支 + 关键词安全测试

**Files:**
- Test: `backend/pkg/bazi/event_narrative_test.go` (append)

- [ ] **Step 2.1: 添加 polarity 分支测试**

追加到 `backend/pkg/bazi/event_narrative_test.go` 末尾：

```go
func TestMakeMinimalFallback_PolarityJi(t *testing.T) {
	ys := YearSignals{Year: 2020, Age: 25, GanZhi: "庚子", DayunGanZhi: "甲寅",
		Signals: []EventSignal{{Type: "用神基底", Source: SourceYongshen, Polarity: PolarityJi}}}
	got := makeMinimalFallback(ys)
	if !strings.Contains(got, "向吉") {
		t.Errorf("ji polarity should mention '向吉'; got %q", got)
	}
	if !strings.Contains(got, "庚子") {
		t.Errorf("expected ganzhi in output; got %q", got)
	}
}

func TestMakeMinimalFallback_PolarityXiong(t *testing.T) {
	ys := YearSignals{Year: 2021, Age: 26, GanZhi: "辛丑", DayunGanZhi: "甲寅",
		Signals: []EventSignal{{Type: "用神基底", Source: SourceYongshen, Polarity: PolarityXiong}}}
	got := makeMinimalFallback(ys)
	if !strings.Contains(got, "偏凶") {
		t.Errorf("xiong polarity should mention '偏凶'; got %q", got)
	}
}

func TestMakeMinimalFallback_NeutralWithDayun(t *testing.T) {
	ys := YearSignals{Year: 2022, Age: 27, GanZhi: "壬寅", DayunGanZhi: "甲寅",
		Signals: []EventSignal{{Type: "用神基底", Source: SourceYongshen, Polarity: PolarityNeutral}}}
	got := makeMinimalFallback(ys)
	if !strings.Contains(got, "按本段大运甲寅方向延展") {
		t.Errorf("neutral polarity with dayun should reference dayun ganzhi; got %q", got)
	}
}

func TestMakeMinimalFallback_NoBasisNoDayun(t *testing.T) {
	// 没有 SourceYongshen 信号、也没有 DayunGanZhi
	ys := YearSignals{Year: 2023, Age: 28, GanZhi: "癸卯"}
	got := makeMinimalFallback(ys)
	if !strings.Contains(got, "无明显波动") {
		t.Errorf("absent dayun should fall to 无明显波动 phrasing; got %q", got)
	}
	if got == "" {
		t.Fatal("must not return empty")
	}
}
```

- [ ] **Step 2.2: 添加关键词安全测试**

```go
func TestMakeMinimalFallback_KeywordSafe(t *testing.T) {
	// 兜底文案的任何 polarity 分支都禁止触发 ValidateYearNarrative 的 28 个关键词。
	forbidden := []string{
		"用神位", "忌神位", "喜神位",
		"伏吟", "反吟", "大运合化", "三会", "三合",
		"受冲", "受刑", "双重命中", "力度倍增",
		"驿马", "桃花", "华盖", "白虎", "丧门", "吊客", "灾煞", "流霞",
		"天医", "天喜", "天乙", "天德", "月德", "文昌", "太极", "福星",
		"红艳", "孤辰", "寡宿", "羊刃", "亡神", "劫煞", "披麻", "咸池",
		"勾绞", "国印",
	}
	cases := []YearSignals{
		{Year: 2020, Age: 25, GanZhi: "庚子", DayunGanZhi: "甲寅",
			Signals: []EventSignal{{Type: "用神基底", Source: SourceYongshen, Polarity: PolarityJi}}},
		{Year: 2021, Age: 26, GanZhi: "辛丑", DayunGanZhi: "甲寅",
			Signals: []EventSignal{{Type: "用神基底", Source: SourceYongshen, Polarity: PolarityXiong}}},
		{Year: 2022, Age: 27, GanZhi: "壬寅", DayunGanZhi: "甲寅",
			Signals: []EventSignal{{Type: "用神基底", Source: SourceYongshen, Polarity: PolarityNeutral}}},
		{Year: 2023, Age: 28, GanZhi: "癸卯"},
	}
	for _, ys := range cases {
		got := makeMinimalFallback(ys)
		for _, kw := range forbidden {
			if strings.Contains(got, kw) {
				t.Errorf("fallback for %s contains forbidden keyword %q: %q", ys.GanZhi, kw, got)
			}
		}
	}
}
```

- [ ] **Step 2.3: 跑测试确认通过**

```bash
cd backend && go test ./pkg/bazi/... -run "TestMakeMinimalFallback" -v
```

期望：全部 5 个测试 PASS。

- [ ] **Step 2.4: Commit**

```bash
cd backend && git add pkg/bazi/event_narrative_test.go
git commit -m "test(bazi): makeMinimalFallback polarity branches + keyword safety"
```

---

## Task 3: 抽出 `fillBlankYearNarratives` 私有函数（可测试性重构）

抽出 `report_service.go:1421-1439` 的循环体到一个独立函数，便于单测。

**Files:**
- Modify: `backend/internal/service/report_service.go` (line 1421-1439 区域)
- Test: `backend/internal/service/report_service_test.go` (append)

- [ ] **Step 3.1: 添加失败的 helper 函数测试**

追加到 `backend/internal/service/report_service_test.go` 末尾：

```go
func TestFillBlankYearNarratives_EmptyNarrativeGetsFallback(t *testing.T) {
	parsed := []parsedYearAI{
		{Year: 2020, GanZhi: "庚子", Narrative: ""},
	}
	signals := []bazi.YearSignals{
		{Year: 2020, Age: 25, GanZhi: "庚子", DayunGanZhi: "甲寅",
			Signals: []bazi.EventSignal{
				{Type: "用神基底", Source: bazi.SourceYongshen, Polarity: bazi.PolarityJi},
			}},
	}
	out := fillBlankYearNarratives(parsed, signals, 1)
	if len(out) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(out))
	}
	if out[0].Narrative == "" {
		t.Error("blank AI narrative should be filled by fallback")
	}
	if !strings.Contains(out[0].Narrative, "庚子") {
		t.Errorf("fallback should reference ganzhi; got %q", out[0].Narrative)
	}
}

func TestFillBlankYearNarratives_ValidAIPreserved(t *testing.T) {
	parsed := []parsedYearAI{
		{Year: 2020, GanZhi: "庚子", Narrative: "庚子年食神高透，事业稳步推进。"},
	}
	signals := []bazi.YearSignals{
		{Year: 2020, Age: 25, GanZhi: "庚子",
			Signals: []bazi.EventSignal{
				{Type: "事业", Evidence: "食神高透", Polarity: bazi.PolarityJi, Source: "天干"},
			}},
	}
	out := fillBlankYearNarratives(parsed, signals, 1)
	if out[0].Narrative != "庚子年食神高透，事业稳步推进。" {
		t.Errorf("valid AI narrative should be preserved verbatim; got %q", out[0].Narrative)
	}
}

func TestFillBlankYearNarratives_ValidatorWipedGetsFallback(t *testing.T) {
	// AI 写了"用神位受冲"但 evidence 没有"用神位" → validator 清空 → 兜底
	parsed := []parsedYearAI{
		{Year: 2020, GanZhi: "庚子", Narrative: "庚子年用神位受冲，运势波动。"},
	}
	signals := []bazi.YearSignals{
		{Year: 2020, Age: 25, GanZhi: "庚子",
			Signals: []bazi.EventSignal{
				{Type: "用神基底", Source: bazi.SourceYongshen, Polarity: bazi.PolarityXiong, Evidence: "日干受克"},
			}},
	}
	out := fillBlankYearNarratives(parsed, signals, 1)
	if out[0].Narrative == "庚子年用神位受冲，运势波动。" {
		t.Error("validator should have wiped the narrative")
	}
	if out[0].Narrative == "" {
		t.Error("wiped narrative should be replaced by fallback, not left empty")
	}
	if !strings.Contains(out[0].Narrative, "偏凶") {
		t.Errorf("xiong basis should produce 偏凶 fallback; got %q", out[0].Narrative)
	}
}
```

确保测试文件 import 包含 `"strings"` 和 `"yuanju/pkg/bazi"`（应已存在，按需补全）。

- [ ] **Step 3.2: 跑测试确认失败**

```bash
cd backend && go test ./internal/service/... -run TestFillBlankYearNarratives -v
```

期望：`undefined: fillBlankYearNarratives` / `undefined: parsedYearAI`（编译失败）

- [ ] **Step 3.3: 在 report_service.go 中提取 helper 函数**

在 `report_service.go` 中找到 `GenerateDayunSummariesStream` 上方（接近 line 870 之前任意位置，或者把它放在文件末尾），添加：

```go
// parsedYearAI 表示 AI 输出 JSON 中单个 year 项目。
// 抽出为命名类型让 fillBlankYearNarratives 可独立测试。
type parsedYearAI struct {
	Year      int    `json:"year"`
	GanZhi    string `json:"ganzhi"`
	Narrative string `json:"narrative"`
}

// yearOut 是入库到 ai_dayun_summaries.years 的最终 JSON shape。
type yearOut struct {
	Year      int    `json:"year"`
	GanZhi    string `json:"ganzhi"`
	Narrative string `json:"narrative"`
}

// fillBlankYearNarratives 校验 AI 输出的逐年 narrative，并对空字符串/校验失败的年
// 调用 RenderYearNarrativeWithFallback 进行兜底，保证每年 narrative 非空。
//
// parsed 与 signals 必须等长且 ganzhi 一一对应（已由上游 GenerateDayunSummariesStream 校验）。
// dayunIndex 仅用于日志定位。
func fillBlankYearNarratives(parsed []parsedYearAI, signals []bazi.YearSignals, dayunIndex int) []yearOut {
	out := make([]yearOut, len(parsed))
	for i, y := range parsed {
		narrative := y.Narrative
		if narrative != "" {
			if ok, reason := ValidateYearNarrative(narrative, signals[i].Signals); !ok {
				log.Printf("[fillBlankYearNarratives] dayun=%d year=%d 校验失败丢弃 narrative：%s",
					dayunIndex, y.Year, reason)
				narrative = ""
			}
		}
		if narrative == "" {
			narrative = bazi.RenderYearNarrativeWithFallback(signals[i])
			log.Printf("[fillBlankYearNarratives] dayun=%d year=%d 使用 template 兜底",
				dayunIndex, y.Year)
		}
		out[i] = yearOut{Year: y.Year, GanZhi: y.GanZhi, Narrative: narrative}
	}
	return out
}
```

- [ ] **Step 3.4: 改写 `GenerateDayunSummariesStream` 调用点**

定位 `report_service.go` 中的：

```go
// 结构校验段（line 1378-1387）前后的 parsed struct 定义改为复用顶层 parsedYearAI：
var parsed struct {
    Themes  []string `json:"themes"`
    Summary string   `json:"summary"`
    Years   []parsedYearAI `json:"years"`
}
```

（也就是把原来内联的 `[]struct{Year int...}` 改成 `[]parsedYearAI`。）

然后定位 line 1421-1439 的循环和 `type yearOut struct` 定义，替换为：

```go
// 护栏 1：逐年校验 narrative；空 narrative 或校验失败的走 template 兜底。
validatedYears := fillBlankYearNarratives(parsed.Years, dySignals, dy.Index)
yearsJSON, _ := json.Marshal(validatedYears)
yearsRaw := json.RawMessage(yearsJSON)
```

注意：原来内联的 `type yearOut struct` 已删除（因为顶层已声明），原来 `validatedYears := make([]yearOut, ...)` 的循环逻辑全部移到 helper 内部。

- [ ] **Step 3.5: 跑测试确认通过**

```bash
cd backend && go test ./internal/service/... -run TestFillBlankYearNarratives -v
```

期望：3 个测试 PASS。

- [ ] **Step 3.6: 跑全包测试确认无回归**

```bash
cd backend && go test ./internal/service/... -count=1
```

期望：全部 PASS（特别注意 `TestDayunSummaryPrompt_*` 系列、`TestComputeAutoGenDayunIndexes_*` 系列、`TestCachedDayunSummaryToStreamItemReturnsCachedItem` 都不受影响）。

- [ ] **Step 3.7: Commit**

```bash
cd backend && git add internal/service/report_service.go internal/service/report_service_test.go
git commit -m "refactor(report): extract fillBlankYearNarratives helper + template fallback wired"
```

---

## Task 4: Prompt 加固"弱信号年安全措辞"块

**Files:**
- Modify: `backend/internal/service/report_service.go` (line 1204-1216 narrative 撰写规则块)

- [ ] **Step 4.1: 编辑 prompt 模板**

定位 `report_service.go` 中 `GenerateDayunSummariesStream` 的 prompt 模板，找到 narrative 撰写规则块（约 line 1204-1226）。在 `- 严禁编造未在 evidence 中出现的神煞或用神位事件` 行前插入：

```text
   - **弱信号年安全措辞**：若该年 evidence 仅含"用神基底"或基底+1 个弱信号，
     可直接使用以下安全句式，禁止省略 narrative：
     · 「该年信号稀疏，运势相对平顺」
     · 「基底为吉/凶/中性，本年无明显波动」
     · 「与大运 {大运干支} 同调，按本段方向延展」
     上述句式不含"用神位/忌神位/伏吟/反吟/神煞名"等需追溯术语，可安全使用。
```

具体的字符串拼接形式遵循该 prompt 已使用的格式（单引号包围中文括号原样保留）。这是 Go 反引号原始字符串字面量，可直接粘贴中文行。

- [ ] **Step 4.2: 跑全包测试确认 prompt fixture 不破坏**

```bash
cd backend && go test ./internal/service/... -count=1
```

期望：全部 PASS。注意 `shishenInjectionTplFixture`（line 440 附近）是 ShishenConfidence 块的独立 fixture，**不**包含 narrative 撰写规则块，因此不受影响。

- [ ] **Step 4.3: 编译确认 Go 字符串字面量合法**

```bash
cd backend && go build ./...
```

期望：无错误。

- [ ] **Step 4.4: Commit**

```bash
cd backend && git add internal/service/report_service.go
git commit -m "feat(prompt): add weak-signal-year safe phrasing guide to dayun summary prompt"
```

---

## Task 5: 升级算法版本号到 `v3.1-narrative-guarded`

**Files:**
- Modify: `backend/internal/repository/algorithm_version.go`

- [ ] **Step 5.1: 修改版本常量和注释**

打开 `backend/internal/repository/algorithm_version.go`，将文件全文替换为：

```go
package repository

// CurrentAlgorithmVersion 当前生成 AI 报告 / 大运总结 所用的算法版本。
//
// 写入新行时使用此常量；老行 algorithm_version IS NULL 被视为 v1 baseline。
// 版本变化时（如 Phase 2 落地）需同步更新此常量并新增 migration。
//
// 版本历史：
//   v1            (NULL) — pre-yongshen-realignment baseline
//   v2-yongshen-shishen   — 喜忌十神 prompt 注入 + algorithm_version 列建立
//   v3-progressive-compressed — lazy-load dayun_indexes 过滤 + YearsData prompt 压缩
//   v3.1-narrative-guarded — AI 空 narrative / validator 清空走 template 兜底，
//                            prompt 追加弱信号年安全措辞指引
const CurrentAlgorithmVersion = "v3.1-narrative-guarded"
```

- [ ] **Step 5.2: 跑全包测试确认无回归**

```bash
cd backend && go test ./... -count=1
```

期望：所有测试 PASS。`algorithm_version` 只被写入路径使用，不被读取断言，所以不会有测试断言失败。

- [ ] **Step 5.3: Commit**

```bash
cd backend && git add internal/repository/algorithm_version.go
git commit -m "chore(version): bump algorithm version to v3.1-narrative-guarded"
```

---

## Task 6: 上线验证

**Files:** 无代码改动；运行手动验证。

- [ ] **Step 6.1: 启动服务**

```bash
cd /Users/liujiming/web/yuanju && docker-compose restart backend
docker logs -f yuanju_backend &
```

确认无 panic、迁移正常、监听 9002。

- [ ] **Step 6.2: 重新生成一个已知有空白年的 chart**

在前端用 1996-02-08 男的盘（chart_id=3c048af0-363a-4b8d-adb3-eff2f80f1dd1）进入"过往事件推算"页面，点击"重新生成"（或删除该 chart 的现有 ai_dayun_summaries 行后再生成）：

```sql
-- 选项：先清除已有数据触发重生
DELETE FROM ai_dayun_summaries WHERE chart_id='3c048af0-363a-4b8d-adb3-eff2f80f1dd1';
```

然后前端触发流式生成。

- [ ] **Step 6.3: 验证 DB 中 0 条 narrative=""**

```bash
docker exec yuanju_postgres psql -U yuanju -d yuanju -c "
WITH expanded AS (
  SELECT algorithm_version, dayun_index, jsonb_array_elements(years) AS y
  FROM ai_dayun_summaries
  WHERE chart_id='3c048af0-363a-4b8d-adb3-eff2f80f1dd1'
)
SELECT algorithm_version,
       COUNT(*) AS total,
       COUNT(*) FILTER (WHERE (y->>'narrative') = '') AS empty_narr
FROM expanded
GROUP BY algorithm_version;
"
```

期望：`v3.1-narrative-guarded` 行 `empty_narr = 0`。

- [ ] **Step 6.4: 抽查"原空白年"的兜底文案**

```bash
docker exec yuanju_postgres psql -U yuanju -d yuanju -c "
WITH expanded AS (
  SELECT dayun_index, jsonb_array_elements(years) AS y
  FROM ai_dayun_summaries
  WHERE chart_id='3c048af0-363a-4b8d-adb3-eff2f80f1dd1'
    AND algorithm_version='v3.1-narrative-guarded'
)
SELECT dayun_index, y->>'year' AS yr, y->>'ganzhi' AS gz, y->>'narrative' AS narr
FROM expanded
WHERE (y->>'year')::int IN (1999, 2008, 2017, 2019, 2028);
"
```

肉眼检查：原空白年要么是 AI 真填了、要么含"信号较稀"/"基调向吉"/"基调偏凶" 等 fallback 短语。绝不应再有空 narrative。

- [ ] **Step 6.5: 前端肉眼检查**

打开 `http://localhost:3000` → 登录 → 进入 1996-02-08 男的盘 → 过往事件推算页面。无痕窗口（Cmd+Shift+N）打开避免缓存。

期望：所有年份卡片都有正文，没有任何"光秃秃干支头"或仅显示 chip 的情况。

---

## Self-Review

**Spec 覆盖：**
- ✅ `RenderYearNarrativeWithFallback` 新增 → Task 1
- ✅ `makeMinimalFallback` 三个 polarity 分支 → Task 1（实现）+ Task 2（测试）
- ✅ 兜底句不触发 28 个 validatedKeywords → Task 2 Step 2.2
- ✅ `GenerateDayunSummariesStream` 调用点改写 → Task 3
- ✅ Prompt 加固"弱信号年安全措辞" → Task 4
- ✅ `CurrentAlgorithmVersion` 升级到 `v3.1-narrative-guarded` → Task 5
- ✅ TestRenderYearNarrativeWithFallback_AlwaysNonEmpty → Task 1 Step 1.1 (`NoSignalsReturnsNonEmpty`)
- ✅ TestMakeMinimalFallback_PolarityBranches → Task 2 Step 2.1
- ✅ TestMakeMinimalFallback_KeywordSafe → Task 2 Step 2.2
- ✅ TestGenerateDayunSummariesStream_FillsBlankNarrative → Task 3 Step 3.1（实现为 `TestFillBlankYearNarratives_*` 三个，因为整体 stream 函数难以纯单测；通过抽出 helper 实现等价覆盖）
- ✅ 存量数据不动 → 无相关任务即满足
- ✅ 无前端改动 → 计划无前端任务

**Placeholder scan：** 无 TBD/TODO/handle edge cases，每步含完整代码或完整命令。

**Type consistency：**
- `parsedYearAI` 在 Task 3 Step 3.1（测试）和 3.3（实现）均使用，名称一致。
- `yearOut` 在 Task 3 Step 3.3 顶层声明，原内联定义在 Step 3.4 删除，无冲突。
- `RenderYearNarrativeWithFallback` 和 `makeMinimalFallback` 在 Task 1 定义、Task 3 实现使用、Task 1+2 测试，签名一致。
- 字段 `Year`/`GanZhi`/`Narrative` 在 `parsedYearAI`、`yearOut`、AI prompt JSON 中均一致（注意 prompt 用 `ganzhi` 小写，Go struct tag `json:"ganzhi"` 已匹配）。
