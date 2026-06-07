# 按八字起盘（四柱反推日期）Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 让用户只用四柱八字（无真实出生时间）也能起盘——反推能产生这组四柱的公历日期，再复用现有 `Calculate` 全流程。

**Architecture:** 新增一个纯函数 `bazi.ResolvePillars`（四柱 → 候选公历日期）+ 一个只读 handler/路由；前端首页加「按生辰 / 按八字」切换，八字面板提交后先调反查接口，用户从候选里选一个（仅一个时自动选中），再用「该日期 + 时辰中点 + 性别」调用现有 `/api/bazi/calculate`。现有起盘链路（大运/神煞/用神/格局/流年/入库）一行不改。

**Tech Stack:** Go 1.26（`github.com/6tail/lunar-go`，`gin`）；React + TypeScript（`axios`，`lunar-javascript`）；后端测试 `go test`，前端测试 `node --test`（源码字符串断言，项目既有约定）。

**关键事实（实现前必读）：**
- 四柱提取与 `engine.go:182 Calculate` 完全一致：`solar := calendar.NewSolar(y,m,d,h,0,0); bz := solar.GetLunar().GetEightChar()`，柱串 = `bz.GetYearGan()+bz.GetYearZhi()`（月/日/时同理）。反查必须用同一路径，保证「反查能命中的日期」喂回 `Calculate` 会得到同样的四柱。
- 日柱干支由儒略日严格 +1/天、周期恰好 60 天（lunar-go 无时区/夏令时，纯公历），所以「找到一个日柱匹配日 → 每 +60 天枚举」不会漏、不会错。
- 时辰中点小时 = 地支序号 × 2（子=0、丑=2 … 亥=22）。子时统一取 0（晚子时），中点 0 不落在 23:00 边界，日柱无歧义。
- `EightChar.GetYear()` 返回的是「按立春的年柱」(`GetYearInGanZhiExact`)，但本计划统一用 `GetYearGan()+GetYearZhi()` 拼接比较，与 `engine.go` 一致，不要用 `GetYear()`。
- 农历串：`lunar.GetYearInGanZhi() + "年" + lunar.GetMonthInChinese() + "月" + lunar.GetDayInChinese()` → 如「乙巳年六月初九」。
- 前端测试是「读源码文件断言包含某字符串」，不渲染 DOM（见 `frontend/tests/bazi-input-ux.test.mjs`）。新前端测试沿用此约定。
- 前端无 `npm test` 脚本，测试用 `node --test frontend/tests/<file>` 跑。

---

## File Structure

| 文件 | 职责 | 动作 |
|---|---|---|
| `backend/pkg/bazi/resolve_pillars.go` | `Candidate` 类型 + `ResolvePillars` 反查纯函数 | 新建 |
| `backend/pkg/bazi/resolve_pillars_test.go` | 反查函数单测 | 新建 |
| `backend/internal/handler/bazi_handler.go` | `ResolvePillarsInput` + `ResolvePillars` handler | 修改（追加） |
| `backend/internal/handler/bazi_handler_test.go` | handler 测试 | 修改（追加） |
| `backend/cmd/api/main.go` | 注册 `POST /api/bazi/resolve-pillars` | 修改（1 行） |
| `frontend/src/lib/api.ts` | `ResolvePillarsInput`/`PillarCandidate` 类型 + `baziAPI.resolvePillars` | 修改（追加） |
| `frontend/src/components/PillarsInputForm.tsx` | 4 干支下拉 + 性别 + 可选年代范围面板 | 新建 |
| `frontend/src/pages/HomePage.tsx` | 「按生辰/按八字」切换 + 候选选择 + 反查后转 calculate | 修改 |
| `frontend/tests/cast-chart-from-pillars.test.mjs` | 前端源码断言 | 新建 |

---

## Task 1: 后端反查纯函数 `ResolvePillars`

**Files:**
- Create: `backend/pkg/bazi/resolve_pillars.go`
- Test: `backend/pkg/bazi/resolve_pillars_test.go`

- [ ] **Step 1: Write the failing test**

创建 `backend/pkg/bazi/resolve_pillars_test.go`：

