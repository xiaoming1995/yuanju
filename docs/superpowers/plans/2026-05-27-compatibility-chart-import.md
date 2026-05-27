# Compatibility Chart Import Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Let users name saved Bazi charts, import saved charts into compatibility readings, and carry chart names into compatibility participant display names.

**Architecture:** Add a nullable `display_name` field to `bazi_charts`, expose a small authenticated update endpoint, and extend existing history/profile payloads. Keep compatibility creation birth-profile based, but accept optional participant display names so imported charts can name "我" and "对方" without introducing chart-reference coupling.

**Tech Stack:** Go + Gin + PostgreSQL migrations for backend; React + TypeScript + Vite with existing static Node tests for frontend.

---

## File Map

- `backend/pkg/database/migrations/00012_add_bazi_chart_display_name.sql`: add nullable chart display name.
- `backend/internal/model/model.go`: add `DisplayName` to `BaziChart`.
- `backend/internal/model/user_profile.go`: add `DisplayName` to `UserProfileChartSummary`.
- `backend/internal/repository/repository.go`: select `display_name`, add `UpdateChartDisplayName`.
- `backend/internal/repository/user_profile_repository.go`: select `display_name` for profile recent charts.
- `backend/internal/handler/bazi_handler.go`: add `UpdateHistoryDisplayName` handler with ownership and validation.
- `backend/cmd/api/main.go`: route `PATCH /api/bazi/history/:id/display-name`.
- `backend/internal/handler/bazi_handler_test.go`: focused handler validation tests.
- `backend/internal/handler/compatibility_handler.go`: decode optional display names.
- `backend/internal/service/compatibility_service.go`: normalize and persist optional display names.
- `backend/internal/handler/compatibility_handler_test.go`: request decoding / JSON compatibility tests.
- `frontend/src/lib/api.ts`: add chart types, display-name update API, compatibility display-name fields.
- `frontend/src/components/birthProfile.ts`: add `chartToBirthProfile` helper and import source types.
- `frontend/src/pages/HistoryPage.tsx`: display/edit chart names and start compatibility import with role choice.
- `frontend/src/pages/HistoryPage.css`: styles for name editing and action row.
- `frontend/src/pages/ResultPage.tsx`: expose chart name editing and compatibility import action on detail pages.
- `frontend/src/pages/ResultPage.css`: styles for chart archive management block.
- `frontend/src/pages/CompatibilityPage.tsx`: import saved charts into self/partner panels and submit display names.
- `frontend/src/pages/CompatibilityPage.css`: import toolbar, source hint, picker modal, role chooser styles.
- `frontend/tests/compatibility-chart-import.test.mjs`: static coverage for the import flow.

---

### Task 1: Backend Chart Display Name Persistence

**Files:**
- Create: `backend/pkg/database/migrations/00012_add_bazi_chart_display_name.sql`
- Modify: `backend/internal/model/model.go`
- Modify: `backend/internal/model/user_profile.go`
- Modify: `backend/internal/repository/repository.go`
- Modify: `backend/internal/repository/user_profile_repository.go`
- Test: existing Go compile/tests

- [ ] **Step 1: Create the migration**

Add this file:

```sql
ALTER TABLE bazi_charts
ADD COLUMN IF NOT EXISTS display_name TEXT;
```

- [ ] **Step 2: Extend backend models**

In `backend/internal/model/model.go`, add `DisplayName` to `BaziChart` immediately after `Gender`:

```go
	DisplayName   string      `json:"display_name"`
```

In `backend/internal/model/user_profile.go`, add `DisplayName` to `UserProfileChartSummary` immediately after `Gender`:

```go
	DisplayName string `json:"display_name"`
```

- [ ] **Step 3: Select display_name in chart repository reads**

In `backend/internal/repository/repository.go`, update `GetChartByID` select list to include:

```sql
		       COALESCE(display_name, '') AS display_name,
```

Place it between `gender` and `year_gan`, then scan into:

```go
		&chart.DisplayName,
```

between `&chart.Gender` and `&chart.YearGan`.

Update `GetChartsByUserID` the same way: select `COALESCE(display_name, '') AS display_name` after `gender`, and scan into `&c.DisplayName` after `&c.Gender`.

- [ ] **Step 4: Add repository update function**

In `backend/internal/repository/repository.go`, add this function after `GetChartsByUserID`:

```go
func UpdateChartDisplayName(chartID, displayName string) error {
	_, err := database.DB.Exec(
		`UPDATE bazi_charts SET display_name=$1 WHERE id=$2`,
		displayName, chartID,
	)
	return err
}
```

- [ ] **Step 5: Select display_name for profile recent charts**

In `backend/internal/repository/user_profile_repository.go`, update `ListRecentChartsForProfile` to select:

```sql
		SELECT id, birth_year, birth_month, birth_day, birth_hour, gender,
		       COALESCE(display_name, '') AS display_name,
		       year_gan, year_zhi, month_gan, month_zhi, day_gan, day_zhi, hour_gan, hour_zhi,
		       COALESCE(yongshen, ''), COALESCE(jishen, ''), created_at
```

Update the scan arguments to include:

```go
			&item.DisplayName,
```

between `&item.Gender` and `&item.YearGan`.

