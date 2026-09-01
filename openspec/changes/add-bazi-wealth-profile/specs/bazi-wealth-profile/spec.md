## ADDED Requirements

### Requirement: Natal wealth profile
The system SHALL compute a deterministic natal wealth profile for Bazi results using existing chart, Ten God, Fuyi, pattern, and natal-assessment data.

#### Scenario: New Bazi calculation returns wealth profile
- **WHEN** the backend calculates a Bazi result with enough deterministic chart data
- **THEN** the result includes `wealth_profile`
- **AND** `wealth_profile.grade` is one of `S`, `A`, `B`, `C`, or `D`
- **AND** `wealth_profile.score` is an integer from 0 to 100
- **AND** `wealth_profile` includes `grade_label`, `wealth_type`, `summary`, `tags`, `risk_flags`, and `evidences`

#### Scenario: Wealth profile is evidence backed
- **WHEN** `wealth_profile` is present
- **THEN** its `evidences` identify deterministic inputs and score impacts for major wealth signals
- **AND** the evidence covers relevant positive and negative drivers such as wealth-star visibility, carrying capacity, wealth-producing chains, favorable or adverse Ten God direction, and wealth-retention risks

### Requirement: Wealth grade remains natal scoped
The system SHALL keep the primary wealth grade scoped to natal wealth structure and SHALL NOT mix current Dayun activation into that grade.

#### Scenario: Current Dayun highlights wealth themes
- **WHEN** a chart has a current Dayun whose Gan/Zhi or Ten God power highlights 正财 or 偏财
- **THEN** `wealth_profile.grade` remains determined by natal wealth structure
- **AND** the profile MAY include a separate `current_hint` describing the current wealth-resource window

#### Scenario: Dayun road is unfavorable
- **WHEN** a current Dayun has money/resource themes but its road or evidence is adverse
- **THEN** the current hint frames the period as visible money/resource activity with risk or pressure
- **AND** it does not upgrade the natal wealth grade

### Requirement: Wealth profile uses conservative labels
The system SHALL present wealth levels as money/resource structure and carrying capacity, not as real-world assets, guaranteed outcomes, or investment advice.

#### Scenario: User views wealth grade copy
- **WHEN** the frontend or AI report displays `wealth_profile`
- **THEN** the copy describes wealth structure, resource visibility, carrying capacity, flow, retention, and risk-control themes
- **AND** the copy does not claim guaranteed wealth, fixed social class, exact income, investment timing, or moral worth

### Requirement: Result overview displays wealth structure
The result page SHALL surface the backend-computed wealth profile near the existing first-screen Bazi overview.

#### Scenario: Wealth profile is available
- **WHEN** the result page receives `wealth_profile`
- **THEN** it displays a compact "财富结构" summary with grade label, grade marker, score meter, summary, and capped tags
- **AND** professional mode provides an entry point to inspect wealth evidence

#### Scenario: Wealth profile is unavailable
- **WHEN** an older or partial result does not include `wealth_profile`
- **THEN** the result page remains usable without throwing an error
- **AND** it does not invent a frontend-only wealth grade