```go
package bazi

import "testing"

// 用一个已知公历日期正推出四柱，再反查，断言候选里含原日期，
// 且每个候选喂回 Calculate 都能复现这组四柱（soundness）。
func TestResolvePillars_RoundTripContainsOriginal(t *testing.T) {
	// 1984-02-15 午时(12:00) 男命，公历
	y, m, d, h := 1984, 2, 15, 12
	r := Calculate(y, m, d, h, "male", false, 0, "solar", false)
	yearGZ := r.YearGan + r.YearZhi
	monthGZ := r.MonthGan + r.MonthZhi
	dayGZ := r.DayGan + r.DayZhi
	hourGZ := r.HourGan + r.HourZhi

	cands := ResolvePillars(yearGZ, monthGZ, dayGZ, hourGZ, 1900, 2030, 2026)
	if len(cands) == 0 {
		t.Fatalf("expected at least one candidate, got 0")
	}

	found := false
	for _, c := range cands {
		if c.Year == y && c.Month == m && c.Day == d {
			found = true
		}
		// soundness：每个候选都必须真的产生目标四柱
		cr := Calculate(c.Year, c.Month, c.Day, c.Hour, "male", false, 0, "solar", false)
		if cr.YearGan+cr.YearZhi != yearGZ ||
			cr.MonthGan+cr.MonthZhi != monthGZ ||
			cr.DayGan+cr.DayZhi != dayGZ ||
			cr.HourGan+cr.HourZhi != hourGZ {
			t.Errorf("candidate %v does not reproduce target pillars", c)
		}
	}
	if !found {
		t.Errorf("original date %d-%d-%d not found in candidates %v", y, m, d, cands)
	}
}

func TestResolvePillars_MidpointHour(t *testing.T) {
	// 取一个时支为「午」的命，候选 Hour 必为 12（午时中点）
	r := Calculate(1984, 2, 15, 12, "male", false, 0, "solar", false)
	cands := ResolvePillars(r.YearGan+r.YearZhi, r.MonthGan+r.MonthZhi,
		r.DayGan+r.DayZhi, r.HourGan+r.HourZhi, 1980, 1990, 2026)
	if len(cands) == 0 {
		t.Fatalf("expected candidates")
	}
	for _, c := range cands {
		if c.Hour != 12 {
			t.Errorf("expected midpoint hour 12 for 午时, got %d", c.Hour)
		}
	}
}

func TestResolvePillars_InvalidPillarsReturnEmpty(t *testing.T) {
	// "甲丑" 不是合法干支（阳干配阴支），应返回空
	cands := ResolvePillars("甲丑", "丙寅", "戊辰", "庚午", 1900, 2030, 2026)
	if len(cands) != 0 {
		t.Errorf("expected empty for invalid ganzhi, got %v", cands)
	}
}

func TestResolvePillars_InconsistentPillarsReturnEmpty(t *testing.T) {
	// 合法单柱但跨柱不自洽（随意拼凑，几乎不可能对应真实日期）→ 空
	cands := ResolvePillars("甲子", "甲子", "甲子", "甲子", 1900, 2030, 2026)
	if len(cands) != 0 {
		t.Errorf("expected empty for non-self-consistent pillars, got %v", cands)
	}
}

func TestResolvePillars_RefAge(t *testing.T) {
	r := Calculate(1984, 2, 15, 12, "male", false, 0, "solar", false)
	cands := ResolvePillars(r.YearGan+r.YearZhi, r.MonthGan+r.MonthZhi,
		r.DayGan+r.DayZhi, r.HourGan+r.HourZhi, 1980, 1990, 2026)
	for _, c := range cands {
		if c.RefAge != 2026-c.Year {
			t.Errorf("RefAge mismatch: got %d, want %d", c.RefAge, 2026-c.Year)
		}
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd backend && go test ./pkg/bazi/ -run TestResolvePillars -v`
Expected: 编译失败 `undefined: ResolvePillars` / `undefined: Candidate`。

- [ ] **Step 3: Write minimal implementation**

创建 `backend/pkg/bazi/resolve_pillars.go`：