- [ ] **Step 6: Run backend compile tests**

Run:

```bash
cd backend && go test ./internal/model ./internal/repository
```

Expected: package tests compile and pass. If repository package has no tests, `go test` still exits 0.

- [ ] **Step 7: Commit backend persistence**

```bash
git add backend/pkg/database/migrations/00012_add_bazi_chart_display_name.sql backend/internal/model/model.go backend/internal/model/user_profile.go backend/internal/repository/repository.go backend/internal/repository/user_profile_repository.go
git commit -m "feat(bazi): persist chart display names"
```

---

### Task 2: Backend Display Name Update Endpoint

**Files:**
- Modify: `backend/internal/handler/bazi_handler.go`
- Modify: `backend/cmd/api/main.go`
- Test: `backend/internal/handler/bazi_handler_test.go`

- [ ] **Step 1: Add failing handler validation tests**

Create or append to `backend/internal/handler/bazi_handler_test.go`:

```go
package handler

import (
	"strings"
	"testing"
)

func TestNormalizeChartDisplayName(t *testing.T) {
	name, err := normalizeChartDisplayName("  小王  ")
	if err != nil {
		t.Fatal(err)
	}
	if name != "小王" {
		t.Fatalf("expected trimmed name, got %q", name)
	}
}

func TestNormalizeChartDisplayName_AllowsEmpty(t *testing.T) {
	name, err := normalizeChartDisplayName("   ")
	if err != nil {
		t.Fatal(err)
	}
	if name != "" {
		t.Fatalf("expected empty name, got %q", name)
	}
}

func TestNormalizeChartDisplayName_RejectsLongName(t *testing.T) {
	_, err := normalizeChartDisplayName(strings.Repeat("命", 21))
	if err == nil {
		t.Fatal("expected long name to be rejected")
	}
	if !strings.Contains(err.Error(), "20") {
		t.Fatalf("expected error to mention length limit, got %v", err)
	}
}
```

- [ ] **Step 2: Run test and verify it fails**

Run:

```bash
cd backend && go test ./internal/handler -run 'TestNormalizeChartDisplayName' -v
```

Expected: FAIL with `undefined: normalizeChartDisplayName`.

- [ ] **Step 3: Implement normalization helper and handler**

In `backend/internal/handler/bazi_handler.go`, add `strings` and `unicode/utf8` imports if missing. Then add this helper and handler after `GetHistoryDetail`:

```go
func normalizeChartDisplayName(input string) (string, error) {
	name := strings.TrimSpace(input)
	if name == "" {
		return "", nil
	}
	if utf8.RuneCountInString(name) > 20 {
		return "", fmt.Errorf("称呼不能超过20个字符")
	}
	return name, nil
}

type updateChartDisplayNameRequest struct {
	DisplayName string `json:"display_name"`
}

func UpdateHistoryDisplayName(c *gin.Context) {
	userID, _ := c.Get("user_id")
	chartID := c.Param("id")

	var req updateChartDisplayNameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求体格式错误"})
		return
	}

	displayName, err := normalizeChartDisplayName(req.DisplayName)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	chart, err := repository.GetChartByID(chartID)
	if err != nil || chart == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "记录不存在"})
		return
	}
	if chart.UserID == nil || *chart.UserID != userID.(string) {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权操作此命盘"})
		return
	}
	if err := repository.UpdateChartDisplayName(chartID, displayName); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存称呼失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"id":           chartID,
			"display_name": displayName,
		},
	})
}
```

- [ ] **Step 4: Wire the route**

In `backend/cmd/api/main.go`, add this route immediately after `bazi.GET("/history/:id", middleware.Auth(), handler.GetHistoryDetail)`:

```go
			bazi.PATCH("/history/:id/display-name", middleware.Auth(), handler.UpdateHistoryDisplayName)
```

- [ ] **Step 5: Run handler tests**

Run:

```bash
cd backend && go test ./internal/handler -run 'TestNormalizeChartDisplayName' -v
```

Expected: PASS.

- [ ] **Step 6: Commit endpoint**

```bash
git add backend/internal/handler/bazi_handler.go backend/internal/handler/bazi_handler_test.go backend/cmd/api/main.go
git commit -m "feat(bazi): add chart display name endpoint"
```

---

### Task 3: Backend Compatibility Participant Display Names

**Files:**
- Modify: `backend/internal/handler/compatibility_handler.go`
- Modify: `backend/internal/service/compatibility_service.go`
- Test: `backend/internal/handler/compatibility_handler_test.go`

- [ ] **Step 1: Add failing request decode test**

Append to `backend/internal/handler/compatibility_handler_test.go`:

```go
func TestCreateCompatibilityReadingRequest_DecodesDisplayNames(t *testing.T) {
	body := []byte(`{
		"self_display_name": "我",
		"partner_display_name": "小王",
		"self": {"year": 1990, "month": 1, "day": 1, "hour": 0, "gender": "male"},
		"partner": {"year": 1992, "month": 6, "day": 15, "hour": 12, "gender": "female"}
	}`)
	var req CreateCompatibilityReadingRequest
	if err := json.Unmarshal(body, &req); err != nil {
		t.Fatal(err)
	}
	if req.SelfDisplayName != "我" {
		t.Fatalf("expected self display name to decode, got %q", req.SelfDisplayName)
	}
	if req.PartnerDisplayName != "小王" {
		t.Fatalf("expected partner display name to decode, got %q", req.PartnerDisplayName)
	}
}
```

