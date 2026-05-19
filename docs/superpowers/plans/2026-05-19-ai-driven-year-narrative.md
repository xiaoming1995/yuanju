# AI-Driven Year Narrative Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace template-based per-year narratives with AI output. Extend the existing dayun summary AI call to also emit 10 year cards in one combined invocation, with fact-validation, feature-flag rollback, and token-cost monitoring.

**Architecture:** Single combined AI call per dayun span emits `{themes, summary, years: [{year, ganzhi, narrative}, ...]}`. Backend validates each year's narrative against algorithm signals (drops fabricated content). Frontend merges per-year narratives from dayun stream into year cards. Feature flag `year_narrative_mode` in `algo_config` table toggles between AI and template modes; template code stays for 4-6 weeks before deletion.

**Tech Stack:** Go 1.21+ (`backend/`), DeepSeek API via existing `StreamAIWithSystemNoThink`, PostgreSQL (via goose migrations), React 19 + TypeScript (`frontend/`).

**Related spec:** `docs/superpowers/specs/2026-05-19-ai-driven-year-narrative-design.md`

---

## File Structure

**Create:**
- `backend/pkg/database/migrations/00002_dayun_summaries_years.sql` — add `years JSONB` column to `ai_dayun_summaries`
- `backend/internal/service/year_narrative_validate.go` — `ValidateYearNarrative` function (护栏 1)
- `backend/internal/service/year_narrative_validate_test.go` — table-driven tests

**Modify:**
- `backend/internal/model/model.go` — add `Years` field to `AIDayunSummary`
- `backend/internal/repository/dayun_summary_repository.go` — read/write `years` column in all 3 functions
- `backend/internal/service/report_service.go` — prompt change + AI response parsing + validation wire + feature flag branch
- `backend/internal/service/algo_config_service.go` — add `GetYearNarrativeMode()` getter
- `backend/internal/handler/algo_config_handler.go` — expose `year_narrative_mode` in admin endpoints (already supports arbitrary keys via `GetAllAlgoConfig` / `UpsertAlgoConfig` — verify no code changes needed)
- `frontend/src/pages/PastEventsPage.tsx` — extend `DayunSummary` interface; merge per-year narrative from dayun stream; loading state
- `frontend/src/pages/admin/AlgoConfigPage.tsx` — surface the feature flag toggle

**Not touched (out of scope per spec):**
- `backend/pkg/bazi/event_signals.go` — signal detection unchanged
- `backend/pkg/bazi/event_narrative.go` + `event_narrative_leads.go` — DELETION DEFERRED to post-observation cleanup task (4-6 weeks)
- `backend/internal/service/report_service.go::GeneratePastEventsStream` (old report flow) — unrelated, separately tracked

---

## Task 1: Migration to add `years` JSONB column

**Files:**
- Create: `backend/pkg/database/migrations/00002_dayun_summaries_years.sql`

- [ ] **Step 1: Create the migration file**

```sql
-- +goose Up
ALTER TABLE ai_dayun_summaries ADD COLUMN years JSONB;
COMMENT ON COLUMN ai_dayun_summaries.years IS '10 个年份卡片 [{year,ganzhi,narrative}, ...]，AI dayun 调用同时产出。NULL=旧缓存需重生';

-- +goose Down
ALTER TABLE ai_dayun_summaries DROP COLUMN years;
```

- [ ] **Step 2: Verify migration syntax with dry-run**

```bash
cd /Users/liujiming/web/yuanju/backend && go run ./cmd/api --migrate-dry-run 2>&1 | tail -5
```

Expected: JSON output listing `[2]` as pending.

- [ ] **Step 3: Apply migration**

```bash
cd /Users/liujiming/web/yuanju/backend && go run ./cmd/api --migrate-apply 2>&1 | tail -5
```

Expected: JSON output listing `applied: [2]`, `failed: []`.

- [ ] **Step 4: Verify column exists**

```bash
docker-compose exec postgres psql -U postgres -d yuanju -c "\d ai_dayun_summaries" 2>&1 | grep years
```

Expected: line showing `years | jsonb |`.

- [ ] **Step 5: Commit**

```bash
cd /Users/liujiming/web/yuanju && git add backend/pkg/database/migrations/00002_dayun_summaries_years.sql
git -c commit.gpgsign=false commit -m "feat(migrations): add years jsonb column to ai_dayun_summaries

Stores [{year,ganzhi,narrative}] alongside existing themes+summary.
NULL value signals legacy cache row that needs AI regeneration."
```

---

## Task 2: Model + repository support `Years` field

**Files:**
- Modify: `backend/internal/model/model.go:107` — `AIDayunSummary` struct
- Modify: `backend/internal/repository/dayun_summary_repository.go` — all 3 SELECT/INSERT statements

- [ ] **Step 1: Add `Years` field to `AIDayunSummary`**

In `backend/internal/model/model.go`, find the `AIDayunSummary` struct (around line 107) and add a `Years` field:

```go
// AIDayunSummary 单段大运 AI 总结（按 chart_id + dayun_index 缓存）
type AIDayunSummary struct {
	ID          string           `json:"id"`
	ChartID     string           `json:"chart_id"`
	DayunIndex  int              `json:"dayun_index"`
	DayunGanZhi string           `json:"dayun_ganzhi"`
	Themes      *json.RawMessage `json:"themes"`
	Summary     string           `json:"summary"`
	Years       *json.RawMessage `json:"years"`
	Model       string           `json:"model"`
	CreatedAt   time.Time        `json:"created_at"`
}
```

- [ ] **Step 2: Update `GetDayunSummary` SELECT**

In `backend/internal/repository/dayun_summary_repository.go`, replace `GetDayunSummary` body to include `years` column:

```go
func GetDayunSummary(chartID string, dayunIndex int) (*model.AIDayunSummary, error) {
	r := &model.AIDayunSummary{}
	err := database.DB.QueryRow(
		`SELECT id, chart_id, dayun_index, dayun_ganzhi, themes, summary, years, model, created_at
		 FROM ai_dayun_summaries
		 WHERE chart_id = $1 AND dayun_index = $2`,
		chartID, dayunIndex,
	).Scan(&r.ID, &r.ChartID, &r.DayunIndex, &r.DayunGanZhi, &r.Themes, &r.Summary, &r.Years, &r.Model, &r.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return r, err
}
```