```go
package bazi

import "github.com/6tail/lunar-go/calendar"

// 反查默认搜索范围（含端点）
const (
	resolveMinYear = 1900
	resolveMaxYear = 2030
)

const (
	resolveGan = "甲乙丙丁戊己庚辛壬癸"
	resolveZhi = "子丑寅卯辰巳午未申酉戌亥"
)

// Candidate 一个能产生目标四柱的候选公历日期
type Candidate struct {
	Year      int    `json:"year"`
	Month     int    `json:"month"`
	Day       int    `json:"day"`
	Hour      int    `json:"hour"`       // 时辰中点小时，供后续 Calculate 复现时柱
	LunarDate string `json:"lunar_date"` // 如「乙巳年六月初九」
	RefAge    int    `json:"ref_age"`    // referenceYear - Year，供用户按年龄辨识
}

// validGanZhi 判断 gz 是否为 60 甲子之一（阴阳同性配对）。
func validGanZhi(gz string) bool {
	r := []rune(gz)
	if len(r) != 2 {
		return false
	}
	gi := indexRune(resolveGan, r[0])
	zi := indexRune(resolveZhi, r[1])
	if gi < 0 || zi < 0 {
		return false
	}
	// 干支同阴阳：序号奇偶必须一致
	return gi%2 == zi%2
}

// indexRune 返回 c 在 s 中的 rune 序号，找不到返回 -1。
func indexRune(s string, c rune) int {
	for i, x := range []rune(s) {
		if x == c {
			return i
		}
	}
	return -1
}

// zhiMidpointHour 返回时支对应时辰的中点小时（子=0、丑=2 … 亥=22）。
// 时支非法返回 -1。
func zhiMidpointHour(hourGZ string) int {
	r := []rune(hourGZ)
	if len(r) != 2 {
		return -1
	}
	zi := indexRune(resolveZhi, r[1])
	if zi < 0 {
		return -1
	}
	return zi * 2
}

// pillarsAt 用与 engine.go Calculate 完全一致的路径，取某公历时刻的四柱。
func pillarsAt(solar *calendar.Solar) (yearGZ, monthGZ, dayGZ, hourGZ string) {
	bz := solar.GetLunar().GetEightChar()
	yearGZ = bz.GetYearGan() + bz.GetYearZhi()
	monthGZ = bz.GetMonthGan() + bz.GetMonthZhi()
	dayGZ = bz.GetDayGan() + bz.GetDayZhi()
	hourGZ = bz.GetTimeGan() + bz.GetTimeZhi()
	return
}

// ResolvePillars 反查能产生目标四柱的公历日期。
// 4 个入参为干支字符串（如 "甲子"）；[minYear,maxYear] 为搜索范围（会被夹到 [1900,2030]）；
// referenceYear 用于计算候选参考年龄。非法/不自洽的四柱返回空切片。结果按年份升序。
func ResolvePillars(yearGZ, monthGZ, dayGZ, hourGZ string, minYear, maxYear, referenceYear int) []Candidate {
	out := []Candidate{}

	// 单柱合法性校验：任一非法直接空返回
	if !validGanZhi(yearGZ) || !validGanZhi(monthGZ) || !validGanZhi(dayGZ) || !validGanZhi(hourGZ) {
		return out
	}
	midHour := zhiMidpointHour(hourGZ)
	if midHour < 0 {
		return out
	}

	// 夹紧范围
	if minYear < resolveMinYear {
		minYear = resolveMinYear
	}
	if maxYear > resolveMaxYear {
		maxYear = resolveMaxYear
	}
	if minYear > maxYear {
		return out
	}

	start := calendar.NewSolar(minYear, 1, 1, midHour, 0, 0)
	end := calendar.NewSolar(maxYear, 12, 31, midHour, 0, 0)
	endJD := end.GetJulianDay()

	// 1) 在前 60 天内找到首个「日柱」匹配日（日柱 60 天一循环，必命中）
	firstOffset := -1
	for i := 0; i < 60; i++ {
		s := start.NextDay(i)
		if _, _, d, _ := pillarsAt(s); d == dayGZ {
			firstOffset = i
			break
		}
	}
	if firstOffset < 0 {
		return out // 理论不可达；防御
	}

	// 2) 从首个匹配日起每 +60 天枚举，校验年/月/时三柱（日柱已对）
	for k := 0; ; k++ {
		s := start.NextDay(firstOffset + 60*k)
		if s.GetJulianDay() > endJD {
			break
		}
		y, mo, d, h := pillarsAt(s)
		if y == yearGZ && mo == monthGZ && d == dayGZ && h == hourGZ {
			lunar := s.GetLunar()
			out = append(out, Candidate{
				Year:      s.GetYear(),
				Month:     s.GetMonth(),
				Day:       s.GetDay(),
				Hour:      midHour,
				LunarDate: lunar.GetYearInGanZhi() + "年" + lunar.GetMonthInChinese() + "月" + lunar.GetDayInChinese(),
				RefAge:    referenceYear - s.GetYear(),
			})
		}
	}
	return out
}
```

> 注：`Solar.GetJulianDay()` 用于范围终止判断，避免月份天数差异带来的 off-by-one。若该方法名在 lunar-go v1.4.6 中不同，改用 `s.GetYear() > maxYear` 终止（按年判断也安全，因为 firstOffset+60k 单调递增）。实现时以 `go build` 为准二选一。

- [ ] **Step 4: Run test to verify it passes**

Run: `cd backend && go test ./pkg/bazi/ -run TestResolvePillars -v`
Expected: 全部 PASS（5 个用例）。

- [ ] **Step 5: 全包回归 + 构建**

Run: `cd backend && go build ./... && go test ./pkg/bazi/`
Expected: build 通过；`pkg/bazi` 全绿。