- [ ] **Step 2: Run test and verify it fails**

Run:

```bash
cd backend && go test ./internal/handler -run 'TestCreateCompatibilityReadingRequest_DecodesDisplayNames' -v
```

Expected: FAIL with `req.SelfDisplayName undefined`.

- [ ] **Step 3: Extend request and service API**

In `backend/internal/handler/compatibility_handler.go`, add fields to `CreateCompatibilityReadingRequest`:

```go
	SelfDisplayName    string                    `json:"self_display_name"`
	PartnerDisplayName string                    `json:"partner_display_name"`
```

Update the service call:

```go
	detail, err := service.CreateCompatibilityReading(
		userID.(string),
		model.CompatibilityBirthProfile(req.Self),
		model.CompatibilityBirthProfile(req.Partner),
		model.CompatibilityContext{
			RelationshipStage: req.RelationshipStage,
			PrimaryQuestion:   req.PrimaryQuestion,
		},
		model.CompatibilityDisplayNames{
			Self:    req.SelfDisplayName,
			Partner: req.PartnerDisplayName,
		},
	)
```

In `backend/internal/model/compatibility.go`, add:

```go
type CompatibilityDisplayNames struct {
	Self    string
	Partner string
}
```

- [ ] **Step 4: Normalize and persist display names**

In `backend/internal/service/compatibility_service.go`, change the function signature:

```go
func CreateCompatibilityReading(userID string, selfProfile, partnerProfile model.CompatibilityBirthProfile, context model.CompatibilityContext, displayNames model.CompatibilityDisplayNames) (*model.CompatibilityDetail, error) {
```

Replace the initial context parsing block with:

```go
	context = normalizeCompatibilityContext(context)
	displayNames = normalizeCompatibilityDisplayNames(displayNames)
```

Add helper functions near `normalizeCompatibilityContext`:

```go
func normalizeCompatibilityDisplayNames(in model.CompatibilityDisplayNames) model.CompatibilityDisplayNames {
	self := strings.TrimSpace(in.Self)
	partner := strings.TrimSpace(in.Partner)
	if self == "" {
		self = "我"
	}
	if partner == "" {
		partner = "对方"
	}
	if len([]rune(self)) > 20 {
		self = string([]rune(self)[:20])
	}
	if len([]rune(partner)) > 20 {
		partner = string([]rune(partner)[:20])
	}
	return model.CompatibilityDisplayNames{Self: self, Partner: partner}
}
```

Then replace participant creation display names:

```go
	if _, err := repository.CreateCompatibilityParticipant(reading.ID, "self", displayNames.Self, selfChart.ChartHash, selfProfile, &selfRaw); err != nil {
		return nil, err
	}
	if _, err := repository.CreateCompatibilityParticipant(reading.ID, "partner", displayNames.Partner, partnerChart.ChartHash, partnerProfile, &partnerRaw); err != nil {
		return nil, err
	}
```

- [ ] **Step 5: Run compatibility handler tests**

Run:

```bash
cd backend && go test ./internal/handler -run 'Compatibility|CreateCompatibilityReadingRequest' -v
```

Expected: PASS.

- [ ] **Step 6: Commit compatibility backend**

```bash
git add backend/internal/handler/compatibility_handler.go backend/internal/handler/compatibility_handler_test.go backend/internal/model/compatibility.go backend/internal/service/compatibility_service.go
git commit -m "feat(compatibility): accept participant display names"
```

---

### Task 4: Frontend API Types and Birth Profile Helper

**Files:**
- Modify: `frontend/src/lib/api.ts`
- Modify: `frontend/src/components/birthProfile.ts`
- Create: `frontend/tests/compatibility-chart-import.test.mjs`

- [ ] **Step 1: Add failing static tests**

Create `frontend/tests/compatibility-chart-import.test.mjs`:

```js
import test from 'node:test'
import assert from 'node:assert/strict'
import fs from 'node:fs'

const read = path => fs.readFileSync(new URL(`../${path}`, import.meta.url), 'utf8')

test('API exposes chart display names and update endpoint', () => {
  const api = read('src/lib/api.ts')
  assert.match(api, /display_name\??:\s*string/)
  assert.match(api, /updateHistoryDisplayName:\s*\(id:\s*string,\s*displayName:\s*string\)/)
  assert.match(api, /\/api\/bazi\/history\/\$\{id\}\/display-name/)
})

test('compatibility create payload supports participant display names', () => {
  const api = read('src/lib/api.ts')
  assert.match(api, /self_display_name\??:\s*string/)
  assert.match(api, /partner_display_name\??:\s*string/)
})

test('birth profile helper converts saved charts into compatibility form values', () => {
  const helper = read('src/components/birthProfile.ts')
  assert.match(helper, /export function chartToBirthProfile/)
  assert.match(helper, /birth_year/)
  assert.match(helper, /calendarType:\s*chart\.calendar_type\s*\|\|\s*'solar'/)
  assert.match(helper, /isLeapMonth:\s*Boolean\(chart\.is_leap_month\)/)
})
```