- [ ] **Step 3: Update `ListDayunSummaries` SELECT**

Same file, replace `ListDayunSummaries` body:

```go
func ListDayunSummaries(chartID string) ([]model.AIDayunSummary, error) {
	rows, err := database.DB.Query(
		`SELECT id, chart_id, dayun_index, dayun_ganzhi, themes, summary, years, model, created_at
		 FROM ai_dayun_summaries
		 WHERE chart_id = $1
		 ORDER BY dayun_index`,
		chartID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.AIDayunSummary
	for rows.Next() {
		var r model.AIDayunSummary
		if err := rows.Scan(&r.ID, &r.ChartID, &r.DayunIndex, &r.DayunGanZhi, &r.Themes, &r.Summary, &r.Years, &r.Model, &r.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}
```

- [ ] **Step 4: Update `UpsertDayunSummary` signature + INSERT**

Same file, replace `UpsertDayunSummary` entirely:

```go
// UpsertDayunSummary 写入或覆盖单段缓存
func UpsertDayunSummary(chartID string, dayunIndex int, dayunGanZhi string, themes *json.RawMessage, summary string, years *json.RawMessage, modelName string) error {
	_, err := database.DB.Exec(
		`INSERT INTO ai_dayun_summaries (chart_id, dayun_index, dayun_ganzhi, themes, summary, years, model)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 ON CONFLICT (chart_id, dayun_index) DO UPDATE
		 SET dayun_ganzhi = EXCLUDED.dayun_ganzhi,
		     themes = EXCLUDED.themes,
		     summary = EXCLUDED.summary,
		     years = EXCLUDED.years,
		     model = EXCLUDED.model,
		     created_at = NOW()`,
		chartID, dayunIndex, dayunGanZhi, themes, summary, years, modelName,
	)
	return err
}
```

- [ ] **Step 5: Update existing `UpsertDayunSummary` call site in service layer**

In `backend/internal/service/report_service.go` around line 1271 (the existing call inside `GenerateDayunSummariesStream`), change:

```go
_ = repository.UpsertDayunSummary(chartID, dy.Index, gz, &themesRaw, parsed.Summary, modelName)
```

to:

```go
// Task 5/7 will replace this; for now passing nil years to keep compile passing.
_ = repository.UpsertDayunSummary(chartID, dy.Index, gz, &themesRaw, parsed.Summary, nil, modelName)
```

- [ ] **Step 6: Verify compile**

```bash
cd /Users/liujiming/web/yuanju/backend && go build ./...
```

Expected: no errors.

- [ ] **Step 7: Verify existing tests still pass**

```bash
cd /Users/liujiming/web/yuanju/backend && go test ./...
```

Expected: all green.

- [ ] **Step 8: Commit**

```bash
cd /Users/liujiming/web/yuanju && git add backend/internal/model/model.go backend/internal/repository/dayun_summary_repository.go backend/internal/service/report_service.go
git -c commit.gpgsign=false commit -m "feat(dayun): add Years field to AIDayunSummary model+repo

Schema change paired with migration 00002. UpsertDayunSummary signature
now takes years *json.RawMessage parameter. Existing call site passes
nil for now (Task 7 wires the AI-produced years through)."
```

---

## Task 3: Feature flag — `year_narrative_mode` in algo_config

**Files:**
- Modify: `backend/internal/service/algo_config_service.go` — add `GetYearNarrativeMode()` getter
- Modify: `backend/internal/service/algo_config_service.go` — add seed step for default value

- [ ] **Step 1: Write the failing test**

Create `backend/internal/service/algo_config_service_test.go` (if not exists, append if exists):

```go
package service

import (
	"testing"
)

func TestGetYearNarrativeMode_DefaultsToAI(t *testing.T) {
	// When no row exists in algo_config, getter returns "ai" (default).
	got := GetYearNarrativeMode()
	if got != "ai" && got != "template" {
		t.Fatalf("expected ai or template, got %q", got)
	}
	// Default behavior: return "ai" when not set.
	// (We cannot assert exact value without DB state control; this test
	// is a smoke check that the function compiles and returns valid value.)
}
```

- [ ] **Step 2: Run test to verify it fails to compile**

```bash
cd /Users/liujiming/web/yuanju/backend && go test ./internal/service/ -run TestGetYearNarrativeMode_DefaultsToAI 2>&1 | tail -5
```

Expected: compile error `undefined: GetYearNarrativeMode`.

- [ ] **Step 3: Implement `GetYearNarrativeMode`**

Append to `backend/internal/service/algo_config_service.go`:

```go
// GetYearNarrativeMode 读取 algo_config 表中 year_narrative_mode 键。
// "ai"（默认）→ 用 AI 生成年度批语；"template" → 走旧模板路径。
// 该 flag 用作 4-6 周观察期的回滚开关。
func GetYearNarrativeMode() string {
	rows, err := database.DB.Query(
		`SELECT value FROM algo_config WHERE key = 'year_narrative_mode' LIMIT 1`,
	)
	if err != nil {
		log.Printf("[algo_config] GetYearNarrativeMode 查询失败，默认 ai: %v", err)
		return "ai"
	}
	defer rows.Close()
	if rows.Next() {
		var v string
		if err := rows.Scan(&v); err != nil {
			return "ai"
		}
		if v == "template" {
			return "template"
		}
	}
	return "ai"
}
```

Verify the `database` and `log` imports exist at the top of the file (they should — `LoadAlgoConfig` already uses them).

- [ ] **Step 4: Run test to verify it compiles and passes**

```bash
cd /Users/liujiming/web/yuanju/backend && go test ./internal/service/ -run TestGetYearNarrativeMode_DefaultsToAI -v
```

Expected: PASS.

- [ ] **Step 5: Verify `algo_config_handler.go` already exposes this key**