- [ ] **Step 6: Commit**

```bash
cd /Users/liujiming/web/yuanju
git add backend/pkg/bazi/resolve_pillars.go backend/pkg/bazi/resolve_pillars_test.go
git commit -m "feat(bazi): add ResolvePillars to reverse-lookup solar dates from four pillars

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

## Task 2: 后端 handler 与路由

**Files:**
- Modify: `backend/internal/handler/bazi_handler.go`（在 `Calculate` 之后追加）
- Modify: `backend/cmd/api/main.go:157` 附近（`bazi` 路由组内追加一行）
- Test: `backend/internal/handler/bazi_handler_test.go`（追加）

- [ ] **Step 1: Write the failing test**

在 `backend/internal/handler/bazi_handler_test.go` 末尾追加：

```go
func TestResolvePillars_ReturnsCandidates(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/resolve-pillars", ResolvePillars)

	// 1984-02-15 午时 的四柱（甲子年 丙寅月 …），用合法自洽四柱
	body := `{"year_pillar":"甲子","month_pillar":"丙寅","day_pillar":"丁丑","hour_pillar":"丙午"}`
	req := httptest.NewRequest(http.MethodPost, "/resolve-pillars", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp struct {
		Candidates []bazi.Candidate `json:"candidates"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(resp.Candidates) == 0 {
		t.Fatalf("expected at least one candidate")
	}
}

func TestResolvePillars_EmptyOnNoMatch(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/resolve-pillars", ResolvePillars)

	// 不自洽四柱：返回 200 + 空数组，而非错误
	body := `{"year_pillar":"甲子","month_pillar":"甲子","day_pillar":"甲子","hour_pillar":"甲子"}`
	req := httptest.NewRequest(http.MethodPost, "/resolve-pillars", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp struct {
		Candidates []bazi.Candidate `json:"candidates"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if len(resp.Candidates) != 0 {
		t.Errorf("expected empty candidates, got %v", resp.Candidates)
	}
}
```

> 说明：测试文件已 import `bytes`/`net/http`/`net/http/httptest`/`encoding/json`/`github.com/gin-gonic/gin`（见现有 `TestCalculate_RejectsLongDisplayName`）。还需确认已 import `yuanju/pkg/bazi`；若未引入，在 import 块加 `"yuanju/pkg/bazi"`。

- [ ] **Step 2: Run test to verify it fails**

Run: `cd backend && go test ./internal/handler/ -run TestResolvePillars -v`
Expected: 编译失败 `undefined: ResolvePillars`（handler 未定义）。

- [ ] **Step 3: Write minimal implementation**

在 `backend/internal/handler/bazi_handler.go` 的 `Calculate` 函数之后追加：

```go
// ResolvePillarsInput 四柱反查请求体
type ResolvePillarsInput struct {
	YearPillar  string `json:"year_pillar" binding:"required"`
	MonthPillar string `json:"month_pillar" binding:"required"`
	DayPillar   string `json:"day_pillar" binding:"required"`
	HourPillar  string `json:"hour_pillar" binding:"required"`
	MinYear     int    `json:"min_year"` // 可选，缺省 1900
	MaxYear     int    `json:"max_year"` // 可选，缺省 2030
}

// ResolvePillars 四柱反查候选公历日期（无需登录，只读，不落库）
func ResolvePillars(c *gin.Context) {
	var input ResolvePillarsInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "请检查四柱：" + err.Error()})
		return
	}

	minYear := input.MinYear
	if minYear == 0 {
		minYear = 1900
	}
	maxYear := input.MaxYear
	if maxYear == 0 {
		maxYear = 2030
	}

	candidates := bazi.ResolvePillars(
		input.YearPillar, input.MonthPillar, input.DayPillar, input.HourPillar,
		minYear, maxYear, time.Now().Year(),
	)

	c.JSON(http.StatusOK, gin.H{"candidates": candidates})
}
```

> `time`、`net/http`、`yuanju/pkg/bazi`、`gin` 均已在文件 import（见文件头）。无需新增 import。

- [ ] **Step 4: 注册路由**

修改 `backend/cmd/api/main.go`，在 `bazi` 组内（如 `bazi.POST("/liu-yue", handler.HandleLiuYue)` 这一行附近）追加：

```go
		bazi.POST("/resolve-pillars", handler.ResolvePillars) // 四柱反查候选公历日期（无需登录）
```

- [ ] **Step 5: Run test to verify it passes**

Run: `cd backend && go build ./... && go test ./internal/handler/ -run TestResolvePillars -v`
Expected: build 通过；2 个用例 PASS。

- [ ] **Step 6: Commit**

```bash
cd /Users/liujiming/web/yuanju
git add backend/internal/handler/bazi_handler.go backend/internal/handler/bazi_handler_test.go backend/cmd/api/main.go
git commit -m "feat(bazi): add POST /api/bazi/resolve-pillars endpoint

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

