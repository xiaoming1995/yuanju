# bazi-chart-archive-naming Specification

## ADDED Requirements

### Requirement: Optional display name collected during chart creation
The chart creation form SHALL provide an optional "档案称呼" text field; if filled (after trim), the value SHALL be persisted into `chart.display_name` in the same `POST /api/bazi/calculate` call. The length rule SHALL match the existing `normalizeChartDisplayName` (UTF-8 rune count ≤ 20).

#### Scenario: User leaves the name empty or whitespace-only
- **WHEN** user submits the chart-creation form with empty or whitespace-only display_name
- **THEN** `chart.display_name` SHALL be saved as empty string
- **AND** the request SHALL still succeed (the field is optional)
- **AND** the result page SHALL render normally with no naming prompt

#### Scenario: User provides a valid name
- **WHEN** user types "小王" and submits
- **THEN** `chart.display_name` SHALL be saved as "小王" (trimmed)
- **AND** the result page SHALL NOT render any UI asking the user to name the chart again
- **AND** the saved name SHALL be visible in subsequent HistoryPage / chart-list views

#### Scenario: User exceeds 20 characters
- **WHEN** display_name (after trim) exceeds 20 characters (UTF-8 rune count)
- **THEN** the form SHALL reject the submission with an inline validation error
- **AND** no request SHALL be sent to the backend

#### Scenario: Backend receives an over-long display_name
- **WHEN** a client bypasses front-end validation and submits a display_name exceeding 20 runes
- **THEN** backend SHALL return HTTP 422 with the same error message as `PATCH /api/bazi/history/:id/display-name`
- **AND** no chart record SHALL be created

### Requirement: Result page must not duplicate naming UI
The chart result page (`ResultPage`) SHALL NOT contain any inline editor for `display_name`, nor any compatibility-launch entry that depended on co-locating with the naming editor.

#### Scenario: User views the result page after creating a chart
- **WHEN** user lands on `/result` after submitting the chart-creation form
- **THEN** the page SHALL NOT render the `.chart-archive-tools` section
- **AND** the page SHALL NOT render "命盘档案 / 档案称呼" labels, inputs, or save buttons
- **AND** the page SHALL NOT render "用此命盘发起合盘 / 作为我 / 作为对方" buttons

### Requirement: Renaming an existing chart is preserved via dedicated endpoint
The ability to rename a chart after creation SHALL remain available via the existing `PATCH /api/bazi/history/:id/display-name` endpoint and its UI in HistoryPage / list views.

#### Scenario: User renames a chart from HistoryPage
- **WHEN** user edits a chart's display_name on HistoryPage (or any list view exposing rename)
- **THEN** the existing PATCH endpoint SHALL be called with the new name
- **AND** the existing client-side normalize / length rules SHALL apply
- **AND** the chart's `display_name` SHALL be updated in place

### Requirement: Compatibility-launch entry points remain accessible from non-result pages
Launching compatibility from a saved chart SHALL remain supported from existing non-result entry points; the result page SHALL NOT be required to expose this entry.

#### Scenario: Launch compatibility from HistoryPage
- **WHEN** user clicks "用此命盘合盘" on any HistoryPage record
- **THEN** the existing role-selection modal SHALL appear unchanged
- **AND** user SHALL be able to pick "作为我" or "作为对方" and proceed to `/compatibility?importChart=...&role=...`

#### Scenario: Launch compatibility from CompatibilityPage picker
- **WHEN** user opens `/compatibility` and clicks "从命盘档案选择"
- **THEN** the existing chart picker SHALL appear unchanged