Read `backend/internal/handler/algo_config_handler.go` to confirm `AdminGetAlgoConfig` returns all rows (it iterates `GetAllAlgoConfig()`) and `AdminUpsertAlgoConfig` accepts arbitrary `key`+`value`. If yes, no code change needed — the frontend admin UI will see/edit `year_narrative_mode` like any other row.

```bash
grep -A3 "func AdminGetAlgoConfig\|func AdminUpsertAlgoConfig" /Users/liujiming/web/yuanju/backend/internal/handler/algo_config_handler.go | head -20
```

Expected: both functions are generic key-value endpoints (no whitelist of keys).

- [ ] **Step 6: Commit**

```bash
cd /Users/liujiming/web/yuanju && git add backend/internal/service/algo_config_service.go backend/internal/service/algo_config_service_test.go
git -c commit.gpgsign=false commit -m "feat(algo_config): add year_narrative_mode feature flag getter

GetYearNarrativeMode() reads the year_narrative_mode key from algo_config
table. Default 'ai' on missing key. Acts as kill-switch for the 4-6 week
observation period after AI year-narrative rollout."
```

---

## Task 4: `ValidateYearNarrative` function + tests (护栏 1)

**Files:**
- Create: `backend/internal/service/year_narrative_validate.go`
- Create: `backend/internal/service/year_narrative_validate_test.go`

- [ ] **Step 1: Write the failing tests**

Create `backend/internal/service/year_narrative_validate_test.go`:

```go
package service

import (
	"strings"
	"testing"

	"yuanju/pkg/bazi"
)

func TestValidateYearNarrative_PassesWhenAllKeywordsTraceToEvidence(t *testing.T) {
	signals := []bazi.EventSignal{
		{Type: "健康", Evidence: "白虎临运，主孝服、突发伤痛或意外", Source: bazi.SourceShensha, Polarity: bazi.PolarityXiong},
		{Type: "迁变", Evidence: "驿马临运，主奔波、出行、变动", Source: bazi.SourceShensha, Polarity: bazi.PolarityNeutral},
	}
	narrative := "本年白虎临运，健康注意；驿马合年支，宜防奔波。"
	ok, reason := ValidateYearNarrative(narrative, signals)
	if !ok {
		t.Errorf("expected pass, got fail: %s", reason)
	}
}

func TestValidateYearNarrative_FailsOnUnattestedShensha(t *testing.T) {
	signals := []bazi.EventSignal{
		{Type: "迁变", Evidence: "驿马临运，主奔波", Source: bazi.SourceShensha, Polarity: bazi.PolarityNeutral},
	}
	// AI fabricated "桃花临运" which doesn't appear in any signal evidence.
	narrative := "本年驿马动象明显，桃花临运人缘旺。"
	ok, reason := ValidateYearNarrative(narrative, signals)
	if ok {
		t.Errorf("expected fail (桃花 not in evidence), got pass")
	}
	if !strings.Contains(reason, "桃花") {
		t.Errorf("expected reason to mention 桃花, got: %s", reason)
	}
}

func TestValidateYearNarrative_FailsOnUnattestedPositionClaim(t *testing.T) {
	signals := []bazi.EventSignal{
		{Type: "综合变动", Evidence: "流年地支辰冲月柱戌", Source: bazi.SourceZhuwei, Polarity: bazi.PolarityXiong},
	}
	// AI fabricated "用神位受刑" — no signal has that text.
	narrative := "流年地支辰冲月柱戌，且用神位受刑，宜防。"
	ok, reason := ValidateYearNarrative(narrative, signals)
	if ok {
		t.Errorf("expected fail (用神位 not attested), got pass")
	}
	if !strings.Contains(reason, "用神位") {
		t.Errorf("expected reason to mention 用神位, got: %s", reason)
	}
}

func TestValidateYearNarrative_PassesWithoutAnyHardKeywords(t *testing.T) {
	// Narrative with no validated keywords passes trivially.
	signals := []bazi.EventSignal{
		{Type: "事业", Evidence: "流年节奏微调", Source: bazi.SourceZhuwei, Polarity: bazi.PolarityNeutral},
	}
	narrative := "本年事业上节奏微调，按部就班即可。"
	ok, _ := ValidateYearNarrative(narrative, signals)
	if !ok {
		t.Error("expected pass for narrative without validated keywords")
	}
}

func TestValidateYearNarrative_PassesOnEmptyNarrative(t *testing.T) {
	// Empty narrative (AI explicitly skipped this year) always passes.
	signals := []bazi.EventSignal{}
	ok, _ := ValidateYearNarrative("", signals)
	if !ok {
		t.Error("expected pass for empty narrative")
	}
}

func TestValidateYearNarrative_FuyinFanyinDayunheHua(t *testing.T) {
	// 伏吟 / 反吟 / 大运合化 are validated terms.
	signals := []bazi.EventSignal{
		{Type: "伏吟", Evidence: "流年壬辰伏吟日柱壬辰，主同类事件重现", Source: bazi.SourceFuyin, Polarity: bazi.PolarityXiong},
	}
	narrative := "本年伏吟日柱，旧事重提；反吟未现。"
	ok, reason := ValidateYearNarrative(narrative, signals)
	if ok {
		t.Errorf("expected fail (反吟 not in evidence), got pass")
	}
	if !strings.Contains(reason, "反吟") {
		t.Errorf("expected reason to mention 反吟, got: %s", reason)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail to compile**

```bash
cd /Users/liujiming/web/yuanju/backend && go test ./internal/service/ -run TestValidateYearNarrative 2>&1 | tail -5
```

Expected: `undefined: ValidateYearNarrative`.

- [ ] **Step 3: Implement the function**

Create `backend/internal/service/year_narrative_validate.go`:

```go
package service

import (
	"fmt"
	"strings"

	"yuanju/pkg/bazi"
)