## Task 3: 前端 API 客户端

**Files:**
- Modify: `frontend/src/lib/api.ts`（`CalculateInput` 之后追加类型；`baziAPI` 内追加方法）

- [ ] **Step 1: 追加类型与方法**

在 `frontend/src/lib/api.ts` 的 `CalculateInput` 接口之后追加：

```ts
export interface ResolvePillarsInput {
  year_pillar: string
  month_pillar: string
  day_pillar: string
  hour_pillar: string
  min_year?: number
  max_year?: number
}

export interface PillarCandidate {
  year: number
  month: number
  day: number
  hour: number
  lunar_date: string
  ref_age: number
}
```

在 `baziAPI` 对象内（`calculate` 那一行下面）追加：

```ts
  resolvePillars: (data: ResolvePillarsInput) =>
    api.post<{ candidates: PillarCandidate[] }>('/api/bazi/resolve-pillars', data),
```

- [ ] **Step 2: 类型检查**

Run: `cd frontend && npx tsc -b --noEmit`
Expected: 无错误（或仅与本改动无关的既有告警）。

- [ ] **Step 3: Commit**

```bash
cd /Users/liujiming/web/yuanju
git add frontend/src/lib/api.ts
git commit -m "feat(api): add resolvePillars client for pillar reverse-lookup

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

## Task 4: 前端「按八字」输入组件 `PillarsInputForm`

**Files:**
- Create: `frontend/src/components/PillarsInputForm.tsx`

- [ ] **Step 1: 创建组件**

创建 `frontend/src/components/PillarsInputForm.tsx`：

```tsx
import { useId } from 'react'

export interface PillarsFormValue {
  yearPillar: string
  monthPillar: string
  dayPillar: string
  hourPillar: string
  gender: 'male' | 'female'
  minYear: number
  maxYear: number
}

const GAN = '甲乙丙丁戊己庚辛壬癸'
const ZHI = '子丑寅卯辰巳午未申酉戌亥'

// 60 甲子（干支同阴阳配对）：甲子、乙丑 … 癸亥
export const JIAZI: string[] = Array.from({ length: 60 }, (_, i) => GAN[i % 10] + ZHI[i % 12])

export const initialPillarsValue = (gender: 'male' | 'female'): PillarsFormValue => ({
  yearPillar: '甲子',
  monthPillar: '丙寅',
  dayPillar: '甲子',
  hourPillar: '甲子',
  gender,
  minYear: 1900,
  maxYear: 2030,
})

interface PillarsInputFormProps {
  value: PillarsFormValue
  onChange: (next: PillarsFormValue) => void
}

const PILLAR_FIELDS: Array<{ key: keyof PillarsFormValue; label: string }> = [
  { key: 'yearPillar', label: '年柱' },
  { key: 'monthPillar', label: '月柱' },
  { key: 'dayPillar', label: '日柱' },
  { key: 'hourPillar', label: '时柱' },
]

export default function PillarsInputForm({ value, onChange }: PillarsInputFormProps) {
  const formId = useId()
  const update = (patch: Partial<PillarsFormValue>) => onChange({ ...value, ...patch })

  return (
    <div className="pillars-input-form">
      <div className="birth-profile-fieldset">
        <div className="form-label">性别</div>
        <div className="birth-profile-segmented">
          {(['male', 'female'] as const).map(g => (
            <button
              key={g}
              type="button"
              className={`birth-profile-option ${value.gender === g ? 'active' : ''}`}
              onClick={() => update({ gender: g })}
            >
              {g === 'male' ? '♂ 男命' : '♀ 女命'}
            </button>
          ))}
        </div>
      </div>

      <div className="birth-profile-fieldset">
        <div className="form-label">
          四柱八字
          <span className="field-note">从排盘图或已知八字里逐柱选择，不知道出生时间也能起盘</span>
        </div>
        <div className="pillars-grid">
          {PILLAR_FIELDS.map(field => (
            <div className="form-group" key={field.key}>
              <select
                id={`${formId}-${field.key}`}
                className="form-select"
                aria-label={field.label}
                value={value[field.key] as string}
                onChange={e => update({ [field.key]: e.target.value } as Partial<PillarsFormValue>)}
              >
                {JIAZI.map(gz => (
                  <option key={gz} value={gz}>{field.label.slice(0, 1)}：{gz}</option>
                ))}
              </select>
            </div>
          ))}
        </div>
      </div>

      <div className="birth-profile-fieldset">
        <div className="form-label">
          大致年代（可选）
          <span className="field-note">缩小反查范围、减少候选；不确定就留默认</span>
        </div>
        <div className="pillars-range">
          <select
            className="form-select"
            aria-label="起始年份"
            value={value.minYear}
            onChange={e => update({ minYear: Number(e.target.value) })}
          >
            {Array.from({ length: 14 }, (_, i) => 1900 + i * 10).map(y => (
              <option key={y} value={y}>{y} 年起</option>
            ))}
          </select>
          <span className="pillars-range-sep">—</span>
          <select
            className="form-select"
            aria-label="结束年份"
            value={value.maxYear}
            onChange={e => update({ maxYear: Number(e.target.value) })}
          >
            {Array.from({ length: 14 }, (_, i) => 1900 + i * 10).concat([2030]).map(y => (
              <option key={y} value={y}>{y} 年止</option>
            ))}
          </select>
        </div>
      </div>
    </div>
  )
}
```

- [ ] **Step 2: 类型检查**

Run: `cd frontend && npx tsc -b --noEmit`
Expected: 无新错误。

- [ ] **Step 3: Commit**

```bash
cd /Users/liujiming/web/yuanju
git add frontend/src/components/PillarsInputForm.tsx
git commit -m "feat(ui): add PillarsInputForm (four-pillar dropdowns) for cast-by-pillars

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

