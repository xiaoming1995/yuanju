## ADDED Requirements

### Requirement: PDF report SHALL include a MingGe interpretation block

The system SHALL render a dedicated `命格解读` block in the exported Bazi PDF whenever the chart result contains `ming_ge`.

#### Scenario: MingGe block appears before analysis overview

- **WHEN** the print layout renders a Bazi result whose data includes `ming_ge`
- **THEN** the PDF content MUST include a `命格解读` block
- **AND** that block MUST appear before `命局分析总览` within the `命理解读` section

#### Scenario: MingGe block shows core MingGe fields

- **WHEN** the `命格解读` block is rendered
- **THEN** it MUST display the main MingGe name from `ming_ge`
- **AND** it MUST display the explanatory text from `ming_ge_desc` when that text is available

### Requirement: PDF MingGe interpretation SHALL support a concise local verdict

The `命格解读` block SHALL optionally show a short `本局落点` line derived from the existing structured analysis text, without requiring new backend fields.

#### Scenario: Structured analysis exists

- **WHEN** the report contains `content_structured.analysis.logic`
- **THEN** the PDF MAY display a concise `本局落点` line derived from the opening sentence of that text
- **AND** it MUST keep that line shorter than the full `命局分析总览` body

#### Scenario: Structured analysis is unavailable

- **WHEN** the report lacks structured `analysis.logic`
- **THEN** the PDF MUST still render the `命格解读` block using `ming_ge` and any available `ming_ge_desc`
- **AND** it MUST omit the `本局落点` line rather than rendering an empty placeholder

### Requirement: PDF MingGe interpretation SHALL degrade safely when MingGe is absent

The print layout SHALL remain valid for reports that do not contain MingGe data.

#### Scenario: MingGe is missing

- **WHEN** the chart result does not contain `ming_ge`
- **THEN** the print layout MUST omit the `命格解读` block
- **AND** it MUST preserve the existing `命局分析总览` and chapter rendering unchanged