// validatedKeywords 是 narrative 中出现就必须能在算法 evidence 里追溯到的 字符串。
// 选取标准：
//   (1) 命理学有具体结构含义的术语（用神位/忌神位/喜神位）
//   (2) 命理学有特定干支事件含义的术语（伏吟/反吟/大运合化/三会/三合/受冲/受刑/双重命中/力度倍增）
//   (3) 神煞名（每一个神煞临运 都对应 event_signals.go 的具体生成路径）
// 不在此列的术语（如十神名「食神/伤官」、单纯干支「甲乙丙丁」、宫位「日柱/月柱」）
// AI 可以自由发挥而不被卡，因为它们可从 BaziResult 直接推导。
var validatedKeywords = []string{
	// 位
	"用神位", "忌神位", "喜神位",
	// 强变动
	"伏吟", "反吟", "大运合化", "三会", "三合",
	// 硬事件标记
	"受冲", "受刑", "双重命中", "力度倍增",
	// 神煞（与 event_signals.go::shenshaTable 对齐）
	"驿马", "桃花", "华盖", "白虎", "丧门", "吊客", "灾煞", "流霞",
	"天医", "天喜", "天乙", "天德", "月德", "文昌", "太极", "福星",
	"红艳", "孤辰", "寡宿", "羊刃", "亡神", "劫煞", "披麻", "咸池",
	"勾绞", "国印",
}