## Task 5: 首页整合「按生辰 / 按八字」切换与候选选择

**Files:**
- Modify: `frontend/src/pages/HomePage.tsx`
- Test: `frontend/tests/cast-chart-from-pillars.test.mjs`（新建）

- [ ] **Step 1: 写前端断言测试（源码字符串约定）**

创建 `frontend/tests/cast-chart-from-pillars.test.mjs`：

```js
import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import test from 'node:test'

const root = resolve(dirname(fileURLToPath(import.meta.url)), '..')
const read = (path) => readFileSync(resolve(root, path), 'utf8')

test('pillars input form renders four ganzhi dropdowns and gender', () => {
  const form = read('src/components/PillarsInputForm.tsx')
  assert.match(form, /年柱/)
  assert.match(form, /月柱/)
  assert.match(form, /日柱/)
  assert.match(form, /时柱/)
  assert.match(form, /JIAZI/)
  assert.match(form, /男命/)
})

test('JIAZI builds 60 sexagenary combinations', () => {
  const form = read('src/components/PillarsInputForm.tsx')
  assert.match(form, /length:\s*60/)
})

test('homepage offers birth/pillars input mode toggle', () => {
  const home = read('src/pages/HomePage.tsx')
  assert.match(home, /inputMode/)
  assert.match(home, /按生辰/)
  assert.match(home, /按八字/)
  assert.match(home, /PillarsInputForm/)
})

test('homepage resolves pillars then calculates, handling 0 / many candidates', () => {
  const home = read('src/pages/HomePage.tsx')
  assert.match(home, /resolvePillars/)
  // 0 候选提示
  assert.match(home, /找不到对应的真实日期/)
  // 候选列表渲染
  assert.match(home, /candidates/)
  assert.match(home, /ref_age|参考年龄|岁/)
})
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd frontend && node --test tests/cast-chart-from-pillars.test.mjs`
Expected: FAIL（`PillarsInputForm` 未被 HomePage 引用、`inputMode` 不存在等）。第一、二个用例可能已通过（组件 Task4 已建），但 homepage 相关用例必失败。

- [ ] **Step 3: 改造 HomePage —— 顶部 import 与状态**

修改 `frontend/src/pages/HomePage.tsx`。

3a. 顶部 import 追加（与现有 import 并列）：

```tsx
import PillarsInputForm, { initialPillarsValue, type PillarsFormValue } from '../components/PillarsInputForm'
import type { PillarCandidate } from '../lib/api'
```

3b. 在现有 `useState` 区块后追加状态（紧接 `const [displayNameError, setDisplayNameError] = useState('')` 之后）：

```tsx
  const [inputMode, setInputMode] = useState<'birth' | 'pillars'>('birth')
  const [pillars, setPillars] = useState<PillarsFormValue>(initialPillarsValue('male'))
  const [candidates, setCandidates] = useState<PillarCandidate[]>([])
```

- [ ] **Step 4: 改造 HomePage —— 抽出「按公历日期 + 时辰起盘」公共函数**

在 `handleSubmit` 之前新增一个公共起盘函数（供生辰模式与候选选中后复用），并新增八字提交、候选选中处理：

