## ADDED Requirements

### Requirement: Day-master ten-god relation matrix
The system SHALL expose a day-master-centered ten-god relation matrix for each calculated bazi chart.

#### Scenario: Chart contains a day master
- **WHEN** a bazi chart is calculated with a valid `day_gan`
- **THEN** the result includes a day-master label containing the day stem and its five-element label

#### Scenario: Heavenly stems are mapped to the day master
- **WHEN** a bazi chart contains year, month, day, and hour heavenly stems
- **THEN** the relation matrix lists each stem with pillar position, stem, element, ten-god label, relationship type, and short plain-language explanation

#### Scenario: Day stem is treated as the reference point
- **WHEN** the relation matrix renders the day heavenly stem
- **THEN** it labels that stem as "日主 / 日元" and explains that all other ten-god labels are calculated relative to it

### Requirement: Hidden-stem ten-god relation matrix
The system SHALL expose hidden-stem ten-god relationships for each earthly branch in the chart.

#### Scenario: Branch has hidden stems
- **WHEN** an earthly branch contains hidden stems
- **THEN** the relation matrix lists each hidden stem with the branch position, hidden stem, ten-god label, relationship type, and short plain-language explanation

#### Scenario: Hidden-stem data is incomplete
- **WHEN** a saved chart snapshot has missing or incomplete hidden-stem ten-god data
- **THEN** the system omits incomplete hidden-stem items or derives them deterministically without showing mismatched stem-to-ten-god pairs

### Requirement: Ten-god plain-language explanations
The system SHALL provide concise deterministic explanations for exact ten-god labels and grouped ten-god meanings.

#### Scenario: User sees an exact ten-god label
- **WHEN** the relation matrix displays a label such as "七杀", "偏印", or "食神"
- **THEN** it also displays a short explanation of what that label commonly represents to the day master

#### Scenario: User sees grouped ten-god meaning
- **WHEN** exact ten-god labels belong to the same practical group
- **THEN** the system groups them under peer, output, wealth, official, or seal meaning without losing the exact label

### Requirement: Result-page relation module
The result page SHALL render a "命主十神关系" module that makes the day-master reference rule understandable before users read dense professional chart data.

#### Scenario: Result page has relation data
- **WHEN** the result page receives a relation matrix
- **THEN** it displays the day-master label, heavenly-stem relationships, and hidden-stem relationships in a readable module near the basic chart section

#### Scenario: Basic chart remains first
- **WHEN** the result page displays bazi calculation details
- **THEN** the "基本排盘" section appears before the "命主十神关系" explanation module

#### Scenario: Mobile result page renders relation data
- **WHEN** the result page is viewed on a mobile viewport
- **THEN** the relation module uses stacked cards or collapsible sections and does not introduce horizontal overflow

#### Scenario: Relation data is absent
- **WHEN** the result page only receives legacy raw bazi fields
- **THEN** it derives the relation module from existing compatible fields or hides only the unavailable rows without breaking the page