- [ ] **Step 2: Run test and verify it fails**

Run:

```bash
cd frontend && node --test tests/compatibility-chart-import.test.mjs
```

Expected: FAIL because the API endpoint and helper do not exist.

- [ ] **Step 3: Add API types and functions**

In `frontend/src/lib/api.ts`, add `display_name?: string` to `UserProfileChartSummary`.

Create a reusable chart summary interface near `CalculateInput`:

```ts
export interface BaziHistoryChart {
  id: string
  birth_year: number
  birth_month: number
  birth_day: number
  birth_hour: number
  gender: 'male' | 'female' | string
  display_name?: string
  year_gan: string
  year_zhi: string
  month_gan: string
  month_zhi: string
  day_gan: string
  day_zhi: string
  hour_gan: string
  hour_zhi: string
  calendar_type?: 'solar' | 'lunar'
  is_leap_month?: boolean
  created_at: string
}
```

Extend `CreateCompatibilityReadingInput`:

```ts
  self_display_name?: string
  partner_display_name?: string
```

Add this method to `baziAPI` after `getHistoryDetail`:

```ts
  updateHistoryDisplayName: (id: string, displayName: string) =>
    api.patch<{ data: { id: string; display_name: string } }>(
      `/api/bazi/history/${id}/display-name`,
      { display_name: displayName },
    ),
```

- [ ] **Step 4: Add birth profile conversion helper**

In `frontend/src/components/birthProfile.ts`, add:

```ts
export interface BirthProfileImportSource {
  chartId: string
  displayName: string
  profile: BirthProfileFormValue
}

export interface BirthProfileChartLike {
  id: string
  birth_year: number
  birth_month: number
  birth_day: number
  birth_hour: number
  gender: string
  display_name?: string
  calendar_type?: 'solar' | 'lunar'
  is_leap_month?: boolean
}

export function chartToBirthProfile(chart: BirthProfileChartLike): BirthProfileFormValue {
  return {
    year: chart.birth_year,
    month: chart.birth_month,
    day: chart.birth_day,
    hour: chart.birth_hour,
    gender: chart.gender === 'female' ? 'female' : 'male',
    calendarType: chart.calendar_type || 'solar',
    isLeapMonth: Boolean(chart.is_leap_month),
  }
}
```

- [ ] **Step 5: Run frontend static test**

Run:

```bash
cd frontend && node --test tests/compatibility-chart-import.test.mjs
```

Expected: PASS.

- [ ] **Step 6: Commit frontend API foundation**

```bash
git add frontend/src/lib/api.ts frontend/src/components/birthProfile.ts frontend/tests/compatibility-chart-import.test.mjs
git commit -m "feat(frontend): add chart import API types"
```

---

### Task 5: History Page Chart Names and Compatibility Launch

**Files:**
- Modify: `frontend/src/pages/HistoryPage.tsx`
- Modify: `frontend/src/pages/HistoryPage.css`
- Test: `frontend/tests/compatibility-chart-import.test.mjs`

- [ ] **Step 1: Extend static tests for history page**

Append to `frontend/tests/compatibility-chart-import.test.mjs`:

```js
test('history page supports chart naming and compatibility launch role choice', () => {
  const page = read('src/pages/HistoryPage.tsx')
  const css = read('src/pages/HistoryPage.css')
  assert.match(page, /editingChartId/)
  assert.match(page, /handleSaveDisplayName/)
  assert.match(page, /compatibilityRoleChart/)
  assert.match(page, /作为我/)
  assert.match(page, /作为对方/)
  assert.match(page, /\/compatibility\?importChart=\$\{compatibilityRoleChart\.id\}&role=/)
  assert.match(css, /history-display-name/)
  assert.match(css, /history-role-dialog/)
})
```

- [ ] **Step 2: Run test and verify it fails**

Run:

```bash
cd frontend && node --test tests/compatibility-chart-import.test.mjs
```

Expected: FAIL on missing `editingChartId` and role dialog markers.

- [ ] **Step 3: Implement state and handlers**

In `HistoryPage.tsx`, import `BaziHistoryChart` type from API and replace local `Chart` interface with extension that includes `display_name?: string`, or use `BaziHistoryChart`.

Add state:

```tsx
  const [editingChartId, setEditingChartId] = useState<string | null>(null)
  const [displayNameDraft, setDisplayNameDraft] = useState('')
  const [displayNameError, setDisplayNameError] = useState('')
  const [compatibilityRoleChart, setCompatibilityRoleChart] = useState<Chart | null>(null)
```

Add helpers:

```tsx
function chartFallbackName(chart: Chart) {
  return `${genderText(chart.gender)} · ${chart.birth_year}年${chart.birth_month}月${chart.birth_day}日`
}

function chartDisplayName(chart: Chart) {
  return chart.display_name?.trim() || chartFallbackName(chart)
}
```

Add handlers inside the component:

```tsx
  const startEditDisplayName = (chart: Chart) => {
    setEditingChartId(chart.id)
    setDisplayNameDraft(chart.display_name || '')
    setDisplayNameError('')
  }

  const handleSaveDisplayName = async (chart: Chart) => {
    const nextName = displayNameDraft.trim()
    if (Array.from(nextName).length > 20) {
      setDisplayNameError('称呼不能超过20个字符')
      return
    }
    try {
      const res = await baziAPI.updateHistoryDisplayName(chart.id, nextName)
      const savedName = res.data.data.display_name
      setCharts(prev => prev.map(item => item.id === chart.id ? { ...item, display_name: savedName } : item))
      setEditingChartId(null)
      setDisplayNameDraft('')
      setDisplayNameError('')
    } catch (err: unknown) {
      setDisplayNameError(err instanceof Error ? err.message : '保存称呼失败')
    }
  }

  const launchCompatibility = (role: 'self' | 'partner') => {
    if (!compatibilityRoleChart) return
    navigate(`/compatibility?importChart=${compatibilityRoleChart.id}&role=${role}`)
  }
```

- [ ] **Step 4: Render name controls and role dialog**

Inside each history card, replace the top title area with a display-name block:

```tsx
                    <div className="history-display-name">
                      <span>{chartDisplayName(c)}</span>
                      <button
                        type="button"
                        className="history-inline-action"
                        onClick={e => {
                          e.preventDefault()
                          startEditDisplayName(c)
                        }}
                      >
                        编辑称呼
                      </button>
                    </div>
```

Render the edit row when `editingChartId === c.id`:

```tsx
                    {editingChartId === c.id && (
                      <div className="history-name-editor" onClick={e => e.preventDefault()}>
                        <input
                          className="form-input"
                          value={displayNameDraft}
                          maxLength={20}
                          placeholder="例如：我 / 小王"
                          onChange={e => setDisplayNameDraft(e.target.value)}
                        />
                        <button type="button" className="btn btn-primary" onClick={() => handleSaveDisplayName(c)}>保存</button>
                        <button type="button" className="btn btn-ghost" onClick={() => setEditingChartId(null)}>取消</button>
                        {displayNameError && <span className="history-name-error">{displayNameError}</span>}
                      </div>
                    )}
```

Add an action button beside “查看命盘”:

```tsx
                  <button
                    type="button"
                    className="history-record-action history-record-action-button"
                    onClick={e => {
                      e.preventDefault()
                      setCompatibilityRoleChart(c)
                    }}
                  >
                    用此命盘合盘
                  </button>
```

Render role dialog at the end of the page:

```tsx
        {compatibilityRoleChart && (
          <div className="history-role-dialog" role="dialog" aria-modal="true">
            <div className="history-role-dialog-panel card">
              <h2 className="serif">选择导入位置</h2>
              <p>{chartDisplayName(compatibilityRoleChart)} 要作为哪一方参与合盘？</p>
              <div className="history-role-actions">
                <button className="btn btn-primary" onClick={() => launchCompatibility('self')}>作为我</button>
                <button className="btn btn-primary" onClick={() => launchCompatibility('partner')}>作为对方</button>
                <button className="btn btn-ghost" onClick={() => setCompatibilityRoleChart(null)}>取消</button>
              </div>
            </div>
          </div>
        )}
```

- [ ] **Step 5: Add CSS**

Add to `HistoryPage.css`:

```css
.history-display-name {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
  font-weight: 700;
  color: var(--text-primary);
}

.history-inline-action,
.history-record-action-button {
  border: 0;
  background: transparent;
  color: var(--text-accent);
  cursor: pointer;
  font: inherit;
}

.history-name-editor {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
  margin-top: 10px;
}

.history-name-editor .form-input {
  min-width: 180px;
}

.history-name-error {
  color: #ef9a9a;
  font-size: 13px;
  align-self: center;
}

.history-role-dialog {
  position: fixed;
  inset: 0;
  z-index: 60;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 20px;
  background: rgba(0, 0, 0, 0.55);
}

.history-role-dialog-panel {
  width: min(420px, 100%);
  padding: 24px;
}

.history-role-actions {
  display: grid;
  gap: 10px;
  margin-top: 18px;
}
```

- [ ] **Step 6: Run frontend test**

Run:

```bash
cd frontend && node --test tests/compatibility-chart-import.test.mjs
```

Expected: PASS.

- [ ] **Step 7: Commit history UI**

```bash
git add frontend/src/pages/HistoryPage.tsx frontend/src/pages/HistoryPage.css frontend/tests/compatibility-chart-import.test.mjs
git commit -m "feat(history): add chart names and compatibility launch"
```

---

### Task 6: Result Page Detail Chart Name Controls

**Files:**
- Modify: `frontend/src/pages/ResultPage.tsx`
- Modify: `frontend/src/pages/ResultPage.css`
- Test: `frontend/tests/compatibility-chart-import.test.mjs`

- [ ] **Step 1: Extend static test**

Append:

```js
test('result page exposes chart archive naming and compatibility import action', () => {
  const page = read('src/pages/ResultPage.tsx')
  const css = read('src/pages/ResultPage.css')
  assert.match(page, /chartDisplayNameDraft/)
  assert.match(page, /handleSaveChartDisplayName/)
  assert.match(page, /用此命盘发起合盘/)
  assert.match(page, /role=self/)
  assert.match(page, /role=partner/)
  assert.match(css, /chart-archive-tools/)
})
```