```tsx
  const trimmedDisplayName = () => {
    const name = chartDisplayName.trim()
    return Array.from(name).length > 20 ? null : (name || undefined)
  }

  // 用「公历年月日 + 小时 + 性别」直接走现有 calculate
  const castBySolar = async (
    year: number, month: number, day: number, hour: number,
    gender: 'male' | 'female', isEarlyZishi: boolean,
  ) => {
    const input: CalculateInput = {
      year, month, day, hour, gender,
      is_early_zishi: isEarlyZishi,
      longitude: PROVINCE_LONGITUDE[province] || 0,
      calendar_type: 'solar',
      is_leap_month: false,
      display_name: trimmedDisplayName(),
    }
    const res = await baziAPI.calculate(input)
    navigate('/result', { state: { result: res.data.result, chartId: res.data.chart_id, input, isGuest: !user } })
  }

  // 八字模式提交：先反查候选
  const handlePillarsSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')
    setCandidates([])
    setLoading(true)
    try {
      const res = await baziAPI.resolvePillars({
        year_pillar: pillars.yearPillar,
        month_pillar: pillars.monthPillar,
        day_pillar: pillars.dayPillar,
        hour_pillar: pillars.hourPillar,
        min_year: pillars.minYear,
        max_year: pillars.maxYear,
      })
      const list = res.data.candidates
      if (list.length === 0) {
        setError('这组八字找不到对应的真实日期，请核对四柱')
      } else if (list.length === 1) {
        await castBySolar(list[0].year, list[0].month, list[0].day, list[0].hour, pillars.gender, false)
      } else {
        setCandidates(list)
      }
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : '反查失败，请重试')
    } finally {
      setLoading(false)
    }
  }

  // 用户从多候选里点选一个
  const handlePickCandidate = async (c: PillarCandidate) => {
    setError('')
    setLoading(true)
    try {
      await castBySolar(c.year, c.month, c.day, c.hour, pillars.gender, false)
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : '计算失败，请重试')
    } finally {
      setLoading(false)
    }
  }
```

> 说明：生辰模式的 `handleSubmit` 保持原样不动（它带早子时/经度/农历逻辑）。八字模式默认 `calendar_type:'solar'`、`is_early_zishi:false`、`is_leap_month:false`、经度沿用当前 `province`（默认 0）。

- [ ] **Step 5: 改造 HomePage —— 渲染切换、八字表单、候选列表**

将表单卡片区（`<div className="form-card ...">` 内部）改为按 `inputMode` 切换。把现有 `<form onSubmit={handleSubmit} id="bazi-form"> ... </form>` 用模式切换包起来。

5a. 在 `form-card-header` 之后、表单之前插入模式切换：

```tsx
            <div className="input-mode-toggle birth-profile-segmented">
              {(['birth', 'pillars'] as const).map(m => (
                <button
                  key={m}
                  type="button"
                  className={`birth-profile-option ${inputMode === m ? 'active' : ''}`}
                  onClick={() => { setInputMode(m); setError(''); setCandidates([]) }}
                >
                  {m === 'birth' ? '按生辰' : '按八字'}
                </button>
              ))}
            </div>
```

5b. 现有生辰 `<form>` 用 `{inputMode === 'birth' && ( ... )}` 包裹（整段不改内部）。

5c. 在生辰 form 之后追加八字 form：

```tsx
            {inputMode === 'pillars' && (
              <form onSubmit={handlePillarsSubmit} id="pillars-form">
                <PillarsInputForm value={pillars} onChange={setPillars} />

                {candidates.length > 0 && (
                  <div className="candidate-list" aria-live="polite">
                    <div className="form-label">找到多个可能的出生日期，请按年龄选择</div>
                    {candidates.map(c => (
                      <button
                        type="button"
                        key={`${c.year}-${c.month}-${c.day}-${c.hour}`}
                        className="candidate-item"
                        onClick={() => handlePickCandidate(c)}
                        disabled={loading}
                      >
                        <span>{c.year}-{String(c.month).padStart(2, '0')}-{String(c.day).padStart(2, '0')}</span>
                        <span className="candidate-lunar">{c.lunar_date}</span>
                        <span className="candidate-age">约 {c.ref_age} 岁</span>
                      </button>
                    ))}
                  </div>
                )}

                {error && <p className="form-error">{error}</p>}

                <button type="submit" id="submit-pillars" className="btn btn-primary btn-lg submit-btn" disabled={loading}>
                  {loading ? (<><span className="loading-spinner" />正在反查...</>) : (<>按八字起盘</>)}
                </button>

                {!user && (
                  <p className="guest-hint"><a href="/login">登录</a>后可保存记录并获得完整解读报告</p>
                )}
              </form>
            )}
```