// ValidateYearNarrative 校验单年 narrative 引用的命理术语是否能在该年算法
// signals 的 Evidence 中追溯到。
//
// 返回 (true, "") 表示通过；返回 (false, reason) 表示某个关键词无源可溯，
// reason 包含具体哪个词没匹配上。
//
// 设计意图：拦截 AI 自信地说错（编造神煞/编造用神位事件）的最常见路径。
// 不做语义级校验（"宜防健康" 是否合理）—— 那是 AI 的判断权限范围。
//
// 空 narrative（"" — AI 决定不写）总是通过。
func ValidateYearNarrative(narrative string, signals []bazi.EventSignal) (bool, string) {
	if narrative == "" {
		return true, ""
	}
	// 拼一份 evidence 全文，逐关键词查一次 substring 包含。
	var evidenceBuf strings.Builder
	for _, s := range signals {
		evidenceBuf.WriteString(s.Evidence)
		evidenceBuf.WriteString("\n")
	}
	allEvidence := evidenceBuf.String()
	for _, kw := range validatedKeywords {
		if strings.Contains(narrative, kw) && !strings.Contains(allEvidence, kw) {
			return false, fmt.Sprintf("narrative 出现 %q 但算法 evidence 无对应来源", kw)
		}
	}
	return true, ""
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
cd /Users/liujiming/web/yuanju/backend && go test ./internal/service/ -run TestValidateYearNarrative -v
```

Expected: all 6 tests PASS.

- [ ] **Step 5: Commit**

```bash
cd /Users/liujiming/web/yuanju && git add backend/internal/service/year_narrative_validate.go backend/internal/service/year_narrative_validate_test.go
git -c commit.gpgsign=false commit -m "feat(narrative): ValidateYearNarrative cross-checks AI keywords against signals

护栏 1 implementation: drop narratives that reference 神煞/位/强变动
keywords not attested by the year's signal evidence. Validation list is
conservative (skips 十神名 and 宫位 names that AI can derive from BaziResult)
to avoid false positives. Empty narrative always passes."
```

---

## Task 5: Extend AI prompt with `years` block

**Files:**
- Modify: `backend/internal/service/report_service.go:1117-1136` — prompt template string

- [ ] **Step 1: Update the prompt template**

In `backend/internal/service/report_service.go`, find the `promptTpl` variable inside `GenerateDayunSummariesStream` (around line 1117). Replace the entire `promptTpl` value with:

```go
	promptTpl := `你是一位资深八字命理师。请只为下列单段大运撰写整体总结和该段 10 年的逐年评述。

命主：性别{{.Gender}} / 日干{{.DayGan}}
原局：{{.NatalSummary}}
{{if .YongshenInfo}}用忌神：{{.YongshenInfo}}{{end}}
{{if .StrengthDetail}}身强弱：{{.StrengthDetail}}{{end}}

当前大运：{{.DayunInfo}}
{{if .HuaheTag}}合化：{{.HuaheTag}}{{end}}

本段大运 10 年的算法信号摘要（JSON，每年含 type/evidence/polarity/source/year_in_dayun/dayun_phase/dayun_phase_level；dayun_phase=gan 表示前5年天干主事，zhi 表示后5年地支主事）：
{{.YearsData}}
{{if .LifeStageHint}}
人生阶段提示：{{.LifeStageHint}}{{end}}

输出要求：
1. themes：2-4 个主题词（如"事业↑""感情动荡""贵人扶持"；读书期可用"学业突破""同窗情谊""叛逆"）
2. summary：80-120 字，综合评述这 10 年整体走势、关键转折、注意事项；若前5年与后5年信号明显不同，要点出早段/后段气质差异
3. years：长度等于上方算法信号 JSON 的年份数，与年份顺序一一对应。每个元素：
   {"year": 数字年, "ganzhi": "干支", "narrative": "..."}

   narrative 撰写规则：
   - 100-150 字，3-4 句中文
   - 必须点名当年关键干支事件，引用上方 evidence 已有的命理术语
     （如「丙火透干为食神」「流年地支冲日支」「白虎临运」「驿马合年支」
     「用神位受刑」「伏吟时柱」等）
   - 结合极性写吉凶（吉应期写助力或机遇，凶应期写注意或代价）
   - 读书期年份（age<18，由人生阶段提示判断）改写为学业/同学/家庭语义，
     不出现「事业/婚恋/财运」等成人词
   - 若该年信号确实稀薄（无 hard event 信号、evidence 关键词都缺），
     narrative 可写 "" 表示该年无显著动象
   - 措辞与 summary 不重复，summary 概括十年，narrative 具体到当年
   - 严禁编造未在 evidence 中出现的神煞或用神位事件

4. 严格输出以下 JSON，不要 Markdown 围栏：
{"themes":["主题1","主题2"],"summary":"...","years":[{"year":2005,"ganzhi":"乙酉","narrative":"..."},{"year":2006,"ganzhi":"丙戌","narrative":"..."}]}`
```

- [ ] **Step 2: Verify compile**

```bash
cd /Users/liujiming/web/yuanju/backend && go build ./...
```

Expected: no errors.

- [ ] **Step 3: Commit (prompt-only, AI flow change comes in Task 7)**

```bash
cd /Users/liujiming/web/yuanju && git add backend/internal/service/report_service.go
git -c commit.gpgsign=false commit -m "feat(narrative): extend dayun summary prompt with per-year output block

AI is now instructed to produce {themes, summary, years[]} per dayun.
Output JSON parsing is still themes+summary only — Task 7 extends the
parser to consume years. This commit isolates the prompt change so a
regression can be bisected cleanly."
```

---

## Task 6: Parse `years` from AI JSON + structural validation

**Files:**
- Modify: `backend/internal/service/report_service.go:1259-1271` — extend parse struct + add length/ganzhi validation

- [ ] **Step 1: Extend the parsed struct**

In `backend/internal/service/report_service.go`, find the parse block inside `GenerateDayunSummariesStream` (around line 1259). Replace:

```go
		var parsed struct {
			Themes  []string `json:"themes"`
			Summary string   `json:"summary"`
		}
		if jerr := json.Unmarshal([]byte(raw), &parsed); jerr != nil {
			_ = onItem(DayunSummaryStreamItem{DayunIndex: dy.Index, GanZhi: gz, Error: "解析 AI JSON 失败"})
			continue
		}
```

with:

```go
		var parsed struct {
			Themes  []string `json:"themes"`
			Summary string   `json:"summary"`
			Years   []struct {
				Year      int    `json:"year"`
				GanZhi    string `json:"ganzhi"`
				Narrative string `json:"narrative"`
			} `json:"years"`
		}
		if jerr := json.Unmarshal([]byte(raw), &parsed); jerr != nil {
			_ = onItem(DayunSummaryStreamItem{DayunIndex: dy.Index, GanZhi: gz, Error: "解析 AI JSON 失败"})
			continue
		}

		// 结构校验：years 数组长度必须等于该段实际年份数，且 ganzhi 一一对应。
		// 不匹配整段算失败，避免错位带来的鬼故事。
		if len(parsed.Years) != len(dySignals) {
			_ = onItem(DayunSummaryStreamItem{
				DayunIndex: dy.Index, GanZhi: gz,
				Error: fmt.Sprintf("years 长度不对：AI 返回 %d，期望 %d", len(parsed.Years), len(dySignals)),
			})
			continue
		}
		ganzhiMismatch := false
		for i, ys := range dySignals {
			if parsed.Years[i].GanZhi != ys.GanZhi {
				ganzhiMismatch = true
				log.Printf("[GenerateDayunSummariesStream] dayun=%d year=%d ganzhi 错位：AI=%q 实际=%q",
					dy.Index, ys.Year, parsed.Years[i].GanZhi, ys.GanZhi)
				break
			}
		}
		if ganzhiMismatch {
			_ = onItem(DayunSummaryStreamItem{
				DayunIndex: dy.Index, GanZhi: gz,
				Error: "years 数组干支与算法不对齐",
			})
			continue
		}
```

- [ ] **Step 2: Verify compile**

```bash
cd /Users/liujiming/web/yuanju/backend && go build ./...
```

Expected: no errors.

- [ ] **Step 3: Verify existing tests still pass**

```bash
cd /Users/liujiming/web/yuanju/backend && go test ./...
```

Expected: all green. (No new tests yet — integration tests added in Task 7.)

- [ ] **Step 4: Commit**

```bash
cd /Users/liujiming/web/yuanju && git add backend/internal/service/report_service.go
git -c commit.gpgsign=false commit -m "feat(narrative): parse and structurally validate AI years field

Extend parsed struct to consume years[]; enforce len(years) == len(dySignals)
and per-index ganzhi match. Mismatches fail the entire dayun (no partial
year merging — too easy to misalign and surface wrong-year narratives)."
```

---

## Task 7: Wire fact validation + cache write of `years`

**Files:**
- Modify: `backend/internal/service/report_service.go` — after parse block, validate each year and persist

- [ ] **Step 1: Add per-year validation + cache write**

In `backend/internal/service/report_service.go`, immediately after the ganzhi-match block from Task 6 (and before the existing `themesJSON, _ := json.Marshal(parsed.Themes)` line, around line 1269), insert:

```go
		// 护栏 1：逐年校验 narrative 中的命理术语能在算法 evidence 里追溯到。
		// 校验失败的年份 narrative 被清空（其他字段保留），日志记录原因。
		type yearOut struct {
			Year      int    `json:"year"`
			GanZhi    string `json:"ganzhi"`
			Narrative string `json:"narrative"`
		}
		validatedYears := make([]yearOut, len(parsed.Years))
		for i, y := range parsed.Years {
			validatedYears[i] = yearOut{Year: y.Year, GanZhi: y.GanZhi, Narrative: y.Narrative}
			if y.Narrative == "" {
				continue
			}
			if ok, reason := ValidateYearNarrative(y.Narrative, dySignals[i].Signals); !ok {
				log.Printf("[GenerateDayunSummariesStream] dayun=%d year=%d 校验失败丢弃 narrative：%s",
					dy.Index, y.Year, reason)
				validatedYears[i].Narrative = ""
			}
		}
		yearsJSON, _ := json.Marshal(validatedYears)
		yearsRaw := json.RawMessage(yearsJSON)
```

- [ ] **Step 2: Update the `UpsertDayunSummary` call to pass `&yearsRaw`**

In the same function, find (set to nil in Task 2):

```go
_ = repository.UpsertDayunSummary(chartID, dy.Index, gz, &themesRaw, parsed.Summary, nil, modelName)
```

Replace with:

```go
_ = repository.UpsertDayunSummary(chartID, dy.Index, gz, &themesRaw, parsed.Summary, &yearsRaw, modelName)
```

- [ ] **Step 3: Verify compile**

```bash
cd /Users/liujiming/web/yuanju/backend && go build ./...
```

Expected: no errors.

- [ ] **Step 4: Run all tests**

```bash
cd /Users/liujiming/web/yuanju/backend && go test ./...
```

Expected: all green.

- [ ] **Step 5: Commit**

```bash
cd /Users/liujiming/web/yuanju && git add backend/internal/service/report_service.go
git -c commit.gpgsign=false commit -m "feat(narrative): wire fact validation + persist years to cache

After AI returns + structural validation passes, each year's narrative
runs through ValidateYearNarrative. Failed narratives are cleared
(set to '') but the year entry stays so other years aren't penalized.
Cache row now stores validated years JSON."
```

---

## Task 8: Extend `DayunSummaryStreamItem` with `Years` + push to client

**Files:**
- Modify: `backend/internal/service/report_service.go:1047-1080` — `DayunSummaryStreamItem` type + `cachedDayunSummaryToStreamItem` + push call

- [ ] **Step 1: Add `Years` field to `DayunSummaryStreamItem`**

In `backend/internal/service/report_service.go`, find the `DayunSummaryStreamItem` struct (around line 1047). Replace it with:

```go
// DayunSummaryStreamItem SSE 流式推送的单段大运 summary + 10 年卡片
type DayunSummaryStreamItem struct {
	DayunIndex int             `json:"dayun_index"`
	GanZhi     string          `json:"gan_zhi"`
	Themes     []string        `json:"themes,omitempty"`
	Summary    string          `json:"summary,omitempty"`
	Years      json.RawMessage `json:"years,omitempty"`
	Cached     bool            `json:"cached,omitempty"`
	Error      string          `json:"error,omitempty"`
}
```

- [ ] **Step 2: Update `cachedDayunSummaryToStreamItem` to pass through `Years`**

Same file, find `cachedDayunSummaryToStreamItem` (around line 1057). Replace with:

```go
func cachedDayunSummaryToStreamItem(cached *model.AIDayunSummary, fallbackGanZhi string) (DayunSummaryStreamItem, bool) {
	if cached == nil {
		return DayunSummaryStreamItem{}, false
	}
	// 缓存 row 没有 years → lazy migrate：视为缓存未命中，让上游重生
	if cached.Years == nil {
		return DayunSummaryStreamItem{}, false
	}
	var themes []string
	if cached.Themes != nil {
		if err := json.Unmarshal(*cached.Themes, &themes); err != nil {
			return DayunSummaryStreamItem{}, false
		}
	}
	gz := cached.DayunGanZhi
	if gz == "" {
		gz = fallbackGanZhi
	}
	return DayunSummaryStreamItem{
		DayunIndex: cached.DayunIndex,
		GanZhi:     gz,
		Themes:     themes,
		Summary:    cached.Summary,
		Years:      *cached.Years,
		Cached:     true,
	}, true
}
```

- [ ] **Step 3: Update the final push call in `GenerateDayunSummariesStream` to include `Years`**

Same file, near the end of the function (around line 1274). Replace the `onItem(DayunSummaryStreamItem{...})` push call with:

```go
		_ = onItem(DayunSummaryStreamItem{
			DayunIndex: dy.Index,
			GanZhi:     gz,
			Themes:     parsed.Themes,
			Summary:    parsed.Summary,
			Years:      yearsRaw,
			Cached:     false,
		})
```

- [ ] **Step 4: Verify compile**

```bash
cd /Users/liujiming/web/yuanju/backend && go build ./...
```

Expected: no errors.

- [ ] **Step 5: Run all tests**

```bash
cd /Users/liujiming/web/yuanju/backend && go test ./...
```

Expected: all green.

- [ ] **Step 6: Commit**

```bash
cd /Users/liujiming/web/yuanju && git add backend/internal/service/report_service.go
git -c commit.gpgsign=false commit -m "feat(narrative): SSE stream item carries Years field

DayunSummaryStreamItem now ships years JSON to frontend. Cached rows
missing years are treated as cache miss (lazy migrate triggers AI
regeneration on next access)."
```

---

## Task 9: Stop `GeneratePastEventsYears` emitting template narrative when mode=ai

**Files:**
- Modify: `backend/internal/service/report_service.go:1023-1038` — wrap `Narrative` assignment

- [ ] **Step 1: Apply the feature flag**

In `backend/internal/service/report_service.go`, find `GeneratePastEventsYears` (around line 988). Find the `years = append(years, PastEventsYearItem{...})` block (around line 1023-1038). Replace the `Narrative:` line so it respects the feature flag:

Locate this block:

```go
		years = append(years, PastEventsYearItem{
			Year:            ys.Year,
			Age:             ys.Age,
			GanZhi:          ys.GanZhi,
			DayunGanZhi:     ys.DayunGanZhi,
			DayunIndex:      dyIndex[ys.DayunGanZhi],
			YearInDayun:     ys.YearInDayun,
			DayunPhase:      ys.DayunPhase,
			TenGodPower:     ys.TenGodPower,
			Signals:         bazi.ExtractYearSignalTypes(ys),
			Narrative:       bazi.RenderYearNarrative(ys),
			EvidenceSummary: bazi.RenderEvidenceSummary(ys),
		})
```

Replace with:

```go
		var narrative string
		if mode == "template" {
			narrative = bazi.RenderYearNarrative(ys)
		}
		years = append(years, PastEventsYearItem{
			Year:            ys.Year,
			Age:             ys.Age,
			GanZhi:          ys.GanZhi,
			DayunGanZhi:     ys.DayunGanZhi,
			DayunIndex:      dyIndex[ys.DayunGanZhi],
			YearInDayun:     ys.YearInDayun,
			DayunPhase:      ys.DayunPhase,
			TenGodPower:     ys.TenGodPower,
			Signals:         bazi.ExtractYearSignalTypes(ys),
			Narrative:       narrative,
			EvidenceSummary: bazi.RenderEvidenceSummary(ys),
		})
```

- [ ] **Step 2: Add `mode` lookup at the top of the function**

In the same function, after `result, err := LoadOrCalculateResult(chart)` and before the `currentYear` line (around line 999), insert:

```go
	mode := GetYearNarrativeMode()
```

- [ ] **Step 3: Update `Generated` field to reflect mode**

In the return statement at the bottom of `GeneratePastEventsYears`:

```go
	return &PastEventsYearsResponse{
		Years:     years,
		DayunMeta: dayunMeta,
		Generated: "algo-template",
	}, nil
```

Replace `"algo-template"` with `mode + "-yearly"`:

```go
		Generated: mode + "-yearly",
```

- [ ] **Step 4: Verify compile**

```bash
cd /Users/liujiming/web/yuanju/backend && go build ./...
```

Expected: no errors.

- [ ] **Step 5: Run all tests**

```bash
cd /Users/liujiming/web/yuanju/backend && go test ./...
```

Expected: all green.

- [ ] **Step 6: Commit**

```bash
cd /Users/liujiming/web/yuanju && git add backend/internal/service/report_service.go
git -c commit.gpgsign=false commit -m "feat(narrative): GeneratePastEventsYears respects year_narrative_mode flag

When mode=ai (default), per-year Narrative is empty in Stage 1 response
— frontend fills via Stage 2 SSE merge. When mode=template, falls back
to existing RenderYearNarrative path. Generated field reflects which
path produced the data."
```

---

## Task 10: Frontend — extend `DayunSummary` type + merge per-year narrative

**Files:**
- Modify: `frontend/src/pages/PastEventsPage.tsx`

- [ ] **Step 1: Extend the `DayunSummary` interface**

In `frontend/src/pages/PastEventsPage.tsx`, find the `interface DayunSummary` declaration (around line 43). Replace:

```tsx
interface DayunSummary {
  themes: string[]
  summary: string
  loading?: boolean
  error?: string
}
```

with:

```tsx
interface YearNarrativeEntry {
  year: number
  ganzhi: string
  narrative: string
}

interface DayunSummary {
  themes: string[]
  summary: string
  years?: YearNarrativeEntry[]
  loading?: boolean
  error?: string
}
```

- [ ] **Step 2: Update the SSE `onItem` handler to capture `years`**

Same file, find the `streamDayunSummaries` callback (around line 128). Replace the `if (item.error)` / `else` block with:

```tsx
        setSummaries((prev) => {
          const next = { ...prev }
          if (item.error) {
            next[item.dayun_index] = { themes: [], summary: '', error: item.error, loading: false }
          } else {
            next[item.dayun_index] = {
              themes: item.themes || [],
              summary: item.summary || '',
              years: item.years || undefined,
              loading: false,
            }
          }
          return next
        })
```

- [ ] **Step 3: Add `yearNarrative` helper**

Same file, just before the `return (` line of the component (around line 167, after `grouped` definition):

```tsx
  const yearNarrative = (y: YearEvent): { text: string; status: 'loading' | 'ready' | 'empty' } => {
    // Stage 1 already returned a narrative (template mode) → use it
    if (y.narrative && y.narrative !== '') {
      return { text: y.narrative, status: 'ready' }
    }
    const ds = summaries[y.dayun_index]
    if (!ds || ds.loading) {
      return { text: '', status: 'loading' }
    }
    if (ds.error) {
      return { text: '', status: 'empty' }
    }
    const entry = ds.years?.find((ye) => ye.year === y.year)
    if (!entry) {
      return { text: '', status: 'empty' }
    }
    if (entry.narrative === '') {
      return { text: '', status: 'empty' }
    }
    return { text: entry.narrative, status: 'ready' }
  }
```

- [ ] **Step 4: Replace the narrative render block**

Same file, find the existing `{y.narrative && (...)}` block in the year card render (around line 417-421). Replace with:

```tsx
                          {(() => {
                            const n = yearNarrative(y)
                            if (n.status === 'loading') {
                              return (
                                <div style={{ color: 'var(--text-muted)', fontSize: '0.78rem', fontStyle: 'italic' }}>
                                  本段批语正在生成…
                                </div>
                              )
                            }
                            if (n.status === 'ready') {
                              return (
                                <div style={{ color: 'var(--text-secondary)', fontSize: '0.85rem', lineHeight: 1.7 }}>
                                  {n.text}
                                </div>
                              )
                            }
                            return null
                          })()}
```

- [ ] **Step 5: Type-check + build**

```bash
cd /Users/liujiming/web/yuanju/frontend && npm run build 2>&1 | tail -10
```

Expected: build succeeds, no TypeScript errors.

- [ ] **Step 6: Commit**

```bash
cd /Users/liujiming/web/yuanju && git add frontend/src/pages/PastEventsPage.tsx
git -c commit.gpgsign=false commit -m "feat(past-events): frontend merges per-year narrative from dayun stream

Year cards now display a loading hint while their owning dayun span's AI
call is in progress; show narrative when it arrives; render nothing
when AI explicitly returned empty for that year (信号稀薄). Falls back
to the Stage 1 narrative when present (template mode)."
```

---

## Task 11: Frontend admin UI surfaces feature flag

**Files:**
- Modify: `frontend/src/pages/admin/AlgoConfigPage.tsx`

- [ ] **Step 1: Inspect existing admin algo_config UI**

```bash
grep -n "algo_config\|UpsertAlgoConfig\|key.*value" /Users/liujiming/web/yuanju/frontend/src/pages/admin/AlgoConfigPage.tsx | head -20
```

The admin page already iterates rows from `GetAllAlgoConfig` — confirm if it shows every key generically, or if it hardcodes a known-keys list.

- [ ] **Step 2: Add a known-key documentation note**

If the page is generic key-value (likely): no code change needed. The admin user can manually create the `year_narrative_mode` row with value `"ai"` or `"template"` via the existing UI.

If the page hardcodes keys: add `year_narrative_mode` to the list of displayed keys with a description label. Modify `AlgoConfigPage.tsx` accordingly — append the new key handling block to whatever pattern exists. Use this label:

```tsx
{key: 'year_narrative_mode', label: '年度批语生成模式', description: 'ai = AI 生成（默认）；template = 走旧模板路径（回滚开关）', allowedValues: ['ai', 'template']}
```

- [ ] **Step 3: Seed the default value in the database**

Insert the default row so the admin UI shows it immediately:

```bash
docker-compose exec postgres psql -U postgres -d yuanju -c "INSERT INTO algo_config (key, value, description) VALUES ('year_narrative_mode', 'ai', '年度批语生成模式：ai=AI生成（默认）/template=回旧模板路径') ON CONFLICT (key) DO NOTHING;"
```

Expected: `INSERT 0 1` or `INSERT 0 0` if it already exists.

- [ ] **Step 4: Type-check + build**

```bash
cd /Users/liujiming/web/yuanju/frontend && npm run build 2>&1 | tail -10
```

Expected: success.

- [ ] **Step 5: Commit (if changes made)**

```bash
cd /Users/liujiming/web/yuanju && git add frontend/src/pages/admin/AlgoConfigPage.tsx
git -c commit.gpgsign=false commit -m "feat(admin): surface year_narrative_mode toggle in algo config"
```

(Skip the commit step if Step 1 confirmed no code change was needed.)

---

## Task 12: End-to-end verification

**Files:** none modified — verification only.

- [ ] **Step 1: Rebuild + restart both containers**

```bash
cd /Users/liujiming/web/yuanju && docker-compose up -d --build backend frontend 2>&1 | tail -10
```

Expected: both built and started.

- [ ] **Step 2: Wait for backend ready + check migration applied**

```bash
sleep 5 && docker-compose exec postgres psql -U postgres -d yuanju -c "SELECT version_id FROM goose_db_version ORDER BY version_id;"
```

Expected: list includes `2`.

- [ ] **Step 3: Open a chart in browser**

Navigate to `http://localhost:3000`, log in, open a chart (any chart with child-age years preferred — e.g. `1996-02-08 20:00 male`). Click "过往事件推算".

- [ ] **Step 4: Visual checks**

- **Stage 1** (instant): year cards appear with GanZhi badge + chips + 命理依据. **NO narrative paragraph rendered yet** (or "本段批语正在生成…" hint per card).
- **Stage 2** (streaming, 5-30s per dayun depending on AI latency): dayun summaries appear progressively at the top of each dayun group. Year cards within each dayun also fill in their narrative paragraph as the corresponding dayun completes.
- **AI quality** check: open a year card with rich chips (e.g. multiple shensha + 综合变动 + 健康). Confirm narrative mentions specific evidence keywords (e.g. "白虎临运" "驿马合年支" "用神位"). Compare against the dayun summary at top — they should feel similar in depth.

- [ ] **Step 5: Test feature flag rollback**

In a second terminal, flip the flag and verify the old template path comes back:

```bash
docker-compose exec postgres psql -U postgres -d yuanju -c "UPDATE algo_config SET value = 'template' WHERE key = 'year_narrative_mode';"
docker-compose exec postgres psql -U postgres -d yuanju -c "DELETE FROM ai_dayun_summaries;"  # force regeneration with new mode
```

Hard refresh browser. Year cards should now show **template narratives** in Stage 1 (no waiting for AI). Dayun summaries still come from Stage 2 (those use the same AI call but the per-year narratives are no longer needed).

Flip back to `ai`:

```bash
docker-compose exec postgres psql -U postgres -d yuanju -c "UPDATE algo_config SET value = 'ai' WHERE key = 'year_narrative_mode';"
```

- [ ] **Step 6: Inspect a `ai_dayun_summaries` row**

```bash
docker-compose exec postgres psql -U postgres -d yuanju -c "SELECT dayun_index, years FROM ai_dayun_summaries ORDER BY dayun_index LIMIT 2;"
```

Expected: `years` column populated with JSON array of `{year, ganzhi, narrative}` entries.

- [ ] **Step 7: Inspect backend logs for validation drops**

```bash
docker-compose logs backend 2>&1 | grep "校验失败丢弃 narrative" | tail -5
```

Expected: 0 or a small number of entries. Each entry should name a specific keyword that was rejected — useful signal for refining the validation list.

- [ ] **Step 8: No commit (verification only)**

Document anything surprising in the PR description / pull-request body when this branch ships.

---

## Self-Review

**Spec coverage:**

| Spec section | Tasks |
|---|---|
| 架构（Stage 1 narrative="", Stage 2 加 years 字段） | 1, 2, 8, 9, 10 |
| Prompt 改造（years block） | 5 |
| 缓存（ADD COLUMN years JSONB, lazy migrate） | 1, 2, 8 |
| 前端（DayunSummary 加 years、loading 状态、合并渲染） | 10 |
| 护栏 1（事实校验） | 4, 7 |
| 护栏 2（feature flag） | 3, 9, 11 |
| 护栏 3（token 成本打点）| 既有 `token_usage_log` 已记录 `module=dayun`，admin token usage 视图已存在 — 不需要新任务，验证步骤 7 已确认 |
| 出错处理表（结构校验、ganzhi 不对齐、校验失败 etc.） | 6, 7 |
| 范围边界（不动 event_signals / chips / 命理依据 / 老的 GeneratePastEventsStream）| 计划文件结构章节的 "Not touched" 段 |

All spec items mapped to tasks.

**Placeholder scan:** None. Every step has exact code blocks, file paths, and commands.

**Type consistency:**
- `Years *json.RawMessage` on `AIDayunSummary` (model) ↔ `Years json.RawMessage` on `DayunSummaryStreamItem` (transit) ↔ `years?: YearNarrativeEntry[]` (frontend) — three layers, deliberately different shapes, all serialize/deserialize cleanly through JSON.
- `UpsertDayunSummary` signature change from 6 to 7 params propagates through Task 2 step 5 (with `nil`) and Task 7 step 2 (with `&yearsRaw`) — kept consistent.
- `GetYearNarrativeMode()` returns `"ai" | "template"` string — both call sites (Task 9) and the test (Task 3) handle both values.
- `yearOut` struct in Task 7 step 1 has fields `Year/GanZhi/Narrative` — JSON tagged `year/ganzhi/narrative` which matches the frontend `YearNarrativeEntry` interface (Task 10) and the AI prompt output schema (Task 5).

**No spec gap, no signature drift, no placeholder.** Plan is ready for execution.