- [ ] **Step 2: Run test and verify it fails**

Run:

```bash
cd frontend && node --test tests/compatibility-chart-import.test.mjs
```

Expected: FAIL on missing `chartDisplayNameDraft`.

- [ ] **Step 3: Add state and save handler**

In `ResultPage.tsx`, ensure the local chart/result state has access to `chart?.display_name` from history detail. Add state:

```tsx
  const [chartDisplayNameDraft, setChartDisplayNameDraft] = useState('')
  const [chartDisplayNameSaving, setChartDisplayNameSaving] = useState(false)
  const [chartDisplayNameError, setChartDisplayNameError] = useState('')
```

When history detail loads and sets chart/result, also set:

```tsx
        setChartDisplayNameDraft(res.data.chart?.display_name || '')
```

Add helper:

```tsx
  const handleSaveChartDisplayName = async () => {
    if (!targetId) return
    const nextName = chartDisplayNameDraft.trim()
    if (Array.from(nextName).length > 20) {
      setChartDisplayNameError('称呼不能超过20个字符')
      return
    }
    setChartDisplayNameSaving(true)
    setChartDisplayNameError('')
    try {
      const res = await baziAPI.updateHistoryDisplayName(targetId, nextName)
      setChartDisplayNameDraft(res.data.data.display_name)
    } catch (err: unknown) {
      setChartDisplayNameError(err instanceof Error ? err.message : '保存称呼失败')
    } finally {
      setChartDisplayNameSaving(false)
    }
  }

  const launchCompatibilityFromResult = (role: 'self' | 'partner') => {
    if (!targetId) return
    navigate(`/compatibility?importChart=${targetId}&role=${role}`)
  }
```

- [ ] **Step 4: Render archive tools for logged-in chart details**

Near the top of the result page content, render only when `!isGuest && targetId`:

```tsx
            {!isGuest && targetId && (
              <section className="chart-archive-tools card">
                <div>
                  <p className="chart-archive-kicker">命盘档案</p>
                  <h2 className="serif">档案称呼</h2>
                </div>
                <div className="chart-archive-editor">
                  <input
                    className="form-input"
                    value={chartDisplayNameDraft}
                    maxLength={20}
                    placeholder="例如：我 / 小王"
                    onChange={e => setChartDisplayNameDraft(e.target.value)}
                  />
                  <button className="btn btn-primary" onClick={handleSaveChartDisplayName} disabled={chartDisplayNameSaving}>
                    {chartDisplayNameSaving ? '保存中...' : '保存称呼'}
                  </button>
                </div>
                {chartDisplayNameError && <p className="chart-archive-error">{chartDisplayNameError}</p>}
                <div className="chart-archive-actions">
                  <button className="btn btn-ghost" onClick={() => launchCompatibilityFromResult('self')}>作为我发起合盘</button>
                  <button className="btn btn-ghost" onClick={() => launchCompatibilityFromResult('partner')}>作为对方发起合盘</button>
                </div>
              </section>
            )}
```

Use visible button text containing “用此命盘发起合盘” in the section heading or action description:

```tsx
                  <span>用此命盘发起合盘</span>
```

- [ ] **Step 5: Add CSS**

Add to `ResultPage.css`:

```css
.chart-archive-tools {
  display: grid;
  gap: 14px;
  margin-bottom: 18px;
  padding: 18px;
}

.chart-archive-kicker {
  margin: 0 0 4px;
  color: var(--text-accent);
  font-size: 13px;
}

.chart-archive-editor,
.chart-archive-actions {
  display: flex;
  gap: 10px;
  flex-wrap: wrap;
}

.chart-archive-editor .form-input {
  min-width: min(260px, 100%);
}

.chart-archive-error {
  margin: 0;
  color: #ef9a9a;
  font-size: 13px;
}
```

- [ ] **Step 6: Run frontend test**

Run:

```bash
cd frontend && node --test tests/compatibility-chart-import.test.mjs
```

Expected: PASS.

- [ ] **Step 7: Commit result page controls**

```bash
git add frontend/src/pages/ResultPage.tsx frontend/src/pages/ResultPage.css frontend/tests/compatibility-chart-import.test.mjs
git commit -m "feat(result): add chart archive actions"
```

---

### Task 7: Compatibility Page Import Flow

**Files:**
- Modify: `frontend/src/pages/CompatibilityPage.tsx`
- Modify: `frontend/src/pages/CompatibilityPage.css`
- Test: `frontend/tests/compatibility-chart-import.test.mjs`

- [ ] **Step 1: Extend static test**

Append:

```js
test('compatibility page imports saved charts into either profile and submits display names', () => {
  const page = read('src/pages/CompatibilityPage.tsx')
  const css = read('src/pages/CompatibilityPage.css')
  assert.match(page, /useSearchParams/)
  assert.match(page, /importChartFromHistory/)
  assert.match(page, /selfImportSource/)
  assert.match(page, /partnerImportSource/)
  assert.match(page, /导入最近命盘/)
  assert.match(page, /从命盘档案选择/)
  assert.match(page, /self_display_name/)
  assert.match(page, /partner_display_name/)
  assert.match(css, /compatibility-import-toolbar/)
  assert.match(css, /compatibility-chart-picker/)
})
```