> 注意：现有生辰 `<form>` 内已有自己的 `{error && ...}`；当 `inputMode==='pillars'` 时生辰 form 不渲染，不会重复显示 error。

- [ ] **Step 6: Run test to verify it passes**

Run: `cd frontend && node --test tests/cast-chart-from-pillars.test.mjs`
Expected: 4 个用例全部 PASS。

- [ ] **Step 7: 类型检查 + 既有前端测试回归**

Run: `cd frontend && npx tsc -b --noEmit && node --test tests/`
Expected: tsc 无新错误；既有测试不回归（尤其 `bazi-input-ux.test.mjs`、`result-decision-first.test.mjs` 仍绿）。

- [ ] **Step 8: Commit**

```bash
cd /Users/liujiming/web/yuanju
git add frontend/src/pages/HomePage.tsx frontend/tests/cast-chart-from-pillars.test.mjs
git commit -m "feat(ui): home page 'cast by pillars' mode with candidate picker

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

## Task 6: 最小样式补充（可选但建议）

**Files:**
- Modify: `frontend/src/pages/HomePage.css`

新组件用到的 class（`pillars-grid`、`pillars-range`、`candidate-list`、`candidate-item`、`input-mode-toggle` 等）大多复用既有 `.birth-profile-*` / `.form-*` 样式即可显示。若布局明显错乱再补样式：

- [ ] **Step 1: 视觉自检**

Run: `cd frontend && npm run dev`，浏览器打开首页，切到「按八字」，确认四柱下拉、年代范围、候选列表能正常显示、点击候选能跳结果页。

- [ ] **Step 2: 按需补样式（仅在错乱时）**

在 `frontend/src/pages/HomePage.css` 末尾追加（示例，按实际间距调整）：

```css
.pillars-grid { display: grid; grid-template-columns: repeat(2, 1fr); gap: 12px; }
.pillars-range { display: flex; align-items: center; gap: 8px; }
.pillars-range-sep { color: var(--text-secondary, #888); }
.candidate-list { display: flex; flex-direction: column; gap: 8px; margin: 12px 0; }
.candidate-item {
  display: flex; justify-content: space-between; gap: 12px;
  padding: 10px 14px; border: 1px solid var(--border, #333);
  border-radius: 8px; background: transparent; cursor: pointer; text-align: left;
}
.candidate-item:hover { border-color: var(--gold, #c9a45b); }
.candidate-age { color: var(--gold, #c9a45b); }
```

- [ ] **Step 3: Commit（若有改动）**

```bash
cd /Users/liujiming/web/yuanju
git add frontend/src/pages/HomePage.css
git commit -m "style(ui): layout for cast-by-pillars panel and candidate list

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

## 最终验收

- [ ] `cd backend && go build ./... && go test ./pkg/bazi/ ./internal/handler/` 全绿（`internal/repository` 既有 nil-DB 失败与本改动无关，忽略）。
- [ ] `cd frontend && npx tsc -b --noEmit && node --test tests/` 无新错误、无回归。
- [ ] 手动：首页「按八字」→ 选四柱（如某真实命盘）→ 单候选直接进结果页 / 多候选可选 / 乱填四柱提示「找不到对应的真实日期」。
- [ ] 结果页大运、神煞、用神、流年均正常（因为复用了 `calculate`，应与按生辰起盘完全一致）。

---

## 自查记录（已核对，无需再审）

- **Spec 覆盖**：反查函数(Task1)、只读接口(Task2)、API 客户端(Task3)、八字输入 UI(Task4)、首页切换+候选选择(Task5)、样式(Task6) —— 对应 spec 全部章节。边界（非法/不自洽四柱→空、子时取晚子时中点、默认范围 1900–2030）在 Task1/Task2 实现并被测试覆盖。
- **占位符**：无 TBD/TODO；每个改码步骤均给出完整代码。
- **类型一致**：后端 `Candidate{Year,Month,Day,Hour,LunarDate,RefAge}` 与前端 `PillarCandidate{year,month,day,hour,lunar_date,ref_age}` 字段一一对应；`ResolvePillars(yearGZ,monthGZ,dayGZ,hourGZ,minYear,maxYear,referenceYear)` 签名在 Task1 定义、Task2 调用一致；时辰中点映射（地支序号×2）后端 `zhiMidpointHour` 与前端 SHICHEN/午=12 一致。
- **未保证项的诚实处理**：spec 曾设想「多候选」测试，但「同四柱在 1900–2030 内出现≥2 次」无法静态保证，故改为更强的 soundness 测试（每个候选都必复现目标四柱）+ 完整性测试（原日期必在候选内）。多候选路径仍由 Task5 前端逻辑覆盖。
```