- [ ] **Step 2: Run test and verify it fails**

Run:

```bash
cd frontend && node --test tests/compatibility-chart-import.test.mjs
```

Expected: FAIL on missing `useSearchParams` and import helpers.

- [ ] **Step 3: Add imports and state**

In `CompatibilityPage.tsx`, update imports:

```tsx
import { useEffect, useMemo, useState } from 'react'
import { useNavigate, useSearchParams } from 'react-router-dom'
```

Import API/chart helper types:

```tsx
import { baziAPI, compatibilityAPI, type BaziHistoryChart, type CompatibilityPrimaryQuestion, type CompatibilityProfileInput, type CompatibilityRelationshipStage } from '../lib/api'
import { chartToBirthProfile, initialBirthProfile, type BirthProfileFormValue, type BirthProfileImportSource } from '../components/birthProfile'
```

Add state:

```tsx
  const [searchParams] = useSearchParams()
  const [historyCharts, setHistoryCharts] = useState<BaziHistoryChart[]>([])
  const [historyLoaded, setHistoryLoaded] = useState(false)
  const [pickerRole, setPickerRole] = useState<'self' | 'partner' | null>(null)
  const [selfImportSource, setSelfImportSource] = useState<BirthProfileImportSource | null>(null)
  const [partnerImportSource, setPartnerImportSource] = useState<BirthProfileImportSource | null>(null)
```

- [ ] **Step 4: Add import helpers**

Add inside component:

```tsx
  const loadHistoryCharts = async () => {
    if (!user) {
      navigate('/login')
      return []
    }
    if (historyLoaded) return historyCharts
    const res = await baziAPI.getHistory()
    const charts = (res.data.charts || []) as BaziHistoryChart[]
    setHistoryCharts(charts)
    setHistoryLoaded(true)
    return charts
  }

  const buildImportSource = (chart: BaziHistoryChart): BirthProfileImportSource => {
    const profile = chartToBirthProfile(chart)
    return {
      chartId: chart.id,
      displayName: chart.display_name?.trim() || '',
      profile,
    }
  }

  const applyImportedChart = (role: 'self' | 'partner', chart: BaziHistoryChart) => {
    const source = buildImportSource(chart)
    if (role === 'self') {
      setSelfProfile(source.profile)
      setSelfImportSource(source)
      setActiveProfile('self')
    } else {
      setPartnerProfile(source.profile)
      setPartnerImportSource(source)
      setActiveProfile('partner')
    }
  }

  const importChartFromHistory = async (role: 'self' | 'partner', chartId: string) => {
    try {
      const res = await baziAPI.getHistoryDetail(chartId)
      applyImportedChart(role, res.data.chart as BaziHistoryChart)
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : '命盘不存在或无权访问')
    }
  }

  const importLatestChart = async (role: 'self' | 'partner') => {
    const charts = await loadHistoryCharts()
    if (charts.length === 0) {
      setError('还没有命盘档案，请先新建命盘')
      return
    }
    applyImportedChart(role, charts[0])
  }
```

Add query import effect:

```tsx
  useEffect(() => {
    const chartId = searchParams.get('importChart')
    const role = searchParams.get('role')
    if (!chartId || (role !== 'self' && role !== 'partner')) return
    if (!user) {
      navigate('/login')
      return
    }
    importChartFromHistory(role, chartId)
  }, [searchParams, user])
```

- [ ] **Step 5: Track manual modifications**

Wrap profile change handlers:

```tsx
  const handleSelfProfileChange = (next: BirthProfileFormValue) => {
    setSelfProfile(next)
  }

  const handlePartnerProfileChange = (next: BirthProfileFormValue) => {
    setPartnerProfile(next)
  }

  const isProfileModified = (source: BirthProfileImportSource | null, current: BirthProfileFormValue) => {
    if (!source) return false
    return JSON.stringify(source.profile) !== JSON.stringify(current)
  }

  const importSourceLabel = (source: BirthProfileImportSource | null, current: BirthProfileFormValue, fallback: string) => {
    if (!source) return ''
    const name = source.displayName || fallback
    return isProfileModified(source, current) ? `已基于${name}修改` : `已导入：${name}`
  }
```

Use `handleSelfProfileChange` and `handlePartnerProfileChange` in `BirthProfileForm`.

- [ ] **Step 6: Render import toolbar and picker**

Above each `BirthProfileForm`, add:

```tsx
            <div className="compatibility-import-toolbar">
              <button type="button" className="btn btn-ghost" onClick={() => importLatestChart('self')}>导入最近命盘</button>
              <button type="button" className="btn btn-ghost" onClick={async () => { await loadHistoryCharts(); setPickerRole('self') }}>从命盘档案选择</button>
            </div>
            {selfImportSource && (
              <p className="compatibility-import-source">
                {importSourceLabel(selfImportSource, selfProfile, '我的命盘')}
              </p>
            )}
```

Use the same structure for partner with `partner` role and fallback `对方命盘`.

Render picker modal near the bottom:

```tsx
        {pickerRole && (
          <div className="compatibility-chart-picker" role="dialog" aria-modal="true">
            <div className="compatibility-chart-picker-panel card">
              <div className="compatibility-chart-picker-head">
                <h2 className="serif">选择命盘档案</h2>
                <button className="btn btn-ghost" onClick={() => setPickerRole(null)}>关闭</button>
              </div>
              {historyCharts.length === 0 ? (
                <div className="compatibility-chart-picker-empty">
                  <p>还没有命盘档案。</p>
                  <button className="btn btn-primary" onClick={() => navigate('/')}>先新建命盘</button>
                </div>
              ) : (
                <div className="compatibility-chart-picker-list">
                  {historyCharts.map(chart => (
                    <button
                      key={chart.id}
                      type="button"
                      className="compatibility-chart-picker-item"
                      onClick={() => {
                        applyImportedChart(pickerRole, chart)
                        setPickerRole(null)
                      }}
                    >
                      <strong>{chart.display_name?.trim() || `${chart.birth_year}年${chart.birth_month}月${chart.birth_day}日`}</strong>
                      <span>{chart.gender === 'female' ? '女命' : '男命'} · {chart.year_gan}{chart.year_zhi} · {chart.month_gan}{chart.month_zhi} · {chart.day_gan}{chart.day_zhi} · {chart.hour_gan}{chart.hour_zhi}</span>
                    </button>
                  ))}
                </div>
              )}
            </div>
          </div>
        )}
```

- [ ] **Step 7: Submit participant display names**

In `handleSubmit`, add:

```tsx
        self_display_name: selfImportSource?.displayName || undefined,
        partner_display_name: partnerImportSource?.displayName || undefined,
```

inside the `compatibilityAPI.createReading` payload.

- [ ] **Step 8: Add CSS**

Add to `CompatibilityPage.css`:

```css
.compatibility-import-toolbar {
  display: flex;
  gap: 10px;
  flex-wrap: wrap;
  margin-bottom: 12px;
}

.compatibility-import-source {
  margin: 0 0 12px;
  color: var(--text-secondary);
  font-size: 13px;
}

.compatibility-chart-picker {
  position: fixed;
  inset: 0;
  z-index: 70;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 20px;
  background: rgba(0, 0, 0, 0.58);
}

.compatibility-chart-picker-panel {
  width: min(620px, 100%);
  max-height: min(720px, 86vh);
  overflow: auto;
  padding: 22px;
}

.compatibility-chart-picker-head {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  align-items: center;
  margin-bottom: 16px;
}

.compatibility-chart-picker-list {
  display: grid;
  gap: 10px;
}

.compatibility-chart-picker-item {
  display: grid;
  gap: 6px;
  width: 100%;
  padding: 14px;
  border: 1px solid var(--border-color);
  border-radius: 8px;
  background: rgba(255, 255, 255, 0.03);
  color: var(--text-primary);
  text-align: left;
  cursor: pointer;
}

.compatibility-chart-picker-item span,
.compatibility-chart-picker-empty {
  color: var(--text-secondary);
  font-size: 13px;
}
```

- [ ] **Step 9: Run frontend test**

Run:

```bash
cd frontend && node --test tests/compatibility-chart-import.test.mjs
```

Expected: PASS.

- [ ] **Step 10: Commit compatibility import UI**

```bash
git add frontend/src/pages/CompatibilityPage.tsx frontend/src/pages/CompatibilityPage.css frontend/tests/compatibility-chart-import.test.mjs
git commit -m "feat(compatibility): import saved charts"
```

---

### Task 8: Final Verification

**Files:**
- Review all changed files.

- [ ] **Step 1: Run focused backend tests**

Run:

```bash
cd backend && go test ./internal/handler ./internal/model ./internal/repository
```

Expected: PASS.

- [ ] **Step 2: Run frontend static test**

Run:

```bash
cd frontend && node --test tests/compatibility-chart-import.test.mjs tests/bazi-input-ux.test.mjs tests/compatibility-context-ux.test.mjs tests/compatibility-personality-fit.test.mjs
```

Expected: PASS.

- [ ] **Step 3: Run frontend lint and build**

Run:

```bash
cd frontend && npm run lint && npm run build
```

Expected: both commands exit 0.

- [ ] **Step 4: Run broader backend tests**

Run:

```bash
cd backend && go test ./...
```

Expected: PASS.

- [ ] **Step 5: Manual browser smoke test**

Start local services with the project’s normal dev setup, then verify:

1. Log in as a normal user.
2. Open `/history`.
3. Edit a chart称呼 and refresh; the称呼 remains.
4. Click “用此命盘合盘”, choose “作为对方”, and confirm `/compatibility?importChart=<id>&role=partner` fills the partner panel.
5. On `/compatibility`, use “导入最近命盘” for “我的生辰”.
6. Use “从命盘档案选择” and import a different chart.
7. Submit compatibility and confirm the result page participant names use the imported display names.
8. Open `/history/:id`, edit the称呼, then use the chart to launch compatibility as “作为我”.

- [ ] **Step 6: Final commit if verification required changes**

If verification required fixes:

```bash
git add backend frontend
git commit -m "fix: polish compatibility chart import"
```

If no changes were needed, do not create an empty commit.
